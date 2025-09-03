package repository

import (
	"errors"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
)

type pollRepo struct {
	db *gorm.DB
}

type PollRepository interface {
	// polls
	Create(poll *models.Poll, options []models.Option) error
	FindPollByID(id int) (*models.Poll, error)
	FindPollByOptionID(id int) (*models.Poll, error)
	FindAllPollsByHomeID(id int) (*[]models.Poll, error)
	ClosePoll(id int) error
	Delete(id int) error

	// votes
	Vote(vote *models.Vote) error
}

func NewPollRepository(db *gorm.DB) PollRepository {
	return &pollRepo{db: db}
}

func (r *pollRepo) Create(poll *models.Poll, options []models.Option) error {
	if err := r.db.Create(poll).Error; err != nil {
		return err
	}

	for i := range options {
		options[i].PollID = poll.ID
		if err := r.db.Create(&options[i]).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *pollRepo) FindPollByID(id int) (*models.Poll, error) {
	var poll models.Poll

	// taking memberships also
	if err := r.db.Preload("Options.Votes.User").First(&poll, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &poll, nil
}

func (r *pollRepo) FindPollByOptionID(id int) (*models.Poll, error) {
	var poll models.Poll
	var option models.Option

	// taking memberships also
	if err := r.db.First(&option, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	if err := r.db.First(&poll, option.PollID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &poll, nil
}

func (r *pollRepo) FindAllPollsByHomeID(id int) (*[]models.Poll, error) {
	var polls []models.Poll

	if err := r.db.Where("home_id = ?", id).Preload("Options.Votes.User").Find(&polls).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &polls, nil
}

func (r *pollRepo) ClosePoll(id int) error {
	var poll models.Poll

	if err := r.db.First(&poll, id).Error; err != nil {
		return err
	}

	poll.Status = "closed"
	return r.db.Save(&poll).Error
}

func (r *pollRepo) Delete(id int) error {
	return r.db.Delete(&models.Poll{}, id).Error
}

func (r *pollRepo) Vote(vote *models.Vote) error {
	var poll models.Poll
	var option models.Option
	if err := r.db.First(&option, vote.OptionID).Error; err != nil {
		return err
	}

	if err := r.db.First(&poll, option.PollID).Error; err != nil {
		return err
	}

	if poll.Status == "closed" {
		return errors.New("poll is closed")
	}

	return r.db.Create(vote).Error
}
