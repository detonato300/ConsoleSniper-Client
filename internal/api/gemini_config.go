package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/denisbrodbeck/machineid"
	"agent_client/internal/security"
)

const (
	configFileName = "ai_config.enc"
	hkdfSalt       = "consolesniper_ai_v1"
	hkdfInfo       = "gemini_api_key_storage"
)

var (
	GeminiApiUrl = "https://generativelanguage.googleapis.com/v1beta/models?key="
)

type AIConfig struct {
	GeminiAPIKey string `json:"gemini_api_key"`
}

// ValidateGeminiKey checks if the provided API key is valid by calling Google's models endpoint.
func ValidateGeminiKey(apiKey string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(GeminiApiUrl + apiKey)
	if err != nil {
		return fmt.Errorf("network error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid API Key (Status: %d)", resp.StatusCode)
	}

	return nil
}

func getConfigPath() string {
	if os.Getenv("GO_TEST_MODE") == "true" {
		return configFileName
	}
	exePath, _ := os.Executable()
	return filepath.Join(filepath.Dir(exePath), configFileName)
}

// SaveAIConfig encrypts and saves the AI configuration to disk.
func SaveAIConfig(apiKey string) error {
	config := AIConfig{GeminiAPIKey: apiKey}
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	mid, err := machineid.ID()
	if err != nil {
		return fmt.Errorf("failed to get machine id: %v", err)
	}

	key, err := security.DeriveSessionKey(mid, hkdfSalt, hkdfInfo)
	if err != nil {
		return fmt.Errorf("key derivation failed: %v", err)
	}

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

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return os.WriteFile(getConfigPath(), ciphertext, 0600)
}

// LoadAIConfig decrypts and loads the AI configuration from disk.
func LoadAIConfig() (*AIConfig, error) {
	ciphertext, err := os.ReadFile(getConfigPath())
	if err != nil {
		return nil, err
	}

	mid, err := machineid.ID()
	if err != nil {
		return nil, fmt.Errorf("failed to get machine id: %v", err)
	}

	key, err := security.DeriveSessionKey(mid, hkdfSalt, hkdfInfo)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	var config AIConfig
	if err := json.Unmarshal(plaintext, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
