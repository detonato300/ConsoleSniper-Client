package api

import (
	"agent_client/internal/security"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_ReadyForTask_Poisoned(t *testing.T) {
	// 1. Setup Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"task": null}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", "test-hwid")

	// 2. Tainted State - High probability of failure
	security.GlobalState.MarkTainted()
	
	// We might need multiple attempts to trigger the 10% chance
	failed := false
	for i := 0; i < 50; i++ {
		_, err := client.ReadyForTask([]string{"scraping"}, "none", 0, "2.0.0")
		if err != nil && strings.Contains(err.Error(), "simulated 500") {
			failed = true
			break
		}
	}

	if !failed {
		t.Error("API should have returned a simulated 500 error at least once in 50 attempts when tainted")
	}
}
