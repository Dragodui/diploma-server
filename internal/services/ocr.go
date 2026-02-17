package services

import (
	"context"
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

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/utils"

	"github.com/google/uuid"
)

const (
	maxImageDownloadSize = 10 * 1024 * 1024 // 10 MB
	downloadTimeout      = 30 * time.Second
)

var (
	// ErrInvalidURL is returned when URL validation fails
	ErrInvalidURL = errors.New("invalid or forbidden URL")

	// Blocked private IP ranges (RFC 1918, loopback, link-local)
	privateIPBlocks = []*net.IPNet{
		// Loopback
		{IP: net.IPv4(127, 0, 0, 0), Mask: net.IPv4Mask(255, 0, 0, 0)},
		// Private networks
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.IPv4Mask(255, 0, 0, 0)},
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.IPv4Mask(255, 240, 0, 0)},
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.IPv4Mask(255, 255, 0, 0)},
		// Link-local
		{IP: net.IPv4(169, 254, 0, 0), Mask: net.IPv4Mask(255, 255, 0, 0)},
	}
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

// validateImageURL validates URL to prevent SSRF attacks
// Blocks: file://, localhost, private IPs, AWS metadata, etc.
func validateImageURL(urlStr string) error {
	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow HTTP and HTTPS schemes
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("forbidden URL scheme: %s (only http/https allowed)", scheme)
	}

	// Get hostname
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return errors.New("URL must have a hostname")
	}

	// Block localhost variations
	lowercaseHost := strings.ToLower(hostname)
	forbiddenHosts := []string{
		"localhost",
		"127.0.0.1",
		"0.0.0.0",
		"::1",
		"[::1]",
		// AWS metadata endpoints
		"169.254.169.254",
		"metadata.google.internal",
		"metadata",
	}

	for _, forbidden := range forbiddenHosts {
		if lowercaseHost == forbidden {
			return fmt.Errorf("forbidden hostname: %s", hostname)
		}
	}

	// Resolve hostname to IP and check for private ranges
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("failed to resolve hostname: %w", err)
	}

	for _, ip := range ips {
		// Check if IP is in private range
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
	// SECURITY: Validate URL to prevent SSRF attacks
	if err := validateImageURL(urlStr); err != nil {
		return "", fmt.Errorf("URL validation failed: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: downloadTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Limit redirects to prevent infinite loops
			if len(via) >= 10 {
				return errors.New("too many redirects")
			}
			// Validate redirect URL as well
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

	// Check Content-Length header to prevent huge downloads
	if resp.ContentLength > maxImageDownloadSize {
		return "", fmt.Errorf("image too large: %d bytes (max %d)", resp.ContentLength, maxImageDownloadSize)
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

	// SECURITY: Limit download size to prevent DoS attacks
	// Use LimitReader to enforce max size even if Content-Length is missing/wrong
	limitedReader := io.LimitReader(resp.Body, maxImageDownloadSize+1)

	written, err := io.Copy(outFile, limitedReader)
	if err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	// Check if we hit the limit (downloaded more than max)
	if written > maxImageDownloadSize {
		os.Remove(tempPath)
		return "", fmt.Errorf("image exceeds maximum size of %d bytes", maxImageDownloadSize)
	}

	return tempPath, nil
}
