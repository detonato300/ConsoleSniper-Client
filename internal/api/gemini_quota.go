package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	quotaFileName = "ai_quota.json"
	dailyLimit    = 1500
	defaultThreshold = 100
)

type QuotaStats struct {
	LastReset date `json:"last_reset"`
	RequestsToday int `json:"requests_today"`
}

type date struct {
	time.Time
}

func (d *date) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

func (d date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format("2006-01-02"))
}

func getQuotaFilePath() string {
	if os.Getenv("GO_TEST_MODE") == "true" {
		return quotaFileName
	}
	exePath, _ := os.Executable()
	return filepath.Join(filepath.Dir(exePath), quotaFileName)
}

// IncrementQuota increments the daily request counter.
func IncrementQuota() error {
	stats := loadQuota()
	now := time.Now()
	
	if stats.LastReset.Format("2006-01-02") != now.Format("2006-01-02") {
		stats.RequestsToday = 0
		stats.LastReset = date{now}
	}
	
	stats.RequestsToday++
	return saveQuota(stats)
}

// GetRemainingQuota returns the number of requests left for today.
func GetRemainingQuota() int {
	stats := loadQuota()
	now := time.Now()
	
	if stats.LastReset.Format("2006-01-02") != now.Format("2006-01-02") {
		return dailyLimit
	}
	
	remaining := dailyLimit - stats.RequestsToday
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsLowQuota returns true if remaining requests are below the threshold.
func IsLowQuota(threshold int) bool {
	if threshold <= 0 {
		threshold = defaultThreshold
	}
	return GetRemainingQuota() < threshold
}

func loadQuota() QuotaStats {
	data, err := os.ReadFile(getQuotaFilePath())
	if err != nil {
		return QuotaStats{LastReset: date{time.Now()}, RequestsToday: 0}
	}
	
	var stats QuotaStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return QuotaStats{LastReset: date{time.Now()}, RequestsToday: 0}
	}
	return stats
}

func saveQuota(stats QuotaStats) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return os.WriteFile(getQuotaFilePath(), data, 0644)
}
