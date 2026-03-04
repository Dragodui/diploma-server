package models

import "time"

type TaskSchedule struct {
	ID                   int       `gorm:"autoIncrement; primaryKey" json:"id"`
	TaskID               int       `gorm:"not null;uniqueIndex" json:"task_id"`
	RecurrenceType       string    `gorm:"not null;size:32" json:"recurrence_type"` // daily, weekly, monthly
	RotationUserIDs      string    `gorm:"not null;type:text" json:"rotation_user_ids"` // JSON array of user IDs in order
	CurrentRotationIndex int       `gorm:"not null;default:0" json:"current_rotation_index"`
	NextRunDate          time.Time `gorm:"not null" json:"next_run_date"`
	IsActive             bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`

	// relations
	Task *Task `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"task,omitempty"`
}

type CreateTaskScheduleRequest struct {
	TaskID         int    `json:"task_id"`
	HomeID         int    `json:"home_id"`
	RecurrenceType string `json:"recurrence_type"` // daily, weekly, monthly
	UserIDs        []int  `json:"user_ids"`
}
