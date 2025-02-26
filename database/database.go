package database

import (
	"context"
)

// Message represents a chat message
type Message struct {
	ID        string `bson:"_id" json:"id"`
	Role      string `bson:"role" json:"role"`
	Content   string `bson:"content" json:"content"`
	ChatID    string `bson:"chat_id" json:"chat_id"`
	CreatedAt int64  `bson:"created_at" json:"created_at"`
}

// Chat represents a conversation
type Chat struct {
	ID        string `bson:"_id" json:"id"`
	Title     string `bson:"title" json:"title"`
	UserID    string `bson:"user_id" json:"user_id"`
	CreatedAt int64  `bson:"created_at" json:"created_at"`
	UpdatedAt int64  `bson:"updated_at" json:"updated_at"`
}

// Document represents a knowledge base document
type Document struct {
	ID        string   `bson:"_id" json:"id"`
	Content   string   `bson:"content" json:"content"`
	Metadata  Metadata `bson:"metadata" json:"metadata"`
	CreatedAt int64    `bson:"created_at" json:"created_at"`
}

// Metadata contains additional document information
type Metadata struct {
	Title  string            `bson:"title" json:"title"`
	Source string            `bson:"source" json:"source"`
	Tags   []string          `bson:"tags" json:"tags"`
	Custom map[string]string `bson:"custom" json:"custom"`
}

// ChatStore defines the interface for chat-related operations
type ChatStore interface {
	CreateChat(ctx context.Context, chat *Chat) error
	GetChat(ctx context.Context, id string) (*Chat, error)
	ListChats(ctx context.Context, userID string) ([]Chat, error)
	DeleteChat(ctx context.Context, id string) error

	CreateMessage(ctx context.Context, message *Message) error
	GetMessages(ctx context.Context, chatID string) ([]Message, error)
	DeleteMessages(ctx context.Context, chatID string) error
}

// VectorDatabase defines the interface for RAG operations
type VectorDatabase interface {
	// Document operations
	UpsertDocument(ctx context.Context, doc *Document, embedding []float32) error
	DeleteDocument(ctx context.Context, id string) error

	// Search operations
	SearchSimilar(ctx context.Context, query string, limit int) ([]Document, []float32, error)
	SearchByMetadata(ctx context.Context, metadata Metadata, limit int) ([]Document, error)
	SearchSimilarWithMetadata(ctx context.Context, query string, metadata Metadata, limit int) ([]Document, []float32, error)
	// Collection operations
	CreateCollection(ctx context.Context, name string, dimension int) error
	DeleteCollection(ctx context.Context, name string) error
}
