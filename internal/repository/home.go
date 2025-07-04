package repository

import (
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/utils"
	"gorm.io/gorm"
)

type HomeRepository interface {
	// home
	Create(h *models.Home) error
	FindByID(id int) (*models.Home, error)
	FindByInviteCode(inviteCode string) (*models.Home, error)
	Delete(id int) error
	IsAdmin(id int, userID int) (bool, error)

	// home memberships
	AddMember(id int, userID int, role string) error
	IsMember(id int, userID int) (bool, error)
	DeleteMember(id int, userID int) error
	GenerateUniqueInviteCode() (string, error)
}

type homeRepo struct {
	db *gorm.DB
}

func NewHomeRepository(db *gorm.DB) HomeRepository {
	return &homeRepo{db}
}

func (r *homeRepo) Create(h *models.Home) error {
	return r.db.Create(h).Error
}

func (r *homeRepo) FindByID(id int) (*models.Home, error) {
	var home models.Home

	// taking memberships also
	if err := r.db.Preload("Memberships").First(&home, id).Error; err != nil {
		return nil, err
	}

	return &home, nil
}

func (r *homeRepo) FindByInviteCode(inviteCode string) (*models.Home, error) {
	var home models.Home

	// taking memberships also
	if err := r.db.Preload("Memberships").Where("invite_code = ?", inviteCode).First(&home).Error; err != nil {
		return nil, err
	}

	return &home, nil
}

func (r *homeRepo) Delete(id int) error {

	if err := r.db.Delete(&models.Home{}, id).Error; err != nil {
		return err
	}

	return nil
}

func (r *homeRepo) AddMember(id int, userID int, role string) error {

	if err := r.db.Create(&models.HomeMembership{
		HomeID: id,
		UserID: userID,
		Role:   role,
	}).Error; err != nil {
		return err
	}

	return nil
}

func (r *homeRepo) IsMember(id int, userID int) (bool, error) {

	var count int64
	if err := r.db.Model(&models.HomeMembership{}).Where("home_id = ? AND user_id = ?", id, userID).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *homeRepo) DeleteMember(id int, userID int) error {

	if err := r.db.Where("home_id = ? AND user_id = ?", id, userID).Delete(&models.HomeMembership{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *homeRepo) GenerateUniqueInviteCode() (string, error) {
	for {
		code := utils.RandString(8)

		var count int64
		if err := r.db.Model(&models.Home{}).
			Where("invite_code = ?", code).
			Count(&count).Error; err != nil {
			return "", err
		}

		if count == 0 {
			return code, nil
		}
	}
}

func (r *homeRepo) IsAdmin(id int, userID int) (bool, error) {
	var count int64
	if err := r.db.Model(&models.HomeMembership{}).Where("home_id = ? AND user_id = ? AND role='admin'", id, userID).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
