package handlers_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test fixtures
var (
	validPoll = &models.Poll{
		ID:       1,
		HomeID:   1,
		Question: "What's for dinner?",
		Type:     "public",
		Status:   "open",
		Options: []models.Option{
			{
				ID:     1,
				PollID: 1,
				Title:  "Pizza",
				Votes: []models.Vote{
					{ID: 1, UserID: 123, OptionID: 1},
				},
			},
			{
				ID:     2,
				PollID: 1,
				Title:  "Pasta",
				Votes:  []models.Vote{},
			},
		},
	}

	validCreatePollRequest = models.CreatePollRequest{
		Question: "What's for dinner?",
		Type:     "public",
		Options: []models.OptionRequest{
			{Title: "Pizza"},
			{Title: "Pasta"},
		},
	}

	validVoteRequest = models.VoteRequest{
		OptionID: 1,
	}
)

func setupPollHandler(mockSvc *mockPollService) *handlers.PollHandler {
	return handlers.NewPollHandler(mockSvc)
}

func setupPollRouter(h *handlers.PollHandler) *chi.Mux {
	r := chi.NewRouter()

	// Add middleware to set user ID for tests
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Mock user ID in context for tests
			r = r.WithContext(utils.WithUserID(r.Context(), 123))
			next.ServeHTTP(w, r)
		})
	})

	// Poll routes
	r.Post("/homes/{home_id}/polls", h.Create)
	r.Get("/homes/{home_id}/polls", h.GetAllByHomeID)
	r.Get("/homes/{home_id}/polls/{poll_id}", h.GetByID)
	r.Patch("/homes/{home_id}/polls/{poll_id}/close", h.Close)
	r.Delete("/homes/{home_id}/polls/{poll_id}", h.Delete)
	r.Post("/homes/{home_id}/polls/{poll_id}/vote", h.Vote)

	return r
}

// Mock service
type mockPollService struct {
	// Polls
	CreateFunc              func(ctx context.Context, homeID int, question, pollType string, options []models.OptionRequest, allowRevote bool, endsAt *time.Time) error
	GetPollByIDFunc         func(ctx context.Context, pollID int) (*models.Poll, error)
	GetAllPollsByHomeIDFunc func(ctx context.Context, homeID int) (*[]models.Poll, error)
	ClosePollFunc           func(ctx context.Context, pollID, homeID int) error
	DeleteFunc              func(ctx context.Context, pollID, homeID int) error

	// Votes
	VoteFunc   func(ctx context.Context, userID, optionID, homeID int) error
	UnvoteFunc func(ctx context.Context, userID, pollID, homeID int) error
}

// Poll methods
func (m *mockPollService) Create(ctx context.Context, homeID int, question, pollType string, options []models.OptionRequest, allowRevote bool, endsAt *time.Time) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, homeID, question, pollType, options, allowRevote, endsAt)
	}
	return nil
}

func (m *mockPollService) GetPollByID(ctx context.Context, pollID int) (*models.Poll, error) {
	if m.GetPollByIDFunc != nil {
		return m.GetPollByIDFunc(ctx, pollID)
	}
	return nil, nil
}

func (m *mockPollService) GetAllPollsByHomeID(ctx context.Context, homeID int) (*[]models.Poll, error) {
	if m.GetAllPollsByHomeIDFunc != nil {
		return m.GetAllPollsByHomeIDFunc(ctx, homeID)
	}
	return nil, nil
}

func (m *mockPollService) ClosePoll(ctx context.Context, pollID, homeID int) error {
	if m.ClosePollFunc != nil {
		return m.ClosePollFunc(ctx, pollID, homeID)
	}
	return nil
}

func (m *mockPollService) Delete(ctx context.Context, pollID, homeID int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, pollID, homeID)
	}
	return nil
}

func (m *mockPollService) Vote(ctx context.Context, userID, optionID, homeID int) error {
	if m.VoteFunc != nil {
		return m.VoteFunc(ctx, userID, optionID, homeID)
	}
	return nil
}

func (m *mockPollService) Unvote(ctx context.Context, userID, pollID, homeID int) error {
	if m.UnvoteFunc != nil {
		return m.UnvoteFunc(ctx, userID, pollID, homeID)
	}
	return nil
}

// POLL TESTS
func TestPollHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		homeID         string
		body           interface{}
		mockFunc       func(ctx context.Context, homeID int, question, pollType string, options []models.OptionRequest, allowRevote bool, endsAt *time.Time) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			homeID: "1",
			body:   validCreatePollRequest,
			mockFunc: func(ctx context.Context, homeID int, question, pollType string, options []models.OptionRequest, allowRevote bool, endsAt *time.Time) error {
				assert.Equal(t, 1, homeID)
				assert.Equal(t, "What's for dinner?", question)
				assert.Equal(t, "public", pollType)
				assert.Len(t, options, 2)
				assert.Equal(t, "Pizza", options[0].Title)
				assert.Equal(t, "Pasta", options[1].Title)
				return nil
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "Poll created successfully",
		},
		{
			name:           "Invalid Home ID",
			homeID:         "invalid",
			body:           validCreatePollRequest,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid home ID",
		},
		{
			name:           "Invalid JSON",
			homeID:         "1",
			body:           "{bad json}",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name:   "Validation Error",
			homeID: "1",
			body: models.CreatePollRequest{
				Question: "", // Empty question should fail validation
				Type:     "public",
				Options:  []models.OptionRequest{{Title: "Option1"}},
			},
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "validation",
		},
		{
			name:   "Service Error",
			homeID: "1",
			body:   validCreatePollRequest,
			mockFunc: func(ctx context.Context, homeID int, question, pollType string, options []models.OptionRequest, allowRevote bool, endsAt *time.Time) error {
				return errors.New("service error")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockPollService{
				CreateFunc: tt.mockFunc,
			}

			h := setupPollHandler(svc)
			r := setupPollRouter(h)

			var req *http.Request
			if tt.name == "Invalid JSON" {
				req = httptest.NewRequest(http.MethodPost, "/homes/"+tt.homeID+"/polls",
					bytes.NewBufferString("{bad json}"))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = makeJSONRequest(http.MethodPost, "/homes/"+tt.homeID+"/polls", tt.body)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestPollHandler_GetAllByHomeID(t *testing.T) {
	tests := []struct {
		name           string
		homeID         string
		mockFunc       func(ctx context.Context, homeID int) (*[]models.Poll, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			homeID: "1",
			mockFunc: func(ctx context.Context, homeID int) (*[]models.Poll, error) {
				require.Equal(t, 1, homeID)
				polls := []models.Poll{*validPoll}
				return &polls, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "What's for dinner?",
		},
		{
			name:           "Invalid Home ID",
			homeID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid home ID",
		},
		{
			name:   "Service Error",
			homeID: "1",
			mockFunc: func(ctx context.Context, homeID int) (*[]models.Poll, error) {
				return nil, errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockPollService{
				GetAllPollsByHomeIDFunc: tt.mockFunc,
			}

			h := setupPollHandler(svc)
			r := setupPollRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/homes/"+tt.homeID+"/polls", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestPollHandler_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		homeID         string
		pollID         string
		mockFunc       func(ctx context.Context, pollID int) (*models.Poll, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			homeID: "1",
			pollID: "1",
			mockFunc: func(ctx context.Context, pollID int) (*models.Poll, error) {
				require.Equal(t, 1, pollID)
				return validPoll, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "What's for dinner?",
		},
		{
			name:           "Invalid Poll ID",
			homeID:         "1",
			pollID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid poll ID",
		},
		{
			name:   "Poll Not Found",
			homeID: "1",
			pollID: "1",
			mockFunc: func(ctx context.Context, pollID int) (*models.Poll, error) {
				return nil, nil
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Poll not found",
		},
		{
			name:   "Service Error",
			homeID: "1",
			pollID: "1",
			mockFunc: func(ctx context.Context, pollID int) (*models.Poll, error) {
				return nil, errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockPollService{
				GetPollByIDFunc: tt.mockFunc,
			}

			h := setupPollHandler(svc)
			r := setupPollRouter(h)

			req := httptest.NewRequest(http.MethodGet,
				"/homes/"+tt.homeID+"/polls/"+tt.pollID, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestPollHandler_Close(t *testing.T) {
	tests := []struct {
		name           string
		homeID         string
		pollID         string
		mockFunc       func(ctx context.Context, pollID, homeID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			homeID: "1",
			pollID: "1",
			mockFunc: func(ctx context.Context, pollID, homeID int) error {
				assert.Equal(t, 1, pollID)
				assert.Equal(t, 1, homeID)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Poll closed successfully",
		},
		{
			name:           "Invalid Home ID",
			homeID:         "invalid",
			pollID:         "1",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid home ID",
		},
		{
			name:           "Invalid Poll ID",
			homeID:         "1",
			pollID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid poll ID",
		},
		{
			name:   "Service Error",
			homeID: "1",
			pollID: "1",
			mockFunc: func(ctx context.Context, pollID, homeID int) error {
				return errors.New("close failed")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "close failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockPollService{
				ClosePollFunc: tt.mockFunc,
			}

			h := setupPollHandler(svc)
			r := setupPollRouter(h)

			req := httptest.NewRequest(http.MethodPatch,
				"/homes/"+tt.homeID+"/polls/"+tt.pollID+"/close", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestPollHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		homeID         string
		pollID         string
		mockFunc       func(ctx context.Context, pollID, homeID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			homeID: "1",
			pollID: "1",
			mockFunc: func(ctx context.Context, pollID, homeID int) error {
				assert.Equal(t, 1, pollID)
				assert.Equal(t, 1, homeID)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Poll deleted successfully",
		},
		{
			name:           "Invalid Home ID",
			homeID:         "invalid",
			pollID:         "1",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid home ID",
		},
		{
			name:           "Invalid Poll ID",
			homeID:         "1",
			pollID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid poll ID",
		},
		{
			name:   "Service Error",
			homeID: "1",
			pollID: "1",
			mockFunc: func(ctx context.Context, pollID, homeID int) error {
				return errors.New("delete failed")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "delete failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockPollService{
				DeleteFunc: tt.mockFunc,
			}

			h := setupPollHandler(svc)
			r := setupPollRouter(h)

			req := httptest.NewRequest(http.MethodDelete,
				"/homes/"+tt.homeID+"/polls/"+tt.pollID, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestPollHandler_Vote(t *testing.T) {
	tests := []struct {
		name           string
		homeID         string
		pollID         string
		body           interface{}
		mockFunc       func(ctx context.Context, userID, optionID, homeID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			homeID: "1",
			pollID: "1",
			body:   validVoteRequest,
			mockFunc: func(ctx context.Context, userID, optionID, homeID int) error {
				assert.Equal(t, 123, userID) // From mock middleware
				assert.Equal(t, 1, optionID)
				assert.Equal(t, 1, homeID)
				return nil
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "Vote submitted successfully",
		},
		{
			name:           "Invalid Home ID",
			homeID:         "invalid",
			pollID:         "1",
			body:           validVoteRequest,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid home ID",
		},
		{
			name:           "Invalid JSON",
			homeID:         "1",
			pollID:         "1",
			body:           "{bad json}",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name:   "Validation Error",
			homeID: "1",
			pollID: "1",
			body: models.VoteRequest{
				OptionID: 0, // Invalid option ID
			},
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "validation",
		},
		{
			name:   "Service Error",
			homeID: "1",
			pollID: "1",
			body:   validVoteRequest,
			mockFunc: func(ctx context.Context, userID, optionID, homeID int) error {
				return errors.New("vote failed")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "vote failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockPollService{
				VoteFunc: tt.mockFunc,
			}

			h := setupPollHandler(svc)
			r := setupPollRouter(h)

			var req *http.Request
			if tt.name == "Invalid JSON" {
				req = httptest.NewRequest(http.MethodPost,
					"/homes/"+tt.homeID+"/polls/"+tt.pollID+"/vote",
					bytes.NewBufferString("{bad json}"))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = makeJSONRequest(http.MethodPost,
					"/homes/"+tt.homeID+"/polls/"+tt.pollID+"/vote", tt.body)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

// Test unauthorized access for vote endpoint
func TestPollHandler_Vote_Unauthorized(t *testing.T) {
	svc := &mockPollService{}
	h := setupPollHandler(svc)

	// Create router without user middleware to simulate unauthorized request
	r := chi.NewRouter()
	r.Post("/homes/{home_id}/polls/{poll_id}/vote", h.Vote)

	req := makeJSONRequest(http.MethodPost, "/homes/1/polls/1/vote", validVoteRequest)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assertJSONResponse(t, rr, http.StatusUnauthorized, "Unauthorized")
}
