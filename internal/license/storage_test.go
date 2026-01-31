package license

import (
	"os"
	"testing"
)

func TestStorageEncryption(t *testing.T) {
	hwid := "test-hwid-123"
	data := "this-is-a-secret-license-key"
	filename := "test_license.enc"
	defer os.Remove(filename)

	// Encrypt and save
	err := SaveLicenseEncrypted(filename, data, hwid)
	if err != nil {
		t.Fatalf("Failed to save encrypted license: %v", err)
	}

	// Decrypt with correct HWID
	decrypted, err := LoadLicenseEncrypted(filename, hwid)
	if err != nil {
		t.Fatalf("Failed to load encrypted license: %v", err)
	}

	if decrypted != data {
		t.Errorf("Decrypted data mismatch, got %s, expected %s", decrypted, data)
	}

	// Try to decrypt with WRONG HWID
	_, err = LoadLicenseEncrypted(filename, "wrong-hwid")
	if err == nil {
		t.Error("Expected error when decrypting with wrong HWID, got nil")
	}
}
