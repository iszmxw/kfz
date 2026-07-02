package response

type BookDTO struct {
	Isbn           string `json:"isbn"`
	NormalizedIsbn string `json:"normalized_isbn"`
	Title          string `json:"title"`
	Author         string `json:"author"`
	Publisher      string `json:"publisher"`
	PublishYear    string `json:"publish_year"`
	CoverURL       string `json:"cover_url"`
}

type MarketPriceDTO struct {
	Source      string   `json:"source"`
	Min         *float64 `json:"min"`
	Avg         *float64 `json:"avg"`
	Max         *float64 `json:"max"`
	SampleCount int      `json:"sample_count"`
	Confidence  string   `json:"confidence"`
	CollectedAt *string  `json:"collected_at"`
}

type DuplicateInfoDTO struct {
	ScannedRecently         bool    `json:"scanned_recently"`
	DuplicateInCurrentBatch bool    `json:"duplicate_in_current_batch"`
	LastScannedAt           *string `json:"last_scanned_at"`
}

type BookCheckResponse struct {
	ScanLogID             string           `json:"scan_log_id"`
	Book                  BookDTO          `json:"book"`
	Decision              string           `json:"decision"`
	Reason                string           `json:"reason"`
	SuggestedRecyclePrice *float64         `json:"suggested_recycle_price"`
	MarketPrice           *MarketPriceDTO  `json:"market_price"`
	Duplicate             DuplicateInfoDTO `json:"duplicate"`
	BatchActionRequired   string           `json:"batch_action_required,omitempty"`
	ScannedAt             string           `json:"scanned_at"`
}

type LatestDecisionDTO struct {
	Decision              string   `json:"decision"`
	Reason                string   `json:"reason"`
	SuggestedRecyclePrice *float64 `json:"suggested_recycle_price"`
	UpdatedAt             string   `json:"updated_at"`
}

type BookDetailResponse struct {
	Book              BookDTO           `json:"book"`
	LatestMarketPrice *MarketPriceDTO   `json:"latest_market_price"`
	LatestDecision    LatestDecisionDTO `json:"latest_decision"`
}
