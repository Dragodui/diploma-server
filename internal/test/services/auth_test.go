package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/pkg/security"
	"github.com/golang-jwt/jwt/v4"
	"github.com/markbates/goth"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock UserRepository
type mockUserRepo struct {
	CreateFunc          func(ctx context.Context, u *models.User) error
	FindByIDFunc        func(ctx context.Context, id int) (*models.User, error)
	FindByNameFunc      func(ctx context.Context, name string) (*models.User, error)
	FindByEmailFunc     func(ctx context.Context, email string) (*models.User, error)
	SetVerifyTokenFunc  func(ctx context.Context, email, token string, expiresAt time.Time) error
	VerifyEmailFunc     func(ctx context.Context, token string) error
	GetByResetTokenFunc func(ctx context.Context, token string) (*models.User, error)
	UpdatePasswordFunc  func(ctx context.Context, userID int, newHash string) error
	SetResetTokenFunc   func(ctx context.Context, email, token string, expiresAt time.Time) error
	UpdateFunc          func(ctx context.Context, user *models.User, updates map[string]interface{}) error
}

func (m *mockUserRepo) Create(ctx context.Context, u *models.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, u)
	}
	return nil
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

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepo) SetVerifyToken(ctx context.Context, email, token string, expiresAt time.Time) error {
	if m.SetVerifyTokenFunc != nil {
		return m.SetVerifyTokenFunc(ctx, email, token, expiresAt)
	}
	return nil
}

func (m *mockUserRepo) VerifyEmail(ctx context.Context, token string) error {
	if m.VerifyEmailFunc != nil {
		return m.VerifyEmailFunc(ctx, token)
	}
	return nil
}

func (m *mockUserRepo) GetByResetToken(ctx context.Context, token string) (*models.User, error) {
	if m.GetByResetTokenFunc != nil {
		return m.GetByResetTokenFunc(ctx, token)
	}
	return nil, nil
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, userID int, newHash string) error {
	if m.UpdatePasswordFunc != nil {
		return m.UpdatePasswordFunc(ctx, userID, newHash)
	}
	return nil
}

func (m *mockUserRepo) SetResetToken(ctx context.Context, email, token string, expiresAt time.Time) error {
	if m.SetResetTokenFunc != nil {
		return m.SetResetTokenFunc(ctx, email, token, expiresAt)
	}
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *models.User, updates map[string]interface{}) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user, updates)
	}
	return nil
}

// Mock Mailer
type mockMailer struct {
	SendFunc func(to, subject, body string) error
}

func (m *mockMailer) Send(to, subject, body string) error {
	if m.SendFunc != nil {
		return m.SendFunc(to, subject, body)
	}
	return nil
}

// Test helpers
func setupAuthService(t *testing.T, repo repository.UserRepository) (*services.AuthService, *mockMailer) {
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	mailer := &mockMailer{}
	jwtSecret := []byte("test-secret-key")
	clientURL := "http://localhost:3000"
	ttl := 24 * time.Hour

	svc := services.NewAuthService(repo, jwtSecret, redisClient, ttl, clientURL, mailer)
	return svc, mailer
}

// Register Tests
func TestAuthService_Register_Success(t *testing.T) {
	repo := &mockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return nil, errors.New("user not found")
		},
		CreateFunc: func(ctx context.Context, u *models.User) error {
			require.Equal(t, "test@example.com", u.Email)
			require.Equal(t, "Test User", u.Name)
			require.NotEqual(t, "password123", u.PasswordHash) // should be hashed
			require.False(t, u.EmailVerified)
			return nil
		},
	}

	svc, _ := setupAuthService(t, repo)
	err := svc.Register(context.Background(), "test@example.com", "password123", "Test User")
	assert.NoError(t, err)
}

func TestAuthService_Register_UserExists(t *testing.T) {
	repo := &mockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return &models.User{Email: email}, nil
		},
	}

	svc, _ := setupAuthService(t, repo)
	err := svc.Register(context.Background(), "existing@example.com", "password123", "Test User")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestAuthService_Register_PasswordHashing(t *testing.T) {
	var savedPasswordHash string
	repo := &mockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return nil, errors.New("not found")
		},
		CreateFunc: func(ctx context.Context, u *models.User) error {
			savedPasswordHash = u.PasswordHash
			return nil
		},
	}

	svc, _ := setupAuthService(t, repo)
	err := svc.Register(context.Background(), "test@example.com", "password123", "Test User")
	require.NoError(t, err)

	// Verify password is hashed (bcrypt)
	assert.NotEqual(t, "password123", savedPasswordHash)
	assert.True(t, len(savedPasswordHash) > 50) // bcrypt hashes are long

	// Verify hash is valid
	isValid := security.ComparePasswords(savedPasswordHash, "password123")
	assert.True(t, isValid)
}

// Login Tests
func TestAuthService_Login_Success(t *testing.T) {
	hashedPassword, _ := security.HashPassword("password123")
	user := &models.User{
		ID:            1,
		Email:         "test@example.com",
		PasswordHash:  hashedPassword,
		Name:          "Test User",
		EmailVerified: true,
	}

	repo := &mockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			if email == "test@example.com" {
				return user, nil
			}
			return nil, errors.New("not found")
		},
	}

	svc, _ := setupAuthService(t, repo)
	token, returnedUser, err := svc.Login(context.Background(), "test@example.com", "password123")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, user.Email, returnedUser.Email)
	assert.Equal(t, user.Name, returnedUser.Name)
	assert.Empty(t, returnedUser.PasswordHash) // password should be cleared
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	hashedPassword, _ := security.HashPassword("password123")
	user := &models.User{
		Email:         "test@example.com",
		PasswordHash:  hashedPassword,
		EmailVerified: true,
	}

	repo := &mockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return user, nil
		},
	}

	svc, _ := setupAuthService(t, repo)
	_, _, err := svc.Login(context.Background(), "test@example.com", "wrongpassword")

	assert.Error(t, err)
	assert.Equal(t, services.ErrInvalidCredentials, err)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return nil, errors.New("user not found")
		},
	}

	svc, _ := setupAuthService(t, repo)

	start := time.Now()
	_, _, err := svc.Login(context.Background(), "nonexistent@example.com", "password123")
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Equal(t, services.ErrInvalidCredentials, err)

	// Timing attack protection: should still take time even when user doesn't exist
	assert.True(t, duration > 10*time.Millisecond, "Login should take time even for non-existent users")
}

func TestAuthService_Login_JWTTokenValid(t *testing.T) {
	hashedPassword, _ := security.HashPassword("password123")
	user := &models.User{
		ID:            42,
		Email:         "test@example.com",
		PasswordHash:  hashedPassword,
		Name:          "Test User",
		EmailVerified: true,
	}

	repo := &mockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return user, nil
		},
	}

	svc, _ := setupAuthService(t, repo)
	token, _, err := svc.Login(context.Background(), "test@example.com", "password123")
	require.NoError(t, err)

	// Parse and validate JWT token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})

	require.NoError(t, err)
	require.True(t, parsedToken.Valid)

	// Check claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t, float64(42), claims["user_id"])
	assert.Equal(t, "test@example.com", claims["email"])
}

// HandleCallback Tests (OAuth)
func TestAuthService_HandleCallback_NewUser(t *testing.T) {
	gothUser := goth.User{
		Email:     "oauth@example.com",
		Name:      "OAuth User",
		AvatarURL: "https://example.com/avatar.jpg",
	}

	repo := &mockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return nil, errors.New("not found")
		},
		CreateFunc: func(ctx context.Context, u *models.User) error {
			require.Equal(t, "oauth@example.com", u.Email)
			require.Equal(t, "OAuth User", u.Name)
			require.Equal(t, "https://example.com/avatar.jpg", u.Avatar)
			require.True(t, u.EmailVerified) // OAuth users are pre-verified
			u.ID = 1
			return nil
		},
	}

	svc, _ := setupAuthService(t, repo)
	token, redirectURL, err := svc.HandleCallback(context.Background(), gothUser)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, "http://localhost:3000/auth/callback", redirectURL)
}

func TestAuthService_HandleCallback_ExistingUser(t *testing.T) {
	existingUser := &models.User{
		ID:            5,
		Email:         "oauth@example.com",
		Name:          "Existing User",
		EmailVerified: true,
	}

	gothUser := goth.User{
		Email:     "oauth@example.com",
		Name:      "OAuth User",
		AvatarURL: "https://example.com/avatar.jpg",
	}

	repo := &mockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return existingUser, nil
		},
		UpdateFunc: func(ctx context.Context, user *models.User, updates map[string]interface{}) error {
			assert.Equal(t, "https://example.com/avatar.jpg", updates["avatar"])
			assert.Equal(t, "OAuth User", updates["name"])
			return nil
		},
	}

	svc, _ := setupAuthService(t, repo)
	token, redirectURL, err := svc.HandleCallback(context.Background(), gothUser)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, "http://localhost:3000/auth/callback", redirectURL)
}

// GoogleSignIn Tests
func TestAuthService_GoogleSignIn_NewUser(t *testing.T) {
	repo := &mockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return nil, errors.New("not found")
		},
		CreateFunc: func(ctx context.Context, u *models.User) error {
			require.Equal(t, "google@example.com", u.Email)
			require.Equal(t, "Google User", u.Name)
			require.Equal(t, "https://example.com/avatar.jpg", u.Avatar)
			require.True(t, u.EmailVerified)
			u.ID = 10
			return nil
		},
	}

	svc, _ := setupAuthService(t, repo)
	token, user, err := svc.GoogleSignIn(context.Background(), "google@example.com", "Google User", "https://example.com/avatar.jpg")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, "google@example.com", user.Email)
	assert.Equal(t, "Google User", user.Name)
	assert.Empty(t, user.PasswordHash) // password should be cleared
}

func TestAuthService_GoogleSignIn_ExistingUser(t *testing.T) {
	existingUser := &models.User{
		ID:            15,
		Email:         "google@example.com",
		Name:          "Old Name",
		EmailVerified: true,
	}

	repo := &mockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return existingUser, nil
		},
		UpdateFunc: func(ctx context.Context, user *models.User, updates map[string]interface{}) error {
			assert.Equal(t, "New Name", updates["name"])
			assert.Equal(t, "https://new-avatar.jpg", updates["avatar"])
			return nil
		},
	}

	svc, _ := setupAuthService(t, repo)
	token, user, err := svc.GoogleSignIn(context.Background(), "google@example.com", "New Name", "https://new-avatar.jpg")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, "google@example.com", user.Email)
}

// VerifyEmail Tests
func TestAuthService_VerifyEmail_Success(t *testing.T) {
	repo := &mockUserRepo{
		VerifyEmailFunc: func(ctx context.Context, token string) error {
			assert.Equal(t, "valid-token", token)
			return nil
		},
	}

	svc, _ := setupAuthService(t, repo)
	err := svc.VerifyEmail(context.Background(), "valid-token")

	assert.NoError(t, err)
}

func TestAuthService_VerifyEmail_InvalidToken(t *testing.T) {
	repo := &mockUserRepo{
		VerifyEmailFunc: func(ctx context.Context, token string) error {
			return errors.New("invalid token")
		},
	}

	svc, _ := setupAuthService(t, repo)
	err := svc.VerifyEmail(context.Background(), "invalid-token")

	assert.Error(t, err)
}

// GetUserByEmail Tests
func TestAuthService_GetUserByEmail_Success(t *testing.T) {
	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Name:  "Test User",
	}

	repo := &mockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return user, nil
		},
	}

	svc, _ := setupAuthService(t, repo)
	result, err := svc.GetUserByEmail(context.Background(), "test@example.com")

	assert.NoError(t, err)
	assert.Equal(t, user.Email, result.Email)
}

func TestAuthService_GetUserByEmail_NotFound(t *testing.T) {
	repo := &mockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return nil, errors.New("not found")
		},
	}

	svc, _ := setupAuthService(t, repo)
	_, err := svc.GetUserByEmail(context.Background(), "nonexistent@example.com")

	assert.Error(t, err)
}
