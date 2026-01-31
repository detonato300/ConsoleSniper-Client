package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	genai "google.golang.org/genai"
)

type GeminiClient struct {
	client *genai.Client
	model  string
}

func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:      apiKey,
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1beta"},
	})
	if err != nil {
		slog.Error("Failed to create Gemini client", "error", err)
		return nil, fmt.Errorf("failed to create gemini client: %v", err)
	}

	return &GeminiClient{
		client: client,
		model:  "gemini-3-flash-preview", // State-of-the-art vision & analysis
	}, nil
}

func (g *GeminiClient) AnalyzeGrid(ctx context.Context, screenshotData []byte, platform string) (string, error) {
	slog.Info("Analyzing grid", "platform", platform)
	marketDesc := "Japanese handheld consoles"
	if platform == "olx" || platform == "allegro" {
		marketDesc = "Polish local marketplace for electronics"
	}

	prompt := fmt.Sprintf(`
<role>
You are a Market Trend Analyst for the Empire Trading Network. Your expertise is in the %s market.
</role>

<task>
Analyze this search result page screenshot.
1. Assess the overall market sentiment (e.g., are prices inflated, is there a surge of new listings, or is it a buyer's market?).
2. Identify specific items that look like potential "Hidden Gems" or are in exceptionally good condition for their price.
3. Output a list of item IDs or URLs that deserve a "Deep-Dive" analysis.
</task>

<output_format>
Respond in JSON format: 
{
  "sentiment": "Concise market summary",
  "candidates": [{"id": "...", "reason": "..."}]
}
</output_format>`, marketDesc)

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
	slog.Info("Performing auction deep-dive")
	prompt := fmt.Sprintf(`
<role>
You are an Elite Hardware Auditor. Your mission is to perform a surgical deep-dive analysis of this Yahoo Japan Auction listing.
</role>

<context>
Description: %s
</context>

<task>
1. **Defect Detection:** Scan the text and images for hidden flaws (e.g., screen yellowing, hinge stress marks, non-original parts, "sticking" buttons).
2. **Visual Inspection:** Evaluate the exterior condition (scratches, dents) and the display quality.
3. **Classification:** Assign a grade (S: Mint, A: Good, B: Used, C: Heavy Wear) or "Working JUNK".
4. **Strategic Recommendation:** Determine if it's a profitable "BUY" and set a maximum bidding limit in JPY.
</task>

<output_format>
Respond in JSON format: 
{
  "grade": "S/A/B/C/JUNK", 
  "risk_score": 0-10, 
  "recommendation": "Detailed expert opinion", 
  "bid_limit_jpy": 0
}
</output_format>`, description)

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
	slog.Info("Refining trend results", "query", query, "count", len(items))
	itemsJSON, _ := json.Marshal(items)
	prompt := fmt.Sprintf(`
<role>
You are a Precision Data Filter. Your goal is to clean market trend data.
</role>

<task>
Review this list of products found for the query: "%s".
Filter this list to include ONLY items that are an EXACT match for the console model requested. 

Rules:
1. Remove unrelated models (e.g., if query is "New 3DS XL", remove "2DS", "Old 3DS", or "New 3DS Small").
2. Remove accessory-only listings (e.g., chargers, cases, boxes, parts) unless the console is included.
3. Keep bundles if the main console is the correct model.
4. Normalize titles for clarity.
</task>

<context>
Items: %s
</context>

<output_format>
Respond in JSON format: 
{
  "refined_items": [{"title": "...", "price": 0, "url": "..."}]
}
</output_format>`, query, string(itemsJSON))

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