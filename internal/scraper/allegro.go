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

func SearchAllegro(query string) ([]interface{}, error) {
	// Sort by Newest First (order=n)
	searchURL := fmt.Sprintf("https://allegro.pl/listing?string=%s&order=n", url.QueryEscape(query))

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
			Array.from(document.querySelectorAll('article')).map(item => {
				let titleEl = item.querySelector('h2');
				let title = titleEl ? titleEl.innerText : "";
				
				// Price logic: look for spans containing price classes
				let priceEl = item.querySelector('span[class*="price"]');
				let priceText = priceEl ? priceEl.innerText : "";
				
				let linkEl = item.querySelector('a');
				let url = linkEl ? linkEl.href : "";
				
				// Image: find the first img in the article
				let imgEl = item.querySelector('img');
				let img = imgEl ? (imgEl.src || imgEl.dataset.src || "") : "";

				return {
					title: title,
					price_text: priceText,
					url: url,
					image_url: img
				};
			}).filter(i => i.title !== "" && i.price_text !== "")
		`, &nodes),
	)
	if err != nil {
		return nil, fmt.Errorf("allegro search failed: %w", err)
	}

	var results []interface{}
	for _, raw := range nodes {
		if raw["title"] == "" || raw["url"] == "" {
			continue
		}

		// Clean price: "1 299,00 zł" -> 1299
		priceClean := regexp.MustCompile(`[^\d]`).ReplaceAllString(strings.Split(raw["price_text"], ",")[0], "")
		price, _ := strconv.Atoi(priceClean)

		results = append(results, map[string]interface{}{
			"title":     raw["title"],
			"price":     price,
			"url":       raw["url"],
			"image_url": raw["image_url"],
			"source":    "allegro",
		})
	}

	return results, nil
}
