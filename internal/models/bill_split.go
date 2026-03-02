package models

type BillSplit struct {
	ID     int     `gorm:"autoIncrement;primaryKey" json:"id"`
	BillID int     `gorm:"not null" json:"bill_id"`
	UserID int     `gorm:"not null" json:"user_id"`
	Amount float64 `gorm:"not null" json:"amount"`
	Paid   bool    `gorm:"default:false" json:"paid"`

	Bill *Bill `gorm:"foreignKey:BillID;constraint:OnDelete:CASCADE" json:"bill,omitempty"`
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

type SplitInput struct {
	UserID int     `json:"user_id"`
	Amount float64 `json:"amount"`
}

type UpdateSplitsRequest struct {
	Splits []SplitInput `json:"splits"`
}
