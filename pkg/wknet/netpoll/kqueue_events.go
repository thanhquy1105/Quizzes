//go:build freebsd || dragonfly || darwin
// +build freebsd dragonfly darwin

package netpoll

import "golang.org/x/sys/unix"

type IOEvent = int16

const (
	InitPollEventsCap = 64

	MaxPollEventsCap = 512

	MinPollEventsCap = 16

	MaxAsyncTasksAtOneTime = 128
)

type eventList struct {
	size   int
	events []unix.Kevent_t
}

func newEventList(size int) *eventList {
	return &eventList{size, make([]unix.Kevent_t, size)}
}

func (el *eventList) expand() {
	if newSize := el.size << 1; newSize <= MaxPollEventsCap {
		el.size = newSize
		el.events = make([]unix.Kevent_t, newSize)
	}
}

func (el *eventList) shrink() {
	if newSize := el.size >> 1; newSize >= MinPollEventsCap {
		el.size = newSize
		el.events = make([]unix.Kevent_t, newSize)
	}
}
