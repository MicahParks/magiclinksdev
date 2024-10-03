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
	// CreateLink creates a secret for the given parameters and stores the pair. The secret is returned to the caller.
	CreateLink(ctx context.Context, args CreateArgs) (secret string, err error)
	// ReadLink finds the creation parameters for the given secret. ErrLinkNotFound is returned if the secret is not
	// found or was deleted/expired. This will automatically expire the link.
	ReadLink(ctx context.Context, secret string) (ReadResponse, error) // TODO Use visited and expiration.
}

var _ Storage = &memoryMagicLink{}

type memoryMagicLink struct {
	links map[string]ReadResponse
	mux   sync.Mutex
}

// NewMemoryStorage creates an in-memory implementation of the MagicLink Storage.
func NewMemoryStorage() Storage {
	return &memoryMagicLink{
		links: map[string]ReadResponse{},
	}
}
func (m *memoryMagicLink) CreateLink(_ context.Context, args CreateArgs) (secret string, err error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	u, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID as secret: %w", err)
	}
	secret = u.String()
	response := ReadResponse{
		CreateArgs: args,
	}
	m.links[secret] = response
	return secret, nil
}
func (m *memoryMagicLink) ReadLink(_ context.Context, secret string) (ReadResponse, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	now := time.Now()
	readResp, ok := m.links[secret]
	if !ok || readResp.Visited != nil || readResp.CreateArgs.Expires.Before(now) {
		return readResp, ErrLinkNotFound
	}
	readResp.Visited = mld.Ptr(now)
	m.links[secret] = readResp
	return readResp, nil
}
