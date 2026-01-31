package security

import (
	"bytes"
	"testing"
)

func TestDeriveSessionKey(t *testing.T) {
	secret := "master-secret"
	salt := "random-salt"
	info := "session-context"

	key1, err := DeriveSessionKey(secret, salt, info)
	if err != nil {
		t.Fatalf("Failed to derive key1: %v", err)
	}

	key2, err := DeriveSessionKey(secret, salt, info)
	if err != nil {
		t.Fatalf("Failed to derive key2: %v", err)
	}

	if !bytes.Equal(key1, key2) {
		t.Error("Derived keys should be identical for the same input")
	}

	if len(key1) != 32 {
		t.Errorf("Expected 32-byte key, got %d", len(key1))
	}

	// Different salt should yield different key
	key3, _ := DeriveSessionKey(secret, "different-salt", info)
	if bytes.Equal(key1, key3) {
		t.Error("Different salt should yield different keys")
	}
}
