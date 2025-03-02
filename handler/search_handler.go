package handler

import (
	"encoding/json"
	"net/http"

	"github.com/tieubaoca/chatbot-be/database"
	"github.com/tieubaoca/chatbot-be/types"
)

type SearchHandler struct {
	vectorDB *database.WeaviateStore
}

type SearchRequest struct {
	Queries []string `json:"queries"`
	Tags    []string `json:"tags,omitempty"`
	Limit   int      `json:"limit,omitempty"`
}

type SearchResponse struct {
	Documents []database.Document `json:"documents"`
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

		var req SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.sendError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Set default limit if not provided
		if req.Limit == 0 {
			req.Limit = 5
		}

		// Search documents
		docs, _, err := h.vectorDB.SearchSimilar(r.Context(), req.Queries, req.Limit)
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

func (h *SearchHandler) sendSuccess(w http.ResponseWriter, docs []database.Document) {
	w.WriteHeader(http.StatusOK)
	searchRes := SearchResponse{
		Documents: docs,
	}
	resData := types.DataResponse{
		Status: "success",
		Data:   searchRes,
	}
	json.NewEncoder(w).Encode(resData)
}
