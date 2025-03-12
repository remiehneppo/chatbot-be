package handler

import (
	"encoding/json"
	"net/http"

	"github.com/tieubaoca/chatbot-be/service"
	"github.com/tieubaoca/chatbot-be/types"
)

type ChatHandler struct {
	aiService *service.OpenAIService
}

func NewChatHandler(aiService *service.OpenAIService) *ChatHandler {
	return &ChatHandler{
		aiService: aiService,
	}
}

func (h *ChatHandler) HandleChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var chatRequest types.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&chatRequest); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		response, err := h.aiService.Chat(r.Context(), chatRequest.Messages)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(
			types.DataResponse{
				Status: true,
				Data: types.ChatResponse{
					Message: response,
				},
			},
		)
	}
}
