package wkutil

import (
	"sync"
	"sync/atomic"
)

type WaitGroupWrapper struct {
	sync.WaitGroup
	Name  string
	count int64
}

func NewWaitGroupWrapper(name string) *WaitGroupWrapper {
	return &WaitGroupWrapper{
		Name: name,
	}
}

func (w *WaitGroupWrapper) Wrap(cb func()) {
	w.Add(1)
	atomic.AddInt64(&w.count, 1)
	go func() {
		cb()
		w.Done()
		atomic.AddInt64(&w.count, -1)
	}()
}

func (w *WaitGroupWrapper) GoroutineCount() int64 {
	return atomic.LoadInt64(&w.count)
}
