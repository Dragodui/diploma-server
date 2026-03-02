package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/models"

	"github.com/google/uuid"
)

const (
	maxImageDownloadSize = 10 * 1024 * 1024 // 10 MB
	downloadTimeout      = 30 * time.Second
	geminiTimeout        = 60 * time.Second
	geminiBaseURL        = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-lite:generateContent"
)

var (
	ErrInvalidURL = errors.New("invalid or forbidden URL")

	privateIPBlocks = []*net.IPNet{
		{IP: net.IPv4(127, 0, 0, 0), Mask: net.IPv4Mask(255, 0, 0, 0)},
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.IPv4Mask(255, 0, 0, 0)},
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.IPv4Mask(255, 240, 0, 0)},
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.IPv4Mask(255, 255, 0, 0)},
		{IP: net.IPv4(169, 254, 0, 0), Mask: net.IPv4Mask(255, 255, 0, 0)},
	}
)

type OCRService struct {
	geminiAPIKey string
	httpClient   *http.Client
}

type IOCRService interface {
	ProcessImage(ctx context.Context, imageURL, language string) (*models.OCRResult, error)
	ProcessFile(ctx context.Context, filePath, language string) (*models.OCRResult, error)
}

func NewOCRService(geminiAPIKey string) *OCRService {
	return &OCRService{
		geminiAPIKey: geminiAPIKey,
		httpClient: &http.Client{
			Timeout: geminiTimeout,
		},
	}
}

// ProcessImage downloads image from URL and processes it with Gemini Vision
func (s *OCRService) ProcessImage(ctx context.Context, imageURL, language string) (*models.OCRResult, error) {
	tempPath, err := downloadImage(ctx, imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer os.Remove(tempPath)

	return s.ProcessFile(ctx, tempPath, language)
}

// ProcessFile processes a local file with Gemini Vision API
func (s *OCRService) ProcessFile(ctx context.Context, filePath, language string) (*models.OCRResult, error) {
	imageData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	mimeType := detectMimeType(filePath)
	result, err := s.analyzeWithGemini(ctx, imageData, mimeType, language)
	if err != nil {
		return nil, fmt.Errorf("Gemini Vision failed: %w", err)
	}

	strResult, _ := json.Marshal(result)
	logger.Info.Printf("OCR Result: %s", string(strResult))

	return result, nil
}

// Gemini API request/response types
type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string        `json:"text,omitempty"`
	InlineData *geminiInline `json:"inlineData,omitempty"`
}

type geminiInline struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiGenerationConfig struct {
	ResponseMimeType string `json:"responseMimeType"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (s *OCRService) analyzeWithGemini(ctx context.Context, imageData []byte, mimeType, language string) (*models.OCRResult, error) {
	prompt := fmt.Sprintf(`Analyze this receipt/bill image. The text is likely in %s.
Extract the following data and return ONLY valid JSON (no markdown, no code fences):
{
  "vendor": "store or company name",
  "date": "date from receipt in original format",
  "total": 0.00,
  "items": [
    {"name": "item name", "quantity": 1, "price": 0.00}
  ],
  "raw_text": "all visible text from the image"
}

Rules:
- "total" must be a number (float), not a string
- "price" is the total price for that line item (quantity * unit price)
- If you cannot determine a field, use empty string for strings, 0 for numbers, [] for items
- Do NOT wrap the response in markdown code blocks`, language)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{
						InlineData: &geminiInline{
							MimeType: mimeType,
							Data:     base64.StdEncoding.EncodeToString(imageData),
						},
					},
					{
						Text: prompt,
					},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			ResponseMimeType: "application/json",
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	apiURL := fmt.Sprintf("%s?key=%s", geminiBaseURL, s.geminiAPIKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Gemini API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode Gemini response: %w", err)
	}

	if geminiResp.Error != nil {
		return nil, fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("empty response from Gemini API")
	}

	resultText := geminiResp.Candidates[0].Content.Parts[0].Text

	var result models.OCRResult
	if err := json.Unmarshal([]byte(resultText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini JSON output: %w (raw: %s)", err, resultText)
	}

	result.Confidence = calculateGeminiConfidence(&result)

	return &result, nil
}

func calculateGeminiConfidence(result *models.OCRResult) float64 {
	score := 0.0
	checks := 4.0

	if result.Vendor != "" {
		score++
	}
	if result.Date != "" {
		score++
	}
	if result.Total > 0 {
		score++
	}
	if len(result.Items) > 0 {
		score++

		// Bonus check: items total vs overall total
		checks++
		itemsTotal := 0.0
		for _, item := range result.Items {
			itemsTotal += item.Price
		}
		if result.Total > 0 && itemsTotal > 0 {
			diff := itemsTotal - result.Total
			if diff < 0 {
				diff = -diff
			}
			if diff/result.Total < 0.1 {
				score++
			}
		}
	}

	return score / checks
}

func detectMimeType(filePath string) string {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	default:
		return "image/jpeg"
	}
}

// validateImageURL validates URL to prevent SSRF attacks
func validateImageURL(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("forbidden URL scheme: %s (only http/https allowed)", scheme)
	}

	hostname := parsedURL.Hostname()
	if hostname == "" {
		return errors.New("URL must have a hostname")
	}

	lowercaseHost := strings.ToLower(hostname)
	forbiddenHosts := []string{
		"localhost",
		"127.0.0.1",
		"0.0.0.0",
		"::1",
		"[::1]",
		"169.254.169.254",
		"metadata.google.internal",
		"metadata",
	}

	for _, forbidden := range forbiddenHosts {
		if lowercaseHost == forbidden {
			return fmt.Errorf("forbidden hostname: %s", hostname)
		}
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("failed to resolve hostname: %w", err)
	}

	for _, ip := range ips {
		for _, block := range privateIPBlocks {
			if block.Contains(ip) {
				return fmt.Errorf("private IP address not allowed: %s resolves to %s", hostname, ip.String())
			}
		}
	}

	return nil
}

// downloadImage downloads image from URL and saves to temp file
func downloadImage(ctx context.Context, urlStr string) (string, error) {
	if err := validateImageURL(urlStr); err != nil {
		return "", fmt.Errorf("URL validation failed: %w", err)
	}

	client := &http.Client{
		Timeout: downloadTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("too many redirects")
			}
			if err := validateImageURL(req.URL.String()); err != nil {
				return fmt.Errorf("redirect validation failed: %w", err)
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch image: status %d", resp.StatusCode)
	}

	if resp.ContentLength > maxImageDownloadSize {
		return "", fmt.Errorf("image too large: %d bytes (max %d)", resp.ContentLength, maxImageDownloadSize)
	}

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

	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, fmt.Sprintf("ocr_%s%s", uuid.New().String(), ext))

	outFile, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer outFile.Close()

	limitedReader := io.LimitReader(resp.Body, maxImageDownloadSize+1)

	written, err := io.Copy(outFile, limitedReader)
	if err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	if written > maxImageDownloadSize {
		os.Remove(tempPath)
		return "", fmt.Errorf("image exceeds maximum size of %d bytes", maxImageDownloadSize)
	}

	return tempPath, nil
}
