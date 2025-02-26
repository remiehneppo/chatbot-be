package services

import "context"

// Message represents a single message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// FunctionHandler is a type for handling function calls
type FunctionHandler func(ctx context.Context, args []byte) (any, error)

// Handle stream responses
type StreamHandler func(response string)

type AIService interface {
	Chat(ctx context.Context, prompt string, messages []Message) (string, error)
}
