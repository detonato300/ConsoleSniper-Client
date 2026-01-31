package worker

import (
	"agent_client/internal/api"
	"agent_client/internal/scraper"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type ScraperInterface interface {
	Search(query string, params map[string]string) ([]interface{}, error)
}

type Worker struct {
	Client *api.Client
}

func (w *Worker) ProcessTask(task *api.Task, scr ScraperInterface) (*WorkerResult, error) {
	start := time.Now()
	var finalData interface{}
	var aiModel string
	var err error

	// Extract common filters from payload
	params := make(map[string]string)
	for k, v := range task.Payload {
		if k != "query" && k != "keyword" && k != "source" {
			params[k] = fmt.Sprintf("%v", v)
		}
	}

	// Dynamic Scraper Selection if not provided
	if scr == nil {
		source, _ := task.Payload["source"].(string)
		if source == "yahoo" {
			scr = scraper.NewYahooScraper()
		} else {
			scr = scraper.NewMercariScraper() // Default to Mercari
		}
	}

	switch task.Type {
	case "mercari_search", "search", "discovery":
		query, _ := task.Payload["query"].(string)
		if query == "" {
			query, _ = task.Payload["keyword"].(string)
		}
		items, serr := scr.Search(query, params)
		if serr != nil {
			err = serr
		} else {
			finalData = map[string]interface{}{"items": items}
		}

	case "grid_analysis":
		url, _ := task.Payload["url"].(string)
		if url == "" {
			return nil, fmt.Errorf("missing URL for grid analysis")
		}

		// 1. Capture Screenshot
		screenshot, serr := scraper.CapturePage(url, 30*time.Second)
		if serr != nil {
			err = serr
		} else {
			// 2. Load AI Config
			cfg, _ := api.LoadAIConfig()
			// 3. Analyze with Gemini
			gClient, _ := api.NewGeminiClient(cfg.GeminiAPIKey)
			aiModel = "gemini-3-flash-preview"
			
			source, _ := task.Payload["source"].(string)
			analysis, serr := gClient.AnalyzeGrid(context.Background(), screenshot, source)
			if serr != nil {
				err = serr
			} else {
				// 4. Update Quota
				api.IncrementQuota()
				var result map[string]interface{}
				json.Unmarshal([]byte(analysis), &result)
				finalData = result
			}
		}

	case "deep_dive":
		url, _ := task.Payload["url"].(string)
		description, _ := task.Payload["description"].(string)
		if url == "" {
			return nil, fmt.Errorf("missing URL for deep dive")
		}

		// 1. Capture Screenshot
		screenshot, serr := scraper.CapturePage(url, 30*time.Second)
		if serr != nil {
			err = serr
		} else {
			// 2. Load AI Config
			cfg, _ := api.LoadAIConfig()
			// 3. Analyze with Gemini
			gClient, _ := api.NewGeminiClient(cfg.GeminiAPIKey)
			aiModel = "gemini-3-flash-preview"
			analysis, serr := gClient.AuctionDeepDive(context.Background(), screenshot, description)
			if serr != nil {
				err = serr
			} else {
				// 4. Update Quota
				api.IncrementQuota()
				var result map[string]interface{}
				json.Unmarshal([]byte(analysis), &result)
				finalData = result
			}
		}
	
	case "priceget":
		ratePLN, rateUSD, serr := scraper.FetchJPYRate()
		if serr != nil {
			err = serr
		} else {
			finalData = map[string]float64{"jpy_rate": ratePLN, "jpy_usd": rateUSD}
		}

	case "trend_check":
		platform, _ := task.Payload["platform"].(string)
		query, _ := task.Payload["query"].(string)
		
		var items []interface{}
		var serr error
		
		if platform == "olx" {
			items, serr = scraper.SearchOLX(query)
		} else if platform == "allegro" {
			items, serr = scraper.SearchAllegro(query)
		} else if platform == "ebay" {
			items, serr = scraper.SearchEbay(query)
		} else {
			serr = fmt.Errorf("unsupported platform: %s", platform)
		}
		
		if serr != nil {
			err = serr
		} else {
			// AI REFINEMENT & VISION (If available)
			cfg, aerr := api.LoadAIConfig()
			if aerr == nil && cfg.GeminiAPIKey != "" {
				gClient, _ := api.NewGeminiClient(cfg.GeminiAPIKey)
				
				// 1. Text Refinement
				if len(items) > 0 {
					refinedJSON, rerr := gClient.RefineTrendResults(context.Background(), query, items)
					if rerr == nil {
						var res struct {
							RefinedItems []interface{} `json:"refined_items"`
						}
						if json.Unmarshal([]byte(refinedJSON), &res) == nil {
							items = res.RefinedItems
						}
					}
				}

				// 2. Vision Analysis (Grid) - For local/major platforms
				var sentiment string
				if platform == "olx" || platform == "allegro" || platform == "ebay" {
					// Capture current search page
					searchURL := ""
					if platform == "olx" {
						searchURL = fmt.Sprintf("https://www.olx.pl/elektronika/gry-konsole/konsole/q-%s/?search[filter_float_price:from]=200&search[order]=created_at:desc", url.PathEscape(query))
					} else if platform == "allegro" {
						searchURL = fmt.Sprintf("https://allegro.pl/listing?string=%s&order=n", url.QueryEscape(query))
					} else if platform == "ebay" {
						searchURL = fmt.Sprintf("https://www.ebay.com/sch/i.html?_nkw=%s&_sop=10", url.PathEscape(query))
					}
					
					if searchURL != "" {
						screenshot, serr := scraper.CapturePage(searchURL, 30*time.Second)
						if serr == nil {
							visionJSON, verr := gClient.AnalyzeGrid(context.Background(), screenshot, platform)
							if verr == nil {
								var vres struct {
									Sentiment string `json:"sentiment"`
								}
								json.Unmarshal([]byte(visionJSON), &vres)
								sentiment = vres.Sentiment
							}
						}
					}
				}

				aiModel = "gemini-3-flash-preview (refiner+vision)"
				finalData = map[string]interface{}{
					"items":     items,
					"sentiment": sentiment,
				}
			} else {
				finalData = map[string]interface{}{"items": items}
			}
		}

	default:
		return nil, fmt.Errorf("unsupported task type: %s", task.Type)
	}

	if err != nil {
		return nil, err
	}

	execTime := time.Since(start).Milliseconds()
	return NewWorkerResult("completed", finalData, execTime, "v3.2.1", aiModel), nil
}
