package models

import "time"

type Task struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	HomeID       int       `json:"home_id"`
	Name         string    `gorm:"not null;size:64" json:"name"`
	Description  string    `gorm:"not null" json:"description"`
	ScheduleType string    `gorm:"not null;size:64" json:"schedule_type"`
	CreatedAt    time.Time `json:"created_at"`

	// relations
	Home *Home `gorm:"foreignKey:HomeID;constraint:OnDelete:CASCADE" json:"home,omitempty"`
}
