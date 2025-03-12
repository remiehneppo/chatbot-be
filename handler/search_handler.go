package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tieubaoca/chatbot-be/database"
	"github.com/tieubaoca/chatbot-be/types"
)

type SearchHandler struct {
	vectorDB *database.WeaviateStore
}

func NewSearchHandler(vectorDB *database.WeaviateStore) *SearchHandler {
	return &SearchHandler{
		vectorDB: vectorDB,
	}
}

func (h *SearchHandler) HandleSearch(c *gin.Context) {

	var req types.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.DataResponse{
			Status:  false,
			Message: "Invalid request body",
		})
		return
	}

	// Set default limit if not provided
	if req.Limit == 0 {
		req.Limit = 5
	}

	// Search documents
	docs, _, err := h.vectorDB.SearchSimilarWithMetadata(c, req.Queries, types.Metadata{Tags: req.Tags}, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.DataResponse{
			Status:  false,
			Message: "Search failed: " + err.Error(),
		})

		return
	}
	c.JSON(http.StatusOK, types.DataResponse{
		Status: true,
		Data:   types.SearchResponse{Documents: docs},
	})
}

func (h *SearchHandler) HandleAskAI(c *gin.Context) {

	var req types.AskAIWithRAGRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.DataResponse{
			Status:  false,
			Message: "Invalid request body",
		})
		return
	}

	// Search documents
	docs, err := h.vectorDB.AskAI(context.Background(), req.Question, req.SearchRequest.Queries, types.Metadata{Tags: req.SearchRequest.Tags}, req.SearchRequest.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.DataResponse{
			Status:  false,
			Message: "Ask AI failed: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, types.DataResponse{
		Status: true,
		Data:   types.SearchResponse{Documents: docs},
	})

}
