package models

import (
	"time"
)

type TaskAssignment struct {
	ID           int        `gorm:"primaryKey" json:"id"`
	TaskID       int        `gorm:"not null" json:"task_id"`
	HomeID       int        `gorm:"not null" json:"home_id"`
	UserID       int        `gorm:"not null" json:"user_id"`
	Status       string     `gorm:"not null;size:64;default:assigned" json:"status"`
	AssignedDate time.Time  `json:"assigned_date"`
	CompleteDate *time.Time `json:"complete_date"`

	// relations
	Task *Task `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"task,omitempty"`
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Home *Home `gorm:"foreignKey:HomeID;constraint:OnDelete:CASCADE" json:"home,omitempty"`
}

type AssignUserRequest struct {
	TaskID int       `json:"task_id"`
	HomeID int       `json:"home_id"`
	UserID int       `json:"user_id"`
	Date   time.Time `json:"date"`
}

type UserIDRequest struct {
	UserID int `json:"user_id"`
}

type AssignmentIDRequest struct {
	AssignmentID int `json:"assignment_id"`
}
