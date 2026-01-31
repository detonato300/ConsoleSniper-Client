package scraper

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func SearchOLX(query string) ([]interface{}, error) {
	// Format URL for electronics/consoles category
	searchURL := fmt.Sprintf("https://www.olx.pl/elektronika/gry-konsole/konsole/q-%s/?search[filter_float_price:from]=200&search[order]=created_at:desc", url.PathEscape(query))

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var nodes []map[string]string

	err := chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		chromedp.Sleep(5*time.Second),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('div[data-cy="l-card"]')).map(card => {
				let title = card.querySelector('h6, h4')?.innerText || "";
				let priceText = card.querySelector('p[data-testid="ad-price"]')?.innerText || "";
				let link = card.querySelector('a')?.href || "";
				let img = card.querySelector('img')?.src || "";
				
				return {
					title: title,
					price_text: priceText,
					url: link,
					image_url: img
				};
			})
		`, &nodes),
	)
	if err != nil {
		return nil, fmt.Errorf("olx search failed: %w", err)
	}

	var results []interface{}
	for _, raw := range nodes {
		if raw["title"] == "" || raw["url"] == "" {
			continue
		}

		// Clean price
		priceClean := regexp.MustCompile(`[^\d]`).ReplaceAllString(strings.ReplaceAll(raw["price_text"], " ", ""), "")
		price, _ := strconv.Atoi(priceClean)

		results = append(results, map[string]interface{}{
			"title":     raw["title"],
			"price":     price,
			"url":       raw["url"],
			"image_url": raw["image_url"],
			"source":    "olx",
		})
	}

	return results, nil
}
