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
	repo repository.HomeRepository
	cache *redis.Client
}

type IHomeService interface {
	CreateHome(name string, userID int) error
	JoinHomeByCode(code string, userID int) error
	GetHomeByID(id int) (*models.Home, error)
	DeleteHome(id int) error
	LeaveHome(homeID int, userID int) error
	RemoveMember(homeID int, userID int, currentUserID int) error
	GetUserHome(userID int) (*models.Home, error)
}

func NewHomeService(repo repository.HomeRepository, cache *redis.Client) *HomeService {
	return &HomeService{repo: repo, cache: cache}
}

func (s *HomeService) CreateHome(name string, userID int) error {
	inviteCode, err := s.repo.GenerateUniqueInviteCode()
	if err != nil {
		return err
	}

	home := &models.Home{
		Name:       name,
		InviteCode: inviteCode,
	}

	if err := s.repo.Create(home); err != nil {
		return err
	}

	return s.repo.AddMember(home.ID, userID, "admin")
}

func (s *HomeService) JoinHomeByCode(code string, userID int) error {
	home, err := s.repo.FindByInviteCode(code)
	if err != nil {
		return errors.New("invalid invite code")
	}

	already, err := s.repo.IsMember(home.ID, userID)
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

	return s.repo.AddMember(home.ID, userID, "member")
}

func (s *HomeService) GetHomeByID(id int) (*models.Home, error) {
	key := utils.GetHomeCacheKey(id)
	// if in cache => returns from cache
	cached, err := utils.GetFromCache[models.Home](key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	// if not in cache => returns from db
	home, err := s.repo.FindByID(id)
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
	if err := s.repo.Delete(id); err != nil {
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

	return s.repo.DeleteMember(homeID, userID)
}

func (s *HomeService) RemoveMember(homeID int, userID int, currentUserID int) error {
	if userID == currentUserID {
		return errors.New("you cannot remove yourself")
	}

	key := utils.GetHomeCacheKey(homeID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return s.repo.DeleteMember(homeID, userID)
}

func (s *HomeService) GetUserHome(userID int) (*models.Home, error) {
	key := utils.GetUserHomeKey(userID)
	cached, err := utils.GetFromCache[models.Home](key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	home, err := s.repo.GetUserHome(userID)
	if err != nil {
		return nil, err
	}

	if err := utils.WriteToCache(key, home, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return home, nil
}
