package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(ctx context.Context, t *models.Task) error
	FindByID(ctx context.Context, id int) (*models.Task, error)
	FindByHomeID(ctx context.Context, homeID int) (*[]models.Task, error)
	Delete(ctx context.Context, id int) error
	ReassignRoom(ctx context.Context, taskID, roomID int) error

	// task assignments
	AssignUser(ctx context.Context, taskID, userID int, date time.Time) error
	FindAssignmentsForUser(ctx context.Context, userID int) (*[]models.TaskAssignment, error)
	FindClosestAssignmentForUser(ctx context.Context, userID int) (*models.TaskAssignment, error)
	FindAssignmentByTaskAndUser(ctx context.Context, taskID, userID int) (*models.TaskAssignment, error)
	FindAssignmentByID(ctx context.Context, assignmentID int) (*models.TaskAssignment, error)
	MarkCompleted(ctx context.Context, assignmentID int) error
	MarkUncompleted(ctx context.Context, assignmentID int) error
	FindUserByAssignmentID(ctx context.Context, assignmentID int) (*models.User, error)
	DeleteAssignment(ctx context.Context, assignmentID int) error
}

type taskRepo struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepo{db}
}

func (r *taskRepo) Create(ctx context.Context, t *models.Task) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *taskRepo) FindByID(ctx context.Context, id int) (*models.Task, error) {
	var task models.Task
	// we need preload to room field was not empty
	err := r.db.WithContext(ctx).Preload("Room").First(&task, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &task, err
}

func (r *taskRepo) FindByHomeID(ctx context.Context, homeID int) (*[]models.Task, error) {
	var tasks []models.Task
	if err := r.db.WithContext(ctx).Preload("Room").Preload("TaskAssignments").Preload("TaskAssignments.User").Where("home_id=?", homeID).Find(&tasks).Error; err != nil {
		return nil, err
	}

	return &tasks, nil
}

func (r *taskRepo) Delete(ctx context.Context, id int) error {
	// Delete associated task assignments first
	if err := r.db.WithContext(ctx).Where("task_id = ?", id).Delete(&models.TaskAssignment{}).Error; err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Delete(&models.Task{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (r *taskRepo) AssignUser(ctx context.Context, taskID, userID int, date time.Time) error {
	var task models.Task
	if err := r.db.WithContext(ctx).First(&task, taskID).Error; err != nil {
		return err
	}
	newTaskAssignment := models.TaskAssignment{
		TaskID:       taskID,
		UserID:       userID,
		Status:       "assigned",
		AssignedDate: date,
	}
	if err := r.db.WithContext(ctx).Create(&newTaskAssignment).Error; err != nil {
		return err
	}

	return nil
}

func (r *taskRepo) FindAssignmentsForUser(ctx context.Context, userID int) (*[]models.TaskAssignment, error) {
	var assignments []models.TaskAssignment

	if err := r.db.WithContext(ctx).Where("user_id=?", userID).Find(&assignments).Error; err != nil {
		return nil, err
	}

	return &assignments, nil
}

func (r *taskRepo) FindClosestAssignmentForUser(ctx context.Context, userID int) (*models.TaskAssignment, error) {
	var assignment models.TaskAssignment

	if err := r.db.WithContext(ctx).Preload("Task").Where("user_id=? AND status != 'completed'", userID).Order("assigned_date asc").First(&assignment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &assignment, nil
}

func (r *taskRepo) MarkUncompleted(ctx context.Context, assignmentID int) error {
	var assignment models.TaskAssignment
	if err := r.db.WithContext(ctx).First(&assignment, assignmentID).Error; err != nil {
		return err
	}

	assignment.Status = "assigned"
	assignment.CompleteDate = nil

	if err := r.db.WithContext(ctx).Save(&assignment).Error; err != nil {
		return err
	}

	return nil
}

func (r *taskRepo) FindAssignmentByTaskAndUser(ctx context.Context, taskID, userID int) (*models.TaskAssignment, error) {
	var assignment models.TaskAssignment

	if err := r.db.WithContext(ctx).Where("task_id = ? AND user_id = ?", taskID, userID).First(&assignment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &assignment, nil
}

func (r *taskRepo) FindAssignmentByID(ctx context.Context, assignmentID int) (*models.TaskAssignment, error) {
	var assignment models.TaskAssignment
	if err := r.db.WithContext(ctx).Preload("Task").First(&assignment, assignmentID).Error; err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *taskRepo) FindUserByAssignmentID(ctx context.Context, assignmentID int) (*models.User, error) {
	var assignment models.TaskAssignment
	if err := r.db.WithContext(ctx).First(&assignment, assignmentID).Error; err != nil {
		return nil, err
	}
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, assignment.UserID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *taskRepo) MarkCompleted(ctx context.Context, assignmentID int) error {
	var assignment models.TaskAssignment
	if err := r.db.WithContext(ctx).First(&assignment, assignmentID).Error; err != nil {
		return err
	}

	now := time.Now()
	assignment.Status = "completed"
	assignment.CompleteDate = &now

	if err := r.db.WithContext(ctx).Save(&assignment).Error; err != nil {
		return err
	}

	return nil
}

func (r *taskRepo) DeleteAssignment(ctx context.Context, assignmentID int) error {
	if err := r.db.WithContext(ctx).Delete(&models.TaskAssignment{}, assignmentID).Error; err != nil {
		return err
	}

	return nil
}

func (r *taskRepo) ReassignRoom(ctx context.Context, taskID, roomID int) error {
	var task models.Task

	if err := r.db.WithContext(ctx).First(&task, taskID).Error; err != nil {
		return err
	}

	task.RoomID = &roomID
	if err := r.db.WithContext(ctx).Save(&task).Error; err != nil {
		return err
	}

	return nil
}

