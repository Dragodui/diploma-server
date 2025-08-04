package services

import (
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

type AuthService struct {
	users     repository.UserRepository
	jwtSecret []byte
	cache     *redis.Client
	ttl       time.Duration
	clientURL string
	mail      utils.Mailer
}

func NewAuthService(repo repository.UserRepository, secret []byte, redis *redis.Client, ttl time.Duration, clientURL string) *AuthService {
	return &AuthService{users: repo, jwtSecret: secret, cache: redis, ttl: ttl, clientURL: clientURL}
}

func (s *AuthService) Register(email, password, name string) error {
	existing, _ := s.users.FindByEmail(email)
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

	return s.users.Create(u)
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.users.FindByEmail(email)

	if err != nil {
		return "", err
	}

	if user == nil {
		return "", ErrInvalidCredentials
	}

	isValidPassword := security.ComparePasswords(user.PasswordHash, password)
	if !isValidPassword {
		return "", ErrInvalidCredentials
	}

	return security.GenerateToken(user.ID, email, s.jwtSecret, s.ttl)
}

func (s *AuthService) HandleCallback(user goth.User) (string, error) {
	u, err := s.users.FindByEmail(user.Email)
	if err != nil {
		return "", err
	}

	token, err := security.GenerateToken(u.ID, user.Email, s.jwtSecret, s.ttl)
	if err != nil {
		return "", err
	}

	// Generate address for redirect for client
	clientURL := s.clientURL + "/oauth-success"
	redirectURL := fmt.Sprintf("%s?token=%s", clientURL, token)

	return redirectURL, nil
}

func (s *AuthService) SendVerificationEmail(userID int, email string) error {
	tok, _ := utils.GenToken(32)
	exp := time.Now().Add(24 * time.Hour)
	if err := s.users.SetVerifyToken(userID, tok, exp); err != nil {
		return err
	}
	link := fmt.Sprintf(s.clientURL+"/verify?token=%s", tok)
	body := fmt.Sprintf("Verify email: <a href=\"%s\">%s</a>", link, link)
	return s.mail.Send(email, "Verify your email", body)
}

func (s *AuthService) VerifyEmail(token string) error {
	return s.users.VerifyEmail(token)
}

func (s *AuthService) SendResetPassword(email string) error {
	tok, _ := utils.GenToken(32)
	exp := time.Now().Add(2 * time.Hour)
	if err := s.users.SetResetToken(email, tok, exp); err != nil {
		return err
	}
	link := fmt.Sprintf(s.clientURL+"/reset-password?token=%s", tok)
	body := fmt.Sprintf("Reset password: <a href=\"%s\">%s</a>", link, link)
	return s.mail.Send(email, "Reset password", body)
}

func (s *AuthService) ResetPassword(token, newPass string) error {
	u, err := s.users.GetByResetToken(token)
	if err != nil {
		return err
	}
	hash, _ := security.HashPassword(newPass)
	return s.users.UpdatePassword(u.ID, string(hash))
}
