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
	repo      repository.UserRepository
	jwtSecret []byte
	cache     *redis.Client
	ttl       time.Duration
	clientURL string
	mail      utils.Mailer
}

type IAuthService interface {
	Register(email, password, name string) error
	Login(email, password string) (string, error)
	HandleCallback(user goth.User) (string, error)
	SendVerificationEmail(email string) error
	VerifyEmail(token string) error
	SendResetPassword(email string) error
	ResetPassword(token, newPass string) error
}

func NewAuthService(repo repository.UserRepository, secret []byte, redis *redis.Client, ttl time.Duration, clientURL string, mail utils.Mailer) *AuthService {
	return &AuthService{repo: repo, jwtSecret: secret, cache: redis, ttl: ttl, clientURL: clientURL, mail: mail}
}

func (s *AuthService) Register(email, password, name string) error {
	existing, _ := s.repo.FindByEmail(email)
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

	return s.repo.Create(u)
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.repo.FindByEmail(email)

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
	u, err := s.repo.FindByEmail(user.Email)
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

func (s *AuthService) SendVerificationEmail(email string) error {
	tok, _ := utils.GenToken(32)
	exp := time.Now().Add(24 * time.Hour)
	if err := s.repo.SetVerifyToken(email, tok, exp); err != nil {
		return err
	}
	link := fmt.Sprintf(s.clientURL+"/verify?token=%s", tok)
	body := fmt.Sprintf("Verify email: <a href=\"%s\">%s</a>", link, link)
	return s.mail.Send(email, "Verify your email", body)
}

func (s *AuthService) VerifyEmail(token string) error {
	return s.repo.VerifyEmail(token)
}

func (s *AuthService) SendResetPassword(email string) error {
	tok, _ := utils.GenToken(32)
	exp := time.Now().Add(2 * time.Hour)
	if err := s.repo.SetResetToken(email, tok, exp); err != nil {
		return err
	}
	link := fmt.Sprintf(s.clientURL+"/reset-password?token=%s", tok)
	body := fmt.Sprintf("Reset password: <a href=\"%s\">%s</a>", link, link)
	return s.mail.Send(email, "Reset password", body)
}

func (s *AuthService) ResetPassword(token, newPass string) error {
	u, err := s.repo.GetByResetToken(token)
	if err != nil {
		return err
	}
	hash, _ := security.HashPassword(newPass)
	return s.repo.UpdatePassword(u.ID, string(hash))
}
