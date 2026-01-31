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

type MercariScraper struct {
	BaseURL   string
	SearchURL string
}

func NewMercariScraper() *MercariScraper {
	return &MercariScraper{
		BaseURL:   "https://buyee.jp",
		SearchURL: "https://asf.buyee.jp/mercari",
	}
}

func (s *MercariScraper) BuildURL(params map[string]string) string {
	q := url.Values{}
	q.Set("lang", "pl")
	q.Set("languageCode", "pl")
	q.Set("currencyCode", "PLN")
	q.Set("status", "on_sale")
	q.Set("items", "100")
	q.Set("searchType", "filter")
	q.Set("translationType", "98")
	for k, v := range params {
		if v != "" && v != "<nil>" {
			q.Set(k, v)
		}
	}
	return fmt.Sprintf("%s?%s", s.SearchURL, q.Encode())
}

func (s *MercariScraper) Search(query string, filters map[string]string) ([]interface{}, error) {
	// Merge query with filters
	params := make(map[string]string)
	for k, v := range filters {
		params[k] = v
	}
	if query != "" {
		params["keyword"] = query
	}

	targetURL := s.BuildURL(params)

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
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var nodes []map[string]string

	err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.Sleep(5*time.Second),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('a[href*="/mercari/item/"]')).map(anchor => {
				let card = anchor.closest('li') || anchor.parentElement;
				let img = card.querySelector('img');
				let title = card.querySelector('div[class*="item_name"], span[class*="simple_name"]')?.innerText || anchor.title || "";
				let price = card.querySelector('div[class*="item_price"], span[class*="simple_price"]')?.innerText || "";
				
				return {
					url: anchor.href || "",
					title: title.trim(),
					price_raw: price.trim(),
					image_url: img?.src || img?.dataset?.src || img?.srcset?.split(' ')[0] || ""
				};
			})
		`, &nodes),
	)
	if err != nil {
		return nil, fmt.Errorf("mercari search failed: %w", err)
	}

	var results []interface{}
	
	// Get exclusion list from filters if available
	excludeStr := filters["exclude"]
	var exclude []string
	if excludeStr != "" {
		exclude = strings.Split(strings.ToLower(excludeStr), ",")
	} else {
		// Default exclusion list
		exclude = []string{"ケース", "カバー", "バッテリー", "フィルム", "タッチペン", "充電器", "空箱", "説明書", "acアダプター", "cable", "adapter", "case", "cover"}
	}

	for _, raw := range nodes {
		if raw["url"] == "" || raw["title"] == "" { continue }
		
		// Filter out based on exclusion list
		titleLower := strings.ToLower(raw["title"])
		shouldExclude := false
		for _, word := range exclude {
			if word != "" && strings.Contains(titleLower, strings.TrimSpace(word)) {
				shouldExclude = true
				break
			}
		}
		if shouldExclude { continue }

		item, err := s.ParseItem(raw)
		if err == nil {
			results = append(results, item)
		}
	}

	return results, nil
}

func (s *MercariScraper) ParseItem(raw map[string]string) (*Item, error) {
	rawURL := raw["url"]
	
	// Alphanumeric ID support (some ASF IDs are hashed)
	// Example: /undefined/mercari/item/m3rBf35J2GjhhYpUn68MSc
	re := regexp.MustCompile(`mercari/item/([a-zA-Z0-9]+)`)
	match := re.FindStringSubmatch(rawURL)
	if len(match) < 2 {
		return nil, fmt.Errorf("failed to extract ID from URL: %s", rawURL)
	}
	id := match[1]

	title := raw["title"]
	priceRaw := raw["price_raw"]
	imgURL := raw["image_url"]

	priceClean := regexp.MustCompile(`[^\d]`).ReplaceAllString(strings.ReplaceAll(priceRaw, "\u00a0", ""), "")
	price, _ := strconv.Atoi(priceClean)

	// Extract USD price if present (e.g. US$ 3,02)
	reUSD := regexp.MustCompile(`US\$\s*([\d.,]+)`)
	matchUSD := reUSD.FindStringSubmatch(priceRaw)
	priceUSD := ""
	if len(matchUSD) >= 2 {
		priceUSD = matchUSD[1]
	}

	// Robust URL reconstruction - always point to official Buyee Mercari
	fullURL := fmt.Sprintf("https://buyee.jp/mercari/item/%s?lang=pl", id)

	return &Item{
		ID:       id,
		Title:    strings.TrimSpace(title),
		Price:    price,
		PriceUSD: priceUSD,
		URL:      fullURL,
		ImageURL: imgURL,
		Status:   "active",
		Source:   "mercari",
	}, nil
}
