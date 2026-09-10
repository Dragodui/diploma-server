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
	Update(ctx context.Context, t *models.Task) error
	Delete(ctx context.Context, id int) error
	ReassignRoom(ctx context.Context, taskID, roomID int) error
	RoomBelongsToHome(ctx context.Context, roomID, homeID int) (bool, error)

	// task assignments
	AssignUser(ctx context.Context, taskID, userID int, date time.Time) error
	FindAssignmentsForUser(ctx context.Context, userID int, homeID int) (*[]models.TaskAssignment, error)
	FindClosestAssignmentForUser(ctx context.Context, userID int) (*models.TaskAssignment, error)
	FindClosestAssignmentForUserInHome(ctx context.Context, userID, homeID int) (*models.TaskAssignment, error)
	FindClosestUnassignedTaskInHome(ctx context.Context, homeID int) (*models.Task, error)
	FindAssignmentsNeedingReminder(ctx context.Context, windowStart, windowEnd time.Time) ([]models.TaskAssignment, error)
	FindAssignmentByTaskAndUser(ctx context.Context, taskID, userID int) (*models.TaskAssignment, error)
	FindAssignmentByID(ctx context.Context, assignmentID int) (*models.TaskAssignment, error)
	MarkCompleted(ctx context.Context, assignmentID int) error
	MarkUncompleted(ctx context.Context, assignmentID int) error
	MarkReminderSent(ctx context.Context, assignmentID int, sentAt time.Time) error
	ResetReminderStateForTask(ctx context.Context, taskID int) error
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
	err := r.db.WithContext(ctx).Preload("Room").Preload("Schedule").Preload("TaskAssignments").First(&task, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &task, err
}

func (r *taskRepo) FindByHomeID(ctx context.Context, homeID int) (*[]models.Task, error) {
	var tasks []models.Task
	if err := r.db.WithContext(ctx).Preload("Room").Preload("Schedule").Preload("TaskAssignments").Preload("TaskAssignments.User").Where("home_id=?", homeID).Find(&tasks).Error; err != nil {
		return nil, err
	}

	return &tasks, nil
}

func (r *taskRepo) Delete(ctx context.Context, id int) error {
	// Delete associated schedule first
	if err := r.db.WithContext(ctx).Where("task_id = ?", id).Delete(&models.TaskSchedule{}).Error; err != nil {
		return err
	}

	// Delete associated task assignments
	if err := r.db.WithContext(ctx).Where("task_id = ?", id).Delete(&models.TaskAssignment{}).Error; err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Delete(&models.Task{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (r *taskRepo) Update(ctx context.Context, t *models.Task) error {
	return r.db.WithContext(ctx).Save(t).Error
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

func (r *taskRepo) FindAssignmentsForUser(ctx context.Context, userID int, homeID int) (*[]models.TaskAssignment, error) {
	var assignments []models.TaskAssignment

	if err := r.db.WithContext(ctx).
		Joins("JOIN tasks ON task_assignments.task_id = tasks.id").
		Where("task_assignments.user_id = ? AND tasks.home_id = ?", userID, homeID).
		Find(&assignments).Error; err != nil {
		return nil, err
	}

	return &assignments, nil
}

func (r *taskRepo) FindClosestAssignmentForUser(ctx context.Context, userID int) (*models.TaskAssignment, error) {
	var assignment models.TaskAssignment

	if err := r.db.WithContext(ctx).
		Preload("Task").
		Joins("JOIN tasks ON task_assignments.task_id = tasks.id").
		Where("task_assignments.user_id = ? AND task_assignments.status != 'completed'", userID).
		Order("tasks.due_date IS NULL, tasks.due_date ASC").
		First(&assignment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &assignment, nil
}

func (r *taskRepo) FindClosestAssignmentForUserInHome(ctx context.Context, userID, homeID int) (*models.TaskAssignment, error) {
	var assignment models.TaskAssignment

	// "Closest" means nearest upcoming due date, not oldest assignment -
	// otherwise a freshly created task with an earlier due date than
	// whatever was assigned first would never surface as next-up.
	// Tasks without a due date sort last instead of first.
	if err := r.db.WithContext(ctx).
		Preload("Task").
		Joins("JOIN tasks ON task_assignments.task_id = tasks.id").
		Where("task_assignments.user_id = ? AND task_assignments.status != 'completed' AND tasks.home_id = ?", userID, homeID).
		Order("tasks.due_date IS NULL, tasks.due_date ASC").
		First(&assignment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &assignment, nil
}

// FindClosestUnassignedTaskInHome finds the nearest-due task in the home that
// has nobody assigned yet - a shared/"anyone can do it" task - for when the
// caller has no personal assignment to show on their dashboard.
func (r *taskRepo) FindClosestUnassignedTaskInHome(ctx context.Context, homeID int) (*models.Task, error) {
	var task models.Task

	if err := r.db.WithContext(ctx).
		Where("home_id = ? AND NOT EXISTS (SELECT 1 FROM task_assignments WHERE task_assignments.task_id = tasks.id)", homeID).
		Order("due_date IS NULL, due_date ASC").
		First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &task, nil
}

func (r *taskRepo) FindAssignmentsNeedingReminder(ctx context.Context, windowStart, windowEnd time.Time) ([]models.TaskAssignment, error) {
	var assignments []models.TaskAssignment

	err := r.db.WithContext(ctx).
		Model(&models.TaskAssignment{}).
		Preload("Task").
		Where("task_assignments.status != ?", "completed").
		Where("task_assignments.reminder_sent_at IS NULL").
		Joins("JOIN tasks ON task_assignments.task_id = tasks.id").
		Where("tasks.due_date IS NOT NULL").
		Where("(tasks.due_date - (tasks.reminder_minutes * interval '1 minute')) > ?", windowStart).
		Where("(tasks.due_date - (tasks.reminder_minutes * interval '1 minute')) <= ?", windowEnd).
		Find(&assignments).Error
	if err != nil {
		return nil, err
	}

	return assignments, nil
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

func (r *taskRepo) MarkReminderSent(ctx context.Context, assignmentID int, sentAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.TaskAssignment{}).
		Where("id = ?", assignmentID).
		Update("reminder_sent_at", sentAt).Error
}

func (r *taskRepo) ResetReminderStateForTask(ctx context.Context, taskID int) error {
	return r.db.WithContext(ctx).
		Model(&models.TaskAssignment{}).
		Where("task_id = ? AND status != ?", taskID, "completed").
		Update("reminder_sent_at", nil).Error
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

func (r *taskRepo) RoomBelongsToHome(ctx context.Context, roomID, homeID int) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Room{}).
		Where("id = ? AND home_id = ?", roomID, homeID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
