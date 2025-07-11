package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

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

func (s *HomeService) CreateHome(name string) error {
	inviteCode, err := s.homes.GenerateUniqueInviteCode()
	if err != nil {
		return err
	}

	return s.homes.Create(&models.Home{
		Name:       name,
		InviteCode: inviteCode,
	})
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
	if err := s.cache.Del(context.Background(), key).Err(); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return s.homes.AddMember(home.ID, userID, "member")
}

func (s *HomeService) GetHomeByID(id int) (*models.Home, error) {
	key := utils.GetHomeCacheKey(id)
	// if in cache => returns from cache
	cached, err := s.cache.Get(context.Background(), key).Result()
	if err == nil && cached != "" {
		var home models.Home
		if err := json.Unmarshal([]byte(cached), &home); err == nil {
			return &home, nil
		}
	}

	// if not in cache => returns from db
	home, err := s.homes.FindByID(id)
	if err != nil {
		return nil, err
	}

	// saves to cache
	data, err := json.Marshal(home)
	if err == nil {
		s.cache.Set(context.Background(), key, data, time.Hour)
	}

	return home, nil
}

func (s *HomeService) DeleteHome(id int) error {
	if err := s.homes.Delete(id); err != nil {
		return err
	}
	// remove from cache
	key := utils.GetHomeCacheKey(id)
	if err := s.cache.Del(context.Background(), key).Err(); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}
	return nil
}

func (s *HomeService) LeaveHome(homeID int, userID int) error {
	key := utils.GetHomeCacheKey(homeID)
	if err := s.cache.Del(context.Background(), key).Err(); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return s.homes.DeleteMember(homeID, userID)
}

func (s *HomeService) RemoveMember(homeID int, userID int, currentUserID int) error {
	if userID == currentUserID {
		return errors.New("you cannot remove yourself")
	}

	key := utils.GetHomeCacheKey(homeID)
	if err := s.cache.Del(context.Background(), key).Err(); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return s.homes.DeleteMember(homeID, userID)
}
