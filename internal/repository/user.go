package repository

import (
	"errors"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(u *models.User) error
	FindByID(id int) (*models.User, error)
	FindByName(name string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	SetVerifyToken(email, token string, expiresAt time.Time) error
	VerifyEmail(token string) error
	GetByResetToken(token string) (*models.User, error)
	UpdatePassword(userID int, newHash string) error
	SetResetToken(email, token string, expiresAt time.Time) error
	Update(user *models.User, updates map[string]interface{}) error
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(u *models.User) error {
	return r.db.Create(u).Error
}

func (r *userRepo) FindByID(id int) (*models.User, error) {
	var u models.User
	err := r.db.Where("id=?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &u, err
}

func (r *userRepo) FindByName(name string) (*models.User, error) {
	var u models.User
	err := r.db.Where("name=?", name).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &u, err
}

func (r *userRepo) FindByEmail(email string) (*models.User, error) {
	var u models.User
	err := r.db.Where("email=?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &u, err
}

func (r *userRepo) SetVerifyToken(email, token string, expiresAt time.Time) error {
	return r.db.Model(&models.User{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"verify_token":      token,
			"verify_expires_at": expiresAt,
		}).Error
}

func (r *userRepo) VerifyEmail(token string) error {
	res := r.db.Model(&models.User{}).Where("verify_token = ? AND verify_expires_at > ?", token, time.Now()).Updates(map[string]interface{}{
		"email_verified":    true,
		"verify_token":      nil,
		"verify_expires_at": nil,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("not found")
	}
	return nil
}

func (r *userRepo) GetByResetToken(token string) (*models.User, error) {
	var u models.User
	err := r.db.Where("reset_token = ? AND reset_expires_at > ?", token, time.Now()).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) SetResetToken(email, token string, expiresAt time.Time) error {
	return r.db.Model(&models.User{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"reset_token":      token,
			"reset_expires_at": expiresAt,
		}).Error
}

func (r *userRepo) UpdatePassword(userID int, newHash string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"password_hash":    newHash,
			"reset_token":      nil,
			"reset_expires_at": nil,
		}).Error
}

func (r *userRepo) Update(user *models.User, updates map[string]interface{}) error {
	return r.db.Model(user).Updates(updates).Error
}
