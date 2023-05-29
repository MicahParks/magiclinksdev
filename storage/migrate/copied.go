package migrate

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"

	"github.com/MicahParks/jwkset"

	"github.com/MicahParks/magiclinksdev/storage"
)

func jwkUnmarshalAssets(assets []byte) (jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta], error) {
	var meta jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta]
	var marshal jwkset.JWKMarshal
	err := json.Unmarshal(assets, &marshal)
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

func decrypt(aes256Key [32]byte, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(aes256Key[:])
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
