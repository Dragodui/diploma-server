package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type TaskScheduleRepository interface {
	Create(ctx context.Context, schedule *models.TaskSchedule) error
	FindByID(ctx context.Context, id int) (*models.TaskSchedule, error)
	FindByTaskID(ctx context.Context, taskID int) (*models.TaskSchedule, error)
	FindByHomeID(ctx context.Context, homeID int) ([]models.TaskSchedule, error)
	FindDueSchedules(ctx context.Context, now time.Time) ([]models.TaskSchedule, error)
	Update(ctx context.Context, schedule *models.TaskSchedule) error
	Delete(ctx context.Context, id int) error
}

type taskScheduleRepo struct {
	db *gorm.DB
}

func NewTaskScheduleRepository(db *gorm.DB) TaskScheduleRepository {
	return &taskScheduleRepo{db}
}

func (r *taskScheduleRepo) Create(ctx context.Context, schedule *models.TaskSchedule) error {
	return r.db.WithContext(ctx).Create(schedule).Error
}

func (r *taskScheduleRepo) FindByID(ctx context.Context, id int) (*models.TaskSchedule, error) {
	var schedule models.TaskSchedule
	err := r.db.WithContext(ctx).Preload("Task").First(&schedule, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &schedule, err
}

func (r *taskScheduleRepo) FindByTaskID(ctx context.Context, taskID int) (*models.TaskSchedule, error) {
	var schedule models.TaskSchedule
	err := r.db.WithContext(ctx).Preload("Task").Where("task_id = ?", taskID).First(&schedule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &schedule, err
}

func (r *taskScheduleRepo) FindByHomeID(ctx context.Context, homeID int) ([]models.TaskSchedule, error) {
	var schedules []models.TaskSchedule
	err := r.db.WithContext(ctx).
		Preload("Task").
		Joins("JOIN tasks ON task_schedules.task_id = tasks.id").
		Where("tasks.home_id = ? AND task_schedules.is_active = ?", homeID, true).
		Find(&schedules).Error
	return schedules, err
}

func (r *taskScheduleRepo) FindDueSchedules(ctx context.Context, now time.Time) ([]models.TaskSchedule, error) {
	var schedules []models.TaskSchedule
	err := r.db.WithContext(ctx).
		Preload("Task").
		Where("is_active = ? AND next_run_date <= ?", true, now).
		Find(&schedules).Error
	return schedules, err
}

func (r *taskScheduleRepo) Update(ctx context.Context, schedule *models.TaskSchedule) error {
	return r.db.WithContext(ctx).Save(schedule).Error
}

func (r *taskScheduleRepo) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.TaskSchedule{}, id).Error
}
