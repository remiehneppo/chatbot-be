package handler

import (
	"context"
	"encoding/json"
	"net/http"

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

func (h *SearchHandler) HandleSearch() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req types.SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.sendError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Set default limit if not provided
		if req.Limit == 0 {
			req.Limit = 5
		}

		// Search documents
		docs, _, err := h.vectorDB.SearchSimilarWithMetadata(r.Context(), req.Queries, types.Metadata{Tags: req.Tags}, req.Limit)
		if err != nil {
			h.sendError(w, "Search failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Send response
		h.sendSuccess(w, docs)
	})
}

func (h *SearchHandler) HandleAskAI() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req types.AskAIWithRAGRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.sendError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Search documents
		docs, err := h.vectorDB.AskAI(context.Background(), req.Question, req.SearchRequest.Queries, types.Metadata{Tags: req.SearchRequest.Tags}, req.SearchRequest.Limit)
		if err != nil {
			h.sendError(w, "Search failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Send response
		h.sendSuccess(w, docs)
	})
}

func (h *SearchHandler) sendError(w http.ResponseWriter, message string, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "error",
		"error":  message,
	})
}

func (h *SearchHandler) sendSuccess(w http.ResponseWriter, docs []types.Document) {
	w.WriteHeader(http.StatusOK)
	searchRes := types.SearchResponse{
		Documents: docs,
	}
	resData := types.DataResponse{
		Status: "success",
		Data:   searchRes,
	}
	json.NewEncoder(w).Encode(resData)
}
