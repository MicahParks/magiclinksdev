package mldtest

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/MicahParks/jwkset"
	"github.com/google/uuid"

	"github.com/MicahParks/magiclinksdev/magiclink"
	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/storage"
)

type testTx struct{}

// Commit helps implement the storage.Tx interface.
func (t testTx) Commit(_ context.Context) error {
	return nil
}

// Rollback helps implement the storage.Tx interface.
func (t testTx) Rollback(_ context.Context) error {
	return nil
}

// TestStorageOptions are the options for creating a test storage.
type TestStorageOptions struct {
	Key   ed25519.PrivateKey
	KeyID string
	SA    model.ServiceAccount
}

type testStorage struct {
	meta jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta]
	sa   map[uuid.UUID]model.ServiceAccount // TODO Need mutex?
}

// NewTestStorage creates a new test storage.
func NewTestStorage(options TestStorageOptions) storage.Storage {
	return &testStorage{
		meta: jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta]{
			ALG: jwkset.AlgEdDSA,
			Custom: storage.JWKSetCustomKeyMeta{
				SigningDefault: true,
			},
			Key:   options.Key,
			KeyID: options.KeyID,
		},
		sa: map[uuid.UUID]model.ServiceAccount{options.SA.UUID: options.SA},
	}
}
func (t *testStorage) Begin(_ context.Context) (storage.Tx, error) {
	return testTx{}, nil
}
func (t *testStorage) Close(_ context.Context) error {
	return nil
}
func (t *testStorage) TestingTruncate(_ context.Context) error {
	return nil
}
func (t *testStorage) CreateAdminSA(_ context.Context, _ model.ValidAdminCreateArgs) error {
	return nil
}
func (t *testStorage) CreateSA(_ context.Context, _ model.ValidServiceAccountCreateArgs) (model.ServiceAccount, error) {
	u := uuid.New()
	apiKey := uuid.New()
	aud := uuid.New()
	sa := model.ServiceAccount{
		UUID:   u,
		APIKey: apiKey,
		Aud:    aud,
		Admin:  false,
	}
	t.sa[u] = sa
	return sa, nil
}
func (t *testStorage) ReadSA(_ context.Context, u uuid.UUID) (model.ServiceAccount, error) {
	sa, ok := t.sa[u]
	if !ok {
		return model.ServiceAccount{}, storage.ErrNotFound
	}
	return sa, nil
}
func (t *testStorage) ReadSAFromAPIKey(_ context.Context, apiKey uuid.UUID) (model.ServiceAccount, error) {
	for _, sa := range t.sa {
		if sa.APIKey == apiKey {
			return sa, nil
		}
	}
	return model.ServiceAccount{}, fmt.Errorf("no service account found with API key %w", storage.ErrNotFound)
}
func (t *testStorage) ReadSigningKey(_ context.Context) (meta jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta], err error) {
	return t.meta, nil
}
func (t *testStorage) ReadSigningKeySet(_ context.Context, _ string) error {
	return nil
}
func (t *testStorage) DeleteKey(_ context.Context, _ string) (ok bool, err error) {
	return true, nil
}
func (t *testStorage) ReadKey(_ context.Context, _ string) (jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta], error) {
	return t.meta, nil
}
func (t *testStorage) SnapshotKeys(_ context.Context) ([]jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta], error) {
	return []jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta]{t.meta}, nil
}
func (t *testStorage) WriteKey(_ context.Context, _ jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta]) error {
	return nil
}
func (t *testStorage) CreateLink(_ context.Context, _ magiclink.CreateArgs[storage.MagicLinkCustomCreateArgs]) (secret string, err error) {
	return uuid.New().String(), nil
}
func (t *testStorage) ReadLink(_ context.Context, _ string) (magiclink.ReadResponse[storage.MagicLinkCustomCreateArgs, storage.MagicLinkCustomReadResponse], error) {
	return magiclink.ReadResponse[storage.MagicLinkCustomCreateArgs, storage.MagicLinkCustomReadResponse]{}, nil
}

// ErrorStorage is a storage.Storage implementation that always returns an error.
type ErrorStorage struct{}

func (e ErrorStorage) Begin(_ context.Context) (storage.Tx, error) {
	return nil, ErrMLDTest
}
func (e ErrorStorage) Close(_ context.Context) error {
	return ErrMLDTest
}
func (e ErrorStorage) TestingTruncate(_ context.Context) error {
	return ErrMLDTest
}
func (e ErrorStorage) CreateAdminSA(_ context.Context, _ model.ValidAdminCreateArgs) error {
	return ErrMLDTest
}
func (e ErrorStorage) CreateSA(_ context.Context, _ model.ValidServiceAccountCreateArgs) (model.ServiceAccount, error) {
	return model.ServiceAccount{}, ErrMLDTest
}
func (e ErrorStorage) ReadSA(_ context.Context, _ uuid.UUID) (model.ServiceAccount, error) {
	return model.ServiceAccount{}, ErrMLDTest
}
func (e ErrorStorage) ReadSAFromAPIKey(_ context.Context, _ uuid.UUID) (model.ServiceAccount, error) {
	return model.ServiceAccount{}, ErrMLDTest
}
func (e ErrorStorage) ReadSigningKey(_ context.Context) (meta jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta], err error) {
	return jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta]{}, ErrMLDTest
}
func (e ErrorStorage) ReadSigningKeySet(_ context.Context, _ string) error {
	return ErrMLDTest
}
func (e ErrorStorage) DeleteKey(_ context.Context, _ string) (ok bool, err error) {
	return true, ErrMLDTest
}
func (e ErrorStorage) ReadKey(_ context.Context, _ string) (jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta], error) {
	return jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta]{}, ErrMLDTest
}
func (e ErrorStorage) SnapshotKeys(_ context.Context) ([]jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta], error) {
	return nil, ErrMLDTest
}
func (e ErrorStorage) WriteKey(_ context.Context, _ jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta]) error {
	return ErrMLDTest
}
func (e ErrorStorage) CreateLink(_ context.Context, _ magiclink.CreateArgs[storage.MagicLinkCustomCreateArgs]) (secret string, err error) {
	return "", ErrMLDTest
}
func (e ErrorStorage) ReadLink(_ context.Context, _ string) (magiclink.ReadResponse[storage.MagicLinkCustomCreateArgs, storage.MagicLinkCustomReadResponse], error) {
	return magiclink.ReadResponse[storage.MagicLinkCustomCreateArgs, storage.MagicLinkCustomReadResponse]{}, ErrMLDTest
}
