package models

import "time"

type ShoppingItem struct {
	ID         int        `gorm:"autoIncrement; primaryKey; " json:"id"`
	CategoryID int        `json:"category_id"`
	Name       string     `json:"name"`
	UploadedBy int        `json:"added_by"`
	IsBought   bool       `json:"is_bought"`
	Image      *string    `json:"image"`
	Link       *string    `json:"link"`
	BoughtDate *time.Time `json:"bought_date"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// relations
	User *User `gorm:"foreignKey:UploadedBy;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

type CreateShoppingItemRequest struct {
	CategoryID int     `json:"category_id" validate:"required"`
	Name       string  `json:"name" validate:"required,min=3"`
	Image      *string `json:"image"`
	Link       *string `json:"link"`
}

type UpdateShoppingItemRequest struct {
	Name     *string    `json:"name,omitempty" validate:"omitempty,min=3"`
	Image    *string    `json:"image,omitempty"`
	Link     *string    `json:"link,omitempty"`
	IsBought *bool      `json:"is_bought,omitempty"`
	BoughtAt *time.Time `json:"bought_date,omitempty"`
}
