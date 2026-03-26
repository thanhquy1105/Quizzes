package keylock

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultCleanInterval = 1 * time.Hour
)

type KeyLock struct {
	locks         map[string]*innerLock
	cleanInterval time.Duration
	stopChan      chan struct{}
	mutex         sync.RWMutex
}

func NewKeyLock() *KeyLock {
	return &KeyLock{
		locks:         make(map[string]*innerLock),
		cleanInterval: defaultCleanInterval,
		stopChan:      make(chan struct{}),
	}
}

func (l *KeyLock) Lock(key string) {
	l.mutex.RLock()
	keyLock, ok := l.locks[key]
	if ok {
		keyLock.add()
	}
	l.mutex.RUnlock()
	if !ok {
		l.mutex.Lock()
		keyLock, ok = l.locks[key]
		if !ok {
			keyLock = newInnerLock()
			l.locks[key] = keyLock
		}
		keyLock.add()
		l.mutex.Unlock()
	}
	keyLock.Lock()
}

func (l *KeyLock) Unlock(key string) {
	l.mutex.RLock()
	keyLock, ok := l.locks[key]
	if ok {
		keyLock.done()
	}
	l.mutex.RUnlock()
	if ok {
		keyLock.Unlock()
	}
}

func (l *KeyLock) Clean() {
	l.mutex.Lock()
	for k, v := range l.locks {
		if v.count == 0 {
			delete(l.locks, k)
		}
	}
	l.mutex.Unlock()
}

func (l *KeyLock) StartCleanLoop() {
	go l.cleanLoop()
}

func (l *KeyLock) StopCleanLoop() {
	close(l.stopChan)
}

func (l *KeyLock) cleanLoop() {
	ticker := time.NewTicker(l.cleanInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.Clean()
		case <-l.stopChan:
			ticker.Stop()
			return
		}
	}
}

type innerLock struct {
	count int64
	sync.Mutex
}

func newInnerLock() *innerLock {
	return &innerLock{}
}

func (il *innerLock) add() {
	atomic.AddInt64(&il.count, 1)
}

func (il *innerLock) done() {
	atomic.AddInt64(&il.count, -1)
}
