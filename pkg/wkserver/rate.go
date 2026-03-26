package wkserver

import (
	"math"
	"sync/atomic"

	"btaskee-quiz/pkg/wklog"
	"go.uber.org/zap"
)

type RateLimiter struct {
	size    uint64
	maxSize uint64
	wklog.Log
}

func NewRateLimiter(max uint64) *RateLimiter {
	return &RateLimiter{
		maxSize: max,
		Log:     wklog.NewWKLog("rateLimiter"),
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
		r.Info("rate limited", zap.Uint64("v", v), zap.Uint64("maxSize", r.maxSize))
		return true
	}
	return false
}
