package util

import "time"

// Clock provides the current time. Production code uses RealClock; tests
// substitute their own implementation to control time deterministically
// instead of sleeping or polling around real wall-clock time.
type Clock interface {
	Now() time.Time
}

// RealClock is a Clock backed by the actual system clock.
type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

// FixedClock is a Clock that always returns Time, until Set or Advance changes it.
// Tests use it (e.g. assigned to MainState.Clock) to control "now" deterministically,
// such as pushing a timestamp past a timeout without sleeping.
type FixedClock struct {
	Time time.Time
}

func NewFixedClock(t time.Time) *FixedClock {
	return &FixedClock{Time: t}
}

func (c *FixedClock) Now() time.Time {
	return c.Time
}

// Set changes the time FixedClock reports.
func (c *FixedClock) Set(t time.Time) {
	c.Time = t
}

// Advance moves the time FixedClock reports forward by d (or backward, if d is negative).
func (c *FixedClock) Advance(d time.Duration) {
	c.Time = c.Time.Add(d)
}
