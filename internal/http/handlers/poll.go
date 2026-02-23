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
// Create godoc
// @Summary      Create a new poll
// @Description  Create a new poll in a home
// @Tags         poll
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        input body models.CreatePollRequest true "Create Poll Request"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/polls [post]
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

	if err := h.svc.Create(r.Context(), homeID, req.Question, req.Type, req.Options, req.AllowRevote, req.EndsAt); err != nil {
		utils.SafeError(w, err, "Failed to create poll", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{"status": true, "message": "Poll created successfully"})
}

// GET /homes/{home_id}/polls
// GetAllByHomeID godoc
// @Summary      Get all polls by home ID
// @Description  Get all polls in a home
// @Tags         poll
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/polls [get]
func (h *PollHandler) GetAllByHomeID(w http.ResponseWriter, r *http.Request) {
	homeID, err := strconv.Atoi(chi.URLParam(r, "home_id"))
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	polls, err := h.svc.GetAllPollsByHomeID(r.Context(), homeID)
	if err != nil {
		utils.SafeError(w, err, "Failed to retrieve polls", http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "polls": polls})
}

// GET /homes/{home_id}/polls/{poll_id}
// GetByID godoc
// @Summary      Get poll by ID
// @Description  Get poll details by ID
// @Tags         poll
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        poll_id path int true "Poll ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/polls/{poll_id} [get]
func (h *PollHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	pollID, err := strconv.Atoi(chi.URLParam(r, "poll_id"))
	if err != nil {
		utils.JSONError(w, "Invalid poll ID", http.StatusBadRequest)
		return
	}

	poll, err := h.svc.GetPollByID(r.Context(), pollID)
	if err != nil {
		utils.SafeError(w, err, "Failed to retrieve poll", http.StatusInternalServerError)
		return
	}
	if poll == nil {
		utils.JSONError(w, "Poll not found", http.StatusNotFound)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "poll": poll})
}

// PATCH /homes/{home_id}/polls/{poll_id}/close
// Close godoc
// @Summary      Close poll
// @Description  Close a poll by ID
// @Tags         poll
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        poll_id path int true "Poll ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/polls/{poll_id}/close [patch]
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

	if err := h.svc.ClosePoll(r.Context(), pollID, homeID); err != nil {
		utils.SafeError(w, err, "Failed to close poll", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Poll closed successfully"})
}

// DELETE /homes/{home_id}/polls/{poll_id}
// Delete godoc
// @Summary      Delete poll
// @Description  Delete a poll by ID
// @Tags         poll
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        poll_id path int true "Poll ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/polls/{poll_id} [delete]
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

	if err := h.svc.Delete(r.Context(), pollID, homeID); err != nil {
		utils.SafeError(w, err, "Failed to delete poll", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Poll deleted successfully"})
}

// POST /homes/{home_id}/polls/{poll_id}/vote
// Vote godoc
// @Summary      Vote in poll
// @Description  Vote in a poll
// @Tags         poll
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        poll_id path int true "Poll ID"
// @Param        input body models.VoteRequest true "Vote Request"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/polls/{poll_id}/vote [post]
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

	if err := h.svc.Vote(r.Context(), userID, req.OptionID, homeID); err != nil {
		utils.SafeError(w, err, "Failed to submit vote", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{"status": true, "message": "Vote submitted successfully"})
}

// DELETE /homes/{home_id}/polls/{poll_id}/vote
// Unvote godoc
// @Summary      Remove vote from poll
// @Description  Remove user's vote from a poll
// @Tags         poll
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        poll_id path int true "Poll ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/polls/{poll_id}/vote [delete]
func (h *PollHandler) Unvote(w http.ResponseWriter, r *http.Request) {
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
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.Unvote(r.Context(), userID, pollID, homeID); err != nil {
		if err == services.ErrRevoteNotAllowed {
			utils.JSONError(w, "Revoting is not allowed for this poll", http.StatusForbidden)
			return
		}
		utils.SafeError(w, err, "Failed to remove vote", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Vote removed successfully"})
}
