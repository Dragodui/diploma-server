package handlers_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock user service
type mockUserService struct {
	GetUserByIDFunc      func(ctx context.Context, userID int) (*models.User, error)
	UpdateUserFunc       func(ctx context.Context, userID int, name string) error
	UpdateUserAvatarFunc func(ctx context.Context, userID int, imagePath string) error
}

func (m *mockUserService) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockUserService) UpdateUser(ctx context.Context, userID int, name string) error {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, userID, name)
	}
	return nil
}

func (m *mockUserService) UpdateUserAvatar(ctx context.Context, userID int, imagePath string) error {
	if m.UpdateUserAvatarFunc != nil {
		return m.UpdateUserAvatarFunc(ctx, userID, imagePath)
	}
	return nil
}

// Mock image service for user handler
type mockImageServiceForUser struct {
	UploadFunc func(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
}

func (m *mockImageServiceForUser) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, file, header)
	}
	return "", nil
}

func (m *mockImageServiceForUser) GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	return "", nil
}

func (m *mockImageServiceForUser) Delete(ctx context.Context, imageURL string) error {
	return nil
}

func setupUserHandler(userSvc *mockUserService, imgSvc *mockImageServiceForUser) *handlers.UserHandler {
	return handlers.NewUserHandler(userSvc, imgSvc)
}

func createMultipartFormRequest(fieldName, fileName, content string, extraFields map[string]string) (*http.Request, error) {
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

	for key, val := range extraFields {
		_ = writer.WriteField(key, val)
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	req := httptest.NewRequest(http.MethodPatch, "/user", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func TestUserHandler_GetMe(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		mockFunc       func(ctx context.Context, userID int) (*models.User, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userID: 123,
			mockFunc: func(ctx context.Context, userID int) (*models.User, error) {
				require.Equal(t, 123, userID)
				return &models.User{ID: 123, Name: "Test User", Email: "test@example.com"}, nil
			},
			expectedStatus: http.StatusAccepted,
			expectedBody:   "Test User",
		},
		{
			name:           "Unauthorized",
			userID:         0,
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:   "User Not Found",
			userID: 123,
			mockFunc: func(ctx context.Context, userID int) (*models.User, error) {
				return nil, nil
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "User not found",
		},
		{
			name:   "Service Error",
			userID: 123,
			mockFunc: func(ctx context.Context, userID int) (*models.User, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userSvc := &mockUserService{
				GetUserByIDFunc: tt.mockFunc,
			}

			h := setupUserHandler(userSvc, nil)

			req := httptest.NewRequest(http.MethodPost, "/user", nil)
			if tt.userID != 0 {
				req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))
			}

			rr := httptest.NewRecorder()
			h.GetMe(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestUserHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		formFields     map[string]string
		hasAvatar      bool
		updateNameFunc func(ctx context.Context, userID int, name string) error
		uploadFunc     func(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
		updateAvatar   func(ctx context.Context, userID int, imagePath string) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success - Update Name Only",
			userID: 123,
			formFields: map[string]string{
				"name": "New Name",
			},
			hasAvatar: false,
			updateNameFunc: func(ctx context.Context, userID int, name string) error {
				assert.Equal(t, 123, userID)
				assert.Equal(t, "New Name", name)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "User updated successfully",
		},
		{
			name:       "Success - Update Avatar Only",
			userID:     123,
			formFields: map[string]string{},
			hasAvatar:  true,
			uploadFunc: func(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
				return "https://s3.amazonaws.com/bucket/avatar.jpg", nil
			},
			updateAvatar: func(ctx context.Context, userID int, imagePath string) error {
				assert.Equal(t, 123, userID)
				assert.Equal(t, "https://s3.amazonaws.com/bucket/avatar.jpg", imagePath)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "User updated successfully",
		},
		{
			name:           "Unauthorized",
			userID:         0,
			formFields:     map[string]string{"name": "Test"},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:           "No Fields to Update",
			userID:         123,
			formFields:     map[string]string{},
			hasAvatar:      false,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "No fields to update",
		},
		{
			name:   "Update Name Failed",
			userID: 123,
			formFields: map[string]string{
				"name": "New Name",
			},
			updateNameFunc: func(ctx context.Context, userID int, name string) error {
				return errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to update name",
		},
		{
			name:       "Upload Avatar Failed",
			userID:     123,
			formFields: map[string]string{},
			hasAvatar:  true,
			uploadFunc: func(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
				return "", errors.New("upload error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to upload avatar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userSvc := &mockUserService{
				UpdateUserFunc:       tt.updateNameFunc,
				UpdateUserAvatarFunc: tt.updateAvatar,
			}
			imgSvc := &mockImageServiceForUser{
				UploadFunc: tt.uploadFunc,
			}

			h := setupUserHandler(userSvc, imgSvc)

			var req *http.Request
			var err error

			if tt.hasAvatar {
				req, err = createMultipartFormRequest("avatar_file", "avatar.jpg", "fake image content", tt.formFields)
				require.NoError(t, err)
			} else {
				req, err = createMultipartFormRequest("", "", "", tt.formFields)
				require.NoError(t, err)
			}

			if tt.userID != 0 {
				req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))
			}

			rr := httptest.NewRecorder()
			h.Update(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

