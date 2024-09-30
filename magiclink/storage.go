package magiclink

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Storage represents the underlying storage for the MagicLink service.
type Storage[CustomReadResponse any] interface {
	// CreateLink creates a secret for the given parameters and stores the pair. The secret is returned to the caller.
	CreateLink(ctx context.Context, args CreateArgs) (secret string, err error)
	// ReadLink finds the creation parameters for the given secret. ErrLinkNotFound is returned if the secret is not
	// found or was deleted/expired. This will automatically expire the link.
	ReadLink(ctx context.Context, secret string) (ReadResponse[CustomReadResponse], error)
}

var _ Storage[any] = &memoryMagicLink[any]{}

type memoryMagicLink[CustomReadResponse any] struct {
	links map[string]ReadResponse[CustomReadResponse]
	mux   sync.Mutex
}

// NewMemoryStorage creates an in-memory implementation of the MagicLink Storage.
func NewMemoryStorage[CustomReadResponse any]() Storage[CustomReadResponse] {
	return &memoryMagicLink[CustomReadResponse]{
		links: map[string]ReadResponse[CustomReadResponse]{},
	}
}
func (m *memoryMagicLink[CustomReadResponse]) CreateLink(_ context.Context, args CreateArgs) (secret string, err error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	u, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID as secret: %w", err)
	}
	secret = u.String()
	var custom CustomReadResponse
	response := ReadResponse[CustomReadResponse]{
		Custom:     custom,
		CreateArgs: args,
	}
	m.links[secret] = response
	return secret, nil
}
func (m *memoryMagicLink[CustomReadResponse]) ReadLink(_ context.Context, secret string) (ReadResponse[CustomReadResponse], error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	args, ok := m.links[secret]
	if !ok {
		return args, ErrLinkNotFound
	}
	delete(m.links, secret)
	return args, nil
}
