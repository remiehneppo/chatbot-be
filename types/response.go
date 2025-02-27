package types

type DataResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type UploadResponse struct {
	OriginalName string `json:"original_name,omitempty"`
}

type ProcessingDocumentStatus struct {
	Status         string  `json:"status"`
	Message        string  `json:"message"`
	Progress       float64 `json:"progress"`
	TotalPages     int     `json:"total_pages"`
	ProcessedPages int     `json:"processed_pages"`
}
