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
	Create(ctx context.Context, homeID int, question, pollType string, options []models.OptionRequest, allowRevote bool, endsAt *time.Time) error
	GetPollByID(ctx context.Context, pollID int) (*models.Poll, error)
	GetAllPollsByHomeID(ctx context.Context, homeID int) (*[]models.Poll, error)
	ClosePoll(ctx context.Context, pollID, homeID int) error
	Delete(ctx context.Context, pollID, homeID int) error

	// votes
	Vote(ctx context.Context, userID, optionID, homeID int) error
	Unvote(ctx context.Context, userID, pollID, homeID int) error
}

func NewPollService(repo repository.PollRepository, cache *redis.Client) *PollService {
	return &PollService{
		repo, cache,
	}
}

// polls
func (s *PollService) Create(ctx context.Context, homeID int, question, pollType string, options []models.OptionRequest, allowRevote bool, endsAt *time.Time) error {
	var optionModels []models.Option
	for _, option := range options {
		optionModels = append(optionModels, models.Option{
			Title: option.Title,
		})
	}

	key := utils.GetAllPollsForHomeKey(homeID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	poll := &models.Poll{
		HomeID:      homeID,
		Question:    question,
		Type:        pollType,
		AllowRevote: allowRevote,
		EndsAt:      endsAt,
	}

	if err := s.repo.Create(ctx, poll, optionModels); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModulePoll,
		Action: event.ActionCreated,
		Data:   poll,
	})

	return nil
}

func (s *PollService) GetPollByID(ctx context.Context, pollID int) (*models.Poll, error) {
	key := utils.GetPollKey(pollID)
	cached, err := utils.GetFromCache[models.Poll](ctx, key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	poll, err := s.repo.FindPollByID(ctx, pollID)
	if poll != nil && err == nil {
		_ = utils.WriteToCache(ctx, key, poll, s.cache)
	}

	return poll, err
}

func (s *PollService) GetAllPollsByHomeID(ctx context.Context, homeID int) (*[]models.Poll, error) {
	key := utils.GetAllPollsForHomeKey(homeID)
	cached, err := utils.GetFromCache[[]models.Poll](ctx, key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	return s.repo.FindAllPollsByHomeID(ctx, homeID)
}

func (s *PollService) ClosePoll(ctx context.Context, pollID, homeID int) error {
	pollsKey := utils.GetPollKey(pollID)
	pollsForHomeKey := utils.GetAllPollsForHomeKey(homeID)

	if err := utils.DeleteFromCache(ctx, pollsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", pollsKey, err)
	}

	if err := utils.DeleteFromCache(ctx, pollsForHomeKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", pollsForHomeKey, err)
	}

	if err := s.repo.ClosePoll(ctx, pollID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModulePoll,
		Action: event.ActionClosed,
		Data:   map[string]int{"id": pollID},
	})

	return nil
}

func (s *PollService) Delete(ctx context.Context, pollID, homeID int) error {
	pollsKey := utils.GetPollKey(pollID)
	pollsForHomeKey := utils.GetAllPollsForHomeKey(homeID)

	if err := utils.DeleteFromCache(ctx, pollsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", pollsKey, err)
	}

	if err := utils.DeleteFromCache(ctx, pollsForHomeKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", pollsForHomeKey, err)
	}

	if err := s.repo.Delete(ctx, pollID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModulePoll,
		Action: event.ActionDeleted,
		Data:   map[string]int{"id": pollID},
	})

	return nil
}

// votes
func (s *PollService) Vote(ctx context.Context, userID, optionID, homeID int) error {
	poll, err := s.repo.FindPollByOptionID(ctx, optionID)
	if err != nil {
		return err
	}

	// delete from cache
	pollsKey := utils.GetPollKey(poll.ID)
	pollsForHomeKey := utils.GetAllPollsForHomeKey(homeID)

	if err := utils.DeleteFromCache(ctx, pollsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", pollsKey, err)
	}

	if err := utils.DeleteFromCache(ctx, pollsForHomeKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", pollsForHomeKey, err)
	}

	vote := &models.Vote{
		UserID:   userID,
		OptionID: optionID,
	}
	if err := s.repo.Vote(ctx, vote); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModulePoll,
		Action: event.ActionVoted,
		Data:   vote,
	})

	return nil
}

func (s *PollService) Unvote(ctx context.Context, userID, pollID, homeID int) error {
	poll, err := s.repo.FindPollByID(ctx, pollID)
	if err != nil {
		return err
	}

	if !poll.AllowRevote {
		return ErrRevoteNotAllowed
	}

	// delete from cache
	pollsKey := utils.GetPollKey(poll.ID)
	pollsForHomeKey := utils.GetAllPollsForHomeKey(homeID)

	if err := utils.DeleteFromCache(ctx, pollsKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", pollsKey, err)
	}

	if err := utils.DeleteFromCache(ctx, pollsForHomeKey, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", pollsForHomeKey, err)
	}

	if err := s.repo.Unvote(ctx, userID, pollID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModulePoll,
		Action: event.ActionUnvoted,
		Data:   map[string]int{"userID": userID, "pollID": pollID},
	})

	return nil
}

