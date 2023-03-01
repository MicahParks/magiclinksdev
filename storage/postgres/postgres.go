package postgres

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	jt "github.com/MicahParks/jsontype"
	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/MicahParks/magiclinksdev/magiclink"
	"github.com/MicahParks/magiclinksdev/model"

	"github.com/MicahParks/magiclinksdev/handle"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
	"github.com/MicahParks/magiclinksdev/storage"
)

const (
	//language=sql
	createServiceAccountQuery = `
INSERT INTO mld.service_account (uuid, api_key, aud, is_admin)
VALUES ($1, $2, $3, $4)
RETURNING id
`
)

var _ storage.Storage = postgres{}

// Config is the configuration for Postgres storage.
type Config struct {
	AES256KeyBase64 string                      `json:"aes256KeyBase64"`
	DSN             string                      `json:"dsn"`
	Health          *jt.JSONType[time.Duration] `json:"health"`
	InitialTimeout  *jt.JSONType[time.Duration] `json:"initialTimeout"`
	MaxIdle         *jt.JSONType[time.Duration] `json:"maxIdle"`
	MinConns        int32                       `json:"minConns"`
	PlaintextClaims bool                        `json:"plaintextClaims"`
	PlaintextJWK    bool                        `json:"plaintextJWK"`
	SemVer          string                      `json:"semver"` // https://pkg.go.dev/golang.org/x/mod/semver
}

// DefaultsAndValidate implements the jsontype.Config interface.
func (c Config) DefaultsAndValidate() (Config, error) {
	if !c.PlaintextJWK || !c.PlaintextClaims {
		if c.AES256KeyBase64 == "" {
			return Config{}, fmt.Errorf("AES256 key must be set when plaintext JWK and claims are disabled: %w", jt.ErrDefaultsAndValidate)
		}
		key, err := base64.StdEncoding.DecodeString(c.AES256KeyBase64)
		if err != nil {
			return Config{}, fmt.Errorf("failed to Base64 decode AES256 key: %s: %w", err, jt.ErrDefaultsAndValidate)
		}
		if len(key) != 32 {
			return Config{}, fmt.Errorf("AES256 key must be 32 bytes, but is %d bytes: %w", len(key), jt.ErrDefaultsAndValidate)
		}
	} else {
		if c.AES256KeyBase64 != "" {
			return Config{}, fmt.Errorf("AES256 key must not be set when plaintext JWK and claims are enabled: %w", jt.ErrDefaultsAndValidate)
		}
	}
	if c.DSN == "" {
		return Config{}, fmt.Errorf("DSN must be set: %w", jt.ErrDefaultsAndValidate)
	}
	if c.Health.Get() == 0 {
		c.Health = jt.New(5 * time.Second)
	}
	if c.InitialTimeout.Get() == 0 {
		c.InitialTimeout = jt.New(5 * time.Second)
	}
	if c.MaxIdle.Get() == 0 {
		c.MaxIdle = jt.New(4 * time.Minute)
	}
	if c.MinConns == 0 {
		c.MinConns = 2
	}
	return c, nil
}

type postgres struct {
	aes256Key       [32]byte
	plaintextClaims bool
	plaintextJWK    bool
	pool            *pgxpool.Pool
}

func newPostgres(pool *pgxpool.Pool, config Config) (postgres, error) {
	store := postgres{
		plaintextClaims: config.PlaintextClaims,
		plaintextJWK:    config.PlaintextJWK,
		pool:            pool,
	}
	if !config.PlaintextJWK || !config.PlaintextClaims {
		key, err := base64.StdEncoding.DecodeString(config.AES256KeyBase64)
		if err != nil {
			return postgres{}, fmt.Errorf("failed to Base64 decode AES256 key: %w", err)
		}
		if len(key) != 32 {
			return postgres{}, fmt.Errorf("AES256 key must be 32 bytes, but is %d bytes: %w", len(key), storage.ErrKeySize)
		}
		copy(store.aes256Key[:], key)
	} else {
		if config.AES256KeyBase64 != "" {
			return postgres{}, fmt.Errorf("AES256 key must not be set when plaintext JWK and claims are enabled: %w", storage.ErrKeySize)
		}
	}
	return store, nil
}
func (p postgres) Begin(ctx context.Context) (storage.Tx, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to being Postgres transaction: %w", err)
	}
	return &Transaction{Tx: tx}, nil
}
func (p postgres) Close(_ context.Context) error {
	p.pool.Close()
	return nil
}
func (p postgres) TestingTruncate(ctx context.Context) error {
	tx := ctx.Value(ctxkey.Tx).(*Transaction).Tx

	//language=sql
	const query = `
TRUNCATE TABLE mld.jwk, mld.link, mld.service_account
`
	_, err := tx.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to truncate magic_links table: %w", err)
	}

	return nil
}

func (p postgres) CreateAdminSA(ctx context.Context, args model.ValidAdminCreateArgs) error {
	tx := ctx.Value(ctxkey.Tx).(*Transaction).Tx

	_, err := tx.Exec(ctx, createServiceAccountQuery, args.UUID, args.APIKey, args.Aud, true)
	if err != nil {
		return fmt.Errorf("failed to create admin service account: %w", err)
	}

	return nil
}
func (p postgres) CreateSA(ctx context.Context, _ model.ValidServiceAccountCreateArgs) (model.ServiceAccount, error) {
	tx := ctx.Value(ctxkey.Tx).(*Transaction).Tx

	apiKey, err := uuid.NewRandom()
	if err != nil {
		return model.ServiceAccount{}, fmt.Errorf("failed to generate API key: %w", err)
	}
	aud, err := uuid.NewRandom()
	if err != nil {
		return model.ServiceAccount{}, fmt.Errorf("failed to generate audience: %w", err)
	}
	saUUID, err := uuid.NewRandom()
	if err != nil {
		return model.ServiceAccount{}, fmt.Errorf("failed to generate service account UUID: %w", err)
	}

	_, err = tx.Exec(ctx, createServiceAccountQuery, saUUID, apiKey, aud, false)
	if err != nil {
		return model.ServiceAccount{}, fmt.Errorf("failed to create service account: %w", err)
	}

	sa := model.ServiceAccount{
		UUID:   saUUID,
		APIKey: apiKey,
		Aud:    aud,
		Admin:  false,
	}

	return sa, nil
}
func (p postgres) ReadSA(ctx context.Context, u uuid.UUID) (model.ServiceAccount, error) {
	tx := ctx.Value(ctxkey.Tx).(*Transaction).Tx

	//language=sql
	const queryAud = `
SELECT api_key, aud, is_admin
FROM mld.service_account
WHERE uuid = $1
`
	sa := model.ServiceAccount{
		UUID: u,
	}
	err := tx.QueryRow(ctx, queryAud, u).Scan(&sa.APIKey, &sa.Aud, &sa.Admin)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ServiceAccount{}, fmt.Errorf("failed to read service account audiences from Postgres using UUID: %w: %w", err, storage.ErrNotFound)
		}
		return model.ServiceAccount{}, fmt.Errorf("failed to read service account audiences from Postgres using UUID: %w", err)
	}

	return sa, nil
}
func (p postgres) ReadSAFromAPIKey(ctx context.Context, apiKey uuid.UUID) (model.ServiceAccount, error) {
	tx := ctx.Value(ctxkey.Tx).(*Transaction).Tx

	//language=sql
	const queryAud = `
SELECT uuid, aud, is_admin
FROM mld.service_account
WHERE api_key = $1
`
	sa := model.ServiceAccount{
		APIKey: apiKey,
	}
	err := tx.QueryRow(ctx, queryAud, apiKey).Scan(&sa.UUID, &sa.Aud, &sa.Admin)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ServiceAccount{}, fmt.Errorf("failed to read service account audiences from Postgres using API key: %w: %w", err, storage.ErrNotFound)
		}
		return model.ServiceAccount{}, fmt.Errorf("failed to read service account audiences from Postgres using API key: %w", err)
	}

	return sa, nil
}
func (p postgres) ReadSigningKey(ctx context.Context) (meta jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta], err error) {
	tx := ctx.Value(ctxkey.Tx).(*Transaction).Tx

	//language=sql
	const query = `
SELECT assets
FROM mld.jwk
WHERE signing_default = TRUE
`
	assets := make([]byte, 0)
	err = tx.QueryRow(ctx, query).Scan(&assets)
	if err != nil {
		return meta, fmt.Errorf("failed to read signing key from Postgres: %w", err)
	}

	meta, err = p.jwkUnmarshalAssets(assets)
	if err != nil {
		return meta, fmt.Errorf("failed to unmarshal signing key JWK assets from Postgres: %w", err)
	}

	return meta, nil
}
func (p postgres) ReadSigningKeySet(ctx context.Context, keyID string) error {
	tx := ctx.Value(ctxkey.Tx).(*Transaction).Tx

	//language=sql
	const query = `
UPDATE mld.jwk
SET signing_default = TRUE
WHERE key_id = $1
`
	_, err := tx.Exec(ctx, query, keyID)
	if err != nil {
		return fmt.Errorf("failed to update signing key: %w", err)
	}

	return nil
}

/*
  Magic link storage.
*/

func (p postgres) CreateLink(ctx context.Context, args magiclink.CreateArgs[storage.MagicLinkCustomCreateArgs]) (secret string, err error) {
	tx := ctx.Value(ctxkey.Tx).(*Transaction).Tx
	sa := ctx.Value(ctxkey.ServiceAccount).(model.ServiceAccount)

	s, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate random UUID: %w", err)
	}

	claims, err := p.claimsMarshal(args.JWTClaims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWT claims: %w", err)
	}

	//language=sql
	const query = `
WITH sa AS (SELECT id FROM mld.service_account WHERE uuid = $1)
INSERT
INTO mld.link (expires, jwt_claims, jwt_key_id, jwt_signing_method, redirect_query_key, redirect_url, secret,
                       sa_id)
VALUES ($2, $3, $4, $5, $6, $7, $8, (SELECT id FROM sa))
`
	_, err = tx.Exec(ctx, query, sa.UUID, args.Custom.Expires, claims, args.JWTKeyID, args.JWTSigningMethod, args.RedirectQueryKey, args.RedirectURL.String(), s)
	if err != nil {
		return "", fmt.Errorf("failed to write magic link to Postgres: %w", err)
	}

	return s.String(), nil
}
func (p postgres) ReadLink(ctx context.Context, secret string) (magiclink.ReadResponse[storage.MagicLinkCustomCreateArgs, storage.MagicLinkCustomReadResponse], error) {
	tx := ctx.Value(ctxkey.Tx).(*Transaction).Tx
	var response magiclink.ReadResponse[storage.MagicLinkCustomCreateArgs, storage.MagicLinkCustomReadResponse]

	u, err := uuid.Parse(secret)
	if err != nil {
		return response, fmt.Errorf("failed to parse UUID: %w", magiclink.ErrLinkNotFound)
	}

	//language=sql
	const query = `
UPDATE mld.link updated
SET visited = COALESCE(older.visited, CURRENT_TIMESTAMP)
FROM mld.link older
WHERE older.id = updated.id
  AND updated.secret = $1
RETURNING updated.expires, updated.jwt_claims, updated.jwt_key_id, updated.jwt_signing_method, updated.redirect_query_key, updated.redirect_url, older.visited
`
	claims := make([]byte, 0)
	var args magiclink.CreateArgs[storage.MagicLinkCustomCreateArgs]
	var visited *time.Time
	var redirectURL string
	err = tx.QueryRow(ctx, query, u.String()).Scan(&args.Custom.Expires, &claims, &args.JWTKeyID, &args.JWTSigningMethod, &args.RedirectQueryKey, &redirectURL, &visited)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response, fmt.Errorf("magic link not found: %w", magiclink.ErrLinkNotFound)
		}
		return response, fmt.Errorf("failed to read magic link from Postgres: %w", err)
	}

	if args.Custom.Expires.Before(time.Now()) {
		return response, fmt.Errorf("magic link expired: %w", magiclink.ErrLinkNotFound)
	}

	if visited != nil {
		return response, fmt.Errorf("magic link already visited: %w", magiclink.ErrLinkNotFound)
	}

	args.JWTClaims, err = p.claimsUnmarshal(claims)
	if err != nil {
		return response, fmt.Errorf("failed to unmarshal JWT claims: %w", err)
	}

	args.RedirectURL, err = url.Parse(redirectURL)
	if err != nil {
		return response, fmt.Errorf("failed to parse redirect URL from Postgres: %w", err)
	}

	response.CreateArgs = args
	response.Custom.Visited = visited
	return response, nil
}

/*
JWK Set Storage
*/

func (p postgres) DeleteKey(ctx context.Context, keyID string) (ok bool, err error) {
	tx := ctx.Value(ctxkey.Tx).(*Transaction).Tx

	//language=sql
	const query = `
DELETE FROM mld.jwk
WHERE key_id = $1
`
	res, err := tx.Exec(ctx, query, keyID)
	if err != nil {
		return false, fmt.Errorf("failed to delete JWK from Postgres: %w", err)
	}
	return res.RowsAffected() == 1, nil
}
func (p postgres) ReadKey(ctx context.Context, keyID string) (jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta], error) {
	tx := ctx.Value(ctxkey.Tx).(*Transaction).Tx
	var meta jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta]

	//language=sql
	const query = `
SELECT assets
FROM mld.jwk
WHERE key_id = $1
`

	assets := make([]byte, 0)
	err := tx.QueryRow(ctx, query, keyID).Scan(&assets)
	if err != nil {
		return meta, fmt.Errorf("failed to read JWK from Postgres: %w", err)
	}

	meta, err = p.jwkUnmarshalAssets(assets)
	if err != nil {
		return meta, fmt.Errorf("failed to unmarshal JWK assets from Postgres: %w", err)
	}

	return meta, nil
}
func (p postgres) SnapshotKeys(ctx context.Context) ([]jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta], error) {
	tx := ctx.Value(ctxkey.Tx).(*Transaction).Tx

	//language=sql
	const query = `
SELECT assets, signing_default
FROM mld.jwk
`
	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWKs from Postgres: %w", err)
	}
	defer rows.Close()

	keys := make([]jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta], 0)
	for rows.Next() {
		assets := make([]byte, 0)
		var signingDefault bool
		err = rows.Scan(&assets, &signingDefault)
		if err != nil {
			return nil, fmt.Errorf("failed to scan JWK from Postgres: %w", err)
		}

		meta, err := p.jwkUnmarshalAssets(assets)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JWK assets from Postgres: %w", err)
		}
		meta.Custom.SigningDefault = signingDefault

		keys = append(keys, meta)
	}

	return keys, nil
}
func (p postgres) WriteKey(ctx context.Context, meta jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta]) error {
	tx := ctx.Value(ctxkey.Tx).(*Transaction).Tx

	assets, err := p.jwkMarshalAssets(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal JWK assets: %w", err)
	}

	//language=sql
	const query = `
INSERT INTO mld.jwk (assets, key_id)
VALUES ($1, $2)
`
	_, err = tx.Exec(ctx, query, assets, meta.KeyID)
	if err != nil {
		return fmt.Errorf("failed to write JWK to Postgres: %w", err)
	}

	return nil
}

func (p postgres) setupCheck(ctx context.Context, config Config) error {
	tx := ctx.Value(ctxkey.Tx).(*Transaction).Tx

	//language=sql
	const query = `
SELECT setup
FROM mld.setup
`
	var s setup
	err := tx.QueryRow(ctx, query).Scan(&s)
	if err != nil {
		return fmt.Errorf("failed to read setup from Postgres: %w", err)
	}

	err = compareSemVer(config.SemVer, s.SemVer)
	if err != nil {
		return fmt.Errorf("failed to compare configuration semver with Postgres semver: %w", err)
	}
	if s.PlaintextClaims != config.PlaintextClaims {
		return fmt.Errorf("%w: plaintext claims configuration mismatch", ErrPostgresSetupCheck)
	}
	if s.PlaintextJWK != config.PlaintextJWK {
		return fmt.Errorf("%w: plaintext JWK configuration mismatch", ErrPostgresSetupCheck)
	}

	return nil
}

func (p postgres) claimsMarshal(claims jwt.Claims) ([]byte, error) {
	data, err := json.Marshal(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to JSON marshal claims: %w", err)
	}
	if !p.plaintextClaims {
		data, err = p.encrypt(data)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt claims: %w", err)
		}
	}
	return data, nil
}
func (p postgres) claimsUnmarshal(data []byte) (handle.SigningBytesClaims, error) {
	var err error
	if !p.plaintextClaims {
		data, err = p.decrypt(data)
		if err != nil {
			return handle.SigningBytesClaims{}, fmt.Errorf("failed to decrypt claims: %w", err)
		}
	}
	var claims handle.SigningBytesClaims
	err = json.Unmarshal(data, &claims.Claims)
	if err != nil {
		return handle.SigningBytesClaims{}, fmt.Errorf("failed to JSON unmarshal claims: %w", err)
	}
	return claims, nil
}
func (p postgres) jwkMarshalAssets(meta jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta]) ([]byte, error) {
	options := jwkset.KeyMarshalOptions{
		AsymmetricPrivate: true,
	}
	marshal, err := jwkset.KeyMarshal(meta, options)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JWK: %w", err)
	}

	assets, err := json.Marshal(marshal)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JWK assets: %w", err)
	}

	if !p.plaintextJWK {
		assets, err = p.encrypt(assets)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt JWK: %w", err)
		}
	}

	return assets, nil
}
func (p postgres) jwkUnmarshalAssets(assets []byte) (jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta], error) {
	var err error
	var meta jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta]

	if !p.plaintextJWK {
		assets, err = p.decrypt(assets)
		if err != nil {
			return meta, fmt.Errorf("failed to decrypt JWK: %w", err)
		}
	}

	var marshal jwkset.JWKMarshal
	err = json.Unmarshal(assets, &marshal)
	if err != nil {
		return meta, fmt.Errorf("failed to unmarshal JWK from encrypted assets in Postgres: %w", err)
	}

	options := jwkset.KeyUnmarshalOptions{
		AsymmetricPrivate: true,
	}
	meta, err = jwkset.KeyUnmarshal[storage.JWKSetCustomKeyMeta](marshal, options)
	if err != nil {
		return meta, fmt.Errorf("failed to unmarshal JWK from Postgres: %w", err)
	}

	return meta, nil
}
func (p postgres) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(p.aes256Key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to read random bytes for nonce: %w", err)
	}

	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}
func (p postgres) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(p.aes256Key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt GCM: %w", err)
	}

	return plaintext, nil
}
