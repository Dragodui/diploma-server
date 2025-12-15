package models

import (
	"time"

	"gorm.io/datatypes"
)

type Bill struct {
	ID             int            `gorm:"autoIncrement;  primaryKey" json:"id"`
	HomeID         int            `json:"home_id"`
	BillCategoryID *int           `json:"bill_category_id"`
	Type           string         `json:"type"` // Kept for backward compatibility or as fallback
	Payed          bool           `json:"is_payed"`
	PaymentDate    *time.Time     `json:"payment_date"`
	TotalAmount    float64        `json:"total_amount"`
	Start          time.Time      `json:"period_start"`
	End            time.Time      `json:"period_end"`
	UploadedBy     int            `json:"uploaded_by"`
	OCRData        datatypes.JSON `json:"ocr_data"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`

	//relations
	Home         *Home         `gorm:"foreignKey:HomeID;constraint:OnDelete:CASCADE" json:"home,omitempty"`
	User         *User         `gorm:"foreignKey:UploadedBy;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	BillCategory *BillCategory `gorm:"foreignKey:BillCategoryID;constraint:OnDelete:SET NULL" json:"bill_category,omitempty"`
}

type CreateBillRequest struct {
	BillType       string         `json:"type"` // Optional if CategoryID is provided
	BillCategoryID *int           `json:"bill_category_id"`
	TotalAmount    float64        `json:"total_amount" validate:"required,gte=0"`
	Start          time.Time      `json:"period_start" validate:"required"`
	End            time.Time      `json:"period_end" validate:"required"`
	OCRData        datatypes.JSON `json:"ocr_data" validate:"required"`
}
