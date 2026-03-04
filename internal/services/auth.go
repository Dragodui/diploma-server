package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Dragodui/diploma-server/internal/metrics"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/Dragodui/diploma-server/pkg/security"
	"github.com/markbates/goth"
	"github.com/redis/go-redis/v9"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrEmailNotVerified = errors.New("email is not verified")

// dummyPasswordHash is a pre-computed bcrypt hash used for timing attack mitigation
const dummyPasswordHash = "$2a$12$6EbCFrJc5PL8YvKp.qZYF.nQq3qY5jLvN8xX9X5jZrN5XqY5jZrN5"

type AuthService struct {
	repo      repository.UserRepository
	jwtSecret []byte
	cache     *redis.Client
	ttl       time.Duration
	clientURL string
	serverURL string
	mail      utils.Mailer
}

type IAuthService interface {
	Register(ctx context.Context, email, password, name string) error
	Login(ctx context.Context, email, password string) (string, *models.User, error)
	Logout(ctx context.Context, tokenStr string) error
	IsTokenBlacklisted(ctx context.Context, tokenStr string) bool
	HandleCallback(ctx context.Context, user goth.User) (token string, redirectURL string, err error)
	GoogleSignIn(ctx context.Context, accessToken string) (string, *models.User, error)
	SendVerificationEmail(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, token string) error
	SendResetPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPass string) error
	GetUserByVerifyToken(ctx context.Context, token string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	ChangePassword(ctx context.Context, userID int, currentPassword, newPassword string) error
}

func NewAuthService(repo repository.UserRepository, secret []byte, redis *redis.Client, ttl time.Duration, clientURL, serverURL string, mail utils.Mailer) *AuthService {
	return &AuthService{repo: repo, jwtSecret: secret, cache: redis, ttl: ttl, clientURL: strings.TrimRight(clientURL, "/"), serverURL: strings.TrimRight(serverURL, "/"), mail: mail}
}

const tokenBlacklistPrefix = "blacklist:"

// Logout blacklists the given JWT token in Redis until it expires.
func (s *AuthService) Logout(ctx context.Context, tokenStr string) error {
	claims, err := security.ParseToken(tokenStr, s.jwtSecret)
	if err != nil {
		return errors.New("invalid token")
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil // already expired
	}
	return s.cache.Set(ctx, tokenBlacklistPrefix+tokenStr, "1", ttl).Err()
}

// IsTokenBlacklisted checks if the token has been revoked via logout.
func (s *AuthService) IsTokenBlacklisted(ctx context.Context, tokenStr string) bool {
	val, err := s.cache.Exists(ctx, tokenBlacklistPrefix+tokenStr).Result()
	return err == nil && val > 0
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) error {
	existing, _ := s.repo.FindByEmail(ctx, email)
	if existing != nil {
		metrics.AuthAttemptsTotal.WithLabelValues("register", "failure").Inc()
		return errors.New("registration failed")
	}

	hash, err := security.HashPassword(password)

	if err != nil {
		metrics.AuthAttemptsTotal.WithLabelValues("register", "failure").Inc()
		return err
	}
	u := &models.User{
		Email:        email,
		Name:         name,
		PasswordHash: hash,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		metrics.AuthAttemptsTotal.WithLabelValues("register", "failure").Inc()
		return err
	}
	metrics.AuthAttemptsTotal.WithLabelValues("register", "success").Inc()
	return nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *models.User, error) {
	user, _ := s.repo.FindByEmail(ctx, email)

	// Timing attack mitigation: always perform password comparison
	// Use dummy hash if user not found to ensure constant-time response
	hashToCompare := dummyPasswordHash
	if user != nil {
		hashToCompare = user.PasswordHash
	}

	// Always execute bcrypt comparison regardless of whether user exists
	isValidPassword := security.ComparePasswords(hashToCompare, password)

	// Check both conditions: user must exist AND password must be valid
	if user == nil || !isValidPassword {
		metrics.AuthAttemptsTotal.WithLabelValues("login", "failure").Inc()
		return "", nil, ErrInvalidCredentials
	}

	// Check email verification after constant-time password check
	if !user.EmailVerified {
		metrics.AuthAttemptsTotal.WithLabelValues("login", "failure").Inc()
		return "", nil, ErrEmailNotVerified
	}

	token, err := security.GenerateToken(user.ID, email, s.jwtSecret, s.ttl)
	if err != nil {
		metrics.AuthAttemptsTotal.WithLabelValues("login", "failure").Inc()
		return "", nil, err
	}

	metrics.AuthAttemptsTotal.WithLabelValues("login", "success").Inc()
	metrics.AuthTokensGenerated.Inc()
	user.PasswordHash = ""
	return token, user, nil
}

func (s *AuthService) HandleCallback(ctx context.Context, user goth.User) (string, string, error) {
	u, err := s.repo.FindByEmail(ctx, user.Email)
	if err != nil {
		return "", "", err
	}
	if u == nil {
		// User does not exist, create a new one
		u = &models.User{
			Email:         user.Email,
			Name:          user.Name,
			PasswordHash:  "",   // No password for OAuth users
			EmailVerified: true, // OAuth users are already verified
			Avatar:        user.AvatarURL,
		}
		if err := s.repo.Create(ctx, u); err != nil {
			return "", "", err
		}
		// Fetch the created user to get the ID
		u, err = s.repo.FindByEmail(ctx, user.Email)
		if err != nil {
			return "", "", err
		}
		if u == nil {
			return "", "", errors.New("user not found")
		}
	}

	token, err := security.GenerateToken(u.ID, user.Email, s.jwtSecret, s.ttl)
	if err != nil {
		metrics.AuthAttemptsTotal.WithLabelValues("oauth", "failure").Inc()
		return "", "", err
	}

	metrics.AuthAttemptsTotal.WithLabelValues("oauth", "success").Inc()
	metrics.AuthTokensGenerated.Inc()

	// Return token separately to be set as HTTP-only cookie
	redirectURL := s.clientURL + "/oauth-success"

	return token, redirectURL, nil
}

// googleUserInfo represents the response from Google's userinfo endpoint.
type googleUserInfo struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// verifyGoogleAccessToken verifies the access token with Google and returns user info.
func verifyGoogleAccessToken(ctx context.Context, accessToken string) (*googleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/userinfo/v2/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify Google token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid Google access token")
	}

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode Google response: %w", err)
	}

	if info.Email == "" {
		return nil, errors.New("Google account has no email")
	}

	return &info, nil
}

// GoogleSignIn handles Google Sign-In from mobile apps by verifying the access token with Google.
func (s *AuthService) GoogleSignIn(ctx context.Context, accessToken string) (string, *models.User, error) {
	// Verify the access token with Google to get trusted user info
	gInfo, err := verifyGoogleAccessToken(ctx, accessToken)
	if err != nil {
		return "", nil, err
	}

	u, err := s.repo.FindByEmail(ctx, gInfo.Email)
	if err != nil || u == nil {
		// User does not exist, create a new one
		u = &models.User{
			Email:         gInfo.Email,
			Name:          gInfo.Name,
			PasswordHash:  "",   // No password for OAuth users
			EmailVerified: true, // OAuth users are already verified
			Avatar:        gInfo.Picture,
		}
		if err := s.repo.Create(ctx, u); err != nil {
			return "", nil, err
		}
		// Fetch the created user to get the ID
		u, err = s.repo.FindByEmail(ctx, gInfo.Email)
		if err != nil {
			return "", nil, err
		}
		if u == nil {
			return "", nil, errors.New("user not found after creation")
		}
	}

	token, err := security.GenerateToken(u.ID, gInfo.Email, s.jwtSecret, s.ttl)
	if err != nil {
		metrics.AuthAttemptsTotal.WithLabelValues("google_signin", "failure").Inc()
		return "", nil, err
	}

	metrics.AuthAttemptsTotal.WithLabelValues("google_signin", "success").Inc()
	metrics.AuthTokensGenerated.Inc()
	u.PasswordHash = ""
	return token, u, nil
}

func (s *AuthService) SendVerificationEmail(ctx context.Context, email string) error {
	tok, err := utils.GenToken(32)
	if err != nil {
		return err
	}
	exp := time.Now().Add(24 * time.Hour)
	if err := s.repo.SetVerifyToken(ctx, email, tok, exp); err != nil {
		return err
	}
	link := fmt.Sprintf(s.serverURL+"/api/auth/verify?token=%s", tok)
	body := fmt.Sprintf("Verify email: <a href=\"%s\">%s</a>", link, link)
	return s.mail.Send(email, "Verify your email", body)
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	return s.repo.VerifyEmail(ctx, token)
}

func (s *AuthService) SendResetPassword(ctx context.Context, email string) error {
	// Generate secure random token
	tok, err := utils.GenToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	exp := time.Now().Add(2 * time.Hour)

	// SetResetToken will return nil even if user doesn't exist
	if err := s.repo.SetResetToken(ctx, email, tok, exp); err != nil {
		// Only log database errors, don't expose to client
		return err
	}

	// Send reset email
	link := fmt.Sprintf(s.clientURL+"/reset-password?token=%s", tok)
	body := fmt.Sprintf("Reset password: <a href=\"%s\">%s</a>", link, link)
	_ = s.mail.Send(email, "Reset password", body)

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPass string) error {
	u, err := s.repo.GetByResetToken(ctx, token)
	if err != nil {
		return err
	}
	hash, err := security.HashPassword(newPass)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return s.repo.UpdatePassword(ctx, u.ID, hash)
}

func (s *AuthService) GetUserByVerifyToken(ctx context.Context, token string) (*models.User, error) {
	u, err := s.repo.GetByVerifyToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, err
}

func (s *AuthService) ChangePassword(ctx context.Context, userID int, currentPassword, newPassword string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if user.PasswordHash == "" {
		return errors.New("cannot change password for OAuth account")
	}

	if !security.ComparePasswords(user.PasswordHash, currentPassword) {
		return ErrInvalidCredentials
	}

	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, userID, string(hash))
}
