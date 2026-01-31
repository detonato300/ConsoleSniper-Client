package security

import (
	"testing"
)

func TestMeasureJitter(t *testing.T) {
	jitter := MeasureJitter()
	if jitter == 0 {
		t.Error("Jitter measurement should not be zero")
	}
	t.Logf("Measured jitter: %d", jitter)
}

func TestIsHighJitter(t *testing.T) {
	// Mock threshold detection
	threshold := uint64(500)
	
	// Low jitter case
	if IsHighJitter(100, threshold) {
		t.Error("100 should not be considered high jitter for threshold 500")
	}
	
	// High jitter case (Unauthorized Access/VM simulation)
	if !IsHighJitter(1000, threshold) {
		t.Error("1000 should be considered high jitter for threshold 500")
	}
}
