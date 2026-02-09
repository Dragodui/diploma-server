package repository

import (
	"context"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *models.Notification) error
	FindByUserID(ctx context.Context, id int) ([]models.Notification, error)
	MarkAsRead(ctx context.Context, id int) error

	CreateHomeNotification(ctx context.Context, n *models.HomeNotification) error
	FindByHomeID(ctx context.Context, id int) ([]models.HomeNotification, error)
	MarkAsReadForHomeNotification(ctx context.Context, id int) error
}

type notificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepo{db}
}

// user notifications
func (r *notificationRepo) Create(ctx context.Context, n *models.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *notificationRepo) FindByUserID(ctx context.Context, id int) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := r.db.WithContext(ctx).Where("to = ?", id).Find(&notifications).Error; err != nil {
		return nil, err
	}

	return notifications, nil

}

func (r *notificationRepo) MarkAsRead(ctx context.Context, id int) error {
	var notification models.Notification

	if err := r.db.WithContext(ctx).First(&notification, id).Error; err != nil {
		return err
	}

	notification.Read = true
	if err := r.db.WithContext(ctx).Save(notification).Error; err != nil {
		return err
	}

	return nil
}

// home notifications
func (r *notificationRepo) CreateHomeNotification(ctx context.Context, n *models.HomeNotification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *notificationRepo) FindByHomeID(ctx context.Context, id int) ([]models.HomeNotification, error) {
	var notifications []models.HomeNotification
	if err := r.db.WithContext(ctx).Where("to = ?", id).Find(&notifications).Error; err != nil {
		return nil, err
	}

	return notifications, nil

}

func (r *notificationRepo) MarkAsReadForHomeNotification(ctx context.Context, id int) error {
	var notification models.Notification

	if err := r.db.WithContext(ctx).First(&notification, id).Error; err != nil {
		return err
	}

	notification.Read = true
	if err := r.db.WithContext(ctx).Save(notification).Error; err != nil {
		return err
	}

	return nil
}

