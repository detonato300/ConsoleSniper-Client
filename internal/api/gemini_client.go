package api

import (
	"context"
	"encoding/json"
	"fmt"

	genai "google.golang.org/genai"
)

type GeminiClient struct {
	client *genai.Client
	model  string
}

func NewGeminiClient(apiKey string) (*GeminiClient, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:      apiKey,
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1beta"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %v", err)
	}

	return &GeminiClient{
		client: client,
		model:  "gemini-3-flash-preview", // State-of-the-art vision & analysis
	}, nil
}

func (g *GeminiClient) AnalyzeGrid(ctx context.Context, screenshotData []byte, platform string) (string, error) {
	marketDesc := "Japanese handheld consoles"
	if platform == "olx" || platform == "allegro" {
		marketDesc = "Polish local marketplace for electronics"
	}

	prompt := fmt.Sprintf(`Analyze this search result page for %s.
1. Assess the overall market sentiment (e.g., prices high/low, many new listings).
2. Identify items that look like potential deals or are in exceptionally good condition.
3. Output a list of item IDs or URLs that deserve a "Deep-Dive" analysis.
Respond in JSON format: {"sentiment": "...", "candidates": [{"id": "...", "reason": "..."}]}`, marketDesc)

	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: prompt},
				{
					InlineData: &genai.Blob{
						MIMEType: "image/png",
						Data:     screenshotData,
					},
				},
			},
		},
	}

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.model, contents, config)
	if err != nil {
		return "", err
	}

	return g.extractText(resp), nil
}

func (g *GeminiClient) AuctionDeepDive(ctx context.Context, screenshotData []byte, description string) (string, error) {
	prompt := fmt.Sprintf(`Perform a deep-dive analysis of this Yahoo Japan Auction listing.
Description: %s

Tasks:
1. Identify any hidden defects mentioned in the text (e.g., screen yellowing, scratches, non-original parts).
2. Evaluate the visual condition from the images.
3. Assign a grade (S, A, B, C) or "Working JUNK".
4. Recommend if it's worth bidding and up to what price (in JPY).

Respond in JSON format: {"grade": "...", "risk_score": 0-10, "recommendation": "...", "bid_limit_jpy": 0}`, description)

	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: prompt},
				{
					InlineData: &genai.Blob{
						MIMEType: "image/png",
						Data:     screenshotData,
					},
				},
			},
		},
	}

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.model, contents, config)
	if err != nil {
		return "", err
	}

	return g.extractText(resp), nil
}

func (g *GeminiClient) extractText(resp *genai.GenerateContentResponse) string {
	var result string
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if part.Text != "" {
					result += part.Text
				}
			}
		}
	}
	return result
}

func (g *GeminiClient) RefineTrendResults(ctx context.Context, query string, items []interface{}) (string, error) {
	itemsJSON, _ := json.Marshal(items)
	prompt := fmt.Sprintf(`Review this list of products found for the query: "%s".
Filter this list to include ONLY items that exactly match the model requested. 
Example: If query is "New 3DS XL", remove "2DS", "Original 3DS", or "New 3DS (Small)".
Keep items that are bundles if the main console is correct.

Items: %s

Respond in JSON format: {"refined_items": [{"title": "...", "price": 0, "url": "..."}]}`, query, string(itemsJSON))

	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{{Text: prompt}},
		},
	}

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.model, contents, config)
	if err != nil {
		return "", err
	}

	return g.extractText(resp), nil
}