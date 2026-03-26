package server

import (
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/panjf2000/gnet/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type GorillaConn struct {
	ws  *websocket.Conn
	mu  sync.Mutex
	ctx interface{}
}

func (c *GorillaConn) RemoteAddr() net.Addr {
	return c.ws.RemoteAddr()
}

func (c *GorillaConn) LocalAddr() net.Addr {
	return c.ws.LocalAddr()
}

func (c *GorillaConn) Write(data []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.ws.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

func (c *GorillaConn) AsyncWrite(data []byte, callback gnet.AsyncCallback) error {
	_, err := c.Write(data)
	if callback != nil {
		callback(NewGnetConn(nil), err) // Gorilla doesn't expose a gnet conn
	}
	return err
}

func (c *GorillaConn) Context() interface{} {
	return c.ctx
}

func (c *GorillaConn) SetContext(ctx interface{}) {
	c.ctx = ctx
}

func (c *GorillaConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ws.Close()
}

// wsBufferReader implements proto.Reader
type wsBufferReader struct {
	buf []byte
}

func (w *wsBufferReader) InboundBuffered() int {
	return len(w.buf)
}

func (w *wsBufferReader) Peek(n int) ([]byte, error) {
	if len(w.buf) < n {
		return nil, io.ErrShortBuffer
	}
	return w.buf[:n], nil
}

func (w *wsBufferReader) Discard(n int) (int, error) {
	if len(w.buf) < n {
		return 0, io.ErrShortBuffer
	}
	w.buf = w.buf[n:]
	return n, nil
}

func (s *Server) StartGorillaWS() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleGorillaWS)
	mux.Handle("/metrics", promhttp.Handler())

	s.gorillaServer = &http.Server{
		Addr:    s.opts.GorillaWSAddr,
		Handler: mux,
	}

	s.Info("Gorilla WS server starting", zap.String("addr", s.opts.GorillaWSAddr))
	err := s.gorillaServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		s.Error("Gorilla WS server error", zap.Error(err))
		return err
	}
	return nil
}

func (s *Server) handleGorillaWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Error("failed to upgrade to websocket", zap.Error(err))
		return
	}

	conn := &GorillaConn{
		ws: ws,
	}

	s.metrics.ActiveConnInc()
	defer func() {
		_ = conn.Close()
		s.onClose(conn)
		s.metrics.ActiveConnDec()
	}()

	if s.opts.MaxIdle > 0 {
		_ = ws.SetReadDeadline(time.Now().Add(s.opts.MaxIdle))
	}
	ws.SetPongHandler(func(string) error {
		if s.opts.MaxIdle > 0 {
			_ = ws.SetReadDeadline(time.Now().Add(s.opts.MaxIdle))
		}
		return nil
	})

	reader := &wsBufferReader{}

	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.Error("websocket read error", zap.Error(err))
			}
			break
		}

		if s.opts.MaxIdle > 0 {
			_ = ws.SetReadDeadline(time.Now().Add(s.opts.MaxIdle))
		}

		reader.buf = append(reader.buf, data...)

		for {
			payload, msgType, _, err := s.proto.Decode(reader)
			if err == io.ErrShortBuffer {
				break
			}
			if err != nil {
				s.Error("proto decode error", zap.Error(err))
				return
			}
			s.handleMsg(conn, msgType, payload)
		}
	}
}
