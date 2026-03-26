package wkutil

import (
	"math"
	"sync/atomic"
)

const (
	gcTick uint64 = 3

	ChangeTickThreashold uint64 = 10
)

type followerState struct {
	tick         uint64
	inMemLogSize uint64
}

type RateLimiter struct {
	size    uint64
	maxSize uint64
}

func NewRateLimiter(max uint64) *RateLimiter {
	return &RateLimiter{
		maxSize: max,
	}
}

func (r *RateLimiter) Enabled() bool {
	return r.maxSize > 0 && r.maxSize != math.MaxUint64
}

func (r *RateLimiter) Increase(sz uint64) {
	atomic.AddUint64(&r.size, sz)
}

func (r *RateLimiter) Decrease(sz uint64) {
	atomic.AddUint64(&r.size, ^(sz - 1))
}

func (r *RateLimiter) Set(sz uint64) {
	atomic.StoreUint64(&r.size, sz)
}

func (r *RateLimiter) Get() uint64 {
	return atomic.LoadUint64(&r.size)
}

func (r *RateLimiter) RateLimited() bool {
	if !r.Enabled() {
		return false
	}
	v := r.Get()
	if v > r.maxSize {
		return true
	}
	return false
}

type InMemRateLimiter struct {
	followerSizes map[uint64]followerState
	rl            RateLimiter
	tick          uint64
	tickLimited   uint64
	limited       bool
}

func NewInMemRateLimiter(maxSize uint64) *InMemRateLimiter {
	return &InMemRateLimiter{

		tick:          1,
		rl:            RateLimiter{maxSize: maxSize},
		followerSizes: make(map[uint64]followerState),
	}
}

func (r *InMemRateLimiter) Enabled() bool {
	return r.rl.Enabled()
}

func (r *InMemRateLimiter) Tick() {
	r.tick++
}

func (r *InMemRateLimiter) GetTick() uint64 {
	return r.tick
}

func (r *InMemRateLimiter) Increase(sz uint64) {
	r.rl.Increase(sz)
}

func (r *InMemRateLimiter) Decrease(sz uint64) {
	r.rl.Decrease(sz)
}

func (r *InMemRateLimiter) Set(sz uint64) {
	r.rl.Set(sz)
}

func (r *InMemRateLimiter) Get() uint64 {
	return r.rl.Get()
}

func (r *InMemRateLimiter) Reset() {
	r.followerSizes = make(map[uint64]followerState)
}

func (r *InMemRateLimiter) SetFollowerState(replicaID uint64, sz uint64) {
	r.followerSizes[replicaID] = followerState{
		tick:         r.tick,
		inMemLogSize: sz,
	}
}

func (r *InMemRateLimiter) RateLimited() bool {
	limited := r.limitedByInMemSize()
	if limited != r.limited {
		if r.tickLimited == 0 || r.tick-r.tickLimited > ChangeTickThreashold {
			r.limited = limited
			r.tickLimited = r.tick
		}
	}
	return r.limited
}

func (r *InMemRateLimiter) limitedByInMemSize() bool {
	if !r.Enabled() {
		return false
	}
	maxInMemSize := uint64(0)
	gc := false
	for _, v := range r.followerSizes {
		if r.tick-v.tick > gcTick {
			gc = true
			continue
		}
		if v.inMemLogSize > maxInMemSize {
			maxInMemSize = v.inMemLogSize
		}
	}
	sz := r.Get()
	if sz > maxInMemSize {
		maxInMemSize = sz
	}
	if gc {
		r.gc()
	}
	if !r.limited {
		return maxInMemSize > r.rl.maxSize
	}
	return maxInMemSize >= (r.rl.maxSize * 7 / 10)
}

func (r *InMemRateLimiter) gc() {
	followerStates := make(map[uint64]followerState)
	for nid, v := range r.followerSizes {
		if r.tick-v.tick > gcTick {
			continue
		}
		followerStates[nid] = v
	}
	r.followerSizes = followerStates
}
