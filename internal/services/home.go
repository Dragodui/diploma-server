package services

import (
	"context"
	"errors"

	"github.com/Dragodui/diploma-server/internal/event"
	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/redis/go-redis/v9"
)

type HomeService struct {
	repo  repository.HomeRepository
	cache *redis.Client
}

type IHomeService interface {
	CreateHome(ctx context.Context, name string, userID int) error
	RegenerateInviteCode(ctx context.Context, homeID int) error
	JoinHomeByCode(ctx context.Context, code string, userID int) error
	GetHomeByID(ctx context.Context, id int) (*models.Home, error)
	DeleteHome(ctx context.Context, id int) error
	LeaveHome(ctx context.Context, homeID int, userID int) error
	RemoveMember(ctx context.Context, homeID int, userID int, currentUserID int) error
	GetUserHome(ctx context.Context, userID int) (*models.Home, error)
}

func NewHomeService(repo repository.HomeRepository, cache *redis.Client) *HomeService {
	return &HomeService{repo: repo, cache: cache}
}

func (s *HomeService) CreateHome(ctx context.Context, name string, userID int) error {
	inviteCode, err := s.repo.GenerateUniqueInviteCode(ctx)
	if err != nil {
		return err
	}

	home := &models.Home{
		Name:       name,
		InviteCode: inviteCode,
	}

	if err := s.repo.Create(ctx, home); err != nil {
		return err
	}

	if err := s.repo.AddMember(ctx, home.ID, userID, "admin"); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleHome,
		Action: event.ActionCreated,
		Data:   home,
	})

	return nil
}

func (s *HomeService) RegenerateInviteCode(ctx context.Context, homeID int) error {
	inviteCode, err := s.repo.GenerateUniqueInviteCode(ctx)
	if err != nil {
		return err
	}

	key := utils.GetHomeCacheKey(homeID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	if err := s.repo.RegenerateCode(ctx, inviteCode, homeID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleHome,
		Action: event.ActionUpdated,
		Data:   map[string]int{"homeID": homeID},
	})

	return nil
}

func (s *HomeService) JoinHomeByCode(ctx context.Context, code string, userID int) error {
	home, err := s.repo.FindByInviteCode(ctx, code)
	if err != nil {
		return errors.New("invalid invite code")
	}

	already, err := s.repo.IsMember(ctx, home.ID, userID)
	if err != nil {
		return err
	}
	if already {
		return errors.New("user already in this home")
	}

	key := utils.GetHomeCacheKey(home.ID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	if err := s.repo.AddMember(ctx, home.ID, userID, "member"); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleHome,
		Action: event.ActionMemberJoined,
		Data:   map[string]int{"homeID": home.ID, "userID": userID},
	})

	return nil
}

func (s *HomeService) GetHomeByID(ctx context.Context, id int) (*models.Home, error) {
	key := utils.GetHomeCacheKey(id)
	// if in cache => returns from cache
	cached, err := utils.GetFromCache[models.Home](ctx, key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	// if not in cache => returns from db
	home, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// saves to cache
	if err := utils.WriteToCache(ctx, key, home, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return home, nil
}

func (s *HomeService) DeleteHome(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// remove from cache
	key := utils.GetHomeCacheKey(id)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleHome,
		Action: event.ActionDeleted,
		Data:   map[string]int{"id": id},
	})

	return nil
}

func (s *HomeService) LeaveHome(ctx context.Context, homeID int, userID int) error {
	key := utils.GetHomeCacheKey(homeID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	if err := s.repo.DeleteMember(ctx, homeID, userID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleHome,
		Action: event.ActionMemberLeft,
		Data:   map[string]int{"homeID": homeID, "userID": userID},
	})

	return nil
}

func (s *HomeService) RemoveMember(ctx context.Context, homeID int, userID int, currentUserID int) error {
	if userID == currentUserID {
		return errors.New("you cannot remove yourself")
	}

	key := utils.GetHomeCacheKey(homeID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	if err := s.repo.DeleteMember(ctx, homeID, userID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleHome,
		Action: event.ActionMemberRemoved,
		Data:   map[string]int{"homeID": homeID, "userID": userID},
	})

	return nil
}

func (s *HomeService) GetUserHome(ctx context.Context, userID int) (*models.Home, error) {
	key := utils.GetUserHomeKey(userID)
	cached, err := utils.GetFromCache[models.Home](ctx, key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	home, err := s.repo.GetUserHome(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := utils.WriteToCache(ctx, key, home, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return home, nil
}

