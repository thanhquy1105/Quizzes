package net

import (
	"bytes"
	"fmt"
	"os"
	"syscall"

	"btaskee-quiz/pkg/log"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type ReactorSub struct {
	eg         *Engine
	idx        int
	ReadBuffer []byte
	log.Log
	cache     bytes.Buffer
	connCount atomic.Int32
}

func NewReactorSub(eg *Engine, index int) *ReactorSub {
	return &ReactorSub{
		eg:         eg,
		idx:        index,
		ReadBuffer: make([]byte, eg.options.ReadBufferSize),
		Log:        log.NewBLog(fmt.Sprintf("ReactorSub-%d", index)),
	}
}

func (r *ReactorSub) Start() error {
	fmt.Println("warnwindowsBUGLinux")
	return nil
}

func (r *ReactorSub) Stop() error {
	return nil
}

func (r *ReactorSub) DeleteFd(conn Conn) error {

	return nil
}

func (r *ReactorSub) ConnInc() {
	r.connCount.Inc()
}
func (r *ReactorSub) ConnDec() {
	r.connCount.Dec()
}

func (r *ReactorSub) AddConn(conn Conn) error {
	r.eg.AddConn(conn)
	r.ConnInc()

	go r.readLoop(conn)
	return nil
}

func (r *ReactorSub) CloseConn(c Conn, er error) (rerr error) {
	r.Debug("connection error", zap.Error(er))
	return c.Close()
}

func (r *ReactorSub) AddWrite(conn Conn) error {
	go conn.Flush()
	return nil
}

func (r *ReactorSub) AddRead(conn Conn) error {
	return nil
}

func (r *ReactorSub) RemoveRead(conn Conn) error {
	return nil
}

func (r *ReactorSub) RemoveWrite(conn Conn) error {
	return nil
}

func (r *ReactorSub) readLoop(conn Conn) {
	for {
		n, err := conn.ReadToInboundBuffer()
		if err != nil {
			if err == syscall.EAGAIN {
				continue
			}
			r.Error("readLoop error", zap.Error(err))
			if err1 := r.CloseConn(conn, err); err1 != nil {
				r.Warn("failed to close conn", zap.Error(err1))
			}
			return
		}
		if n == 0 {
			r.CloseConn(conn, os.NewSyscallError("read", syscall.ECONNRESET))
			return
		}
		if err = r.eg.eventHandler.OnData(conn); err != nil {
			if err == syscall.EAGAIN {
				continue
			}
			if err1 := r.CloseConn(conn, err); err1 != nil {
				r.Warn("failed to close conn", zap.Error(err1))
			}
			r.Warn("failed to call OnData", zap.Error(err))
			return
		}
	}

}
