package services

import (
	"fmt"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/services/homeassistant"
	"github.com/redis/go-redis/v9"
)

type ISmartHomeService interface {
	// Config management
	ConnectHA(homeID int, url, token string) error
	DisconnectHA(homeID int) error
	GetHAConfig(homeID int) (*models.HomeAssistantConfig, error)
	TestConnection(homeID int) error

	// Device management
	AddDevice(homeID int, entityID, name string, deviceType string, roomID *int, icon *string) error
	RemoveDevice(deviceID int) error
	UpdateDevice(deviceID int, name string, roomID *int, icon *string) error
	GetDevices(homeID int) ([]models.SmartDevice, error)
	GetDevicesByRoom(roomID int) ([]models.SmartDevice, error)
	GetDeviceByID(deviceID int) (*models.SmartDevice, error)

	// Device control & state
	GetDeviceState(homeID int, entityID string) (*homeassistant.HAState, error)
	GetAllStates(homeID int) ([]homeassistant.HAState, error)
	ControlDevice(homeID int, entityID string, service string, data map[string]interface{}) error

	// Discovery
	DiscoverDevices(homeID int) ([]homeassistant.HAState, error)
}

type SmartHomeService struct {
	repo  repository.SmartHomeRepository
	cache *redis.Client
}

func NewSmartHomeService(repo repository.SmartHomeRepository, cache *redis.Client) ISmartHomeService {
	return &SmartHomeService{repo: repo, cache: cache}
}

// getHAClient creates a Home Assistant client for a given home
func (s *SmartHomeService) getHAClient(homeID int) (*homeassistant.HAClient, error) {
	config, err := s.repo.GetConfigByHomeID(homeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get HA config: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("Home Assistant not connected")
	}
	if !config.IsActive {
		return nil, fmt.Errorf("Home Assistant connection is inactive")
	}

	return homeassistant.NewHAClient(config.URL, config.Token), nil
}

// Config management

func (s *SmartHomeService) ConnectHA(homeID int, url, token string) error {
	// Test connection first
	client := homeassistant.NewHAClient(url, token)
	if err := client.CheckConnection(); err != nil {
		return fmt.Errorf("failed to connect to Home Assistant: %w", err)
	}

	// Check if config already exists
	existing, err := s.repo.GetConfigByHomeID(homeID)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update existing config
		existing.URL = url
		existing.Token = token
		existing.IsActive = true
		return s.repo.UpdateConfig(existing)
	}

	// Create new config
	config := &models.HomeAssistantConfig{
		HomeID:   homeID,
		URL:      url,
		Token:    token,
		IsActive: true,
	}
	return s.repo.CreateConfig(config)
}

func (s *SmartHomeService) DisconnectHA(homeID int) error {
	return s.repo.DeleteConfig(homeID)
}

func (s *SmartHomeService) GetHAConfig(homeID int) (*models.HomeAssistantConfig, error) {
	return s.repo.GetConfigByHomeID(homeID)
}

func (s *SmartHomeService) TestConnection(homeID int) error {
	client, err := s.getHAClient(homeID)
	if err != nil {
		return err
	}
	return client.CheckConnection()
}

// Device management

func (s *SmartHomeService) AddDevice(homeID int, entityID, name string, deviceType string, roomID *int, icon *string) error {
	// Check if device already exists
	existing, err := s.repo.GetDeviceByEntityID(homeID, entityID)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("device with entity_id %s already added", entityID)
	}

	// Verify entity exists in HA
	client, err := s.getHAClient(homeID)
	if err != nil {
		return err
	}

	state, err := client.GetState(entityID)
	if err != nil {
		return fmt.Errorf("entity not found in Home Assistant: %w", err)
	}

	// Use friendly name if no name provided
	if name == "" {
		name = state.GetFriendlyName()
	}

	// Determine device type from entity_id if not provided
	if deviceType == "" {
		deviceType = homeassistant.GetEntityDomain(entityID)
	}

	device := &models.SmartDevice{
		HomeID:   homeID,
		RoomID:   roomID,
		EntityID: entityID,
		Name:     name,
		Type:     deviceType,
		Icon:     icon,
	}

	return s.repo.CreateDevice(device)
}

func (s *SmartHomeService) RemoveDevice(deviceID int) error {
	return s.repo.DeleteDevice(deviceID)
}

func (s *SmartHomeService) UpdateDevice(deviceID int, name string, roomID *int, icon *string) error {
	device, err := s.repo.GetDeviceByID(deviceID)
	if err != nil {
		return err
	}
	if device == nil {
		return fmt.Errorf("device not found")
	}

	device.Name = name
	device.RoomID = roomID
	device.Icon = icon

	return s.repo.UpdateDevice(device)
}

func (s *SmartHomeService) GetDevices(homeID int) ([]models.SmartDevice, error) {
	return s.repo.GetDevicesByHomeID(homeID)
}

func (s *SmartHomeService) GetDevicesByRoom(roomID int) ([]models.SmartDevice, error) {
	return s.repo.GetDevicesByRoomID(roomID)
}

func (s *SmartHomeService) GetDeviceByID(deviceID int) (*models.SmartDevice, error) {
	return s.repo.GetDeviceByID(deviceID)
}

// Device control & state

func (s *SmartHomeService) GetDeviceState(homeID int, entityID string) (*homeassistant.HAState, error) {
	client, err := s.getHAClient(homeID)
	if err != nil {
		return nil, err
	}
	return client.GetState(entityID)
}

func (s *SmartHomeService) GetAllStates(homeID int) ([]homeassistant.HAState, error) {
	client, err := s.getHAClient(homeID)
	if err != nil {
		return nil, err
	}

	// Get all states from HA
	allStates, err := client.GetStates()
	if err != nil {
		return nil, err
	}

	// Get added devices for this home
	devices, err := s.repo.GetDevicesByHomeID(homeID)
	if err != nil {
		return nil, err
	}

	// Create a set of added entity IDs
	addedEntities := make(map[string]bool)
	for _, d := range devices {
		addedEntities[d.EntityID] = true
	}

	// Filter states to only include added devices
	var filteredStates []homeassistant.HAState
	for _, state := range allStates {
		if addedEntities[state.EntityID] {
			filteredStates = append(filteredStates, state)
		}
	}

	return filteredStates, nil
}

func (s *SmartHomeService) ControlDevice(homeID int, entityID string, service string, data map[string]interface{}) error {
	client, err := s.getHAClient(homeID)
	if err != nil {
		return err
	}
	return client.CallServiceForEntity(entityID, service, data)
}

// Discovery

func (s *SmartHomeService) DiscoverDevices(homeID int) ([]homeassistant.HAState, error) {
	client, err := s.getHAClient(homeID)
	if err != nil {
		return nil, err
	}

	states, err := client.GetStates()
	if err != nil {
		return nil, err
	}

	// Filter out system/internal entities
	var devices []homeassistant.HAState
	for _, state := range states {
		domain := homeassistant.GetEntityDomain(state.EntityID)
		// Include common device domains
		switch domain {
		case "light", "switch", "climate", "fan", "cover", "lock", "media_player",
			"vacuum", "camera", "sensor", "binary_sensor", "input_boolean", "scene":
			if state.IsAvailable() {
				devices = append(devices, state)
			}
		}
	}

	return devices, nil
}
