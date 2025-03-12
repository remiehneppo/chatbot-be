package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	services "github.com/tieubaoca/chatbot-be/service"
	"github.com/tieubaoca/chatbot-be/types"
)

type UploadHandler struct {
	fileService *services.FileService
}

func NewUploadHandler(fileService *services.FileService) *UploadHandler {
	return &UploadHandler{
		fileService: fileService,
	}
}

func (h *UploadHandler) UploadDocumentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			h.sendError(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}
		// Giới hạn kích thước file 10MB
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			h.sendError(w, "File too large", http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			h.sendError(w, "Invalid file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// get request data
		metadata := r.FormValue("metadata")
		var req types.UploadRequest
		err = json.Unmarshal([]byte(metadata), &req)
		if err != nil {
			h.sendError(w, "Invalid metadata", http.StatusBadRequest)
			return
		}
		statusChan := make(chan types.ProcessingDocumentStatus)
		errChan := make(chan error)
		defer close(statusChan)
		defer close(errChan)
		go func() {
			errChan <- h.fileService.UploadFile(req, header, statusChan)
		}()

		for {
			select {
			case status := <-statusChan:
				jsonStatus, _ := json.Marshal(status)
				fmt.Fprintf(w, string(jsonStatus))
				flusher.Flush()
			case err := <-errChan:
				if err != nil {
					h.sendError(w, err.Error(), http.StatusInternalServerError)
				} else {
					h.sendSuccess(w, req.Title)
				}
			}
		}
	})

}

func (h *UploadHandler) sendError(w http.ResponseWriter, message string, status int) {
	w.WriteHeader(status)

	res := types.DataResponse{
		Status:  false,
		Message: message,
	}
	json.NewEncoder(w).Encode(res)
}

func (h *UploadHandler) sendSuccess(w http.ResponseWriter, originalName string) {
	w.WriteHeader(http.StatusOK)
	res := types.DataResponse{
		Status: true,
		Data: types.UploadResponse{
			OriginalName: originalName,
		},
	}
	json.NewEncoder(w).Encode(res)
}
