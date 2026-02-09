package repository

import (
	"context"

	"github.com/Dragodui/diploma-server/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IBillCategoryRepository interface {
	Create(ctx context.Context, category *models.BillCategory) error
	GetByHomeID(ctx context.Context, homeID int) ([]models.BillCategory, error)
	Update(ctx context.Context, category *models.BillCategory, updates map[string]interface{}) (*models.BillCategory, error)
	Delete(ctx context.Context, id int) error
	GetByID(ctx context.Context, id int) (*models.BillCategory, error)
}

type BillCategoryRepository struct {
	db *gorm.DB
}

func NewBillCategoryRepository(db *gorm.DB) *BillCategoryRepository {
	return &BillCategoryRepository{db: db}
}

func (r *BillCategoryRepository) Create(ctx context.Context, category *models.BillCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *BillCategoryRepository) GetByHomeID(ctx context.Context, homeID int) ([]models.BillCategory, error) {
	var categories []models.BillCategory
	err := r.db.WithContext(ctx).Where("home_id = ?", homeID).Find(&categories).Error
	return categories, err
}

func (r *BillCategoryRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.BillCategory{}, id).Error
}

func (r *BillCategoryRepository) GetByID(ctx context.Context, id int) (*models.BillCategory, error) {
	var category models.BillCategory
	err := r.db.WithContext(ctx).First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *BillCategoryRepository) Update(ctx context.Context, category *models.BillCategory, updates map[string]interface{}) (*models.BillCategory, error) {
	err := r.db.WithContext(ctx).Model(category).Clauses(clause.Returning{}).Updates(updates).Error
	if err != nil {
		return nil, err
	}
	return category, nil
}

