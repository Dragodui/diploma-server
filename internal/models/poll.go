package models

import "time"

type Option struct {
	ID        int       `gorm:"autoIncrement; primaryKey" json:"id"`
	PollID    int       `json:"poll_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	Votes []Vote `gorm:"foreignKey:OptionID;constraint:OnDelete:CASCADE" json:"votes,omitempty"`
	Poll  *Poll  `gorm:"foreignKey:PollID;constraint;OnDelete:CASCADE" json:"poll,omitempty"`
}

type Vote struct {
	ID       int `gorm:"autoIncrement; primaryKey" json:"id"`
	UserID   int `json:"user_id"`
	OptionID int `json:"option_id"`

	// relations
	Option *Option `gorm:"foreignKey:OptionID;constraint;OnDelete:CASCADE" json:"option,omitempty"`
	User   *User   `gorm:"foreignKey:UserID;constraint;OnDelete:CASCADE" json:"user,omitempty"`
}

type Poll struct {
	ID       int    `gorm:"autoIncrement; primaryKey" json:"id"`
	HomeID   int    `json:"home_id"`
	Question string `json:"question"`
	Type     string `gorm:"default:public" json:"type"`                  // public/anonymous
	Status   string `gorm:"not null;size:64;default:open" json:"status"` // open/closed

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// relations
	Home    *Home    `gorm:"foreignKey:HomeID;constraint;OnDelete:CASCADE" json:"home,omitempty"`
	Options []Option `gorm:"foreignKey:PollID;constraint:OnDelete:CASCADE" json:"options,omitempty"`
}

type OptionRequest struct {
	Title string `json:"title" validate:"required"`
}

type CreatePollRequest struct {
	Question string          `json:"question" validate:"required"`
	Type     string          `json:"type" validate:"required,oneof=single multiple"`
	Options  []OptionRequest `json:"options" validate:"min=1,dive"`
}

type VoteRequest struct {
	OptionID int `json:"option_id" validate:"required"`
}
