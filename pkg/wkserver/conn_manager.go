package wkserver

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

func (c *ConnManager) AddConn(uid string, conn Conn) {
	c.connMapLock.Lock()
	defer c.connMapLock.Unlock()
	c.connMap[uid] = conn
}

func (c *ConnManager) GetConn(uid string) Conn {
	c.connMapLock.RLock()
	defer c.connMapLock.RUnlock()
	return c.connMap[uid]
}

func (c *ConnManager) RemoveConn(uid string) {
	c.connMapLock.Lock()
	defer c.connMapLock.Unlock()
	delete(c.connMap, uid)
}
