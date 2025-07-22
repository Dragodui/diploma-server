package services

import (
	"errors"
	"log"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/redis/go-redis/v9"
)

type TaskService struct {
	tasks repository.TaskRepository
	cache *redis.Client
}

func NewTaskService(repo repository.TaskRepository, cache *redis.Client) *TaskService {
	return &TaskService{tasks: repo, cache: cache}
}

func (s *TaskService) CreateTask(homeID int, roomID *int, name, description, scheduleType string) error {
	if err := s.tasks.Create(&models.Task{
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

func (s *TaskService) GetTaskByID(taskID int) (*models.Task, error) {
	key := utils.GetTaskKey(taskID)

	// try to get from cache
	cached, err := utils.GetFromCache[models.Task](key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	task, err := s.tasks.FindByID(taskID)
	if err != nil {
		return nil, err
	}

	// save to cache
	if err := utils.WriteToCache(key, task, s.cache); err != nil {
		log.Printf("Failed to write to cache [%s]: %v", key, err)
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

	tasks, err := s.tasks.FindByHomeID(homeID)
	if err != nil {
		return nil, err
	}

	// save to cache
	if err := utils.WriteToCache(key, tasks, s.cache); err != nil {
		log.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return tasks, nil
}

func (s *TaskService) DeleteTask(taskID int) error {
	// find task to get homeID
	task, err := s.tasks.FindByID(taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	if err := s.tasks.Delete(taskID); err != nil {
		return err
	}

	// delete task from cache
	taskKey := utils.GetTaskKey(taskID)
	if err := utils.DeleteFromCache(taskKey, s.cache); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", taskKey, err)
	}

	// delete tasks for home from cache
	homeTasksKey := utils.GetTasksForHomeKey(task.HomeID)
	if err := utils.DeleteFromCache(homeTasksKey, s.cache); err != nil {
		log.Printf("Failed to delete redis cache for home %d: %v", task.HomeID, err)
	}

	return nil
}

func (s *TaskService) AssignUser(taskID, userID, homeID int, date time.Time) error {
	// delete task from cache
	key := utils.GetTaskKey(taskID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		log.Printf("Failed to delete redis cache for task %d: %v", taskID, err)
	}

	// delete tasks for home from cache
	homeTasksKey := utils.GetTasksForHomeKey(homeID)
	if err := utils.DeleteFromCache(homeTasksKey, s.cache); err != nil {
		log.Printf("Failed to delete redis cache for home %d: %v", homeID, err)
	}

	if err := s.tasks.AssignUser(taskID, userID, date); err != nil {
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

	assignments, err := s.tasks.FindAssignmentsForUser(userID)
	if err != nil {
		return nil, err
	}

	// save to cache
	if err := utils.WriteToCache(key, assignments, s.cache); err != nil {
		log.Printf("Failed to write to cache [%s]: %v", key, err)
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
	assignment, err := s.tasks.FindClosestAssignmentForUser(userID)
	if err != nil {
		return nil, err
	}

	// save to cache
	if err := utils.WriteToCache(key, assignment, s.cache); err != nil {
		log.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return assignment, nil
}

func (s *TaskService) MarkAssignmentCompleted(assignmentID int) error {
	// delete assignment from cache
	key := utils.GetAssignmentKey(assignmentID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	user, err := s.tasks.FindUserByAssignmentID(assignmentID)
	if err != nil {
		return err
	}

	// delete user assignments from cache
	userAssignmentsKey := utils.GetAssignmentsForUserKey(user.ID)
	if err := utils.DeleteFromCache(userAssignmentsKey, s.cache); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", userAssignmentsKey, err)
	}

	// delete closest user assignment from cache
	userClosestAssignmentsKey := utils.GetClosestAssignmentsForUserKey(user.ID)
	if err := utils.DeleteFromCache(userClosestAssignmentsKey, s.cache); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", userClosestAssignmentsKey, err)
	}
	if err := s.tasks.MarkCompleted(assignmentID); err != nil {
		return err
	}

	return nil
}

func (s *TaskService) DeleteAssignment(assignmentID int) error {
	// delete assignment from cache
	key := utils.GetAssignmentKey(assignmentID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	user, err := s.tasks.FindUserByAssignmentID(assignmentID)
	if err != nil {
		return err
	}

	// delete user assignments from cache
	userAssignmentsKey := utils.GetAssignmentsForUserKey(user.ID)
	if err := utils.DeleteFromCache(userAssignmentsKey, s.cache); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", userAssignmentsKey, err)
	}

	// delete closest user assignment from cache
	userClosestAssignmentsKey := utils.GetClosestAssignmentsForUserKey(user.ID)
	if err := utils.DeleteFromCache(userClosestAssignmentsKey, s.cache); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", userClosestAssignmentsKey, err)
	}
	if err := s.tasks.DeleteAssignment(assignmentID); err != nil {
		return err
	}

	return nil
}

func (s *TaskService) ReassignRoom(taskID, roomID int) error {
	// delete from cache
	taskKey := utils.GetTaskKey(taskID)
	if err := utils.DeleteFromCache(taskKey, s.cache); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", taskKey, err)
	}

	roomKey := utils.GetRoomKey(roomID)
	if err := utils.DeleteFromCache(roomKey, s.cache); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", roomKey, err)
	}

	if err := s.tasks.ReassignRoom(taskID, roomID); err != nil {
		return err
	}

	return nil
}
