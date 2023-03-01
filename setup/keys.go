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
func CreateKeysIfNotExists(ctx context.Context, store storage.Storage) (keys []jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta], existed bool, err error) {
	snapshot, err := store.SnapshotKeys(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get snapshot of keys: %w", err)
	}
	if len(snapshot) > 0 {
		defaultEdDSA := false
		haveEdDSA := false
		haveRS256 := false
		createdKeys := make([]jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta], len(snapshot))
		for i, meta := range snapshot {
			switch meta.ALG {
			case jwkset.AlgEdDSA:
				haveEdDSA = true
				if meta.Custom.SigningDefault {
					defaultEdDSA = true
				}
			case jwkset.AlgRS256:
				haveRS256 = true
			}
			createdKeys[i] = meta
		}
		if !(defaultEdDSA && haveEdDSA && haveRS256) {
			return nil, false, fmt.Errorf("%w: expected to have an EdDSA key as the default and an RS256 key", ErrJWKSet)
		}
		return createdKeys, true, nil
	}

	_, edPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, false, fmt.Errorf("failed to generate EdDSA key: %w", err)
	}

	rsaPrivate, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, false, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	keys = []jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta]{
		{
			ALG: jwkset.AlgEdDSA,
			Custom: storage.JWKSetCustomKeyMeta{
				SigningDefault: true,
			},
			Key: edPrivate,
		},
		{
			ALG: jwkset.AlgRS256,
			Key: rsaPrivate,
		},
	}
	for i, meta := range keys {
		u, err := uuid.NewRandom()
		if err != nil {
			return nil, false, fmt.Errorf("failed to generate UUID: %w", err)
		}
		meta.KeyID = u.String()
		err = store.WriteKey(ctx, meta)
		if err != nil {
			return nil, false, fmt.Errorf("failed to write key to storage: %w", err)
		}
		if meta.Custom.SigningDefault {
			err = store.ReadSigningKeySet(ctx, meta.KeyID)
			if err != nil {
				return nil, false, fmt.Errorf("failed to set signing default: %w", err)
			}
		}
		keys[i] = meta
	}

	return keys, false, nil
}
