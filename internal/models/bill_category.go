package models

import "time"

type BillCategory struct {
	ID        int       `gorm:"autoIncrement; primaryKey" json:"id"`
	HomeID    int       `gorm:"not null" json:"home_id"`
	Name      string    `gorm:"not null;size:64" json:"name"`
	Color     string    `gorm:"size:32;default:'#FBEB9E'" json:"color"` // Hex color
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// relations
	Home *Home `gorm:"foreignKey:HomeID;constraint:OnDelete:CASCADE" json:"home,omitempty"`
}

type CreateBillCategoryRequest struct {
	Name   string `json:"name" validate:"required,min=2,max=64"`
	Color  string `json:"color" validate:"omitempty,hexcolor"`
	HomeID int    `json:"home_id" validate:"required"`
}
