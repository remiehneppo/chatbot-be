package service

import (
	"context"

	"github.com/tieubaoca/chatbot-be/types"
)

type AIService interface {
	Chat(ctx context.Context, prompt string, messages []types.Message) (string, error)
}
