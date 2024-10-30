package mldtest

import (
	"context"
)

// NopLimiter is a rate limiter that does nothing.
type NopLimiter struct{}

// Wait implements the rlimit.RateLimiter interface.
func (n NopLimiter) Wait(_ context.Context, _ string) error {
	return nil
}
