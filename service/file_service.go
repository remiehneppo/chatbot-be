package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tieubaoca/chatbot-be/database"
	"github.com/tieubaoca/chatbot-be/types"
)

type FileService struct {
	uploadDir  string
	vectorDB   *database.WeaviateStore
	pdfService *PDFService
}

func NewFileService(
	uploadDir string,
	vectorDB *database.WeaviateStore,
	pdfService *PDFService,
) *FileService {
	// Tạo thư mục nếu chưa tồn tại
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(err)
	}
	return &FileService{
		uploadDir:  uploadDir,
		vectorDB:   vectorDB,
		pdfService: pdfService,
	}
}

func (s *FileService) UploadFile(req types.UploadRequest, file *multipart.FileHeader, c chan<- types.ProcessingDocumentStatus) error {
	// Kiểm tra phần mở rộng file
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" && ext != ".doc" && ext != ".docx" {
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	// Mở file được upload
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Tạo tên file mới với format: originalname_timestamp.extension
	originalName := strings.TrimSuffix(req.Title, ext)
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%s_%d%s", originalName, timestamp, ext)

	// Thay thế các ký tự không hợp lệ trong tên file
	filename = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			return r
		}
		return '_'
	}, filename)

	dst, err := os.Create(filepath.Join(s.uploadDir, filename))
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy nội dung file
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	// Process PDF và lưu vào vector DB
	if ext == ".pdf" {
		chunkChan := make(chan types.DocumentChunk)
		go s.pdfService.ProcessPDF(filepath.Join(s.uploadDir, filename+ext), req, chunkChan)
		for chunk := range chunkChan {
			c <- types.ProcessingDocumentStatus{
				Status:         "processing",
				Message:        "Processing document",
				Progress:       float64(chunk.Metadata.PageNum) / float64(chunk.Metadata.TotalPages),
				TotalPages:     chunk.Metadata.TotalPages,
				ProcessedPages: chunk.Metadata.PageNum,
			}
		}
		c <- types.ProcessingDocumentStatus{
			Status:  "completed",
			Message: "Done processing PDF",
		}
		fmt.Println("Done processing PDF")
	}

	return nil
}
