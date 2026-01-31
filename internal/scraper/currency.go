package scraper

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// FetchJPYRate retrieves the JPY/PLN and JPY/USD exchange rates from Buyee using headless Chrome.
func FetchJPYRate() (float64, float64, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	url := "https://asf.buyee.jp/mercari?keyword=nintendo&lang=pl&currencyCode=PLN"
	var bodyText string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(5*time.Second),
		chromedp.Text("body", &bodyText, chromedp.ByQuery),
	)
	if err != nil {
		return 0, 0, fmt.Errorf("chromedp failed: %w", err)
	}

	ratePLN := 0.027
	rateUSD := 0.0068

	// 1. Extract JPY -> PLN
	rePLN := regexp.MustCompile(`(?i)([\d\s,\x{00a0}]+)\s*(?:JPY|jenów|yen).*?([\d\s,.\x{00a0}]+)\s*PLN`)
	matchPLN := rePLN.FindStringSubmatch(bodyText)
	if len(matchPLN) >= 3 {
		yenStr := regexp.MustCompile(`[^\d]`).ReplaceAllString(strings.ReplaceAll(matchPLN[1], "\u00a0", ""), "")
		plnStr := regexp.MustCompile(`[^\d.]`).ReplaceAllString(strings.ReplaceAll(matchPLN[2], "\u00a0", ""), "")
		yen, _ := strconv.ParseFloat(yenStr, 64)
		pln, _ := strconv.ParseFloat(plnStr, 64)
		if yen > 0 { ratePLN = pln / yen }
	}

	// 2. Extract JPY -> USD (Format: US$ 3,02)
	reUSD := regexp.MustCompile(`(?i)([\d\s,\x{00a0}]+)\s*(?:JPY|jenów|yen).*?US\$\s*([\d\s,.\x{00a0}]+)`)
	matchUSD := reUSD.FindStringSubmatch(bodyText)
	if len(matchUSD) >= 3 {
		yenStr := regexp.MustCompile(`[^\d]`).ReplaceAllString(strings.ReplaceAll(matchUSD[1], "\u00a0", ""), "")
		usdStr := regexp.MustCompile(`[^\d.]`).ReplaceAllString(strings.ReplaceAll(matchUSD[2], "\u00a0", ""), "")
		yen, _ := strconv.ParseFloat(yenStr, 64)
		usd, _ := strconv.ParseFloat(usdStr, 64)
		if yen > 0 { rateUSD = usd / yen }
	}

	return ratePLN, rateUSD, nil
}
