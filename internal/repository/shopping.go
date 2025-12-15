package repository

import (
	"errors"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type ShoppingRepository interface {
	// categories
	CreateCategory(c *models.ShoppingCategory) error
	FindAllCategories(homeID int) (*[]models.ShoppingCategory, error)
	FindCategoryByID(id int) (*models.ShoppingCategory, error)
	DeleteCategory(id int) error
	EditCategory(category *models.ShoppingCategory, updates map[string]interface{}) error

	// items
	CreateItem(i *models.ShoppingItem) error
	FindItemByID(id int) (*models.ShoppingItem, error)
	FindItemsByCategoryID(id int) ([]models.ShoppingItem, error)
	DeleteItem(id int) error
	MarkIsBought(id int) error
	EditItem(item *models.ShoppingItem, updates map[string]interface{}) error
}

type shoppingRepo struct {
	db *gorm.DB
}

func NewShoppingRepository(db *gorm.DB) ShoppingRepository {
	return &shoppingRepo{db}
}

// categories
func (r *shoppingRepo) CreateCategory(c *models.ShoppingCategory) error {
	return r.db.Create(c).Error
}

func (r *shoppingRepo) FindAllCategories(homeID int) (*[]models.ShoppingCategory, error) {
	var categories []models.ShoppingCategory

	if err := r.db.Where("home_id=?", homeID).Find(&categories).Error; err != nil {
		return nil, err
	}

	return &categories, nil
}

func (r *shoppingRepo) FindCategoryByID(id int) (*models.ShoppingCategory, error) {
	var category models.ShoppingCategory
	if err := r.db.Preload("Items").First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &category, nil
}

func (r *shoppingRepo) EditCategory(category *models.ShoppingCategory, updates map[string]interface{}) error {
	return r.db.Model(category).Updates(updates).Error
}

func (r *shoppingRepo) DeleteCategory(id int) error {
	// Delete items first
	if err := r.db.Where("category_id = ?", id).Delete(&models.ShoppingItem{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&models.ShoppingCategory{}, id).Error
}

// items
func (r *shoppingRepo) CreateItem(i *models.ShoppingItem) error {
	return r.db.Create(i).Error
}

func (r *shoppingRepo) FindItemsByCategoryID(id int) ([]models.ShoppingItem, error) {
	var items []models.ShoppingItem
	if err := r.db.Where("category_id = ?", id).First(&items).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return items, nil
}

func (r *shoppingRepo) FindItemByID(id int) (*models.ShoppingItem, error) {
	var item models.ShoppingItem
	if err := r.db.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &item, nil
}

func (r *shoppingRepo) DeleteItem(id int) error {
	return r.db.Delete(&models.ShoppingItem{}, id).Error
}

func (r *shoppingRepo) MarkIsBought(id int) error {
	var item models.ShoppingItem

	if err := r.db.First(&item, id).Error; err != nil {
		return err
	}

	item.IsBought = true
	now := time.Now()
	item.BoughtDate = &now

	if err := r.db.Save(&item).Error; err != nil {
		return err
	}

	return nil
}

func (r *shoppingRepo) EditItem(item *models.ShoppingItem, updates map[string]interface{}) error {
	return r.db.Model(item).Updates(updates).Error
}
