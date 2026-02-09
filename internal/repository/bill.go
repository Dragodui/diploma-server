package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type BillRepository interface {
	Create(ctx context.Context, b *models.Bill) error
	FindByID(ctx context.Context, id int) (*models.Bill, error)
	FindByHomeID(ctx context.Context, homeID int) ([]models.Bill, error)
	Delete(ctx context.Context, id int) error
	MarkPayed(ctx context.Context, id int) error
}

type billRepo struct {
	db *gorm.DB
}

func NewBillRepository(db *gorm.DB) BillRepository {
	return &billRepo{db}
}

func (r *billRepo) Create(ctx context.Context, b *models.Bill) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *billRepo) FindByID(ctx context.Context, id int) (*models.Bill, error) {
	var bill models.Bill

	if err := r.db.WithContext(ctx).First(&bill, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &bill, nil
}

func (r *billRepo) FindByHomeID(ctx context.Context, homeID int) ([]models.Bill, error) {
	var bills []models.Bill

	if err := r.db.WithContext(ctx).Where("home_id = ?", homeID).Order("created_at DESC").Find(&bills).Error; err != nil {
		return nil, err
	}

	return bills, nil
}

func (r *billRepo) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Bill{}, id).Error
}

func (r *billRepo) MarkPayed(ctx context.Context, id int) error {
	var bill models.Bill
	if err := r.db.WithContext(ctx).First(&bill, id).Error; err != nil {
		return err
	}

	bill.Payed = true
	now := time.Now()
	bill.PaymentDate = &now
	if err := r.db.WithContext(ctx).Save(&bill).Error; err != nil {
		return err
	}

	return nil
}
