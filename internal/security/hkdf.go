package security

import (
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/hkdf"
)

// DeriveSessionKey derives a 32-byte session key from a secret and salt.
func DeriveSessionKey(secret, salt, info string) ([]byte, error) {
	hash := sha256.New
	hkdfReader := hkdf.New(hash, []byte(secret), []byte(salt), []byte(info))

	key := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, key); err != nil {
		return nil, err
	}

	return key, nil
}
