package models

import "time"

type Room struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	HomeID    int       `json:"home_id"`
	Name      string    `gorm:"not null;size:64" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	// relations
	Home  *Home  `gorm:"foreignKey:HomeID;constraint:OnDelete:CASCADE" json:"home,omitempty"`
	Tasks []Task `gorm:"foreignKey:RoomID" json:"tasks,omitempty"`
}

type CreateRoomRequest struct {
	HomeID int    `json:"home_id"`
	Name   string `json:"name"`
}
