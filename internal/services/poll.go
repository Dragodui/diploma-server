package services

import (
	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/redis/go-redis/v9"
)

type PollService struct {
	repo  repository.PollRepository
	cache *redis.Client
}

type IPollService interface {
	// polls
	Create(homeID int, question, pollType string, options []models.OptionRequest) error
	GetPollByID(pollID int) (*models.Poll, error)
	GetAllPollsByHomeID(homeID int) (*[]models.Poll, error)
	ClosePoll(pollID, homeID int) error
	Delete(pollID, homeID int) error

	// votes
	Vote(userID, optionID, homeID int) error
}

func NewPollService(repo repository.PollRepository, cache *redis.Client) *PollService {
	return &PollService{
		repo, cache,
	}
}

// polls
func (s *PollService) Create(homeID int, question, pollType string, options []models.OptionRequest) error {
	var optionModels []models.Option
	for _, option := range options {
		optionModels = append(optionModels, models.Option{
			Title: option.Title,
		})
	}

	key := utils.GetAllPollsForHomeKey(homeID)
	if err := utils.DeleteFromCache(key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	err := s.repo.Create(&models.Poll{
		HomeID:   homeID,
		Question: question,
		Type:     pollType,
	}, optionModels)

	return err
}

func (s *PollService) GetPollByID(pollID int) (*models.Poll, error) {
	key := utils.GetPollKey(pollID)
	cached, err := utils.GetFromCache[models.Poll](key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	poll, err := s.repo.FindPollByID(pollID)
	if poll != nil && err == nil {
		_ = utils.WriteToCache(key, poll, s.cache)
	}

	return poll, err
}

func (s *PollService) GetAllPollsByHomeID(homeID int) (*[]models.Poll, error) {
	key := utils.GetAllPollsForHomeKey(homeID)
	cached, err := utils.GetFromCache[[]models.Poll](key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	return s.repo.FindAllPollsByHomeID(homeID)
}

func (s *PollService) ClosePoll(pollID, homeID int) error {
	pollsKey := utils.GetPollKey(pollID)
	pollsForHomeKey := utils.GetAllPollsForHomeKey(homeID)

	if err := utils.DeleteFromCache(pollsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", pollsKey, err)
	}

	if err := utils.DeleteFromCache(pollsForHomeKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", pollsForHomeKey, err)
	}

	return s.repo.ClosePoll(pollID)
}

func (s *PollService) Delete(pollID, homeID int) error {
	pollsKey := utils.GetPollKey(pollID)
	pollsForHomeKey := utils.GetAllPollsForHomeKey(homeID)

	if err := utils.DeleteFromCache(pollsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", pollsKey, err)
	}

	if err := utils.DeleteFromCache(pollsForHomeKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", pollsForHomeKey, err)
	}

	return s.repo.Delete(pollID)
}

// votes
func (s *PollService) Vote(userID, optionID, homeID int) error {
	poll, err := s.repo.FindPollByOptionID(optionID)
	if err != nil {
		return err
	}

	// delete from cache
	pollsKey := utils.GetPollKey(poll.ID)
	pollsForHomeKey := utils.GetAllPollsForHomeKey(homeID)

	if err := utils.DeleteFromCache(pollsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", pollsKey, err)
	}

	if err := utils.DeleteFromCache(pollsForHomeKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", pollsForHomeKey, err)
	}

	return s.repo.Vote(&models.Vote{
		UserID:   userID,
		OptionID: optionID,
	})
}
