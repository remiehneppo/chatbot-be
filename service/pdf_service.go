package service

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/tieubaoca/chatbot-be/types"
)

var tesseditCharWhitelist = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789.,?!()[]{}<>:;\"'"

// PDFService handles PDF processing operations
type PDFService struct {
	maxChunkSize int // Maximum size of each text chunk
	overlapSize  int // Size of overlap between chunks
}

var DefaultDocumentServiceConfig = types.DocumentServiceConfig{
	MaxChunkSize: 1024,
	OverlapSize:  128,
}

// PDFChunk represents a processed chunk of PDF text with metadata

// NewPDFService creates a new PDF service with configurable chunk sizes
func NewPDFService(config types.DocumentServiceConfig) *PDFService {

	return &PDFService{
		maxChunkSize: config.MaxChunkSize,
		overlapSize:  config.OverlapSize,
	}
}

// ProcessPDF reads and chunks a PDF file
// Parameters:
//   - filePath: Path to the PDF file
//   - c: Channel to send processed chunks
//
// Returns:
//   - error: Error if processing fails
func (s *PDFService) ProcessPDF(filePath string, req types.UploadRequest, c chan<- types.DocumentChunk) error {
	defer close(c)
	// Get total pages
	totalPages, err := getNumPages(filePath)
	if err != nil {
		return err
	}
	log.Println("Total pages: ", totalPages)
	lastText := ""
	// Process each page
	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		// Extract text from current page
		text, err := s.extractText(filePath, pageNum)
		if err != nil {
			log.Printf("Warning: failed to extract text from page %d: %v", pageNum, err)
			continue // Skip failed pages instead of returning error
		}

		// Clean text
		text = lastText + " " + s.cleanText(text)

		// Skip empty text
		if text == "" {
			continue
		}

		// Create metadata for this page
		metadata := types.DocumentMetadata{
			Source:     req.Source,
			Title:      req.Title + ".pdf",
			PageNum:    pageNum,
			TotalPages: totalPages,
		}
		// Create chunks for this page
		pageChunks, newLastText := s.createChunks(text, metadata)
		if len(pageChunks) == 0 {
			lastText = newLastText
			continue
		}
		if len(newLastText) > 0 {
			lastText = newLastText
			for i := 0; i < len(pageChunks)-1; i++ {
				c <- pageChunks[i]
			}
		} else {
			for _, chunk := range pageChunks {
				c <- chunk
			}
		}

	}

	return nil
}

// getFileNameWithoutExt extracts filename without extension from a file path
func GetFileNameWithoutExt(filepath string) string {
	// Get base filename from path
	base := filepath[strings.LastIndex(filepath, "/")+1:]

	// Remove extension
	if idx := strings.LastIndex(base, "."); idx != -1 {
		base = base[:idx]
	}

	return base
}

// extractText attempts to extract text from a specific page using multiple methods
func (s *PDFService) extractText(filePath string, pageNumber int) (string, error) {
	text, err := s.extractTextWithPdftotext(filePath, pageNumber)
	if err != nil || text == "" {
		text, err = s.extractTextWithTesseract(filePath, pageNumber)
		if err != nil {
			return "", fmt.Errorf("failed to extract text: %w", err)
		}
	}
	return text, nil
}

// createChunks splits text into overlapping chunks with proper sentence boundaries
func (s *PDFService) createChunks(text string, metadata types.DocumentMetadata) ([]types.DocumentChunk, string) {
	var chunks []types.DocumentChunk
	textLen := len(text)
	lastText := ""
	// Return early if text fits in one chunk
	if textLen <= s.maxChunkSize {
		lastText = text
		return []types.DocumentChunk{
			{
				Content:  text,
				Page:     metadata.PageNum,
				Metadata: metadata,
			},
		}, lastText
	}

	currentPos := 0
	for currentPos < textLen {
		// Calculate end position for current chunk
		chunkEnd := currentPos + s.maxChunkSize
		if chunkEnd >= textLen {
			// Handle last chunk
			chunk := strings.TrimSpace(text[currentPos:])
			if chunk != "" {
				chunks = append(chunks, types.DocumentChunk{
					Content:  chunk,
					Page:     metadata.PageNum,
					Metadata: metadata,
				})
				lastText = chunk
			}
			break
		}

		// Find nearest sentence end
		sentenceEnd := chunkEnd
		for i := chunkEnd; i > currentPos; i-- {
			if text[i] == '.' || text[i] == '?' || text[i] == '!' {
				sentenceEnd = i + 1
				break
			}
		}

		// If no sentence end found, use word boundary
		if sentenceEnd == chunkEnd {
			for i := chunkEnd; i > currentPos; i-- {
				if text[i] == ' ' {
					sentenceEnd = i
					break
				}
			}
		}

		chunk := strings.TrimSpace(text[currentPos:sentenceEnd])
		if chunk != "" {
			chunks = append(chunks, types.DocumentChunk{
				Content:  chunk,
				Page:     metadata.PageNum,
				Metadata: metadata,
			})
		}

		// Update position for next chunk
		currentPos = sentenceEnd - s.overlapSize
		if currentPos < 0 {
			currentPos = 0
		}
		// Ensure we make progress
		previousPos := currentPos
		if currentPos <= previousPos {
			currentPos = sentenceEnd
		}
	}

	return chunks, lastText
}

// extractTextWithPdftotext extracts text using pdftotext utility
// Parameters:
//   - filepath: Path to the PDF file
//   - pageNumber: Page number to extract text from
//
// Returns:
//   - string: Extracted text
//   - error: Error if extraction fails
func (s *PDFService) extractTextWithPdftotext(filepath string, pageNumber int) (string, error) {
	log.Println("Try extracting with pdftotext")
	pdftotextCmd := exec.Command("pdftotext", "-f", strconv.Itoa(pageNumber),
		"-l", strconv.Itoa(pageNumber),
		"-enc", "UTF-8", "-nopgbrk",
		filepath, "-")
	var txtOut bytes.Buffer
	pdftotextCmd.Stdout = &txtOut

	if err := pdftotextCmd.Run(); err != nil {
		log.Printf("Error executing pdftotext command for page %d: %v", pageNumber, err)
	}
	pageText := txtOut.String()
	if trimmed := strings.TrimSpace(pageText); len(trimmed) > 0 {
		return trimmed, nil
	} else {
		return "", fmt.Errorf("got nothing at page %d", pageNumber)
	}
}

// extractTextWithTesseract extracts text using OCR when pdftotext fails
// Parameters:
//   - pdfPath: Path to the PDF file
//   - pageNumber: Page number to extract text from
//
// Returns:
//   - string: Extracted text
//   - error: Error if extraction fails
func (s *PDFService) extractTextWithTesseract(pdfPath string, pageNumber int) (string, error) {
	log.Println("Try extracting with tesseract")
	//check if temp directory exists
	if _, err := os.Stat("temp"); os.IsNotExist(err) {
		os.Mkdir("temp", os.ModePerm)
	}
	tempFolder := filepath.Join("temp", GetFileNameWithoutExt(pdfPath))
	if _, err := os.Stat(tempFolder); err == nil {
		os.RemoveAll(tempFolder)
	}
	err := os.Mkdir(tempFolder, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempFolder)

	convertCmd := exec.Command("pdftoppm", "-f", strconv.Itoa(pageNumber), "-l", strconv.Itoa(pageNumber), "-png", pdfPath, filepath.Join(tempFolder, "page"))
	if err := convertCmd.Run(); err != nil {
		log.Fatalf("Error converting page %d to image: %v", pageNumber, err)
	}
	pattern := filepath.Join(tempFolder, "page-*.png")
	file, err := filepath.Glob(pattern)
	if err != nil || len(file) == 0 {
		return "", fmt.Errorf("failed to read image files: %w", err)
	}
	imageFile := file[0]
	ocrCmd := exec.Command("tesseract",
		imageFile,
		"stdout",
		"-l", "vie+rus+eng", // Add both language packs
		"--oem", "3", // Use LSTM OCR Engine Mode
		"--psm", "3",
	// "c", fmt.Sprintf("tessedit_char_whitelist=%s", tesseditCharWhitelist),
	) // Auto-detect page segmentation mode
	var ocrOut bytes.Buffer
	ocrCmd.Stdout = &ocrOut
	if err := ocrCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run tesseract: %w", err)
	}
	ocrText := ocrOut.String()
	if trimmed := strings.TrimSpace(ocrText); len(trimmed) > 0 {
		return trimmed, nil
	} else {
		return "", fmt.Errorf("got nothing at page %d", pageNumber)
	}
}

// getNumPages uses pdfinfo to get the total number of pages in a PDF file
// Parameters:
//   - pdfPath: Path to the PDF file
//
// Returns:
//   - int: Number of pages
//   - error: Error if page count cannot be determined
func getNumPages(pdfPath string) (int, error) {
	cmd := exec.Command("pdfinfo", pdfPath)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("error running pdfinfo: %v", err)
	}

	scanner := bufio.NewScanner(&out)
	re := regexp.MustCompile(`Pages:\s+(\d+)`)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := re.FindStringSubmatch(line); len(matches) == 2 {
			return strconv.Atoi(matches[1])
		}
	}

	return 0, fmt.Errorf("unable to determine page count from pdfinfo")
}

func (s *PDFService) cleanText(text string) string {

	replacements := map[string]string{
		"\u0000": "",   // Null character
		"\ufffd": "",   // Unicode replacement character
		"\u001b": "",   // Escape character
		"\r":     "",   // Carriage return
		"\f":     "\n", // Form feed to newline
		"  ":     " ",  // Multiple spaces to single space
		"":      "",   // Apple logo
		"‡":      "",   // Double dagger
		"†":      "",   // Dagger
	}
	// Apply replacements
	cleaned := text
	for old, new := range replacements {
		cleaned = strings.ReplaceAll(cleaned, old, new)
	}

	// Trim leading/trailing whitespace
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}
