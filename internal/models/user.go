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
	ID           int       `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"size:64;not null;unique" json:"email"`
	Name         string    `gorm:"size:64;not null" json:"name"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Avatar       string    `json:"avatar"`
	CreatedAt    time.Time `json:"created_at"`
}
