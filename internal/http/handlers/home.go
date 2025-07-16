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
	svc *services.HomeService
}

func NewHomeHandler(svc *services.HomeService) *HomeHandler {
	return &HomeHandler{svc: svc}
}

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
	if err := h.svc.CreateHome(req.Name, userID); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]string{"message": "Created successfully"})
}

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

	if err := h.svc.JoinHomeByCode(req.Code, userID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Joined successfully"})
}

func (h *HomeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		http.Error(w, "invalid home ID", http.StatusBadRequest)
		return
	}
	home, err := h.svc.GetHomeByID(homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]*models.Home{
		"home": home,
	})
}

func (h *HomeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		http.Error(w, "invalid home ID", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteHome(homeID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"message": "Deleted successfully"})
}

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
		http.Error(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.LeaveHome(homeID, userID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Leaved successfully"})
}

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
		http.Error(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(req.UserID)
	if err != nil {
		http.Error(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	currentUserID := middleware.GetUserID(r)
	if currentUserID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.RemoveMember(homeID, userID, currentUserID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "User removed successfully"})
}
