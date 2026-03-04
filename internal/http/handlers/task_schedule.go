package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/go-chi/chi/v5"
)

type TaskScheduleHandler struct {
	svc      services.ITaskScheduleService
	homeRepo repository.HomeRepository
}

func NewTaskScheduleHandler(svc services.ITaskScheduleService, homeRepo repository.HomeRepository) *TaskScheduleHandler {
	return &TaskScheduleHandler{svc: svc, homeRepo: homeRepo}
}

// CreateSchedule godoc
// @Summary      Create a recurring schedule for a task
// @Description  Admin creates a schedule that rotates task assignments between users on a daily/weekly/monthly basis
// @Tags         task-schedule
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        input body models.CreateTaskScheduleRequest true "Create Schedule Request"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Router       /homes/{home_id}/tasks/schedules [post]
func (h *TaskScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	var req models.CreateTaskScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	req.HomeID = homeID

	schedule, err := h.svc.CreateSchedule(r.Context(), req.TaskID, req.HomeID, req.RecurrenceType, req.UserIDs)
	if err != nil {
		utils.SafeError(w, err, "Failed to create schedule", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{"status": true, "schedule": schedule})
}

// GetScheduleByTaskID godoc
// @Summary      Get schedule for a task
// @Description  Get the recurring schedule configuration for a task
// @Tags         task-schedule
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        task_id path int true "Task ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /homes/{home_id}/tasks/{task_id}/schedule [get]
func (h *TaskScheduleHandler) GetScheduleByTaskID(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		utils.JSONError(w, "invalid task ID", http.StatusBadRequest)
		return
	}

	schedule, err := h.svc.GetScheduleByTaskID(r.Context(), taskID)
	if err != nil {
		utils.SafeError(w, err, "Failed to get schedule", http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "schedule": schedule})
}

// GetSchedulesByHomeID godoc
// @Summary      Get all schedules for a home
// @Description  Get all active recurring task schedules in a home
// @Tags         task-schedule
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /homes/{home_id}/tasks/schedules [get]
func (h *TaskScheduleHandler) GetSchedulesByHomeID(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}

	schedules, err := h.svc.GetSchedulesByHomeID(r.Context(), homeID)
	if err != nil {
		utils.SafeError(w, err, "Failed to get schedules", http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "schedules": schedules})
}

// DeleteSchedule godoc
// @Summary      Delete a schedule
// @Description  Delete a recurring task schedule (admin only)
// @Tags         task-schedule
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        schedule_id path int true "Schedule ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Router       /homes/{home_id}/tasks/schedules/{schedule_id} [delete]
func (h *TaskScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	scheduleIDStr := chi.URLParam(r, "schedule_id")
	scheduleID, err := strconv.Atoi(scheduleIDStr)
	if err != nil {
		utils.JSONError(w, "invalid schedule ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteSchedule(r.Context(), scheduleID); err != nil {
		utils.SafeError(w, err, "Failed to delete schedule", http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Schedule deleted successfully"})
}
