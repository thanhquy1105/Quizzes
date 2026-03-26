//go:build linux
// +build linux

package netpoll

import (
	"golang.org/x/sys/unix"
)

type IOEvent = uint32

const (
	InitPollEventsCap = 128

	MaxPollEventsCap = 1024

	MinPollEventsCap = 32

	MaxAsyncTasksAtOneTime = 256

	ErrEvents = unix.EPOLLERR | unix.EPOLLHUP | unix.EPOLLRDHUP

	OutEvents = ErrEvents | unix.EPOLLOUT

	InEvents = ErrEvents | unix.EPOLLIN | unix.EPOLLPRI
)

type eventList struct {
	size   int
	events []unix.EpollEvent
}

func newEventList(size int) *eventList {
	return &eventList{size, make([]unix.EpollEvent, size)}
}

func (el *eventList) expand() {
	if newSize := el.size << 1; newSize <= MaxPollEventsCap {
		el.size = newSize
		el.events = make([]unix.EpollEvent, newSize)
	}
}

func (el *eventList) shrink() {
	if newSize := el.size >> 1; newSize >= MinPollEventsCap {
		el.size = newSize
		el.events = make([]unix.EpollEvent, newSize)
	}
}
