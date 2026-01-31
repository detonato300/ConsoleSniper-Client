package license

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"os"
)

func deriveKey(hwid string) []byte {
	hash := sha256.Sum256([]byte(hwid + "ConsoleSniperSalt"))
	return hash[:]
}

func SaveLicenseEncrypted(filepath string, licenseData string, hwid string) error {
	key := deriveKey(hwid)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(licenseData), nil)
	return os.WriteFile(filepath, ciphertext, 0600)
}

func LoadLicenseEncrypted(filepath string, hwid string) (string, error) {
	ciphertext, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	key := deriveKey(hwid)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", io.ErrUnexpectedEOF
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
