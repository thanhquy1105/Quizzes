package wkserver

import (
	"sync/atomic"
)

type metrics struct {
	recvMsgCount atomic.Uint64
	recvMsgBytes atomic.Uint64
	sendMsgCount atomic.Uint64
	sendMsgBytes atomic.Uint64
}

func newMetrics() *metrics {
	return &metrics{}
}

func (m *metrics) recvMsgCountAdd(v uint64) uint64 {
	return m.recvMsgCount.Add(v)
}

func (m *metrics) recvMsgCountSub(v uint64) uint64 {
	return m.recvMsgCount.Add(-v)
}

func (m *metrics) recvMsgBytesAdd(v uint64) {
	m.recvMsgBytes.Add(v)
}

func (m *metrics) recvMsgBytesSub(v uint64) {
	m.recvMsgBytes.Add(-v)
}

func (m *metrics) sendMsgCountAdd(v uint64) uint64 {
	return m.sendMsgCount.Add(v)
}

func (m *metrics) sendMsgCountSub(v uint64) uint64 {
	return m.sendMsgCount.Add(-v)
}

func (m *metrics) sendMsgBytesAdd(v uint64) {
	m.sendMsgBytes.Add(v)
}

func (m *metrics) sendMsgBytesSub(v uint64) {
	m.sendMsgBytes.Add(-v)
}

func (m *metrics) printMetrics(prefix string) {

}
