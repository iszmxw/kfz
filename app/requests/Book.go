package requests

type BookCheck struct {
	Isbn            string `json:"isbn"`
	BatchID         string `json:"batch_id"`
	ClientRequestID string `json:"client_request_id"`
}

type BookDetail struct {
	Isbn string `json:"isbn"`
}
