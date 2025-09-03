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

type ShoppingHandler struct {
	svc services.IShoppingService
}

func NewShoppingHandler(svc services.IShoppingService) *ShoppingHandler {
	return &ShoppingHandler{svc}
}

// categories
func (h *ShoppingHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCategoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.CreateCategory(req.Name, req.Icon, homeID); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]string{"message": "Created successfully"})
}

func (h *ShoppingHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	categories, err := h.svc.FindAllCategoriesForHome(homeID)

	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string][]models.ShoppingCategory{
		"categories": *categories,
	})

}

func (h *ShoppingHandler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := chi.URLParam(r, "category_id")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		utils.JSONError(w, "invalid category ID", http.StatusBadRequest)
		return
	}

	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	category, err := h.svc.FindCategoryByID(categoryID, homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]*models.ShoppingCategory{
		"category": category,
	})
}

func (h *ShoppingHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := chi.URLParam(r, "category_id")
	homeIDStr := chi.URLParam(r, "home_id")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		utils.JSONError(w, "invalid category ID", http.StatusBadRequest)
		return
	}
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteCategory(categoryID, homeID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Deleted successfully",
	})
}

func (h *ShoppingHandler) EditCategory(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateShoppingCategoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	categoryIDStr := chi.URLParam(r, "category_id")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		utils.JSONError(w, "invalid category ID", http.StatusBadRequest)
		return
	}

	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.EditCategory(categoryID, homeID, req.Name, req.Icon); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Deleted successfully",
	})
}

// items
func (h *ShoppingHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req models.CreateShoppingItemRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	// validation
	if err := utils.Validate.Struct(req); err != nil {
		utils.JSONValidationErrors(w, err)
		return
	}
	userID := middleware.GetUserID(r)

	if err := h.svc.CreateItem(req.CategoryID, userID, req.Name, req.Image, req.Link); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Created successfully",
	})
}

func (h *ShoppingHandler) GetItemByID(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		utils.JSONError(w, "invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := h.svc.FindItemByID(itemID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]*models.ShoppingItem{
		"item": item,
	})
}

func (h *ShoppingHandler) GetItemsByCategoryID(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := chi.URLParam(r, "category_id")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		utils.JSONError(w, "invalid category ID", http.StatusBadRequest)
		return
	}

	items, err := h.svc.FindItemsByCategoryID(categoryID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string][]models.ShoppingItem{
		"items": items,
	})
}

func (h *ShoppingHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		utils.JSONError(w, "invalid item ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteItem(itemID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Deleted successfully",
	})
}

func (h *ShoppingHandler) MarkIsBought(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		utils.JSONError(w, "invalid item ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkIsBought(itemID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Marked successfully",
	})
}

func (h *ShoppingHandler) EditItem(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateShoppingItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "invalid data", http.StatusBadRequest)
		return
	}

	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		utils.JSONError(w, "invalid item ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.EditItem(itemID, req.Name, req.Image, req.Link, req.IsBought, req.BoughtAt); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Edited successfully",
	})
}
