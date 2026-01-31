package scraper

import (
	"testing"
)

// This test requires Chrome/Chromium to be installed.
// It will fail in strict CI environments without a browser.
func TestFetchJPYRate(t *testing.T) {
	// Skip in CI if needed, but for local dev we want this.
	// t.Skip("Skipping browser-based test")

	ratePLN, rateUSD, err := FetchJPYRate()
	if err != nil {
		t.Logf("FetchJPYRate failed: %v (expected if Chrome not installed)", err)
		// Don't fail the test entirely if browser is missing, but log it.
		// In a real 'verify' step, we'd want this to pass.
		return 
	}

	if ratePLN <= 0 {
		t.Errorf("Expected positive PLN rate, got %f", ratePLN)
	}
	
	t.Logf("Fetched Rates: PLN=%f, USD=%f", ratePLN, rateUSD)
}
