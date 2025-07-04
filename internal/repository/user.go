package repository

import (
	"errors"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(u *models.User) error
	FindByID(id int) (*models.User, error)
	FindByName(name string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
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
