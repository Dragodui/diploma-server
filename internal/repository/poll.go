package repository

import (
	"context"
	"errors"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type pollRepo struct {
	db *gorm.DB
}

type PollRepository interface {
	// polls
	Create(ctx context.Context, poll *models.Poll, options []models.Option) error
	FindPollByID(ctx context.Context, id int) (*models.Poll, error)
	FindPollByOptionID(ctx context.Context, id int) (*models.Poll, error)
	FindAllPollsByHomeID(ctx context.Context, id int) (*[]models.Poll, error)
	ClosePoll(ctx context.Context, id int) error
	Delete(ctx context.Context, id int) error

	// votes
	Vote(ctx context.Context, vote *models.Vote) error
	Unvote(ctx context.Context, userID, pollID int) error
}

func NewPollRepository(db *gorm.DB) PollRepository {
	return &pollRepo{db: db}
}

func (r *pollRepo) Create(ctx context.Context, poll *models.Poll, options []models.Option) error {
	if err := r.db.WithContext(ctx).Create(poll).Error; err != nil {
		return err
	}

	for i := range options {
		options[i].PollID = poll.ID
		if err := r.db.WithContext(ctx).Create(&options[i]).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *pollRepo) FindPollByID(ctx context.Context, id int) (*models.Poll, error) {
	var poll models.Poll

	// taking memberships also
	if err := r.db.WithContext(ctx).Preload("Options.Votes.User").First(&poll, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &poll, nil
}

func (r *pollRepo) FindPollByOptionID(ctx context.Context, id int) (*models.Poll, error) {
	var poll models.Poll
	var option models.Option

	// taking memberships also
	if err := r.db.WithContext(ctx).First(&option, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	if err := r.db.WithContext(ctx).First(&poll, option.PollID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &poll, nil
}

func (r *pollRepo) FindAllPollsByHomeID(ctx context.Context, id int) (*[]models.Poll, error) {
	var polls []models.Poll

	if err := r.db.WithContext(ctx).Where("home_id = ?", id).Preload("Options.Votes.User").Find(&polls).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &polls, nil
}

func (r *pollRepo) ClosePoll(ctx context.Context, id int) error {
	var poll models.Poll

	if err := r.db.WithContext(ctx).First(&poll, id).Error; err != nil {
		return err
	}

	poll.Status = "closed"
	return r.db.WithContext(ctx).Save(&poll).Error
}

func (r *pollRepo) Delete(ctx context.Context, id int) error {
	// Delete votes associated with options of this poll
	// We need to find options first to delete votes
	var options []models.Option
	if err := r.db.WithContext(ctx).Where("poll_id = ?", id).Find(&options).Error; err != nil {
		return err
	}

	for _, option := range options {
		if err := r.db.WithContext(ctx).Where("option_id = ?", option.ID).Delete(&models.Vote{}).Error; err != nil {
			return err
		}
	}

	// Delete options
	if err := r.db.WithContext(ctx).Where("poll_id = ?", id).Delete(&models.Option{}).Error; err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&models.Poll{}, id).Error
}

func (r *pollRepo) Vote(ctx context.Context, vote *models.Vote) error {
	var poll models.Poll
	var option models.Option
	if err := r.db.WithContext(ctx).First(&option, vote.OptionID).Error; err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).First(&poll, option.PollID).Error; err != nil {
		return err
	}

	if poll.Status == "closed" {
		return errors.New("poll is closed")
	}

	return r.db.WithContext(ctx).Create(vote).Error
}

func (r *pollRepo) Unvote(ctx context.Context, userID, pollID int) error {
	// Find all options for this poll
	var options []models.Option
	if err := r.db.WithContext(ctx).Where("poll_id = ?", pollID).Find(&options).Error; err != nil {
		return err
	}

	// Delete user's votes from all options of this poll
	optionIDs := make([]int, len(options))
	for i, opt := range options {
		optionIDs[i] = opt.ID
	}

	return r.db.WithContext(ctx).Where("user_id = ? AND option_id IN ?", userID, optionIDs).Delete(&models.Vote{}).Error
}

