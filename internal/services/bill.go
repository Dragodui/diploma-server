package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Dragodui/diploma-server/internal/event"
	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/metrics"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/datatypes"
)

type BillService struct {
	repo     repository.BillRepository
	cache    *redis.Client
	notifSvc INotificationService
}

type IBillService interface {
	CreateBill(ctx context.Context, billType string, billCategoryID *int, description string, receiptImage *string, totalAmount float64, start, end time.Time,
		ocrData datatypes.JSON, homeID, uploadedBy int, splits []models.SplitInput) error
	GetBillByID(ctx context.Context, id int) (*models.Bill, error)
	GetBillsByHomeID(ctx context.Context, homeID int, categoryID *int) ([]models.Bill, error)
	Delete(ctx context.Context, id int) error
	MarkBillPayed(ctx context.Context, id int) error
	UpdateSplits(ctx context.Context, billID int, splits []models.SplitInput) error
	MarkSplitPaid(ctx context.Context, splitID int) error
	GetSplitByID(ctx context.Context, splitID int) (*models.BillSplit, error)
}

func NewBillService(repo repository.BillRepository, cache *redis.Client, notifSvc INotificationService) *BillService {
	return &BillService{repo: repo, cache: cache, notifSvc: notifSvc}
}

func validateSplits(splits []models.SplitInput, totalAmount float64) error {
	var sum float64
	for _, sp := range splits {
		if sp.Amount <= 0 {
			return fmt.Errorf("split amount must be greater than 0")
		}
		sum += sp.Amount
	}
	if sum > totalAmount {
		return fmt.Errorf("total split amount (%.2f) exceeds bill total (%.2f)", sum, totalAmount)
	}
	return nil
}

func (s *BillService) CreateBill(ctx context.Context, billType string, billCategoryID *int, description string, receiptImage *string, totalAmount float64, start, end time.Time,
	ocrData datatypes.JSON, homeID, uploadedBy int, splits []models.SplitInput) error {

	if len(splits) > 0 {
		if err := validateSplits(splits, totalAmount); err != nil {
			return err
		}
	}

	bill := &models.Bill{
		HomeID:         homeID,
		UploadedBy:     uploadedBy,
		Type:           billType,
		BillCategoryID: billCategoryID,
		Description:    description,
		ReceiptImage:   receiptImage,
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

	// Create splits if provided
	if len(splits) > 0 {
		billSplits := make([]models.BillSplit, len(splits))
		for i, sp := range splits {
			billSplits[i] = models.BillSplit{
				UserID: sp.UserID,
				Amount: sp.Amount,
			}
		}
		if err := s.repo.CreateSplits(ctx, bill.ID, billSplits); err != nil {
			return err
		}
	}

	metrics.BillsTotal.Inc()
	metrics.BillOperationsTotal.WithLabelValues("create").Inc()

	// Notify home about new expense
	fromID := uploadedBy
	desc := fmt.Sprintf("New expense added: $%.2f", totalAmount)
	if description != "" {
		desc = fmt.Sprintf("New expense added: %s ($%.2f)", description, totalAmount)
	}
	_ = s.notifSvc.CreateHomeNotification(ctx, &fromID, homeID, desc)

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

	bill, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if bill == nil {
		return nil, errors.New("bill not found")
	}

	return bill, nil
}

func (s *BillService) GetBillsByHomeID(ctx context.Context, homeID int, categoryID *int) ([]models.Bill, error) {
	return s.repo.FindByHomeID(ctx, homeID, categoryID)
}

func (s *BillService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	metrics.BillsTotal.Dec()
	metrics.BillOperationsTotal.WithLabelValues("delete").Inc()

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

	metrics.BillOperationsTotal.WithLabelValues("mark_paid").Inc()

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
	if bill == nil {
		return errors.New("bill not found")
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

func (s *BillService) UpdateSplits(ctx context.Context, billID int, splits []models.SplitInput) error {
	if len(splits) > 0 {
		bill, err := s.repo.FindByID(ctx, billID)
		if err != nil {
			return err
		}
		if bill == nil {
			return errors.New("bill not found")
		}
		if err := validateSplits(splits, bill.TotalAmount); err != nil {
			return err
		}
	}

	billSplits := make([]models.BillSplit, len(splits))
	for i, sp := range splits {
		billSplits[i] = models.BillSplit{
			UserID: sp.UserID,
			Amount: sp.Amount,
		}
	}

	if err := s.repo.UpdateSplits(ctx, billID, billSplits); err != nil {
		return err
	}

	// Invalidate cache
	key := utils.GetBillKey(billID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleBill,
		Action: event.ActionUpdated,
		Data:   map[string]int{"billID": billID},
	})

	return nil
}

func (s *BillService) GetSplitByID(ctx context.Context, splitID int) (*models.BillSplit, error) {
	return s.repo.FindSplitByID(ctx, splitID)
}

func (s *BillService) MarkSplitPaid(ctx context.Context, splitID int) error {
	if err := s.repo.MarkSplitPaid(ctx, splitID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleBill,
		Action: event.ActionUpdated,
		Data:   map[string]int{"splitID": splitID},
	})

	return nil
}
