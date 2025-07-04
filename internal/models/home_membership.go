package models

import "time"

type HomeMembership struct {
	ID       int       `gorm:"primaryKey" json:"id"`
	HomeID   int       `gorm:"not null" json:"home_id"`
	UserID   int       `gorm:"not null" json:"user_id"`
	Role     string    `gorm:"size:64;not null" json:"role"`
	JoinedAt time.Time `json:"joined_at"`

	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Home *Home `gorm:"foreignKey:HomeID;constraint:OnDelete:CASCADE" json:"home,omitempty"`
}

type JoinRequest struct {
	Code string
}

type LeaveRequest struct {
	HomeID string
}

type RemoveMemberRequest struct {
	HomeID string
	UserID string
}
