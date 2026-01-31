package main

import (
	"agent_client/internal/scraper"
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	ratePLN, rateUSD, err := scraper.FetchJPYRate()
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("Chromedp Error: %v\n", err)
	} else {
		fmt.Printf("Chromedp Success: RatePLN=%.4f, RateUSD=%.4f, Time=%v\n", ratePLN, rateUSD, duration)
	}
}

