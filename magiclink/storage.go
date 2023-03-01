package magiclink

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Storage represents the underlying storage for the MagicLink service.
type Storage[CustomCreateArgs, CustomReadResults, CustomKeyMeta any] interface {
	// CreateLink creates a secret for the given parameters and stores the pair. The secret is returned to the caller.
	CreateLink(ctx context.Context, args CreateArgs[CustomCreateArgs]) (secret string, err error)
	// ReadLink finds the creation parameters for the given secret. ErrLinkNotFound is returned if the secret is not
	// found or was deleted/expired. Depending on the implementation, this may or may not delete/expire the pair.
	ReadLink(ctx context.Context, secret string) (ReadResponse[CustomCreateArgs, CustomReadResults], error)
}

var _ Storage[any, any, any] = &memoryMagicLink[any, any, any]{}

type memoryMagicLink[CustomCreateArgs, CustomReadResults, CustomKeyMeta any] struct {
	links map[string]ReadResponse[CustomCreateArgs, CustomReadResults]
	mux   sync.Mutex
}

// NewMemoryStorage creates an in-memory implementation of the MagicLink Storage.
func NewMemoryStorage[CustomCreateArgs, CustomReadResults, CustomKeyMeta any]() Storage[CustomCreateArgs, CustomReadResults, CustomKeyMeta] {
	return &memoryMagicLink[CustomCreateArgs, CustomReadResults, CustomKeyMeta]{
		links: map[string]ReadResponse[CustomCreateArgs, CustomReadResults]{},
	}
}

func (m *memoryMagicLink[CustomCreateArgs, CustomReadResults, CustomKeyMeta]) CreateLink(_ context.Context, args CreateArgs[CustomCreateArgs]) (secret string, err error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	u, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID as secret: %w", err)
	}
	secret = u.String()
	var custom CustomReadResults
	response := ReadResponse[CustomCreateArgs, CustomReadResults]{
		Custom:     custom,
		CreateArgs: args,
	}
	m.links[secret] = response
	return secret, nil
}

func (m *memoryMagicLink[CustomCreateArgs, CustomReadResults, CustomKeyMeta]) ReadLink(_ context.Context, secret string) (ReadResponse[CustomCreateArgs, CustomReadResults], error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	args, ok := m.links[secret]
	if !ok {
		return args, ErrLinkNotFound
	}
	delete(m.links, secret)
	return args, nil
}
