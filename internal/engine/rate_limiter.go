package engine

import (
	"sync"
	"time"

	"github.com/sol-rpc/sol/internal/discord"
)

// RateLimiter enforces Discord Rich Presence update intervals (e.g. 15s)
// and debounces rapid state changes so the latest state is always dispatched.
type RateLimiter struct {
	mu           sync.Mutex
	interval     time.Duration
	lastSent     time.Time
	pending      *discord.Activity
	hasPending   bool
	timer        *time.Timer
	dispatchFunc func(activity *discord.Activity) error
}

// NewRateLimiter creates a new rate limiter with the specified minimum interval.
func NewRateLimiter(interval time.Duration, dispatch func(activity *discord.Activity) error) *RateLimiter {
	return &RateLimiter{
		interval:     interval,
		dispatchFunc: dispatch,
	}
}

// Submit requests a Discord Rich Presence update. If the interval has elapsed,
// it dispatches immediately; otherwise, it queues the latest activity for deferred dispatch.
func (rl *RateLimiter) Submit(activity *discord.Activity) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastSent)

	if elapsed >= rl.interval {
		// Interval has passed, dispatch immediately
		rl.lastSent = now
		rl.hasPending = false
		if rl.timer != nil {
			rl.timer.Stop()
		}
		_ = rl.dispatchFunc(activity)
		return
	}

	// Queue pending update
	rl.pending = activity
	rl.hasPending = true

	waitDuration := rl.interval - elapsed
	if rl.timer != nil {
		rl.timer.Stop()
	}

	rl.timer = time.AfterFunc(waitDuration, func() {
		rl.flush()
	})
}

// flush executes the pending activity when the timer expires.
func (rl *RateLimiter) flush() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if !rl.hasPending || rl.pending == nil {
		return
	}

	rl.lastSent = time.Now()
	activityToDispatch := rl.pending
	rl.hasPending = false
	rl.pending = nil

	_ = rl.dispatchFunc(activityToDispatch)
}

// Stop cancels any pending scheduled dispatches.
func (rl *RateLimiter) Stop() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.timer != nil {
		rl.timer.Stop()
	}
	rl.hasPending = false
	rl.pending = nil
}
