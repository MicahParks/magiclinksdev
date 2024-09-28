package storage

import (
	"context"
	"errors"

	"github.com/MicahParks/jwkset"
	"github.com/google/uuid"

	"github.com/MicahParks/magiclinksdev/magiclink"
	"github.com/MicahParks/magiclinksdev/model"
)

var (
	// ErrKeySize is returned when the symmetric key for the database is the incorrect size.
	ErrKeySize = errors.New("symmetric key for database is incorrect size")
	// ErrNotFound is returned when a record is not found.
	ErrNotFound = errors.New("not found")
)

// Storage is the interface for magiclinksdev storage.
type Storage interface {
	Begin(ctx context.Context) (Tx, error)
	Close(ctx context.Context) error
	TestingTruncate(ctx context.Context) error

	CreateAdminSA(ctx context.Context, args model.ValidAdminCreateArgs) error
	CreateSA(ctx context.Context, args model.ValidServiceAccountCreateArgs) (model.ServiceAccount, error)
	ReadSA(ctx context.Context, u uuid.UUID) (model.ServiceAccount, error)
	ReadSAFromAPIKey(ctx context.Context, apiKey uuid.UUID) (model.ServiceAccount, error)
	ReadSigningKey(ctx context.Context, options ReadSigningKeyOptions) (jwk jwkset.JWK, err error)
	UpdateDefaultSigningKey(ctx context.Context, keyID string) error

	jwkset.Storage
	magiclink.Storage[MagicLinkCustomCreateArgs, MagicLinkCustomReadResponse]
}

// Tx is the interface for a transaction.
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
