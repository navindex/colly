package cache

import (
	"time"
)

// ------------------------------------------------------------------------

type expByHeader struct{}

type expByDuration struct {
	duration time.Duration
}

type expByDate struct {
	expiry time.Time
}

// ------------------------------------------------------------------------

func NewExpirationByHeader() *expByHeader {
	return &expByHeader{}
}

// ------------------------------------------------------------------------

func NewExpirationByDuration(duration time.Duration) *expByDuration {
	if duration == 0 {
		return nil
	}

	return &expByDuration{
		duration: duration,
	}
}

// ------------------------------------------------------------------------

func NewExpirationByDate(expiry time.Time) *expByDate {
	if expiry.IsZero() || expiry.Before(time.Now()) {
		return nil
	}

	return &expByDate{
		expiry: expiry,
	}
}
