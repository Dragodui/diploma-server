package repository

import (
	"errors"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type RoomRepository interface {
	Create(r *models.Room) error
	FindByID(id int) (*models.Room, error)
	Delete(id int) error
	FindByHomeID(homeID int) (*[]models.Room, error)
}

type roomRepo struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepo{db}
}

func (r *roomRepo) Create(room *models.Room) error {
	return r.db.Create(room).Error
}

func (r *roomRepo) FindByID(id int) (*models.Room, error) {
	var room models.Room
	if err := r.db.First(&room, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

func (r *roomRepo) Delete(id int) error {
	return r.db.Delete(&models.Room{}, id).Error
}

func (r *roomRepo) FindByHomeID(homeID int) (*[]models.Room, error) {
	var rooms []models.Room

	if err := r.db.Where("home_id=?", homeID).Find(&rooms).Error; err != nil {
		return nil, err
	}

	return &rooms, nil
}
