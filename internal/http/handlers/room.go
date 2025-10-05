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

type RoomHandler struct {
	svc services.IRoomService
}

func NewRoomHandler(svc services.IRoomService) *RoomHandler {
	return &RoomHandler{svc}
}

func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.svc.CreateRoom(req.Name, req.HomeID); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{"status": true, "message": "Created successfully"})
}

func (h *RoomHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "room_id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		utils.JSONError(w, "invalid room ID", http.StatusBadRequest)
		return
	}

	room, err := h.svc.GetRoomByID(roomID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true,
		"room": room,
	})
}

func (h *RoomHandler) GetByHomeID(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}
	rooms, err := h.svc.GetRoomsByHomeID(homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true,
		"rooms": rooms,
	})
}

func (h *RoomHandler) Delete(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "room_id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		utils.JSONError(w, "invalid room ID", http.StatusBadRequest)
		return
	}

	err = h.svc.DeleteRoom(roomID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Deleted successfully"})
}
