package server

import "go.uber.org/atomic"

type ConnContext struct {
	uid   atomic.String
	token atomic.String
}

func NewConnContext(uid string, token string) *ConnContext {
	c := &ConnContext{}
	c.uid.Store(uid)
	c.token.Store(token)
	return c
}

func (c *ConnContext) UID() string {
	return c.uid.Load()
}

func (c *ConnContext) Token() string {
	return c.token.Load()
}
