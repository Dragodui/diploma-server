package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Dragodui/diploma-server/internal/http/middleware"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/go-chi/chi/v5"
)

type BillHandler struct {
	svc *services.BillService
}

func NewBillHandler(svc *services.BillService) *BillHandler {
	return &BillHandler{svc: svc}
}

func (h *BillHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// validation
	if err := utils.Validate.Struct(req); err != nil {
		utils.JSONValidationErrors(w, err)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.CreateBill(req.BillType, req.TotalAmount, req.Start, req.End, req.OCRData, req.HomeID, userID); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]string{"message": "Created successfully"})
}

func (h *BillHandler) GetById(w http.ResponseWriter, r *http.Request) {
	billIDStr := chi.URLParam(r, "bill_id")
	billID, err := strconv.Atoi(billIDStr)
	if err != nil {
		http.Error(w, "invalid home ID", http.StatusBadRequest)
		return
	}
	bill, err := h.svc.GetBillByID(billID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]*models.Bill{
		"bill": bill,
	})
}

func (h *BillHandler) Delete(w http.ResponseWriter, r *http.Request) {
	billIDStr := chi.URLParam(r, "bill_id")
	billID, err := strconv.Atoi(billIDStr)
	if err != nil {
		http.Error(w, "invalid home ID", http.StatusBadRequest)
		return
	}
	if err := h.svc.Delete(billID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"message": "Deleted successfully"})
}

func (h *BillHandler) MarkPayed(w http.ResponseWriter, r *http.Request) {
	billIDStr := chi.URLParam(r, "bill_id")
	billID, err := strconv.Atoi(billIDStr)
	if err != nil {
		http.Error(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkBillPayed(billID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Updated successfully"})
}
