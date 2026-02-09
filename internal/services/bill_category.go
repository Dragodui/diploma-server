package services

import (
	"context"

	"github.com/Dragodui/diploma-server/internal/event"
	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/redis/go-redis/v9"
)

type IBillCategoryService interface {
	CreateCategory(ctx context.Context, homeID int, name, color string) error
	GetCategories(ctx context.Context, homeID int) ([]models.BillCategory, error)
	UpdateCategory(ctx context.Context, categoryID int, name, color *string) (*models.BillCategory, error)
	DeleteCategory(ctx context.Context, id int, homeID int) error 
}

type BillCategoryService struct {
	cache *redis.Client
	repo  repository.IBillCategoryRepository
}

func NewBillCategoryService(repo repository.IBillCategoryRepository, cache *redis.Client) *BillCategoryService {
	return &BillCategoryService{repo: repo, cache: cache}
}

func (s *BillCategoryService) CreateCategory(ctx context.Context, homeID int, name, color string) error {
	category := &models.BillCategory{
		HomeID: homeID,
		Name:   name,
		Color:  color,
	}

	key := utils.GetBillCategoriesKey(homeID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	if err := s.repo.Create(ctx, category); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleBillCategory,
		Action: event.ActionCreated,
		Data:   category,
	})

	return nil
}

func (s *BillCategoryService) GetCategories(ctx context.Context, homeID int) ([]models.BillCategory, error) {
	key := utils.GetBillCategoriesKey(homeID)
	
	cached, err := utils.GetFromCache[[]models.BillCategory](ctx, key, s.cache)
	if cached != nil && err == nil {
		return *cached, nil
	}

	categories, err := s.repo.GetByHomeID(ctx, homeID)
	if err != nil {
		return nil, err
	}

	if len(categories) > 0 {
		_ = utils.WriteToCache(ctx, key, categories, s.cache)
	}

	return categories, nil
}

func (s *BillCategoryService) UpdateCategory(ctx context.Context, categoryID int, name, color *string) (*models.BillCategory, error) {
	category, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	key := utils.GetBillCategoriesKey(category.HomeID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	updates := map[string]interface{}{}
	if name != nil {
		updates["name"] = *name
	}
	if color != nil {
		updates["color"] = *color
	}

	newCategory, err := s.repo.Update(ctx, category, updates)
	if err != nil {
		return nil, err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleBillCategory,
		Action: event.ActionUpdated,
		Data:   newCategory,
	})

	return newCategory, nil
}

func (s *BillCategoryService) DeleteCategory(ctx context.Context, id int, homeID int) error {
	key := utils.GetBillCategoriesKey(homeID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleBillCategory,
		Action: event.ActionDeleted,
		Data:   map[string]int{"id": id},
	})

	return nil
}