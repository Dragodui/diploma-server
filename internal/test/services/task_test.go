package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock TaskRepository
type mockTaskRepo struct {
	CreateFunc                       func(ctx context.Context, t *models.Task) error
	FindByIDFunc                     func(ctx context.Context, id int) (*models.Task, error)
	FindByHomeIDFunc                 func(ctx context.Context, homeID int) (*[]models.Task, error)
	DeleteFunc                       func(ctx context.Context, id int) error
	ReassignRoomFunc                 func(ctx context.Context, taskID, roomID int) error
	AssignUserFunc                   func(ctx context.Context, taskID, userID int, date time.Time) error
	FindAssignmentsForUserFunc       func(ctx context.Context, userID int) (*[]models.TaskAssignment, error)
	FindClosestAssignmentForUserFunc func(ctx context.Context, userID int) (*models.TaskAssignment, error)
	FindAssignmentByTaskAndUserFunc  func(ctx context.Context, taskID, userID int) (*models.TaskAssignment, error)
	FindAssignmentByIDFunc           func(ctx context.Context, assignmentID int) (*models.TaskAssignment, error)
	MarkCompletedFunc                func(ctx context.Context, assignmentID int) error
	MarkUncompletedFunc              func(ctx context.Context, assignmentID int) error
	FindUserByAssignmentIDFunc       func(ctx context.Context, assignmentID int) (*models.User, error)
	DeleteAssignmentFunc             func(ctx context.Context, assignmentID int) error
}

func (m *mockTaskRepo) Create(ctx context.Context, t *models.Task) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, t)
	}
	return nil
}

func (m *mockTaskRepo) FindByID(ctx context.Context, id int) (*models.Task, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockTaskRepo) FindByHomeID(ctx context.Context, homeID int) (*[]models.Task, error) {
	if m.FindByHomeIDFunc != nil {
		return m.FindByHomeIDFunc(ctx, homeID)
	}
	return &[]models.Task{}, nil
}

func (m *mockTaskRepo) Delete(ctx context.Context, id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *mockTaskRepo) ReassignRoom(ctx context.Context, taskID, roomID int) error {
	if m.ReassignRoomFunc != nil {
		return m.ReassignRoomFunc(ctx, taskID, roomID)
	}
	return nil
}

func (m *mockTaskRepo) AssignUser(ctx context.Context, taskID, userID int, date time.Time) error {
	if m.AssignUserFunc != nil {
		return m.AssignUserFunc(ctx, taskID, userID, date)
	}
	return nil
}

func (m *mockTaskRepo) FindAssignmentsForUser(ctx context.Context, userID int) (*[]models.TaskAssignment, error) {
	if m.FindAssignmentsForUserFunc != nil {
		return m.FindAssignmentsForUserFunc(ctx, userID)
	}
	return &[]models.TaskAssignment{}, nil
}

func (m *mockTaskRepo) FindClosestAssignmentForUser(ctx context.Context, userID int) (*models.TaskAssignment, error) {
	if m.FindClosestAssignmentForUserFunc != nil {
		return m.FindClosestAssignmentForUserFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockTaskRepo) FindAssignmentByTaskAndUser(ctx context.Context, taskID, userID int) (*models.TaskAssignment, error) {
	if m.FindAssignmentByTaskAndUserFunc != nil {
		return m.FindAssignmentByTaskAndUserFunc(ctx, taskID, userID)
	}
	return nil, nil
}

func (m *mockTaskRepo) FindAssignmentByID(ctx context.Context, assignmentID int) (*models.TaskAssignment, error) {
	if m.FindAssignmentByIDFunc != nil {
		return m.FindAssignmentByIDFunc(ctx, assignmentID)
	}
	return nil, nil
}

func (m *mockTaskRepo) MarkCompleted(ctx context.Context, assignmentID int) error {
	if m.MarkCompletedFunc != nil {
		return m.MarkCompletedFunc(ctx, assignmentID)
	}
	return nil
}

func (m *mockTaskRepo) MarkUncompleted(ctx context.Context, assignmentID int) error {
	if m.MarkUncompletedFunc != nil {
		return m.MarkUncompletedFunc(ctx, assignmentID)
	}
	return nil
}

func (m *mockTaskRepo) FindUserByAssignmentID(ctx context.Context, assignmentID int) (*models.User, error) {
	if m.FindUserByAssignmentIDFunc != nil {
		return m.FindUserByAssignmentIDFunc(ctx, assignmentID)
	}
	return nil, nil
}

func (m *mockTaskRepo) DeleteAssignment(ctx context.Context, assignmentID int) error {
	if m.DeleteAssignmentFunc != nil {
		return m.DeleteAssignmentFunc(ctx, assignmentID)
	}
	return nil
}

// Test helpers
func setupTaskService(t *testing.T, repo repository.TaskRepository) *services.TaskService {
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	return services.NewTaskService(repo, redisClient)
}

// CreateTask Tests
func TestTaskService_CreateTask_Success(t *testing.T) {
	roomID := 5
	dueDate := time.Now().Add(24 * time.Hour)

	repo := &mockTaskRepo{
		CreateFunc: func(ctx context.Context, task *models.Task) error {
			require.Equal(t, "Clean Kitchen", task.Name)
			require.Equal(t, "Deep clean the kitchen", task.Description)
			require.Equal(t, 1, task.HomeID)
			require.Equal(t, &roomID, task.RoomID)
			require.Equal(t, "once", task.ScheduleType)
			require.NotNil(t, task.DueDate)
			return nil
		},
	}

	svc := setupTaskService(t, repo)
	err := svc.CreateTask(context.Background(), 1, &roomID, "Clean Kitchen", "Deep clean the kitchen", "once", &dueDate)
	assert.NoError(t, err)
}

func TestTaskService_CreateTask_WithoutRoomID(t *testing.T) {
	repo := &mockTaskRepo{
		CreateFunc: func(ctx context.Context, task *models.Task) error {
			require.Equal(t, "General Task", task.Name)
			require.Nil(t, task.RoomID)
			return nil
		},
	}

	svc := setupTaskService(t, repo)
	err := svc.CreateTask(context.Background(), 1, nil, "General Task", "Not room-specific", "weekly", nil)
	assert.NoError(t, err)
}

func TestTaskService_CreateTask_RepositoryError(t *testing.T) {
	repo := &mockTaskRepo{
		CreateFunc: func(ctx context.Context, task *models.Task) error {
			return errors.New("database error")
		},
	}

	svc := setupTaskService(t, repo)
	err := svc.CreateTask(context.Background(), 1, nil, "Task", "Description", "once", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

// GetTaskByID Tests
func TestTaskService_GetTaskByID_Success(t *testing.T) {
	expectedTask := &models.Task{
		ID:          10,
		Name:        "Clean Kitchen",
		Description: "Deep clean",
		HomeID:      1,
	}

	repo := &mockTaskRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*models.Task, error) {
			require.Equal(t, 10, id)
			return expectedTask, nil
		},
	}

	svc := setupTaskService(t, repo)
	task, err := svc.GetTaskByID(context.Background(), 10)

	assert.NoError(t, err)
	assert.Equal(t, expectedTask.ID, task.ID)
	assert.Equal(t, expectedTask.Name, task.Name)
}

func TestTaskService_GetTaskByID_NotFound(t *testing.T) {
	repo := &mockTaskRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*models.Task, error) {
			return nil, errors.New("task not found")
		},
	}

	svc := setupTaskService(t, repo)
	_, err := svc.GetTaskByID(context.Background(), 999)

	assert.Error(t, err)
}

// GetTasksByHomeID Tests
func TestTaskService_GetTasksByHomeID_Success(t *testing.T) {
	expectedTasks := &[]models.Task{
		{ID: 1, Name: "Task 1", HomeID: 1},
		{ID: 2, Name: "Task 2", HomeID: 1},
		{ID: 3, Name: "Task 3", HomeID: 1},
	}

	repo := &mockTaskRepo{
		FindByHomeIDFunc: func(ctx context.Context, homeID int) (*[]models.Task, error) {
			require.Equal(t, 1, homeID)
			return expectedTasks, nil
		},
	}

	svc := setupTaskService(t, repo)
	tasks, err := svc.GetTasksByHomeID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, *tasks, 3)
	assert.Equal(t, "Task 1", (*tasks)[0].Name)
}

func TestTaskService_GetTasksByHomeID_Empty(t *testing.T) {
	repo := &mockTaskRepo{
		FindByHomeIDFunc: func(ctx context.Context, homeID int) (*[]models.Task, error) {
			return &[]models.Task{}, nil
		},
	}

	svc := setupTaskService(t, repo)
	tasks, err := svc.GetTasksByHomeID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, *tasks, 0)
}

// DeleteTask Tests
func TestTaskService_DeleteTask_Success(t *testing.T) {
	repo := &mockTaskRepo{
		DeleteFunc: func(ctx context.Context, id int) error {
			require.Equal(t, 10, id)
			return nil
		},
	}

	svc := setupTaskService(t, repo)
	err := svc.DeleteTask(context.Background(), 10)
	assert.NoError(t, err)
}

func TestTaskService_DeleteTask_NotFound(t *testing.T) {
	repo := &mockTaskRepo{
		DeleteFunc: func(ctx context.Context, id int) error {
			return errors.New("task not found")
		},
	}

	svc := setupTaskService(t, repo)
	err := svc.DeleteTask(context.Background(), 999)
	assert.Error(t, err)
}

// AssignUser Tests
func TestTaskService_AssignUser_Success(t *testing.T) {
	assignDate := time.Now().Add(24 * time.Hour)

	repo := &mockTaskRepo{
		AssignUserFunc: func(ctx context.Context, taskID, userID int, date time.Time) error {
			require.Equal(t, 10, taskID)
			require.Equal(t, 5, userID)
			require.Equal(t, assignDate.Unix(), date.Unix())
			return nil
		},
	}

	svc := setupTaskService(t, repo)
	err := svc.AssignUser(context.Background(), 10, 5, 1, assignDate)
	assert.NoError(t, err)
}

// GetAssignmentsForUser Tests
func TestTaskService_GetAssignmentsForUser_Success(t *testing.T) {
	expectedAssignments := &[]models.TaskAssignment{
		{ID: 1, TaskID: 1, UserID: 5},
		{ID: 2, TaskID: 2, UserID: 5},
	}

	repo := &mockTaskRepo{
		FindAssignmentsForUserFunc: func(ctx context.Context, userID int) (*[]models.TaskAssignment, error) {
			require.Equal(t, 5, userID)
			return expectedAssignments, nil
		},
	}

	svc := setupTaskService(t, repo)
	assignments, err := svc.GetAssignmentsForUser(context.Background(), 5)

	assert.NoError(t, err)
	assert.Len(t, *assignments, 2)
}

func TestTaskService_GetAssignmentsForUser_Empty(t *testing.T) {
	repo := &mockTaskRepo{
		FindAssignmentsForUserFunc: func(ctx context.Context, userID int) (*[]models.TaskAssignment, error) {
			return &[]models.TaskAssignment{}, nil
		},
	}

	svc := setupTaskService(t, repo)
	assignments, err := svc.GetAssignmentsForUser(context.Background(), 5)

	assert.NoError(t, err)
	assert.Len(t, *assignments, 0)
}

// GetClosestAssignmentForUser Tests
func TestTaskService_GetClosestAssignmentForUser_Success(t *testing.T) {
	closestAssignment := &models.TaskAssignment{
		ID:           1,
		TaskID:       10,
		UserID:       5,
		AssignedDate: time.Now().Add(2 * time.Hour),
	}

	repo := &mockTaskRepo{
		FindClosestAssignmentForUserFunc: func(ctx context.Context, userID int) (*models.TaskAssignment, error) {
			require.Equal(t, 5, userID)
			return closestAssignment, nil
		},
	}

	svc := setupTaskService(t, repo)
	assignment, err := svc.GetClosestAssignmentForUser(context.Background(), 5)

	assert.NoError(t, err)
	assert.Equal(t, closestAssignment.ID, assignment.ID)
	assert.Equal(t, closestAssignment.TaskID, assignment.TaskID)
}

func TestTaskService_GetClosestAssignmentForUser_NotFound(t *testing.T) {
	repo := &mockTaskRepo{
		FindClosestAssignmentForUserFunc: func(ctx context.Context, userID int) (*models.TaskAssignment, error) {
			return nil, errors.New("no assignments found")
		},
	}

	svc := setupTaskService(t, repo)
	_, err := svc.GetClosestAssignmentForUser(context.Background(), 5)

	assert.Error(t, err)
}

// MarkAssignmentCompleted Tests
func TestTaskService_MarkAssignmentCompleted_Success(t *testing.T) {
	repo := &mockTaskRepo{
		MarkCompletedFunc: func(ctx context.Context, assignmentID int) error {
			require.Equal(t, 10, assignmentID)
			return nil
		},
	}

	svc := setupTaskService(t, repo)
	err := svc.MarkAssignmentCompleted(context.Background(), 10)
	assert.NoError(t, err)
}

func TestTaskService_MarkAssignmentCompleted_NotFound(t *testing.T) {
	repo := &mockTaskRepo{
		MarkCompletedFunc: func(ctx context.Context, assignmentID int) error {
			return errors.New("assignment not found")
		},
	}

	svc := setupTaskService(t, repo)
	err := svc.MarkAssignmentCompleted(context.Background(), 999)
	assert.Error(t, err)
}

// MarkAssignmentUncompleted Tests
func TestTaskService_MarkAssignmentUncompleted_Success(t *testing.T) {
	repo := &mockTaskRepo{
		MarkUncompletedFunc: func(ctx context.Context, assignmentID int) error {
			require.Equal(t, 10, assignmentID)
			return nil
		},
	}

	svc := setupTaskService(t, repo)
	err := svc.MarkAssignmentUncompleted(context.Background(), 10)
	assert.NoError(t, err)
}

// DeleteAssignment Tests
func TestTaskService_DeleteAssignment_Success(t *testing.T) {
	repo := &mockTaskRepo{
		DeleteAssignmentFunc: func(ctx context.Context, assignmentID int) error {
			require.Equal(t, 10, assignmentID)
			return nil
		},
	}

	svc := setupTaskService(t, repo)
	err := svc.DeleteAssignment(context.Background(), 10)
	assert.NoError(t, err)
}

func TestTaskService_DeleteAssignment_NotFound(t *testing.T) {
	repo := &mockTaskRepo{
		DeleteAssignmentFunc: func(ctx context.Context, assignmentID int) error {
			return errors.New("assignment not found")
		},
	}

	svc := setupTaskService(t, repo)
	err := svc.DeleteAssignment(context.Background(), 999)
	assert.Error(t, err)
}

// ReassignRoom Tests
func TestTaskService_ReassignRoom_Success(t *testing.T) {
	repo := &mockTaskRepo{
		ReassignRoomFunc: func(ctx context.Context, taskID, roomID int) error {
			require.Equal(t, 10, taskID)
			require.Equal(t, 3, roomID)
			return nil
		},
	}

	svc := setupTaskService(t, repo)
	err := svc.ReassignRoom(context.Background(), 10, 3)
	assert.NoError(t, err)
}

func TestTaskService_ReassignRoom_TaskNotFound(t *testing.T) {
	repo := &mockTaskRepo{
		ReassignRoomFunc: func(ctx context.Context, taskID, roomID int) error {
			return errors.New("task not found")
		},
	}

	svc := setupTaskService(t, repo)
	err := svc.ReassignRoom(context.Background(), 999, 3)
	assert.Error(t, err)
}
