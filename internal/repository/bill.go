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
	FindByHomeID(ctx context.Context, homeID int, categoryID *int) ([]models.Bill, error)
	Delete(ctx context.Context, id int) error
	MarkPayed(ctx context.Context, id int) error
	CreateSplits(ctx context.Context, billID int, splits []models.BillSplit) error
	UpdateSplits(ctx context.Context, billID int, splits []models.BillSplit) error
	MarkSplitPaid(ctx context.Context, splitID int) error
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

	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("BillSplits").
		Preload("BillSplits.User").
		Preload("BillCategory").
		First(&bill, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &bill, nil
}

func (r *billRepo) FindByHomeID(ctx context.Context, homeID int, categoryID *int) ([]models.Bill, error) {
	var bills []models.Bill

	query := r.db.WithContext(ctx).Where("home_id = ?", homeID)
	if categoryID != nil {
		query = query.Where("bill_category_id = ?", *categoryID)
	}

	if err := query.
		Preload("User").
		Preload("BillSplits").
		Preload("BillSplits.User").
		Preload("BillCategory").
		Order("created_at DESC").
		Find(&bills).Error; err != nil {
		return nil, err
	}

	return bills, nil
}

func (r *billRepo) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Where("bill_id = ?", id).Delete(&models.BillSplit{}).Error; err != nil {
		return err
	}
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

func (r *billRepo) CreateSplits(ctx context.Context, billID int, splits []models.BillSplit) error {
	for i := range splits {
		splits[i].BillID = billID
	}
	return r.db.WithContext(ctx).Create(&splits).Error
}

func (r *billRepo) UpdateSplits(ctx context.Context, billID int, splits []models.BillSplit) error {
	// Delete existing splits and create new ones
	if err := r.db.WithContext(ctx).Where("bill_id = ?", billID).Delete(&models.BillSplit{}).Error; err != nil {
		return err
	}
	if len(splits) == 0 {
		return nil
	}
	for i := range splits {
		splits[i].BillID = billID
	}
	return r.db.WithContext(ctx).Create(&splits).Error
}

func (r *billRepo) MarkSplitPaid(ctx context.Context, splitID int) error {
	return r.db.WithContext(ctx).Model(&models.BillSplit{}).Where("id = ?", splitID).Update("paid", true).Error
}
