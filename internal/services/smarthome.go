package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/services/homeassistant"
	"github.com/redis/go-redis/v9"
)

type ISmartHomeService interface {
	// Config management
	ConnectHA(ctx context.Context, homeID int, url, token string) error
	DisconnectHA(ctx context.Context, homeID int) error
	GetHAConfig(ctx context.Context, homeID int) (*models.HomeAssistantConfig, error)
	TestConnection(ctx context.Context, homeID int) error

	// Device management
	AddDevice(ctx context.Context, homeID int, entityID, name string, deviceType string, roomID *int, icon *string) error
	RemoveDevice(ctx context.Context, deviceID int) error
	UpdateDevice(ctx context.Context, deviceID, homeID int, name string, roomID *int, icon *string) error
	GetDevices(ctx context.Context, homeID int) ([]models.SmartDevice, error)
	GetDevicesByRoom(ctx context.Context, roomID int) ([]models.SmartDevice, error)
	GetDeviceByID(ctx context.Context, deviceID, homeID int) (*models.SmartDevice, error)

	// Device control & state
	GetDeviceState(ctx context.Context, homeID int, entityID string) (*homeassistant.HAState, error)
	GetAllStates(ctx context.Context, homeID int) ([]homeassistant.HAState, error)
	ControlDevice(ctx context.Context, homeID int, entityID string, service string, data map[string]interface{}) error

	// Discovery
	DiscoverDevices(ctx context.Context, homeID int) ([]homeassistant.HAState, error)
}

type SmartHomeService struct {
	repo  repository.SmartHomeRepository
	cache *redis.Client
}

func NewSmartHomeService(repo repository.SmartHomeRepository, cache *redis.Client) ISmartHomeService {
	return &SmartHomeService{repo: repo, cache: cache}
}

// getHAClient creates a Home Assistant client for a given home
func (s *SmartHomeService) getHAClient(ctx context.Context, homeID int) (*homeassistant.HAClient, error) {
	config, err := s.repo.GetConfigByHomeID(ctx, homeID)
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

func (s *SmartHomeService) ConnectHA(ctx context.Context, homeID int, url, token string) error {
	// Test Connection first
	client := homeassistant.NewHAClient(url, token)
	if err := client.CheckConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to Home Assistant: %w", err)
	}

	// Check if config already exists
	existing, err := s.repo.GetConfigByHomeID(ctx, homeID)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update existing config
		existing.URL = url
		existing.Token = token
		existing.IsActive = true
		return s.repo.UpdateConfig(ctx, existing)
	}

	// Create new config
	config := &models.HomeAssistantConfig{
		HomeID:   homeID,
		URL:      url,
		Token:    token,
		IsActive: true,
	}
	return s.repo.CreateConfig(ctx, config)
}

func (s *SmartHomeService) DisconnectHA(ctx context.Context, homeID int) error {
	return s.repo.DeleteConfig(ctx, homeID)
}

func (s *SmartHomeService) GetHAConfig(ctx context.Context, homeID int) (*models.HomeAssistantConfig, error) {
	return s.repo.GetConfigByHomeID(ctx, homeID)
}

func (s *SmartHomeService) TestConnection(ctx context.Context, homeID int) error {
	client, err := s.getHAClient(ctx, homeID)
	if err != nil {
		return err
	}
	return client.CheckConnection(ctx)
}

// Device management

func (s *SmartHomeService) AddDevice(ctx context.Context, homeID int, entityID, name string, deviceType string, roomID *int, icon *string) error {
	// Check if device already exists
	existing, err := s.repo.GetDeviceByEntityID(ctx, homeID, entityID)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("device with entity_id %s already added", entityID)
	}

	// Verify entity exists in HA
	client, err := s.getHAClient(ctx, homeID)
	if err != nil {
		return err
	}

	state, err := client.GetState(ctx, entityID)
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

	return s.repo.CreateDevice(ctx, device)
}

func (s *SmartHomeService) RemoveDevice(ctx context.Context, deviceID int) error {
	return s.repo.DeleteDevice(ctx, deviceID)
}

func (s *SmartHomeService) UpdateDevice(ctx context.Context, deviceID, homeID int, name string, roomID *int, icon *string) error {
	device, err := s.GetDeviceByID(ctx, deviceID, homeID)
	if err != nil {
		return err
	}
	if device == nil {
		return fmt.Errorf("device not found")
	}

	device.Name = name
	device.RoomID = roomID
	device.Icon = icon

	return s.repo.UpdateDevice(ctx, device)
}

func (s *SmartHomeService) GetDevices(ctx context.Context, homeID int) ([]models.SmartDevice, error) {
	return s.repo.GetDevicesByHomeID(ctx, homeID)
}

func (s *SmartHomeService) GetDevicesByRoom(ctx context.Context, roomID int) ([]models.SmartDevice, error) {
	return s.repo.GetDevicesByRoomID(ctx, roomID)
}

func (s *SmartHomeService) GetDeviceByID(ctx context.Context, deviceID, homeID int) (*models.SmartDevice, error) {
	device, err := s.repo.GetDeviceByID(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	if device.HomeID != homeID {
		return nil, errors.New("device is not from your home")
	}

	return device, nil
}

// Device control & state

func (s *SmartHomeService) GetDeviceState(ctx context.Context, homeID int, entityID string) (*homeassistant.HAState, error) {
	client, err := s.getHAClient(ctx, homeID)
	if err != nil {
		return nil, err
	}
	return client.GetState(ctx, entityID)
}

func (s *SmartHomeService) GetAllStates(ctx context.Context, homeID int) ([]homeassistant.HAState, error) {
	client, err := s.getHAClient(ctx, homeID)
	if err != nil {
		return nil, err
	}

	// Get all states from HA
	allStates, err := client.GetStates(ctx)
	if err != nil {
		return nil, err
	}

	// Get added devices for this home
	devices, err := s.repo.GetDevicesByHomeID(ctx, homeID)
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

func (s *SmartHomeService) ControlDevice(ctx context.Context, homeID int, entityID string, service string, data map[string]interface{}) error {
	client, err := s.getHAClient(ctx, homeID)
	if err != nil {
		return err
	}
	return client.CallServiceForEntity(ctx, entityID, service, data)
}

// Discovery

func (s *SmartHomeService) DiscoverDevices(ctx context.Context, homeID int) ([]homeassistant.HAState, error) {
	client, err := s.getHAClient(ctx, homeID)
	if err != nil {
		return nil, err
	}

	states, err := client.GetStates(ctx)
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
