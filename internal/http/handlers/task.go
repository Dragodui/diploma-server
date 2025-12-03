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
	svc services.ITaskService
}

func NewTaskHandler(svc services.ITaskService) *TaskHandler {
	return &TaskHandler{svc}
}

// Create godoc
// @Summary      Create a new task
// @Description  Create a new task in a home
// @Tags         task
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        input body models.CreateTaskRequest true "Create Task Request"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/tasks [post]
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.svc.CreateTask(req.HomeID, req.RoomID, req.Name, req.Description, req.ScheduleType); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{"status": true, "message": "Created successfully"})
}

// GetByID godoc
// @Summary      Get task by ID
// @Description  Get task details by ID
// @Tags         task
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        task_id path int true "Task ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/tasks/{task_id} [get]
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		utils.JSONError(w, "invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := h.svc.GetTaskByID(taskID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true,
		"task": task,
	})
}

// GetTasksByHomeID godoc
// @Summary      Get tasks by home ID
// @Description  Get all tasks in a home
// @Tags         task
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/tasks [get]
func (h *TaskHandler) GetTasksByHomeID(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "invalid home ID", http.StatusBadRequest)
		return
	}
	tasks, err := h.svc.GetTasksByHomeID(homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true,
		"tasks": tasks,
	})
}

// DeleteTask godoc
// @Summary      Delete task
// @Description  Delete a task by ID
// @Tags         task
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        task_id path int true "Task ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/tasks/{task_id} [delete]
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		utils.JSONError(w, "invalid task ID", http.StatusBadRequest)
		return
	}

	err = h.svc.DeleteTask(taskID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Deleted successfully"})
}

// AssignUser godoc
// @Summary      Assign user to task
// @Description  Assign a user to a task
// @Tags         task
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        task_id path int true "Task ID"
// @Param        input body models.AssignUserRequest true "Assign User Request"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/tasks/{task_id}/assign [post]
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

	utils.JSON(w, http.StatusCreated, map[string]interface{}{"status": true, "message": "Created successfully"})
}

// GetAssignmentsForUser godoc
// @Summary      Get assignments for user
// @Description  Get all assignments for a user in a home
// @Tags         task
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        user_id path int true "User ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/users/{user_id}/assignments [get]
func (h *TaskHandler) GetAssignmentsForUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		utils.JSONError(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	assignments, err := h.svc.GetAssignmentsForUser(userID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true,
		"assignments": assignments,
	})
}

// GetClosestAssignmentForUser godoc
// @Summary      Get closest assignment for user
// @Description  Get the closest assignment for a user in a home
// @Tags         task
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        user_id path int true "User ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/users/{user_id}/assignments/closest [get]
func (h *TaskHandler) GetClosestAssignmentForUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		utils.JSONError(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	assignment, err := h.svc.GetClosestAssignmentForUser(userID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true,
		"assignment": assignment,
	})
}

// MarkAssignmentCompleted godoc
// @Summary      Mark assignment as completed
// @Description  Mark an assignment as completed
// @Tags         task
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        task_id path int true "Task ID"
// @Param        input body models.AssignmentIDRequest true "Assignment ID Request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/tasks/{task_id}/mark-completed [patch]
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

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Marked successfully"})
}

// DeleteAssignment godoc
// @Summary      Delete assignment
// @Description  Delete an assignment by ID
// @Tags         task
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        task_id path int true "Task ID"
// @Param        assignment_id path int true "Assignment ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /homes/{home_id}/tasks/{task_id}/assignments/{assignment_id} [delete]
func (h *TaskHandler) DeleteAssignment(w http.ResponseWriter, r *http.Request) {
	assignmentIDStr := chi.URLParam(r, "assignment_id")
	assignmentID, err := strconv.Atoi(assignmentIDStr)
	if err != nil {
		utils.JSONError(w, "invalid assignment ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteAssignment(assignmentID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Deleted successfully"})
}

// ReassignRoom godoc
// @Summary      Reassign room for task
// @Description  Reassign a room for a task
// @Tags         task
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        task_id path int true "Task ID"
// @Param        input body models.ReassignRoomRequest true "Reassign Room Request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /homes/{home_id}/tasks/{task_id}/reassign-room [patch]
func (h *TaskHandler) ReassignRoom(w http.ResponseWriter, r *http.Request) {
	var req models.ReassignRoomRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.svc.ReassignRoom(req.TaskID, req.RoomID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Updated successfully"})
}
