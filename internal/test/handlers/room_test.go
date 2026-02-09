package handlers_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock service
type mockRoomService struct {
	CreateRoomFunc       func(ctx context.Context, name string, homeID int) error
	GetRoomByIDFunc      func(ctx context.Context, roomID int) (*models.Room, error)
	GetRoomsByHomeIDFunc func(ctx context.Context, homeID int) (*[]models.Room, error)
	DeleteRoomFunc       func(ctx context.Context, roomID int) error
}

func (m *mockRoomService) CreateRoom(ctx context.Context, name string, homeID int) error {
	if m.CreateRoomFunc != nil {
		return m.CreateRoomFunc(ctx, name, homeID)
	}
	return nil
}

func (m *mockRoomService) GetRoomByID(ctx context.Context, roomID int) (*models.Room, error) {
	if m.GetRoomByIDFunc != nil {
		return m.GetRoomByIDFunc(ctx, roomID)
	}
	return nil, nil
}

func (m *mockRoomService) GetRoomsByHomeID(ctx context.Context, homeID int) (*[]models.Room, error) {
	if m.GetRoomsByHomeIDFunc != nil {
		return m.GetRoomsByHomeIDFunc(ctx, homeID)
	}
	return nil, nil
}

func (m *mockRoomService) DeleteRoom(ctx context.Context, roomID int) error {
	if m.DeleteRoomFunc != nil {
		return m.DeleteRoomFunc(ctx, roomID)
	}
	return nil
}

// Test fixtures
var validCreateRoomRequest = models.CreateRoomRequest{
	Name:   "Kitchen",
	HomeID: 1,
}

func setupRoomHandler(svc *mockRoomService) *handlers.RoomHandler {
	return handlers.NewRoomHandler(svc)
}

func setupRoomRouter(h *handlers.RoomHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/rooms/{room_id}", h.GetByID)
	r.Get("/homes/{home_id}/rooms", h.GetByHomeID)
	r.Delete("/rooms/{room_id}", h.Delete)
	return r
}

func TestRoomHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		mockFunc       func(ctx context.Context, name string, homeID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			body: validCreateRoomRequest,
			mockFunc: func(ctx context.Context, name string, homeID int) error {
				assert.Equal(t, "Kitchen", name)
				assert.Equal(t, 1, homeID)
				return nil
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "Created successfully",
		},
		{
			name:           "Invalid JSON",
			body:           "{bad json}",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name: "Service Error",
			body: validCreateRoomRequest,
			mockFunc: func(ctx context.Context, name string, homeID int) error {
				return errors.New("service error")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockRoomService{
				CreateRoomFunc: tt.mockFunc,
			}

			h := setupRoomHandler(svc)

			var req *http.Request
			if tt.name == "Invalid JSON" {
				req = httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewBufferString("{bad json}"))
			} else {
				req = makeJSONRequest(http.MethodPost, "/rooms", tt.body)
			}

			rr := httptest.NewRecorder()
			h.Create(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestRoomHandler_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		roomID         string
		mockFunc       func(ctx context.Context, roomID int) (*models.Room, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			roomID: "1",
			mockFunc: func(ctx context.Context, roomID int) (*models.Room, error) {
				require.Equal(t, 1, roomID)
				return &models.Room{ID: 1, Name: "Kitchen"}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Kitchen",
		},
		{
			name:           "Invalid ID",
			roomID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid room ID",
		},
		{
			name:   "Service Error",
			roomID: "1",
			mockFunc: func(ctx context.Context, roomID int) (*models.Room, error) {
				return nil, errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockRoomService{
				GetRoomByIDFunc: tt.mockFunc,
			}

			h := setupRoomHandler(svc)
			r := setupRoomRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/rooms/"+tt.roomID, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestRoomHandler_GetByHomeID(t *testing.T) {
	tests := []struct {
		name           string
		homeID         string
		mockFunc       func(ctx context.Context, homeID int) (*[]models.Room, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			homeID: "1",
			mockFunc: func(ctx context.Context, homeID int) (*[]models.Room, error) {
				require.Equal(t, 1, homeID)
				rooms := []models.Room{{ID: 1, Name: "Kitchen"}, {ID: 2, Name: "Bedroom"}}
				return &rooms, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Kitchen",
		},
		{
			name:           "Invalid ID",
			homeID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid home ID",
		},
		{
			name:   "Service Error",
			homeID: "1",
			mockFunc: func(ctx context.Context, homeID int) (*[]models.Room, error) {
				return nil, errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockRoomService{
				GetRoomsByHomeIDFunc: tt.mockFunc,
			}

			h := setupRoomHandler(svc)
			r := setupRoomRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/homes/"+tt.homeID+"/rooms", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestRoomHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		roomID         string
		mockFunc       func(ctx context.Context, roomID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			roomID: "1",
			mockFunc: func(ctx context.Context, roomID int) error {
				require.Equal(t, 1, roomID)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Deleted successfully",
		},
		{
			name:           "Invalid ID",
			roomID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid room ID",
		},
		{
			name:   "Service Error",
			roomID: "1",
			mockFunc: func(ctx context.Context, roomID int) error {
				return errors.New("delete failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "delete failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockRoomService{
				DeleteRoomFunc: tt.mockFunc,
			}

			h := setupRoomHandler(svc)
			r := setupRoomRouter(h)

			req := httptest.NewRequest(http.MethodDelete, "/rooms/"+tt.roomID, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

