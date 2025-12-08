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
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

// Mock service
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

// Test fixtures
var (
	testStartTime       = time.Now()
	testEndTime         = testStartTime.Add(24 * time.Hour)
	testOCRData, _      = json.Marshal([]byte("{" + "test ocr data" + "}"))
	validBillRequest    = models.CreateBillRequest{
		BillType:    "electricity",
		TotalAmount: 100.50,
		Start:       testStartTime,
		End:         testEndTime,
		OCRData:     testOCRData,
		HomeID:      1,
	}
)

func setupBillHandler(svc *mockBillService) *handlers.BillHandler {
	return handlers.NewBillHandler(svc)
}

func setupBillRouter(h *handlers.BillHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/bills/{bill_id}", h.GetByID)
	r.Delete("/bills/{bill_id}", h.Delete)
	r.Put("/bills/{bill_id}/mark-payed", h.MarkPayed)
	return r
}

func TestBillHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		userID         int
		mockFunc       func(billType string, totalAmount float64, start, end time.Time, ocrData datatypes.JSON, homeID, userID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			body:   validBillRequest,
			userID: 123,
			mockFunc: func(billType string, totalAmount float64, start, end time.Time, ocrData datatypes.JSON, homeID, userID int) error {
				assert.Equal(t, "electricity", billType)
				assert.Equal(t, 100.50, totalAmount)
				assert.Equal(t, 1, homeID)
				assert.Equal(t, 123, userID)
				return nil
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "Created successfully",
		},
		{
			name:           "Invalid JSON",
			body:           "{bad json}",
			userID:         123,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name:           "Unauthorized",
			body:           validBillRequest,
			userID:         0,
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:   "Service Error",
			body:   validBillRequest,
			userID: 123,
			mockFunc: func(billType string, totalAmount float64, start, end time.Time, ocrData datatypes.JSON, homeID, userID int) error {
				return errors.New("service error")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockBillService{
				CreateBillFunc: tt.mockFunc,
			}

			h := setupBillHandler(svc)

			var req *http.Request
			if tt.name == "Invalid JSON" {
				req = httptest.NewRequest(http.MethodPost, "/bills", bytes.NewBufferString("{bad json}"))
			} else {
				req = makeJSONRequest(http.MethodPost, "/bills", tt.body)
			}

			if tt.userID != 0 {
				req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))
			}

			rr := httptest.NewRecorder()
			h.Create(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestBillHandler_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		billID         string
		mockFunc       func(billID int) (*models.Bill, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			billID: "1",
			mockFunc: func(billID int) (*models.Bill, error) {
				require.Equal(t, 1, billID)
				return &models.Bill{ID: 1, Type: "electricity", TotalAmount: 100.50}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "electricity",
		},
		{
			name:           "Invalid ID",
			billID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid bill ID",
		},
		{
			name:   "Service Error",
			billID: "1",
			mockFunc: func(billID int) (*models.Bill, error) {
				return nil, errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockBillService{
				GetBillByIDFunc: tt.mockFunc,
			}

			h := setupBillHandler(svc)
			r := setupBillRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/bills/"+tt.billID, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestBillHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		billID         string
		mockFunc       func(billID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			billID: "1",
			mockFunc: func(billID int) error {
				require.Equal(t, 1, billID)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Deleted successfully",
		},
		{
			name:           "Invalid ID",
			billID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid bill ID",
		},
		{
			name:   "Service Error",
			billID: "1",
			mockFunc: func(billID int) error {
				return errors.New("delete failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "delete failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockBillService{
				DeleteFunc: tt.mockFunc,
			}

			h := setupBillHandler(svc)
			r := setupBillRouter(h)

			req := httptest.NewRequest(http.MethodDelete, "/bills/"+tt.billID, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestBillHandler_MarkPayed(t *testing.T) {
	tests := []struct {
		name           string
		billID         string
		mockFunc       func(billID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			billID: "1",
			mockFunc: func(billID int) error {
				require.Equal(t, 1, billID)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Updated successfully",
		},
		{
			name:           "Invalid ID",
			billID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid bill ID",
		},
		{
			name:   "Service Error",
			billID: "1",
			mockFunc: func(billID int) error {
				return errors.New("update failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "update failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockBillService{
				MarkBillPayedFunc: tt.mockFunc,
			}

			h := setupBillHandler(svc)
			r := setupBillRouter(h)

			req := httptest.NewRequest(http.MethodPut, "/bills/"+tt.billID+"/mark-payed", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}
