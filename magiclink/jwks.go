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
	storage     jwkset.Storage
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
		options := jwkset.JWKOptions{
			Marshal: jwkset.JWKMarshalOptions{
				Private: true,
			},
			Metadata: jwkset.JWKMetadataOptions{
				ALG: jwkset.AlgEdDSA,
				KID: u.String(),
			},
		}
		jwk, err := jwkset.NewJWKFromKey(private, options)
		if err != nil {
			return nil, fmt.Errorf("failed to create JWK from generated EdDSA key: %w", err)
		}
		err = store.KeyWrite(ctx, jwk)
		if err != nil {
			return nil, fmt.Errorf("failed to write generated EdDSA key to storage: %w", err)
		}
	}

	initialCache, err := store.JSONPublic(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial JWK Set as JSON: %w", err)
	}

	cacheRefresh := config.CacheRefresh
	if cacheRefresh == 0 {
		cacheRefresh = DefaultJWKSCacheRefresh
	}

	jCache := &jwksCache{
		cached:      initialCache,
		storage:     store,
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
	body, err := j.storage.JSONPublic(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh the JWK Set: %w", err)
	}
	j.cached = body
	cpy := make(json.RawMessage, len(j.cached))
	copy(cpy, j.cached)
	j.lastRefresh = time.Now()

	return cpy, nil
}
