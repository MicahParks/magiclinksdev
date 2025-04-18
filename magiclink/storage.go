package magiclink

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	mld "github.com/MicahParks/magiclinksdev"
)

// Storage represents the underlying storage for the MagicLink service.
type Storage interface {
	// MagicLinkCreate creates a secret for the given parameters and stores the pair. The secret is returned to the
	// caller.
	MagicLinkCreate(ctx context.Context, params CreateParams) (secret string, err error)
	// MagicLinkRead finds the creation parameters for the given secret. ErrLinkNotFound is returned if the secret is
	// not found or was deleted/expired. This will automatically expire the link.
	MagicLinkRead(ctx context.Context, secret string) (ReadResult, error)
}

var _ Storage = &memoryMagicLink{}

type memoryMagicLink struct {
	links map[string]ReadResult
	mux   sync.Mutex
}

// NewMemoryStorage creates an in-memory implementation of the MagicLink Storage.
func NewMemoryStorage() Storage {
	return &memoryMagicLink{
		links: map[string]ReadResult{},
	}
}
func (m *memoryMagicLink) MagicLinkCreate(_ context.Context, args CreateParams) (secret string, err error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	u, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID as secret: %w", err)
	}
	secret = u.String()
	response := ReadResult{
		CreateParams: args,
	}
	m.links[secret] = response
	return secret, nil
}
func (m *memoryMagicLink) MagicLinkRead(_ context.Context, secret string) (ReadResult, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	now := time.Now()
	readResp, ok := m.links[secret]
	if !ok || readResp.Visited != nil || readResp.CreateParams.Expires.Before(now) {
		return readResp, ErrLinkNotFound
	}
	readResp.Visited = mld.Ptr(now)
	m.links[secret] = readResp
	return readResp, nil
}
