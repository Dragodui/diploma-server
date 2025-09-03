package handlers_test

import (
	"bytes"
	"encoding/json"
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
)

type mockHomeRepo struct {
	IsMemberFunc func(homeID, userID int) (bool, error)
	IsAdminFunc  func(homeID, userID int) (bool, error)
}

// unused function to implement HomeRepository
func (m *mockHomeRepo) Create(h *models.Home) error                              { return nil }
func (m *mockHomeRepo) FindByID(id int) (*models.Home, error)                    { return nil, nil }
func (m *mockHomeRepo) FindByInviteCode(inviteCode string) (*models.Home, error) { return nil, nil }
func (m *mockHomeRepo) Delete(id int) error                                      { return nil }
func (m *mockHomeRepo) AddMember(id int, userID int, role string) error          { return nil }
func (m *mockHomeRepo) DeleteMember(id int, userID int) error                    { return nil }
func (m *mockHomeRepo) GenerateUniqueInviteCode() (string, error)                { return "CODE1234", nil }
func (m *mockHomeRepo) GetUserHome(userID int) (*models.Home, error)             { return nil, nil }

func (m *mockHomeRepo) IsMember(homeID, userID int) (bool, error) {
	if m.IsMemberFunc != nil {
		return m.IsMemberFunc(homeID, userID)
	}
	return false, nil
}

func (m *mockHomeRepo) IsAdmin(homeID, userID int) (bool, error) {
	if m.IsAdminFunc != nil {
		return m.IsAdminFunc(homeID, userID)
	}
	return false, nil
}

type mockHomeService struct {
	CreateHomeFunc     func(name string, userID int) error
	JoinHomeByCodeFunc func(code string, userID int) error
	GetUserHomeFunc    func(userID int) (*models.Home, error)
	GetHomeByIDFunc    func(userID int) (*models.Home, error)
	DeleteHomeFunc     func(homeID int) error
	LeaveHomeFunc      func(homeID, userID int) error
	RemoveMemberFunc   func(homeID, userID, currentUserID int) error
}

func (m *mockHomeService) CreateHome(name string, userID int) error {
	return m.CreateHomeFunc(name, userID)
}

func (m *mockHomeService) JoinHomeByCode(code string, userID int) error {
	return m.JoinHomeByCodeFunc(code, userID)
}

func (m *mockHomeService) GetUserHome(userID int) (*models.Home, error) {
	return m.GetUserHomeFunc(userID)
}

func (m *mockHomeService) GetHomeByID(homeID int) (*models.Home, error) {
	return m.GetHomeByIDFunc(homeID)
}

func (m *mockHomeService) DeleteHome(homeID int) error {
	return m.DeleteHomeFunc(homeID)
}

func (m *mockHomeService) LeaveHome(homeID, userID int) error { return m.LeaveHomeFunc(homeID, userID) }
func (m *mockHomeService) RemoveMember(homeID, userID, currentUserID int) error {
	return m.RemoveMemberFunc(homeID, userID, currentUserID)
}

// POST /homes/create
func TestHomeHandler_Create_Success(t *testing.T) {
	svc := &mockHomeService{
		CreateHomeFunc: func(name string, userID int) error {
			assert.Equal(t, "Test Home", name)
			assert.Equal(t, 123, userID)
			return nil
		},
	}
	h := handlers.NewHomeHandler(svc)
	reqBody, _ := json.Marshal(models.CreateHomeRequest{Name: "Test Home"})
	req := httptest.NewRequest(http.MethodPost, "/homes", bytes.NewReader(reqBody))
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Contains(t, rr.Body.String(), "Created successfully")
}

func TestHomeHandler_Create_InvalidJSON(t *testing.T) {
	svc := &mockHomeService{
		CreateHomeFunc: func(name string, userID int) error {
			return nil
		},
	}

	h := handlers.NewHomeHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/homes", bytes.NewBufferString("{bad json}"))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid JSON")
}

func TestHomeHandler_Create_Unauthorized(t *testing.T) {
	svc := &mockHomeService{}
	h := handlers.NewHomeHandler(svc)

	reqBody, _ := json.Marshal(models.CreateHomeRequest{Name: "Test Home"})
	req := httptest.NewRequest(http.MethodPost, "/homes", bytes.NewReader(reqBody))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Unauthorized")

}

// POST /homes/join
func TestHomeHandler_Join_Success(t *testing.T) {
	svc := &mockHomeService{
		JoinHomeByCodeFunc: func(code string, userID int) error {
			assert.Equal(t, "TESTCODE", code)
			assert.Equal(t, 123, userID)
			return nil
		},
	}

	h := handlers.NewHomeHandler(svc)

	reqBody, _ := json.Marshal(models.JoinRequest{
		Code: "TESTCODE",
	})

	req := httptest.NewRequest(http.MethodPost, "/join", bytes.NewBuffer(reqBody))
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()
	h.Join(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Joined successfully")
}

func TestHomeHandler_Join_InvalidJSON(t *testing.T) {
	svc := &mockHomeService{
		JoinHomeByCodeFunc: func(code string, userID int) error { return nil },
	}

	h := handlers.NewHomeHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/join", bytes.NewBufferString("{bad Json}"))
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	h.Join(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid JSON")
}

func TestHomeHandler_Join_Unauthorized(t *testing.T) {
	svc := &mockHomeService{}
	h := handlers.NewHomeHandler(svc)

	reqBody, _ := json.Marshal(models.JoinRequest{
		Code: "TESTCODE",
	})
	req := httptest.NewRequest(http.MethodPost, "/join", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	h.Join(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Unauthorized")
}

// GET /homes/my
func TestHomeHandler_My_Success(t *testing.T) {
	svc := &mockHomeService{
		GetUserHomeFunc: func(userID int) (*models.Home, error) {
			assert.Equal(t, 123, userID)
			return &models.Home{ID: 1, Name: "TestHome"}, nil
		},
	}

	h := handlers.NewHomeHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/my", nil)
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	h.GetUserHome(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "TestHome")
}

func TestHomeHandler_My_Error(t *testing.T) {
	svc := &mockHomeService{
		GetUserHomeFunc: func(userID int) (*models.Home, error) {
			return nil, errors.New("test error")
		},
	}

	h := handlers.NewHomeHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/my", nil)
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	h.GetUserHome(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHomeHandler_My_NotFound(t *testing.T) {
	svc := &mockHomeService{
		GetUserHomeFunc: func(userID int) (*models.Home, error) {
			return nil, nil
		},
	}

	h := handlers.NewHomeHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/my", nil)
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	h.GetUserHome(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "Home not found")
}

func TestHomeHandler_My_Unauthorized(t *testing.T) {
	svc := &mockHomeService{}

	h := handlers.NewHomeHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/my", nil)

	rr := httptest.NewRecorder()

	h.GetUserHome(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Unauthorized")
}

// GET /homes/{id}
func TestHomeHandler_Get_Success(t *testing.T) {
	svc := &mockHomeService{
		GetHomeByIDFunc: func(homeID int) (*models.Home, error) {
			assert.Equal(t, 1, homeID)
			return &models.Home{ID: 1, Name: "TestHome"}, nil
		},
	}

	h := handlers.NewHomeHandler(svc)

	mockRepo := &mockHomeRepo{
		IsMemberFunc: func(homeID, userID int) (bool, error) {
			assert.Equal(t, 1, homeID)
			assert.Equal(t, 123, userID)
			return true, nil
		},
	}

	r := chi.NewRouter()
	r.With(middleware.RequireMember(mockRepo)).Get("/homes/{home_id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/homes/1", nil)
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "TestHome")
}

func TestHomeHandler_Get_Error(t *testing.T) {
	svc := &mockHomeService{
		GetHomeByIDFunc: func(homeID int) (*models.Home, error) {
			assert.Equal(t, 1, homeID)
			return nil, errors.New("test error")
		},
	}

	h := handlers.NewHomeHandler(svc)

	mockRepo := &mockHomeRepo{
		IsMemberFunc: func(homeID, userID int) (bool, error) {
			assert.Equal(t, 1, homeID)
			assert.Equal(t, 123, userID)
			return true, nil
		},
	}

	r := chi.NewRouter()
	r.With(middleware.RequireMember(mockRepo)).Get("/homes/{home_id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/homes/1", nil)
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

}

func TestHomeHandler_Get_NotFound(t *testing.T) {
	svc := &mockHomeService{
		GetHomeByIDFunc: func(homeID int) (*models.Home, error) {
			assert.Equal(t, 1, homeID)
			return nil, nil
		},
	}

	h := handlers.NewHomeHandler(svc)

	mockRepo := &mockHomeRepo{
		IsMemberFunc: func(homeID, userID int) (bool, error) {
			assert.Equal(t, 1, homeID)
			assert.Equal(t, 123, userID)
			return true, nil
		},
	}

	r := chi.NewRouter()
	r.With(middleware.RequireMember(mockRepo)).Get("/homes/{home_id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/homes/1", nil)
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "Home not found")
}

func TestHomeHandler_Get_Unauthorized(t *testing.T) {
	svc := &mockHomeService{
		GetHomeByIDFunc: func(homeID int) (*models.Home, error) {
			return &models.Home{ID: homeID, Name: "TestHome"}, nil
		},
	}

	h := handlers.NewHomeHandler(svc)

	mockRepo := &mockHomeRepo{
		IsMemberFunc: func(homeID, userID int) (bool, error) {
			return false, nil
		},
	}

	r := chi.NewRouter()
	r.With(middleware.RequireMember(mockRepo)).Get("/homes/{home_id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/homes/1", nil)
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "you are not a member")
}

// DELETE /homes/{home_id}
func TestHomeHandler_Delete_Success(t *testing.T) {
	svc := &mockHomeService{
		DeleteHomeFunc: func(homeID int) error {
			return nil
		},
	}

	h := handlers.NewHomeHandler(svc)

	mockRepo := &mockHomeRepo{
		IsAdminFunc: func(homeID, userID int) (bool, error) {
			assert.Equal(t, 1, homeID)
			assert.Equal(t, 123, userID)
			return true, nil
		},
		IsMemberFunc: func(homeID, userID int) (bool, error) {
			assert.Equal(t, 1, homeID)
			assert.Equal(t, 123, userID)
			return true, nil
		},
	}

	r := chi.NewRouter()
	r.With(middleware.RequireAdmin(mockRepo)).Delete("/homes/{home_id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/homes/1", nil)
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Deleted successfully")

}

func TestHomeHandler_Delete_Error(t *testing.T) {
	svc := &mockHomeService{
		DeleteHomeFunc: func(homeID int) error {
			return errors.New("delete failed")
		},
	}

	h := handlers.NewHomeHandler(svc)

	mockRepo := &mockHomeRepo{
		IsAdminFunc:  func(homeID, userID int) (bool, error) { return true, nil },
		IsMemberFunc: func(homeID, userID int) (bool, error) { return true, nil },
	}

	r := chi.NewRouter()
	r.With(middleware.RequireAdmin(mockRepo)).Delete("/homes/{home_id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/homes/1", nil)
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "delete failed")
}

func TestHomeHandler_Delete_Unauthorized(t *testing.T) {
	svc := &mockHomeService{
		DeleteHomeFunc: func(homeID int) error { return nil },
	}

	h := handlers.NewHomeHandler(svc)

	mockRepo := &mockHomeRepo{
		IsAdminFunc: func(homeID, userID int) (bool, error) {
			return false, nil
		},
	}

	r := chi.NewRouter()
	r.With(middleware.RequireAdmin(mockRepo)).Delete("/homes/{home_id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/homes/1", nil)
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "you are not a")
}

// POST /homes/leave
func TestHomeHandler_Leave_Success(t *testing.T) {
	svc := &mockHomeService{
		LeaveHomeFunc: func(homeID, userID int) error {
			assert.Equal(t, 1, homeID)
			assert.Equal(t, 123, userID)
			return nil
		},
	}

	h := handlers.NewHomeHandler(svc)

	reqBody, _ := json.Marshal(models.LeaveRequest{HomeID: "1"})
	req := httptest.NewRequest(http.MethodPost, "/homes/leave", bytes.NewBuffer(reqBody))
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	h.Leave(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Leaved successfully")
}

func TestHomeHandler_Leave_Error(t *testing.T) {
	svc := &mockHomeService{
		LeaveHomeFunc: func(homeID, userID int) error {
			return errors.New("leave failed")
		},
	}

	h := handlers.NewHomeHandler(svc)

	reqBody, _ := json.Marshal(models.LeaveRequest{HomeID: "1"})
	req := httptest.NewRequest(http.MethodPost, "/homes/leave", bytes.NewBuffer(reqBody))
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	h.Leave(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "leave failed")
}

func TestHomeHandler_Leave_Unauthorized(t *testing.T) {
	svc := &mockHomeService{}

	h := handlers.NewHomeHandler(svc)

	reqBody, _ := json.Marshal(models.LeaveRequest{HomeID: "1"})
	req := httptest.NewRequest(http.MethodPost, "/homes/leave", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	h.Leave(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Unauthorized")
}

// POST /homes/remove-member
func TestHomeHandler_RemoveMember_Success(t *testing.T) {
	svc := &mockHomeService{
		RemoveMemberFunc: func(homeID, userID, currentUserID int) error {
			assert.Equal(t, 1, homeID)
			assert.Equal(t, 2, userID)
			assert.Equal(t, 123, currentUserID)
			return nil
		},
	}

	h := handlers.NewHomeHandler(svc)

	reqBody, _ := json.Marshal(models.RemoveMemberRequest{HomeID: "1", UserID: "2"})
	req := httptest.NewRequest(http.MethodPost, "/homes/remove-member", bytes.NewBuffer(reqBody))
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	h.RemoveMember(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "User removed successfully")
}

func TestHomeHandler_RemoveMember_Error(t *testing.T) {
	svc := &mockHomeService{
		RemoveMemberFunc: func(homeID, userID, currentUserID int) error {
			return errors.New("remove failed")
		},
	}

	h := handlers.NewHomeHandler(svc)

	reqBody, _ := json.Marshal(models.RemoveMemberRequest{HomeID: "1", UserID: "2"})
	req := httptest.NewRequest(http.MethodPost, "/homes/remove-member", bytes.NewBuffer(reqBody))
	req = req.WithContext(utils.WithUserID(req.Context(), 123))
	rr := httptest.NewRecorder()

	h.RemoveMember(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "remove failed")
}

func TestHomeHandler_RemoveMember_Unauthorized(t *testing.T) {
	svc := &mockHomeService{}

	h := handlers.NewHomeHandler(svc)

	reqBody, _ := json.Marshal(models.RemoveMemberRequest{HomeID: "1", UserID: "2"})
	req := httptest.NewRequest(http.MethodPost, "/homes/remove-member", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	h.RemoveMember(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Unauthorized")
}
