package ws

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"btaskee-quiz/internal/server/proto"
	"btaskee-quiz/pkg/log"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// EventHandler receives network events from the websocket transprot
type EventHandler interface {
	OnMessage(conn interface{}, msgType proto.MsgType, data []byte)
	OnClose(conn interface{})
	ActiveConnInc()
	ActiveConnDec()
}

type Server struct {
	addr          string
	maxIdle       time.Duration
	decoder       proto.Protocol
	handler       EventHandler
	gorillaServer *http.Server
	certFile      string
	keyFile       string
	log.Log
}

func NewServer(addr string, maxIdle time.Duration, decoder proto.Protocol, handler EventHandler, logger log.Log, certFile, keyFile string) *Server {
	return &Server{
		addr:     addr,
		maxIdle:  maxIdle,
		decoder:  decoder,
		handler:  handler,
		Log:      logger,
		certFile: certFile,
		keyFile:  keyFile,
	}
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

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleGorillaWS)
	mux.Handle("/metrics", promhttp.Handler())

	s.gorillaServer = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	var err error
	if s.certFile != "" && s.keyFile != "" {
		s.Info("Gorilla WS server starting (WSS)", zap.String("addr", s.addr), zap.String("cert", s.certFile))
		err = s.gorillaServer.ListenAndServeTLS(s.certFile, s.keyFile)
	} else {
		s.Info("Gorilla WS server starting (WS)", zap.String("addr", s.addr))
		err = s.gorillaServer.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		s.Error("Gorilla WS server error", zap.Error(err))
		return err
	}
	return nil
}

func (s *Server) Stop() error {
	if s.gorillaServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		return s.gorillaServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) handleGorillaWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Error("failed to upgrade to websocket", zap.Error(err))
		return
	}

	gConn := &GorillaConn{
		ws: conn,
	}

	if s.handler != nil {
		s.handler.ActiveConnInc()
	}

	defer func() {
		_ = gConn.Close()
		if s.handler != nil {
			s.handler.OnClose(gConn)
			s.handler.ActiveConnDec()
		}
	}()

	if s.maxIdle > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(s.maxIdle))
	}
	conn.SetPongHandler(func(string) error {
		if s.maxIdle > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(s.maxIdle))
		}
		return nil
	})

	reader := &wsBufferReader{}

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.Error("websocket read error", zap.Error(err))
			}
			break
		}

		if s.maxIdle > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(s.maxIdle))
		}

		reader.buf = append(reader.buf, data...)

		for {
			payload, msgType, _, err := s.decoder.Decode(reader)
			if err == io.ErrShortBuffer {
				break
			}
			if err != nil {
				s.Error("proto decode error", zap.Error(err))
				return
			}
			if s.handler != nil {
				s.handler.OnMessage(gConn, msgType, payload)
			}
		}
	}
}
