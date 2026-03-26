package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"btaskee-quiz/pkg/log"

	"github.com/RussellLuo/timingwheel"

	"btaskee-quiz/pkg/net"
	"btaskee-quiz/pkg/server/proto"

	"github.com/lni/goutils/syncutil"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet/v2"
	"go.etcd.io/etcd/pkg/v3/wait"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Server struct {
	proto  proto.Protocol
	engine gnet.Engine
	gnet.BuiltinEventEngine
	opts         *Options
	routeMapLock sync.RWMutex
	routeMap     map[string]Handler
	log.Log
	requestPool *ants.Pool
	messagePool *ants.Pool
	reqIDGen    atomic.Uint64
	w           wait.Wait
	connManager *ConnManager
	metrics     *metrics

	timingWheel *timingwheel.TimingWheel

	requestObjPool *sync.Pool
	stopper        *syncutil.Stopper
	batchRead      int

	wsEngine *net.Engine
	ready    chan struct{}
	gorillaServer *http.Server
}

func New(addr string, ops ...Option) *Server {
	opts := NewOptions()
	opts.Addr = addr
	if len(ops) > 0 {
		for _, op := range ops {
			op(opts)
		}
	}

	s := &Server{
		proto:       proto.New(),
		opts:        opts,
		routeMap:    make(map[string]Handler),
		Log:         log.NewBLog("Server"),
		w:           wait.New(),
		stopper:     syncutil.NewStopper(),
		connManager: NewConnManager(),
		metrics:     newMetrics(),
		batchRead:   100,
		timingWheel: timingwheel.NewTimingWheel(opts.TimingWheelTick, opts.TimingWheelSize),
		ready:       make(chan struct{}),
		requestObjPool: &sync.Pool{
			New: func() any {

				return &proto.Request{}
			},
		},
	}

	log.Configure(&log.Options{
		Level: zapcore.InfoLevel,
	})

	requestPool, err := ants.NewPool(opts.RequestPoolSize, ants.WithNonblocking(true), ants.WithPanicHandler(func(i interface{}) {
		s.Panic("request pool panic", zap.Any("panic", i), zap.Stack("stack"))
	}))
	if err != nil {
		s.Panic("new request pool error", zap.Error(err))
	}
	s.requestPool = requestPool

	messagePool, err := ants.NewPool(opts.MessagePoolSize, ants.WithNonblocking(true), ants.WithPanicHandler(func(i interface{}) {
		s.Panic("message pool panic", zap.Any("panic", i), zap.Stack("stack"))
	}))
	if err != nil {
		s.Panic("new message pool error", zap.Error(err))
	}
	s.messagePool = messagePool

	s.routeMap[opts.ConnPath] = func(ctx *Context) {
		req := ctx.ConnReq()
		ctx.WriteConnack(&proto.Connack{
			Id:     req.Id,
			Status: proto.StatusOK,
		})
	}

	if s.opts.WSAddr != "" || s.opts.WSSAddr != "" {
		s.wsEngine = net.NewEngine(
			net.WithAddr("tcp://0.0.0.0:0"),
			net.WithWSAddr(s.opts.WSAddr),
			net.WithWSSAddr(s.opts.WSSAddr),
			net.WithWSTLSConfig(s.opts.WSTLSConfig),
		)
	}

	return s
}

func (s *Server) Start() error {
	s.timingWheel.Start()

	s.Schedule(time.Minute*1, func() {
		s.metrics.printMetrics(fmt.Sprintf("Server:%s", s.opts.Addr))
	})

	errChan := make(chan error, 1)
	go func() {
		err := gnet.Run(s, s.opts.Addr, gnet.WithTicker(true), gnet.WithReuseAddr(true))
		if err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-s.ready:
	case <-time.After(time.Second * 5):
		return fmt.Errorf("gnet start timeout")
	}

	if s.opts.GorillaWSAddr != "" {
		go func() {
			err := s.StartGorillaWS()
			if err != nil {
				s.Error("failed to start gorilla ws server", zap.Error(err))
			}
		}()
	}

	if s.wsEngine != nil {
		s.wsEngine.OnConnect(func(conn net.Conn) error {
			return s.onWSConnect(conn)
		})
		s.wsEngine.OnData(func(conn net.Conn) error {
			return s.onWSData(conn)
		})
		s.wsEngine.OnClose(func(conn net.Conn) {
			s.onWSClose(conn)
		})
		err := s.wsEngine.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) Stop() {
	s.stopper.Stop()
	s.timingWheel.Stop()
	if s.wsEngine != nil {
		_ = s.wsEngine.Stop()
	}
	if s.gorillaServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_ = s.gorillaServer.Shutdown(ctx)
	}
	timeCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := s.engine.Stop(timeCtx)
	if err != nil {
		s.Warn("stop is error", zap.Error(err))
	}
}

func (s *Server) Schedule(interval time.Duration, f func()) *timingwheel.Timer {
	return s.timingWheel.ScheduleFunc(&everyScheduler{
		Interval: interval,
	}, f)
}

func (s *Server) Route(p string, h Handler) {
	s.routeMapLock.Lock()
	defer s.routeMapLock.Unlock()
	s.routeMap[p] = h
}

func (s *Server) OnMessage(h func(conn Conn, msg *proto.Message)) {
	s.opts.OnMessage = h
}

func (s *Server) Options() *Options {
	return s.opts
}

func (s *Server) RequestPoolRunning() int {
	return s.requestPool.Running()
}

func (s *Server) MessagePoolRunning() int {
	return s.messagePool.Running()
}

func (s *Server) onWSConnect(conn net.Conn) error {
	conn.SetMaxIdle(time.Second * 120)
	return nil
}

func (s *Server) onWSData(conn net.Conn) error {
	s.onTraffic(NewNetConn(conn))
	return nil
}

func (s *Server) onWSClose(conn net.Conn) {
	s.onClose(NewNetConn(conn))
}

func (s *Server) onTraffic(c Conn) {
	for i := 0; i < s.batchRead; i++ {
		data, msgType, _, err := s.proto.Decode(NewNetShim(c))
		if err == io.ErrShortBuffer {
			break
		}
		if err != nil {
			s.Error("ws decode error", zap.Error(err))
			_ = c.Close()
			return
		}
		s.handleMsg(c, msgType, data)
	}
}

func (s *Server) onClose(conn Conn) {
	ctx := conn.Context()
	if ctx == nil {
		return
	}
	connCtx := ctx.(*connContext)

	s.connManager.RemoveConn(connCtx.uid.Load())
	s.routeMapLock.RLock()
	h, ok := s.routeMap[s.opts.ClosePath]
	s.routeMapLock.RUnlock()
	if ok {
		ct := NewContext(conn)
		ct.proto = s.proto
		h(ct)
	}
}

type GnetConn struct {
	gnet.Conn
}

func NewGnetConn(c gnet.Conn) *GnetConn {
	return &GnetConn{Conn: c}
}

func (g *GnetConn) AsyncWrite(data []byte, callback gnet.AsyncCallback) error {
	return g.Conn.AsyncWrite(data, callback)
}

func (g *GnetConn) Write(data []byte) (int, error) {
	err := g.Conn.AsyncWrite(data, nil)
	return len(data), err
}

type NetConn struct {
	net.Conn
}

func NewNetConn(c net.Conn) *NetConn {
	return &NetConn{Conn: c}
}

func (w *NetConn) AsyncWrite(data []byte, _ gnet.AsyncCallback) error {
	if wsConn, ok := w.Conn.(net.IWSConn); ok {
		return wsConn.WriteServerBinary(data)
	}
	_, err := w.Conn.Write(data)
	return err
}

func (w *NetConn) Write(data []byte) (int, error) {
	if wsConn, ok := w.Conn.(net.IWSConn); ok {
		err := wsConn.WriteServerBinary(data)
		return len(data), err
	}
	return w.Conn.Write(data)
}

func (w *NetConn) Peek(n int) ([]byte, error) {
	return w.Conn.Peek(n)
}

func (w *NetConn) Discard(n int) (int, error) {
	return w.Conn.Discard(n)
}

func (w *NetConn) InboundBuffered() int {
	if w.Conn.InboundBuffer() == nil {
		return 0
	}
	return w.Conn.InboundBuffer().BoundBufferSize()
}

type NetShim struct {
	conn *NetConn
}

func NewNetShim(c Conn) *NetShim {
	return &NetShim{conn: c.(*NetConn)}
}

func (w *NetShim) InboundBuffered() int {
	return w.conn.InboundBuffered()
}

func (w *NetShim) Peek(n int) ([]byte, error) {
	return w.conn.Peek(n)
}

func (w *NetShim) Discard(n int) (int, error) {
	return w.conn.Discard(n)
}

type everyScheduler struct {
	Interval time.Duration
}

func (s *everyScheduler) Next(prev time.Time) time.Time {
	return prev.Add(s.Interval)
}
