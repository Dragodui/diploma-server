package handlers_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/markbates/goth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock service
type mockAuthService struct {
	RegisterFunc              func(email, password, name string) error
	LoginFunc                 func(email, password string) (string, *models.User, error)
	HandleCallbackFunc        func(user goth.User) (string, error)
	SendVerificationEmailFunc func(email string) error
	VerifyEmailFunc           func(token string) error
	SendResetPasswordFunc     func(email string) error
	ResetPasswordFunc         func(token, newPass string) error
	GetUserByVerifyTokenFunc  func(token string) (*models.User, error)
	GetUserByEmailFunc        func(email string) (*models.User, error)
}

// GoogleSignIn implements services.IAuthService.
func (m *mockAuthService) GoogleSignIn(email string, name string, avatar string) (string, *models.User, error) {
	panic("unimplemented")
}

func (m *mockAuthService) Register(email, password, name string) error {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(email, password, name)
	}
	return nil
}

func (m *mockAuthService) Login(email, password string) (string, *models.User, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(email, password)
	}
	return "", nil, nil
}

func (m *mockAuthService) HandleCallback(user goth.User) (string, error) {
	if m.HandleCallbackFunc != nil {
		return m.HandleCallbackFunc(user)
	}
	return "", nil
}

func (m *mockAuthService) SendVerificationEmail(email string) error {
	if m.SendVerificationEmailFunc != nil {
		return m.SendVerificationEmailFunc(email)
	}
	return nil
}

func (m *mockAuthService) VerifyEmail(token string) error {
	if m.VerifyEmailFunc != nil {
		return m.VerifyEmailFunc(token)
	}
	return nil
}

func (m *mockAuthService) SendResetPassword(email string) error {
	if m.SendResetPasswordFunc != nil {
		return m.SendResetPasswordFunc(email)
	}
	return nil
}

func (m *mockAuthService) ResetPassword(token, newPass string) error {
	if m.ResetPasswordFunc != nil {
		return m.ResetPasswordFunc(token, newPass)
	}
	return nil
}

func (m *mockAuthService) GetUserByVerifyToken(token string) (*models.User, error) {
	if m.GetUserByVerifyTokenFunc != nil {
		return m.GetUserByVerifyTokenFunc(token)
	}
	return nil, nil
}

func (m *mockAuthService) GetUserByEmail(email string) (*models.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(email)
	}
	return nil, nil
}

// Test fixtures
var (
	validRegisterInput = models.RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	validLoginInput = models.LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}
)

func setupAuthHandler(svc *mockAuthService) *handlers.AuthHandler {
	return handlers.NewAuthHandler(svc)
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name                 string
		body                 interface{}
		registerFunc         func(email, password, name string) error
		sendVerificationFunc func(email string) error
		expectedStatus       int
		expectedBody         string
	}{
		{
			name: "Success",
			body: validRegisterInput,
			registerFunc: func(email, password, name string) error {
				assert.Equal(t, "test@example.com", email)
				assert.Equal(t, "password123", password)
				assert.Equal(t, "Test User", name)
				return nil
			},
			sendVerificationFunc: func(email string) error {
				return nil
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "Registered successfully",
		},
		{
			name:           "Invalid JSON",
			body:           "{bad json}",
			registerFunc:   nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name: "Validation Error - Missing Email",
			body: models.RegisterInput{
				Password: "password123",
				Name:     "Test User",
			},
			registerFunc:   nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Email",
		},
		{
			name: "User Already Exists",
			body: validRegisterInput,
			registerFunc: func(email, password, name string) error {
				return errors.New("user already exists")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "user already exists",
		},
		{
			name: "Send Verification Failed",
			body: validRegisterInput,
			registerFunc: func(email, password, name string) error {
				return nil
			},
			sendVerificationFunc: func(email string) error {
				return errors.New("mail error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to send verification email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthService{
				RegisterFunc:              tt.registerFunc,
				SendVerificationEmailFunc: tt.sendVerificationFunc,
			}

			h := setupAuthHandler(svc)

			var req *http.Request
			if tt.name == "Invalid JSON" {
				req = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("{bad json}"))
			} else {
				req = makeJSONRequest(http.MethodPost, "/auth/register", tt.body)
			}

			rr := httptest.NewRecorder()
			h.Register(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		getUserByEmail func(email string) (*models.User, error)
		loginFunc      func(email, password string) (string, *models.User, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			body: validLoginInput,
			getUserByEmail: func(email string) (*models.User, error) {
				return &models.User{ID: 1, Email: email, EmailVerified: true}, nil
			},
			loginFunc: func(email, password string) (string, *models.User, error) {
				require.Equal(t, "test@example.com", email)
				require.Equal(t, "password123", password)
				return "jwt-token-123", &models.User{ID: 1, Email: email}, nil
			},
			expectedStatus: http.StatusAccepted,
			expectedBody:   "jwt-token-123",
		},
		{
			name:           "Invalid JSON",
			body:           "{bad json}",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name: "Validation Error - Missing Password",
			body: models.LoginInput{
				Email: "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Password",
		},
		{
			name: "User Not Found",
			body: validLoginInput,
			getUserByEmail: func(email string) (*models.User, error) {
				return nil, errors.New("user not found")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "user not found",
		},
		{
			name: "Email Not Verified",
			body: validLoginInput,
			getUserByEmail: func(email string) (*models.User, error) {
				return &models.User{ID: 1, Email: email, EmailVerified: false}, nil
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Email is not verified",
		},
		{
			name: "Invalid Credentials",
			body: validLoginInput,
			getUserByEmail: func(email string) (*models.User, error) {
				return &models.User{ID: 1, Email: email, EmailVerified: true}, nil
			},
			loginFunc: func(email, password string) (string, *models.User, error) {
				return "", nil, errors.New("invalid credentials")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthService{
				GetUserByEmailFunc: tt.getUserByEmail,
				LoginFunc:          tt.loginFunc,
			}

			h := setupAuthHandler(svc)

			var req *http.Request
			if tt.name == "Invalid JSON" {
				req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("{bad json}"))
			} else {
				req = makeJSONRequest(http.MethodPost, "/auth/login", tt.body)
			}

			rr := httptest.NewRecorder()
			h.Login(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestAuthHandler_VerifyEmail(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		verifyFunc     func(token string) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:  "Success",
			token: "valid-token",
			verifyFunc: func(token string) error {
				require.Equal(t, "valid-token", token)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Email verified",
		},
		{
			name:  "Invalid Token",
			token: "invalid-token",
			verifyFunc: func(token string) error {
				return errors.New("invalid token")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Incorrect or expired token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthService{
				VerifyEmailFunc: tt.verifyFunc,
			}

			h := setupAuthHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/auth/verify?token="+tt.token, nil)
			rr := httptest.NewRecorder()

			h.VerifyEmail(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestAuthHandler_ForgotPassword(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			email:          "test@example.com",
			expectedStatus: http.StatusOK,
			expectedBody:   "Reset link was sended",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthService{
				SendResetPasswordFunc: func(email string) error {
					return nil
				},
			}

			h := setupAuthHandler(svc)

			form := url.Values{}
			form.Add("email", tt.email)
			req := httptest.NewRequest(http.MethodPost, "/auth/forgot", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()
			h.ForgotPassword(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestAuthHandler_ResetPassword(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		password       string
		resetFunc      func(token, newPass string) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "Success",
			token:    "valid-token",
			password: "newpassword123",
			resetFunc: func(token, newPass string) error {
				require.Equal(t, "valid-token", token)
				require.Equal(t, "newpassword123", newPass)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Password changed successfully",
		},
		{
			name:     "Invalid Token",
			token:    "invalid-token",
			password: "newpassword123",
			resetFunc: func(token, newPass string) error {
				return errors.New("invalid token")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Incorrect or expired token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthService{
				ResetPasswordFunc: tt.resetFunc,
			}

			h := setupAuthHandler(svc)

			form := url.Values{}
			form.Add("token", tt.token)
			form.Add("password", tt.password)
			req := httptest.NewRequest(http.MethodPost, "/auth/reset", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()
			h.ResetPassword(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestAuthHandler_RegenerateVerify(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		sendFunc       func(email string) error
		expectedStatus int
	}{
		{
			name:  "Success",
			email: "test@example.com",
			sendFunc: func(email string) error {
				require.Equal(t, "test@example.com", email)
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Failed to Send",
			email: "test@example.com",
			sendFunc: func(email string) error {
				return errors.New("mail error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthService{
				SendVerificationEmailFunc: tt.sendFunc,
			}

			h := setupAuthHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/auth/verify/regenerate?email="+tt.email, nil)
			rr := httptest.NewRecorder()

			h.RegenerateVerify(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
