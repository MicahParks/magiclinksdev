package magiclink_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/MicahParks/jwkset"
)

var (
	ecdsaKID = "ec"
	eddsaKID = "ed"
	rsaKID   = "r"
	hmacKID  = "h"
)

func makeCases[CustomKeyMeta any](t *testing.T) []testCase[CustomKeyMeta] {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	noGivenJWKSStoreCase := testCase[CustomKeyMeta]{
		createArgs: []createArg{
			{},
		},
		name: "No Given JWKS store",
	}

	ec, ed, r := testKeys(t)
	jwksStoreWithAllKeys := jwkset.NewMemoryStorage[CustomKeyMeta]()
	err := jwksStoreWithAllKeys.WriteKey(ctx, jwkset.NewKey[CustomKeyMeta](ec, ecdsaKID))
	if err != nil {
		t.Fatalf("Failed to write EC key to JWKS store: %s", err)
	}
	err = jwksStoreWithAllKeys.WriteKey(ctx, jwkset.NewKey[CustomKeyMeta](ed, eddsaKID))
	if err != nil {
		t.Fatalf("Failed to write ED key to JWKS store: %s", err)
	}
	err = jwksStoreWithAllKeys.WriteKey(ctx, jwkset.NewKey[CustomKeyMeta](r, rsaKID))
	if err != nil {
		t.Fatalf("Failed to write RSA key to JWKS store: %s", err)
	}
	err = jwksStoreWithAllKeys.WriteKey(ctx, jwkset.NewKey[CustomKeyMeta]([]byte("my-hmac-secret"), hmacKID))
	if err != nil {
		t.Fatalf("Failed to write HMAC key to JWKS store: %s", err)
	}

	fourTypesOfKeys := testCase[CustomKeyMeta]{
		createArgs: []createArg{
			{
				JWTKeyID: &ecdsaKID,
			},
			{
				JWTKeyID: &eddsaKID,
			},
			{
				JWTKeyID: &rsaKID,
			},
			{
				JWTKeyID: &hmacKID,
			},
		},
		setupParam: setupArgs[CustomKeyMeta]{
			jwksStore: jwksStoreWithAllKeys,
		},
		name: "Four types of keys present",
	}

	getJWKS := testCase[CustomKeyMeta]{
		setupParam: setupArgs[CustomKeyMeta]{
			jwksGet:   true,
			jwksStore: jwksStoreWithAllKeys,
		},
		name: "Get JWK Set JSON",
	}

	getJWKSCacheRefresh := testCase[CustomKeyMeta]{
		setupParam: setupArgs[CustomKeyMeta]{
			jwksGet:          true,
			jwksGetDelay:     51 * time.Millisecond,
			jwksCacheRefresh: 50 * time.Millisecond,
			jwksStore:        jwksStoreWithAllKeys,
		},
		name: "Get JWK Set JSON after a cache refresh",
	}

	return []testCase[CustomKeyMeta]{
		noGivenJWKSStoreCase,
		fourTypesOfKeys,
		getJWKS,
		getJWKSCacheRefresh,
	}
}

func testKeys(t *testing.T) (*ecdsa.PrivateKey, ed25519.PrivateKey, *rsa.PrivateKey) {
	block, _ := pem.Decode([]byte(ecdsaTestKey))
	ec, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse ECDSA test key: %s", err)
	}
	block, _ = pem.Decode([]byte(eddsaTestKey))
	ed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse EdDSA test key: %s", err)
	}
	block, _ = pem.Decode([]byte(rsaTestKey))
	r, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse RSA test key: %s", err)
	}
	return ec, ed.(ed25519.PrivateKey), r
}

const (
	eddsaTestKey = `-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEILS9sWfJXDG4XX4tt4sKpWo5SPQPd6W9m6maZCNXKPW8
-----END PRIVATE KEY-----
`
	ecdsaTestKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIBIZqJrQCxsoTesKEZZjKZ74EWcQV26vIGQPxkEks0iWoAoGCCqGSM49
AwEHoUQDQgAE/AfJrI0dHT7yPTyjpjS9qC5UWj7GcboNImPnt+taGhaLtQyZdjxY
mtop2MLKszBf8gDVORnRYSS+hf9x2AzNzw==
-----END EC PRIVATE KEY-----`
	rsaTestKey = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`
)
