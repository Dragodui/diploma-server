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
	svc services.IBillService
}

func NewBillHandler(svc services.IBillService) *BillHandler {
	return &BillHandler{svc}
}

// GetByHomeID godoc
// @Summary      Get bills by home ID
// @Description  Get all bills in a home
// @Tags         bill
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/bills [get]
func (h *BillHandler) GetByHomeID(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	bills, err := h.svc.GetBillsByHomeID(r.Context(), homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": true,
		"bills":  bills,
	})
}

// Create godoc
// @Summary      Create a new bill
// @Description  Create a new bill in a home
// @Tags         bill
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        input body models.CreateBillRequest true "Create Bill Request"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/bills [post]
func (h *BillHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid home id", http.StatusBadRequest)
	}
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

	if err := h.svc.CreateBill(r.Context(), req.BillType, req.BillCategoryID, req.TotalAmount, req.Start, req.End, req.OCRData, homeID, userID); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{"status": true, "message": "Created successfully"})
}

// GetByID godoc
// @Summary      Get bill by ID
// @Description  Get bill details by ID
// @Tags         bill
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        bill_id path int true "Bill ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/bills/{bill_id} [get]
func (h *BillHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	billIDStr := chi.URLParam(r, "bill_id")
	billID, err := strconv.Atoi(billIDStr)
	if err != nil {
		utils.JSONError(w, "invalid bill ID", http.StatusBadRequest)
		return
	}
	bill, err := h.svc.GetBillByID(r.Context(), billID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": true,
		"bill":   bill,
	})
}

// Delete godoc
// @Summary      Delete bill
// @Description  Delete a bill by ID
// @Tags         bill
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        bill_id path int true "Bill ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/bills/{bill_id} [delete]
func (h *BillHandler) Delete(w http.ResponseWriter, r *http.Request) {
	billIDStr := chi.URLParam(r, "bill_id")
	billID, err := strconv.Atoi(billIDStr)
	if err != nil {
		utils.JSONError(w, "invalid bill ID", http.StatusBadRequest)
		return
	}
	if err := h.svc.Delete(r.Context(), billID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Deleted successfully"})
}

// MarkPayed godoc
// @Summary      Mark bill as payed
// @Description  Mark a bill as payed
// @Tags         bill
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        bill_id path int true "Bill ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/bills/{bill_id} [patch]
func (h *BillHandler) MarkPayed(w http.ResponseWriter, r *http.Request) {
	billIDStr := chi.URLParam(r, "bill_id")
	billID, err := strconv.Atoi(billIDStr)
	if err != nil {
		utils.JSONError(w, "invalid bill ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkBillPayed(r.Context(), billID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Updated successfully"})
}
