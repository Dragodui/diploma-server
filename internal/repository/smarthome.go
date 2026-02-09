package repository

import (
	"context"
	"errors"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type SmartHomeRepository interface {
	// Config operations
	CreateConfig(ctx context.Context, config *models.HomeAssistantConfig) error
	GetConfigByHomeID(ctx context.Context, homeID int) (*models.HomeAssistantConfig, error)
	UpdateConfig(ctx context.Context, config *models.HomeAssistantConfig) error
	DeleteConfig(ctx context.Context, homeID int) error

	// Device operations
	CreateDevice(ctx context.Context, device *models.SmartDevice) error
	GetDeviceByID(ctx context.Context, id int) (*models.SmartDevice, error)
	GetDevicesByHomeID(ctx context.Context, homeID int) ([]models.SmartDevice, error)
	GetDevicesByRoomID(ctx context.Context, roomID int) ([]models.SmartDevice, error)
	GetDeviceByEntityID(ctx context.Context, homeID int, entityID string) (*models.SmartDevice, error)
	UpdateDevice(ctx context.Context, device *models.SmartDevice) error
	DeleteDevice(ctx context.Context, id int) error
}

type smartHomeRepo struct {
	db *gorm.DB
}

func NewSmartHomeRepository(db *gorm.DB) SmartHomeRepository {
	return &smartHomeRepo{db}
}

// Config operations

func (r *smartHomeRepo) CreateConfig(ctx context.Context, config *models.HomeAssistantConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

func (r *smartHomeRepo) GetConfigByHomeID(ctx context.Context, homeID int) (*models.HomeAssistantConfig, error) {
	var config models.HomeAssistantConfig
	if err := r.db.WithContext(ctx).Where("home_id = ?", homeID).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *smartHomeRepo) UpdateConfig(ctx context.Context, config *models.HomeAssistantConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *smartHomeRepo) DeleteConfig(ctx context.Context, homeID int) error {
	return r.db.WithContext(ctx).Where("home_id = ?", homeID).Delete(&models.HomeAssistantConfig{}).Error
}

// Device operations

func (r *smartHomeRepo) CreateDevice(ctx context.Context, device *models.SmartDevice) error {
	return r.db.WithContext(ctx).Create(device).Error
}

func (r *smartHomeRepo) GetDeviceByID(ctx context.Context, id int) (*models.SmartDevice, error) {
	var device models.SmartDevice
	if err := r.db.WithContext(ctx).Preload("Room").First(&device, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

func (r *smartHomeRepo) GetDevicesByHomeID(ctx context.Context, homeID int) ([]models.SmartDevice, error) {
	var devices []models.SmartDevice
	if err := r.db.WithContext(ctx).Preload("Room").Where("home_id = ?", homeID).Order("created_at DESC").Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *smartHomeRepo) GetDevicesByRoomID(ctx context.Context, roomID int) ([]models.SmartDevice, error) {
	var devices []models.SmartDevice
	if err := r.db.WithContext(ctx).Preload("Room").Where("room_id = ?", roomID).Order("created_at DESC").Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *smartHomeRepo) GetDeviceByEntityID(ctx context.Context, homeID int, entityID string) (*models.SmartDevice, error) {
	var device models.SmartDevice
	if err := r.db.WithContext(ctx).Where("home_id = ? AND entity_id = ?", homeID, entityID).First(&device).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

func (r *smartHomeRepo) UpdateDevice(ctx context.Context, device *models.SmartDevice) error {
	return r.db.WithContext(ctx).Save(device).Error
}

func (r *smartHomeRepo) DeleteDevice(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.SmartDevice{}, id).Error
}

