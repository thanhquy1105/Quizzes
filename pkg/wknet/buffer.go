package wknet

type Buffer interface {
	IsEmpty() bool

	Write(data []byte) (int, error)

	Read(data []byte) (int, error)

	BoundBufferSize() int

	Peek(n int) (head []byte, tail []byte)
	PeekBytes(p []byte) int

	Discard(n int) (int, error)

	Release() error
}

type InboundBuffer interface {
	Buffer
}

type OutboundBuffer interface {
	Buffer
}

type DefualtBuffer struct {
	ringBuffer *RingBuffer
}

func NewDefaultBuffer() *DefualtBuffer {
	return &DefualtBuffer{
		ringBuffer: &RingBuffer{},
	}
}

func (d *DefualtBuffer) IsEmpty() bool {
	return d.ringBuffer.IsEmpty()
}

func (d *DefualtBuffer) Write(data []byte) (int, error) {
	return d.ringBuffer.Write(data)
}
func (d *DefualtBuffer) Read(data []byte) (int, error) {
	return d.ringBuffer.Read(data)
}
func (d *DefualtBuffer) BoundBufferSize() int {
	return d.ringBuffer.Buffered()
}

func (d *DefualtBuffer) Peek(n int) (head []byte, tail []byte) {
	return d.ringBuffer.Peek(n)
}

func (d *DefualtBuffer) PeekBytes(p []byte) int {
	head, tail := d.ringBuffer.Peek(-1)
	if len(head) > 0 && len(tail) > 0 {
		return copy(p, append(head, tail...))
	}

	if len(head) > 0 {
		return copy(p, head)
	}
	if len(tail) > 0 {
		return copy(p, tail)
	}
	return 0
}

func (d *DefualtBuffer) Discard(n int) (int, error) {
	return d.ringBuffer.Discard(n)
}

func (d *DefualtBuffer) Release() error {
	d.ringBuffer.Done()
	d.ringBuffer.Reset()
	return nil
}
