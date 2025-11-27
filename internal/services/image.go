package services

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type ImageService struct {
	s3Client   *s3.Client
	uploader   *manager.Uploader
	bucketName string
	region     string
}

type IImageService interface {
	Upload(file multipart.File, header *multipart.FileHeader) (string, error)
	GetPresignedURL(key string, expiration time.Duration) (string, error)
	Delete(imageURL string) error
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
func (s *ImageService) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ext := filepath.Ext(header.Filename)
	newName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	result, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(newName),
		Body:        file,
		ContentType: aws.String(contentType),
		Metadata: map[string]string{
			"original-filename": header.Filename,
			"uploaded-at":       time.Now().Format(time.RFC3339),
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	publicURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		s.bucketName, s.region, newName)

	log.Printf("Uploaded to S3: %s (Size: %d bytes, ETag: %s)",
		newName, header.Size, *result.ETag)

	return publicURL, nil
}

func (s *ImageService) Delete(imageURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

func (s *ImageService) GetPresignedURL(key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.s3Client)

	request, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
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
