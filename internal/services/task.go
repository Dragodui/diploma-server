package services

import (
	"context"
	"encoding/json"
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

func (s *TaskService) CreateTask(homeID int, name, description, scheduleType string) error {
	if err := s.tasks.Create(&models.Task{
		Name:         name,
		Description:  description,
		HomeID:       homeID,
		ScheduleType: scheduleType,
	}); err != nil {
		return err
	}

	return nil
}

func (s *TaskService) GetTaskByID(taskID int) (*models.Task, error) {
	key := utils.GetTaskKey(taskID)

	// try to get from cache
	cached, err := s.cache.Get(context.Background(), key).Result()
	if err == nil && cached != "" {
		var task models.Task
		if err := json.Unmarshal([]byte(cached), &task); err == nil {
			return &task, nil
		}
	}

	task, err := s.tasks.FindByID(taskID)
	if err != nil {
		return nil, err
	}

	// save to cache
	data, err := json.Marshal(task)
	if err == nil {
		s.cache.Set(context.Background(), key, data, time.Hour)
	}

	return task, nil
}

func (s *TaskService) GetTasksByHomeID(homeID int) (*[]models.Task, error) {
	key := utils.GetTasksForHomeKey(homeID)

	// try to get from cache
	cached, err := s.cache.Get(context.Background(), key).Result()
	if err == nil && cached != "" {
		var tasks []models.Task
		if err := json.Unmarshal([]byte(cached), &tasks); err == nil {
			return &tasks, nil
		}
	}

	tasks, err := s.tasks.FindByHomeID(homeID)
	if err != nil {
		return nil, err
	}

	// save to cache
	data, err := json.Marshal(tasks)
	if err == nil {
		s.cache.Set(context.Background(), key, data, time.Hour)
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
	if err := s.cache.Del(context.Background(), taskKey).Err(); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", taskKey, err)
	}

	// delete tasks for home from cache
	homeTasksKey := utils.GetTasksForHomeKey(task.HomeID)
	if err := s.cache.Del(context.Background(), homeTasksKey).Err(); err != nil {
		log.Printf("Failed to delete redis cache for home %d: %v", task.HomeID, err)
	}

	return nil
}

func (s *TaskService) AssignUser(taskID, userID, homeID int, date time.Time) error {
	// delete task from cache
	key := utils.GetTaskKey(taskID)
	if err := s.cache.Del(context.Background(), key).Err(); err != nil {
		log.Printf("Failed to delete redis cache for task %d: %v", taskID, err)
	}

	// delete tasks for home from cache
	homeTasksKey := utils.GetTasksForHomeKey(homeID)
	if err := s.cache.Del(context.Background(), homeTasksKey).Err(); err != nil {
		log.Printf("Failed to delete redis cache for home %d: %v", homeID, err)
	}

	if err := s.tasks.AssignUser(taskID, userID, homeID, date); err != nil {
		return err
	}

	return nil
}

func (s *TaskService) GetAssignmentsForUser(userID int) (*[]models.TaskAssignment, error) {
	// get assignments from cache if exists
	key := utils.GetAssignmentsForUserKey(userID)
	cached, err := s.cache.Get(context.Background(), key).Result()
	if cached != "" && err == nil {
		var assignments []models.TaskAssignment
		if err := json.Unmarshal([]byte(cached), &assignments); err == nil {
			return &assignments, nil
		}
	}

	assignments, err := s.tasks.FindAssignmentsForUser(userID)
	if err != nil {
		return nil, err
	}

	// save to cache
	data, err := json.Marshal(assignments)
	if err == nil {
		s.cache.Set(context.Background(), key, data, time.Hour)
	}

	return assignments, nil

}

func (s *TaskService) GetClosestAssignmentForUser(userID int) (*models.TaskAssignment, error) {
	// get assignment form cache if exists
	key := utils.GetClosestAssignmentsForUserKey(userID)
	cached, err := s.cache.Get(context.Background(), key).Result()
	if cached != "" && err == nil {
		var assignment models.TaskAssignment
		if err := json.Unmarshal([]byte(cached), &assignment); err == nil {
			return &assignment, nil
		}
	}
	assignment, err := s.tasks.FindClosestAssignmentForUser(userID)
	if err != nil {
		return nil, err
	}

	// save to cache
	data, err := json.Marshal(assignment)
	if err == nil {
		s.cache.Set(context.Background(), key, data, time.Hour)

	}

	return assignment, nil
}

func (s *TaskService) MarkAssignmentCompleted(assignmentID int) error {
	// delete assignment from cache
	key := utils.GetAssignmentKey(assignmentID)
	if err := s.cache.Del(context.Background(), key); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	user, err := s.tasks.FindUserByAssignmentID(assignmentID)
	if err != nil {
		return err
	}

	// delete user assignments from cache
	userAssignmentsKey := utils.GetAssignmentsForUserKey(user.ID)
	if err := s.cache.Del(context.Background(), userAssignmentsKey); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", userAssignmentsKey, err)
	}

	// delete closest user assignment from cache
	userClosestAssignmentsKey := utils.GetClosestAssignmentsForUserKey(user.ID)
	if err := s.cache.Del(context.Background(), userClosestAssignmentsKey); err != nil {
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
	if err := s.cache.Del(context.Background(), key); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	user, err := s.tasks.FindUserByAssignmentID(assignmentID)
	if err != nil {
		return err
	}

	// delete user assignments from cache
	userAssignmentsKey := utils.GetAssignmentsForUserKey(user.ID)
	if err := s.cache.Del(context.Background(), userAssignmentsKey); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", userAssignmentsKey, err)
	}

	// delete closest user assignment from cache
	userClosestAssignmentsKey := utils.GetClosestAssignmentsForUserKey(user.ID)
	if err := s.cache.Del(context.Background(), userClosestAssignmentsKey); err != nil {
		log.Printf("Failed to delete redis cache for key %s: %v", userClosestAssignmentsKey, err)
	}
	if err := s.tasks.DeleteAssignment(assignmentID); err != nil {
		return err
	}

	return nil
}
