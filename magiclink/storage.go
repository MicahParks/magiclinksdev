package magiclink

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Storage represents the underlying storage for the MagicLink service.
type Storage[CustomCreateArgs, CustomReadResponse any] interface {
	// CreateLink creates a secret for the given parameters and stores the pair. The secret is returned to the caller.
	CreateLink(ctx context.Context, args CreateArgs[CustomCreateArgs]) (secret string, err error)
	// ReadLink finds the creation parameters for the given secret. ErrLinkNotFound is returned if the secret is not
	// found or was deleted/expired. This will automatically expire the link.
	ReadLink(ctx context.Context, secret string) (ReadResponse[CustomCreateArgs, CustomReadResponse], error)
}

var _ Storage[any, any] = &memoryMagicLink[any, any, any]{}

type memoryMagicLink[CustomCreateArgs, CustomReadResponse any] struct {
	links map[string]ReadResponse[CustomCreateArgs, CustomReadResponse]
	mux   sync.Mutex
}

// NewMemoryStorage creates an in-memory implementation of the MagicLink Storage.
func NewMemoryStorage[CustomCreateArgs, CustomReadResponse any]() Storage[CustomCreateArgs, CustomReadResponse] {
	return &memoryMagicLink[CustomCreateArgs, CustomReadResponse]{
		links: map[string]ReadResponse[CustomCreateArgs, CustomReadResponse]{},
	}
}
func (m *memoryMagicLink[CustomCreateArgs, CustomReadResponse]) CreateLink(_ context.Context, args CreateArgs[CustomCreateArgs]) (secret string, err error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	u, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID as secret: %w", err)
	}
	secret = u.String()
	var custom CustomReadResponse
	response := ReadResponse[CustomCreateArgs, CustomReadResponse]{
		Custom:     custom,
		CreateArgs: args,
	}
	m.links[secret] = response
	return secret, nil
}
func (m *memoryMagicLink[CustomCreateArgs, CustomReadResponse]) ReadLink(_ context.Context, secret string) (ReadResponse[CustomCreateArgs, CustomReadResponse], error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	args, ok := m.links[secret]
	if !ok {
		return args, ErrLinkNotFound
	}
	delete(m.links, secret)
	return args, nil
}
