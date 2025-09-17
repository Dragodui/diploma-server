package handlers

import (
	"net/http"
	"strconv"

	"github.com/Dragodui/diploma-server/internal/http/middleware"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/go-chi/chi/v5"
)

type NotificationHandler struct {
	svc services.INotificationService
}

func NewNotificationHandler(svc services.INotificationService) *NotificationHandler {
	return &NotificationHandler{svc}
}

func (h *NotificationHandler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	notifications, err := h.svc.GetByUserID(userID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string][]models.Notification{
		"notifications": notifications,
	})
}

func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	notificationIDStr := chi.URLParam(r, "notification_id")
	userID := middleware.GetUserID(r)
	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		utils.JSONError(w, "invalid notification ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkAsRead(notificationID, userID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Marked as read",
	})
}

func (h *NotificationHandler) GetByHomeID(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	notifications, err := h.svc.GetByHomeID(homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string][]models.HomeNotification{
		"notifications": notifications,
	})
}

func (h *NotificationHandler) MarkAsReadForHome(w http.ResponseWriter, r *http.Request) {
	notificationIDStr := chi.URLParam(r, "notification_id")
	userID := middleware.GetUserID(r)
	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		utils.JSONError(w, "invalid notification ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkAsReadForHomeNotification(notificationID, userID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Marked as read",
	})
}
