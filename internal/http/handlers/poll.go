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

type PollHandler struct {
	svc services.IPollService
}

func NewPollHandler(svc services.IPollService) *PollHandler {
	return &PollHandler{svc}
}

// POST /homes/{home_id}/polls
func (h *PollHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := utils.Validate.Struct(req); err != nil {
		utils.JSONValidationErrors(w, err)
		return
	}

	homeID, err := strconv.Atoi(chi.URLParam(r, "home_id"))
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.Create(homeID, req.Question, req.Type, req.Options); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{"status": true, "message": "Poll created successfully"})
}

// GET /homes/{home_id}/polls
func (h *PollHandler) GetAllByHomeID(w http.ResponseWriter, r *http.Request) {
	homeID, err := strconv.Atoi(chi.URLParam(r, "home_id"))
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	polls, err := h.svc.GetAllPollsByHomeID(homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "polls": polls})
}

// GET /homes/{home_id}/polls/{poll_id}
func (h *PollHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	pollID, err := strconv.Atoi(chi.URLParam(r, "poll_id"))
	if err != nil {
		utils.JSONError(w, "Invalid poll ID", http.StatusBadRequest)
		return
	}

	poll, err := h.svc.GetPollByID(pollID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if poll == nil {
		utils.JSONError(w, "Poll not found", http.StatusNotFound)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "poll": poll})
}

// PATCH /homes/{home_id}/polls/{poll_id}/close
func (h *PollHandler) Close(w http.ResponseWriter, r *http.Request) {
	homeID, err := strconv.Atoi(chi.URLParam(r, "home_id"))
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}
	pollID, err := strconv.Atoi(chi.URLParam(r, "poll_id"))
	if err != nil {
		utils.JSONError(w, "Invalid poll ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.ClosePoll(pollID, homeID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Poll closed successfully"})
}

// DELETE /homes/{home_id}/polls/{poll_id}
func (h *PollHandler) Delete(w http.ResponseWriter, r *http.Request) {
	homeID, err := strconv.Atoi(chi.URLParam(r, "home_id"))
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}
	pollID, err := strconv.Atoi(chi.URLParam(r, "poll_id"))
	if err != nil {
		utils.JSONError(w, "Invalid poll ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.Delete(pollID, homeID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Poll deleted successfully"})
}

// POST /homes/{home_id}/polls/{poll_id}/vote
func (h *PollHandler) Vote(w http.ResponseWriter, r *http.Request) {
	var req models.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := utils.Validate.Struct(req); err != nil {
		utils.JSONValidationErrors(w, err)
		return
	}

	homeID, err := strconv.Atoi(chi.URLParam(r, "home_id"))
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.Vote(userID, req.OptionID, homeID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{"status": true, "message": "Vote submitted successfully"})
}
