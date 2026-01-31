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

type YahooScraper struct {
	BaseURL   string
}

func NewYahooScraper() *YahooScraper {
	return &YahooScraper{
		BaseURL:   "https://buyee.jp",
	}
}

func (s *YahooScraper) BuildURL(query string, params map[string]string) string {
	q := url.Values{}
	q.Set("status", "active")
	q.Set("lang", "pl")
	q.Set("currencyCode", "PLN")
	
	// Default: Newest first
	q.Set("s", "created_time")
	q.Set("o", "d")
	
	q.Set("translationType", "98")
	q.Set("page", "1")
	
	if query != "" {
		q.Set("keyword", query)
	}
	
	for k, v := range params {
		if v != "" && v != "<nil>" {
			// Map 'sort' to 's' and 'order' to 'o' if they come from old API style
			if k == "sort" { k = "s" }
			if k == "order" { k = "o" }
			q.Set(k, v)
		}
	}
	
	encodedQuery := url.PathEscape(query)
	if encodedQuery == "" {
		encodedQuery = "all"
	}
	return fmt.Sprintf("%s/item/search/query/%s?%s", s.BaseURL, encodedQuery, q.Encode())
}

func (s *YahooScraper) Search(query string, filters map[string]string) ([]interface{}, error) {
	targetURL := s.BuildURL(query, filters)

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
			Array.from(document.querySelectorAll('li[class*="item"], .auction-item-list-item')).map(item => {
				const link = item.querySelector('a[href*="/item/"]');
				if (!link) return null;
				const img = item.querySelector('img');
				const imgSrc = img?.src || img?.dataset?.src || img?.srcset?.split(' ')[0] || "";
				
				return {
					url: link.href || "",
					title: item.querySelector('a[class*="name"]')?.innerText || link.title || img?.alt || "",
					price_text: item.innerText || "",
					image_url: imgSrc
				};
			}).filter(i => i !== null)
		`, &nodes),
	)
	if err != nil {
		return nil, fmt.Errorf("yahoo search failed: %w", err)
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

func (s *YahooScraper) ParseItem(raw map[string]string) (*Item, error) {
	rawURL := raw["url"]
	if rawURL == "" {
		return nil, fmt.Errorf("missing url")
	}
	
	reID := regexp.MustCompile(`auction/([^/?]+)`)
	match := reID.FindStringSubmatch(rawURL)
	if len(match) < 2 {
		return nil, fmt.Errorf("failed to extract ID from URL")
	}
	id := match[1]

	title := raw["title"]
	priceText := raw["price_text"]
	imgURL := raw["image_url"]

	// Improved regex parsing for Yahoo prices
	// Matches digits/spaces/commas followed by currency labels
	reJPY := regexp.MustCompile(`(?i)([\d\s,]+)\s*(?:jen|yen|JPY)`)
	matchJPY := reJPY.FindStringSubmatch(priceText)
	
	currentPrice := 0
	if len(matchJPY) >= 2 {
		cleanPrice := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(matchJPY[1], ",", ""), " ", ""))
		currentPrice, _ = strconv.Atoi(cleanPrice)
	}

	if currentPrice == 0 {
		// Fallback to any number that looks like a price if the above fails
		reNum := regexp.MustCompile(`[\d\s,]{2,}`)
		match := reNum.FindString(priceText)
		cleanPrice := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(match, ",", ""), " ", ""))
		currentPrice, _ = strconv.Atoi(cleanPrice)
	}

	// Extract USD price if present (e.g. US$ 3,02)
	reUSD := regexp.MustCompile(`US\$\s*([\d.,]+)`)
	matchUSD := reUSD.FindStringSubmatch(priceText)
	priceUSD := ""
	if len(matchUSD) >= 2 {
		priceUSD = matchUSD[1]
	}

	fullURL := rawURL
	if !strings.HasPrefix(fullURL, "http") {
		fullURL = s.BaseURL + rawURL
	}

	return &Item{
		ID:       id,
		Title:    strings.TrimSpace(title),
		Price:    currentPrice,
		PriceUSD: priceUSD,
		URL:      fullURL,
		ImageURL: imgURL,
		Status:   "active",
		Source:   "yahoo",
	}, nil
}
