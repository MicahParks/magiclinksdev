package magiclink

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/google/uuid"
)

// DefaultJWKSCacheRefresh is the default time to wait before refreshing the JWKS cache.
const DefaultJWKSCacheRefresh = 5 * time.Minute

type jwksCache struct {
	cached      json.RawMessage
	jwks        jwkset.Storage
	lastRefresh time.Time
	refresh     time.Duration
	mux         sync.RWMutex
}

func newJWKSCache(ctx context.Context, config JWKSArgs) (*jwksCache, error) {
	store := config.Store
	if store == nil {
		store = jwkset.NewMemoryStorage()
		_, private, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ed25519 key for empty JWKS: %w", err)
		}
		u, err := uuid.NewRandom()
		if err != nil {
			return nil, fmt.Errorf("failed to generate UUID for generated RSA key: %w", err)
		}
		meta := jwkset.NewKey(private, u.String())
		err = store.WriteKey(ctx, meta)
		if err != nil {
			return nil, fmt.Errorf("failed to store generated RSA key: %w", err)
		}
	}

	jwkSet := jwkset.JWKSet{
		Store: store,
	}

	initialCache, err := jwkSet.JSONPublic(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial JWK Set as JSON: %w", err)
	}

	cacheRefresh := config.CacheRefresh
	if cacheRefresh == 0 {
		cacheRefresh = DefaultJWKSCacheRefresh
	}

	jCache := &jwksCache{
		cached:      initialCache,
		jwks:        jwkSet,
		lastRefresh: time.Now(),
		refresh:     cacheRefresh,
	}

	return jCache, nil
}

func (j *jwksCache) get(ctx context.Context) (json.RawMessage, error) {
	j.mux.RLock()
	since := time.Since(j.lastRefresh)
	if since <= j.refresh {
		cpy := make(json.RawMessage, len(j.cached))
		copy(cpy, j.cached)
		j.mux.RUnlock()
		return cpy, nil
	}
	j.mux.RUnlock()

	j.mux.Lock()
	defer j.mux.Unlock()
	body, err := j.jwks.JSONPublic(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh the JWK Set: %w", err)
	}
	j.cached = body
	cpy := make(json.RawMessage, len(j.cached))
	copy(cpy, j.cached)
	j.lastRefresh = time.Now()

	return cpy, nil
}
