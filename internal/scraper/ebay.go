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

func SearchEbay(query string) ([]interface{}, error) {
	// Sort by Newest First (_sop=10), Items per page 60 (_ipg=60)
	searchURL := fmt.Sprintf("https://www.ebay.com/sch/i.html?_nkw=%s&_sop=10&_ipg=60", url.PathEscape(query))

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
			Array.from(document.querySelectorAll('.s-item')).map(item => {
				let titleEl = item.querySelector('.s-item__title');
				if (!titleEl) return null;
				let title = titleEl.innerText || "";
				if (title.toLowerCase().includes("shop on ebay")) return null;

				let priceEl = item.querySelector('.s-item__price');
				let priceText = priceEl ? priceEl.innerText : "";
				
				let linkEl = item.querySelector('.s-item__link');
				let url = linkEl ? linkEl.href : "";
				
				let imgEl = item.querySelector('.s-item__image-img img, .s-item__image-img');
				let img = imgEl ? (imgEl.src || imgEl.dataset.src) : "";

				return {
					title: title,
					price_text: priceText,
					url: url,
					image_url: img
				};
			}).filter(i => i !== null)
		`, &nodes),
	)
	if err != nil {
		return nil, fmt.Errorf("ebay search failed: %w", err)
	}

	var results []interface{}
	for _, raw := range nodes {
		if raw["title"] == "" || raw["url"] == "" {
			continue
		}

		// eBay prices can be "$200.00" or range "$200 to $300"
		// We take the first number as primary price
		priceClean := regexp.MustCompile(`[\d.,]+`).FindString(strings.ReplaceAll(raw["price_text"], ",", ""))
		price, _ := strconv.ParseFloat(priceClean, 64)

		results = append(results, map[string]interface{}{
			"title":     raw["title"],
			"price":     int(price),
			"price_usd": priceClean,
			"url":       raw["url"],
			"image_url": raw["image_url"],
			"source":    "ebay",
		})
	}

	return results, nil
}
