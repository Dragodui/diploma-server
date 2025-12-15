package services

import (
	"encoding/json"
	"errors"
	"time"

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
	CreateTask(homeID int, roomID *int, name, description, scheduleType string) error
	CreateTaskWithAssignment(homeID int, roomID *int, name, description, scheduleType string, userID int) error
	GetTaskByID(taskID int) (*models.Task, error)
	GetTasksByHomeID(homeID int) (*[]models.Task, error)
	DeleteTask(taskID int) error
	AssignUser(taskID, userID, homeID int, date time.Time) error
	GetAssignmentsForUser(userID int) (*[]models.TaskAssignment, error)
	GetClosestAssignmentForUser(userID int) (*models.TaskAssignment, error)
	MarkAssignmentCompleted(assignmentID int) error
	MarkTaskCompletedForUser(taskID, userID, homeID int) error
	DeleteAssignment(assignmentID int) error
	ReassignRoom(taskID, roomID int) error
}

func NewTaskService(repo repository.TaskRepository, cache *redis.Client) *TaskService {
	return &TaskService{repo: repo, cache: cache}
}

func (s *TaskService) CreateTask(homeID int, roomID *int, name, description, scheduleType string) error {
	tasksKey := utils.GetTasksForHomeKey(homeID)
	if err := utils.DeleteFromCache(tasksKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", tasksKey, err)
	}

	if err := s.repo.Create(&models.Task{
		Name:         name,
		Description:  description,
		HomeID:       homeID,
		RoomID:       roomID,
		ScheduleType: scheduleType,
	}); err != nil {
		return err
	}

	return nil
}

func (s *TaskService) CreateTaskWithAssignment(homeID int, roomID *int, name, description, scheduleType string, userID int) error {
	tasksKey := utils.GetTasksForHomeKey(homeID)
	if err := utils.DeleteFromCache(tasksKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", tasksKey, err)
	}

	task := &models.Task{
		Name:         name,
		Description:  description,
		HomeID:       homeID,
		RoomID:       roomID,
		ScheduleType: scheduleType,
	}

	if err := s.repo.Create(task); err != nil {
		return err
	}

	// Auto-assign to user
	if err := s.repo.AssignUser(task.ID, userID, time.Now()); err != nil {
		return err
	}

	return nil
}

func (s *TaskService) GetTaskByID(taskID int) (*models.Task, error) {
	key := utils.GetTaskKey(taskID)

	// try to get from cache
	cached, err := utils.GetFromCache[models.Task](key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	task, err := s.repo.FindByID(taskID)
	if err != nil {
		return nil, err
	}

	// save to cache
	if err := utils.WriteToCache(key, task, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return task, nil
}

func (s *TaskService) GetTasksByHomeID(homeID int) (*[]models.Task, error) {
	key := utils.GetTasksForHomeKey(homeID)

	// try to get from cache
	cached, err := utils.GetFromCache[[]models.Task](key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	tasks, err := s.repo.FindByHomeID(homeID)
	if err != nil {
		return nil, err
	}

	// save to cache
	if err := utils.WriteToCache(key, tasks, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return tasks, nil
}

func (s *TaskService) DeleteTask(taskID int) error {
	// find task to get homeID
	task, err := s.repo.FindByID(taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	if err := s.repo.Delete(taskID); err != nil {
		return err
	}

	// delete task from cache
	taskKey := utils.GetTaskKey(taskID)
	if err := utils.DeleteFromCache(taskKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", taskKey, err)
	}

	// delete tasks for home from cache
	homeTasksKey := utils.GetTasksForHomeKey(task.HomeID)
	if err := utils.DeleteFromCache(homeTasksKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for home %d: %v", task.HomeID, err)
	}

	return nil
}

func (s *TaskService) AssignUser(taskID, userID, homeID int, date time.Time) error {
	// delete task from cache
	key := utils.GetTaskKey(taskID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for task %d: %v", taskID, err)
	}

	// delete tasks for home from cache
	homeTasksKey := utils.GetTasksForHomeKey(homeID)
	if err := utils.DeleteFromCache(homeTasksKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for home %d: %v", homeID, err)
	}

	if err := s.repo.AssignUser(taskID, userID, date); err != nil {
		return err
	}

	return nil
}

func (s *TaskService) GetAssignmentsForUser(userID int) (*[]models.TaskAssignment, error) {
	// get assignments from cache if exists
	key := utils.GetAssignmentsForUserKey(userID)
	cached, err := utils.GetFromCache[[]models.TaskAssignment](key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	assignments, err := s.repo.FindAssignmentsForUser(userID)
	if err != nil {
		return nil, err
	}

	// save to cache
	if err := utils.WriteToCache(key, assignments, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return assignments, nil

}

func (s *TaskService) GetClosestAssignmentForUser(userID int) (*models.TaskAssignment, error) {
	// get assignment form cache if exists
	key := utils.GetClosestAssignmentsForUserKey(userID)
	cached, err := utils.GetFromCache[models.TaskAssignment](key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}
	assignment, err := s.repo.FindClosestAssignmentForUser(userID)
	ass_str, _ := json.Marshal(assignment)
	logger.Info.Printf(string(ass_str))
	if err != nil {
		return nil, err
	}

	// save to cache
	if err := utils.WriteToCache(key, assignment, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return assignment, nil
}

func (s *TaskService) MarkAssignmentCompleted(assignmentID int) error {
	// delete assignment from cache
	key := utils.GetAssignmentKey(assignmentID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	user, err := s.repo.FindUserByAssignmentID(assignmentID)
	if err != nil {
		return err
	}

	// delete user assignments from cache
	userAssignmentsKey := utils.GetAssignmentsForUserKey(user.ID)
	if err := utils.DeleteFromCache(userAssignmentsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", userAssignmentsKey, err)
	}

	// delete closest user assignment from cache
	userClosestAssignmentsKey := utils.GetClosestAssignmentsForUserKey(user.ID)
	if err := utils.DeleteFromCache(userClosestAssignmentsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", userClosestAssignmentsKey, err)
	}
	if err := s.repo.MarkCompleted(assignmentID); err != nil {
		return err
	}

	return nil
}

func (s *TaskService) MarkTaskCompletedForUser(taskID, userID, homeID int) error {
	// Find assignment by task and user
	assignment, err := s.repo.FindAssignmentByTaskAndUser(taskID, userID)
	if err != nil {
		return err
	}

	// If no assignment exists, create one and mark it completed
	if assignment == nil {
		if err := s.repo.AssignUser(taskID, userID, time.Now()); err != nil {
			return err
		}
		// Find the newly created assignment
		assignment, err = s.repo.FindAssignmentByTaskAndUser(taskID, userID)
		if err != nil {
			return err
		}
	}

	// Clear caches
	key := utils.GetAssignmentKey(assignment.ID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	userAssignmentsKey := utils.GetAssignmentsForUserKey(userID)
	if err := utils.DeleteFromCache(userAssignmentsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", userAssignmentsKey, err)
	}

	userClosestAssignmentsKey := utils.GetClosestAssignmentsForUserKey(userID)
	if err := utils.DeleteFromCache(userClosestAssignmentsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", userClosestAssignmentsKey, err)
	}

	tasksKey := utils.GetTasksForHomeKey(homeID)
	if err := utils.DeleteFromCache(tasksKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", tasksKey, err)
	}

	return s.repo.MarkCompleted(assignment.ID)
}

func (s *TaskService) DeleteAssignment(assignmentID int) error {
	// delete assignment from cache
	key := utils.GetAssignmentKey(assignmentID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	user, err := s.repo.FindUserByAssignmentID(assignmentID)
	if err != nil {
		return err
	}

	// delete user assignments from cache
	userAssignmentsKey := utils.GetAssignmentsForUserKey(user.ID)
	if err := utils.DeleteFromCache(userAssignmentsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", userAssignmentsKey, err)
	}

	// delete closest user assignment from cache
	userClosestAssignmentsKey := utils.GetClosestAssignmentsForUserKey(user.ID)
	if err := utils.DeleteFromCache(userClosestAssignmentsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", userClosestAssignmentsKey, err)
	}
	if err := s.repo.DeleteAssignment(assignmentID); err != nil {
		return err
	}

	return nil
}

func (s *TaskService) ReassignRoom(taskID, roomID int) error {
	// delete from cache
	taskKey := utils.GetTaskKey(taskID)
	if err := utils.DeleteFromCache(taskKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", taskKey, err)
	}

	roomKey := utils.GetRoomKey(roomID)
	if err := utils.DeleteFromCache(roomKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", roomKey, err)
	}

	task, err := s.repo.FindByID(taskID)
	if err != nil {
		return err
	}
	homeTasksKey := utils.GetTasksForHomeKey(task.HomeID)
	if err := utils.DeleteFromCache(homeTasksKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", homeTasksKey, err)
	}

	if err := s.repo.ReassignRoom(taskID, roomID); err != nil {
		return err
	}

	return nil
}
