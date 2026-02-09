package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Dragodui/diploma-server/internal/http/middleware"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/go-chi/chi/v5"
)

type HomeHandler struct {
	svc services.IHomeService
}

func NewHomeHandler(svc services.IHomeService) *HomeHandler {
	return &HomeHandler{svc}
}

// Create godoc
// @Summary      Create a new home
// @Description  Create a new home with a name
// @Tags         home
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body models.CreateHomeRequest true "Create Home Request"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/create [post]
func (h *HomeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateHomeRequest

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
	if err := h.svc.CreateHome(r.Context(), req.Name, userID); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{"status": true, "message": "Created successfully"})
}

// RegenerateInviteCode godoc
// @Summary      Regenerate invite code
// @Description  Regenerate invite code for a home
// @Tags         home
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/regenerate_code [post]
func (h *HomeHandler) RegenerateInviteCode(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.RegenerateInviteCode(r.Context(), homeID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Invite code regenerated successfully"})
}

// Join godoc
// @Summary      Join a home
// @Description  Join a home with an invite code
// @Tags         home
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body models.JoinRequest true "Join Request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/join [post]
func (h *HomeHandler) Join(w http.ResponseWriter, r *http.Request) {
	var req models.JoinRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// validation
	if err := utils.Validate.Struct(req); err != nil {
		utils.JSONValidationErrors(w, err)
		return
	}

	if strings.TrimSpace(req.Code) == "" {
		utils.JSONError(w, "Invite code is required", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.JoinHomeByCode(r.Context(), req.Code, userID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Joined successfully"})
}

// GetUserHome godoc
// @Summary      Get user's home
// @Description  Get the home the user belongs to
// @Tags         home
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]models.Home
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /homes/my [get]
func (h HomeHandler) GetUserHome(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	home, err := h.svc.GetUserHome(r.Context(), userID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if home == nil {
		utils.JSONError(w, "Home not found", http.StatusNotFound)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]*models.Home{
		"home": home,
	})
}

// GetByID godoc
// @Summary      Get home by ID
// @Description  Get home details by ID
// @Tags         home
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  map[string]models.Home
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id} [get]
func (h *HomeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}
	home, err := h.svc.GetHomeByID(r.Context(), homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if home == nil {
		utils.JSONError(w, "Home not found", http.StatusNotFound)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]*models.Home{
		"home": home,
	})
}

// Delete godoc
// @Summary      Delete home
// @Description  Delete a home by ID
// @Tags         home
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id} [delete]
func (h *HomeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteHome(r.Context(), homeID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Deleted successfully"})
}

// Leave godoc
// @Summary      Leave home
// @Description  Leave the current home
// @Tags         home
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        input body models.LeaveRequest true "Leave Request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/leave [post]
func (h *HomeHandler) Leave(w http.ResponseWriter, r *http.Request) {
	var req models.LeaveRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// validation
	if err := utils.Validate.Struct(req); err != nil {
		utils.JSONValidationErrors(w, err)
		return
	}

	homeID, err := strconv.Atoi(req.HomeID)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.LeaveHome(r.Context(), homeID, userID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Leaved successfully"})
}

// RemoveMember godoc
// @Summary      Remove member
// @Description  Remove a member from the home
// @Tags         home
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        user_id path int true "User ID"
// @Param        input body models.RemoveMemberRequest true "Remove Member Request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/members/{user_id} [delete]
func (h *HomeHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	var req models.RemoveMemberRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// validation
	if err := utils.Validate.Struct(req); err != nil {
		utils.JSONValidationErrors(w, err)
		return
	}

	homeID, err := strconv.Atoi(req.HomeID)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(req.UserID)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	currentUserID := middleware.GetUserID(r)
	if currentUserID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.RemoveMember(r.Context(), homeID, userID, currentUserID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "User removed successfully"})
}
