package services

import (
	"context"
	"errors"
	"time"

	"github.com/Dragodui/diploma-server/internal/event"
	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/redis/go-redis/v9"
)

var ErrRevoteNotAllowed = errors.New("revoting is not allowed for this poll")

type PollService struct {
	repo  repository.PollRepository
	cache *redis.Client
}

type IPollService interface {
	// polls
	Create(homeID int, question, pollType string, options []models.OptionRequest, allowRevote bool, endsAt *time.Time) error
	GetPollByID(pollID int) (*models.Poll, error)
	GetAllPollsByHomeID(homeID int) (*[]models.Poll, error)
	ClosePoll(pollID, homeID int) error
	Delete(pollID, homeID int) error

	// votes
	Vote(userID, optionID, homeID int) error
	Unvote(userID, pollID, homeID int) error
}

func NewPollService(repo repository.PollRepository, cache *redis.Client) *PollService {
	return &PollService{
		repo, cache,
	}
}

// polls
func (s *PollService) Create(homeID int, question, pollType string, options []models.OptionRequest, allowRevote bool, endsAt *time.Time) error {
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

	poll := &models.Poll{
		HomeID:      homeID,
		Question:    question,
		Type:        pollType,
		AllowRevote: allowRevote,
		EndsAt:      endsAt,
	}

	if err := s.repo.Create(poll, optionModels); err != nil {
		return err
	}

	event.SendEvent(context.Background(), s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModulePoll,
		Action: event.ActionCreated,
		Data:   poll,
	})

	return nil
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

	if err := s.repo.ClosePoll(pollID); err != nil {
		return err
	}

	event.SendEvent(context.Background(), s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModulePoll,
		Action: event.ActionClosed,
		Data:   map[string]int{"id": pollID},
	})

	return nil
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

	if err := s.repo.Delete(pollID); err != nil {
		return err
	}

	event.SendEvent(context.Background(), s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModulePoll,
		Action: event.ActionDeleted,
		Data:   map[string]int{"id": pollID},
	})

	return nil
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

	vote := &models.Vote{
		UserID:   userID,
		OptionID: optionID,
	}
	if err := s.repo.Vote(vote); err != nil {
		return err
	}

	event.SendEvent(context.Background(), s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModulePoll,
		Action: event.ActionVoted,
		Data:   vote,
	})

	return nil
}

func (s *PollService) Unvote(userID, pollID, homeID int) error {
	poll, err := s.repo.FindPollByID(pollID)
	if err != nil {
		return err
	}

	if !poll.AllowRevote {
		return ErrRevoteNotAllowed
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

	if err := s.repo.Unvote(userID, pollID); err != nil {
		return err
	}

	event.SendEvent(context.Background(), s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModulePoll,
		Action: event.ActionUnvoted,
		Data:   map[string]int{"userID": userID, "pollID": pollID},
	})

	return nil
}
