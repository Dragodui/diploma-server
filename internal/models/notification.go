package models

import "time"

type Notification struct {
	ID          int       `gorm:"autoIncrement;primaryKey" json:"id"`
	From        *int      `json:"from"` // userID
	To          int       `json:"to"`   // userID
	Description string    `json:"description"`
	Read        bool      `json:"read"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	// relations
	UserFrom *User `gorm:"foreignKey:From;constraint:OnDelete:CASCADE" json:"user_from,omitempty"`
	UserTo   User  `gorm:"foreignKey:To;constraint:OnDelete:CASCADE" json:"user_to,omitempty"`
}

type HomeNotification struct {
	ID          int       `gorm:"autoIncrement;primaryKey" json:"id"`
	From        *int      `json:"from"` // userID
	HomeID      int       `json:"home_id"`
	Description string    `json:"description"`
	Read        bool      `json:"read"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	// relations
	UserFrom *User `gorm:"foreignKey:From;constraint:OnDelete:CASCADE" json:"user_from,omitempty"`
	Home     Home  `gorm:"foreignKey:HomeID;constraint:OnDelete:CASCADE" json:"home,omitempty"`
}
