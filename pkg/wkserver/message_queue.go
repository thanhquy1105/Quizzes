package wkserver

import (
	"sync"

	"btaskee-quiz/pkg/wklog"
	"btaskee-quiz/pkg/wkserver/proto"
	"github.com/panjf2000/gnet/v2"
	"go.uber.org/zap"
)

type message struct {
	msgType proto.MsgType
	data    []byte
	conn    gnet.Conn
}

func (m *message) size() int {
	return len(m.data)
}

type messageQueue struct {
	ch            chan struct{}
	rl            *RateLimiter
	lazyFreeCycle uint64
	size          uint64
	left          []*message
	right         []*message
	nodrop        []*message
	mu            sync.Mutex
	leftInWrite   bool
	idx           uint64
	oldIdx        uint64
	cycle         uint64
	wklog.Log
}

func (q *messageQueue) add(msg *message) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.idx >= q.size {
		return false
	}

	if !q.tryAdd(msg) {
		return false
	}

	w := q.targetQueue()
	w[q.idx] = msg
	q.idx++
	return true

}

func (q *messageQueue) mustAdd(msg *message) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.nodrop = append(q.nodrop, msg)
}

func (q *messageQueue) get() []*message {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.cycle++
	sz := q.idx
	q.idx = 0
	t := q.targetQueue()
	q.leftInWrite = !q.leftInWrite
	q.gc()
	q.oldIdx = sz
	if q.rl.Enabled() {
		q.rl.Set(0)
	}
	if len(q.nodrop) == 0 {
		return t[:sz]
	}

	var result []*message
	if len(q.nodrop) > 0 {
		ssm := q.nodrop
		q.nodrop = make([]*message, 0)
		result = append(result, ssm...)
	}
	return append(result, t[:sz]...)
}

func (q *messageQueue) targetQueue() []*message {
	var t []*message
	if q.leftInWrite {
		t = q.left
	} else {
		t = q.right
	}
	return t
}

func (q *messageQueue) tryAdd(msg *message) bool {
	if !q.rl.Enabled() {
		return true
	}
	if q.rl.RateLimited() {
		q.Warn("rate limited dropped", zap.Uint8("msgType", msg.msgType.Uint8()))
		return false
	}
	q.rl.Increase(uint64(msg.size()))
	return true
}

func (q *messageQueue) gc() {
	if q.lazyFreeCycle > 0 {
		oldq := q.targetQueue()
		if q.lazyFreeCycle == 1 {
			for i := uint64(0); i < q.oldIdx; i++ {
				oldq[i].data = nil
			}
		} else if q.cycle%q.lazyFreeCycle == 0 {
			for i := uint64(0); i < q.size; i++ {
				oldq[i].data = nil
			}
		}
	}
}

func (q *messageQueue) notify() {
	if q.ch != nil {
		select {
		case q.ch <- struct{}{}:
		default:
		}
	}
}

func (q *messageQueue) notifyCh() <-chan struct{} {
	return q.ch
}
