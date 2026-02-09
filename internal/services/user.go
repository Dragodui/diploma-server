package services

import (
	"context"

	"github.com/Dragodui/diploma-server/internal/event"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/redis/go-redis/v9"
)

type UserService struct {
	repo  repository.UserRepository
	cache *redis.Client
}

type IUserService interface {
	GetUserByID(ctx context.Context, userID int) (*models.User, error)
	UpdateUser(ctx context.Context, userID int, name string) error
	UpdateUserAvatar(ctx context.Context, userID int, imagePath string) error
}

func NewUserService(repo repository.UserRepository, redis *redis.Client) *UserService {
	return &UserService{repo: repo, cache: redis}
}

func (s *UserService) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	return s.repo.FindByID(ctx, userID)
}

func (s *UserService) UpdateUser(ctx context.Context, userID int, name string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{}
	updates["name"] = name

	if err := s.repo.Update(ctx, user, updates); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleUser,
		Action: event.ActionUpdated,
		Data:   user,
	})

	return nil
}

func (s *UserService) UpdateUserAvatar(ctx context.Context, userID int, imagePath string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{}
	updates["avatar"] = imagePath

	if err := s.repo.Update(ctx, user, updates); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleUser,
		Action: event.ActionUpdated,
		Data:   user,
	})

	return nil
}

