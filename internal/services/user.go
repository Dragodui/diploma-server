package services

import (
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/redis/go-redis/v9"
)

type UserService struct {
	repo  repository.UserRepository
	cache *redis.Client
}

type IUserService interface {
	GetUserByID(userID int) (*models.User, error)
	UpdateUser(userID int, name string) error
	UpdateUserAvatar(userID int, imagePath string) error
}

func NewUserService(repo repository.UserRepository, redis *redis.Client) *UserService {
	return &UserService{repo: repo, cache: redis}
}

func (s *UserService) GetUserByID(userID int) (*models.User, error) {
	return s.repo.FindByID(userID)
}

func (s *UserService) UpdateUser(userID int, name string) error {
	user, err := s.repo.FindByID(userID)

	if err != nil {
		return err
	}
	updates := map[string]interface{}{}

	updates["name"] = name

	return s.repo.Update(user, updates)
}

func (s *UserService) UpdateUserAvatar(userID int, imagePath string) error {
	user, err := s.repo.FindByID(userID)

	if err != nil {
		return err
	}
	updates := map[string]interface{}{}

	updates["avatar"] = imagePath

	return s.repo.Update(user, updates)
}
