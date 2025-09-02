package models

import "time"

type HomeMembership struct {
	ID        int       `gorm:"autoIncrement; primaryKey" json:"id"`
	HomeID    int       `gorm:"not null" json:"home_id"`
	UserID    int       `gorm:"not null" json:"user_id"`
	Role      string    `gorm:"size:64;not null" json:"role"`
	JoinedAt  time.Time `gorm:"autoCreateTime" json:"joined_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// relations
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Home *Home `gorm:"foreignKey:HomeID;constraint:OnDelete:CASCADE" json:"home,omitempty"`
}

type JoinRequest struct {
	Code string `json:"code" validate:"required"`
}

type LeaveRequest struct {
	HomeID string `json:"home_id" validate:"required"`
}

type RemoveMemberRequest struct {
	HomeID string `json:"home_id" validate:"required"`
	UserID string `json:"user_id" validate:"required"`
}
