package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CopyFileWithTimestamp copies a file to the destination directory with a timestamp suffix
// Returns the destination path and error if any
func CopyFileWithTimestamp(sourcePath, uploadDir string) (string, error) {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %v", err)
	}
	defer sourceFile.Close()

	// Create destination filename with timestamp
	originalName := filepath.Base(sourcePath)
	ext := filepath.Ext(originalName)
	baseFileName := strings.TrimSuffix(originalName, ext)
	timestamp := time.Now().Unix()
	destFileName := fmt.Sprintf("%s_%d%s", baseFileName, timestamp, ext)
	destPath := filepath.Join(uploadDir, destFileName)

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destFile.Close()

	// Copy the file
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return "", fmt.Errorf("failed to copy file: %v", err)
	}

	return destPath, nil
}
