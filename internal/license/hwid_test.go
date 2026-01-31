package license

import (
	"testing"
)

func TestGetEntropicHWID(t *testing.T) {
	id, err := GetEntropicHWID()
	if err != nil {
		t.Fatalf("Failed to get Entropic HWID: %v", err)
	}
	
	if len(id) < 32 {
		t.Errorf("Entropic HWID too short, got length %d", len(id))
	}
	
	// Ensure it's deterministic
	id2, _ := GetEntropicHWID()
	if id != id2 {
		t.Errorf("Entropic HWID should be consistent, got %s and %s", id, id2)
	}
}