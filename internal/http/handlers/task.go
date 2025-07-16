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

type TaskHandler struct {
	svc *services.TaskService
}

func NewTaskHandler(svc *services.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.svc.CreateTask(req.HomeID, req.Name, req.Description, req.ScheduleType); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]string{"message": "Created successfully"})
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := h.svc.GetTaskByID(taskID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]*models.Task{
		"task": task,
	})
}

func (h *TaskHandler) GetTasksByHomeID(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		http.Error(w, "invalid home ID", http.StatusBadRequest)
		return
	}
	tasks, err := h.svc.GetTasksByHomeID(homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]*[]models.Task{
		"tasks": tasks,
	})
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "invalid task ID", http.StatusBadRequest)
		return
	}

	err = h.svc.DeleteTask(taskID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Deleted successfully"})
}

func (h *TaskHandler) AssignUser(w http.ResponseWriter, r *http.Request) {
	var assignUserRequest models.AssignUserRequest
	if err := json.NewDecoder(r.Body).Decode(&assignUserRequest); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.svc.AssignUser(assignUserRequest.TaskID, assignUserRequest.UserID, assignUserRequest.HomeID, assignUserRequest.Date); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]string{"message": "Created successfully"})
}

func (h *TaskHandler) GetAssignmentsForUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	assignments, err := h.svc.GetAssignmentsForUser(userID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]*[]models.TaskAssignment{
		"assignments": assignments,
	})
}

func (h *TaskHandler) GetClosestAssignmentForUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	assignment, err := h.svc.GetClosestAssignmentForUser(userID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]*models.TaskAssignment{
		"assignment": assignment,
	})
}

func (h *TaskHandler) MarkAssignmentCompleted(w http.ResponseWriter, r *http.Request) {
	var assignmentRequest models.AssignmentIDRequest
	if err := json.NewDecoder(r.Body).Decode(&assignmentRequest); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.svc.MarkAssignmentCompleted(assignmentRequest.AssignmentID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Marked successfully"})
}

func (h *TaskHandler) DeleteAssignment(w http.ResponseWriter, r *http.Request) {
	assignmentIDStr := chi.URLParam(r, "assignment_id")
	assignmentID, err := strconv.Atoi(assignmentIDStr)
	if err != nil {
		http.Error(w, "invalid assignment ID", http.StatusBadRequest)
		return
	}

	err = h.svc.DeleteAssignment(assignmentID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Deleted successfully"})
}
