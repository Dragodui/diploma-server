package services

import (
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
)

type IBillCategoryService interface {
	CreateCategory(homeID int, name, color string) error
	GetCategories(homeID int) ([]models.BillCategory, error)
	DeleteCategory(id int) error
}

type BillCategoryService struct {
	repo repository.IBillCategoryRepository
}

func NewBillCategoryService(repo repository.IBillCategoryRepository) *BillCategoryService {
	return &BillCategoryService{repo: repo}
}

func (s *BillCategoryService) CreateCategory(homeID int, name, color string) error {
	category := &models.BillCategory{
		HomeID: homeID,
		Name:   name,
		Color:  color,
	}
	return s.repo.Create(category)
}

func (s *BillCategoryService) GetCategories(homeID int) ([]models.BillCategory, error) {
	return s.repo.GetByHomeID(homeID)
}

func (s *BillCategoryService) DeleteCategory(id int) error {
	return s.repo.Delete(id)
}
