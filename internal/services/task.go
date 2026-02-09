package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Dragodui/diploma-server/internal/event"
	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/redis/go-redis/v9"
)

type TaskService struct {
	repo  repository.TaskRepository
	cache *redis.Client
}

type ITaskService interface {
	CreateTask(ctx context.Context, homeID int, roomID *int, name, description, scheduleType string, dueDate *time.Time) error
	CreateTaskWithAssignment(ctx context.Context, homeID int, roomID *int, name, description, scheduleType string, dueDate *time.Time, userID int) error
	GetTaskByID(ctx context.Context, taskID int) (*models.Task, error)
	GetTasksByHomeID(ctx context.Context, homeID int) (*[]models.Task, error)
	DeleteTask(ctx context.Context, taskID int) error
	AssignUser(ctx context.Context, taskID, userID, homeID int, date time.Time) error
	GetAssignmentsForUser(ctx context.Context, userID int) (*[]models.TaskAssignment, error)
	GetClosestAssignmentForUser(ctx context.Context, userID int) (*models.TaskAssignment, error)
	MarkAssignmentCompleted(ctx context.Context, assignmentID int) error
	MarkAssignmentUncompleted(ctx context.Context, assignmentID int) error
	MarkTaskCompletedForUser(ctx context.Context, taskID, userID, homeID int) error
	DeleteAssignment(ctx context.Context, assignmentID int) error
	ReassignRoom(ctx context.Context, taskID, roomID int) error
}

func NewTaskService(repo repository.TaskRepository, cache *redis.Client) *TaskService {
	return &TaskService{repo: repo, cache: cache}
}

func (s *TaskService) CreateTask(ctx context.Context, homeID int, roomID *int, name, description, scheduleType string, dueDate *time.Time) error {
	tasksKey := utils.GetTasksForHomeKey(homeID)
	if err := utils.DeleteFromCache(ctx, tasksKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", tasksKey, err)
	}

	task := &models.Task{
		Name:         name,
		Description:  description,
		HomeID:       homeID,
		RoomID:       roomID,
		ScheduleType: scheduleType,
		DueDate:      dueDate,
	}
	if err := s.repo.Create(ctx, task); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleTask,
		Action: event.ActionCreated,
		Data:   task,
	})

	return nil
}

func (s *TaskService) CreateTaskWithAssignment(ctx context.Context, homeID int, roomID *int, name, description, scheduleType string, dueDate *time.Time, userID int) error {
	tasksKey := utils.GetTasksForHomeKey(homeID)
	if err := utils.DeleteFromCache(ctx, tasksKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", tasksKey, err)
	}

	task := &models.Task{
		Name:         name,
		Description:  description,
		HomeID:       homeID,
		RoomID:       roomID,
		ScheduleType: scheduleType,
		DueDate:      dueDate,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return err
	}

	// Auto-assign to user
	if err := s.repo.AssignUser(ctx, task.ID, userID, time.Now()); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleTask,
		Action: event.ActionCreated,
		Data:   task,
	})

	return nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, taskID int) (*models.Task, error) {
	key := utils.GetTaskKey(taskID)

	// try to get from cache
	cached, err := utils.GetFromCache[models.Task](ctx, key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	task, err := s.repo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// save to cache
	if err := utils.WriteToCache(ctx, key, task, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return task, nil
}

func (s *TaskService) GetTasksByHomeID(ctx context.Context, homeID int) (*[]models.Task, error) {
	key := utils.GetTasksForHomeKey(homeID)

	// try to get from cache
	cached, err := utils.GetFromCache[[]models.Task](ctx, key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	tasks, err := s.repo.FindByHomeID(ctx, homeID)
	if err != nil {
		return nil, err
	}

	// save to cache
	if err := utils.WriteToCache(ctx, key, tasks, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return tasks, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, taskID int) error {
	// find task to get homeID
	task, err := s.repo.FindByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	if err := s.repo.Delete(ctx, taskID); err != nil {
		return err
	}

	// delete task from cache
	taskKey := utils.GetTaskKey(taskID)
	if err := utils.DeleteFromCache(ctx, taskKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", taskKey, err)
	}

	// delete tasks for home from cache
	homeTasksKey := utils.GetTasksForHomeKey(task.HomeID)
	if err := utils.DeleteFromCache(ctx, homeTasksKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for home %d: %v", task.HomeID, err)
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleTask,
		Action: event.ActionDeleted,
		Data:   task,
	})

	return nil
}

func (s *TaskService) AssignUser(ctx context.Context, taskID, userID, homeID int, date time.Time) error {
	// delete task from cache
	key := utils.GetTaskKey(taskID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for task %d: %v", taskID, err)
	}

	// delete tasks for home from cache
	homeTasksKey := utils.GetTasksForHomeKey(homeID)
	if err := utils.DeleteFromCache(ctx, homeTasksKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for home %d: %v", homeID, err)
	}

	if err := s.repo.AssignUser(ctx, taskID, userID, date); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleTask,
		Action: event.ActionAssigned,
		Data:   map[string]int{"taskID": taskID, "userID": userID},
	})

	return nil
}

func (s *TaskService) GetAssignmentsForUser(ctx context.Context, userID int) (*[]models.TaskAssignment, error) {
	// get assignments from cache if exists
	key := utils.GetAssignmentsForUserKey(userID)
	cached, err := utils.GetFromCache[[]models.TaskAssignment](ctx, key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	assignments, err := s.repo.FindAssignmentsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// save to cache
	if err := utils.WriteToCache(ctx, key, assignments, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return assignments, nil

}

func (s *TaskService) GetClosestAssignmentForUser(ctx context.Context, userID int) (*models.TaskAssignment, error) {
	// get assignment form cache if exists
	key := utils.GetClosestAssignmentsForUserKey(userID)
	cached, err := utils.GetFromCache[models.TaskAssignment](ctx, key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}
	assignment, err := s.repo.FindClosestAssignmentForUser(ctx, userID)
	ass_str, _ := json.Marshal(assignment)
	logger.Info.Printf("%s", string(ass_str))
	if err != nil {
		return nil, err
	}

	// save to cache
	if err := utils.WriteToCache(ctx, key, assignment, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return assignment, nil
}

func (s *TaskService) MarkAssignmentCompleted(ctx context.Context, assignmentID int) error {
	// delete assignment from cache
	key := utils.GetAssignmentKey(assignmentID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	assignment, err := s.repo.FindAssignmentByID(ctx, assignmentID)
	if err != nil {
		return err
	}

	// delete user assignments from cache
	userAssignmentsKey := utils.GetAssignmentsForUserKey(assignment.UserID)
	if err := utils.DeleteFromCache(ctx, userAssignmentsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", userAssignmentsKey, err)
	}

	// delete closest user assignment from cache
	userClosestAssignmentsKey := utils.GetClosestAssignmentsForUserKey(assignment.UserID)
	if err := utils.DeleteFromCache(ctx, userClosestAssignmentsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", userClosestAssignmentsKey, err)
	}

	// delete home tasks cache
	homeTasksKey := utils.GetTasksForHomeKey(assignment.Task.HomeID)
	if err := utils.DeleteFromCache(ctx, homeTasksKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for home %d: %v", assignment.Task.HomeID, err)
	}

	if err := s.repo.MarkCompleted(ctx, assignmentID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleTask,
		Action: event.ActionCompleted,
		Data:   assignment,
	})

	return nil
}

func (s *TaskService) MarkAssignmentUncompleted(ctx context.Context, assignmentID int) error {
	// delete assignment from cache
	key := utils.GetAssignmentKey(assignmentID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	assignment, err := s.repo.FindAssignmentByID(ctx, assignmentID)
	if err != nil {
		return err
	}

	// delete user assignments from cache
	userAssignmentsKey := utils.GetAssignmentsForUserKey(assignment.UserID)
	if err := utils.DeleteFromCache(ctx, userAssignmentsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", userAssignmentsKey, err)
	}

	// delete closest user assignment from cache
	userClosestAssignmentsKey := utils.GetClosestAssignmentsForUserKey(assignment.UserID)
	if err := utils.DeleteFromCache(ctx, userClosestAssignmentsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", userClosestAssignmentsKey, err)
	}

	// delete home tasks cache
	homeTasksKey := utils.GetTasksForHomeKey(assignment.Task.HomeID)
	if err := utils.DeleteFromCache(ctx, homeTasksKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for home %d: %v", assignment.Task.HomeID, err)
	}

	if err := s.repo.MarkUncompleted(ctx, assignmentID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleTask,
		Action: event.ActionUncompleted,
		Data:   assignment,
	})

	return nil
}

func (s *TaskService) MarkTaskCompletedForUser(ctx context.Context, taskID, userID, homeID int) error {
	// Find assignment by task and user
	assignment, err := s.repo.FindAssignmentByTaskAndUser(ctx, taskID, userID)
	if err != nil {
		return err
	}

	// If no assignment exists, create one and mark it completed
	if assignment == nil {
		if err := s.repo.AssignUser(ctx, taskID, userID, time.Now()); err != nil {
			return err
		}
		// Find the newly created assignment
		assignment, err = s.repo.FindAssignmentByTaskAndUser(ctx, taskID, userID)
		if err != nil {
			return err
		}
	}

	// Clear caches
	key := utils.GetAssignmentKey(assignment.ID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	userAssignmentsKey := utils.GetAssignmentsForUserKey(userID)
	if err := utils.DeleteFromCache(ctx, userAssignmentsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", userAssignmentsKey, err)
	}

	userClosestAssignmentsKey := utils.GetClosestAssignmentsForUserKey(userID)
	if err := utils.DeleteFromCache(ctx, userClosestAssignmentsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", userClosestAssignmentsKey, err)
	}

	tasksKey := utils.GetTasksForHomeKey(homeID)
	if err := utils.DeleteFromCache(ctx, tasksKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	if err := s.repo.MarkCompleted(ctx, assignment.ID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleTask,
		Action: event.ActionCompleted,
		Data:   assignment,
	})

	return nil
}

func (s *TaskService) DeleteAssignment(ctx context.Context, assignmentID int) error {
	// delete assignment from cache
	key := utils.GetAssignmentKey(assignmentID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	user, err := s.repo.FindUserByAssignmentID(ctx, assignmentID)
	if err != nil {
		return err
	}

	// delete user assignments from cache
	userAssignmentsKey := utils.GetAssignmentsForUserKey(user.ID)
	if err := utils.DeleteFromCache(ctx, userAssignmentsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", userAssignmentsKey, err)
	}

	// delete closest user assignment from cache
	userClosestAssignmentsKey := utils.GetClosestAssignmentsForUserKey(user.ID)
	if err := utils.DeleteFromCache(ctx, userClosestAssignmentsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", userClosestAssignmentsKey, err)
	}
	if err := s.repo.DeleteAssignment(ctx, assignmentID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleTask,
		Action: event.ActionDeleted,
		Data:   map[string]int{"assignmentID": assignmentID},
	})

	return nil
}

func (s *TaskService) ReassignRoom(ctx context.Context, taskID, roomID int) error {
	// delete from cache
	taskKey := utils.GetTaskKey(taskID)
	if err := utils.DeleteFromCache(ctx, taskKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", taskKey, err)
	}

	roomKey := utils.GetRoomKey(roomID)
	if err := utils.DeleteFromCache(ctx, roomKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", roomKey, err)
	}

	task, err := s.repo.FindByID(ctx, taskID)
	if err != nil {
		return err
	}
	homeTasksKey := utils.GetTasksForHomeKey(task.HomeID)
	if err := utils.DeleteFromCache(ctx, homeTasksKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", homeTasksKey, err)
	}

	if err := s.repo.ReassignRoom(ctx, taskID, roomID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleTask,
		Action: event.ActionUpdated,
		Data:   task,
	})

	return nil
}
