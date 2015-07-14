package clock

import (
	"time"
)

// A replaceable clock interface. Alternative implementations can run faster than real time.
type Clock interface {
	Now() time.Time
	NewTimer(d time.Duration) *time.Timer
}

var clock Clock = NewClock()

type realClock struct {
}

func GetClock() Clock {
	return clock
}

func SetClock(c Clock) {
	clock = c
}

func NewClock() Clock {
	return &realClock{}
}

func (c *realClock) Now() time.Time {
	return time.Now()
}

func (c *realClock) NewTimer(d time.Duration) *time.Timer {
	return time.NewTimer(d)
}

type simulatedClock struct {
	epoch     time.Time
	epochDiff time.Duration
	speedup   int
}

const DEFAULT_EPOCH_DIFF = 86400 * 365 * 10

// Answer a clock in which real time after the epoch time is sped up by the specified amount.
func NewSimulatedClock(epoch time.Time, speedup int, epochDiff time.Duration) Clock {
	return &simulatedClock{
		epoch:     epoch,
		speedup:   speedup,
		epochDiff: epochDiff,
	}
}

func (c *simulatedClock) Now() time.Time {
	return c.simulatedTime(time.Now())
}

func (c *simulatedClock) simulatedTime(realTime time.Time) time.Time {
	return c.epoch.Add(realTime.Sub(c.epoch) * time.Duration(c.speedup)).Add(-c.epochDiff)
}

func (c *simulatedClock) NewTimer(d time.Duration) *time.Timer {
	t := time.NewTimer(d / time.Duration(c.speedup))
	ch := make(chan time.Time)
	pseudo := &time.Timer{
		C: ch,
	}
	go func() {
		t := <-t.C
		ch <- c.simulatedTime(t)
	}()
	return pseudo
}
