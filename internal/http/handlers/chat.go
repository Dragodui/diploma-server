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

type ChatHandler struct {
	svc      services.IChatService
	homeRepo repository.HomeRepository
}

func NewChatHandler(svc services.IChatService, homeRepo repository.HomeRepository) *ChatHandler {
	return &ChatHandler{svc: svc, homeRepo: homeRepo}
}

// SendMessage godoc
// @Summary      Send a chat message
// @Description  Post a message to the home chat, optionally mentioning users, tasks, bills, shopping items and categories
// @Tags         chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        input body models.CreateChatMessageRequest true "Create Chat Message Request"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/chat [post]
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	homeID, err := strconv.Atoi(chi.URLParam(r, "home_id"))
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	var req models.CreateChatMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	message, err := h.svc.SendMessage(r.Context(), homeID, userID, req)
	if err != nil {
		utils.SafeError(w, err, "Failed to send message", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"status": true,
		"data":   message,
	})
}

// GetMessages godoc
// @Summary      Get chat messages
// @Description  Get a page of home chat messages, newest first; pass before_id to page into history
// @Tags         chat
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        limit query int false "Page size (default 50, max 200)"
// @Param        before_id query int false "Return messages older than this message ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/chat [get]
func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	homeID, err := strconv.Atoi(chi.URLParam(r, "home_id"))
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	limit := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, _ = strconv.Atoi(raw)
	}

	var beforeID *int
	if raw := r.URL.Query().Get("before_id"); raw != "" {
		if parsed, convErr := strconv.Atoi(raw); convErr == nil {
			beforeID = &parsed
		}
	}

	messages, err := h.svc.GetMessages(r.Context(), homeID, limit, beforeID)
	if err != nil {
		utils.SafeError(w, err, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":   true,
		"messages": messages,
	})
}

// UpdateMessage godoc
// @Summary      Edit a chat message
// @Description  Edit your own chat message
// @Tags         chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        message_id path int true "Message ID"
// @Param        input body models.UpdateChatMessageRequest true "Update Chat Message Request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/chat/{message_id} [put]
func (h *ChatHandler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	homeID, err := strconv.Atoi(chi.URLParam(r, "home_id"))
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	messageID, err := strconv.Atoi(chi.URLParam(r, "message_id"))
	if err != nil {
		utils.JSONError(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateChatMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	message, err := h.svc.UpdateMessage(r.Context(), messageID, homeID, userID, req)
	if err != nil {
		utils.SafeError(w, err, "Failed to update message", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   message,
	})
}

// DeleteMessage godoc
// @Summary      Delete a chat message
// @Description  Delete your own chat message
// @Tags         chat
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        message_id path int true "Message ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/chat/{message_id} [delete]
func (h *ChatHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	homeID, err := strconv.Atoi(chi.URLParam(r, "home_id"))
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	messageID, err := strconv.Atoi(chi.URLParam(r, "message_id"))
	if err != nil {
		utils.JSONError(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteMessage(r.Context(), messageID, homeID, userID); err != nil {
		utils.SafeError(w, err, "Failed to delete message", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true})
}

// MarkRead godoc
// @Summary      Mark chat messages as read
// @Description  Mark every message up to and including last_message_id as read by the caller
// @Tags         chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        input body models.MarkChatReadRequest true "Mark Chat Read Request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/chat/read [post]
func (h *ChatHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	homeID, err := strconv.Atoi(chi.URLParam(r, "home_id"))
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	var req models.MarkChatReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkRead(r.Context(), homeID, userID, req.LastMessageID); err != nil {
		utils.SafeError(w, err, "Failed to mark messages as read", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true})
}

// GetUnreadCount godoc
// @Summary      Get unread chat message count
// @Description  Number of chat messages in the home the caller hasn't read yet
// @Tags         chat
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/chat/unread [get]
func (h *ChatHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	homeID, err := strconv.Atoi(chi.URLParam(r, "home_id"))
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	count, err := h.svc.GetUnreadCount(r.Context(), homeID, userID)
	if err != nil {
		utils.SafeError(w, err, "Failed to get unread count", http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": true,
		"count":  count,
	})
}
