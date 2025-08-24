package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type mockRoomService struct {
	CreateRoomFunc       func(name string, homeID int) error
	GetRoomByIDFunc      func(roomID int) (*models.Room, error)
	GetRoomsByHomeIDFunc func(homeID int) (*[]models.Room, error)
	DeleteRoomFunc       func(roomID int) error
}

func (m *mockRoomService) CreateRoom(name string, homeID int) error {
	if m.CreateRoomFunc != nil {
		return m.CreateRoomFunc(name, homeID)
	}
	return nil
}

func (m *mockRoomService) GetRoomByID(roomID int) (*models.Room, error) {
	if m.GetRoomByIDFunc != nil {
		return m.GetRoomByIDFunc(roomID)
	}
	return nil, nil
}

func (m *mockRoomService) GetRoomsByHomeID(homeID int) (*[]models.Room, error) {
	if m.GetRoomsByHomeIDFunc != nil {
		return m.GetRoomsByHomeIDFunc(homeID)
	}
	return nil, nil
}

func (m *mockRoomService) DeleteRoom(roomID int) error {
	if m.DeleteRoomFunc != nil {
		return m.DeleteRoomFunc(roomID)
	}
	return nil
}

// POST /rooms/create
func TestRoomHandler_Create_Success(t *testing.T) {
	svc := &mockRoomService{
		CreateRoomFunc: func(name string, homeID int) error {
			assert.Equal(t, "Kitchen", name)
			assert.Equal(t, 1, homeID)
			return nil
		},
	}

	h := handlers.NewRoomHandler(svc)

	reqBody, _ := json.Marshal(models.CreateRoomRequest{
		Name:   "Kitchen",
		HomeID: 1,
	})

	req := httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewReader(reqBody))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Contains(t, rr.Body.String(), "Created successfully")
}

func TestRoomHandler_Create_InvalidJSON(t *testing.T) {
	svc := &mockRoomService{}
	h := handlers.NewRoomHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewBufferString("{bad json}"))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid JSON")
}

func TestRoomHandler_Create_ServiceError(t *testing.T) {
	svc := &mockRoomService{
		CreateRoomFunc: func(name string, homeID int) error {
			return errors.New("service error")
		},
	}

	h := handlers.NewRoomHandler(svc)

	reqBody, _ := json.Marshal(models.CreateRoomRequest{
		Name:   "Kitchen",
		HomeID: 1,
	})

	req := httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewReader(reqBody))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid data")
}

// GET /rooms/{room_id}
func TestRoomHandler_GetByID_Success(t *testing.T) {
	svc := &mockRoomService{
		GetRoomByIDFunc: func(roomID int) (*models.Room, error) {
			assert.Equal(t, 1, roomID)
			return &models.Room{ID: 1, Name: "Kitchen"}, nil
		},
	}

	h := handlers.NewRoomHandler(svc)

	r := chi.NewRouter()
	r.Get("/rooms/{room_id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/rooms/1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Kitchen")
}

func TestRoomHandler_GetByID_InvalidID(t *testing.T) {
	svc := &mockRoomService{}
	h := handlers.NewRoomHandler(svc)

	r := chi.NewRouter()
	r.Get("/rooms/{room_id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/rooms/invalid", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid room ID")
}

func TestRoomHandler_GetByID_ServiceError(t *testing.T) {
	svc := &mockRoomService{
		GetRoomByIDFunc: func(roomID int) (*models.Room, error) {
			return nil, errors.New("service error")
		},
	}

	h := handlers.NewRoomHandler(svc)

	r := chi.NewRouter()
	r.Get("/rooms/{room_id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/rooms/1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "service error")
}

// GET /homes/{home_id}/rooms
func TestRoomHandler_GetByHomeID_Success(t *testing.T) {
	svc := &mockRoomService{
		GetRoomsByHomeIDFunc: func(homeID int) (*[]models.Room, error) {
			assert.Equal(t, 1, homeID)
			rooms := []models.Room{{ID: 1, Name: "Kitchen"}, {ID: 2, Name: "Bedroom"}}
			return &rooms, nil
		},
	}

	h := handlers.NewRoomHandler(svc)

	r := chi.NewRouter()
	r.Get("/homes/{home_id}/rooms", h.GetByHomeID)

	req := httptest.NewRequest(http.MethodGet, "/homes/1/rooms", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Kitchen")
	assert.Contains(t, rr.Body.String(), "Bedroom")
}

func TestRoomHandler_GetByHomeID_InvalidID(t *testing.T) {
	svc := &mockRoomService{}
	h := handlers.NewRoomHandler(svc)

	r := chi.NewRouter()
	r.Get("/homes/{home_id}/rooms", h.GetByHomeID)

	req := httptest.NewRequest(http.MethodGet, "/homes/invalid/rooms", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid home ID")
}

func TestRoomHandler_GetByHomeID_ServiceError(t *testing.T) {
	svc := &mockRoomService{
		GetRoomsByHomeIDFunc: func(homeID int) (*[]models.Room, error) {
			return nil, errors.New("service error")
		},
	}

	h := handlers.NewRoomHandler(svc)

	r := chi.NewRouter()
	r.Get("/homes/{home_id}/rooms", h.GetByHomeID)

	req := httptest.NewRequest(http.MethodGet, "/homes/1/rooms", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "service error")
}

// DELETE /rooms/{room_id}
func TestRoomHandler_Delete_Success(t *testing.T) {
	svc := &mockRoomService{
		DeleteRoomFunc: func(roomID int) error {
			assert.Equal(t, 1, roomID)
			return nil
		},
	}

	h := handlers.NewRoomHandler(svc)

	r := chi.NewRouter()
	r.Delete("/rooms/{room_id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/rooms/1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Deleted successfully")
}

func TestRoomHandler_Delete_InvalidID(t *testing.T) {
	svc := &mockRoomService{}
	h := handlers.NewRoomHandler(svc)

	r := chi.NewRouter()
	r.Delete("/rooms/{room_id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/rooms/invalid", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid room ID")
}

func TestRoomHandler_Delete_ServiceError(t *testing.T) {
	svc := &mockRoomService{
		DeleteRoomFunc: func(roomID int) error {
			return errors.New("delete failed")
		},
	}

	h := handlers.NewRoomHandler(svc)

	r := chi.NewRouter()
	r.Delete("/rooms/{room_id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/rooms/1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "delete failed")
}
