package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/Dragodui/diploma-server/pkg/security"
	"github.com/markbates/goth"
	"github.com/redis/go-redis/v9"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

// dummyPasswordHash is a pre-computed bcrypt hash used for timing attack mitigation
const dummyPasswordHash = "$2a$12$6EbCFrJc5PL8YvKp.qZYF.nQq3qY5jLvN8xX9X5jZrN5XqY5jZrN5"

type AuthService struct {
	repo      repository.UserRepository
	jwtSecret []byte
	cache     *redis.Client
	ttl       time.Duration
	clientURL string
	mail      utils.Mailer
}

type IAuthService interface {
	Register(ctx context.Context, email, password, name string) error
	Login(ctx context.Context, email, password string) (string, *models.User, error)
	HandleCallback(ctx context.Context, user goth.User) (token string, redirectURL string, err error)
	GoogleSignIn(ctx context.Context, email, name, avatar string) (string, *models.User, error)
	SendVerificationEmail(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, token string) error
	SendResetPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPass string) error
	GetUserByVerifyToken(ctx context.Context, token string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

func NewAuthService(repo repository.UserRepository, secret []byte, redis *redis.Client, ttl time.Duration, clientURL string, mail utils.Mailer) *AuthService {
	return &AuthService{repo: repo, jwtSecret: secret, cache: redis, ttl: ttl, clientURL: clientURL, mail: mail}
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) error {
	existing, _ := s.repo.FindByEmail(ctx, email)
	if existing != nil {
		return errors.New("user already exists")
	}

	hash, err := security.HashPassword(password)

	if err != nil {
		return err
	}
	u := &models.User{
		Email:        email,
		Name:         name,
		PasswordHash: hash,
	}

	return s.repo.Create(ctx, u)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *models.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)

	if err != nil {
		return "", nil, err
	}

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
		return "", nil, ErrInvalidCredentials
	}

	token, err := security.GenerateToken(user.ID, email, s.jwtSecret, s.ttl)
	if err != nil {
		return "", nil, err
	}

	user.PasswordHash = ""
	return token, user, nil
}

func (s *AuthService) HandleCallback(ctx context.Context, user goth.User) (string, string, error) {
	u, err := s.repo.FindByEmail(ctx, user.Email)
	if err != nil || u == nil {
		// User does not exist, create a new one
		u = &models.User{
			Email:         user.Email,
			Name:          user.Name,
			PasswordHash:  "", // No password for OAuth users
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
	}

	token, err := security.GenerateToken(u.ID, user.Email, s.jwtSecret, s.ttl)
	if err != nil {
		return "", "", err
	}

	// Return token separately to be set as HTTP-only cookie
	redirectURL := s.clientURL + "/oauth-success"

	return token, redirectURL, nil
}

// GoogleSignIn handles Google Sign-In from mobile apps using user info from Google
func (s *AuthService) GoogleSignIn(ctx context.Context, email, name, avatar string) (string, *models.User, error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil || u == nil {
		// User does not exist, create a new one
		u = &models.User{
			Email:         email,
			Name:          name,
			PasswordHash:  "", // No password for OAuth users
			EmailVerified: true, // OAuth users are already verified
			Avatar:        avatar,
		}
		if err := s.repo.Create(ctx, u); err != nil {
			return "", nil, err
		}
		// Fetch the created user to get the ID
		u, err = s.repo.FindByEmail(ctx, email)
		if err != nil {
			return "", nil, err
		}
	}

	token, err := security.GenerateToken(u.ID, email, s.jwtSecret, s.ttl)
	if err != nil {
		return "", nil, err
	}

	u.PasswordHash = ""
	return token, u, nil
}

func (s *AuthService) SendVerificationEmail(ctx context.Context, email string) error {
	tok, err := utils.GenToken(32)
	if (err != nil) {
		return err
	}
	exp := time.Now().Add(24 * time.Hour)
	if err := s.repo.SetVerifyToken(ctx, email, tok, exp); err != nil {
		return err
	}
	link := fmt.Sprintf(s.clientURL+"/verify?token=%s", tok)
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
	hash, _ := security.HashPassword(newPass)
	return s.repo.UpdatePassword(ctx, u.ID, string(hash))
}

func (s *AuthService) GetUserByVerifyToken(ctx context.Context, token string) (*models.User, error) {
	u, err := s.repo.GetByResetToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.repo.FindByEmail(ctx, email)
}

