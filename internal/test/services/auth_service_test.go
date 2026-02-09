package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock user repository
type mockUserRepo struct {
	CreateFunc           func(ctx context.Context, user *models.User) error
	FindByEmailFunc      func(ctx context.Context, email string) (*models.User, error)
	FindByIDFunc         func(ctx context.Context, id int) (*models.User, error)
	FindByNameFunc       func(ctx context.Context, name string) (*models.User, error)
	UpdateFunc           func(ctx context.Context, user *models.User, updates map[string]interface{}) error
	SetVerifyTokenFunc   func(ctx context.Context, email, token string, exp time.Time) error
	VerifyEmailFunc      func(ctx context.Context, token string) error
	SetResetTokenFunc    func(ctx context.Context, email, token string, exp time.Time) error
	GetByResetTokenFunc  func(ctx context.Context, token string) (*models.User, error)
	UpdatePasswordFunc   func(ctx context.Context, userID int, hash string) error
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id int) (*models.User, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepo) FindByName(ctx context.Context, name string) (*models.User, error) {
	if m.FindByNameFunc != nil {
		return m.FindByNameFunc(ctx, name)
	}
	return nil, nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *models.User, updates map[string]interface{}) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user, updates)
	}
	return nil
}

func (m *mockUserRepo) SetVerifyToken(ctx context.Context, email, token string, exp time.Time) error {
	if m.SetVerifyTokenFunc != nil {
		return m.SetVerifyTokenFunc(ctx, email, token, exp)
	}
	return nil
}

func (m *mockUserRepo) VerifyEmail(ctx context.Context, token string) error {
	if m.VerifyEmailFunc != nil {
		return m.VerifyEmailFunc(ctx, token)
	}
	return nil
}

func (m *mockUserRepo) SetResetToken(ctx context.Context, email, token string, exp time.Time) error {
	if m.SetResetTokenFunc != nil {
		return m.SetResetTokenFunc(ctx, email, token, exp)
	}
	return nil
}

func (m *mockUserRepo) GetByResetToken(ctx context.Context, token string) (*models.User, error) {
	if m.GetByResetTokenFunc != nil {
		return m.GetByResetTokenFunc(ctx, token)
	}
	return nil, nil
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, userID int, hash string) error {
	if m.UpdatePasswordFunc != nil {
		return m.UpdatePasswordFunc(ctx, userID, hash)
	}
	return nil
}

// Mock mailer
type mockMailer struct {
	SendFunc func(to, subject, body string) error
}

func (m *mockMailer) Send(to, subject, body string) error {
	if m.SendFunc != nil {
		return m.SendFunc(to, subject, body)
	}
	return nil
}

var testSecret = []byte("test-secret-key")

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		password      string
		userName      string
		findByEmail   func(ctx context.Context, email string) (*models.User, error)
		create        func(ctx context.Context, user *models.User) error
		expectedError string
	}{
		{
			name:     "Success",
			email:    "test@example.com",
			password: "password123",
			userName: "Test User",
			findByEmail: func(ctx context.Context, email string) (*models.User, error) {
				return nil, nil // User doesn't exist
			},
			create: func(ctx context.Context, user *models.User) error {
				assert.Equal(t, "test@example.com", user.Email)
				assert.Equal(t, "Test User", user.Name)
				assert.NotEmpty(t, user.PasswordHash)
				return nil
			},
			expectedError: "",
		},
		{
			name:     "User Already Exists",
			email:    "existing@example.com",
			password: "password123",
			userName: "Test User",
			findByEmail: func(ctx context.Context, email string) (*models.User, error) {
				return &models.User{ID: 1, Email: email}, nil
			},
			create:        nil,
			expectedError: "user already exists",
		},
		{
			name:     "Create Error",
			email:    "test@example.com",
			password: "password123",
			userName: "Test User",
			findByEmail: func(ctx context.Context, email string) (*models.User, error) {
				return nil, nil
			},
			create: func(ctx context.Context, user *models.User) error {
				return errors.New("database error")
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{
				FindByEmailFunc: tt.findByEmail,
				CreateFunc:      tt.create,
			}
			mailer := &mockMailer{}

			svc := services.NewAuthService(repo, testSecret, nil, time.Hour, "http://localhost", mailer)

			err := svc.Register(context.Background(), tt.email, tt.password, tt.userName)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		password      string
		findByEmail   func(ctx context.Context, email string) (*models.User, error)
		expectedError string
		expectToken   bool
	}{
		{
			name:     "Success",
			email:    "test@example.com",
			password: "password123",
			findByEmail: func(ctx context.Context, email string) (*models.User, error) {
				// Using bcrypt hash for "password123"
				return &models.User{
					ID:           1,
					Email:        email,
					PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye21K0R.1234567890abcdefghij", // This won't match
				}, nil
			},
			expectedError: "invalid credentials",
			expectToken:   false,
		},
		{
			name:     "User Not Found",
			email:    "notfound@example.com",
			password: "password123",
			findByEmail: func(ctx context.Context, email string) (*models.User, error) {
				return nil, errors.New("not found")
			},
			expectedError: "not found",
			expectToken:   false,
		},
		{
			name:     "User Is Nil",
			email:    "nil@example.com",
			password: "password123",
			findByEmail: func(ctx context.Context, email string) (*models.User, error) {
				return nil, nil
			},
			expectedError: "invalid credentials",
			expectToken:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{
				FindByEmailFunc: tt.findByEmail,
			}
			mailer := &mockMailer{}

			svc := services.NewAuthService(repo, testSecret, nil, time.Hour, "http://localhost", mailer)

			token, user, err := svc.Login(context.Background(), tt.email, tt.password)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.NotNil(t, user)
			}
		})
	}
}

func TestAuthService_VerifyEmail(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		verifyFunc    func(ctx context.Context, token string) error
		expectedError string
	}{
		{
			name:  "Success",
			token: "valid-token",
			verifyFunc: func(ctx context.Context, token string) error {
				assert.Equal(t, "valid-token", token)
				return nil
			},
			expectedError: "",
		},
		{
			name:  "Invalid Token",
			token: "invalid-token",
			verifyFunc: func(ctx context.Context, token string) error {
				return errors.New("invalid token")
			},
			expectedError: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{
				VerifyEmailFunc: tt.verifyFunc,
			}
			mailer := &mockMailer{}

			svc := services.NewAuthService(repo, testSecret, nil, time.Hour, "http://localhost", mailer)

			err := svc.VerifyEmail(context.Background(), tt.token)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuthService_ResetPassword(t *testing.T) {
	tests := []struct {
		name             string
		token            string
		newPassword      string
		getByResetToken  func(ctx context.Context, token string) (*models.User, error)
		updatePassword   func(ctx context.Context, userID int, hash string) error
		expectedError    string
	}{
		{
			name:        "Success",
			token:       "valid-reset-token",
			newPassword: "newpassword123",
			getByResetToken: func(ctx context.Context, token string) (*models.User, error) {
				assert.Equal(t, "valid-reset-token", token)
				return &models.User{ID: 1, Email: "test@example.com"}, nil
			},
			updatePassword: func(ctx context.Context, userID int, hash string) error {
				assert.Equal(t, 1, userID)
				assert.NotEmpty(t, hash)
				return nil
			},
			expectedError: "",
		},
		{
			name:        "Invalid Token",
			token:       "invalid-token",
			newPassword: "newpassword123",
			getByResetToken: func(ctx context.Context, token string) (*models.User, error) {
				return nil, errors.New("token not found")
			},
			expectedError: "token not found",
		},
		{
			name:        "Update Password Error",
			token:       "valid-token",
			newPassword: "newpassword123",
			getByResetToken: func(ctx context.Context, token string) (*models.User, error) {
				return &models.User{ID: 1}, nil
			},
			updatePassword: func(ctx context.Context, userID int, hash string) error {
				return errors.New("database error")
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{
				GetByResetTokenFunc: tt.getByResetToken,
				UpdatePasswordFunc:  tt.updatePassword,
			}
			mailer := &mockMailer{}

			svc := services.NewAuthService(repo, testSecret, nil, time.Hour, "http://localhost", mailer)

			err := svc.ResetPassword(context.Background(), tt.token, tt.newPassword)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuthService_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		findByEmail   func(ctx context.Context, email string) (*models.User, error)
		expectedUser  *models.User
		expectedError string
	}{
		{
			name:  "Success",
			email: "test@example.com",
			findByEmail: func(ctx context.Context, email string) (*models.User, error) {
				return &models.User{ID: 1, Email: email, Name: "Test User"}, nil
			},
			expectedUser:  &models.User{ID: 1, Email: "test@example.com", Name: "Test User"},
			expectedError: "",
		},
		{
			name:  "User Not Found",
			email: "notfound@example.com",
			findByEmail: func(ctx context.Context, email string) (*models.User, error) {
				return nil, errors.New("not found")
			},
			expectedUser:  nil,
			expectedError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{
				FindByEmailFunc: tt.findByEmail,
			}
			mailer := &mockMailer{}

			svc := services.NewAuthService(repo, testSecret, nil, time.Hour, "http://localhost", mailer)

			user, err := svc.GetUserByEmail(context.Background(), tt.email)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedUser.ID, user.ID)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
			}
		})
	}
}

