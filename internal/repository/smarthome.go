package repository

import (
	"errors"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type SmartHomeRepository interface {
	// Config operations
	CreateConfig(config *models.HomeAssistantConfig) error
	GetConfigByHomeID(homeID int) (*models.HomeAssistantConfig, error)
	UpdateConfig(config *models.HomeAssistantConfig) error
	DeleteConfig(homeID int) error

	// Device operations
	CreateDevice(device *models.SmartDevice) error
	GetDeviceByID(id int) (*models.SmartDevice, error)
	GetDevicesByHomeID(homeID int) ([]models.SmartDevice, error)
	GetDevicesByRoomID(roomID int) ([]models.SmartDevice, error)
	GetDeviceByEntityID(homeID int, entityID string) (*models.SmartDevice, error)
	UpdateDevice(device *models.SmartDevice) error
	DeleteDevice(id int) error
}

type smartHomeRepo struct {
	db *gorm.DB
}

func NewSmartHomeRepository(db *gorm.DB) SmartHomeRepository {
	return &smartHomeRepo{db}
}

// Config operations

func (r *smartHomeRepo) CreateConfig(config *models.HomeAssistantConfig) error {
	return r.db.Create(config).Error
}

func (r *smartHomeRepo) GetConfigByHomeID(homeID int) (*models.HomeAssistantConfig, error) {
	var config models.HomeAssistantConfig
	if err := r.db.Where("home_id = ?", homeID).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *smartHomeRepo) UpdateConfig(config *models.HomeAssistantConfig) error {
	return r.db.Save(config).Error
}

func (r *smartHomeRepo) DeleteConfig(homeID int) error {
	return r.db.Where("home_id = ?", homeID).Delete(&models.HomeAssistantConfig{}).Error
}

// Device operations

func (r *smartHomeRepo) CreateDevice(device *models.SmartDevice) error {
	return r.db.Create(device).Error
}

func (r *smartHomeRepo) GetDeviceByID(id int) (*models.SmartDevice, error) {
	var device models.SmartDevice
	if err := r.db.Preload("Room").First(&device, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

func (r *smartHomeRepo) GetDevicesByHomeID(homeID int) ([]models.SmartDevice, error) {
	var devices []models.SmartDevice
	if err := r.db.Preload("Room").Where("home_id = ?", homeID).Order("created_at DESC").Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *smartHomeRepo) GetDevicesByRoomID(roomID int) ([]models.SmartDevice, error) {
	var devices []models.SmartDevice
	if err := r.db.Preload("Room").Where("room_id = ?", roomID).Order("created_at DESC").Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *smartHomeRepo) GetDeviceByEntityID(homeID int, entityID string) (*models.SmartDevice, error) {
	var device models.SmartDevice
	if err := r.db.Where("home_id = ? AND entity_id = ?", homeID, entityID).First(&device).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

func (r *smartHomeRepo) UpdateDevice(device *models.SmartDevice) error {
	return r.db.Save(device).Error
}

func (r *smartHomeRepo) DeleteDevice(id int) error {
	return r.db.Delete(&models.SmartDevice{}, id).Error
}
