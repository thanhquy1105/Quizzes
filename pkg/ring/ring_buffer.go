package ring

import (
	"errors"
	"io"
	"reflect"
	"unsafe"

	bsPool "github.com/panjf2000/gnet/v2/pkg/pool/byteslice"
)

const (
	MinRead = 512

	DefaultBufferSize   = 1024
	bufferGrowThreshold = 4 * 1024
)

var ErrIsEmpty = errors.New("ring-buffer is empty")

type Buffer struct {
	bs      [][]byte
	buf     []byte
	size    int
	r       int
	w       int
	isEmpty bool
}

func New(size int) *Buffer {
	if size == 0 {
		return &Buffer{bs: make([][]byte, 2), isEmpty: true}
	}
	size = CeilToPowerOfTwo(size)
	return &Buffer{
		bs:      make([][]byte, 2),
		buf:     make([]byte, size),
		size:    size,
		isEmpty: true,
	}
}

const (
	bitSize       = 32 << (^uint(0) >> 63)
	maxintHeadBit = 1 << (bitSize - 2)
)

func CeilToPowerOfTwo(n int) int {
	if n&maxintHeadBit != 0 && n > maxintHeadBit {
		panic("argument is too large")
	}

	if n <= 2 {
		return 2
	}

	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++

	return n
}

func (rb *Buffer) Peek(n int) (head []byte, tail []byte) {
	if rb.isEmpty {
		return
	}

	if n <= 0 {
		return rb.peekAll()
	}

	if rb.w > rb.r {
		m := rb.w - rb.r
		if m > n {
			m = n
		}
		head = rb.buf[rb.r : rb.r+m]
		return
	}

	m := rb.size - rb.r + rb.w
	if m > n {
		m = n
	}

	if rb.r+m <= rb.size {
		head = rb.buf[rb.r : rb.r+m]
	} else {
		c1 := rb.size - rb.r
		head = rb.buf[rb.r:]
		c2 := m - c1
		tail = rb.buf[:c2]
	}

	return
}

func (rb *Buffer) peekAll() (head []byte, tail []byte) {
	if rb.isEmpty {
		return
	}

	if rb.w > rb.r {
		head = rb.buf[rb.r:rb.w]
		return
	}

	head = rb.buf[rb.r:]
	if rb.w != 0 {
		tail = rb.buf[:rb.w]
	}

	return
}

func (rb *Buffer) Discard(n int) (discarded int, err error) {
	if n <= 0 {
		return 0, nil
	}

	discarded = rb.Buffered()
	if n < discarded {
		rb.r = (rb.r + n) % rb.size
		return n, nil
	}
	rb.Reset()
	return
}

func (rb *Buffer) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	if rb.isEmpty {
		return 0, ErrIsEmpty
	}

	if rb.w > rb.r {
		n = rb.w - rb.r
		if n > len(p) {
			n = len(p)
		}
		copy(p, rb.buf[rb.r:rb.r+n])
		rb.r += n
		if rb.r == rb.w {
			rb.Reset()
		}
		return
	}

	n = rb.size - rb.r + rb.w
	if n > len(p) {
		n = len(p)
	}

	if rb.r+n <= rb.size {
		copy(p, rb.buf[rb.r:rb.r+n])
	} else {
		c1 := rb.size - rb.r
		copy(p, rb.buf[rb.r:])
		c2 := n - c1
		copy(p[c1:], rb.buf[:c2])
	}
	rb.r = (rb.r + n) % rb.size
	if rb.r == rb.w {
		rb.Reset()
	}

	return
}

func (rb *Buffer) ReadByte() (b byte, err error) {
	if rb.isEmpty {
		return 0, ErrIsEmpty
	}
	b = rb.buf[rb.r]
	rb.r++
	if rb.r == rb.size {
		rb.r = 0
	}
	if rb.r == rb.w {
		rb.Reset()
	}

	return
}

func (rb *Buffer) Write(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		return
	}

	free := rb.Available()
	if n > free {
		rb.grow(rb.size + n - free)
	}

	if rb.w >= rb.r {
		c1 := rb.size - rb.w
		if c1 >= n {
			copy(rb.buf[rb.w:], p)
			rb.w += n
		} else {
			copy(rb.buf[rb.w:], p[:c1])
			c2 := n - c1
			copy(rb.buf, p[c1:])
			rb.w = c2
		}
	} else {
		copy(rb.buf[rb.w:], p)
		rb.w += n
	}

	if rb.w == rb.size {
		rb.w = 0
	}

	rb.isEmpty = false

	return
}

func (rb *Buffer) WriteByte(c byte) error {
	if rb.Available() < 1 {
		rb.grow(1)
	}
	rb.buf[rb.w] = c
	rb.w++

	if rb.w == rb.size {
		rb.w = 0
	}
	rb.isEmpty = false

	return nil
}

func (rb *Buffer) Buffered() int {
	if rb.r == rb.w {
		if rb.isEmpty {
			return 0
		}
		return rb.size
	}

	if rb.w > rb.r {
		return rb.w - rb.r
	}

	return rb.size - rb.r + rb.w
}

func (rb *Buffer) Len() int {
	return len(rb.buf)
}

func (rb *Buffer) Cap() int {
	return rb.size
}

func (rb *Buffer) Available() int {
	if rb.r == rb.w {
		if rb.isEmpty {
			return rb.size
		}
		return 0
	}

	if rb.w < rb.r {
		return rb.r - rb.w
	}

	return rb.size - rb.w + rb.r
}

func (rb *Buffer) WriteString(s string) (int, error) {
	return rb.Write(StringToBytes(s))
}

func (rb *Buffer) Bytes() []byte {
	if rb.isEmpty {
		return nil
	} else if rb.w == rb.r {
		var bb []byte
		bb = append(bb, rb.buf[rb.r:]...)
		bb = append(bb, rb.buf[:rb.w]...)
		return bb
	}

	var bb []byte
	if rb.w > rb.r {
		bb = append(bb, rb.buf[rb.r:rb.w]...)
		return bb
	}

	bb = append(bb, rb.buf[rb.r:]...)

	if rb.w != 0 {
		bb = append(bb, rb.buf[:rb.w]...)
	}

	return bb
}

func (rb *Buffer) ReadFrom(r io.Reader) (n int64, err error) {
	var m int
	for {
		if rb.Available() < MinRead {
			rb.grow(rb.Buffered() + MinRead)
		}

		if rb.w >= rb.r {
			m, err = r.Read(rb.buf[rb.w:])
			if m < 0 {
				panic("RingBuffer.ReadFrom: reader returned negative count from Read")
			}
			rb.isEmpty = false
			rb.w = (rb.w + m) % rb.size
			n += int64(m)
			if err == io.EOF {
				return n, nil
			}
			if err != nil {
				return
			}
			m, err = r.Read(rb.buf[:rb.r])
			if m < 0 {
				panic("RingBuffer.ReadFrom: reader returned negative count from Read")
			}
			rb.w = (rb.w + m) % rb.size
			n += int64(m)
			if err == io.EOF {
				return n, nil
			}
			if err != nil {
				return
			}
		} else {
			m, err = r.Read(rb.buf[rb.w:rb.r])
			if m < 0 {
				panic("RingBuffer.ReadFrom: reader returned negative count from Read")
			}
			rb.isEmpty = false
			rb.w = (rb.w + m) % rb.size
			n += int64(m)
			if err == io.EOF {
				return n, nil
			}
			if err != nil {
				return
			}
		}
	}
}

func (rb *Buffer) WriteTo(w io.Writer) (int64, error) {
	if rb.isEmpty {
		return 0, ErrIsEmpty
	}

	if rb.w > rb.r {
		n := rb.w - rb.r
		m, err := w.Write(rb.buf[rb.r : rb.r+n])
		if m > n {
			panic("RingBuffer.WriteTo: invalid Write count")
		}
		rb.r += m
		if rb.r == rb.w {
			rb.Reset()
		}
		if err != nil {
			return int64(m), err
		}
		if !rb.isEmpty {
			return int64(m), io.ErrShortWrite
		}
		return int64(m), nil
	}

	n := rb.size - rb.r + rb.w
	if rb.r+n <= rb.size {
		m, err := w.Write(rb.buf[rb.r : rb.r+n])
		if m > n {
			panic("RingBuffer.WriteTo: invalid Write count")
		}
		rb.r = (rb.r + m) % rb.size
		if rb.r == rb.w {
			rb.Reset()
		}
		if err != nil {
			return int64(m), err
		}
		if !rb.isEmpty {
			return int64(m), io.ErrShortWrite
		}
		return int64(m), nil
	}

	var cum int64
	c1 := rb.size - rb.r
	m, err := w.Write(rb.buf[rb.r:])
	if m > c1 {
		panic("RingBuffer.WriteTo: invalid Write count")
	}
	rb.r = (rb.r + m) % rb.size
	if err != nil {
		return int64(m), err
	}
	if m < c1 {
		return int64(m), io.ErrShortWrite
	}
	cum += int64(m)
	c2 := n - c1
	m, err = w.Write(rb.buf[:c2])
	if m > c2 {
		panic("RingBuffer.WriteTo: invalid Write count")
	}
	rb.r = m
	cum += int64(m)
	if rb.r == rb.w {
		rb.Reset()
	}
	if err != nil {
		return cum, err
	}
	if !rb.isEmpty {
		return cum, io.ErrShortWrite
	}
	return cum, nil
}

func (rb *Buffer) IsFull() bool {
	return rb.r == rb.w && !rb.isEmpty
}

func (rb *Buffer) IsEmpty() bool {
	return rb.isEmpty
}

func (rb *Buffer) Reset() {
	rb.isEmpty = true
	rb.r, rb.w = 0, 0
}

func (rb *Buffer) grow(newCap int) {
	if n := rb.size; n == 0 {
		if newCap <= DefaultBufferSize {
			newCap = DefaultBufferSize
		} else {
			newCap = CeilToPowerOfTwo(newCap)
		}
	} else {
		doubleCap := n + n
		if newCap <= doubleCap {
			if n < bufferGrowThreshold {
				newCap = doubleCap
			} else {

				for 0 < n && n < newCap {
					n += n / 4
				}

				if n > 0 {
					newCap = n
				}
			}
		}
	}
	newBuf := bsPool.Get(newCap)
	oldLen := rb.Buffered()
	_, _ = rb.Read(newBuf)
	bsPool.Put(rb.buf)
	rb.buf = newBuf
	rb.r = 0
	rb.w = oldLen
	rb.size = newCap
	if rb.w > 0 {
		rb.isEmpty = false
	}
}

func StringToBytes(s string) (b []byte) {

	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))

	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))

	bh.Data, bh.Len, bh.Cap = sh.Data, sh.Len, sh.Len
	return b
}
