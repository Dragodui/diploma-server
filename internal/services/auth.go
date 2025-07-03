package services

import (
	"errors"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/pkg/security"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	users     repository.UserRepository
	jwtSecret []byte
	ttl       time.Duration
}

func NewAuthService(repo repository.UserRepository, secret []byte, ttl time.Duration) *AuthService {
	return &AuthService{users: repo, jwtSecret: secret, ttl: ttl}
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
	user, _ := s.users.FindByEmail(email)
	if user == nil {
		return "", ErrInvalidCredentials
	}

	isValidPassword := security.ComparePasswords(user.PasswordHash, password)
	if !isValidPassword {
		return "", ErrInvalidCredentials
	}

	return security.GenerateToken(email, s.jwtSecret, s.ttl)
}
