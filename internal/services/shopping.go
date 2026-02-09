package services

import (
	"context"
	"errors"
	"time"

	"github.com/Dragodui/diploma-server/internal/event"
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
	CreateCategory(ctx context.Context, name string, icon *string, color string, homeID int) error
	FindAllCategoriesForHome(ctx context.Context, homeID int) (*[]models.ShoppingCategory, error)
	FindCategoryByID(ctx context.Context, categoryID, homeID int) (*models.ShoppingCategory, error)
	DeleteCategory(ctx context.Context, categoryID, homeID int) error
	EditCategory(ctx context.Context, categoryID, homeID int, name, icon, color *string) error

	// items
	CreateItem(ctx context.Context, categoryID, userID int, name string, image, link *string) error
	FindItemByID(ctx context.Context, itemID int) (*models.ShoppingItem, error)
	FindItemsByCategoryID(ctx context.Context, categoryID int) ([]models.ShoppingItem, error)
	DeleteItem(ctx context.Context, itemID int) error
	MarkIsBought(ctx context.Context, itemID int) error
	EditItem(ctx context.Context, itemID int, name, image, link *string, isBought *bool, boughtAt *time.Time) error
}

func NewShoppingService(repo repository.ShoppingRepository, cache *redis.Client) *ShoppingService {
	return &ShoppingService{
		repo,
		cache,
	}
}

// categories
func (s *ShoppingService) CreateCategory(ctx context.Context, name string, icon *string, color string, homeID int) error {
	key := utils.GetAllCategoriesForHomeKey(homeID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}
	category := &models.ShoppingCategory{
		Name:   name,
		Icon:   icon,
		Color:  color,
		HomeID: homeID,
	}
	if err := s.repo.CreateCategory(ctx, category); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleShoppingCategory,
		Action: event.ActionCreated,
		Data:   category,
	})

	return nil
}

func (s *ShoppingService) FindAllCategoriesForHome(ctx context.Context, homeID int) (*[]models.ShoppingCategory, error) {
	key := utils.GetAllCategoriesForHomeKey(homeID)
	cached, err := utils.GetFromCache[[]models.ShoppingCategory](ctx, key, s.cache)

	if cached != nil && err == nil {
		return cached, nil
	}

	categories, err := s.repo.FindAllCategories(ctx, homeID)

	if err != nil {
		return nil, err
	}

	if err := utils.WriteToCache(ctx, key, categories, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return categories, nil
}

func (s *ShoppingService) FindCategoryByID(ctx context.Context, categoryID, homeID int) (*models.ShoppingCategory, error) {
	key := utils.GetCategoryKey(categoryID)
	cached, err := utils.GetFromCache[models.ShoppingCategory](ctx, key, s.cache)

	if cached != nil && err == nil {
		return cached, nil
	}

	category, err := s.repo.FindCategoryByID(ctx, categoryID)

	if category.HomeID != homeID {
		return nil, errCategoryNotBelongsToHome
	}

	if err != nil {
		return nil, err
	}

	if err := utils.WriteToCache(ctx, key, category, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return category, nil
}

func (s *ShoppingService) DeleteCategory(ctx context.Context, categoryID, homeID int) error {
	// Remove from cache
	categoryKey := utils.GetCategoryKey(categoryID)
	categoriesForHomeKey := utils.GetAllCategoriesForHomeKey(homeID)

	if err := utils.DeleteFromCache(ctx, categoryKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", categoryKey, err)
	}

	if err := utils.DeleteFromCache(ctx, categoriesForHomeKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", categoryKey, err)
	}

	if err := s.repo.DeleteCategory(ctx, categoryID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleShoppingCategory,
		Action: event.ActionDeleted,
		Data:   map[string]int{"id": categoryID},
	})

	return nil
}

func (s *ShoppingService) EditCategory(ctx context.Context, categoryID, homeID int, name, icon, color *string) error {
	category, err := s.repo.FindCategoryByID(ctx, categoryID)

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
	if color != nil {
		updates["color"] = *color
	}

	// Remove from cache
	categoryKey := utils.GetCategoryKey(categoryID)
	categoriesForHomeKey := utils.GetAllCategoriesForHomeKey(homeID)

	if err := utils.DeleteFromCache(ctx, categoryKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", categoryKey, err)
	}

	if err := utils.DeleteFromCache(ctx, categoriesForHomeKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", categoryKey, err)
	}

	if err := s.repo.EditCategory(ctx, category, updates); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleShoppingCategory,
		Action: event.ActionUpdated,
		Data:   category,
	})

	return nil
}

// items
func (s *ShoppingService) CreateItem(ctx context.Context, categoryID int, userID int, name string, image, link *string) error {
	// Remove cache
	key := utils.GetCategoryKey(categoryID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	item := &models.ShoppingItem{
		CategoryID: categoryID,
		Name:       name,
		Image:      image,
		Link:       link,
		UploadedBy: userID,
	}
	if err := s.repo.CreateItem(ctx, item); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleShoppingItem,
		Action: event.ActionCreated,
		Data:   item,
	})

	return nil
}

func (s *ShoppingService) FindItemsByCategoryID(ctx context.Context, categoryID int) ([]models.ShoppingItem, error) {
	return s.repo.FindItemsByCategoryID(ctx, categoryID)
}

func (s *ShoppingService) FindItemByID(ctx context.Context, itemID int) (*models.ShoppingItem, error) {
	return s.repo.FindItemByID(ctx, itemID)
}

func (s *ShoppingService) DeleteItem(ctx context.Context, itemID int) error {
	// Remove cache
	item, err := s.repo.FindItemByID(ctx, itemID)
	if err != nil {
		return err
	}
	categoryID := item.CategoryID

	key := utils.GetCategoryKey(categoryID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	if err := s.repo.DeleteItem(ctx, itemID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleShoppingItem,
		Action: event.ActionDeleted,
		Data:   item,
	})

	return nil
}

func (s *ShoppingService) MarkIsBought(ctx context.Context, itemID int) error {
	// Remove cache
	item, err := s.repo.FindItemByID(ctx, itemID)
	if err != nil {
		return err
	}
	categoryID := item.CategoryID

	key := utils.GetCategoryKey(categoryID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	if err := s.repo.MarkIsBought(ctx, itemID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleShoppingItem,
		Action: event.ActionUpdated,
		Data:   item,
	})

	return nil
}

func (s *ShoppingService) EditItem(ctx context.Context, itemID int, name, image, link *string, isBought *bool, boughtAt *time.Time) error {
	item, err := s.repo.FindItemByID(ctx, itemID)
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
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	if err := s.repo.EditItem(ctx, item, updates); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleShoppingItem,
		Action: event.ActionUpdated,
		Data:   item,
	})

	return nil
}

