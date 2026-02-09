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
// CreateCategory godoc
// @Summary      Create a new shopping category
// @Description  Create a new shopping category in a home
// @Tags         shopping
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        input body models.CreateCategoryRequest true "Create Category Request"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/shopping/categories [post]
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

	if err := h.svc.CreateCategory(r.Context(), req.Name, req.Icon, req.Color, homeID); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{"status": true, "message": "Created successfully"})
}

// GetAllCategories godoc
// @Summary      Get all shopping categories
// @Description  Get all shopping categories in a home
// @Tags         shopping
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/shopping/categories/all [get]
func (h *ShoppingHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	categories, err := h.svc.FindAllCategoriesForHome(r.Context(), homeID)

	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true,
		"categories": *categories,
	})

}

// GetCategoryByID godoc
// @Summary      Get shopping category by ID
// @Description  Get shopping category details by ID
// @Tags         shopping
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        category_id path int true "Category ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/shopping/categories/{category_id} [get]
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

	category, err := h.svc.FindCategoryByID(r.Context(), categoryID, homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true,
		"category": category,
	})
}

// DeleteCategory godoc
// @Summary      Delete shopping category
// @Description  Delete a shopping category by ID
// @Tags         shopping
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        category_id path int true "Category ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/shopping/categories/{category_id} [delete]
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

	if err := h.svc.DeleteCategory(r.Context(), categoryID, homeID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Deleted successfully"})
}

// EditCategory godoc
// @Summary      Edit shopping category
// @Description  Edit a shopping category by ID
// @Tags         shopping
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        category_id path int true "Category ID"
// @Param        input body models.UpdateShoppingCategoryRequest true "Update Category Request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/shopping/categories/{category_id} [put]
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

	if err := h.svc.EditCategory(r.Context(), categoryID, homeID, req.Name, req.Icon, req.Color); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Updated successfully"})
}

// items
// CreateItem godoc
// @Summary      Create a new shopping item
// @Description  Create a new shopping item in a category
// @Tags         shopping
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        input body models.CreateShoppingItemRequest true "Create Item Request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/shopping/items [post]
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

	if err := h.svc.CreateItem(r.Context(), req.CategoryID, userID, req.Name, req.Image, req.Link); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Created successfully"})
}

// GetItemByID godoc
// @Summary      Get shopping item by ID
// @Description  Get shopping item details by ID
// @Tags         shopping
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        item_id path int true "Item ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/shopping/items/{item_id} [get]
func (h *ShoppingHandler) GetItemByID(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		utils.JSONError(w, "invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := h.svc.FindItemByID(r.Context(), itemID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true,
		"item": item,
	})
}

// GetItemsByCategoryID godoc
// @Summary      Get shopping items by category ID
// @Description  Get all shopping items in a category
// @Tags         shopping
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        category_id path int true "Category ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/shopping/categories/{category_id}/items [get]
func (h *ShoppingHandler) GetItemsByCategoryID(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := chi.URLParam(r, "category_id")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		utils.JSONError(w, "invalid category ID", http.StatusBadRequest)
		return
	}

	items, err := h.svc.FindItemsByCategoryID(r.Context(), categoryID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true,
		"items": items,
	})
}

// DeleteItem godoc
// @Summary      Delete shopping item
// @Description  Delete a shopping item by ID
// @Tags         shopping
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        item_id path int true "Item ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/shopping/items/{item_id} [delete]
func (h *ShoppingHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		utils.JSONError(w, "invalid item ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteItem(r.Context(), itemID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Deleted successfully"})
}

// MarkIsBought godoc
// @Summary      Mark shopping item as bought
// @Description  Mark a shopping item as bought
// @Tags         shopping
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        item_id path int true "Item ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/shopping/items/{item_id} [patch]
func (h *ShoppingHandler) MarkIsBought(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		utils.JSONError(w, "invalid item ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkIsBought(r.Context(), itemID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Marked successfully"})
}

// EditItem godoc
// @Summary      Edit shopping item
// @Description  Edit a shopping item by ID
// @Tags         shopping
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        item_id path int true "Item ID"
// @Param        input body models.UpdateShoppingItemRequest true "Update Item Request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/shopping/items/{item_id} [put]
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

	if err := h.svc.EditItem(r.Context(), itemID, req.Name, req.Image, req.Link, req.IsBought, req.BoughtAt); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Edited successfully"})
}
