package api

import (
	"os"
	"testing"
	"time"
)

func TestQuotaMonitoring(t *testing.T) {
	os.Setenv("GO_TEST_MODE", "true")
	defer os.Remove(quotaFileName)

	// Initial
	rem := GetRemainingQuota()
	if rem != dailyLimit {
		t.Errorf("Expected initial quota %d, got %d", dailyLimit, rem)
	}

	// Increment
	IncrementQuota()
	rem = GetRemainingQuota()
	if rem != dailyLimit-1 {
		t.Errorf("Expected quota %d after 1 req, got %d", dailyLimit-1, rem)
	}

	// Low quota check
	if IsLowQuota(dailyLimit) == false {
		t.Error("Expected IsLowQuota to be true for threshold == dailyLimit")
	}
}

func TestQuotaReset(t *testing.T) {
	os.Setenv("GO_TEST_MODE", "true")
	defer os.Remove(quotaFileName)

	// Save manual stats from "yesterday"
	yesterday := time.Now().AddDate(0, 0, -1)
	stats := QuotaStats{
		LastReset: date{yesterday},
		RequestsToday: 500,
	}
	saveQuota(stats)

	// Should reset on GetRemainingQuota
	rem := GetRemainingQuota()
	if rem != dailyLimit {
		t.Errorf("Expected reset quota %d, got %d", dailyLimit, rem)
	}
}
