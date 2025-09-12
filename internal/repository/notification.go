package repository

import (
	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(n *models.Notification) error
	FindByUserID(id int) ([]models.Notification, error)
	MarkAsRead(id int) error

	CreateHomeNotification(n *models.HomeNotification) error
	FindByHomeID(id int) ([]models.HomeNotification, error)
	MarkAsReadForHomeNotification(id int) error
}

type notificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepo{db}
}

// user notifications
func (r *notificationRepo) Create(n *models.Notification) error {
	return r.db.Create(n).Error
}

func (r *notificationRepo) FindByUserID(id int) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := r.db.Where("to = ?", id).Find(&notifications).Error; err != nil {
		return nil, err
	}

	return notifications, nil

}

func (r *notificationRepo) MarkAsRead(id int) error {
	var notification models.Notification

	if err := r.db.First(&notification, id).Error; err != nil {
		return err
	}

	notification.Read = true
	if err := r.db.Save(notification).Error; err != nil {
		return err
	}

	return nil
}

// home notifications
func (r *notificationRepo) CreateHomeNotification(n *models.Notification) error {
	return r.db.Create(n).Error
}

func (r *notificationRepo) FindByHomeID(id int) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := r.db.Where("to = ?", id).Find(&notifications).Error; err != nil {
		return nil, err
	}

	return notifications, nil

}

func (r *notificationRepo) MarkAsReadForHomeNotification(id int) error {
	var notification models.Notification

	if err := r.db.First(&notification, id).Error; err != nil {
		return err
	}

	notification.Read = true
	if err := r.db.Save(notification).Error; err != nil {
		return err
	}

	return nil
}

