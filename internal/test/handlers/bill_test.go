package handlers_test

import (
	"bytes"
	"context"
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
	CreateBillFunc    func(ctx context.Context, billType string, billCategoryID *int, description string, receiptImage *string, totalAmount float64, start, end time.Time, ocrData datatypes.JSON, homeID, userID int, splits []models.SplitInput) error
	GetBillByIDFunc   func(ctx context.Context, billID int) (*models.Bill, error)
	GetBillsByHomeIDFunc func(ctx context.Context, homeID int, categoryID *int) ([]models.Bill, error)
	DeleteFunc        func(ctx context.Context, billID int) error
	MarkBillPayedFunc func(ctx context.Context, billID int) error
	UpdateSplitsFunc  func(ctx context.Context, billID int, splits []models.SplitInput) error
	MarkSplitPaidFunc func(ctx context.Context, splitID int) error
}

func (m *mockBillService) CreateBill(ctx context.Context, billType string, billCategoryID *int, description string, receiptImage *string, totalAmount float64, start, end time.Time, ocrData datatypes.JSON, homeID, userID int, splits []models.SplitInput) error {
	if m.CreateBillFunc != nil {
		return m.CreateBillFunc(ctx, billType, billCategoryID, description, receiptImage, totalAmount, start, end, ocrData, homeID, userID, splits)
	}
	return nil
}

func (m *mockBillService) GetBillByID(ctx context.Context, billID int) (*models.Bill, error) {
	if m.GetBillByIDFunc != nil {
		return m.GetBillByIDFunc(ctx, billID)
	}
	return nil, nil
}

func (m *mockBillService) GetBillsByHomeID(ctx context.Context, homeId int, categoryID *int) ([]models.Bill, error) {
	if m.GetBillsByHomeIDFunc != nil {
		return m.GetBillsByHomeIDFunc(ctx, homeId, categoryID)
	}
	return nil, nil
}

func (m *mockBillService) Delete(ctx context.Context, billID int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, billID)
	}
	return nil
}

func (m *mockBillService) MarkBillPayed(ctx context.Context, billID int) error {
	if m.MarkBillPayedFunc != nil {
		return m.MarkBillPayedFunc(ctx, billID)
	}
	return nil
}

func (m *mockBillService) UpdateSplits(ctx context.Context, billID int, splits []models.SplitInput) error {
	if m.UpdateSplitsFunc != nil {
		return m.UpdateSplitsFunc(ctx, billID, splits)
	}
	return nil
}

func (m *mockBillService) MarkSplitPaid(ctx context.Context, splitID int) error {
	if m.MarkSplitPaidFunc != nil {
		return m.MarkSplitPaidFunc(ctx, splitID)
	}
	return nil
}

func (m *mockBillService) GetSplitByID(ctx context.Context, splitID int) (*models.BillSplit, error) {
	return &models.BillSplit{ID: splitID, UserID: 123, BillID: 1, Amount: 50.0}, nil
}

// Test fixtures
var (
	testStartTime    = time.Now()
	testEndTime      = testStartTime.Add(24 * time.Hour)
	testOCRData, _   = json.Marshal([]byte("{" + "test ocr data" + "}"))
	validBillRequest = models.CreateBillRequest{
		BillType:    "electricity",
		TotalAmount: 100.50,
		Start:       testStartTime,
		End:         testEndTime,
		OCRData:     testOCRData,
	}
)

func setupBillHandler(svc *mockBillService) *handlers.BillHandler {
	return handlers.NewBillHandler(svc, nil)
}

func setupBillRouter(h *handlers.BillHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(utils.WithUserID(r.Context(), 123))
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/bills/{bill_id}", h.GetByID)
	r.Delete("/homes/{home_id}/bills/{bill_id}", h.Delete)
	r.Put("/bills/{bill_id}/mark-payed", h.MarkPayed)
	r.Put("/homes/{home_id}/bills/{bill_id}/splits", h.UpdateSplits)
	r.Patch("/homes/{home_id}/bills/{bill_id}/splits/{split_id}/paid", h.MarkSplitPaid)
	return r
}

func TestBillHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		userID         int
		mockFunc       func(ctx context.Context, billType string, billCategoryID *int, description string, receiptImage *string, totalAmount float64, start, end time.Time, ocrData datatypes.JSON, homeID, userID int, splits []models.SplitInput) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			body:   validBillRequest,
			userID: 123,
			mockFunc: func(ctx context.Context, billType string, billCategoryID *int, description string, receiptImage *string, totalAmount float64, start, end time.Time, ocrData datatypes.JSON, homeID, userID int, splits []models.SplitInput) error {
				assert.Equal(t, "electricity", billType)
				assert.Nil(t, billCategoryID)
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
			mockFunc: func(ctx context.Context, billType string, billCategoryID *int, description string, receiptImage *string, totalAmount float64, start, end time.Time, ocrData datatypes.JSON, homeID, userID int, splits []models.SplitInput) error {
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

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("home_id", "1")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

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
		mockFunc       func(ctx context.Context, billID int) (*models.Bill, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			billID: "1",
			mockFunc: func(ctx context.Context, billID int) (*models.Bill, error) {
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
			mockFunc: func(ctx context.Context, billID int) (*models.Bill, error) {
				return nil, errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to retrieve bill",
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
		mockFunc       func(ctx context.Context, billID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			billID: "1",
			mockFunc: func(ctx context.Context, billID int) error {
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
			mockFunc: func(ctx context.Context, billID int) error {
				return errors.New("delete failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to delete bill",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockBillService{
				DeleteFunc: tt.mockFunc,
				GetBillByIDFunc: func(ctx context.Context, billID int) (*models.Bill, error) {
					return &models.Bill{ID: billID, UploadedBy: 123, HomeID: 1}, nil
				},
			}

			h := setupBillHandler(svc)
			r := setupBillRouter(h)

			req := httptest.NewRequest(http.MethodDelete, "/homes/1/bills/"+tt.billID, nil)
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
		mockFunc       func(ctx context.Context, billID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			billID: "1",
			mockFunc: func(ctx context.Context, billID int) error {
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
			mockFunc: func(ctx context.Context, billID int) error {
				return errors.New("update failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to mark bill as paid",
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

func TestBillHandler_UpdateSplits(t *testing.T) {
	tests := []struct {
		name           string
		billID         string
		body           interface{}
		mockFunc       func(ctx context.Context, billID int, splits []models.SplitInput) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			billID: "1",
			body: models.UpdateSplitsRequest{
				Splits: []models.SplitInput{
					{UserID: 2, Amount: 50.0},
					{UserID: 3, Amount: 50.0},
				},
			},
			mockFunc: func(ctx context.Context, billID int, splits []models.SplitInput) error {
				require.Equal(t, 1, billID)
				require.Len(t, splits, 2)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Splits updated",
		},
		{
			name:           "Invalid Bill ID",
			billID:         "invalid",
			body:           models.UpdateSplitsRequest{},
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid bill ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockBillService{
				UpdateSplitsFunc: tt.mockFunc,
				GetBillByIDFunc: func(ctx context.Context, billID int) (*models.Bill, error) {
					return &models.Bill{ID: billID, UploadedBy: 123, HomeID: 1}, nil
				},
			}

			h := setupBillHandler(svc)
			r := setupBillRouter(h)

			req := makeJSONRequest(http.MethodPut, "/homes/1/bills/"+tt.billID+"/splits", tt.body)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestBillHandler_MarkSplitPaid(t *testing.T) {
	tests := []struct {
		name           string
		splitID        string
		mockFunc       func(ctx context.Context, splitID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Success",
			splitID: "5",
			mockFunc: func(ctx context.Context, splitID int) error {
				require.Equal(t, 5, splitID)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Split marked as paid",
		},
		{
			name:           "Invalid Split ID",
			splitID:        "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid split ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockBillService{
				MarkSplitPaidFunc: tt.mockFunc,
			}

			h := setupBillHandler(svc)
			r := setupBillRouter(h)

			req := httptest.NewRequest(http.MethodPatch, "/homes/1/bills/1/splits/"+tt.splitID+"/paid", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}
