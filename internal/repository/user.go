package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, u *models.User) error
	FindByID(ctx context.Context, id int) (*models.User, error)
	FindByName(ctx context.Context, name string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	SetVerifyToken(ctx context.Context, email, token string, expiresAt time.Time) error
	VerifyEmail(ctx context.Context, token string) error
	GetByResetToken(ctx context.Context, token string) (*models.User, error)
	UpdatePassword(ctx context.Context, userID int, newHash string) error
	SetResetToken(ctx context.Context, email, token string, expiresAt time.Time) error
	Update(ctx context.Context, user *models.User, updates map[string]interface{}) error
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, u *models.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *userRepo) FindByID(ctx context.Context, id int) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).Where("id=?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &u, err
}

func (r *userRepo) FindByName(ctx context.Context, name string) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).Where("name=?", name).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &u, err
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).Where("email=?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &u, err
}

func (r *userRepo) SetVerifyToken(ctx context.Context, email, token string, expiresAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"verify_token":      token,
			"verify_expires_at": expiresAt,
		}).Error
}

func (r *userRepo) VerifyEmail(ctx context.Context, token string) error {
	res := r.db.WithContext(ctx).Model(&models.User{}).Where("verify_token = ? AND verify_expires_at > ?", token, time.Now()).Updates(map[string]any{
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

func (r *userRepo) GetByResetToken(ctx context.Context, token string) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).Where("reset_token = ? AND reset_expires_at > ?", token, time.Now()).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) SetResetToken(ctx context.Context, email, token string, expiresAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&models.User{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"reset_token":      token,
			"reset_expires_at": expiresAt,
		})

	if result.Error != nil {
		return result.Error
	}

	// We check RowsAffected to know if user exists
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepo) UpdatePassword(ctx context.Context, userID int, newHash string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"password_hash":    newHash,
			"reset_token":      nil,
			"reset_expires_at": nil,
		}).Error
}

func (r *userRepo) Update(ctx context.Context, user *models.User, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(user).Updates(updates).Error
}
