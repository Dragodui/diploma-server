package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Dragodui/diploma-server/internal/http/middleware"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/go-chi/chi/v5"
)

type BillHandler struct {
	svc      services.IBillService
	homeRepo repository.HomeRepository
}

func NewBillHandler(svc services.IBillService, homeRepo repository.HomeRepository) *BillHandler {
	return &BillHandler{svc: svc, homeRepo: homeRepo}
}

// GetByHomeID godoc
// @Summary      Get bills by home ID
// @Description  Get all bills in a home
// @Tags         bill
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        category_id query int false "Filter by category ID"
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

	var categoryID *int
	if catStr := r.URL.Query().Get("category_id"); catStr != "" {
		catID, err := strconv.Atoi(catStr)
		if err != nil {
			utils.JSONError(w, "invalid category_id", http.StatusBadRequest)
			return
		}
		categoryID = &catID
	}

	bills, err := h.svc.GetBillsByHomeID(r.Context(), homeID, categoryID)
	if err != nil {
		utils.SafeError(w, err, "Failed to retrieve bills", http.StatusInternalServerError)
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
		return
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

	if err := h.svc.CreateBill(r.Context(), req.BillType, req.BillCategoryID, req.Description, req.TotalAmount, req.Start, req.End, req.OCRData, homeID, userID, req.Splits); err != nil {
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
		utils.SafeError(w, err, "Failed to retrieve bill", http.StatusInternalServerError)
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
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	billIDStr := chi.URLParam(r, "bill_id")
	billID, err := strconv.Atoi(billIDStr)
	if err != nil {
		utils.JSONError(w, "invalid bill ID", http.StatusBadRequest)
		return
	}

	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	// Check ownership or admin
	bill, err := h.svc.GetBillByID(r.Context(), billID)
	if err != nil {
		utils.SafeError(w, err, "Failed to find bill", http.StatusInternalServerError)
		return
	}
	if bill.UploadedBy != userID {
		isAdmin, _ := h.homeRepo.IsAdmin(r.Context(), homeID, userID)
		if !isAdmin {
			utils.JSONError(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	if err := h.svc.Delete(r.Context(), billID); err != nil {
		utils.SafeError(w, err, "Failed to delete bill", http.StatusInternalServerError)
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
		utils.SafeError(w, err, "Failed to mark bill as paid", http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Updated successfully"})
}

// UpdateSplits godoc
// @Summary      Update bill splits
// @Description  Update how a bill is split between users (uploader or admin only)
// @Tags         bill
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        bill_id path int true "Bill ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Router       /homes/{home_id}/bills/{bill_id}/splits [put]
func (h *BillHandler) UpdateSplits(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	billIDStr := chi.URLParam(r, "bill_id")
	billID, err := strconv.Atoi(billIDStr)
	if err != nil {
		utils.JSONError(w, "invalid bill ID", http.StatusBadRequest)
		return
	}

	// Check ownership or admin
	bill, err := h.svc.GetBillByID(r.Context(), billID)
	if err != nil {
		utils.SafeError(w, err, "Failed to find bill", http.StatusInternalServerError)
		return
	}
	if bill.UploadedBy != userID {
		isAdmin, _ := h.homeRepo.IsAdmin(r.Context(), homeID, userID)
		if !isAdmin {
			utils.JSONError(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	var req models.UpdateSplitsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateSplits(r.Context(), billID, req.Splits); err != nil {
		utils.SafeError(w, err, "Failed to update splits", http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Splits updated"})
}

// MarkSplitPaid godoc
// @Summary      Mark a split as paid
// @Description  Mark a single user's bill split as paid
// @Tags         bill
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        bill_id path int true "Bill ID"
// @Param        split_id path int true "Split ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /homes/{home_id}/bills/{bill_id}/splits/{split_id}/paid [patch]
func (h *BillHandler) MarkSplitPaid(w http.ResponseWriter, r *http.Request) {
	splitIDStr := chi.URLParam(r, "split_id")
	splitID, err := strconv.Atoi(splitIDStr)
	if err != nil {
		utils.JSONError(w, "invalid split ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkSplitPaid(r.Context(), splitID); err != nil {
		utils.SafeError(w, err, "Failed to mark split as paid", http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Split marked as paid"})
}
