package models

import (
	"time"

	"gorm.io/datatypes"
)

type Bill struct {
	ID          int            `gorm:"primaryKey" json:"id"`
	HomeID      int            `json:"home_id"`
	Type        string         `json:"type"`
	Payed       bool           `json:"is_payed"`
	PaymentDate *time.Time     `json:"payment_date"`
	TotalAmount float64            `json:"total_amount"`
	Start       time.Time      `json:"period_start"`
	End         time.Time      `json:"period_end"`
	UploadedBy  int            `json:"uploaded_by"`
	OCRData     datatypes.JSON `json:"ocr_data"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`

	//relations
	Home *Home `gorm:"foreignKey:HomeID;constraint:OnDelete:CASCADE" json:"home,omitempty"`
	User *User `gorm:"foreignKey:UploadedBy;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

type CreateBillRequest struct {
	BillType    string         `json:"type" validate:"required,min=3"`
	TotalAmount float64            `json:"total_amount" validate:"required,gte=0"`
	HomeID      int            `json:"home_id" validate:"required"`
	Start       time.Time      `json:"period_start" validate:"required"`
	End         time.Time      `json:"period_end" validate:"required"`
	OCRData     datatypes.JSON `json:"ocr_data" validate:"required"`
}
