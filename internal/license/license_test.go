package license

import (
	"testing"
)

func TestVerifyLicense(t *testing.T) {
	publicKeyHex := "99d527d1e8d01d83071466c7ed7398f9ecb110823d0904ed7f96e65e733e8b62"
	validLicense := "eyJleHBpcmVzIjogIm5ldmVyIiwgImh3aWRfbG9jayI6IGZhbHNlLCAidGllciI6ICJwcmVtaXVtIiwgInVzZXJfaWQiOiAxfQ==.uZbFC/E9cKr8GKTjwE0ho0VxhEMWvlsKs6ctdTjKZMJyq9ie/zr5Pic0k8smZp8RPjSCUZbFqPJHeQEXFet2Bw=="

	// Valid case
	payload, err := VerifyLicense(validLicense, publicKeyHex)
	if err != nil {
		t.Fatalf("Failed to verify valid license: %v", err)
	}

	if payload.Tier != "premium" {
		t.Errorf("Expected tier premium, got %s", payload.Tier)
	}

	// Invalid signature case
	invalidLicense := validLicense + "extra"
	_, err = VerifyLicense(invalidLicense, publicKeyHex)
	if err == nil {
		t.Error("Expected error for invalid signature, got nil")
	}

	// Tampered payload case
	tamperedPayload := "eyJleHBpcmVzIjogIm5ldmVyIiwgImh3aWRfbG9jayI6IGZhbHNlLCAidGllciI6ICJmcmVlIiwgInVzZXJfaWQiOiAxfQ==.u3zciK15VNS8EXEnyFI9hRwn4K79bR05gFzlvbId5z9BFUU8fWDvojTmkjKupw0YNDEBvo7z/IwLR0ezL/U9AQ=="
	_, err = VerifyLicense(tamperedPayload, publicKeyHex)
	if err == nil {
		t.Error("Expected error for tampered payload, got nil")
	}
}
