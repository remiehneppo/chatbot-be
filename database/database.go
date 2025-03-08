package database

import (
	"context"

	"github.com/tieubaoca/chatbot-be/types"
)

// VectorDatabase defines the interface for RAG operations
type VectorDatabase interface {
	// Document operations
	UpsertDocument(ctx context.Context, doc *types.Document, embedding []float32) error
	DeleteDocument(ctx context.Context, id string) error

	// Search operations
	SearchSimilar(ctx context.Context, query string, limit int) ([]types.Document, []float32, error)
	SearchByMetadata(ctx context.Context, metadata types.Metadata, limit int) ([]types.Document, error)
	SearchSimilarWithMetadata(ctx context.Context, query string, metadata types.Metadata, limit int) ([]types.Document, []float32, error)
	// Collection operations
	CreateCollection(ctx context.Context, name string, dimension int) error
	DeleteCollection(ctx context.Context, name string) error
}
