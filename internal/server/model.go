package server

import (
	"net"
	"time"

	"btaskee-quiz/pkg/log"
	"btaskee-quiz/internal/server/proto"

	"github.com/panjf2000/gnet/v2"
	"go.uber.org/zap"
)

type Conn interface {
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	Write(data []byte) (int, error)
	AsyncWrite(data []byte, callback gnet.AsyncCallback) error
	Context() interface{}
	SetContext(interface{})
	Close() error
}

type Handler func(c *Context)

type Context struct {
	conn    Conn
	req     *proto.Request
	connReq *proto.Connect
	proto   proto.Protocol
	log.Log
}

func NewContext(conn Conn) *Context {

	return &Context{
		conn:  conn,
		proto: proto.New(),
		Log:   log.NewBLog("Context"),
	}
}

func (c *Context) Write(data []byte) {
	var id uint64 = 0
	if c.req != nil {
		id = c.req.Id
	}
	resp := &proto.Response{
		Id:        id,
		Status:    proto.StatusOK,
		Body:      data,
		Timestamp: time.Now().UnixMilli(),
	}
	respData, err := resp.Marshal()
	if err != nil {
		c.Debug("marshal is error", zap.Error(err))
		return
	}
	msgData, err := c.proto.Encode(respData, proto.MsgTypeResp)
	if err != nil {
		c.Debug("encode is error", zap.Error(err))
		return
	}
	err = c.conn.AsyncWrite(msgData, nil)
	if err != nil {
		c.Debug("WriteToOutboundBuffer is error", zap.Error(err))
		return
	}
}

func (c *Context) WriteOk() {
	var id uint64 = 0
	if c.req != nil {
		id = c.req.Id
	}
	resp := &proto.Response{
		Id:     id,
		Status: proto.StatusOK,
	}
	respData, err := resp.Marshal()
	if err != nil {
		c.Debug("marshal is error", zap.Error(err))
		return
	}
	msgData, err := c.proto.Encode(respData, proto.MsgTypeResp)
	if err != nil {
		c.Debug("encode is error", zap.Error(err))
		return
	}
	err = c.conn.AsyncWrite(msgData, nil)
	if err != nil {
		c.Debug("WriteToOutboundBuffer is error", zap.Error(err))
		return
	}
}

func (c *Context) WriteErr(err error) {
	c.WriteErrorAndStatus(err, proto.StatusError)
}

func (c *Context) WriteErrorAndStatus(err error, status proto.Status) {
	var id uint64 = 0
	if c.req != nil {
		id = c.req.Id
	}
	resp := &proto.Response{
		Id:     id,
		Status: status,
		Body:   []byte(err.Error()),
	}
	respData, err := resp.Marshal()
	if err != nil {
		c.Debug("marshal is error", zap.Error(err))
		return
	}
	msgData, err := c.proto.Encode(respData, proto.MsgTypeResp)
	if err != nil {
		c.Debug("encode is error", zap.Error(err))
		return
	}
	err = c.conn.AsyncWrite(msgData, nil)
	if err != nil {
		c.Debug("WriteToOutboundBuffer is error", zap.Error(err))
		return
	}
}

func (c *Context) Body() []byte {
	return c.req.Body
}

func (c *Context) ConnReq() *proto.Connect {
	return c.connReq
}

func (c *Context) WriteConnack(connack *proto.Connack) {
	data, err := connack.Marshal()
	if err != nil {
		c.Info("marshal is error", zap.Error(err))
		return
	}
	msgData, err := c.proto.Encode(data, proto.MsgTypeConnack)
	if err != nil {
		c.Info("encode is error", zap.Error(err))
		return
	}
	err = c.conn.AsyncWrite(msgData, nil)
	if err != nil {
		c.Info("asyncWrite is error", zap.Error(err))
		return
	}
}

func (c *Context) Conn() Conn {

	return c.conn
}
