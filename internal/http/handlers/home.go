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

	if err := h.svc.CreateHome(req.Name); err != nil {
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

	if strings.TrimSpace(req.Code) == "" {
		utils.JSONError(w, "Invite code is required", http.StatusBadRequest)
		return
	}

	userId := middleware.GetUserID(r)
	if userId == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.JoinHomeByCode(req.Code, userId); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Joined successfully"})
}

func (h *HomeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	homeIdStr := chi.URLParam(r, "id")
	homeId, err := strconv.Atoi(homeIdStr)
	if err != nil {
		http.Error(w, "invalid home ID", http.StatusBadRequest)
		return
	}
	home, err := h.svc.GetHomeByID(homeId)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]*models.Home{
		"home": home,
	})
}

func (h *HomeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	homeIdStr := chi.URLParam(r, "id")
	homeId, err := strconv.Atoi(homeIdStr)
	if err != nil {
		http.Error(w, "invalid home ID", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteHome(homeId); err != nil {
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

	homeId, err := strconv.Atoi(req.HomeID)
	if err != nil {
		http.Error(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	userId := middleware.GetUserID(r)
	if userId == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.LeaveHome(homeId, userId); err != nil {
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

	homeId, err := strconv.Atoi(req.HomeID)
	if err != nil {
		http.Error(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	userId, err := strconv.Atoi(req.UserID)
	if err != nil {
		http.Error(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	currentUserId := middleware.GetUserID(r)
	if currentUserId == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.RemoveMember(homeId, userId, currentUserId); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "User removed successfully"})
}
