package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type ShoppingRepository interface {
	// categories
	CreateCategory(ctx context.Context, c *models.ShoppingCategory) error
	FindAllCategories(ctx context.Context, homeID int) (*[]models.ShoppingCategory, error)
	FindCategoryByID(ctx context.Context, id int) (*models.ShoppingCategory, error)
	DeleteCategory(ctx context.Context, id int) error
	EditCategory(ctx context.Context, category *models.ShoppingCategory, updates map[string]interface{}) error

	// items
	CreateItem(ctx context.Context, i *models.ShoppingItem) error
	FindItemByID(ctx context.Context, id int) (*models.ShoppingItem, error)
	FindItemsByCategoryID(ctx context.Context, id int) ([]models.ShoppingItem, error)
	DeleteItem(ctx context.Context, id int) error
	MarkIsBought(ctx context.Context, id int) error
	EditItem(ctx context.Context, item *models.ShoppingItem, updates map[string]interface{}) error
}

type shoppingRepo struct {
	db *gorm.DB
}

func NewShoppingRepository(db *gorm.DB) ShoppingRepository {
	return &shoppingRepo{db}
}

// categories
func (r *shoppingRepo) CreateCategory(ctx context.Context, c *models.ShoppingCategory) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *shoppingRepo) FindAllCategories(ctx context.Context, homeID int) (*[]models.ShoppingCategory, error) {
	var categories []models.ShoppingCategory

	if err := r.db.WithContext(ctx).Where("home_id=?", homeID).Find(&categories).Error; err != nil {
		return nil, err
	}

	return &categories, nil
}

func (r *shoppingRepo) FindCategoryByID(ctx context.Context, id int) (*models.ShoppingCategory, error) {
	var category models.ShoppingCategory
	if err := r.db.WithContext(ctx).Preload("Items").First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &category, nil
}

func (r *shoppingRepo) EditCategory(ctx context.Context, category *models.ShoppingCategory, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(category).Updates(updates).Error
}

func (r *shoppingRepo) DeleteCategory(ctx context.Context, id int) error {
	// Delete items first
	if err := r.db.WithContext(ctx).Where("category_id = ?", id).Delete(&models.ShoppingItem{}).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Delete(&models.ShoppingCategory{}, id).Error
}

// items
func (r *shoppingRepo) CreateItem(ctx context.Context, i *models.ShoppingItem) error {
	return r.db.WithContext(ctx).Create(i).Error
}

func (r *shoppingRepo) FindItemsByCategoryID(ctx context.Context, id int) ([]models.ShoppingItem, error) {
	var items []models.ShoppingItem
	// Use Find() instead of First() to get all items, not just one
	if err := r.db.WithContext(ctx).Where("category_id = ?", id).Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (r *shoppingRepo) FindItemByID(ctx context.Context, id int) (*models.ShoppingItem, error) {
	var item models.ShoppingItem
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &item, nil
}

func (r *shoppingRepo) DeleteItem(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.ShoppingItem{}, id).Error
}

func (r *shoppingRepo) MarkIsBought(ctx context.Context, id int) error {
	var item models.ShoppingItem

	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return err
	}

	item.IsBought = true
	now := time.Now()
	item.BoughtDate = &now

	if err := r.db.WithContext(ctx).Save(&item).Error; err != nil {
		return err
	}

	return nil
}

func (r *shoppingRepo) EditItem(ctx context.Context, item *models.ShoppingItem, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(item).Updates(updates).Error
}

