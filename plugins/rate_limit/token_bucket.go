package rate_limit

import (
	"sync/atomic"
	"time"
	"unsafe"
)

type Clock interface {
	Now() time.Time
}

type simpleClock struct {
}

func (clock *simpleClock) Now() time.Time {
	return time.Now()
}

type Limiter interface {
	Take() bool
	Undo()
	UpdateRate(rate int, per time.Duration)
	WithClock(clock Clock) Limiter
}

type state struct {
	allowance uint64
	lastCheck uint64
}

type tokenBucketLimiter struct {
	clock Clock
	rate  uint64
	max   uint64
	unit  uint64
	state unsafe.Pointer
}

func (limiter *tokenBucketLimiter) WithClock(clock Clock) Limiter {
	limiter.clock = clock
	return limiter
}

func unixNano(clock Clock) uint64 {
	return uint64(clock.Now().UnixNano())
}

func NewTokenBucketLimiter(rate int, per time.Duration) Limiter {
	nano := uint64(per)
	if nano < 1 {
		nano = uint64(time.Second)
	}
	if rate < 1 {
		rate = 1
	}
	clock := &simpleClock{}
	unit := nano
	allowance := uint64(rate) * unit
	lastCheck := unixNano(clock)
	initState := state{
		allowance: allowance,
		lastCheck: lastCheck,
	}

	ul := &tokenBucketLimiter{
		clock: clock,
		rate:  uint64(rate),
		max:   uint64(rate) * unit,
		unit:  unit,
	}
	atomic.StorePointer(&ul.state, unsafe.Pointer(&initState))
	return ul
}

func (limiter *tokenBucketLimiter) UpdateRate(rate int, per time.Duration) {
	nano := uint64(per)
	if nano < 1 {
		nano = uint64(time.Second)
	}
	if rate < 1 {
		rate = 1
	}
	unit := nano
	max := uint64(rate) * unit
	atomic.StoreUint64(&limiter.rate, uint64(rate))
	atomic.StoreUint64(&limiter.max, max)
}

func (limiter *tokenBucketLimiter) Take() bool {
	var newState state
	changeStateSuccess := false
	for !changeStateSuccess {
		now := unixNano(limiter.clock)
		previousStatePointer := atomic.LoadPointer(&limiter.state)
		oldState := (*state)(previousStatePointer)

		newState = state{}
		newState.allowance = oldState.allowance
		newState.lastCheck = now

		// time passed after last check
		passed := now - oldState.lastCheck

		rate := atomic.LoadUint64(&limiter.rate)
		// after passed time, there are more allowance
		current := newState.allowance + passed*rate
		newState.allowance = current

		if max := atomic.LoadUint64(&limiter.max); current > max {
			current = max
			newState.allowance = max
		}

		if current < limiter.unit {
			return false
		}

		// available to take
		newState.allowance -= limiter.unit
		changeStateSuccess = atomic.CompareAndSwapPointer(&limiter.state, previousStatePointer, unsafe.Pointer(&newState))
	}
	return true
}

func (limiter *tokenBucketLimiter) Undo() {
	var newState state
	changeStateSuccess := false
	for !changeStateSuccess {
		previousStatePointer := atomic.LoadPointer(&limiter.state)
		oldState := (*state)(previousStatePointer)
		newState = state{}
		newState.lastCheck = oldState.lastCheck
		newState.allowance = oldState.allowance + limiter.unit

		if max := atomic.LoadUint64(&limiter.max); newState.allowance > max {
			newState.allowance = max
		}
		changeStateSuccess = atomic.CompareAndSwapPointer(&limiter.state, previousStatePointer, unsafe.Pointer(&newState))
	}
}
