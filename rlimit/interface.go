package rlimit

import "context"

// RateLimiter is an interface for rate limiting.
type RateLimiter interface {
	Wait(ctx context.Context, serviceAccountUUID string) error
}
