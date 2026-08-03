package recycle

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestDecideUsesEnabledRuleValues(t *testing.T) {
	rule := Rule{
		MinAcceptAvgPrice: decimal.RequireFromString("20.00"),
		RecycleRate:       decimal.RequireFromString("0.25"),
		MinSampleCount:    5,
	}
	book := BookInput{Exists: true, Source: "manual"}
	price := PriceInput{
		HasPrice:    true,
		AvgPrice:    decimalPtr("40.00"),
		SampleCount: 6,
		Confidence:  "MEDIUM",
	}

	result := Decide(book, price, rule)

	if result.Decision != DecisionAccept {
		t.Fatalf("decision = %s, want %s", result.Decision, DecisionAccept)
	}
	if result.SuggestedRecyclePrice == nil {
		t.Fatal("suggested price = nil, want decimal")
	}
	if !result.SuggestedRecyclePrice.Equal(decimal.RequireFromString("10.00")) {
		t.Fatalf("suggested = %s, want 10.00", result.SuggestedRecyclePrice.String())
	}
}

func TestDecideAcceptsExistingBookWithLowSampleByDefault(t *testing.T) {
	rule := Rule{
		MinAcceptAvgPrice: decimal.RequireFromString("10.00"),
		RecycleRate:       decimal.RequireFromString("0.30"),
		MinSampleCount:    5,
	}
	result := Decide(
		BookInput{Exists: true, Source: "manual"},
		PriceInput{HasPrice: true, AvgPrice: decimalPtr("50.00"), SampleCount: 4, Confidence: "HIGH"},
		rule,
	)

	if result.Decision != DecisionAccept {
		t.Fatalf("decision = %s, want %s", result.Decision, DecisionAccept)
	}
	if result.SuggestedRecyclePrice == nil || !result.SuggestedRecyclePrice.Equal(decimal.RequireFromString("15.00")) {
		t.Fatalf("suggested = %v, want 15.00", result.SuggestedRecyclePrice)
	}
}

func TestDecideAcceptsExistingBookWithoutPrice(t *testing.T) {
	result := Decide(
		BookInput{Exists: true, Source: "manual"},
		PriceInput{},
		DefaultRule(),
	)

	if result.Decision != DecisionAccept {
		t.Fatalf("decision = %s, want %s", result.Decision, DecisionAccept)
	}
	if result.SuggestedRecyclePrice != nil {
		t.Fatalf("suggested = %v, want nil", result.SuggestedRecyclePrice)
	}
}

func TestDefaultRuleMatchesExistingBehavior(t *testing.T) {
	rule := DefaultRule()
	result := Decide(
		BookInput{Exists: true, Source: "mock"},
		PriceInput{HasPrice: true, AvgPrice: decimalPtr("24.50"), SampleCount: 8, Confidence: "HIGH"},
		rule,
	)

	if result.Decision != DecisionAccept {
		t.Fatalf("decision = %s, want %s", result.Decision, DecisionAccept)
	}
	if result.SuggestedRecyclePrice == nil || !result.SuggestedRecyclePrice.Equal(decimal.RequireFromString("7.35")) {
		t.Fatalf("suggested = %v, want 7.35", result.SuggestedRecyclePrice)
	}
}

func decimalPtr(value string) *decimal.Decimal {
	amount := decimal.RequireFromString(value)
	return &amount
}
