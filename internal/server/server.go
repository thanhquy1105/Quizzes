package server

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"btaskee-quiz/pkg/log"

	"btaskee-quiz/internal/server/proto"
	"btaskee-quiz/internal/transport/ws"

	"github.com/RussellLuo/timingwheel"

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
	metrics     *Metrics

	timingWheel *timingwheel.TimingWheel

	requestObjPool *sync.Pool
	batchRead      int

	ready       chan struct{}
	wsTransport *ws.Server
}

func New(addr string, ops ...Option) *Server {
	opts := NewOptions()
	opts.TCPAddr = addr
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

	return s
}

func (s *Server) Start() error {
	s.timingWheel.Start()

	s.Schedule(time.Minute*1, func() {
		s.metrics.PrintMetrics(fmt.Sprintf("Server:%s", s.opts.TCPAddr))
	})

	errChan := make(chan error, 1)
	go func() {
		err := gnet.Run(s, s.opts.TCPAddr, gnet.WithTicker(true), gnet.WithReuseAddr(true))
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

	if s.opts.WSAddr != "" {
		s.wsTransport = ws.NewServer(
			s.opts.WSAddr,
			s.opts.MaxIdle,
			s.proto,
			&wsTransportHandler{server: s},
			log.NewBLog("GorillaWS"),
		)
		go func() {
			err := s.wsTransport.Start()
			if err != nil {
				s.Error("failed to start gorilla ws server", zap.Error(err))
			}
		}()
	}

	return nil
}

func (s *Server) Stop() {
	s.timingWheel.Stop()
	if s.wsTransport != nil {
		_ = s.wsTransport.Stop()
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

func (s *Server) Metrics() *Metrics {
	return s.metrics
}

func (s *Server) OnTraffic(c gnet.Conn) gnet.Action {
	for i := 0; i < s.batchRead; i++ {
		data, msgType, _, err := s.proto.Decode(c)
		if err == io.ErrShortBuffer {
			break
		}
		if err != nil {
			s.Error("tcp decode error", zap.Error(err))
			_ = c.Close()
			return gnet.Close
		}
		s.handleMsg(NewGnetConn(c), msgType, data)
	}
	return gnet.None
}

func (s *Server) onClose(conn Conn) {
	ctx := conn.Context()
	if ctx == nil {
		return
	}
	connCtx, ok := ctx.(*ConnContext)
	if !ok {
		s.Error("invalid connection context type", zap.Any("ctx", ctx))
		return
	}

	s.connManager.RemoveConn(connCtx.Username())
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

type everyScheduler struct {
	Interval time.Duration
}

func (s *everyScheduler) Next(prev time.Time) time.Time {
	return prev.Add(s.Interval)
}

type wsTransportHandler struct {
	server *Server
}

func (h *wsTransportHandler) OnMessage(conn interface{}, msgType proto.MsgType, data []byte) {
	h.server.handleMsg(conn.(Conn), msgType, data)
}

func (h *wsTransportHandler) OnClose(conn interface{}) {
	h.server.onClose(conn.(Conn))
}

func (h *wsTransportHandler) ActiveConnInc() {
	if metrics := h.server.Metrics(); metrics != nil {
		metrics.ActiveConnInc()
	}
}

func (h *wsTransportHandler) ActiveConnDec() {
	if metrics := h.server.Metrics(); metrics != nil {
		metrics.ActiveConnDec()
	}
}
