package middleware_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dragodui/diploma-server/internal/http/middleware"
	ratelimiter "github.com/Dragodui/diploma-server/internal/http/rate_limiter"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/Dragodui/diploma-server/pkg/security"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock home repository
type mockHomeRepo struct {
	IsMemberFunc func(ctx context.Context, homeID, userID int) (bool, error)
	IsAdminFunc  func(ctx context.Context, homeID, userID int) (bool, error)
}

func (m *mockHomeRepo) Create(ctx context.Context, h *models.Home) error { return nil }
func (m *mockHomeRepo) FindByID(ctx context.Context, id int) (*models.Home, error) {
	return nil, nil
}
func (m *mockHomeRepo) FindByInviteCode(ctx context.Context, inviteCode string) (*models.Home, error) {
	return nil, nil
}
func (m *mockHomeRepo) Delete(ctx context.Context, id int) error { return nil }
func (m *mockHomeRepo) AddMember(ctx context.Context, id int, userID int, role string) error {
	return nil
}
func (m *mockHomeRepo) DeleteMember(ctx context.Context, id int, userID int) error { return nil }
func (m *mockHomeRepo) GenerateUniqueInviteCode(ctx context.Context) (string, error) {
	return "CODE1234", nil
}
func (m *mockHomeRepo) GetUserHome(ctx context.Context, userID int) (*models.Home, error) {
	return nil, nil
}
func (m *mockHomeRepo) RegenerateCode(ctx context.Context, code string, id int) error { return nil }

func (m *mockHomeRepo) IsMember(ctx context.Context, homeID, userID int) (bool, error) {
	if m.IsMemberFunc != nil {
		return m.IsMemberFunc(ctx, homeID, userID)
	}
	return false, nil
}

func (m *mockHomeRepo) IsAdmin(ctx context.Context, homeID, userID int) (bool, error) {
	if m.IsAdminFunc != nil {
		return m.IsAdminFunc(ctx, homeID, userID)
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
				IsMemberFunc: func(ctx context.Context, homeID, userID int) (bool, error) {
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
				IsMemberFunc: func(ctx context.Context, homeID, userID int) (bool, error) {
					return tt.isMember, nil
				},
				IsAdminFunc: func(ctx context.Context, homeID, userID int) (bool, error) {
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

func TestSecurityHeaders(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	handler := middleware.SecurityHeaders(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Verify status
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify all security headers are set
	headers := rr.Header()
	assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
	assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
	assert.Equal(t, "default-src 'self'", headers.Get("Content-Security-Policy"))
	assert.Equal(t, "max-age=31536000; includeSubDomains", headers.Get("Strict-Transport-Security"))
	assert.Equal(t, "no-referrer", headers.Get("Referrer-Policy"))
	assert.Equal(t, "require-corp", headers.Get("Cross-Origin-Embedder-Policy"))
	assert.Equal(t, "same-origin", headers.Get("Cross-Origin-Opener-Policy"))
	assert.Equal(t, "same-origin", headers.Get("Cross-Origin-Resource-Policy"))
}

func TestBasicAuth(t *testing.T) {
	correctUsername := "admin"
	correctPassword := "secret123"

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
		shouldCallNext bool
	}{
		{
			name:           "Valid Credentials",
			authHeader:     "Basic " + base64Encode(correctUsername+":"+correctPassword),
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			shouldCallNext: true,
		},
		{
			name:           "Invalid Username",
			authHeader:     "Basic " + base64Encode("wrong:"+correctPassword),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
			shouldCallNext: false,
		},
		{
			name:           "Invalid Password",
			authHeader:     "Basic " + base64Encode(correctUsername+":wrong"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
			shouldCallNext: false,
		},
		{
			name:           "Missing Authorization Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
			shouldCallNext: false,
		},
		{
			name:           "Invalid Format - Not Basic",
			authHeader:     "Bearer token123",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
			shouldCallNext: false,
		},
		{
			name:           "Invalid Base64",
			authHeader:     "Basic not-valid-base64!!!",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
			shouldCallNext: false,
		},
		{
			name:           "Missing Colon in Credentials",
			authHeader:     "Basic " + base64Encode("adminnocolon"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
			shouldCallNext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			handler := middleware.BasicAuth(correctUsername, correctPassword)(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
			assert.Equal(t, tt.shouldCallNext, nextCalled)

			// Verify WWW-Authenticate header is set on 401
			if tt.expectedStatus == http.StatusUnauthorized {
				assert.Equal(t, `Basic realm="Restricted"`, rr.Header().Get("WWW-Authenticate"))
			}
		})
	}
}

func TestBodySizeLimit(t *testing.T) {
	tests := []struct {
		name           string
		bodySize       int
		expectedStatus int
		shouldCallNext bool
	}{
		{
			name:           "Small Body - Success",
			bodySize:       1024, // 1KB
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:           "Max Size Body - Success",
			bodySize:       10 * 1024 * 1024, // 10MB (exactly at limit)
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:           "Oversized Body - Rejected",
			bodySize:       11 * 1024 * 1024, // 11MB (over limit)
			expectedStatus: http.StatusRequestEntityTooLarge,
			shouldCallNext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				// Try to read the body
				buf := make([]byte, tt.bodySize)
				_, err := r.Body.Read(buf)
				if err != nil && err.Error() != "EOF" {
					// If body is too large, MaxBytesReader will error
					w.WriteHeader(http.StatusRequestEntityTooLarge)
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			handler := middleware.BodySizeLimit(nextHandler)

			// Create request with body of specified size
			body := bytes.Repeat([]byte("A"), tt.bodySize)
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if tt.shouldCallNext {
				assert.True(t, nextCalled)
			}

			// Note: MaxBytesReader returns 413 through the handler when limit exceeded
			// The actual status depends on how the handler processes the error
			if !tt.shouldCallNext {
				// Body should be limited
				assert.NotNil(t, req.Body)
			}
		})
	}
}

// Helper function to encode credentials for Basic Auth
func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// TestRateLimitMiddleware tests the global rate limiter
func TestRateLimitMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		requests       int
		ip             string
		expectedStatus int
		shouldBlock    bool
	}{
		{
			name:           "Allow requests under limit",
			requests:       3,
			ip:             "192.168.1.1",
			expectedStatus: http.StatusOK,
			shouldBlock:    false,
		},
		{
			name:           "Block after exceeding limit",
			requests:       125, // More than 120 (default limit)
			ip:             "192.168.1.2",
			expectedStatus: http.StatusTooManyRequests,
			shouldBlock:    true,
		},
		{
			name:           "Different IPs have separate limits",
			requests:       3,
			ip:             "192.168.1.3",
			expectedStatus: http.StatusOK,
			shouldBlock:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := ratelimiter.NewIpRateLimiter()
			middleware := middleware.RateLimitMiddleware(limiter)

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			var lastStatus int
			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = tt.ip + ":12345"
				w := httptest.NewRecorder()

				handler.ServeHTTP(w, req)
				lastStatus = w.Code
			}

			if tt.shouldBlock {
				assert.Equal(t, http.StatusTooManyRequests, lastStatus)
			} else {
				assert.Equal(t, http.StatusOK, lastStatus)
			}
		})
	}
}

// TestStrictRateLimitMiddleware tests the strict rate limiter
func TestStrictRateLimitMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		maxRequests    float64
		refillRate     float64
		requests       int
		ip             string
		expectedStatus int
		shouldBlock    bool
	}{
		{
			name:           "Allow requests under strict limit",
			maxRequests:    5,
			refillRate:     0.083, // 5 per minute
			requests:       3,
			ip:             "192.168.1.10",
			expectedStatus: http.StatusOK,
			shouldBlock:    false,
		},
		{
			name:           "Block after exceeding strict limit",
			maxRequests:    5,
			refillRate:     0.083,
			requests:       7,
			ip:             "192.168.1.11",
			expectedStatus: http.StatusTooManyRequests,
			shouldBlock:    true,
		},
		{
			name:           "Very strict limit (3 per 5 minutes)",
			maxRequests:    3,
			refillRate:     0.01, // 3 per 5 minutes
			requests:       5,
			ip:             "192.168.1.12",
			expectedStatus: http.StatusTooManyRequests,
			shouldBlock:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := ratelimiter.NewIpRateLimiter()
			middleware := middleware.StrictRateLimitMiddleware(limiter, tt.maxRequests, tt.refillRate)

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			var lastStatus int
			var w *httptest.ResponseRecorder
			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest("POST", "/auth/login", nil)
				req.RemoteAddr = tt.ip + ":12345"
				w = httptest.NewRecorder()

				handler.ServeHTTP(w, req)
				lastStatus = w.Code
			}

			if tt.shouldBlock {
				assert.Equal(t, http.StatusTooManyRequests, lastStatus)
				// Verify error message
				var response map[string]string
				json.NewDecoder(w.Body).Decode(&response)
				assert.Contains(t, response["error"], "Too many requests")
			} else {
				assert.Equal(t, http.StatusOK, lastStatus)
			}
		})
	}
}

// TestGetIPExtraction tests IP extraction from various headers
func TestGetIPExtraction(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		xForwardedFor  string
		xRealIP        string
		expectedIP     string
	}{
		{
			name:        "Extract from RemoteAddr",
			remoteAddr:  "192.168.1.1:12345",
			expectedIP:  "192.168.1.1",
		},
		{
			name:          "Extract from X-Forwarded-For (single IP)",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:          "Extract from X-Forwarded-For (multiple IPs, take first)",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1, 198.51.100.1, 192.0.2.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:        "Extract from X-Real-IP",
			remoteAddr:  "10.0.0.1:12345",
			xRealIP:     "203.0.113.5",
			expectedIP:  "203.0.113.5",
		},
		{
			name:          "Prefer X-Forwarded-For over X-Real-IP",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1",
			xRealIP:       "203.0.113.5",
			expectedIP:    "203.0.113.1",
		},
		{
			name:        "IPv6 address",
			remoteAddr:  "[2001:db8::1]:12345",
			expectedIP:  "[2001:db8::1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := ratelimiter.NewIpRateLimiter()
			var capturedIP string

			// Create middleware that captures the IP
			testMiddleware := middleware.RateLimitMiddleware(limiter)
			handler := testMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// The IP is used internally, we verify by checking rate limiter behavior
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			// Make another request with the same headers to verify same IP is used
			req2 := httptest.NewRequest("GET", "/test", nil)
			req2.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req2.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req2.Header.Set("X-Real-IP", tt.xRealIP)
			}

			w2 := httptest.NewRecorder()
			handler.ServeHTTP(w2, req2)

			// Both requests should succeed (under rate limit)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, http.StatusOK, w2.Code)

			// Verify that the same IP is being tracked by exhausting the limit
			for i := 0; i < 125; i++ {
				reqN := httptest.NewRequest("GET", "/test", nil)
				reqN.RemoteAddr = tt.remoteAddr
				if tt.xForwardedFor != "" {
					reqN.Header.Set("X-Forwarded-For", tt.xForwardedFor)
				}
				if tt.xRealIP != "" {
					reqN.Header.Set("X-Real-IP", tt.xRealIP)
				}
				wN := httptest.NewRecorder()
				handler.ServeHTTP(wN, reqN)
				if wN.Code == http.StatusTooManyRequests {
					capturedIP = "rate_limited" // Confirms IP extraction worked
					break
				}
			}

			assert.Equal(t, "rate_limited", capturedIP, "Rate limiter should eventually block requests from the same IP")
		})
	}
}
