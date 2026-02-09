package repository

import (
	"context"
	"errors"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type RoomRepository interface {
	Create(ctx context.Context, room *models.Room) error
	FindByID(ctx context.Context, id int) (*models.Room, error)
	Delete(ctx context.Context, id int) error
	FindByHomeID(ctx context.Context, homeID int) (*[]models.Room, error)
}

type roomRepo struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepo{db}
}

func (r *roomRepo) Create(ctx context.Context, room *models.Room) error {
	return r.db.WithContext(ctx).Create(room).Error
}

func (r *roomRepo) FindByID(ctx context.Context, id int) (*models.Room, error) {
	var room models.Room
	if err := r.db.WithContext(ctx).First(&room, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

func (r *roomRepo) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Room{}, id).Error
}

func (r *roomRepo) FindByHomeID(ctx context.Context, homeID int) (*[]models.Room, error) {
	var rooms []models.Room

	if err := r.db.WithContext(ctx).Where("home_id=?", homeID).Find(&rooms).Error; err != nil {
		return nil, err
	}

	return &rooms, nil
}

