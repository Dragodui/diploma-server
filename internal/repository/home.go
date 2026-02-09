package repository

import (
	"context"
	"errors"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/utils"
	"gorm.io/gorm"
)

type HomeRepository interface {
	// home
	Create(ctx context.Context, h *models.Home) error
	RegenerateCode(ctx context.Context, code string, id int) error
	FindByID(ctx context.Context, id int) (*models.Home, error)
	FindByInviteCode(ctx context.Context, inviteCode string) (*models.Home, error)
	Delete(ctx context.Context, id int) error
	IsAdmin(ctx context.Context, id int, userID int) (bool, error)

	// home memberships
	AddMember(ctx context.Context, id int, userID int, role string) error
	IsMember(ctx context.Context, id int, userID int) (bool, error)
	DeleteMember(ctx context.Context, id int, userID int) error
	GenerateUniqueInviteCode(ctx context.Context) (string, error)
	GetUserHome(ctx context.Context, userID int) (*models.Home, error)
}

type homeRepo struct {
	db *gorm.DB
}

func NewHomeRepository(db *gorm.DB) HomeRepository {
	return &homeRepo{db}
}

func (r *homeRepo) Create(ctx context.Context, h *models.Home) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *homeRepo) RegenerateCode(ctx context.Context, code string, id int) error {
	var home models.Home
	if err := r.db.WithContext(ctx).First(&home, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	home.InviteCode = code
	return r.db.WithContext(ctx).Save(&home).Error

}

func (r *homeRepo) FindByID(ctx context.Context, id int) (*models.Home, error) {
	var home models.Home

	// taking memberships also
	if err := r.db.WithContext(ctx).Preload("Memberships").Preload("Memberships.User").First(&home, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &home, nil
}

func (r *homeRepo) FindByInviteCode(ctx context.Context, inviteCode string) (*models.Home, error) {
	var home models.Home

	// taking memberships also
	if err := r.db.WithContext(ctx).Preload("Memberships").Preload("Memberships.User").Where("invite_code = ?", inviteCode).First(&home).Error; err != nil {
		return nil, err
	}

	return &home, nil
}

func (r *homeRepo) Delete(ctx context.Context, id int) error {
	// 1. Delete HomeMemberships
	if err := r.db.WithContext(ctx).Where("home_id = ?", id).Delete(&models.HomeMembership{}).Error; err != nil {
		return err
	}

	// 2. Delete HomeNotifications
	if err := r.db.WithContext(ctx).Where("home_id = ?", id).Delete(&models.HomeNotification{}).Error; err != nil {
		return err
	}

	// 3. Delete Bills
	if err := r.db.WithContext(ctx).Where("home_id = ?", id).Delete(&models.Bill{}).Error; err != nil {
		return err
	}

	// 4. Delete Tasks (and assignments)
	var tasks []models.Task
	if err := r.db.WithContext(ctx).Where("home_id = ?", id).Find(&tasks).Error; err != nil {
		return err
	}
	for _, task := range tasks {
		if err := r.db.WithContext(ctx).Where("task_id = ?", task.ID).Delete(&models.TaskAssignment{}).Error; err != nil {
			return err
		}
	}
	if err := r.db.WithContext(ctx).Where("home_id = ?", id).Delete(&models.Task{}).Error; err != nil {
		return err
	}

	// 5. Delete ShoppingCategories (and items)
	var categories []models.ShoppingCategory
	if err := r.db.WithContext(ctx).Where("home_id = ?", id).Find(&categories).Error; err != nil {
		return err
	}
	for _, cat := range categories {
		if err := r.db.WithContext(ctx).Where("category_id = ?", cat.ID).Delete(&models.ShoppingItem{}).Error; err != nil {
			return err
		}
	}
	if err := r.db.WithContext(ctx).Where("home_id = ?", id).Delete(&models.ShoppingCategory{}).Error; err != nil {
		return err
	}

	// 6. Delete Polls (and options/votes)
	var polls []models.Poll
	if err := r.db.WithContext(ctx).Where("home_id = ?", id).Find(&polls).Error; err != nil {
		return err
	}
	for _, poll := range polls {
		var options []models.Option
		if err := r.db.WithContext(ctx).Where("poll_id = ?", poll.ID).Find(&options).Error; err != nil {
			return err
		}
		for _, opt := range options {
			if err := r.db.WithContext(ctx).Where("option_id = ?", opt.ID).Delete(&models.Vote{}).Error; err != nil {
				return err
			}
		}
		if err := r.db.WithContext(ctx).Where("poll_id = ?", poll.ID).Delete(&models.Option{}).Error; err != nil {
			return err
		}
	}
	if err := r.db.WithContext(ctx).Where("home_id = ?", id).Delete(&models.Poll{}).Error; err != nil {
		return err
	}

	// 7. Delete Rooms
	if err := r.db.WithContext(ctx).Where("home_id = ?", id).Delete(&models.Room{}).Error; err != nil {
		return err
	}

	// 8. Delete Home
	return r.db.WithContext(ctx).Delete(&models.Home{}, id).Error
}

func (r *homeRepo) AddMember(ctx context.Context, id int, userID int, role string) error {

	if err := r.db.WithContext(ctx).Create(&models.HomeMembership{
		HomeID: id,
		UserID: userID,
		Role:   role,
	}).Error; err != nil {
		return err
	}

	return nil
}

func (r *homeRepo) IsMember(ctx context.Context, id int, userID int) (bool, error) {

	var count int64
	if err := r.db.WithContext(ctx).Model(&models.HomeMembership{}).Where("home_id = ? AND user_id = ?", id, userID).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *homeRepo) DeleteMember(ctx context.Context, id int, userID int) error {

	if err := r.db.WithContext(ctx).Where("home_id = ? AND user_id = ?", id, userID).Delete(&models.HomeMembership{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *homeRepo) GenerateUniqueInviteCode(ctx context.Context) (string, error) {
	for {
		code := utils.RandString(8)

		var count int64
		if err := r.db.WithContext(ctx).Model(&models.Home{}).
			Where("invite_code = ?", code).
			Count(&count).Error; err != nil {
			return "", err
		}

		if count == 0 {
			return code, nil
		}
	}
}

func (r *homeRepo) IsAdmin(ctx context.Context, id int, userID int) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.HomeMembership{}).Where("home_id = ? AND user_id = ? AND role='admin'", id, userID).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *homeRepo) GetUserHome(ctx context.Context, userID int) (*models.Home, error) {
	var home models.Home

	if err := r.db.WithContext(ctx).Model(&models.Home{}).Joins("JOIN home_memberships ON home_memberships.home_id = homes.id").Where("home_memberships.user_id = ?", userID).Preload("Memberships").Preload("Memberships.User").First(&home).Error; err != nil {
		return nil, err
	}

	return &home, nil
}

