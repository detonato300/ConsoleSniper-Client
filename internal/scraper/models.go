package scraper

type Item struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Price     int    `json:"price"`
	PriceUSD  string `json:"price_usd"`
	URL       string `json:"url"`
	ImageURL  string `json:"image_url"`
	Status    string `json:"status"`
	Source    string `json:"source"`
	IsHighValue bool `json:"is_high_value"`
	HighValueTags []string `json:"high_value_tags"`
}
