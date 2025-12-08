package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// Mock notification service
type mockNotificationService struct {
	CreateFunc                       func(from *int, to int, description string) error
	GetByUserIDFunc                  func(userID int) ([]models.Notification, error)
	MarkAsReadFunc                   func(notificationID, userID int) error
	CreateHomeNotificationFunc       func(from *int, homeID int, description string) error
	GetByHomeIDFunc                  func(homeID int) ([]models.HomeNotification, error)
	MarkAsReadForHomeNotificationFunc func(notificationID, homeID int) error
}

func (m *mockNotificationService) Create(from *int, to int, description string) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(from, to, description)
	}
	return nil
}

func (m *mockNotificationService) GetByUserID(userID int) ([]models.Notification, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(userID)
	}
	return nil, nil
}

func (m *mockNotificationService) MarkAsRead(notificationID, userID int) error {
	if m.MarkAsReadFunc != nil {
		return m.MarkAsReadFunc(notificationID, userID)
	}
	return nil
}

func (m *mockNotificationService) CreateHomeNotification(from *int, homeID int, description string) error {
	if m.CreateHomeNotificationFunc != nil {
		return m.CreateHomeNotificationFunc(from, homeID, description)
	}
	return nil
}

func (m *mockNotificationService) GetByHomeID(homeID int) ([]models.HomeNotification, error) {
	if m.GetByHomeIDFunc != nil {
		return m.GetByHomeIDFunc(homeID)
	}
	return nil, nil
}

func (m *mockNotificationService) MarkAsReadForHomeNotification(notificationID, homeID int) error {
	if m.MarkAsReadForHomeNotificationFunc != nil {
		return m.MarkAsReadForHomeNotificationFunc(notificationID, homeID)
	}
	return nil
}

func setupNotificationHandler(svc *mockNotificationService) *handlers.NotificationHandler {
	return handlers.NewNotificationHandler(svc)
}

func setupNotificationRouter(h *handlers.NotificationHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/homes/notifications", h.GetByUserID)
	r.Delete("/homes/notifications/{notification_id}", h.MarkAsRead)
	r.Get("/homes/{home_id}/notifications", h.GetByHomeID)
	r.Delete("/homes/{home_id}/notifications/{notification_id}", h.MarkAsReadForHome)
	return r
}

func TestNotificationHandler_GetByUserID(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		mockFunc       func(userID int) ([]models.Notification, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userID: 123,
			mockFunc: func(userID int) ([]models.Notification, error) {
				require.Equal(t, 123, userID)
				return []models.Notification{
					{ID: 1, To: 123, Description: "Test notification"},
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Test notification",
		},
		{
			name:   "Empty List",
			userID: 123,
			mockFunc: func(userID int) ([]models.Notification, error) {
				return []models.Notification{}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "notifications",
		},
		{
			name:   "Service Error",
			userID: 123,
			mockFunc: func(userID int) ([]models.Notification, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockNotificationService{
				GetByUserIDFunc: tt.mockFunc,
			}

			h := setupNotificationHandler(svc)
			r := setupNotificationRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/homes/notifications", nil)
			req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestNotificationHandler_MarkAsRead(t *testing.T) {
	tests := []struct {
		name           string
		notificationID string
		userID         int
		mockFunc       func(notificationID, userID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			notificationID: "1",
			userID:         123,
			mockFunc: func(notificationID, userID int) error {
				require.Equal(t, 1, notificationID)
				require.Equal(t, 123, userID)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Marked as read",
		},
		{
			name:           "Invalid ID",
			notificationID: "invalid",
			userID:         123,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid notification ID",
		},
		{
			name:           "Service Error",
			notificationID: "1",
			userID:         123,
			mockFunc: func(notificationID, userID int) error {
				return errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockNotificationService{
				MarkAsReadFunc: tt.mockFunc,
			}

			h := setupNotificationHandler(svc)
			r := setupNotificationRouter(h)

			req := httptest.NewRequest(http.MethodDelete, "/homes/notifications/"+tt.notificationID, nil)
			req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestNotificationHandler_GetByHomeID(t *testing.T) {
	tests := []struct {
		name           string
		homeID         string
		mockFunc       func(homeID int) ([]models.HomeNotification, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			homeID: "1",
			mockFunc: func(homeID int) ([]models.HomeNotification, error) {
				require.Equal(t, 1, homeID)
				return []models.HomeNotification{
					{ID: 1, HomeID: 1, Description: "Home notification"},
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Home notification",
		},
		{
			name:           "Invalid Home ID",
			homeID:         "invalid",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid home ID",
		},
		{
			name:   "Service Error",
			homeID: "1",
			mockFunc: func(homeID int) ([]models.HomeNotification, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockNotificationService{
				GetByHomeIDFunc: tt.mockFunc,
			}

			h := setupNotificationHandler(svc)
			r := setupNotificationRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/homes/"+tt.homeID+"/notifications", nil)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestNotificationHandler_MarkAsReadForHome(t *testing.T) {
	tests := []struct {
		name           string
		homeID         string
		notificationID string
		userID         int
		mockFunc       func(notificationID, userID int) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			homeID:         "1",
			notificationID: "1",
			userID:         123,
			mockFunc: func(notificationID, userID int) error {
				require.Equal(t, 1, notificationID)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Marked as read",
		},
		{
			name:           "Invalid Notification ID",
			homeID:         "1",
			notificationID: "invalid",
			userID:         123,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid notification ID",
		},
		{
			name:           "Service Error",
			homeID:         "1",
			notificationID: "1",
			userID:         123,
			mockFunc: func(notificationID, userID int) error {
				return errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockNotificationService{
				MarkAsReadForHomeNotificationFunc: tt.mockFunc,
			}

			h := setupNotificationHandler(svc)
			r := setupNotificationRouter(h)

			req := httptest.NewRequest(http.MethodDelete, "/homes/"+tt.homeID+"/notifications/"+tt.notificationID, nil)
			req = req.WithContext(utils.WithUserID(req.Context(), tt.userID))

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
		})
	}
}
