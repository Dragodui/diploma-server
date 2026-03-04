package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/metrics"
	"github.com/Dragodui/diploma-server/internal/models"
)

const (
	geminiTimeout = 60 * time.Second
	geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-lite:generateContent"
)

type OCRService struct {
	geminiAPIKey string
	httpClient   *http.Client
}

type IOCRService interface {
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

// ProcessFile processes a local file with Gemini Vision API
func (s *OCRService) ProcessFile(ctx context.Context, filePath, language string) (*models.OCRResult, error) {
	start := time.Now()

	imageData, err := os.ReadFile(filePath)
	if err != nil {
		metrics.OcrRequestsTotal.WithLabelValues("error").Inc()
		metrics.OcrProcessingDuration.Observe(time.Since(start).Seconds())
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	mimeType := detectMimeType(filePath)
	result, err := s.analyzeWithGemini(ctx, imageData, mimeType, language)
	duration := time.Since(start).Seconds()
	metrics.OcrProcessingDuration.Observe(duration)
	if err != nil {
		metrics.OcrRequestsTotal.WithLabelValues("error").Inc()
		return nil, fmt.Errorf("Gemini Vision failed: %w", err)
	}

	metrics.OcrRequestsTotal.WithLabelValues("success").Inc()

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
	case ".pdf":
		return "application/pdf"
	default:
		return "image/jpeg"
	}
}

