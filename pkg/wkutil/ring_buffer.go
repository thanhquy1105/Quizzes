package wkutil

import (
	"io"

	rbPool "btaskee-quiz/pkg/pool/ringbuffer"
	"btaskee-quiz/pkg/ring"
)

type RingBuffer struct {
	rb *ring.Buffer
}

func (b *RingBuffer) instance() *ring.Buffer {
	if b.rb == nil {
		b.rb = rbPool.Get()
	}

	return b.rb
}

func (b *RingBuffer) Done() {
	if b.rb != nil {
		rbPool.Put(b.rb)
		b.rb = nil
	}
}

func (b *RingBuffer) done() {
	if b.rb != nil && b.rb.IsEmpty() {
		rbPool.Put(b.rb)
		b.rb = nil
	}
}

func (b *RingBuffer) Peek(n int) (head []byte, tail []byte) {
	if b.rb == nil {
		return nil, nil
	}
	return b.rb.Peek(n)
}

func (b *RingBuffer) Discard(n int) (int, error) {
	if b.rb == nil {
		return 0, ring.ErrIsEmpty
	}

	defer b.done()
	return b.rb.Discard(n)
}

func (b *RingBuffer) Read(p []byte) (int, error) {
	if b.rb == nil {
		return 0, ring.ErrIsEmpty
	}

	defer b.done()
	return b.rb.Read(p)
}

func (b *RingBuffer) ReadByte() (byte, error) {
	if b.rb == nil {
		return 0, ring.ErrIsEmpty
	}

	defer b.done()
	return b.rb.ReadByte()
}

func (b *RingBuffer) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return b.instance().Write(p)
}

func (b *RingBuffer) WriteByte(c byte) error {
	return b.instance().WriteByte(c)
}

func (b *RingBuffer) Buffered() int {
	if b.rb == nil {
		return 0
	}
	return b.rb.Buffered()
}

func (b *RingBuffer) Len() int {
	if b.rb == nil {
		return 0
	}
	return b.rb.Len()
}

func (b *RingBuffer) Cap() int {
	if b.rb == nil {
		return 0
	}
	return b.rb.Cap()
}

func (b *RingBuffer) Available() int {
	if b.rb == nil {
		return 0
	}
	return b.rb.Available()
}

func (b *RingBuffer) WriteString(s string) (int, error) {
	if len(s) == 0 {
		return 0, nil
	}
	return b.instance().WriteString(s)
}

func (b *RingBuffer) Bytes() []byte {
	if b.rb == nil {
		return nil
	}
	return b.rb.Bytes()
}

func (b *RingBuffer) ReadFrom(r io.Reader) (int64, error) {
	return b.instance().ReadFrom(r)
}

func (b *RingBuffer) WriteTo(w io.Writer) (int64, error) {
	if b.rb == nil {
		return 0, ring.ErrIsEmpty
	}

	defer b.done()
	return b.instance().WriteTo(w)
}

func (b *RingBuffer) IsFull() bool {
	if b.rb == nil {
		return false
	}
	return b.rb.IsFull()
}

func (b *RingBuffer) IsEmpty() bool {
	if b.rb == nil {
		return true
	}
	return b.rb.IsEmpty()
}

func (b *RingBuffer) Reset() {
	if b.rb == nil {
		return
	}
	b.rb.Reset()
}
