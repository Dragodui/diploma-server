package models

import "time"

type ShoppingCategory struct {
	ID        int       `gorm:"autoIncrement; primaryKey" json:"id"`
	HomeID    int       `json:"home_id"`
	CreatedBy int       `json:"created_by"`
	Name      string    `json:"name"`
	Icon      *string   `json:"icon"`
	Color     string    `gorm:"size:32;default:'#D8D4FC'" json:"color"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// relations
	Home    *Home          `gorm:"foreignKey:HomeID;constraint;OnDelete:CASCADE" json:"home,omitempty"`
	Creator *User          `gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE" json:"creator,omitempty"`
	Items   []ShoppingItem `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

type CreateCategoryRequest struct {
	Name  string  `json:"name" validate:"required,min=1,max=64"`
	Icon  *string `json:"icon"`
	Color string  `json:"color" validate:"omitempty,hexcolor"`
}

type UpdateShoppingCategoryRequest struct {
	Name  *string `json:"name"`
	Icon  *string `json:"icon"`
	Color *string `json:"color" validate:"omitempty,hexcolor"`
}
