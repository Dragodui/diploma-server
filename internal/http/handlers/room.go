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

// Create godoc
// @Summary      Create a new room
// @Description  Create a new room in a home
// @Tags         room
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        input body models.CreateRoomRequest true "Create Room Request"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/rooms [post]
func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.svc.CreateRoom(r.Context(), req.Name, req.HomeID); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{"status": true, "message": "Created successfully"})
}

// GetByID godoc
// @Summary      Get room by ID
// @Description  Get room details by ID
// @Tags         room
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        room_id path int true "Room ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/rooms/{room_id} [get]
func (h *RoomHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "room_id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		utils.JSONError(w, "invalid room ID", http.StatusBadRequest)
		return
	}

	room, err := h.svc.GetRoomByID(r.Context(), roomID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true,
		"room": room,
	})
}

// GetByHomeID godoc
// @Summary      Get rooms by home ID
// @Description  Get all rooms in a home
// @Tags         room
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/rooms [get]
func (h *RoomHandler) GetByHomeID(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}
	rooms, err := h.svc.GetRoomsByHomeID(r.Context(), homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true,
		"rooms": rooms,
	})
}

// Delete godoc
// @Summary      Delete room
// @Description  Delete a room by ID
// @Tags         room
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        room_id path int true "Room ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/rooms/{room_id} [delete]
func (h *RoomHandler) Delete(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "room_id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		utils.JSONError(w, "invalid room ID", http.StatusBadRequest)
		return
	}

	err = h.svc.DeleteRoom(r.Context(), roomID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Deleted successfully"})
}
