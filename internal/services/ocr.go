package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/utils"

	"github.com/google/uuid"
)

type OCRService struct{}

type IOCRService interface {
	ProcessImage(ctx context.Context, imageURL, language string) (*models.OCRResult, error)
	ProcessFile(ctx context.Context, filePath, language string) (*models.OCRResult, error)
}

func NewOCRService() *OCRService {
	return &OCRService{}
}

// ProcessImage downloads image from URL and processes it with OCR
func (s *OCRService) ProcessImage(ctx context.Context, imageURL, language string) (*models.OCRResult, error) {
	tempPath, err := downloadImage(ctx, imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer os.Remove(tempPath)

	return s.ProcessFile(ctx, tempPath, language)
}

// ProcessFile processes a local file with OCR
func (s *OCRService) ProcessFile(ctx context.Context, filePath, language string) (*models.OCRResult, error) {
	text, err := utils.ExtractTextFromImage(filePath, language)
	if err != nil {
		return nil, fmt.Errorf("OCR extraction failed: %w", err)
	}

	result := utils.ParseReceipt(text)

	return result, nil
}

// downloadImage downloads image from URL and saves to temp file
func downloadImage(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch image: status %d", resp.StatusCode)
	}

	// Determine file extension from content type
	ext := ".jpg"
	contentType := resp.Header.Get("Content-Type")
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	case "image/webp":
		ext = ".webp"
	case "image/bmp":
		ext = ".bmp"
	}

	// Create temp file
	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, fmt.Sprintf("ocr_%s%s", uuid.New().String(), ext))

	outFile, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	return tempPath, nil
}
