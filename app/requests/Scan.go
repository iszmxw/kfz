package requests

type ScanHistory struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Decision string `json:"decision"`
	Isbn     string `json:"isbn"`
	BatchID  string `json:"batch_id"`
}
