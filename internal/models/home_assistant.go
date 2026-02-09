package models

import "time"

// HomeAssistantConfig stores HA connection settings for a Home
type HomeAssistantConfig struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	HomeID    int       `gorm:"uniqueIndex;not null" json:"home_id"`
	URL       string    `gorm:"not null;size:512" json:"url"`
	Token     string    `gorm:"not null;size:512" json:"-"` // Hidden from JSON
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Home *Home `gorm:"foreignKey:HomeID;constraint:OnDelete:CASCADE" json:"home,omitempty"`
}

// SmartDevice represents a HA device assigned to a room
type SmartDevice struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	HomeID    int       `gorm:"not null;index" json:"home_id"`
	RoomID    *int      `gorm:"index" json:"room_id"`
	EntityID  string    `gorm:"not null;size:256" json:"entity_id"`
	Name      string    `gorm:"not null;size:128" json:"name"`
	Type      string    `gorm:"not null;size:64" json:"type"`
	Icon      *string   `gorm:"size:64" json:"icon"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	Home *Home `gorm:"foreignKey:HomeID;constraint:OnDelete:CASCADE" json:"home,omitempty"`
	Room *Room `gorm:"foreignKey:RoomID;constraint:OnDelete:SET NULL" json:"room,omitempty"`
}

// ConnectHARequest for connecting Home Assistant
type ConnectHARequest struct {
	URL   string `json:"url" validate:"required,url"`
	Token string `json:"token" validate:"required,min=10"`
}

// AddDeviceRequest for adding a device to the app
type AddDeviceRequest struct {
	EntityID string  `json:"entity_id" validate:"required"`
	Name     string  `json:"name" validate:"required,min=1,max=128"`
	RoomID   *int    `json:"room_id"`
	Icon     *string `json:"icon"`
}

// UpdateDeviceRequest for updating device details
type UpdateDeviceRequest struct {
	Name   string  `json:"name" validate:"required,min=1,max=128"`
	RoomID *int    `json:"room_id"`
	Icon   *string `json:"icon"`
}

// ControlDeviceRequest for controlling a device
type ControlDeviceRequest struct {
	Service string                 `json:"service" validate:"required"`
	Data    map[string]interface{} `json:"data"`
}
