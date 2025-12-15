package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/internal/utils"

	"github.com/go-chi/chi/v5"
)

type BillCategoryHandler struct {
	svc services.IBillCategoryService
}

func NewBillCategoryHandler(svc services.IBillCategoryService) *BillCategoryHandler {
	return &BillCategoryHandler{svc: svc}
}

// Create godoc
// @Summary Create a new bill category
// @Description Create a new bill category for a home
// @Tags BillCategory
// @Accept json
// @Produce json
// @Param home_id path int true "Home ID"
// @Param request body models.CreateBillCategoryRequest true "Create Bill Category Request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /homes/{home_id}/bill-categories [post]
func (h *BillCategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	homeID, err := strconv.Atoi(chi.URLParam(r, "home_id"))
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	var req models.CreateBillCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.svc.CreateCategory(homeID, req.Name, req.Color); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"status":  true,
		"message": "Category created successfully",
	})
}

// GetAll godoc
// @Summary Get all bill categories
// @Description Get all bill categories for a home
// @Tags BillCategory
// @Produce json
// @Param home_id path int true "Home ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /homes/{home_id}/bill-categories [get]
func (h *BillCategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	homeID, err := strconv.Atoi(chi.URLParam(r, "home_id"))
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	categories, err := h.svc.GetCategories(homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":     true,
		"categories": categories,
	})
}

// Delete godoc
// @Summary Delete a bill category
// @Description Delete a bill category by ID
// @Tags BillCategory
// @Produce json
// @Param home_id path int true "Home ID"
// @Param category_id path int true "Category ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /homes/{home_id}/bill-categories/{category_id} [delete]
func (h *BillCategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	categoryID, err := strconv.Atoi(chi.URLParam(r, "category_id"))
	if err != nil {
		utils.JSONError(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteCategory(categoryID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "Category deleted successfully",
	})
}
