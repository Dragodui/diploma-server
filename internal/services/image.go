package services

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type ImageService struct {
}

type IImageService interface {
	Upload(file multipart.File, header *multipart.FileHeader) (string, error)
}

func NewImageService() *ImageService {
	return &ImageService{}
}

func (s *ImageService) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	uploadDir := "./uploads"
	os.MkdirAll(uploadDir, os.ModePerm)

	ext := filepath.Ext(header.Filename)
	newName := uuid.New().String() + ext

	filePath := uploadDir + newName
	out, err := os.Create(filePath)
	if err != nil {

		return "", err

	}
	defer out.Close()
	io.Copy(out, file)

	publicPath := "/uploads/" + newName

	return publicPath, nil
}
