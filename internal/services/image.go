package services

import (
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

const (
	maxFileSize       = 10 * 1024 * 1024 // 10 MB
	maxImageWidth     = 10000            // max width in pixels
	maxImageHeight    = 10000            // max height in pixels
)

var (
	forbiddenExtensions = []string{".php", ".exe", ".sh", ".bat", ".cmd", ".com", ".pif", ".scr", ".vbs", ".js"}
)

type ImageService struct {
	s3Client   *s3.Client
	uploader   *manager.Uploader
	bucketName string
	region     string
}

type IImageService interface {
	Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
	GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error)
	Delete(ctx context.Context, imageURL string) error
}

func NewImageService(bucketName, region string) (*ImageService, error) {
	if bucketName == "" || region == "" {
		return nil, fmt.Errorf("bucket name and region are required")
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(region))

	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

	log.Printf("S3 configured: bucket=%s, region=%s", bucketName, region)

	return &ImageService{
		s3Client:   client,
		uploader:   uploader,
		bucketName: bucketName,
		region:     region,
	}, nil
}
func (s *ImageService) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	// Check file size
	if header.Size > maxFileSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxFileSize)
	}

	if header.Size == 0 {
		return "", errors.New("file is empty")
	}

	// Get and normalize extension
	ext := strings.ToLower(filepath.Ext(header.Filename))

	// Check for forbidden extensions (case-insensitive)
	for _, forbiddenExt := range forbiddenExtensions {
		if ext == forbiddenExt {
			return "", errors.New("forbidden file extension")
		}
	}

	// Check for double extensions (e.g., .jpg.php)
	baseFilename := strings.TrimSuffix(header.Filename, ext)
	if strings.Contains(baseFilename, ".") {
		secondExt := strings.ToLower(filepath.Ext(baseFilename))
		for _, forbiddenExt := range forbiddenExtensions {
			if secondExt == forbiddenExt {
				return "", errors.New("forbidden double extension detected")
			}
		}
	}

	// Validate image file content and get detected MIME type
	detectedContentType, width, height, err := validateImageFile(file)
	if err != nil {
		return "", err
	}

	// Check image dimensions
	if width > maxImageWidth || height > maxImageHeight {
		return "", fmt.Errorf("image dimensions (%dx%d) exceed maximum allowed (%dx%d)",
			width, height, maxImageWidth, maxImageHeight)
	}

	// Verify extension matches detected content type
	expectedExt, ok := allowedTypes[detectedContentType]
	if !ok {
		return "", fmt.Errorf("unsupported content type: %s", detectedContentType)
	}

	// Allow .jpg for .jpeg and vice versa
	if !(ext == expectedExt ||
		(ext == ".jpg" && expectedExt == ".jpeg") ||
		(ext == ".jpeg" && expectedExt == ".jpg")) {
		return "", fmt.Errorf("file extension %s does not match detected content type %s", ext, detectedContentType)
	}

	// Generate new filename with validated extension
	newName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Use detected content type instead of client-provided header
	result, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(newName),
		Body:        file,
		ContentType: aws.String(detectedContentType),
		Metadata: map[string]string{
			"uploaded-at": time.Now().Format(time.RFC3339),
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	publicURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		s.bucketName, s.region, newName)

	log.Printf("Uploaded to S3: %s (Size: %d bytes, Dimensions: %dx%d, Type: %s, ETag: %s)",
		newName, header.Size, width, height, detectedContentType, *result.ETag)

	return publicURL, nil
}

func (s *ImageService) Delete(ctx context.Context, imageURL string) error {
	key := filepath.Base(imageURL)

	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	log.Printf("Deleted from S3: %s", key)
	return nil
}

func (s *ImageService) GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.s3Client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

var allowedTypes = map[string]string{
	"image/jpeg": ".jpeg",
	"image/png":  ".png",
	"image/gif":  ".gif",
}

// validateImageFile validates the image file and returns content type, width, height, and error
func validateImageFile(file multipart.File) (contentType string, width int, height int, err error) {
	// Read first 512 bytes to detect MIME type
	buffer := make([]byte, 512)
	n, readErr := file.Read(buffer)
	if readErr != nil {
		return "", 0, 0, fmt.Errorf("failed to read file: %w", readErr)
	}

	// Reset file pointer to beginning
	if _, seekErr := file.Seek(0, 0); seekErr != nil {
		return "", 0, 0, fmt.Errorf("failed to reset file pointer: %w", seekErr)
	}

	// Detect content type from file signature
	detectedType := http.DetectContentType(buffer[:n])

	// Check if type is allowed
	_, allowed := allowedTypes[detectedType]
	if !allowed {
		return "", 0, 0, fmt.Errorf("unsupported file type: %s", detectedType)
	}

	// Decode image to verify it's a valid image and get dimensions
	img, format, decodeErr := image.DecodeConfig(file)
	if decodeErr != nil {
		return "", 0, 0, fmt.Errorf("invalid or corrupted image file: %w", decodeErr)
	}

	// Reset file pointer again for S3 upload
	if _, seekErr := file.Seek(0, 0); seekErr != nil {
		return "", 0, 0, fmt.Errorf("failed to reset file pointer after decode: %w", seekErr)
	}

	// Verify format matches detected type
	expectedFormat := map[string]string{
		"image/jpeg": "jpeg",
		"image/png":  "png",
		"image/gif":  "gif",
	}

	if expectedFormat[detectedType] != format {
		return "", 0, 0, fmt.Errorf("format mismatch: detected %s but decoded as %s", detectedType, format)
	}

	return detectedType, img.Width, img.Height, nil
}
