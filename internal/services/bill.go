package services

import (
	"context"
	"time"

	"github.com/Dragodui/diploma-server/internal/event"
	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/datatypes"
)

type BillService struct {
	repo  repository.BillRepository
	cache *redis.Client
}

type IBillService interface {
	CreateBill(ctx context.Context, billType string, billCategoryID *int, totalAmount float64, start, end time.Time,
		ocrData datatypes.JSON, homeID, uploadedBy int) error
	GetBillByID(ctx context.Context, id int) (*models.Bill, error)
	GetBillsByHomeID(ctx context.Context, homeID int) ([]models.Bill, error)
	Delete(ctx context.Context, id int) error
	MarkBillPayed(ctx context.Context, id int) error
}

func NewBillService(repo repository.BillRepository, cache *redis.Client) *BillService {
	return &BillService{repo: repo, cache: cache}
}

func (s *BillService) CreateBill(ctx context.Context, billType string, billCategoryID *int, totalAmount float64, start, end time.Time,
	ocrData datatypes.JSON, homeID, uploadedBy int) error {

	bill := &models.Bill{
		HomeID:         homeID,
		UploadedBy:     uploadedBy,
		Type:           billType,
		BillCategoryID: billCategoryID,
		TotalAmount:    totalAmount,
		Start:          start,
		End:            end,
		Payed:          false,
		OCRData:        ocrData,
		CreatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, bill); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleBill,
		Action: event.ActionCreated,
		Data:   bill,
	})

	return nil
}

func (s *BillService) GetBillByID(ctx context.Context, id int) (*models.Bill, error) {
	key := utils.GetBillKey(id)

	// get bill from cache
	cached, err := utils.GetFromCache[models.Bill](ctx, key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	return s.repo.FindByID(ctx, id)
}

func (s *BillService) GetBillsByHomeID(ctx context.Context, homeID int) ([]models.Bill, error) {
	return s.repo.FindByHomeID(ctx, homeID)
}

func (s *BillService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	key := utils.GetBillKey(id)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleBill,
		Action: event.ActionDeleted,
		Data:   map[string]int{"id": id},
	})

	return nil
}

func (s *BillService) MarkBillPayed(ctx context.Context, id int) error {
	// change payed status
	if err := s.repo.MarkPayed(ctx, id); err != nil {
		return err
	}

	// remove from cache
	key := utils.GetBillKey(id)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	// get new bill data
	bill, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// write to cache
	if err := utils.WriteToCache(ctx, key, bill, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleBill,
		Action: event.ActionMarkedPayed,
		Data:   bill,
	})

	return nil
}

