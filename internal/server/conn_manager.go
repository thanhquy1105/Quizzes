package server

import (
	"sync"
)

type ConnManager struct {
	connMapLock sync.RWMutex
	connMap     map[string]Conn
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connMap: make(map[string]Conn),
	}
}

func (c *ConnManager) AddConn(username string, conn Conn) {
	c.connMapLock.Lock()
	defer c.connMapLock.Unlock()
	c.connMap[username] = conn
}

func (c *ConnManager) GetConn(username string) Conn {
	c.connMapLock.RLock()
	defer c.connMapLock.RUnlock()
	return c.connMap[username]
}

func (c *ConnManager) RemoveConn(username string) {
	c.connMapLock.Lock()
	defer c.connMapLock.Unlock()
	delete(c.connMap, username)
}
