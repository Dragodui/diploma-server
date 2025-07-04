package models

import "time"

type TaskAssignment struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	HomeID       int       `gorm:"not null" json:"home_id"`
	UserID       int       `gorm:"not null" json:"user_id"`
	Status       string    `gorm:"not null;size:64" json:"status"`
	AssignedDate time.Time `json:"assigned_date"`

	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Home *Home `gorm:"foreignKey:HomeID;constraint:OnDelete:CASCADE" json:"home,omitempty"`
}
