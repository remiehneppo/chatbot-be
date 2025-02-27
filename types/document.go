package types

type DocumentChunk struct {
	Content  string           // The actual text content
	Page     int              // Page number where the chunk is from
	Metadata DocumentMetadata // Associated metadata for the chunk
}

// PDFMetadata contains metadata information for PDF chunks
type DocumentMetadata struct {
	Title      string // Title of the PDF document
	Source     string // Source file path
	PageNum    int    // Current page number
	TotalPages int    // Total number of pages in the document
}

// PDFServiceConfig contains configuration options for PDF processing
type DocumentServiceConfig struct {
	MaxChunkSize int // Maximum size for text chunks
	OverlapSize  int // Size of overlap between chunks
}

type UploadRequest struct {
	Title  string   `json:"title"`
	Source string   `json:"source"`
	Tags   []string `json:"tags"`
}
