package storage

import (
	"context"
	"errors"

	"github.com/MicahParks/jwkset"
	"github.com/google/uuid"

	"github.com/MicahParks/magiclinksdev/magiclink"
	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/otp"
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

	SAAdminCreate(ctx context.Context, args model.ValidAdminCreateParams) error
	SACreate(ctx context.Context, args model.ValidServiceAccountCreateParams) (model.ServiceAccount, error)
	SARead(ctx context.Context, u uuid.UUID) (model.ServiceAccount, error)
	SAReadFromAPIKey(ctx context.Context, apiKey uuid.UUID) (model.ServiceAccount, error)
	SigningKeyRead(ctx context.Context, options ReadSigningKeyOptions) (jwk jwkset.JWK, err error)
	SigningKeyDefaultRead(ctx context.Context) (jwk jwkset.JWK, err error)
	SigningKeyDefaultUpdate(ctx context.Context, keyID string) error

	jwkset.Storage
	magiclink.Storage
	otp.Storage
}

// Tx is the interface for a transaction.
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
