package repository

import (
	"errors"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(t *models.Task) error
	FindByID(id int) (*models.Task, error)
	FindByHomeID(homeID int) (*[]models.Task, error)
	Delete(id int) error
	ReassignRoom(taskID, roomID int) error

	// task assignments
	AssignUser(taskID, userID int, date time.Time) error
	FindAssignmentsForUser(userID int) (*[]models.TaskAssignment, error)
	FindClosestAssignmentForUser(userID int) (*models.TaskAssignment, error)
	FindAssignmentByTaskAndUser(taskID, userID int) (*models.TaskAssignment, error)
	MarkCompleted(assignmentID int) error
	FindUserByAssignmentID(assignmentID int) (*models.User, error)
	DeleteAssignment(assignmentID int) error
}

type taskRepo struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepo{db}
}

func (r *taskRepo) Create(t *models.Task) error {
	return r.db.Create(t).Error
}

func (r *taskRepo) FindByID(id int) (*models.Task, error) {
	var task models.Task
	// we need preload to room field was not empty
	err := r.db.Preload("Room").First(&task, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &task, err
}

func (r *taskRepo) FindByHomeID(homeID int) (*[]models.Task, error) {
	var tasks []models.Task
	if err := r.db.Preload("Room").Preload("TaskAssignments").Preload("TaskAssignments.User").Where("home_id=?", homeID).Find(&tasks).Error; err != nil {
		return nil, err
	}

	return &tasks, nil
}

func (r *taskRepo) Delete(id int) error {
	if err := r.db.Delete(&models.Task{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (r *taskRepo) AssignUser(taskID, userID int, date time.Time) error {
	var task models.Task
	if err := r.db.First(&task, taskID).Error; err != nil {
		return err
	}
	newTaskAssignment := models.TaskAssignment{
		TaskID:       taskID,
		UserID:       userID,
		Status:       "assigned",
		AssignedDate: date,
	}
	if err := r.db.Create(&newTaskAssignment).Error; err != nil {
		return err
	}

	return nil
}

func (r *taskRepo) FindAssignmentsForUser(userID int) (*[]models.TaskAssignment, error) {
	var assignments []models.TaskAssignment

	if err := r.db.Where("user_id=?", userID).Find(&assignments).Error; err != nil {
		return nil, err
	}

	return &assignments, nil
}

func (r *taskRepo) FindClosestAssignmentForUser(userID int) (*models.TaskAssignment, error) {
	var assignment models.TaskAssignment

	if err := r.db.Where("user_id=?", userID).Order("assigned_date desc").First(&assignment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &assignment, nil
}

func (r *taskRepo) FindAssignmentByTaskAndUser(taskID, userID int) (*models.TaskAssignment, error) {
	var assignment models.TaskAssignment

	if err := r.db.Where("task_id = ? AND user_id = ?", taskID, userID).First(&assignment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &assignment, nil
}

func (r *taskRepo) FindUserByAssignmentID(assignmentID int) (*models.User, error) {
	var assignment models.TaskAssignment
	if err := r.db.First(&assignment, assignmentID).Error; err != nil {
		return nil, err
	}
	var user models.User
	if err := r.db.First(&user, assignment.UserID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *taskRepo) MarkCompleted(assignmentID int) error {
	var assignment models.TaskAssignment
	if err := r.db.First(&assignment, assignmentID).Error; err != nil {
		return err
	}

	now := time.Now()
	assignment.Status = "completed"
	assignment.CompleteDate = &now

	if err := r.db.Save(&assignment).Error; err != nil {
		return err
	}

	return nil
}

func (r *taskRepo) DeleteAssignment(assignmentID int) error {
	if err := r.db.Delete(&models.TaskAssignment{}, assignmentID).Error; err != nil {
		return err
	}

	return nil
}

func (r *taskRepo) ReassignRoom(taskID, roomID int) error {
	var task models.Task

	if err := r.db.First(&task, taskID).Error; err != nil {
		return err
	}

	task.RoomID = &roomID
	if err := r.db.Save(&task).Error; err != nil {
		return err
	}

	return nil
}
