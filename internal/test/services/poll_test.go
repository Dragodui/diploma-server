package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock PollRepository
type mockPollRepo struct {
	CreateFunc             func(ctx context.Context, poll *models.Poll, options []models.Option) error
	FindPollByIDFunc       func(ctx context.Context, id int) (*models.Poll, error)
	FindPollByOptionIDFunc func(ctx context.Context, id int) (*models.Poll, error)
	FindAllPollsByHomeIDFunc func(ctx context.Context, id int) (*[]models.Poll, error)
	ClosePollFunc          func(ctx context.Context, id int) error
	DeleteFunc             func(ctx context.Context, id int) error
	VoteFunc               func(ctx context.Context, vote *models.Vote) error
	UnvoteFunc             func(ctx context.Context, userID, pollID int) error
}

func (m *mockPollRepo) Create(ctx context.Context, poll *models.Poll, options []models.Option) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, poll, options)
	}
	return nil
}

func (m *mockPollRepo) FindPollByID(ctx context.Context, id int) (*models.Poll, error) {
	if m.FindPollByIDFunc != nil {
		return m.FindPollByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockPollRepo) FindPollByOptionID(ctx context.Context, id int) (*models.Poll, error) {
	if m.FindPollByOptionIDFunc != nil {
		return m.FindPollByOptionIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockPollRepo) FindAllPollsByHomeID(ctx context.Context, id int) (*[]models.Poll, error) {
	if m.FindAllPollsByHomeIDFunc != nil {
		return m.FindAllPollsByHomeIDFunc(ctx, id)
	}
	return &[]models.Poll{}, nil
}

func (m *mockPollRepo) ClosePoll(ctx context.Context, id int) error {
	if m.ClosePollFunc != nil {
		return m.ClosePollFunc(ctx, id)
	}
	return nil
}

func (m *mockPollRepo) Delete(ctx context.Context, id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *mockPollRepo) Vote(ctx context.Context, vote *models.Vote) error {
	if m.VoteFunc != nil {
		return m.VoteFunc(ctx, vote)
	}
	return nil
}

func (m *mockPollRepo) Unvote(ctx context.Context, userID, pollID int) error {
	if m.UnvoteFunc != nil {
		return m.UnvoteFunc(ctx, userID, pollID)
	}
	return nil
}

// Test helpers
func setupPollService(t *testing.T, repo repository.PollRepository) *services.PollService {
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	return services.NewPollService(repo, redisClient)
}

// Create Poll Tests
func TestPollService_Create_Success(t *testing.T) {
	options := []models.OptionRequest{
		{Title: "Option A"},
		{Title: "Option B"},
		{Title: "Option C"},
	}
	endsAt := time.Now().Add(24 * time.Hour)

	repo := &mockPollRepo{
		CreateFunc: func(ctx context.Context, poll *models.Poll, opts []models.Option) error {
			require.Equal(t, "What should we order?", poll.Question)
			require.Equal(t, 1, poll.HomeID)
			require.Equal(t, "public", poll.Type)
			require.True(t, poll.AllowRevote)
			require.Len(t, opts, 3)
			require.Equal(t, "Option A", opts[0].Title)
			return nil
		},
	}

	svc := setupPollService(t, repo)
	err := svc.Create(context.Background(), 1, "What should we order?", "public", options, true, &endsAt)

	assert.NoError(t, err)
}

func TestPollService_Create_RepositoryError(t *testing.T) {
	options := []models.OptionRequest{
		{Title: "Option A"},
		{Title: "Option B"},
	}

	repo := &mockPollRepo{
		CreateFunc: func(ctx context.Context, poll *models.Poll, opts []models.Option) error {
			return errors.New("database error")
		},
	}

	svc := setupPollService(t, repo)
	err := svc.Create(context.Background(), 1, "Question?", "public", options, false, nil)

	assert.Error(t, err)
}

// GetPollByID Tests
func TestPollService_GetPollByID_Success(t *testing.T) {
	expectedPoll := &models.Poll{
		ID:       1,
		Question: "What should we order?",
		HomeID:   1,
		Status:   "open",
	}

	repo := &mockPollRepo{
		FindPollByIDFunc: func(ctx context.Context, id int) (*models.Poll, error) {
			require.Equal(t, 1, id)
			return expectedPoll, nil
		},
	}

	svc := setupPollService(t, repo)
	poll, err := svc.GetPollByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedPoll.Question, poll.Question)
	assert.Equal(t, "open", poll.Status)
}

func TestPollService_GetPollByID_NotFound(t *testing.T) {
	repo := &mockPollRepo{
		FindPollByIDFunc: func(ctx context.Context, id int) (*models.Poll, error) {
			return nil, errors.New("poll not found")
		},
	}

	svc := setupPollService(t, repo)
	_, err := svc.GetPollByID(context.Background(), 999)

	assert.Error(t, err)
}

// GetAllPollsByHomeID Tests
func TestPollService_GetAllPollsByHomeID_Success(t *testing.T) {
	expectedPolls := &[]models.Poll{
		{ID: 1, Question: "Poll 1", HomeID: 1, Status: "open"},
		{ID: 2, Question: "Poll 2", HomeID: 1, Status: "closed"},
	}

	repo := &mockPollRepo{
		FindAllPollsByHomeIDFunc: func(ctx context.Context, homeID int) (*[]models.Poll, error) {
			require.Equal(t, 1, homeID)
			return expectedPolls, nil
		},
	}

	svc := setupPollService(t, repo)
	polls, err := svc.GetAllPollsByHomeID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, *polls, 2)
	assert.Equal(t, "open", (*polls)[0].Status)
}

func TestPollService_GetAllPollsByHomeID_Empty(t *testing.T) {
	repo := &mockPollRepo{
		FindAllPollsByHomeIDFunc: func(ctx context.Context, homeID int) (*[]models.Poll, error) {
			return &[]models.Poll{}, nil
		},
	}

	svc := setupPollService(t, repo)
	polls, err := svc.GetAllPollsByHomeID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, *polls, 0)
}

// ClosePoll Tests
func TestPollService_ClosePoll_Success(t *testing.T) {
	repo := &mockPollRepo{
		ClosePollFunc: func(ctx context.Context, id int) error {
			require.Equal(t, 1, id)
			return nil
		},
	}

	svc := setupPollService(t, repo)
	err := svc.ClosePoll(context.Background(), 1, 1)

	assert.NoError(t, err)
}

// DeletePoll Tests
func TestPollService_Delete_Success(t *testing.T) {
	repo := &mockPollRepo{
		DeleteFunc: func(ctx context.Context, id int) error {
			require.Equal(t, 1, id)
			return nil
		},
	}

	svc := setupPollService(t, repo)
	err := svc.Delete(context.Background(), 1, 1)

	assert.NoError(t, err)
}

func TestPollService_Delete_NotFound(t *testing.T) {
	repo := &mockPollRepo{
		DeleteFunc: func(ctx context.Context, id int) error {
			return errors.New("poll not found")
		},
	}

	svc := setupPollService(t, repo)
	err := svc.Delete(context.Background(), 999, 1)

	assert.Error(t, err)
}

// Vote Tests
func TestPollService_Vote_Success(t *testing.T) {
	poll := &models.Poll{
		ID:     1,
		HomeID: 1,
		Status: "open",
	}

	repo := &mockPollRepo{
		FindPollByOptionIDFunc: func(ctx context.Context, id int) (*models.Poll, error) {
			require.Equal(t, 2, id)
			return poll, nil
		},
		VoteFunc: func(ctx context.Context, vote *models.Vote) error {
			require.Equal(t, 5, vote.UserID)
			require.Equal(t, 2, vote.OptionID)
			return nil
		},
	}

	svc := setupPollService(t, repo)
	err := svc.Vote(context.Background(), 5, 2, 1)

	assert.NoError(t, err)
}

func TestPollService_Vote_PollClosed(t *testing.T) {
	poll := &models.Poll{
		ID:     1,
		HomeID: 1,
		Status: "closed",
	}

	repo := &mockPollRepo{
		FindPollByOptionIDFunc: func(ctx context.Context, id int) (*models.Poll, error) {
			return poll, nil
		},
	}

	svc := setupPollService(t, repo)
	err := svc.Vote(context.Background(), 5, 2, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}

// Unvote Tests
func TestPollService_Unvote_Success(t *testing.T) {
	repo := &mockPollRepo{
		UnvoteFunc: func(ctx context.Context, userID, pollID int) error {
			require.Equal(t, 5, userID)
			require.Equal(t, 1, pollID)
			return nil
		},
	}

	svc := setupPollService(t, repo)
	err := svc.Unvote(context.Background(), 5, 1, 1)

	assert.NoError(t, err)
}

func TestPollService_Unvote_RepositoryError(t *testing.T) {
	repo := &mockPollRepo{
		UnvoteFunc: func(ctx context.Context, userID, pollID int) error {
			return errors.New("vote not found")
		},
	}

	svc := setupPollService(t, repo)
	err := svc.Unvote(context.Background(), 5, 1, 1)

	assert.Error(t, err)
}
