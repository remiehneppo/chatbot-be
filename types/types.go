package types

import (
	"context"

	"github.com/tieubaoca/chatbot-be/database"
)

const (
	TypeWebsocketPing       = "ping"
	TypeWebsocketPong       = "pong"
	TypeWebsocketChat       = "chat"
	TypeWebsocketProcessing = "processing"
	TypeWebsocketError      = "error"
)

type WebsocketRequest struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type WebSocketChatPayload struct {
	Messages []Message `json:"messages"`
}

type WebSocketResponse struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type WebSocketChatResponse struct {
	Message string `json:"message"`
}

type WebSocketProcessingResponse struct {
	Message string `json:"message"`
}

// Message represents a single message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// FunctionHandler is a type for handling function calls
type FunctionHandler func(ctx context.Context, args []byte) (any, error)

// Handle stream responses
type StreamHandler func(response string)

type AskAIWithRAGRequest struct {
	Question      string        `json:"question"`
	SearchRequest SearchRequest `json:"search_request"`
}

type SearchRequest struct {
	Queries []string `json:"queries"`
	Tags    []string `json:"tags,omitempty"`
	Limit   int      `json:"limit,omitempty"`
}

type SearchResponse struct {
	Documents []database.Document `json:"documents"`
}
