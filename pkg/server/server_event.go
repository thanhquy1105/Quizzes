package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"btaskee-quiz/pkg/log"
	"btaskee-quiz/pkg/server/proto"

	"github.com/panjf2000/gnet/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (s *Server) handleMsg(conn Conn, msgType proto.MsgType, data []byte) {
	s.Debug("Received message", zap.String("type", msgType.String()), zap.Int("len", len(data)))
	s.metrics.recvMsgBytesAdd(uint64(len(data)))
	s.metrics.recvMsgCountAdd(1)

	if msgType == proto.MsgTypeHeartbeat {
		s.handleHeartbeat(conn)
	} else if msgType == proto.MsgTypeConnect {
		req := &proto.Connect{}
		err := req.Unmarshal(data)
		if err != nil {
			s.Error("unmarshal connack error", zap.Error(err))
			return
		}
		s.handleConnack(conn, req)
	} else if msgType == proto.MsgTypeRequest {

		req := s.requestObjPool.Get().(*proto.Request)
		err := req.Unmarshal(data)
		if err != nil {
			s.Error("unmarshal request error", zap.Error(err))
			return
		}

		if s.requestPool.Running() > s.opts.RequestPoolSize-10 {
			s.Warn("request pool will full", zap.Int("running", s.requestPool.Running()), zap.Int("size", s.opts.RequestPoolSize))
		}
		err = s.requestPool.Submit(func() {

			s.handleRequest(conn, req)

			s.releaseRequest(req)
		})
		if err != nil {
			s.Error("submit request error", zap.Error(err))
		}
	} else if msgType == proto.MsgTypeResp {

		resp := &proto.Response{}
		err := resp.Unmarshal(data)
		if err != nil {
			s.Error("unmarshal resp error", zap.Error(err))
			return
		}
		s.handleResp(conn, resp)
	} else if msgType == proto.MsgTypeMessage {

		msg := &proto.Message{}
		err := msg.Unmarshal(data)
		if err != nil {
			s.Error("unmarshal message error", zap.Error(err))
			return
		}
		if s.opts.MessagePoolOn {
			if s.messagePool.Running() > s.opts.MessagePoolSize-10 {
				s.Warn("message pool will full", zap.Int("running", s.messagePool.Running()), zap.Int("size", s.opts.MessagePoolSize))
			}
			s.processSingleMessage(conn, msg)
		} else {
			s.handleMessage(conn, msg)
		}

	} else if msgType == proto.MsgTypeBatchMessage {

		s.handleBatchMessage(conn, data)
	} else {
		s.Error("unknown msg type", zap.Uint8("msgType", msgType.Uint8()))
	}
}

func (s *Server) releaseRequest(r *proto.Request) {
	r.Reset()
	s.requestObjPool.Put(r)
}

func (s *Server) handleHeartbeat(conn Conn) {
	data, err := s.proto.Encode([]byte{proto.MsgTypeHeartbeat.Uint8()}, proto.MsgTypeHeartbeat)
	if err != nil {
		s.Error("encode heartbeat error", zap.Error(err))
		return
	}
	_, err = conn.Write(data)
	if err != nil {
		s.Debug("write heartbeat error", zap.Error(err))
	}
}

func (s *Server) handleConnack(conn Conn, req *proto.Connect) {

	s.Info("", zap.String("from", req.Uid))
	conn.SetContext(newConnContext(req.Uid))
	s.connManager.AddConn(req.Uid, conn)

	s.routeMapLock.RLock()
	h, ok := s.routeMap[s.opts.ConnPath]
	s.routeMapLock.RUnlock()
	if !ok {
		s.Info("route not found", zap.String("path", s.opts.ConnPath))
		return
	}
	ctx := NewContext(conn)
	ctx.connReq = req
	ctx.proto = s.proto
	h(ctx)
}

func (s *Server) handleResp(_ Conn, resp *proto.Response) {
	if s.w.IsRegistered(resp.Id) {
		s.w.Trigger(resp.Id, resp)
	} else {
		s.Error("resp id not found", zap.Uint64("id", resp.Id))
	}
}

func (s *Server) handleMessage(conn Conn, msg *proto.Message) {
	if s.opts.OnMessage != nil {
		s.opts.OnMessage(conn, msg)
	}
}

func (s *Server) handleBatchMessage(conn Conn, data []byte) {

	batchMsg := &proto.BatchMessage{}
	err := batchMsg.Decode(data)
	if err != nil {
		s.Error("Failed to decode batch message", zap.Error(err))
		return
	}

	if s.opts.LogDetailOn {
		s.Info("Received batch message",
			zap.Uint32("count", batchMsg.Count),
			zap.Int("totalSize", len(data)))
	}

	s.metrics.recvMsgCountAdd(uint64(batchMsg.Count - 1))

	for i, msg := range batchMsg.Messages {
		if msg == nil {
			s.Warn("Null message in batch", zap.Int("index", i))
			continue
		}

		err := s.processSingleMessage(conn, msg)
		if err != nil {
			s.Error("Failed to process message from batch",
				zap.Error(err),
				zap.Int("index", i),
				zap.Uint32("msgType", msg.MsgType))

		}
	}

	if s.opts.LogDetailOn {
		s.Debug("Batch message processing completed",
			zap.Uint32("processedCount", batchMsg.Count))
	}
}

func (s *Server) processSingleMessage(conn Conn, msg *proto.Message) error {
	err := s.messagePool.Submit(func() {
		s.handleMessage(conn, msg)
	})
	if err != nil {
		return fmt.Errorf("failed to submit message: %w", err)
	}
	return nil
}

func (s *Server) handleRequest(conn Conn, req *proto.Request) {
	s.Debug("Handling request", zap.String("path", req.Path), zap.Uint64("id", req.Id))
	s.routeMapLock.RLock()
	handler, ok := s.routeMap[req.Path]
	s.routeMapLock.RUnlock()
	if !ok {
		s.Debug("route not found", zap.String("path", req.Path))
		return
	}
	start := time.Now()
	ctx := NewContext(conn)
	ctx.req = req
	ctx.proto = s.proto
	handler(ctx)

	if log.Level() == zapcore.DebugLevel {
		cost := time.Since(start)
		s.Debug("request path", zap.Uint64("id", req.Id), zap.String("path", req.Path), zap.Duration("cost", cost))
	}

}

func (s *Server) Request(uid string, p string, body []byte) (*proto.Response, error) {
	conn := s.connManager.GetConn(uid)
	if conn == nil {
		return nil, errors.New("conn is nil")
	}
	r := &proto.Request{
		Id:   s.reqIDGen.Inc(),
		Path: p,
		Body: body,
	}

	if s.opts.OnRequest != nil {
		s.opts.OnRequest(conn, r)
	}

	data, err := r.Marshal()
	if err != nil {
		return nil, err
	}
	msgData, err := s.proto.Encode(data, proto.MsgTypeRequest)
	if err != nil {
		return nil, err
	}
	ch := s.w.Register(r.Id)
	err = conn.AsyncWrite(msgData, nil)
	if err != nil {
		return nil, err
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), s.opts.RequestTimeout)
	defer cancel()
	select {
	case x := <-ch:
		if x == nil {
			return nil, errors.New("unknown error")
		}
		resp := x.(*proto.Response)
		if s.opts.OnResponse != nil {
			s.opts.OnResponse(conn, resp)
		}
		return resp, nil
	case <-timeoutCtx.Done():
		s.w.Trigger(r.Id, nil)
		return nil, timeoutCtx.Err()
	}
}

func (s *Server) RequestAsync(uid string, p string, body []byte) error {
	conn := s.connManager.GetConn(uid)
	if conn == nil {
		return errors.New("conn is nil")
	}
	r := &proto.Request{
		Id:   s.reqIDGen.Inc(),
		Path: p,
		Body: body,
	}
	data, err := r.Marshal()
	if err != nil {
		return err
	}
	msgData, err := s.proto.Encode(data, proto.MsgTypeRequest)
	if err != nil {
		return err
	}
	err = conn.AsyncWrite(msgData, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Send(uid string, msg *proto.Message) error {
	conn := s.connManager.GetConn(uid)
	if conn == nil {
		return errors.New("conn is nil")
	}
	data, err := msg.Marshal()
	if err != nil {
		return err
	}
	msgData, err := s.proto.Encode(data, proto.MsgTypeMessage)
	if err != nil {
		return err
	}
	err = conn.AsyncWrite(msgData, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) OnBoot(eng gnet.Engine) (action gnet.Action) {
	s.engine = eng
	close(s.ready)
	return
}

func (s *Server) OnClose(conn gnet.Conn, err error) (action gnet.Action) {

	ctx := conn.Context()
	if ctx == nil {
		return
	}
	connCtx := ctx.(*connContext)

	s.connManager.RemoveConn(connCtx.uid.Load())
	s.routeMapLock.RLock()
	h, ok := s.routeMap[s.opts.ClosePath]
	s.routeMapLock.RUnlock()
	if !ok {
		s.Debug("route not found", zap.String("path", s.opts.ClosePath))
		return
	}

	ct := NewContext(NewGnetConn(conn))
	ct.proto = s.proto

	h(ct)

	return
}

func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	delay = time.Millisecond * 100
	return
}
