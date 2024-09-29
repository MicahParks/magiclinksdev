package setup

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/MicahParks/jwkset"
	"github.com/google/uuid"

	"github.com/MicahParks/magiclinksdev/storage"
)

// ErrJWKSet is returned when the JWK Set does not align with setup expectations.
var ErrJWKSet = errors.New("JWK Set did not align with setup expectations")

// CreateKeysIfNotExists creates the keys if they do not exist.
func CreateKeysIfNotExists(ctx context.Context, store storage.Storage) (keys []jwkset.JWK, existed bool, err error) {
	allKeys, err := store.KeyReadAll(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read all JWKs: %w", err)
	}
	if len(allKeys) > 0 {
		defaultEdDSA := false
		haveEdDSA := false
		haveRS256 := false
		existingKeys := make([]jwkset.JWK, len(allKeys))
		for i, jwk := range allKeys {
			switch jwk.Marshal().ALG {
			case jwkset.AlgEdDSA:
				haveEdDSA = true
				// if jwk.Custom.SigningDefault { // TODO Check for signing default.
				// 	defaultEdDSA = true
				// }
			case jwkset.AlgRS256:
				haveRS256 = true
			}
			existingKeys[i] = jwk
		}
		if !(defaultEdDSA && haveEdDSA && haveRS256) {
			return nil, false, fmt.Errorf("%w: expected to have an EdDSA key as the default and an RS256 key", ErrJWKSet)
		}
		return existingKeys, true, nil
	}

	_, edPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, false, fmt.Errorf("failed to generate EdDSA key: %w", err)
	}

	rsaPrivate, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, false, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	jwkOptions := jwkset.JWKOptions{
		Marshal: jwkset.JWKMarshalOptions{
			Private: true,
		},
	}
	jwkOptions.Metadata.ALG = jwkset.AlgEdDSA
	jwkOptions.Metadata.KID = uuid.New().String()
	jwk, err := jwkset.NewJWKFromKey(edPrivate, jwkOptions)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create EdDSA JWK: %w", err)
	}
	err = store.KeyWrite(ctx, jwk)
	if err != nil {
		return nil, false, fmt.Errorf("failed to write EdDSA JWK: %w", err)
	}

	jwkOptions.Metadata.ALG = jwkset.AlgRS256
	jwkOptions.Metadata.KID = uuid.New().String()
	jwk, err = jwkset.NewJWKFromKey(rsaPrivate, jwkOptions)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create RSA JWK: %w", err)
	}
	err = store.KeyWrite(ctx, jwk)
	if err != nil {
		return nil, false, fmt.Errorf("failed to write RSA JWK: %w", err)
	}

	// TODO Signing default and update signing default.

	return keys, false, nil
}
