package repository

import (
	"errors"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type BillRepository interface {
	Create(b *models.Bill) error
	FindByID(id int) (*models.Bill, error)
	Delete(id int) error
	MarkPayed(id int) error
}

type billRepo struct {
	db *gorm.DB
}

func NewBillRepository(db *gorm.DB) BillRepository {
	return &billRepo{db}
}

func (r *billRepo) Create(b *models.Bill) error {
	return r.db.Create(b).Error
}

func (r *billRepo) FindByID(id int) (*models.Bill, error) {
	var bill models.Bill

	if err := r.db.First(&bill, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &bill, nil

}

func (r *billRepo) Delete(id int) error {
	return r.db.Delete(&models.Bill{}, id).Error
}

func (r *billRepo) MarkPayed(id int) error {
	var bill models.Bill
	if err := r.db.First(&bill, id).Error; err != nil {
		return err
	}

	bill.Payed = true
	now := time.Now()
	bill.PaymentDate = &now
	if err := r.db.Save(&bill).Error; err != nil {
		return err
	}

	return nil
}
