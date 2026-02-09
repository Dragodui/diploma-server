package services

import (
	"context"

	"github.com/Dragodui/diploma-server/internal/event"
	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/redis/go-redis/v9"
)

type RoomService struct {
	repo repository.RoomRepository
	cache *redis.Client
}

type IRoomService interface {
	CreateRoom(name string, homeID int) error
	GetRoomByID(roomID int) (*models.Room, error)
	GetRoomsByHomeID(homeID int) (*[]models.Room, error)
	DeleteRoom(roomID int) error
}

func NewRoomService(repo repository.RoomRepository, cache *redis.Client) *RoomService {
	return &RoomService{repo: repo, cache: cache}
}

func (s *RoomService) CreateRoom(name string, homeID int) error {
	// delete homes rooms from cache
	key := utils.GetRoomsForHomeKey(homeID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	room := &models.Room{
		Name:   name,
		HomeID: homeID,
	}
	if err := s.repo.Create(room); err != nil {
		return err
	}

	event.SendEvent(context.Background(), s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleRoom,
		Action: event.ActionCreated,
		Data:   room,
	})

	return nil
}

func (s *RoomService) GetRoomByID(roomID int) (*models.Room, error) {
	key := utils.GetRoomKey(roomID)

	// try to get from cache
	cached, err := utils.GetFromCache[models.Room](key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	room, err := s.repo.FindByID(roomID)
	if err != nil {
		return nil, err
	}

	if err := utils.WriteToCache(key, room, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return room, nil
}

func (s *RoomService) GetRoomsByHomeID(homeID int) (*[]models.Room, error) {
	key := utils.GetRoomsForHomeKey(homeID)
	cached, err := utils.GetFromCache[[]models.Room](key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	rooms, err := s.repo.FindByHomeID(homeID)
	if err != nil {
		return nil, err
	}

	if err := utils.WriteToCache(key, rooms, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return rooms, nil
}

func (s *RoomService) DeleteRoom(roomID int) error {
	// delete from cache
	roomKey := utils.GetRoomKey(roomID)
	room, err := s.repo.FindByID(roomID)
	if err != nil {
		return err
	}
	homeID := room.HomeID
	roomsKey := utils.GetRoomsForHomeKey(homeID)
	if err := utils.DeleteFromCache(roomKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", roomKey, err)
	}
	if err := utils.DeleteFromCache(roomsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", roomsKey, err)
	}

	if err := s.repo.Delete(roomID); err != nil {
		return err
	}

	event.SendEvent(context.Background(), s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleRoom,
		Action: event.ActionDeleted,
		Data:   room,
	})

	return nil
}
