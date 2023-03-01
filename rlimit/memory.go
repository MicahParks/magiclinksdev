package rlimit

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

// Config is the configuration for the in-memory rate limiter.
type Config struct {
	Burst      uint    `json:"burst"`
	RefillRate float64 `json:"refillRate"`
}

// DefaultsAndValidate implements the jsontype.Config interface.
func (c Config) DefaultsAndValidate() (Config, error) {
	if c.Burst == 0 {
		c.Burst = 5
	}
	if c.RefillRate == 0 {
		c.RefillRate = float64(rate.Inf)
	}
	return c, nil
}

type memory struct {
	burst  uint
	m      map[string]*rate.Limiter
	mux    sync.RWMutex
	refill rate.Limit
}

// NewMemory returns a new RateLimiter that uses an in-memory implementation.
func NewMemory(conf Config) RateLimiter {
	m := &memory{
		burst:  conf.Burst,
		m:      make(map[string]*rate.Limiter),
		refill: rate.Limit(conf.RefillRate),
	}
	return m
}

// Wait implements the RateLimiter interface.
func (m *memory) Wait(ctx context.Context, key string) error {
	m.mux.RLock()
	limiter, ok := m.m[key]
	m.mux.RUnlock()
	if !ok {
		limiter = rate.NewLimiter(m.refill, int(m.burst))
		m.mux.Lock()
		m.m[key] = limiter
		m.mux.Unlock()
	}
	return limiter.Wait(ctx)
}
