package middleware_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dragodui/diploma-server/internal/http/middleware"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/Dragodui/diploma-server/pkg/security"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock home repository
type mockHomeRepo struct {
	IsMemberFunc func(homeID, userID int) (bool, error)
	IsAdminFunc  func(homeID, userID int) (bool, error)
}

func (m *mockHomeRepo) Create(h *models.Home) error                              { return nil }
func (m *mockHomeRepo) FindByID(id int) (*models.Home, error)                    { return nil, nil }
func (m *mockHomeRepo) FindByInviteCode(inviteCode string) (*models.Home, error) { return nil, nil }
func (m *mockHomeRepo) Delete(id int) error                                      { return nil }
func (m *mockHomeRepo) AddMember(id int, userID int, role string) error          { return nil }
func (m *mockHomeRepo) DeleteMember(id int, userID int) error                    { return nil }
func (m *mockHomeRepo) GenerateUniqueInviteCode() (string, error)                { return "CODE1234", nil }
func (m *mockHomeRepo) GetUserHome(userID int) (*models.Home, error)             { return nil, nil }
func (m *mockHomeRepo) RegenerateCode(code string, id int) error                 { return nil }

func (m *mockHomeRepo) IsMember(homeID, userID int) (bool, error) {
	if m.IsMemberFunc != nil {
		return m.IsMemberFunc(homeID, userID)
	}
	return false, nil
}

func (m *mockHomeRepo) IsAdmin(homeID, userID int) (bool, error) {
	if m.IsAdminFunc != nil {
		return m.IsAdminFunc(homeID, userID)
	}
	return false, nil
}

var testJWTSecret = []byte("test-secret-key-for-testing-purposes")

func TestJWTAuth(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
		shouldCallNext bool
	}{
		{
			name:           "Valid Token",
			authHeader:     "", // Will be set dynamically
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			shouldCallNext: true,
		},
		{
			name:           "Missing Token",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "missing token",
			shouldCallNext: false,
		},
		{
			name:           "Invalid Token Format",
			authHeader:     "InvalidFormat token123",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "missing token",
			shouldCallNext: false,
		},
		{
			name:           "Invalid Token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid token",
			shouldCallNext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				userID := middleware.GetUserID(r)
				if userID > 0 {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"message":"success"}`))
				}
			})

			handler := middleware.JWTAuth(testJWTSecret)(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)

			if tt.name == "Valid Token" {
				token, err := security.GenerateToken(123, "test@example.com", testJWTSecret, time.Hour)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+token)
			} else if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
			assert.Equal(t, tt.shouldCallNext, nextCalled)
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		expectedID int
	}{
		{
			name:       "Valid User ID",
			userID:     123,
			expectedID: 123,
		},
		{
			name:       "No User ID in Context",
			userID:     0,
			expectedID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.userID != 0 {
				req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))
			}

			userID := middleware.GetUserID(req)
			assert.Equal(t, tt.expectedID, userID)
		})
	}
}

func TestRequireMember(t *testing.T) {
	tests := []struct {
		name           string
		homeID         string
		userID         int
		isMember       bool
		useBody        bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Is Member - URL Param",
			homeID:         "1",
			userID:         123,
			isMember:       true,
			useBody:        false,
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "Not Member - URL Param",
			homeID:         "1",
			userID:         123,
			isMember:       false,
			useBody:        false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "you are not a member",
		},
		{
			name:           "Is Member - Body",
			homeID:         "",
			userID:         123,
			isMember:       true,
			useBody:        true,
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "Not Member - Body",
			homeID:         "",
			userID:         123,
			isMember:       false,
			useBody:        true,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "you are not a member",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockHomeRepo{
				IsMemberFunc: func(homeID, userID int) (bool, error) {
					return tt.isMember, nil
				},
			}

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"message":"success"}`))
			})

			r := chi.NewRouter()

			if tt.homeID != "" {
				r.With(middleware.RequireMember(mockRepo)).Get("/homes/{home_id}", nextHandler)
			} else {
				r.With(middleware.RequireMember(mockRepo)).Post("/action", nextHandler)
			}

			var req *http.Request
			if tt.useBody {
				body := map[string]int{"home_id": 1}
				jsonBody, _ := json.Marshal(body)
				req = httptest.NewRequest(http.MethodPost, "/action", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(http.MethodGet, "/homes/"+tt.homeID, nil)
			}

			req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	tests := []struct {
		name           string
		homeID         string
		userID         int
		isMember       bool
		isAdmin        bool
		useBody        bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Is Admin - URL Param (checks member)",
			homeID:         "1",
			userID:         123,
			isMember:       true,
			isAdmin:        true,
			useBody:        false,
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "Not Member - URL Param",
			homeID:         "1",
			userID:         123,
			isMember:       false,
			isAdmin:        false,
			useBody:        false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "you are not a member",
		},
		{
			name:           "Is Admin - Body",
			homeID:         "",
			userID:         123,
			isMember:       true,
			isAdmin:        true,
			useBody:        true,
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "Not Admin - Body",
			homeID:         "",
			userID:         123,
			isMember:       true,
			isAdmin:        false,
			useBody:        true,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "you are not an admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockHomeRepo{
				IsMemberFunc: func(homeID, userID int) (bool, error) {
					return tt.isMember, nil
				},
				IsAdminFunc: func(homeID, userID int) (bool, error) {
					return tt.isAdmin, nil
				},
			}

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"message":"success"}`))
			})

			r := chi.NewRouter()

			if tt.homeID != "" {
				r.With(middleware.RequireAdmin(mockRepo)).Get("/homes/{home_id}/admin", nextHandler)
			} else {
				r.With(middleware.RequireAdmin(mockRepo)).Post("/admin-action", nextHandler)
			}

			var req *http.Request
			if tt.useBody {
				body := map[string]int{"home_id": 1}
				jsonBody, _ := json.Marshal(body)
				req = httptest.NewRequest(http.MethodPost, "/admin-action", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(http.MethodGet, "/homes/"+tt.homeID+"/admin", nil)
			}

			req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}
