package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

func (h *ChatHandler) HandleChat(c *gin.Context) {

	var chatRequest types.ChatRequest
	if err := c.ShouldBindJSON(&chatRequest); err != nil {
		c.JSON(http.StatusBadRequest, types.DataResponse{
			Status:  false,
			Message: "Invalid request body",
		})
	}

	response, err := h.aiService.Chat(c, chatRequest.Messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.DataResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK,
		types.DataResponse{
			Status: true,
			Data: types.ChatResponse{
				Message: response,
			},
		},
	)

}
