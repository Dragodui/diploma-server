package models

import "time"

type ShoppingCategory struct {
	ID        int       `gorm:"autoIncrement; primaryKey" json:"id"`
	HomeID    int       `json:"home_id"`
	Name      string    `json:"name"`
	Icon      *string   `json:"icon"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// relations
	Home  *Home          `gorm:"foreignKey:HomeID;constraint;OnDelete:CASCADE" json:"home,omitempty"`
	Items []ShoppingItem `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

type CreateCategoryRequest struct {
	Name   string  `json:"name"`
	Icon   *string `json:"icon"`
}

type UpdateShoppingCategoryRequest struct {
	Name       *string `json:"name"`
	Icon       *string `json:"icon"`
}
