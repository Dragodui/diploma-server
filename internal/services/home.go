package services

import (
	"errors"

	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/redis/go-redis/v9"
)

type HomeService struct {
	homes repository.HomeRepository
	cache *redis.Client
}

func NewHomeService(repo repository.HomeRepository, cache *redis.Client) *HomeService {
	return &HomeService{homes: repo, cache: cache}
}

func (s *HomeService) CreateHome(name string, userID int) error {
	inviteCode, err := s.homes.GenerateUniqueInviteCode()
	if err != nil {
		return err
	}

	home := &models.Home{
		Name:       name,
		InviteCode: inviteCode,
	}

	if err := s.homes.Create(home); err != nil {
		return err
	}

	return s.homes.AddMember(home.ID, userID, "admin")
}

func (s *HomeService) JoinHomeByCode(code string, userID int) error {
	home, err := s.homes.FindByInviteCode(code)
	if err != nil {
		return errors.New("invalid invite code")
	}

	already, err := s.homes.IsMember(home.ID, userID)
	if err != nil {
		return err
	}
	if already {
		return errors.New("user already in this home")
	}

	key := utils.GetHomeCacheKey(home.ID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return s.homes.AddMember(home.ID, userID, "member")
}

func (s *HomeService) GetHomeByID(id int) (*models.Home, error) {
	key := utils.GetHomeCacheKey(id)
	// if in cache => returns from cache
	cached, err := utils.GetFromCache[models.Home](key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	// if not in cache => returns from db
	home, err := s.homes.FindByID(id)
	if err != nil {
		return nil, err
	}

	// saves to cache
	if err := utils.WriteToCache(key, home, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return home, nil
}

func (s *HomeService) DeleteHome(id int) error {
	if err := s.homes.Delete(id); err != nil {
		return err
	}

	// remove from cache
	key := utils.GetHomeCacheKey(id)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}
	return nil
}

func (s *HomeService) LeaveHome(homeID int, userID int) error {
	key := utils.GetHomeCacheKey(homeID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return s.homes.DeleteMember(homeID, userID)
}

func (s *HomeService) RemoveMember(homeID int, userID int, currentUserID int) error {
	if userID == currentUserID {
		return errors.New("you cannot remove yourself")
	}

	key := utils.GetHomeCacheKey(homeID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return s.homes.DeleteMember(homeID, userID)
}

func (s *HomeService) GetUserHome(userID int) (*models.Home, error) {
	key := utils.GetUserHomeKey(userID)
	cached, err := utils.GetFromCache[models.Home](key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	home, err := s.homes.GetUserHome(userID)
	if err != nil {
		return nil, err
	}

	if err := utils.WriteToCache(key, home, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return home, nil
}
