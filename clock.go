package main

import "time"

// Clock is our timekeeping device, this is used by cachedFetcher to abstract source of time.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

type testClock struct {
	NowValue time.Time
}

func (clock testClock) Now() time.Time { return clock.NowValue }
