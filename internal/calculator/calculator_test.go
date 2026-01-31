package calculator

import (
	"agent_client/internal/security"
	"testing"
)

func TestCalculateTotal_Poisoning(t *testing.T) {
	priceJPY := 10000
	weight := 1000
	rate := 0.027
	options := []ShippingOption{
		{Method: "EMS", PriceJPY: 3000, Delivery: "3-5 days"},
	}

	// 1. Clean State
	security.GlobalState.Tainted = false
	resClean := CalculateTotal(priceJPY, weight, rate, options)
	
	// 2. Tainted State
	security.GlobalState.MarkTainted()
	resTainted := CalculateTotal(priceJPY, weight, rate, options)

	if resClean.ItemPLN == resTainted.ItemPLN {
		t.Error("Poisoned result should differ from clean result when tainted")
	}
	
	t.Logf("Clean: %.2f, Tainted: %.2f", resClean.ItemPLN, resTainted.ItemPLN)
}