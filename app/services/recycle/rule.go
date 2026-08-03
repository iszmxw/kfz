package recycle

import "github.com/shopspring/decimal"

const (
	DecisionAccept     = "ACCEPT"
	DecisionReject     = "REJECT"
	DecisionNeedReview = "NEED_REVIEW"
	ConfidenceNone     = "NONE"
)

type Rule struct {
	MinAcceptAvgPrice decimal.Decimal
	RecycleRate       decimal.Decimal
	MinSampleCount    int
}

type BookInput struct {
	Exists bool
	Source string
}

type PriceInput struct {
	HasPrice    bool
	AvgPrice    *decimal.Decimal
	SampleCount int
	Confidence  string
}

type DecisionResult struct {
	Decision              string
	Reason                string
	SuggestedRecyclePrice *decimal.Decimal
}

func DefaultRule() Rule {
	return Rule{
		MinAcceptAvgPrice: decimal.RequireFromString("10.00"),
		RecycleRate:       decimal.RequireFromString("0.30"),
		MinSampleCount:    3,
	}
}

func Decide(book BookInput, price PriceInput, rule Rule) DecisionResult {
	if rule.MinSampleCount <= 0 {
		rule.MinSampleCount = 3
	}
	if rule.MinAcceptAvgPrice.IsZero() {
		rule.MinAcceptAvgPrice = decimal.RequireFromString("10.00")
	}
	if rule.RecycleRate.IsZero() {
		rule.RecycleRate = decimal.RequireFromString("0.30")
	}

	if !book.Exists || book.Source == "scan" {
		return DecisionResult{Decision: DecisionReject, Reason: "未找到书籍基础信息"}
	}
	if !price.HasPrice || price.AvgPrice == nil || price.SampleCount <= 0 {
		return DecisionResult{Decision: DecisionAccept, Reason: "书籍存在，默认建议回收"}
	}

	suggested := price.AvgPrice.Mul(rule.RecycleRate).Round(2)
	return DecisionResult{
		Decision:              DecisionAccept,
		Reason:                "书籍存在，默认建议回收",
		SuggestedRecyclePrice: &suggested,
	}
}
