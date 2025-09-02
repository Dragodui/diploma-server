package models

import "time"

type RegisterInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=3"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type Login struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type User struct {
	ID              int        `gorm:"autoIncrement; primaryKey" json:"id"`
	Email           string     `gorm:"size:64;not null;unique" json:"email"`
	EmailVerified   bool       `db:"email_verified"`
	VerifyToken     *string    `db:"verify_token"`
	VerifyExpiresAt *time.Time `db:"verify_expires_at"`
	ResetToken      *string    `db:"reset_token"`
	ResetExpiresAt  *time.Time `db:"reset_expires_at"`
	Name            string     `gorm:"size:64;not null" json:"name"`
	PasswordHash    string     `gorm:"not null" json:"-"`
	Avatar          string     `json:"avatar"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// relations
	Memberships     []HomeMembership `gorm:"foreignKey:UserID" json:"memberships,omitempty"`
	TaskAssignments []TaskAssignment `gorm:"foreignKey:UserID" json:"task_assignments,omitempty"`
}
