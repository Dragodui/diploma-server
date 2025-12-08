package handlers_test

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/stretchr/testify/require"
)

// Mock image service
type mockImageService struct {
	UploadFunc          func(file multipart.File, header *multipart.FileHeader) (string, error)
	GetPresignedURLFunc func(key string, expiration time.Duration) (string, error)
	DeleteFunc          func(imageURL string) error
}

func (m *mockImageService) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	if m.UploadFunc != nil {
		return m.UploadFunc(file, header)
	}
	return "", nil
}

func (m *mockImageService) GetPresignedURL(key string, expiration time.Duration) (string, error) {
	if m.GetPresignedURLFunc != nil {
		return m.GetPresignedURLFunc(key, expiration)
	}
	return "", nil
}

func (m *mockImageService) Delete(imageURL string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(imageURL)
	}
	return nil
}

func setupImageHandler(svc *mockImageService) *handlers.ImageHandler {
	return handlers.NewImageHandler(svc)
}

func createImageUploadRequest(fieldName, fileName, content string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if fileName != "" {
		part, err := writer.CreateFormFile(fieldName, fileName)
		if err != nil {
			return nil, err
		}
		_, err = io.WriteString(part, content)
		if err != nil {
			return nil, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func TestImageHandler_UploadImage(t *testing.T) {
	tests := []struct {
		name           string
		hasFile        bool
		fieldName      string
		fileName       string
		uploadFunc     func(file multipart.File, header *multipart.FileHeader) (string, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "Success",
			hasFile:   true,
			fieldName: "image",
			fileName:  "test.jpg",
			uploadFunc: func(file multipart.File, header *multipart.FileHeader) (string, error) {
				require.Equal(t, "test.jpg", header.Filename)
				return "https://s3.amazonaws.com/bucket/test.jpg", nil
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "File uploaded successfully",
		},
		{
			name:           "Missing File",
			hasFile:        false,
			fieldName:      "image",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "File required",
		},
		{
			name:      "Wrong Field Name",
			hasFile:   true,
			fieldName: "wrong_field",
			fileName:  "test.jpg",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "File required",
		},
		{
			name:      "Upload Service Error",
			hasFile:   true,
			fieldName: "image",
			fileName:  "test.jpg",
			uploadFunc: func(file multipart.File, header *multipart.FileHeader) (string, error) {
				return "", errors.New("S3 upload failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "S3 upload failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockImageService{
				UploadFunc: tt.uploadFunc,
			}

			h := setupImageHandler(svc)

			var req *http.Request
			var err error

			if tt.hasFile {
				req, err = createImageUploadRequest(tt.fieldName, tt.fileName, "fake image content")
			} else {
				// Create empty multipart request
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()
				req = httptest.NewRequest(http.MethodPost, "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
			}
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			h.UploadImage(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}
