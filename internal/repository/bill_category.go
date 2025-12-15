package repository

import (
	"github.com/Dragodui/diploma-server/internal/models"

	"gorm.io/gorm"
)

type IBillCategoryRepository interface {
	Create(category *models.BillCategory) error
	GetByHomeID(homeID int) ([]models.BillCategory, error)
	Delete(id int) error
	GetByID(id int) (*models.BillCategory, error)
}

type BillCategoryRepository struct {
	db *gorm.DB
}

func NewBillCategoryRepository(db *gorm.DB) *BillCategoryRepository {
	return &BillCategoryRepository{db: db}
}

func (r *BillCategoryRepository) Create(category *models.BillCategory) error {
	return r.db.Create(category).Error
}

func (r *BillCategoryRepository) GetByHomeID(homeID int) ([]models.BillCategory, error) {
	var categories []models.BillCategory
	err := r.db.Where("home_id = ?", homeID).Find(&categories).Error
	return categories, err
}

func (r *BillCategoryRepository) Delete(id int) error {
	return r.db.Delete(&models.BillCategory{}, id).Error
}

func (r *BillCategoryRepository) GetByID(id int) (*models.BillCategory, error) {
	var category models.BillCategory
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}
