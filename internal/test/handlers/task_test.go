package handlers_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock service
type mockTaskService struct {
	CreateTaskFunc                  func(homeID int, roomID *int, name, description, scheduleType string, dueDate *time.Time) error
	CreateTaskWithAssignmentFunc    func(homeID int, roomID *int, name, description, scheduleType string, dueDate *time.Time, userID int) error
	GetTaskByIDFunc                 func(taskID int) (*models.Task, error)
	GetTasksByHomeIDFunc            func(homeID int) (*[]models.Task, error)
	DeleteTaskFunc                  func(taskID int) error
	AssignUserFunc                  func(taskID, userID, homeID int, date time.Time) error
	GetAssignmentsForUserFunc       func(userID int) (*[]models.TaskAssignment, error)
	GetClosestAssignmentForUserFunc func(userID int) (*models.TaskAssignment, error)
	MarkAssignmentCompletedFunc     func(assignmentID int) error
	MarkAssignmentUncompletedFunc   func(assignmentID int) error
	MarkTaskCompletedForUserFunc    func(taskID, userID, homeID int) error
	DeleteAssignmentFunc            func(assignmentID int) error
	ReassignRoomFunc                func(taskID, roomID int) error
}

func (m *mockTaskService) CreateTaskWithAssignment(homeID int, roomID *int, name, description, scheduleType string, dueDate *time.Time, userID int) error {
	if m.CreateTaskFunc != nil {
		return m.CreateTaskWithAssignmentFunc(homeID, roomID, name, description, scheduleType, dueDate, userID)
	}
	return nil
}

func (m *mockTaskService) CreateTask(homeID int, roomID *int, name, description, scheduleType string, dueDate *time.Time) error {
	if m.CreateTaskFunc != nil {
		return m.CreateTaskFunc(homeID, roomID, name, description, scheduleType, dueDate)
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

func (m *mockTaskService) MarkAssignmentUncompleted(assignmentID int) error {
	if m.MarkAssignmentUncompletedFunc != nil {
		return m.MarkAssignmentUncompletedFunc(assignmentID)
	}
	return nil
}

func (m *mockTaskService) MarkTaskCompletedForUser(taskID, userID, homeID int) error {
	if m.MarkTaskCompletedForUserFunc != nil {
		return m.MarkTaskCompletedForUserFunc(taskID, userID, homeID)
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

// Test fixtures
var (
	testRoomID         = 2
	validCreateTaskReq = models.CreateTaskRequest{
		HomeID:       1,
		RoomID:       &testRoomID,
		Name:         "Clean Kitchen",
		Description:  "Daily cleaning",
		ScheduleType: "daily",
	}
	validAssignUserReq = models.AssignUserRequest{
		TaskID: 1,
		UserID: 2,
		HomeID: 3,
		Date:   time.Now(),
	}
	validReassignRoomReq = models.ReassignRoomRequest{
		TaskID: 1,
		RoomID: 2,
	}
)

func setupTaskHandler(svc *mockTaskService) *handlers.TaskHandler {
	return handlers.NewTaskHandler(svc)
}

func setupTaskRouter(h *handlers.TaskHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/tasks/{task_id}", h.GetByID)
	r.Get("/homes/{home_id}/tasks", h.GetTasksByHomeID)
	r.Delete("/tasks/{task_id}", h.DeleteTask)
	r.Get("/users/{user_id}/assignments", h.GetAssignmentsForUser)
	r.Get("/users/{user_id}/assignments/closest", h.GetClosestAssignmentForUser)
	r.Delete("/assignments/{assignment_id}", h.DeleteAssignment)
	return r
}

func TestTaskHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		mockFunc       func(homeID int, roomID *int, name, description, scheduleType string, dueDate *time.Time) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			body: validCreateTaskReq,
			mockFunc: func(homeID int, roomID *int, name, description, scheduleType string, dueDate *time.Time) error {
				assert.Equal(t, 1, homeID)
				assert.Equal(t, "Clean Kitchen", name)
				assert.Equal(t, "Daily cleaning", description)
				assert.Equal(t, "daily", scheduleType)
				assert.Nil(t, dueDate)
				return nil
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "Created successfully",
		},
		{
			name:           "Invalid JSON",
			body:           "{bad json}",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name: "Service Error",
			body: validCreateTaskReq,
			mockFunc: func(homeID int, roomID *int, name, description, scheduleType string, dueDate *time.Time) error {
				return errors.New("service error")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTaskService{
				CreateTaskFunc: tt.mockFunc,
			}

			h := setupTaskHandler(svc)

			var req *http.Request
			if tt.name == "Invalid JSON" {
				req = httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString("{bad json}"))
			} else {
				req = makeJSONRequest(http.MethodPost, "/tasks", tt.body)
			}

			rr := httptest.NewRecorder()
			h.Create(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestTaskHandler_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		mockFunc       func(taskID int) (*models.Task, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			taskID: "1",
			mockFunc: func(taskID int) (*models.Task, error) {
				require.Equal(t, 1, taskID)
				return &models.Task{ID: 1, Name: "Clean Kitchen"}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Clean Kitchen",
		},
		{
			name:           "Invalid ID",
			taskID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid task ID",
		},
		{
			name:   "Service Error",
			taskID: "1",
			mockFunc: func(taskID int) (*models.Task, error) {
				return nil, errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTaskService{
				GetTaskByIDFunc: tt.mockFunc,
			}

			h := setupTaskHandler(svc)
			r := setupTaskRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/tasks/"+tt.taskID, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestTaskHandler_GetTasksByHomeID(t *testing.T) {
	tests := []struct {
		name           string
		homeID         string
		mockFunc       func(homeID int) (*[]models.Task, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			homeID: "1",
			mockFunc: func(homeID int) (*[]models.Task, error) {
				require.Equal(t, 1, homeID)
				tasks := []models.Task{
					{ID: 1, Name: "Clean Kitchen"},
					{ID: 2, Name: "Vacuum Living Room"},
				}
				return &tasks, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Clean Kitchen",
		},
		{
			name:           "Invalid ID",
			homeID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid home ID",
		},
		{
			name:   "Service Error",
			homeID: "1",
			mockFunc: func(homeID int) (*[]models.Task, error) {
				return nil, errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTaskService{
				GetTasksByHomeIDFunc: tt.mockFunc,
			}

			h := setupTaskHandler(svc)
			r := setupTaskRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/homes/"+tt.homeID+"/tasks", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		mockFunc       func(taskID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			taskID: "1",
			mockFunc: func(taskID int) error {
				require.Equal(t, 1, taskID)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Deleted successfully",
		},
		{
			name:           "Invalid ID",
			taskID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid task ID",
		},
		{
			name:   "Service Error",
			taskID: "1",
			mockFunc: func(taskID int) error {
				return errors.New("delete failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "delete failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTaskService{
				DeleteTaskFunc: tt.mockFunc,
			}

			h := setupTaskHandler(svc)
			r := setupTaskRouter(h)

			req := httptest.NewRequest(http.MethodDelete, "/tasks/"+tt.taskID, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestTaskHandler_AssignUser(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		mockFunc       func(taskID, userID, homeID int, date time.Time) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			body: validAssignUserReq,
			mockFunc: func(taskID, userID, homeID int, date time.Time) error {
				assert.Equal(t, 1, taskID)
				assert.Equal(t, 2, userID)
				assert.Equal(t, 3, homeID)
				return nil
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "Created successfully",
		},
		{
			name:           "Invalid JSON",
			body:           "{bad json}",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name: "Service Error",
			body: validAssignUserReq,
			mockFunc: func(taskID, userID, homeID int, date time.Time) error {
				return errors.New("assign failed")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTaskService{
				AssignUserFunc: tt.mockFunc,
			}

			h := setupTaskHandler(svc)

			var req *http.Request
			if tt.name == "Invalid JSON" {
				req = httptest.NewRequest(http.MethodPost, "/tasks/assign", bytes.NewBufferString("{bad json}"))
			} else {
				req = makeJSONRequest(http.MethodPost, "/tasks/assign", tt.body)
			}

			rr := httptest.NewRecorder()
			h.AssignUser(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestTaskHandler_GetAssignmentsForUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockFunc       func(userID int) (*[]models.TaskAssignment, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userID: "123",
			mockFunc: func(userID int) (*[]models.TaskAssignment, error) {
				require.Equal(t, 123, userID)
				assignments := []models.TaskAssignment{
					{ID: 1, TaskID: 1, UserID: 123},
					{ID: 2, TaskID: 2, UserID: 123},
				}
				return &assignments, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "assignments",
		},
		{
			name:           "Invalid ID",
			userID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid user ID",
		},
		{
			name:   "Service Error",
			userID: "123",
			mockFunc: func(userID int) (*[]models.TaskAssignment, error) {
				return nil, errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTaskService{
				GetAssignmentsForUserFunc: tt.mockFunc,
			}

			h := setupTaskHandler(svc)
			r := setupTaskRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID+"/assignments", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestTaskHandler_GetClosestAssignmentForUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockFunc       func(userID int) (*models.TaskAssignment, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userID: "123",
			mockFunc: func(userID int) (*models.TaskAssignment, error) {
				require.Equal(t, 123, userID)
				return &models.TaskAssignment{ID: 1, TaskID: 1, UserID: 123}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "assignment",
		},
		{
			name:           "Invalid ID",
			userID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid user ID",
		},
		{
			name:   "Service Error",
			userID: "123",
			mockFunc: func(userID int) (*models.TaskAssignment, error) {
				return nil, errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTaskService{
				GetClosestAssignmentForUserFunc: tt.mockFunc,
			}

			h := setupTaskHandler(svc)
			r := setupTaskRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID+"/assignments/closest", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestTaskHandler_MarkAssignmentCompleted(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		mockFunc       func(assignmentID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			body: models.AssignmentIDRequest{AssignmentID: 1},
			mockFunc: func(assignmentID int) error {
				require.Equal(t, 1, assignmentID)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Marked successfully",
		},
		{
			name:           "Invalid JSON",
			body:           "{bad json}",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name: "Service Error",
			body: models.AssignmentIDRequest{AssignmentID: 1},
			mockFunc: func(assignmentID int) error {
				return errors.New("mark failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "mark failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTaskService{
				MarkAssignmentCompletedFunc: tt.mockFunc,
			}

			h := setupTaskHandler(svc)

			var req *http.Request
			if tt.name == "Invalid JSON" {
				req = httptest.NewRequest(http.MethodPut, "/assignments/mark-completed", bytes.NewBufferString("{bad json}"))
			} else {
				req = makeJSONRequest(http.MethodPut, "/assignments/mark-completed", tt.body)
			}

			rr := httptest.NewRecorder()
			h.MarkAssignmentCompleted(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestTaskHandler_DeleteAssignment(t *testing.T) {
	tests := []struct {
		name           string
		assignmentID   string
		mockFunc       func(assignmentID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:         "Success",
			assignmentID: "1",
			mockFunc: func(assignmentID int) error {
				require.Equal(t, 1, assignmentID)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Deleted successfully",
		},
		{
			name:           "Invalid ID",
			assignmentID:   "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid assignment ID",
		},
		{
			name:         "Service Error",
			assignmentID: "1",
			mockFunc: func(assignmentID int) error {
				return errors.New("delete failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "delete failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTaskService{
				DeleteAssignmentFunc: tt.mockFunc,
			}

			h := setupTaskHandler(svc)
			r := setupTaskRouter(h)

			req := httptest.NewRequest(http.MethodDelete, "/assignments/"+tt.assignmentID, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestTaskHandler_ReassignRoom(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		mockFunc       func(taskID, roomID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			body: validReassignRoomReq,
			mockFunc: func(taskID, roomID int) error {
				assert.Equal(t, 1, taskID)
				assert.Equal(t, 2, roomID)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Updated successfully",
		},
		{
			name:           "Invalid JSON",
			body:           "{bad json}",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name: "Service Error",
			body: validReassignRoomReq,
			mockFunc: func(taskID, roomID int) error {
				return errors.New("reassign failed")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "reassign failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTaskService{
				ReassignRoomFunc: tt.mockFunc,
			}

			h := setupTaskHandler(svc)

			var req *http.Request
			if tt.name == "Invalid JSON" {
				req = httptest.NewRequest(http.MethodPut, "/tasks/reassign-room", bytes.NewBufferString("{bad json}"))
			} else {
				req = makeJSONRequest(http.MethodPut, "/tasks/reassign-room", tt.body)
			}

			rr := httptest.NewRecorder()
			h.ReassignRoom(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}
