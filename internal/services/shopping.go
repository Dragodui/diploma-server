package services

import (
	"errors"
	"time"

	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/redis/go-redis/v9"
)

var errCategoryNotBelongsToHome error = errors.New("this category does not belongs to this home")

type ShoppingService struct {
	repo  repository.ShoppingRepository
	cache *redis.Client
}

type IShoppingService interface {
	// categories
	CreateCategory(name string, icon *string, homeID int) error
	FindAllCategoriesForHome(homeID int) (*[]models.ShoppingCategory, error)
	FindCategoryByID(categoryID, homeID int) (*models.ShoppingCategory, error)
	DeleteCategory(categoryID, homeID int) error
	EditCategory(categoryID, homeID int, name, icon *string) error

	// items
	CreateItem(categoryID, userID int, name string, image, link *string) error
	FindItemByID(itemID int) (*models.ShoppingItem, error)
	FindItemsByCategoryID(categoryID int) ([]models.ShoppingItem, error)
	DeleteItem(itemID int) error
	MarkIsBought(itemID int) error
	EditItem(itemID int, name, image, link *string, isBought *bool, boughtAt *time.Time) error
}

func NewShoppingService(repo repository.ShoppingRepository, cache *redis.Client) *ShoppingService {
	return &ShoppingService{
		repo,
		cache,
	}
}

// categories
func (s *ShoppingService) CreateCategory(name string, icon *string, homeID int) error {
	key := utils.GetAllCategoriesForHomeKey(homeID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}
	return s.repo.CreateCategory(&models.ShoppingCategory{
		Name:   name,
		Icon:   icon,
		HomeID: homeID,
	})
}

func (s *ShoppingService) FindAllCategoriesForHome(homeID int) (*[]models.ShoppingCategory, error) {
	key := utils.GetAllCategoriesForHomeKey(homeID)
	cached, err := utils.GetFromCache[[]models.ShoppingCategory](key, s.cache)

	if cached != nil && err == nil {
		return cached, nil
	}

	categories, err := s.repo.FindAllCategories(homeID)

	if err != nil {
		return nil, err
	}

	if err := utils.WriteToCache(key, categories, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return categories, nil
}

func (s *ShoppingService) FindCategoryByID(categoryID, homeID int) (*models.ShoppingCategory, error) {
	key := utils.GetCategoryKey(categoryID)
	cached, err := utils.GetFromCache[models.ShoppingCategory](key, s.cache)

	if cached != nil && err == nil {
		return cached, nil
	}

	category, err := s.repo.FindCategoryByID(categoryID)

	if category.HomeID != homeID {
		return nil, errCategoryNotBelongsToHome
	}

	if err != nil {
		return nil, err
	}

	if err := utils.WriteToCache(key, category, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return category, nil
}

func (s *ShoppingService) DeleteCategory(categoryID, homeID int) error {
	// Remove from cache
	categoryKey := utils.GetCategoryKey(categoryID)
	categoriesForHomeKey := utils.GetAllCategoriesForHomeKey(homeID)

	if err := utils.DeleteFromCache(categoryKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", categoryKey, err)
	}

	if err := utils.DeleteFromCache(categoriesForHomeKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", categoryKey, err)
	}

	return s.repo.DeleteCategory(categoryID)
}

func (s *ShoppingService) EditCategory(categoryID, homeID int, name, icon *string) error {
	category, err := s.repo.FindCategoryByID(categoryID)

	if err != nil {
		return err
	}
	updates := map[string]interface{}{}

	if icon != nil {
		updates["icon"] = icon
	}
	if name != nil {
		updates["name"] = *name
	}

	// Remove from cache
	categoryKey := utils.GetCategoryKey(categoryID)
	categoriesForHomeKey := utils.GetAllCategoriesForHomeKey(homeID)

	if err := utils.DeleteFromCache(categoryKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", categoryKey, err)
	}

	if err := utils.DeleteFromCache(categoriesForHomeKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", categoryKey, err)
	}

	return s.repo.EditCategory(category, updates)

}

// items
func (s *ShoppingService) CreateItem(categoryID int, userID int, name string, image, link *string) error {
	// Remove cache
	key := utils.GetCategoryKey(categoryID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return s.repo.CreateItem(&models.ShoppingItem{
		CategoryID: categoryID,
		Name:       name,
		Image:      image,
		Link:       link,
		AddedBy:    userID,
	})
}

func (s *ShoppingService) FindItemsByCategoryID(categoryID int) ([]models.ShoppingItem, error) {
	return s.repo.FindItemsByCategoryID(categoryID)
}

func (s *ShoppingService) FindItemByID(itemID int) (*models.ShoppingItem, error) {
	return s.repo.FindItemByID(itemID)
}

func (s *ShoppingService) DeleteItem(itemID int) error {
	// Remove cache
	item, err := s.repo.FindItemByID(itemID)
	if err != nil {
		return err
	}
	categoryID := item.CategoryID

	key := utils.GetCategoryKey(categoryID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return s.repo.DeleteItem(itemID)
}

func (s *ShoppingService) MarkIsBought(itemID int) error {
	// Remove cache
	item, err := s.repo.FindItemByID(itemID)
	if err != nil {
		return err
	}
	categoryID := item.CategoryID

	key := utils.GetCategoryKey(categoryID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return s.repo.MarkIsBought(itemID)
}

func (s *ShoppingService) EditItem(itemID int, name, image, link *string, isBought *bool, boughtAt *time.Time) error {
	item, err := s.repo.FindItemByID(itemID)
	if err != nil {
		return err
	}
	updates := map[string]interface{}{}

	if name != nil {
		updates["name"] = *name
	}
	if image != nil {
		updates["image"] = image
	}
	if link != nil {
		updates["link"] = link
	}
	if isBought != nil {
		updates["is_bought"] = *isBought
		if *isBought {
			now := time.Now()
			updates["bought_date"] = &now
		} else {
			updates["bought_date"] = nil
		}
	}
	if boughtAt != nil {
		updates["bought_date"] = boughtAt
	}

	// Remove cache
	categoryID := item.CategoryID

	key := utils.GetCategoryKey(categoryID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return s.repo.EditItem(item, updates)
}
