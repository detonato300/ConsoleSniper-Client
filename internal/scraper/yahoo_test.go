package scraper

import (
	"testing"
)

func TestYahooScraper_BuildURL(t *testing.T) {
	s := NewYahooScraper()
	params := map[string]string{
		"keyword": "3ds",
		"aucminprice": "1000",
	}
	url := s.BuildURL("3ds", params)
	
	expected := "https://buyee.jp/item/search/query/3ds?aucminprice=1000&currencyCode=PLN&keyword=3ds&lang=pl&o=d&page=1&s=created_time&status=active&translationType=98"
	if url != expected {
		t.Errorf("Expected URL %s, got %s", expected, url)
	}
}

func TestYahooScraper_ParseItem(t *testing.T) {
	s := NewYahooScraper()
	
	rawItem := map[string]string{
		"url": "/item/yahoo/auction/x123456789",
		"title": "Yahoo Auction Item",
		"price_text": "Aktualna cena\n1,500 jenów\nCena Kup Teraz\n3,000 jenów",
		"info_text": "Liczba ofert 5\nPOZOSTAŁY CZAS 1 dzień",
		"image_url": "http://example.com/yahoo.jpg",
	}
	
	item, err := s.ParseItem(rawItem)
	if err != nil {
		t.Fatalf("Failed to parse item: %v", err)
	}
	
	if item.ID != "x123456789" {
		t.Errorf("Expected ID x123456789, got %s", item.ID)
	}
	
	if item.Price != 1500 {
		t.Errorf("Expected price 1500, got %d", item.Price)
	}
}