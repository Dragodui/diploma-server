package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type mockTaskService struct {
	CreateTaskFunc                  func(homeID int, roomID *int, name, description, scheduleType string) error
	GetTaskByIDFunc                 func(taskID int) (*models.Task, error)
	GetTasksByHomeIDFunc            func(homeID int) (*[]models.Task, error)
	DeleteTaskFunc                  func(taskID int) error
	AssignUserFunc                  func(taskID, userID, homeID int, date time.Time) error
	GetAssignmentsForUserFunc       func(userID int) (*[]models.TaskAssignment, error)
	GetClosestAssignmentForUserFunc func(userID int) (*models.TaskAssignment, error)
	MarkAssignmentCompletedFunc     func(assignmentID int) error
	DeleteAssignmentFunc            func(assignmentID int) error
	ReassignRoomFunc                func(taskID, roomID int) error
}

func (m *mockTaskService) CreateTask(homeID int, roomID *int, name, description, scheduleType string) error {
	if m.CreateTaskFunc != nil {
		return m.CreateTaskFunc(homeID, roomID, name, description, scheduleType)
	}
	return nil
}

func (m *mockTaskService) GetTaskByID(taskID int) (*models.Task, error) {
	if m.GetTaskByIDFunc != nil {
		return m.GetTaskByIDFunc(taskID)
	}
	return nil, nil
}

func (m *mockTaskService) GetTasksByHomeID(homeID int) (*[]models.Task, error) {
	if m.GetTasksByHomeIDFunc != nil {
		return m.GetTasksByHomeIDFunc(homeID)
	}
	return nil, nil
}

func (m *mockTaskService) DeleteTask(taskID int) error {
	if m.DeleteTaskFunc != nil {
		return m.DeleteTaskFunc(taskID)
	}
	return nil
}

func (m *mockTaskService) AssignUser(taskID, userID, homeID int, date time.Time) error {
	if m.AssignUserFunc != nil {
		return m.AssignUserFunc(taskID, userID, homeID, date)
	}
	return nil
}

func (m *mockTaskService) GetAssignmentsForUser(userID int) (*[]models.TaskAssignment, error) {
	if m.GetAssignmentsForUserFunc != nil {
		return m.GetAssignmentsForUserFunc(userID)
	}
	return nil, nil
}

func (m *mockTaskService) GetClosestAssignmentForUser(userID int) (*models.TaskAssignment, error) {
	if m.GetClosestAssignmentForUserFunc != nil {
		return m.GetClosestAssignmentForUserFunc(userID)
	}
	return nil, nil
}

func (m *mockTaskService) MarkAssignmentCompleted(assignmentID int) error {
	if m.MarkAssignmentCompletedFunc != nil {
		return m.MarkAssignmentCompletedFunc(assignmentID)
	}
	return nil
}

func (m *mockTaskService) DeleteAssignment(assignmentID int) error {
	if m.DeleteAssignmentFunc != nil {
		return m.DeleteAssignmentFunc(assignmentID)
	}
	return nil
}

func (m *mockTaskService) ReassignRoom(taskID, roomID int) error {
	if m.ReassignRoomFunc != nil {
		return m.ReassignRoomFunc(taskID, roomID)
	}
	return nil
}

// POST /tasks/create
func TestTaskHandler_Create_Success(t *testing.T) {
	svc := &mockTaskService{
		CreateTaskFunc: func(homeID int, roomID *int, name, description, scheduleType string) error {
			assert.Equal(t, 1, homeID)
			// assert.Equal(t, 2, roomID)
			assert.Equal(t, "Clean Kitchen", name)
			assert.Equal(t, "Daily cleaning", description)
			assert.Equal(t, "daily", scheduleType)
			return nil
		},
	}

	h := handlers.NewTaskHandler(svc)

	testRoomID := 2
	reqBody, _ := json.Marshal(models.CreateTaskRequest{
		HomeID:       1,
		RoomID:       &testRoomID,
		Name:         "Clean Kitchen",
		Description:  "Daily cleaning",
		ScheduleType: "daily",
	})

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(reqBody))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Contains(t, rr.Body.String(), "Created successfully")
}

func TestTaskHandler_Create_InvalidJSON(t *testing.T) {
	svc := &mockTaskService{}
	h := handlers.NewTaskHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString("{bad json}"))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid JSON")
}

func TestTaskHandler_Create_ServiceError(t *testing.T) {
	svc := &mockTaskService{
		CreateTaskFunc: func(homeID int, roomID *int, name, description, scheduleType string) error {
			return errors.New("service error")
		},
	}

	h := handlers.NewTaskHandler(svc)

	testRoomID := 2
	reqBody, _ := json.Marshal(models.CreateTaskRequest{
		HomeID:       1,
		RoomID:       &testRoomID,
		Name:         "Clean Kitchen",
		Description:  "Daily cleaning",
		ScheduleType: "daily",
	})

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(reqBody))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid data")
}

// GET /tasks/{task_id}
func TestTaskHandler_GetByID_Success(t *testing.T) {
	svc := &mockTaskService{
		GetTaskByIDFunc: func(taskID int) (*models.Task, error) {
			assert.Equal(t, 1, taskID)
			return &models.Task{ID: 1, Name: "Clean Kitchen"}, nil
		},
	}

	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Get("/tasks/{task_id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Clean Kitchen")
}

func TestTaskHandler_GetByID_InvalidID(t *testing.T) {
	svc := &mockTaskService{}
	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Get("/tasks/{task_id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/tasks/invalid", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid task ID")
}

func TestTaskHandler_GetByID_ServiceError(t *testing.T) {
	svc := &mockTaskService{
		GetTaskByIDFunc: func(taskID int) (*models.Task, error) {
			return nil, errors.New("service error")
		},
	}

	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Get("/tasks/{task_id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "service error")
}

// GET /homes/{home_id}/tasks
func TestTaskHandler_GetTasksByHomeID_Success(t *testing.T) {
	svc := &mockTaskService{
		GetTasksByHomeIDFunc: func(homeID int) (*[]models.Task, error) {
			assert.Equal(t, 1, homeID)
			tasks := []models.Task{
				{ID: 1, Name: "Clean Kitchen"},
				{ID: 2, Name: "Vacuum Living Room"},
			}
			return &tasks, nil
		},
	}

	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Get("/homes/{home_id}/tasks", h.GetTasksByHomeID)

	req := httptest.NewRequest(http.MethodGet, "/homes/1/tasks", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Clean Kitchen")
	assert.Contains(t, rr.Body.String(), "Vacuum Living Room")
}

func TestTaskHandler_GetTasksByHomeID_InvalidID(t *testing.T) {
	svc := &mockTaskService{}
	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Get("/homes/{home_id}/tasks", h.GetTasksByHomeID)

	req := httptest.NewRequest(http.MethodGet, "/homes/invalid/tasks", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid home ID")
}

func TestTaskHandler_GetTasksByHomeID_ServiceError(t *testing.T) {
	svc := &mockTaskService{
		GetTasksByHomeIDFunc: func(homeID int) (*[]models.Task, error) {
			return nil, errors.New("service error")
		},
	}

	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Get("/homes/{home_id}/tasks", h.GetTasksByHomeID)

	req := httptest.NewRequest(http.MethodGet, "/homes/1/tasks", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "service error")
}

// DELETE /tasks/{task_id}
func TestTaskHandler_DeleteTask_Success(t *testing.T) {
	svc := &mockTaskService{
		DeleteTaskFunc: func(taskID int) error {
			assert.Equal(t, 1, taskID)
			return nil
		},
	}

	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Delete("/tasks/{task_id}", h.DeleteTask)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Deleted successfully")
}

func TestTaskHandler_DeleteTask_InvalidID(t *testing.T) {
	svc := &mockTaskService{}
	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Delete("/tasks/{task_id}", h.DeleteTask)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/invalid", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid task ID")
}

func TestTaskHandler_DeleteTask_ServiceError(t *testing.T) {
	svc := &mockTaskService{
		DeleteTaskFunc: func(taskID int) error {
			return errors.New("delete failed")
		},
	}

	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Delete("/tasks/{task_id}", h.DeleteTask)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "delete failed")
}

// POST /tasks/assign
func TestTaskHandler_AssignUser_Success(t *testing.T) {
	assignDate := time.Now()

	svc := &mockTaskService{
		AssignUserFunc: func(taskID, userID, homeID int, date time.Time) error {
			assert.Equal(t, 1, taskID)
			assert.Equal(t, 2, userID)
			assert.Equal(t, 3, homeID)
			return nil
		},
	}

	h := handlers.NewTaskHandler(svc)

	reqBody, _ := json.Marshal(models.AssignUserRequest{
		TaskID: 1,
		UserID: 2,
		HomeID: 3,
		Date:   assignDate,
	})

	req := httptest.NewRequest(http.MethodPost, "/tasks/assign", bytes.NewReader(reqBody))
	rr := httptest.NewRecorder()

	h.AssignUser(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Contains(t, rr.Body.String(), "Created successfully")
}

func TestTaskHandler_AssignUser_InvalidJSON(t *testing.T) {
	svc := &mockTaskService{}
	h := handlers.NewTaskHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/tasks/assign", bytes.NewBufferString("{bad json}"))
	rr := httptest.NewRecorder()

	h.AssignUser(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid JSON")
}

func TestTaskHandler_AssignUser_ServiceError(t *testing.T) {
	assignDate := time.Now()

	svc := &mockTaskService{
		AssignUserFunc: func(taskID, userID, homeID int, date time.Time) error {
			return errors.New("assign failed")
		},
	}

	h := handlers.NewTaskHandler(svc)

	reqBody, _ := json.Marshal(models.AssignUserRequest{
		TaskID: 1,
		UserID: 2,
		HomeID: 3,
		Date:   assignDate,
	})

	req := httptest.NewRequest(http.MethodPost, "/tasks/assign", bytes.NewReader(reqBody))
	rr := httptest.NewRecorder()

	h.AssignUser(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid data")
}

// GET /users/{user_id}/assignments
func TestTaskHandler_GetAssignmentsForUser_Success(t *testing.T) {
	svc := &mockTaskService{
		GetAssignmentsForUserFunc: func(userID int) (*[]models.TaskAssignment, error) {
			assert.Equal(t, 123, userID)
			assignments := []models.TaskAssignment{
				{ID: 1, TaskID: 1, UserID: 123},
				{ID: 2, TaskID: 2, UserID: 123},
			}
			return &assignments, nil
		},
	}

	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Get("/users/{user_id}/assignments", h.GetAssignmentsForUser)

	req := httptest.NewRequest(http.MethodGet, "/users/123/assignments", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "assignments")
}

func TestTaskHandler_GetAssignmentsForUser_InvalidID(t *testing.T) {
	svc := &mockTaskService{}
	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Get("/users/{user_id}/assignments", h.GetAssignmentsForUser)

	req := httptest.NewRequest(http.MethodGet, "/users/invalid/assignments", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid user ID")
}

func TestTaskHandler_GetAssignmentsForUser_ServiceError(t *testing.T) {
	svc := &mockTaskService{
		GetAssignmentsForUserFunc: func(userID int) (*[]models.TaskAssignment, error) {
			return nil, errors.New("service error")
		},
	}

	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Get("/users/{user_id}/assignments", h.GetAssignmentsForUser)

	req := httptest.NewRequest(http.MethodGet, "/users/123/assignments", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "service error")
}

// GET /users/{user_id}/assignments/closest
func TestTaskHandler_GetClosestAssignmentForUser_Success(t *testing.T) {
	svc := &mockTaskService{
		GetClosestAssignmentForUserFunc: func(userID int) (*models.TaskAssignment, error) {
			assert.Equal(t, 123, userID)
			return &models.TaskAssignment{ID: 1, TaskID: 1, UserID: 123}, nil
		},
	}

	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Get("/users/{user_id}/assignments/closest", h.GetClosestAssignmentForUser)

	req := httptest.NewRequest(http.MethodGet, "/users/123/assignments/closest", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "assignment")
}

func TestTaskHandler_GetClosestAssignmentForUser_InvalidID(t *testing.T) {
	svc := &mockTaskService{}
	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Get("/users/{user_id}/assignments/closest", h.GetClosestAssignmentForUser)

	req := httptest.NewRequest(http.MethodGet, "/users/invalid/assignments/closest", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid user ID")
}

func TestTaskHandler_GetClosestAssignmentForUser_ServiceError(t *testing.T) {
	svc := &mockTaskService{
		GetClosestAssignmentForUserFunc: func(userID int) (*models.TaskAssignment, error) {
			return nil, errors.New("service error")
		},
	}

	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Get("/users/{user_id}/assignments/closest", h.GetClosestAssignmentForUser)

	req := httptest.NewRequest(http.MethodGet, "/users/123/assignments/closest", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "service error")
}

// PUT /assignments/mark-completed
func TestTaskHandler_MarkAssignmentCompleted_Success(t *testing.T) {
	svc := &mockTaskService{
		MarkAssignmentCompletedFunc: func(assignmentID int) error {
			assert.Equal(t, 1, assignmentID)
			return nil
		},
	}

	h := handlers.NewTaskHandler(svc)

	reqBody, _ := json.Marshal(models.AssignmentIDRequest{
		AssignmentID: 1,
	})

	req := httptest.NewRequest(http.MethodPut, "/assignments/mark-completed", bytes.NewReader(reqBody))
	rr := httptest.NewRecorder()

	h.MarkAssignmentCompleted(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Marked successfully")
}

func TestTaskHandler_MarkAssignmentCompleted_InvalidJSON(t *testing.T) {
	svc := &mockTaskService{}
	h := handlers.NewTaskHandler(svc)

	req := httptest.NewRequest(http.MethodPut, "/assignments/mark-completed", bytes.NewBufferString("{bad json}"))
	rr := httptest.NewRecorder()

	h.MarkAssignmentCompleted(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid JSON")
}

func TestTaskHandler_MarkAssignmentCompleted_ServiceError(t *testing.T) {
	svc := &mockTaskService{
		MarkAssignmentCompletedFunc: func(assignmentID int) error {
			return errors.New("mark failed")
		},
	}

	h := handlers.NewTaskHandler(svc)

	reqBody, _ := json.Marshal(models.AssignmentIDRequest{
		AssignmentID: 1,
	})

	req := httptest.NewRequest(http.MethodPut, "/assignments/mark-completed", bytes.NewReader(reqBody))
	rr := httptest.NewRecorder()

	h.MarkAssignmentCompleted(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "mark failed")
}

// DELETE /assignments/{assignment_id}
func TestTaskHandler_DeleteAssignment_Success(t *testing.T) {
	svc := &mockTaskService{
		DeleteAssignmentFunc: func(assignmentID int) error {
			assert.Equal(t, 1, assignmentID)
			return nil
		},
	}

	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Delete("/assignments/{assignment_id}", h.DeleteAssignment)

	req := httptest.NewRequest(http.MethodDelete, "/assignments/1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Deleted successfully")
}

func TestTaskHandler_DeleteAssignment_InvalidID(t *testing.T) {
	svc := &mockTaskService{}
	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Delete("/assignments/{assignment_id}", h.DeleteAssignment)

	req := httptest.NewRequest(http.MethodDelete, "/assignments/invalid", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid assignment ID")
}

func TestTaskHandler_DeleteAssignment_ServiceError(t *testing.T) {
	svc := &mockTaskService{
		DeleteAssignmentFunc: func(assignmentID int) error {
			return errors.New("delete failed")
		},
	}

	h := handlers.NewTaskHandler(svc)

	r := chi.NewRouter()
	r.Delete("/assignments/{assignment_id}", h.DeleteAssignment)

	req := httptest.NewRequest(http.MethodDelete, "/assignments/1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "delete failed")
}

// PUT /tasks/reassign-room
func TestTaskHandler_ReassignRoom_Success(t *testing.T) {
	svc := &mockTaskService{
		ReassignRoomFunc: func(taskID, roomID int) error {
			assert.Equal(t, 1, taskID)
			// assert.Equal(t, 2, roomID)
			return nil
		},
	}

	h := handlers.NewTaskHandler(svc)

	reqBody, _ := json.Marshal(models.ReassignRoomRequest{
		TaskID: 1,
		RoomID: 2,
	})

	req := httptest.NewRequest(http.MethodPut, "/tasks/reassign-room", bytes.NewReader(reqBody))
	rr := httptest.NewRecorder()

	h.ReassignRoom(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Updated successfully")
}

func TestTaskHandler_ReassignRoom_InvalidJSON(t *testing.T) {
	svc := &mockTaskService{}
	h := handlers.NewTaskHandler(svc)

	req := httptest.NewRequest(http.MethodPut, "/tasks/reassign-room", bytes.NewBufferString("{bad json}"))
	rr := httptest.NewRecorder()

	h.ReassignRoom(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid JSON")
}