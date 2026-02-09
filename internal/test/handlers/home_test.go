package handlers_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/http/middleware"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock repository
type mockHomeRepo struct {
	IsMemberFunc func(ctx context.Context, homeID, userID int) (bool, error)
	IsAdminFunc  func(ctx context.Context, homeID, userID int) (bool, error)
}

func (m *mockHomeRepo) Create(ctx context.Context, h *models.Home) error { return nil }
func (m *mockHomeRepo) FindByID(ctx context.Context, id int) (*models.Home, error) {
	return nil, nil
}
func (m *mockHomeRepo) FindByInviteCode(ctx context.Context, inviteCode string) (*models.Home, error) {
	return nil, nil
}
func (m *mockHomeRepo) Delete(ctx context.Context, id int) error { return nil }
func (m *mockHomeRepo) AddMember(ctx context.Context, id int, userID int, role string) error {
	return nil
}
func (m *mockHomeRepo) DeleteMember(ctx context.Context, id int, userID int) error { return nil }
func (m *mockHomeRepo) GenerateUniqueInviteCode(ctx context.Context) (string, error) {
	return "CODE1234", nil
}
func (m *mockHomeRepo) GetUserHome(ctx context.Context, userID int) (*models.Home, error) {
	return nil, nil
}
func (m *mockHomeRepo) RegenerateCode(ctx context.Context, code string, id int) error { return nil }

func (m *mockHomeRepo) IsMember(ctx context.Context, homeID, userID int) (bool, error) {
	if m.IsMemberFunc != nil {
		return m.IsMemberFunc(ctx, homeID, userID)
	}
	return false, nil
}

func (m *mockHomeRepo) IsAdmin(ctx context.Context, homeID, userID int) (bool, error) {
	if m.IsAdminFunc != nil {
		return m.IsAdminFunc(ctx, homeID, userID)
	}
	return false, nil
}

// Mock service
type mockHomeService struct {
	CreateHomeFunc           func(ctx context.Context, name string, userID int) error
	RegenerateInviteCodeFunc func(ctx context.Context, homeID int) error
	JoinHomeByCodeFunc       func(ctx context.Context, code string, userID int) error
	GetUserHomeFunc          func(ctx context.Context, userID int) (*models.Home, error)
	GetHomeByIDFunc          func(ctx context.Context, userID int) (*models.Home, error)
	DeleteHomeFunc           func(ctx context.Context, homeID int) error
	LeaveHomeFunc            func(ctx context.Context, homeID, userID int) error
	RemoveMemberFunc         func(ctx context.Context, homeID, userID, currentUserID int) error
}

func (m *mockHomeService) CreateHome(ctx context.Context, name string, userID int) error {
	return m.CreateHomeFunc(ctx, name, userID)
}

func (m *mockHomeService) JoinHomeByCode(ctx context.Context, code string, userID int) error {
	return m.JoinHomeByCodeFunc(ctx, code, userID)
}

func (m *mockHomeService) GetUserHome(ctx context.Context, userID int) (*models.Home, error) {
	return m.GetUserHomeFunc(ctx, userID)
}

func (m *mockHomeService) GetHomeByID(ctx context.Context, homeID int) (*models.Home, error) {
	return m.GetHomeByIDFunc(ctx, homeID)
}

func (m *mockHomeService) DeleteHome(ctx context.Context, homeID int) error {
	return m.DeleteHomeFunc(ctx, homeID)
}

func (m *mockHomeService) RegenerateInviteCode(ctx context.Context, homeID int) error {
	return m.RegenerateInviteCodeFunc(ctx, homeID)
}

func (m *mockHomeService) LeaveHome(ctx context.Context, homeID, userID int) error {
	return m.LeaveHomeFunc(ctx, homeID, userID)
}

func (m *mockHomeService) RemoveMember(ctx context.Context, homeID, userID, currentUserID int) error {
	return m.RemoveMemberFunc(ctx, homeID, userID, currentUserID)
}

// Test fixtures
var (
	validCreateHomeRequest = models.CreateHomeRequest{Name: "Test Home"}
	validJoinRequest       = models.JoinRequest{Code: "TESTCODE"}
	validLeaveRequest      = models.LeaveRequest{HomeID: "1"}
	validRemoveMemberReq   = models.RemoveMemberRequest{HomeID: "1", UserID: "2"}
)

func setupHomeHandler(svc *mockHomeService) *handlers.HomeHandler {
	return handlers.NewHomeHandler(svc)
}

func TestHomeHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		userID         int
		mockFunc       func(ctx context.Context, name string, userID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			body:   validCreateHomeRequest,
			userID: 123,
			mockFunc: func(ctx context.Context, name string, userID int) error {
				assert.Equal(t, "Test Home", name)
				assert.Equal(t, 123, userID)
				return nil
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "Created successfully",
		},
		{
			name:           "Invalid JSON",
			body:           "{bad json}",
			userID:         123,
			mockFunc:       func(ctx context.Context, name string, userID int) error { return nil },
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name:           "Unauthorized",
			body:           validCreateHomeRequest,
			userID:         0,
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockHomeService{
				CreateHomeFunc: tt.mockFunc,
			}

			h := setupHomeHandler(svc)

			var req *http.Request
			if tt.name == "Invalid JSON" {
				req = httptest.NewRequest(http.MethodPost, "/homes", bytes.NewBufferString("{bad json}"))
			} else {
				req = makeJSONRequest(http.MethodPost, "/homes", tt.body)
			}

			if tt.userID != 0 {
				req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))
			}

			rr := httptest.NewRecorder()
			h.Create(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestHomeHandler_Join(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		userID         int
		mockFunc       func(ctx context.Context, code string, userID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			body:   validJoinRequest,
			userID: 123,
			mockFunc: func(ctx context.Context, code string, userID int) error {
				assert.Equal(t, "TESTCODE", code)
				assert.Equal(t, 123, userID)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Joined successfully",
		},
		{
			name:           "Invalid JSON",
			body:           "{bad Json}",
			userID:         123,
			mockFunc:       func(ctx context.Context, code string, userID int) error { return nil },
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name:           "Unauthorized",
			body:           validJoinRequest,
			userID:         0,
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockHomeService{
				JoinHomeByCodeFunc: tt.mockFunc,
			}

			h := setupHomeHandler(svc)

			var req *http.Request
			if tt.name == "Invalid JSON" {
				req = httptest.NewRequest(http.MethodPost, "/join", bytes.NewBufferString("{bad Json}"))
			} else {
				req = makeJSONRequest(http.MethodPost, "/join", tt.body)
			}

			if tt.userID != 0 {
				req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))
			}

			rr := httptest.NewRecorder()
			h.Join(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestHomeHandler_GetUserHome(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		mockFunc       func(ctx context.Context, userID int) (*models.Home, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userID: 123,
			mockFunc: func(ctx context.Context, userID int) (*models.Home, error) {
				require.Equal(t, 123, userID)
				return &models.Home{ID: 1, Name: "TestHome"}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "TestHome",
		},
		{
			name:   "Error",
			userID: 123,
			mockFunc: func(ctx context.Context, userID int) (*models.Home, error) {
				return nil, errors.New("test error")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:   "Not Found",
			userID: 123,
			mockFunc: func(ctx context.Context, userID int) (*models.Home, error) {
				return nil, nil
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Home not found",
		},
		{
			name:           "Unauthorized",
			userID:         0,
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockHomeService{
				GetUserHomeFunc: tt.mockFunc,
			}

			h := setupHomeHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/my", nil)
			if tt.userID != 0 {
				req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))
			}

			rr := httptest.NewRecorder()
			h.GetUserHome(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestHomeHandler_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		homeID         string
		userID         int
		isMember       bool
		mockFunc       func(ctx context.Context, homeID int) (*models.Home, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "Success",
			homeID:   "1",
			userID:   123,
			isMember: true,
			mockFunc: func(ctx context.Context, homeID int) (*models.Home, error) {
				require.Equal(t, 1, homeID)
				return &models.Home{ID: 1, Name: "TestHome"}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "TestHome",
		},
		{
			name:     "Error",
			homeID:   "1",
			userID:   123,
			isMember: true,
			mockFunc: func(ctx context.Context, homeID int) (*models.Home, error) {
				return nil, errors.New("test error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
		},
		{
			name:     "Not Found",
			homeID:   "1",
			userID:   123,
			isMember: true,
			mockFunc: func(ctx context.Context, homeID int) (*models.Home, error) {
				return nil, nil
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Home not found",
		},
		{
			name:     "Unauthorized - Not Member",
			homeID:   "1",
			userID:   123,
			isMember: false,
			mockFunc: func(ctx context.Context, homeID int) (*models.Home, error) {
				return &models.Home{ID: 1, Name: "TestHome"}, nil
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "you are not a member",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockHomeService{
				GetHomeByIDFunc: tt.mockFunc,
			}

			h := setupHomeHandler(svc)

			mockRepo := &mockHomeRepo{
				IsMemberFunc: func(ctx context.Context, homeID, userID int) (bool, error) {
					return tt.isMember, nil
				},
			}

			r := chi.NewRouter()
			r.With(middleware.RequireMember(mockRepo)).Get("/homes/{home_id}", h.GetByID)

			req := httptest.NewRequest(http.MethodGet, "/homes/"+tt.homeID, nil)
			req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestHomeHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		homeID         string
		userID         int
		isAdmin        bool
		isMember       bool
		mockFunc       func(ctx context.Context, homeID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "Success",
			homeID:   "1",
			userID:   123,
			isAdmin:  true,
			isMember: true,
			mockFunc: func(ctx context.Context, homeID int) error {
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Deleted successfully",
		},
		{
			name:     "Error",
			homeID:   "1",
			userID:   123,
			isAdmin:  true,
			isMember: true,
			mockFunc: func(ctx context.Context, homeID int) error {
				return errors.New("delete failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "delete failed",
		},
		{
			name:     "Unauthorized - Not Admin",
			homeID:   "1",
			userID:   123,
			isAdmin:  false,
			isMember: false,
			mockFunc: func(ctx context.Context, homeID int) error { return nil },
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "you are not a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockHomeService{
				DeleteHomeFunc: tt.mockFunc,
			}

			h := setupHomeHandler(svc)

			mockRepo := &mockHomeRepo{
				IsAdminFunc:  func(ctx context.Context, homeID, userID int) (bool, error) { return tt.isAdmin, nil },
				IsMemberFunc: func(ctx context.Context, homeID, userID int) (bool, error) { return tt.isMember, nil },
			}

			r := chi.NewRouter()
			r.With(middleware.RequireAdmin(mockRepo)).Delete("/homes/{home_id}", h.Delete)

			req := httptest.NewRequest(http.MethodDelete, "/homes/"+tt.homeID, nil)
			req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestHomeHandler_Leave(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		userID         int
		mockFunc       func(ctx context.Context, homeID, userID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			body:   validLeaveRequest,
			userID: 123,
			mockFunc: func(ctx context.Context, homeID, userID int) error {
				assert.Equal(t, 1, homeID)
				assert.Equal(t, 123, userID)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Leaved successfully",
		},
		{
			name:   "Error",
			body:   validLeaveRequest,
			userID: 123,
			mockFunc: func(ctx context.Context, homeID, userID int) error {
				return errors.New("leave failed")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "leave failed",
		},
		{
			name:           "Unauthorized",
			body:           validLeaveRequest,
			userID:         0,
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockHomeService{
				LeaveHomeFunc: tt.mockFunc,
			}

			h := setupHomeHandler(svc)

			req := makeJSONRequest(http.MethodPost, "/homes/leave", tt.body)
			if tt.userID != 0 {
				req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))
			}

			rr := httptest.NewRecorder()
			h.Leave(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestHomeHandler_RemoveMember(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		userID         int
		mockFunc       func(ctx context.Context, homeID, userID, currentUserID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			body:   validRemoveMemberReq,
			userID: 123,
			mockFunc: func(ctx context.Context, homeID, userID, currentUserID int) error {
				assert.Equal(t, 1, homeID)
				assert.Equal(t, 2, userID)
				assert.Equal(t, 123, currentUserID)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "User removed successfully",
		},
		{
			name:   "Error",
			body:   validRemoveMemberReq,
			userID: 123,
			mockFunc: func(ctx context.Context, homeID, userID, currentUserID int) error {
				return errors.New("remove failed")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "remove failed",
		},
		{
			name:           "Unauthorized",
			body:           validRemoveMemberReq,
			userID:         0,
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockHomeService{
				RemoveMemberFunc: tt.mockFunc,
			}

			h := setupHomeHandler(svc)

			req := makeJSONRequest(http.MethodPost, "/homes/remove-member", tt.body)
			if tt.userID != 0 {
				req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))
			}

			rr := httptest.NewRecorder()
			h.RemoveMember(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

