package scraper

import (
	"testing"
)

func TestMercariScraper_Params(t *testing.T) {
	s := NewMercariScraper()
	params := map[string]string{
		"keyword": "gameboy",
		"price_min": "1000",
	}
	url := s.BuildURL(params)
	
	// Buyee Mercari URL with all default parameters (sorted alphabetically by url.Values)
	expected := "https://asf.buyee.jp/mercari?currencyCode=PLN&items=100&keyword=gameboy&lang=pl&languageCode=pl&price_min=1000&searchType=filter&status=on_sale&translationType=98"
	if url != expected {
		t.Errorf("Expected URL %s, got %s", expected, url)
	}
}

func TestMercariScraper_ParseItem(t *testing.T) {
	s := NewMercariScraper()
	
	// Mock JS evaluation result
	rawItem := map[string]string{
		"url": "/mercari/item/m123456789",
		"title": "Test Item",
		"price_raw": "1,234 jenów",
		"image_url": "http://example.com/img.jpg",
	}
	
	item, err := s.ParseItem(rawItem)
	if err != nil {
		t.Fatalf("Failed to parse item: %v", err)
	}
	
	if item.ID != "m123456789" {
		t.Errorf("Expected ID m123456789, got %s", item.ID)
	}
	
	if item.Price != 1234 {
		t.Errorf("Expected price 1234, got %d", item.Price)
	}
	
	expectedURL := "https://buyee.jp/mercari/item/m123456789?lang=pl"
	if item.URL != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, item.URL)
	}
}