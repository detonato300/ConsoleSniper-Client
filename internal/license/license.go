package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

type LicensePayload struct {
	UserID   int    `json:"user_id"`
	Tier     string `json:"tier"`
	Expires  string `json:"expires"`
	HWIDLock bool   `json:"hwid_lock"`
}

func VerifyLicense(licenseStr string, publicKeyHex string) (*LicensePayload, error) {
	licenseStr = strings.TrimSpace(licenseStr)
	parts := strings.Split(licenseStr, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid license format")
	}

	payloadB64 := strings.TrimSpace(parts[0])
	signatureB64 := strings.TrimSpace(parts[1])

	payloadBytes, err := base64.StdEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, errors.New("failed to decode payload")
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, errors.New("failed to decode signature")
	}

	publicKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, errors.New("failed to decode public key hex")
	}

	if len(publicKeyBytes) != ed25519.PublicKeySize {
		return nil, errors.New("invalid public key size")
	}

	pub := ed25519.PublicKey(publicKeyBytes)

	if !ed25519.Verify(pub, payloadBytes, signatureBytes) {
		return nil, errors.New("invalid signature")
	}

	var payload LicensePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, errors.New("failed to parse payload JSON")
	}

	return &payload, nil
}
