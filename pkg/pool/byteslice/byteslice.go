package byteslice

import (
	"math"
	"math/bits"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

var builtinPool Pool

type Pool struct {
	pools [32]sync.Pool
}

func Get(size int) []byte {
	return builtinPool.Get(size)
}

func Put(buf []byte) {
	builtinPool.Put(buf)
}

func (p *Pool) Get(size int) (buf []byte) {
	if size <= 0 {
		return nil
	}
	if size > math.MaxInt32 {
		return make([]byte, size)
	}
	idx := index(uint32(size))
	ptr, _ := p.pools[idx].Get().(unsafe.Pointer)
	if ptr == nil {
		return make([]byte, 1<<idx)[:size]
	}
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Data = uintptr(ptr)
	sh.Len = size
	sh.Cap = 1 << idx
	runtime.KeepAlive(ptr)
	return
}

func (p *Pool) Put(buf []byte) {
	size := cap(buf)
	if size == 0 || size > math.MaxInt32 {
		return
	}
	idx := index(uint32(size))
	if size != 1<<idx {
		idx--
	}

	p.pools[idx].Put(unsafe.Pointer(&buf[:1][0]))
}

func index(n uint32) uint32 {
	return uint32(bits.Len32(n - 1))
}
