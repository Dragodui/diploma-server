package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

type mockBillService struct {
	CreateBillFunc    func(billType string, totalAmount float64, start, end time.Time, ocrData datatypes.JSON, homeID, userID int) error
	GetBillByIDFunc   func(billID int) (*models.Bill, error)
	DeleteFunc        func(billID int) error
	MarkBillPayedFunc func(billID int) error
}

func (m *mockBillService) CreateBill(billType string, totalAmount float64, start, end time.Time, ocrData datatypes.JSON, homeID, userID int) error {
	if m.CreateBillFunc != nil {
		return m.CreateBillFunc(billType, totalAmount, start, end, ocrData, homeID, userID)
	}
	return nil
}

func (m *mockBillService) GetBillByID(billID int) (*models.Bill, error) {
	if m.GetBillByIDFunc != nil {
		return m.GetBillByIDFunc(billID)
	}
	return nil, nil
}

func (m *mockBillService) Delete(billID int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(billID)
	}
	return nil
}

func (m *mockBillService) MarkBillPayed(billID int) error {
	if m.MarkBillPayedFunc != nil {
		return m.MarkBillPayedFunc(billID)
	}
	return nil
}

// POST /bills/create
func TestBillHandler_Create_Success(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(24 * time.Hour)

	svc := &mockBillService{
		CreateBillFunc: func(billType string, totalAmount float64, start, end time.Time, ocrData datatypes.JSON, homeID, userID int) error {
			assert.Equal(t, "electricity", billType)
			assert.Equal(t, 100.50, totalAmount)
			assert.Equal(t, 1, homeID)
			assert.Equal(t, 123, userID)
			return nil
		},
	}

	h := handlers.NewBillHandler(svc)

	testJson, _ := json.Marshal([]byte("{" + "test ocr data" + "}"))
	reqBody, _ := json.Marshal(models.CreateBillRequest{
		BillType:    "electricity",
		TotalAmount: 100.50,
		Start:       startTime,
		End:         endTime,
		OCRData:     testJson,
		HomeID:      1,
	})

	req := httptest.NewRequest(http.MethodPost, "/bills", bytes.NewReader(reqBody))
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Contains(t, rr.Body.String(), "Created successfully")
}

func TestBillHandler_Create_InvalidJSON(t *testing.T) {
	svc := &mockBillService{}
	h := handlers.NewBillHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/bills", bytes.NewBufferString("{bad json}"))
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid JSON")
}

func TestBillHandler_Create_Unauthorized(t *testing.T) {
	svc := &mockBillService{}
	h := handlers.NewBillHandler(svc)

	reqBody, _ := json.Marshal(models.CreateBillRequest{
		BillType:    "electricity",
		TotalAmount: 100.50,
		HomeID:      1,
	})

	req := httptest.NewRequest(http.MethodPost, "/bills", bytes.NewReader(reqBody))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Unauthorized")
}

func TestBillHandler_Create_ServiceError(t *testing.T) {
	svc := &mockBillService{
		CreateBillFunc: func(billType string, totalAmount float64, start, end time.Time, ocrData datatypes.JSON, homeID, userID int) error {
			return errors.New("service error")
		},
	}

	h := handlers.NewBillHandler(svc)

	reqBody, _ := json.Marshal(models.CreateBillRequest{
		BillType:    "electricity",
		TotalAmount: 100.50,
		Start:       time.Now(),
		End:         time.Now().Add(24 * time.Hour),
		HomeID:      1,
	})

	req := httptest.NewRequest(http.MethodPost, "/bills", bytes.NewReader(reqBody))
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid data")
}

// GET /bills/{bill_id}
func TestBillHandler_GetByID_Success(t *testing.T) {
	svc := &mockBillService{
		GetBillByIDFunc: func(billID int) (*models.Bill, error) {
			assert.Equal(t, 1, billID)
			return &models.Bill{ID: 1, Type: "electricity", TotalAmount: 100.50}, nil
		},
	}

	h := handlers.NewBillHandler(svc)

	r := chi.NewRouter()
	r.Get("/bills/{bill_id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/bills/1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "electricity")
}

func TestBillHandler_GetByID_InvalidID(t *testing.T) {
	svc := &mockBillService{}
	h := handlers.NewBillHandler(svc)

	r := chi.NewRouter()
	r.Get("/bills/{bill_id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/bills/invalid", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid bill ID")
}

func TestBillHandler_GetByID_ServiceError(t *testing.T) {
	svc := &mockBillService{
		GetBillByIDFunc: func(billID int) (*models.Bill, error) {
			return nil, errors.New("service error")
		},
	}

	h := handlers.NewBillHandler(svc)

	r := chi.NewRouter()
	r.Get("/bills/{bill_id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/bills/1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "service error")
}

// DELETE /bills/{bill_id}
func TestBillHandler_Delete_Success(t *testing.T) {
	svc := &mockBillService{
		DeleteFunc: func(billID int) error {
			assert.Equal(t, 1, billID)
			return nil
		},
	}

	h := handlers.NewBillHandler(svc)

	r := chi.NewRouter()
	r.Delete("/bills/{bill_id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/bills/1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Deleted successfully")
}

func TestBillHandler_Delete_InvalidID(t *testing.T) {
	svc := &mockBillService{}
	h := handlers.NewBillHandler(svc)

	r := chi.NewRouter()
	r.Delete("/bills/{bill_id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/bills/invalid", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid bill ID")
}

func TestBillHandler_Delete_ServiceError(t *testing.T) {
	svc := &mockBillService{
		DeleteFunc: func(billID int) error {
			return errors.New("delete failed")
		},
	}

	h := handlers.NewBillHandler(svc)

	r := chi.NewRouter()
	r.Delete("/bills/{bill_id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/bills/1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "delete failed")
}

// PUT /bills/{bill_id}/mark-payed
func TestBillHandler_MarkPayed_Success(t *testing.T) {
	svc := &mockBillService{
		MarkBillPayedFunc: func(billID int) error {
			assert.Equal(t, 1, billID)
			return nil
		},
	}

	h := handlers.NewBillHandler(svc)

	r := chi.NewRouter()
	r.Put("/bills/{bill_id}/mark-payed", h.MarkPayed)

	req := httptest.NewRequest(http.MethodPut, "/bills/1/mark-payed", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Updated successfully")
}

func TestBillHandler_MarkPayed_InvalidID(t *testing.T) {
	svc := &mockBillService{}
	h := handlers.NewBillHandler(svc)

	r := chi.NewRouter()
	r.Put("/bills/{bill_id}/mark-payed", h.MarkPayed)

	req := httptest.NewRequest(http.MethodPut, "/bills/invalid/mark-payed", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid bill ID")
}

func TestBillHandler_MarkPayed_ServiceError(t *testing.T) {
	svc := &mockBillService{
		MarkBillPayedFunc: func(billID int) error {
			return errors.New("update failed")
		},
	}

	h := handlers.NewBillHandler(svc)

	r := chi.NewRouter()
	r.Put("/bills/{bill_id}/mark-payed", h.MarkPayed)

	req := httptest.NewRequest(http.MethodPut, "/bills/1/mark-payed", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "update failed")
}
