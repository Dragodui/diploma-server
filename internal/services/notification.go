package services

import (
	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/redis/go-redis/v9"
)

type NotificationService struct {
	repo  repository.NotificationRepository
	cache *redis.Client
}

type INotificationService interface {
	// user notifications
	Create(from *int, to int, description string) error
	GetByUserID(userID int) ([]models.Notification, error)
	MarkAsRead(notificationID, userID int) error

	// home notifications
	CreateHomeNotification(from *int, homeID int, description string) error
	GetByHomeID(homeID int) ([]models.HomeNotification, error)
	MarkAsReadForHomeNotification(notificationID, homeID int) error
}

func NewNotificationService(repo repository.NotificationRepository, cache *redis.Client) *NotificationService {
	return &NotificationService{repo: repo, cache: cache}
}

func (s *NotificationService) Create(from *int, to int, description string) error {
	// remove from cache
	key := utils.GetUserNotificationsKey(to)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return s.repo.Create(&models.Notification{
		From:        from,
		To:          to,
		Description: description,
	})
}

func (s *NotificationService) GetByUserID(userID int) ([]models.Notification, error) {
	key := utils.GetUserNotificationsKey(userID)
	cached, err := utils.GetFromCache[[]models.Notification](key, s.cache)
	if cached != nil && err == nil {
		return *cached, err
	}

	notifications, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	if err := utils.WriteToCache(key, notifications, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return notifications, err
}

func (s *NotificationService) MarkAsRead(notificationID, userID int) error {
	key := utils.GetUserNotificationsKey(userID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return s.repo.MarkAsRead(notificationID)
}

func (s *NotificationService) CreateHomeNotification(from *int, homeID int, description string) error {
	key := utils.GetHomeNotificationsKey(homeID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return s.repo.CreateHomeNotification(&models.HomeNotification{
		From:        from,
		HomeID:      homeID,
		Description: description,
	})
}

func (s *NotificationService) GetByHomeID(homeID int) ([]models.HomeNotification, error) {
	key := utils.GetHomeNotificationsKey(homeID)
	cached, err := utils.GetFromCache[[]models.HomeNotification](key, s.cache)
	if cached != nil && err == nil {
		return *cached, err
	}

	notifications, err := s.repo.FindByHomeID(homeID)
	if err != nil {
		return nil, err
	}

	if err := utils.WriteToCache(key, notifications, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return notifications, err
}

func (s *NotificationService) MarkAsReadForHomeNotification(notificationID, homeID int) error {
	key := utils.GetHomeNotificationsKey(homeID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return s.repo.MarkAsReadForHomeNotification(notificationID)
}
