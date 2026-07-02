package response

type ScanHistoryItem struct {
	ScanLogID             string   `json:"scan_log_id"`
	Isbn                  string   `json:"isbn"`
	Title                 string   `json:"title"`
	Decision              string   `json:"decision"`
	Reason                string   `json:"reason"`
	SuggestedRecyclePrice *float64 `json:"suggested_recycle_price"`
	Confidence            string   `json:"confidence"`
	ScannedAt             string   `json:"scanned_at"`
}

type ScanHistoryResponse struct {
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int64             `json:"total"`
	Items    []ScanHistoryItem `json:"items"`
}
