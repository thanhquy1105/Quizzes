package ws

import (
	"net"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/panjf2000/gnet/v2"
)

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
		callback(nil, err) // Gorilla doesn't expose a gnet conn
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
