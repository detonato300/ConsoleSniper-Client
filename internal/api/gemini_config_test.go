package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestValidateGeminiKey(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "valid-key" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	originalUrl := GeminiApiUrl
	GeminiApiUrl = server.URL + "/?key="
	defer func() { GeminiApiUrl = originalUrl }()

	if err := ValidateGeminiKey("valid-key"); err != nil {
		t.Errorf("Expected valid key to pass, got: %v", err)
	}

	if err := ValidateGeminiKey("invalid-key"); err == nil {
		t.Error("Expected invalid key to fail, got nil")
	}
}

func TestAIConfigStorage(t *testing.T) {
	os.Setenv("GO_TEST_MODE", "true")
	testKey := "test-gemini-api-key-12345"
	
	// Ensure file is removed after test
	defer os.Remove("ai_config.enc")

	err := SaveAIConfig(testKey)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loaded, err := LoadAIConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loaded.GeminiAPIKey != testKey {
		t.Errorf("Expected key %s, got %s", testKey, loaded.GeminiAPIKey)
	}
}

func TestAIConfigTampering(t *testing.T) {
	testKey := "secure-key"
	defer os.Remove("ai_config.enc")

	if err := SaveAIConfig(testKey); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Tamper with the file
	data, _ := os.ReadFile("ai_config.enc")
	data[len(data)-1] ^= 0xFF // Flip last byte
	os.WriteFile("ai_config.enc", data, 0600)

	_, err := LoadAIConfig()
	if err == nil {
		t.Error("Expected error when loading tampered config, got nil")
	}
}

func TestAIConfigEmptyKey(t *testing.T) {
	defer os.Remove("ai_config.enc")
	if err := SaveAIConfig(""); err != nil {
		t.Fatalf("Failed to save empty key: %v", err)
	}
	loaded, _ := LoadAIConfig()
	if loaded.GeminiAPIKey != "" {
		t.Errorf("Expected empty key, got %s", loaded.GeminiAPIKey)
	}
}
