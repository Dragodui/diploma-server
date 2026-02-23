package handlers

import (
	"net/http"
	"strconv"

	"github.com/Dragodui/diploma-server/internal/http/middleware"
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

// GetByUserID godoc
// @Summary      Get notifications by user ID
// @Description  Get all notifications for the current user
// @Tags         notification
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/notifications [get]
func (h *NotificationHandler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	notifications, err := h.svc.GetByUserID(r.Context(), userID)
	if err != nil {
		utils.SafeError(w, err, "Failed to retrieve notifications", http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":        true,
		"notifications": notifications,
	})
}

// MarkAsRead godoc
// @Summary      Mark notification as read
// @Description  Mark a notification as read
// @Tags         notification
// @Produce      json
// @Security     BearerAuth
// @Param        notification_id path int true "Notification ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/notifications/{notification_id} [delete]
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	notificationIDStr := chi.URLParam(r, "notification_id")
	userID := middleware.GetUserID(r)
	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		utils.JSONError(w, "invalid notification ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkAsRead(r.Context(), notificationID, userID); err != nil {
		utils.SafeError(w, err, "Failed to mark notification as read", http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "Marked as read",
	})
}

// GetByHomeID godoc
// @Summary      Get notifications by home ID
// @Description  Get all notifications for a home
// @Tags         notification
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/notifications [get]
func (h *NotificationHandler) GetByHomeID(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	notifications, err := h.svc.GetByHomeID(r.Context(), homeID)
	if err != nil {
		utils.SafeError(w, err, "Failed to retrieve notifications", http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":        true,
		"notifications": notifications,
	})
}

// MarkAsReadForHome godoc
// @Summary      Mark home notification as read
// @Description  Mark a home notification as read
// @Tags         notification
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        notification_id path int true "Notification ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/notifications/{notification_id} [delete]
func (h *NotificationHandler) MarkAsReadForHome(w http.ResponseWriter, r *http.Request) {
	notificationIDStr := chi.URLParam(r, "notification_id")
	userID := middleware.GetUserID(r)
	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		utils.JSONError(w, "invalid notification ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkAsReadForHomeNotification(r.Context(), notificationID, userID); err != nil {
		utils.SafeError(w, err, "Failed to mark notification as read", http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Marked as read"})
}
