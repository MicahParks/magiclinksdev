package mldtest

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"

	"github.com/MicahParks/jwkset"
	"github.com/google/uuid"

	"github.com/MicahParks/magiclinksdev/magiclink"
	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/otp"
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

var _ storage.Storage = &testStorage{}

type testStorage struct {
	jwk jwkset.JWK
	sa  map[uuid.UUID]model.ServiceAccount // TODO Need mutex?
}

func (t *testStorage) toMemory(ctx context.Context) (jwkset.Storage, error) {
	m := jwkset.NewMemoryStorage()
	err := m.KeyWrite(ctx, t.jwk)
	if err != nil {
		return nil, fmt.Errorf("failed to write key to memory: %w", err)
	}
	return m, nil
}

// NewTestStorage creates a new test storage.
func NewTestStorage(options TestStorageOptions) storage.Storage {
	jwkOptions := jwkset.JWKOptions{
		Marshal: jwkset.JWKMarshalOptions{
			Private: true,
		},
		Metadata: jwkset.JWKMetadataOptions{
			ALG: jwkset.AlgEdDSA,
			KID: options.KeyID,
		},
	}
	jwk, err := jwkset.NewJWKFromKey(options.Key, jwkOptions)
	if err != nil {
		panic(err)
	}

	return &testStorage{
		jwk: jwk,
		sa:  map[uuid.UUID]model.ServiceAccount{options.SA.UUID: options.SA},
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
func (t *testStorage) SAAdminCreate(_ context.Context, _ model.ValidAdminCreateParams) error {
	return nil
}
func (t *testStorage) SACreate(_ context.Context, _ model.ValidServiceAccountCreateParams) (model.ServiceAccount, error) {
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
func (t *testStorage) SARead(_ context.Context, u uuid.UUID) (model.ServiceAccount, error) {
	sa, ok := t.sa[u]
	if !ok {
		return model.ServiceAccount{}, storage.ErrNotFound
	}
	return sa, nil
}
func (t *testStorage) SAReadFromAPIKey(_ context.Context, apiKey uuid.UUID) (model.ServiceAccount, error) {
	for _, sa := range t.sa {
		if sa.APIKey == apiKey {
			return sa, nil
		}
	}
	return model.ServiceAccount{}, fmt.Errorf("no service account found with API key %w", storage.ErrNotFound)
}
func (t *testStorage) SigningKeyRead(_ context.Context, _ storage.ReadSigningKeyOptions) (meta jwkset.JWK, err error) {
	return t.jwk, nil
}
func (t *testStorage) SigningKeyDefaultRead(_ context.Context) (jwk jwkset.JWK, err error) {
	return t.jwk, nil
}
func (t *testStorage) SigningKeyDefaultUpdate(_ context.Context, _ string) error {
	return nil
}
func (t *testStorage) KeyDelete(_ context.Context, _ string) (ok bool, err error) {
	return true, nil
}
func (t *testStorage) KeyRead(_ context.Context, _ string) (jwkset.JWK, error) {
	return t.jwk, nil
}
func (t *testStorage) KeyReadAll(_ context.Context) ([]jwkset.JWK, error) {
	return []jwkset.JWK{t.jwk}, nil
}
func (t *testStorage) KeyWrite(_ context.Context, _ jwkset.JWK) error {
	return nil
}
func (t *testStorage) JSON(ctx context.Context) (json.RawMessage, error) {
	m, err := t.toMemory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to memory storage: %w", err)
	}
	return m.JSON(ctx)
}
func (t *testStorage) JSONPublic(ctx context.Context) (json.RawMessage, error) {
	m, err := t.toMemory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to memory storage: %w", err)
	}
	return m.JSONPublic(ctx)
}
func (t *testStorage) JSONPrivate(ctx context.Context) (json.RawMessage, error) {
	m, err := t.toMemory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to memory storage: %w", err)
	}
	return m.JSONPrivate(ctx)
}
func (t *testStorage) JSONWithOptions(ctx context.Context, marshalOptions jwkset.JWKMarshalOptions, validationOptions jwkset.JWKValidateOptions) (json.RawMessage, error) {
	m, err := t.toMemory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to memory storage: %w", err)
	}
	return m.JSONWithOptions(ctx, marshalOptions, validationOptions)
}
func (t *testStorage) Marshal(ctx context.Context) (jwkset.JWKSMarshal, error) {
	m, err := t.toMemory(ctx)
	if err != nil {
		return jwkset.JWKSMarshal{}, fmt.Errorf("failed to convert to memory storage: %w", err)
	}
	return m.Marshal(ctx)
}
func (t *testStorage) MarshalWithOptions(ctx context.Context, marshalOptions jwkset.JWKMarshalOptions, validationOptions jwkset.JWKValidateOptions) (jwkset.JWKSMarshal, error) {
	m, err := t.toMemory(ctx)
	if err != nil {
		return jwkset.JWKSMarshal{}, fmt.Errorf("failed to convert to memory storage: %w", err)
	}
	return m.MarshalWithOptions(ctx, marshalOptions, validationOptions)
}
func (t *testStorage) MagicLinkCreate(_ context.Context, _ magiclink.CreateParams) (secret string, err error) {
	return uuid.New().String(), nil
}
func (t *testStorage) MagicLinkRead(_ context.Context, _ string) (magiclink.ReadResult, error) {
	return magiclink.ReadResult{}, nil
}
func (t *testStorage) OTPCreate(_ context.Context, _ otp.CreateParams) (otp.CreateResult, error) {
	return otp.CreateResult{}, nil
}
func (t *testStorage) OTPValidate(_ context.Context, _, _ string) error {
	return nil
}

// ErrorStorage is a storage.Storage implementation that always returns an error.
type ErrorStorage struct{}

func (e ErrorStorage) Begin(ctx context.Context) (storage.Tx, error) {
	return nil, ErrMLDTest
}
func (e ErrorStorage) Close(ctx context.Context) error {
	return ErrMLDTest
}
func (e ErrorStorage) TestingTruncate(ctx context.Context) error {
	return ErrMLDTest
}
func (e ErrorStorage) SAAdminCreate(ctx context.Context, args model.ValidAdminCreateParams) error {
	return ErrMLDTest
}
func (e ErrorStorage) SACreate(ctx context.Context, args model.ValidServiceAccountCreateParams) (model.ServiceAccount, error) {
	return model.ServiceAccount{}, ErrMLDTest
}
func (e ErrorStorage) SARead(ctx context.Context, u uuid.UUID) (model.ServiceAccount, error) {
	return model.ServiceAccount{}, ErrMLDTest
}
func (e ErrorStorage) SAReadFromAPIKey(ctx context.Context, apiKey uuid.UUID) (model.ServiceAccount, error) {
	return model.ServiceAccount{}, ErrMLDTest
}
func (e ErrorStorage) SigningKeyRead(ctx context.Context, options storage.ReadSigningKeyOptions) (jwk jwkset.JWK, err error) {
	return jwkset.JWK{}, ErrMLDTest
}
func (e ErrorStorage) SigningKeyDefaultRead(ctx context.Context) (jwk jwkset.JWK, err error) {
	return jwkset.JWK{}, ErrMLDTest
}
func (e ErrorStorage) SigningKeyDefaultUpdate(ctx context.Context, keyID string) error {
	return ErrMLDTest
}
func (e ErrorStorage) KeyDelete(ctx context.Context, keyID string) (ok bool, err error) {
	return false, ErrMLDTest
}
func (e ErrorStorage) KeyRead(ctx context.Context, keyID string) (jwkset.JWK, error) {
	return jwkset.JWK{}, ErrMLDTest
}
func (e ErrorStorage) KeyReadAll(ctx context.Context) ([]jwkset.JWK, error) {
	return nil, ErrMLDTest
}
func (e ErrorStorage) KeyWrite(ctx context.Context, jwk jwkset.JWK) error {
	return ErrMLDTest
}
func (e ErrorStorage) JSON(ctx context.Context) (json.RawMessage, error) {
	return nil, ErrMLDTest
}
func (e ErrorStorage) JSONPublic(ctx context.Context) (json.RawMessage, error) {
	return nil, ErrMLDTest
}
func (e ErrorStorage) JSONPrivate(ctx context.Context) (json.RawMessage, error) {
	return nil, ErrMLDTest
}
func (e ErrorStorage) JSONWithOptions(ctx context.Context, marshalOptions jwkset.JWKMarshalOptions, validationOptions jwkset.JWKValidateOptions) (json.RawMessage, error) {
	return nil, ErrMLDTest
}
func (e ErrorStorage) Marshal(ctx context.Context) (jwkset.JWKSMarshal, error) {
	return jwkset.JWKSMarshal{}, ErrMLDTest
}
func (e ErrorStorage) MarshalWithOptions(ctx context.Context, marshalOptions jwkset.JWKMarshalOptions, validationOptions jwkset.JWKValidateOptions) (jwkset.JWKSMarshal, error) {
	return jwkset.JWKSMarshal{}, ErrMLDTest
}
func (e ErrorStorage) MagicLinkCreate(ctx context.Context, params magiclink.CreateParams) (secret string, err error) {
	return "", ErrMLDTest
}
func (e ErrorStorage) MagicLinkRead(ctx context.Context, secret string) (magiclink.ReadResult, error) {
	return magiclink.ReadResult{}, ErrMLDTest
}
func (e ErrorStorage) OTPCreate(ctx context.Context, params otp.CreateParams) (otp.CreateResult, error) {
	return otp.CreateResult{}, ErrMLDTest
}
func (e ErrorStorage) OTPValidate(ctx context.Context, id, o string) error {
	return ErrMLDTest
}
