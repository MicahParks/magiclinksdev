package mldtest

import (
	"context"
)

// ErrorLimiter is a rate limiter that always returns an error.
type ErrorLimiter struct{}

// Wait implements the rlimit.RateLimiter interface.
func (e ErrorLimiter) Wait(_ context.Context, _ string) error {
	return ErrMLDTest
}

// NopLimiter is a rate limiter that does nothing.
type NopLimiter struct{}

// Wait implements the rlimit.RateLimiter interface.
func (n NopLimiter) Wait(_ context.Context, _ string) error {
	return nil
}
