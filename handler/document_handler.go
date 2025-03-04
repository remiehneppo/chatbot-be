package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type DocumentHandler struct {
	uploadDir string
}

func NewPDFHandler(uploadDir string) *DocumentHandler {
	return &DocumentHandler{
		uploadDir: uploadDir,
	}
}

func (h *DocumentHandler) ServeDocument() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get filename from query parameter
		requestedName := r.URL.Query().Get("file")
		if requestedName == "" {
			http.Error(w, "File parameter is required", http.StatusBadRequest)
			return
		}

		// Validate filename extension
		if filepath.Ext(requestedName) != ".pdf" {
			http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
			return
		}

		// Find actual file with timestamp
		actualFile, err := h.findFileWithTimestamp(requestedName)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		filePath := filepath.Join(h.uploadDir, actualFile)

		// Set appropriate headers
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s", requestedName))

		// Stream the file to the client
		http.ServeFile(w, r, filePath)
	})
}

func (h *DocumentHandler) findFileWithTimestamp(requestedName string) (string, error) {
	files, err := os.ReadDir(h.uploadDir)
	if err != nil {
		return "", err
	}

	baseName := strings.TrimSuffix(requestedName, ".pdf")
	for _, file := range files {
		name := file.Name()
		if !strings.HasSuffix(name, ".pdf") {
			continue
		}

		nameWithoutExt := strings.TrimSuffix(name, ".pdf")
		if nameWithoutExt == baseName {
			return name, nil
		}
		// Find last underscore position
		lastUnderscoreIdx := strings.LastIndex(nameWithoutExt, "_")
		if lastUnderscoreIdx == -1 {
			continue
		}

		// Get potential timestamp part
		timestampPart := nameWithoutExt[lastUnderscoreIdx+1:]
		fileBaseName := nameWithoutExt[:lastUnderscoreIdx]

		// Validate if it's a timestamp (Unix timestamp is typically 10 or 13 digits)
		if len(timestampPart) == 10 || len(timestampPart) == 13 {
			if _, err := strconv.ParseInt(timestampPart, 10, 64); err == nil {
				if fileBaseName == baseName {
					return name, nil
				}
			}
		}
	}

	return "", fmt.Errorf("file not found: %s", requestedName)
}
