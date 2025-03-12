package types

type DataResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type PaginateResponse struct {
	Total    int64       `json:"total"`
	Elements interface{} `json:"elements"`
	Page     int64       `json:"page"`
	Limit    int64       `json:"limit"`
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
