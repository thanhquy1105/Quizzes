//go:build linux || freebsd || dragonfly || darwin
// +build linux freebsd dragonfly darwin

package ring

import (
	"btaskee-quiz/pkg/wknet/io"
	"golang.org/x/sys/unix"
)

func (rb *Buffer) CopyFromSocket(fd int) (n int, err error) {
	if rb.r == rb.w {
		if !rb.isEmpty {
			rb.grow(rb.size + rb.size/2)
			n, err = unix.Read(fd, rb.buf[rb.w:])
			if n > 0 {
				rb.w = (rb.w + n) % rb.size
			}
			return
		}
		rb.r, rb.w = 0, 0
		n, err = unix.Read(fd, rb.buf)
		if n > 0 {
			rb.w = (rb.w + n) % rb.size
			rb.isEmpty = false
		}
		return
	}
	if rb.w < rb.r {
		n, err = unix.Read(fd, rb.buf[rb.w:rb.r])
		if n > 0 {
			rb.w = (rb.w + n) % rb.size
		}
		return
	}
	rb.bs[0] = rb.buf[rb.w:]
	rb.bs[1] = rb.buf[:rb.r]
	n, err = io.Readv(fd, rb.bs)
	if n > 0 {
		rb.w = (rb.w + n) % rb.size
	}
	return
}

func (rb *Buffer) Rewind() (n int) {
	if rb.IsEmpty() {
		rb.Reset()
		return
	}
	if rb.w == 0 {
		if rb.r < rb.size-rb.r {
			rb.grow(rb.size + rb.size - rb.r)
			return rb.size - rb.r
		}
		n = copy(rb.buf, rb.buf[rb.r:])
		rb.r = 0
		rb.w = n
	} else if rb.size-rb.w < DefaultBufferSize {
		if rb.r < rb.w-rb.r {
			rb.grow(rb.size + rb.w - rb.r)
			return rb.w - rb.r
		}
		n = copy(rb.buf, rb.buf[rb.r:rb.w])
		rb.r = 0
		rb.w = n
	}
	return
}
