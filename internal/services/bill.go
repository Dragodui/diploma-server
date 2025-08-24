package services

import (
	"time"

	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/datatypes"
)

type BillService struct {
	bills repository.BillRepository
	cache *redis.Client
}

type IBillService interface {
	CreateBill(billType string, totalAmount int, start, end time.Time,
		ocrData datatypes.JSON, homeID, uploadedBy int) error
	GetBillByID(id int) (*models.Bill, error)
	Delete(id int) error
	MarkBillPayed(id int) error
}

func NewBillService(repo repository.BillRepository, cache *redis.Client) *BillService {
	return &BillService{bills: repo, cache: cache}
}

func (s *BillService) CreateBill(billType string, totalAmount int, start, end time.Time,
	ocrData datatypes.JSON, homeID, uploadedBy int) error {

	bill := &models.Bill{
		HomeID:      homeID,
		UploadedBy:  uploadedBy,
		Type:        billType,
		TotalAmount: totalAmount,
		Start:       start,
		End:         end,
		Payed:       false,
		OCRData:     ocrData,
		CreatedAt:   time.Now(),
	}

	return s.bills.Create(bill)

}

func (s *BillService) GetBillByID(id int) (*models.Bill, error) {
	key := utils.GetBillKey(id)

	// get bill from cache
	cached, err := utils.GetFromCache[models.Bill](key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	return s.bills.FindByID(id)
}

func (s *BillService) Delete(id int) error {
	if err := s.bills.Delete(id); err != nil {
		return err
	}

	key := utils.GetBillKey(id)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return nil
}

func (s *BillService) MarkBillPayed(id int) error {
	// change payed status
	if err := s.bills.MarkPayed(id); err != nil {
		return err
	}

	// remove from cache
	key := utils.GetBillKey(id)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	// get new bill data
	bill, err := s.bills.FindByID(id)
	if err != nil {
		return err
	}

	// write to cache
	if err := utils.WriteToCache(key, bill, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return nil
}
