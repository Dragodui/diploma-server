package services

import (
	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/redis/go-redis/v9"
)

type RoomService struct {
	rooms repository.RoomRepository
	cache *redis.Client
}

func NewRoomService(repo repository.RoomRepository, cache *redis.Client) *RoomService {
	return &RoomService{rooms: repo, cache: cache}
}

func (s *RoomService) CreateRoom(name string, homeID int) error {
	// delete homes rooms from cache
	roomsKey := utils.GetRoomsForHomeKey(homeID)
	if err := utils.DeleteFromCache(roomsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", roomsKey, err)
	}

	if err := s.rooms.Create(&models.Room{
		Name:   name,
		HomeID: homeID,
	}); err != nil {
		return err
	}
	return nil
}

func (s *RoomService) GetRoomByID(roomID int) (*models.Room, error) {
	key := utils.GetRoomKey(roomID)
	// try to get from cache

	cached, err := utils.GetFromCache[models.Room](key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	room, err := s.rooms.FindByID(roomID)
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

	rooms, err := s.rooms.FindByHomeID(homeID)
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
	room, err := s.rooms.FindByID(roomID)
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

	if err := s.rooms.Delete(roomID); err != nil {
		return err
	}

	return nil
}
