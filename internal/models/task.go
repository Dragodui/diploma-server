package models

import "time"

type Task struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	HomeID       int       `json:"home_id"`
	RoomID       *int      `json:"room_id"`
	Name         string    `gorm:"not null;size:64" json:"name"`
	Description  string    `gorm:"not null" json:"description"`
	ScheduleType string    `gorm:"not null;size:64" json:"schedule_type"`
	CreatedAt    time.Time `json:"created_at"`

	// relations
	Home            *Home            `gorm:"foreignKey:HomeID;constraint:OnDelete:CASCADE" json:"home,omitempty"`
	Room            *Room            `gorm:"foreignKey:RoomID;constraint:OnDelete:SET NULL" json:"room,omitempty"`
	TaskAssignments []TaskAssignment `gorm:"foreignKey:TaskID" json:"assignments,omitempty"`
}

type CreateTaskRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	ScheduleType string `json:"schedule_type"`
	HomeID       int    `json:"home_id"`
	RoomID       *int   `json:"room_id,omitempty"`
}

type ReassignRoomRequest struct {
	TaskID int `json:"task_id"`
	RoomID int `json:"room_id"`
}
