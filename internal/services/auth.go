package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/pkg/security"
	"github.com/markbates/goth"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	users     repository.UserRepository
	jwtSecret []byte
	ttl       time.Duration
	clientURL string
}

func NewAuthService(repo repository.UserRepository, secret []byte, ttl time.Duration, clientURL string) *AuthService {
	return &AuthService{users: repo, jwtSecret: secret, ttl: ttl, clientURL: clientURL}
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

	return security.GenerateToken(email, s.jwtSecret, s.ttl)
}

func (s *AuthService) HandleCallback(user goth.User) (string, error) {

	token, err := security.GenerateToken(user.Email, s.jwtSecret, s.ttl)
	if err != nil {
		return "", err
	}

	// Generate address for redirect for client
	clientURL := s.clientURL + "/oauth-success"
	redirectURL := fmt.Sprintf("%s?token=%s", clientURL, token)

	return redirectURL, nil
}
