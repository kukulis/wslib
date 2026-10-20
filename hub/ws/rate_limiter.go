package ws

import "time"

// RateLimiter decides whether one more event is allowed right now. Implementations
// hold their own policy and state; callers just call Allow() on every attempt.
type RateLimiter interface {
	Allow() bool
}

// SlidingWindowLimiter allows at most `max` calls to Allow() within any trailing
// `window` duration. Not safe for concurrent use - intended for a single owner,
// e.g. one Client's readPump goroutine.
type SlidingWindowLimiter struct {
	window time.Duration
	max    int
	times  []time.Time
}

var _ RateLimiter = (*SlidingWindowLimiter)(nil)

func NewSlidingWindowLimiter(window time.Duration, max int) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		window: window,
		max:    max,
	}
}

func (l *SlidingWindowLimiter) Allow() bool {
	now := time.Now()
	cutoff := now.Add(-l.window)

	i := 0
	for i < len(l.times) && l.times[i].Before(cutoff) {
		i++
	}
	l.times = l.times[i:]

	if len(l.times) >= l.max {
		return false
	}

	l.times = append(l.times, now)
	return true
}
