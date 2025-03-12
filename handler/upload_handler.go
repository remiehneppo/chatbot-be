package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
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

func (h *UploadHandler) UploadDocumentHandler(c *gin.Context) {

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, types.DataResponse{
			Status:  false,
			Message: "Invalid file",
		})
		return
	}
	defer file.Close()

	// Get metadata from form
	metadata := c.Request.FormValue("metadata")
	var req types.UploadRequest
	if err := json.Unmarshal([]byte(metadata), &req); err != nil {
		c.JSON(http.StatusBadRequest, types.DataResponse{
			Status:  false,
			Message: "Invalid metadata",
		})
		return
	}

	const maxSize = 10 << 20
	if header.Size > maxSize {
		c.JSON(http.StatusBadRequest, types.DataResponse{
			Status:  false,
			Message: "File too large",
		})
		return
	}

	statusChan := make(chan types.ProcessingDocumentStatus)
	errChan := make(chan error)
	defer close(statusChan)
	defer close(errChan)
	go func() {
		errChan <- h.fileService.UploadFile(req, header, statusChan)
	}()
	// Create a channel to detect client disconnect
	clientGone := c.Writer.CloseNotify()
	for {
		select {
		case <-clientGone:
			return // Client disconnected
		case status := <-statusChan:
			jsonStatus, err := json.Marshal(status)
			if err != nil {
				continue
			}
			c.SSEvent("message", string(jsonStatus))
			c.Writer.Flush()
		case err := <-errChan:
			if err != nil {
				c.JSON(http.StatusInternalServerError, types.DataResponse{
					Status:  false,
					Message: err.Error(),
				})
			} else {
				c.JSON(http.StatusOK, types.DataResponse{
					Status: true,
					Data: types.UploadResponse{
						OriginalName: req.Title,
					},
				})
			}
			return
		}
	}

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
