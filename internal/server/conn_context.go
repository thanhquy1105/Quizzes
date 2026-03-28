package server

import "go.uber.org/atomic"

type ConnContext struct {
	username atomic.String
	token    atomic.String
}

func NewConnContext(username string, token string) *ConnContext {
	c := &ConnContext{}
	c.username.Store(username)
	c.token.Store(token)
	return c
}

func (c *ConnContext) Username() string {
	return c.username.Load()
}

func (c *ConnContext) Token() string {
	return c.token.Load()
}
