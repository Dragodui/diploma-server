package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test fixtures
var (
	validCategory = &models.ShoppingCategory{
		ID:     1,
		Name:   "Groceries",
		Icon:   stringPtr("🛒"),
		Color:  "#ffffff",
		HomeID: 1,
	}

	validItem = &models.ShoppingItem{
		ID:         1,
		CategoryID: 1,
		Name:       "Milk",
		Image:      stringPtr("milk.jpg"),
		Link:       stringPtr("http://example.com"),
		UploadedBy: 123,
		IsBought:   false,
	}

	validCreateCategoryRequest = models.CreateCategoryRequest{
		Name:  "Groceries",
		Icon:  stringPtr("🛒"),
		Color: "#ffffff",
	}

	validCreateItemRequest = models.CreateShoppingItemRequest{
		CategoryID: 1,
		Name:       "Milk",
		Image:      stringPtr("milk.jpg"),
		Link:       stringPtr("http://example.com"),
	}

	validUpdateCategoryRequest = models.UpdateShoppingCategoryRequest{
		Name:  stringPtr("Updated Category"),
		Icon:  stringPtr("🆕"),
		Color: stringPtr("#000000"),
	}

	validUpdateItemRequest = models.UpdateShoppingItemRequest{
		Name:     stringPtr("Updated Item"),
		Image:    stringPtr("updated.jpg"),
		Link:     stringPtr("http://updated.com"),
		IsBought: boolPtr(true),
		BoughtAt: timePtr(time.Now()),
	}
)

// Helper functions
func stringPtr(s string) *string     { return &s }
func boolPtr(b bool) *bool           { return &b }
func timePtr(t time.Time) *time.Time { return &t }

func setupShoppingHandler(mockSvc *mockShoppingService) *handlers.ShoppingHandler {
	return handlers.NewShoppingHandler(mockSvc)
}

func makeJSONRequest(method, url string, body interface{}) *http.Request {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, url, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func setupShoppingRouter(h *handlers.ShoppingHandler) *chi.Mux {
	r := chi.NewRouter()
	// Categories
	r.Post("/homes/{home_id}/categories", h.CreateCategory)
	r.Get("/homes/{home_id}/categories", h.GetAllCategories)
	r.Get("/homes/{home_id}/categories/{category_id}", h.GetCategoryByID)
	r.Put("/homes/{home_id}/categories/{category_id}", h.EditCategory)
	r.Delete("/homes/{home_id}/categories/{category_id}", h.DeleteCategory)

	// Items
	r.Post("/items", h.CreateItem)
	r.Get("/items/{item_id}", h.GetItemByID)
	r.Get("/categories/{category_id}/items", h.GetItemsByCategoryID)
	r.Put("/items/{item_id}", h.EditItem)
	r.Delete("/items/{item_id}", h.DeleteItem)
	r.Put("/items/{item_id}/mark-bought", h.MarkIsBought)

	return r
}

func assertJSONResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int, shouldContain string) {
	t.Helper()
	assert.Equal(t, expectedStatus, rr.Code)
	if shouldContain != "" {
		assert.Contains(t, rr.Body.String(), shouldContain)
	}
}

func assertJSONEqual(t *testing.T, rr *httptest.ResponseRecorder, expected interface{}) {
	t.Helper()
	expectedJSON, _ := json.Marshal(expected)
	assert.JSONEq(t, string(expectedJSON), rr.Body.String())
}

// Mock service
type mockShoppingService struct {
	// Categories
	CreateCategoryFunc           func(ctx context.Context, name string, icon *string, color string, homeID int) error
	FindAllCategoriesForHomeFunc func(ctx context.Context, homeID int) (*[]models.ShoppingCategory, error)
	FindCategoryByIDFunc         func(ctx context.Context, categoryID, homeID int) (*models.ShoppingCategory, error)
	DeleteCategoryFunc           func(ctx context.Context, categoryID, homeID int) error
	EditCategoryFunc             func(ctx context.Context, categoryID, homeID int, name, icon, color *string) error

	// Items
	CreateItemFunc            func(ctx context.Context, categoryID, userID int, name string, image, link *string) error
	FindItemByIDFunc          func(ctx context.Context, itemID int) (*models.ShoppingItem, error)
	FindItemsByCategoryIDFunc func(ctx context.Context, categoryID int) ([]models.ShoppingItem, error)
	DeleteItemFunc            func(ctx context.Context, itemID int) error
	MarkIsBoughtFunc          func(ctx context.Context, itemID int) error
	EditItemFunc              func(ctx context.Context, itemID int, name, image, link *string, isBought *bool, boughtAt *time.Time) error
}

// Category methods
func (m *mockShoppingService) CreateCategory(ctx context.Context, name string, icon *string, color string, homeID int) error {
	if m.CreateCategoryFunc != nil {
		return m.CreateCategoryFunc(ctx, name, icon, color, homeID)
	}
	return nil
}

func (m *mockShoppingService) FindAllCategoriesForHome(ctx context.Context, homeID int) (*[]models.ShoppingCategory, error) {
	if m.FindAllCategoriesForHomeFunc != nil {
		return m.FindAllCategoriesForHomeFunc(ctx, homeID)
	}
	return nil, nil
}

func (m *mockShoppingService) FindCategoryByID(ctx context.Context, categoryID, homeID int) (*models.ShoppingCategory, error) {
	if m.FindCategoryByIDFunc != nil {
		return m.FindCategoryByIDFunc(ctx, categoryID, homeID)
	}
	return nil, nil
}

func (m *mockShoppingService) DeleteCategory(ctx context.Context, categoryID, homeID int) error {
	if m.DeleteCategoryFunc != nil {
		return m.DeleteCategoryFunc(ctx, categoryID, homeID)
	}
	return nil
}

func (m *mockShoppingService) EditCategory(ctx context.Context, categoryID, homeID int, name, icon, color *string) error {
	if m.EditCategoryFunc != nil {
		return m.EditCategoryFunc(ctx, categoryID, homeID, name, icon, color)
	}
	return nil
}

// Item methods
func (m *mockShoppingService) CreateItem(ctx context.Context, categoryID, userID int, name string, image, link *string) error {
	if m.CreateItemFunc != nil {
		return m.CreateItemFunc(ctx, categoryID, userID, name, image, link)
	}
	return nil
}

func (m *mockShoppingService) FindItemByID(ctx context.Context, itemID int) (*models.ShoppingItem, error) {
	if m.FindItemByIDFunc != nil {
		return m.FindItemByIDFunc(ctx, itemID)
	}
	return nil, nil
}

func (m *mockShoppingService) FindItemsByCategoryID(ctx context.Context, categoryID int) ([]models.ShoppingItem, error) {
	if m.FindItemsByCategoryIDFunc != nil {
		return m.FindItemsByCategoryIDFunc(ctx, categoryID)
	}
	return nil, nil
}

func (m *mockShoppingService) DeleteItem(ctx context.Context, itemID int) error {
	if m.DeleteItemFunc != nil {
		return m.DeleteItemFunc(ctx, itemID)
	}
	return nil
}

func (m *mockShoppingService) MarkIsBought(ctx context.Context, itemID int) error {
	if m.MarkIsBoughtFunc != nil {
		return m.MarkIsBoughtFunc(ctx, itemID)
	}
	return nil
}

func (m *mockShoppingService) EditItem(ctx context.Context, itemID int, name, image, link *string, isBought *bool, boughtAt *time.Time) error {
	if m.EditItemFunc != nil {
		return m.EditItemFunc(ctx, itemID, name, image, link, isBought, boughtAt)
	}
	return nil
}

// CATEGORY TESTS
func TestShoppingHandler_Categories(t *testing.T) {
	t.Run("CreateCategory", func(t *testing.T) {
		tests := []struct {
			name           string
			homeID         string
			body           interface{}
			mockFunc       func(ctx context.Context, name string, icon *string, color string, homeID int) error
			expectedStatus int
			expectedBody   string
		}{
			{
				name:   "Success",
				homeID: "1",
				body:   validCreateCategoryRequest,
				mockFunc: func(ctx context.Context, name string, icon *string, color string, homeID int) error {
					assert.Equal(t, "Groceries", name)
					assert.Equal(t, "🛒", *icon)
					assert.Equal(t, "#ffffff", color)
					assert.Equal(t, 1, homeID)
					return nil
				},
				expectedStatus: http.StatusCreated,
				expectedBody:   "Created successfully",
			},
			{
				name:           "Invalid Home ID",
				homeID:         "invalid",
				body:           validCreateCategoryRequest,
				mockFunc:       nil,
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "invalid home ID",
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
				name:   "Service Error",
				homeID: "1",
				body:   validCreateCategoryRequest,
				mockFunc: func(ctx context.Context, name string, icon *string, color string, homeID int) error {
					return errors.New("service error")
				},
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "Invalid data",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := &mockShoppingService{
					CreateCategoryFunc: tt.mockFunc,
				}

				h := setupShoppingHandler(svc)
				r := setupShoppingRouter(h)

				var req *http.Request
				if tt.name == "Invalid JSON" {
					req = httptest.NewRequest(http.MethodPost, "/homes/"+tt.homeID+"/categories",
						bytes.NewBufferString("{bad json}"))
				} else {
					req = makeJSONRequest(http.MethodPost, "/homes/"+tt.homeID+"/categories", tt.body)
				}

				rr := httptest.NewRecorder()
				r.ServeHTTP(rr, req)

				assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
			})
		}
	})

	t.Run("GetAllCategories", func(t *testing.T) {
		tests := []struct {
			name           string
			homeID         string
			mockFunc       func(ctx context.Context, homeID int) (*[]models.ShoppingCategory, error)
			expectedStatus int
			expectedBody   string
		}{
			{
				name:   "Success",
				homeID: "1",
				mockFunc: func(ctx context.Context, homeID int) (*[]models.ShoppingCategory, error) {
					require.Equal(t, 1, homeID)
					categories := []models.ShoppingCategory{*validCategory}
					return &categories, nil
				},
				expectedStatus: http.StatusOK,
				expectedBody:   "Groceries",
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
				mockFunc: func(ctx context.Context, homeID int) (*[]models.ShoppingCategory, error) {
					return nil, errors.New("service error")
				},
				expectedStatus: http.StatusInternalServerError,
				expectedBody:   "service error",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := &mockShoppingService{
					FindAllCategoriesForHomeFunc: tt.mockFunc,
				}

				h := setupShoppingHandler(svc)
				r := setupShoppingRouter(h)

				req := httptest.NewRequest(http.MethodGet, "/homes/"+tt.homeID+"/categories", nil)
				rr := httptest.NewRecorder()

				r.ServeHTTP(rr, req)

				assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
			})
		}
	})

	t.Run("GetCategoryByID", func(t *testing.T) {
		tests := []struct {
			name           string
			categoryID     string
			homeID         string
			mockFunc       func(ctx context.Context, categoryID, homeID int) (*models.ShoppingCategory, error)
			expectedStatus int
			expectedBody   string
		}{
			{
				name:       "Success",
				categoryID: "1",
				homeID:     "1",
				mockFunc: func(ctx context.Context, categoryID, homeID int) (*models.ShoppingCategory, error) {
					require.Equal(t, 1, categoryID)
					return validCategory, nil
				},
				expectedStatus: http.StatusOK,
				expectedBody:   "Groceries",
			},
			{
				name:           "Invalid Category ID",
				categoryID:     "invalid",
				homeID:         "1",
				mockFunc:       nil,
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "invalid category ID",
			},
			{
				name:           "Invalid Home ID",
				categoryID:     "1",
				homeID:         "invalid",
				mockFunc:       nil,
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "invalid home ID",
			},
			{
				name:       "Service Error",
				categoryID: "1",
				homeID:     "1",
				mockFunc: func(ctx context.Context, categoryID, homeID int) (*models.ShoppingCategory, error) {
					return nil, errors.New("service error")
				},
				expectedStatus: http.StatusInternalServerError,
				expectedBody:   "service error",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := &mockShoppingService{
					FindCategoryByIDFunc: tt.mockFunc,
				}

				h := setupShoppingHandler(svc)
				r := setupShoppingRouter(h)

				req := httptest.NewRequest(http.MethodGet,
					"/homes/"+tt.homeID+"/categories/"+tt.categoryID, nil)
				rr := httptest.NewRecorder()

				r.ServeHTTP(rr, req)

				assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
			})
		}
	})

	t.Run("EditCategory", func(t *testing.T) {
		tests := []struct {
			name           string
			categoryID     string
			homeID         string
			body           interface{}
			mockFunc       func(ctx context.Context, categoryID, homeID int, name, icon, color *string) error
			expectedStatus int
			expectedBody   string
		}{
			{
				name:       "Success",
				categoryID: "1",
				homeID:     "1",
				body:       validUpdateCategoryRequest,
				mockFunc: func(ctx context.Context, categoryID, homeID int, name, icon, color *string) error {
					assert.Equal(t, 1, categoryID)
					assert.Equal(t, 1, homeID)
					assert.Equal(t, "Updated Category", *name)
					assert.Equal(t, "#000000", *color)
					assert.Equal(t, "🆕", *icon)
					return nil
				},
				expectedStatus: http.StatusOK,
				expectedBody:   "Updated successfully",
			},
			{
				name:           "Invalid Category ID",
				categoryID:     "invalid",
				homeID:         "1",
				body:           validUpdateCategoryRequest,
				mockFunc:       nil,
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "invalid category ID",
			},
			{
				name:           "Invalid Home ID",
				categoryID:     "1",
				homeID:         "invalid",
				body:           validUpdateCategoryRequest,
				mockFunc:       nil,
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "invalid home ID",
			},
			{
				name:           "Invalid JSON",
				categoryID:     "1",
				homeID:         "1",
				body:           "{bad json}",
				mockFunc:       nil,
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "Invalid data",
			},
			{
				name:       "Service Error",
				categoryID: "1",
				homeID:     "1",
				body:       validUpdateCategoryRequest,
				mockFunc: func(ctx context.Context, categoryID, homeID int, name, icon, color *string) error {
					return errors.New("edit failed")
				},
				expectedStatus: http.StatusInternalServerError,
				expectedBody:   "edit failed",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := &mockShoppingService{
					EditCategoryFunc: tt.mockFunc,
				}

				h := setupShoppingHandler(svc)
				r := setupShoppingRouter(h)

				var req *http.Request
				if tt.name == "Invalid JSON" {
					req = httptest.NewRequest(http.MethodPut,
						"/homes/"+tt.homeID+"/categories/"+tt.categoryID,
						bytes.NewBufferString("{bad json}"))
				} else {
					req = makeJSONRequest(http.MethodPut,
						"/homes/"+tt.homeID+"/categories/"+tt.categoryID, tt.body)
				}

				rr := httptest.NewRecorder()
				r.ServeHTTP(rr, req)

				assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
			})
		}
	})

	t.Run("DeleteCategory", func(t *testing.T) {
		tests := []struct {
			name           string
			categoryID     string
			homeID         string
			mockFunc       func(ctx context.Context, categoryID, homeID int) error
			expectedStatus int
			expectedBody   string
		}{
			{
				name:       "Success",
				categoryID: "1",
				homeID:     "1",
				mockFunc: func(ctx context.Context, categoryID, homeID int) error {
					assert.Equal(t, 1, categoryID)
					assert.Equal(t, 1, homeID)
					return nil
				},
				expectedStatus: http.StatusOK,
				expectedBody:   "Deleted successfully",
			},
			{
				name:           "Invalid Category ID",
				categoryID:     "invalid",
				homeID:         "1",
				mockFunc:       nil,
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "invalid category ID",
			},
			{
				name:           "Invalid Home ID",
				categoryID:     "1",
				homeID:         "invalid",
				mockFunc:       nil,
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "invalid home ID",
			},
			{
				name:       "Service Error",
				categoryID: "1",
				homeID:     "1",
				mockFunc: func(ctx context.Context, categoryID, homeID int) error {
					return errors.New("delete failed")
				},
				expectedStatus: http.StatusInternalServerError,
				expectedBody:   "delete failed",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := &mockShoppingService{
					DeleteCategoryFunc: tt.mockFunc,
				}

				h := setupShoppingHandler(svc)
				r := setupShoppingRouter(h)

				req := httptest.NewRequest(http.MethodDelete,
					"/homes/"+tt.homeID+"/categories/"+tt.categoryID, nil)
				rr := httptest.NewRecorder()

				r.ServeHTTP(rr, req)

				assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
			})
		}
	})
}

// ITEM TESTS
func TestShoppingHandler_Items(t *testing.T) {
	t.Run("CreateItem", func(t *testing.T) {
		tests := []struct {
			name           string
			body           interface{}
			mockFunc       func(ctx context.Context, categoryID, userID int, name string, image, link *string) error
			expectedStatus int
			expectedBody   string
		}{
			{
				name: "Success",
				body: validCreateItemRequest,
				mockFunc: func(ctx context.Context, categoryID, userID int, name string, image, link *string) error {
					assert.Equal(t, 1, categoryID)
					assert.Equal(t, "Milk", name)
					assert.Equal(t, "milk.jpg", *image)
					assert.Equal(t, "http://example.com", *link)
					return nil
				},
				expectedStatus: http.StatusOK,
				expectedBody:   "Created successfully",
			},
			{
				name:           "Invalid JSON",
				body:           "{bad json}",
				mockFunc:       nil,
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "Invalid data",
			},
			{
				name: "Service Error",
				body: validCreateItemRequest,
				mockFunc: func(ctx context.Context, categoryID, userID int, name string, image, link *string) error {
					return errors.New("service error")
				},
				expectedStatus: http.StatusInternalServerError,
				expectedBody:   "service error",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := &mockShoppingService{
					CreateItemFunc: tt.mockFunc,
				}

				h := setupShoppingHandler(svc)

				var req *http.Request
				if tt.name == "Invalid JSON" {
					req = httptest.NewRequest(http.MethodPost, "/items",
						bytes.NewBufferString("{bad json}"))
				} else {
					req = makeJSONRequest(http.MethodPost, "/items", tt.body)
				}

				rr := httptest.NewRecorder()
				h.CreateItem(rr, req)

				assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
			})
		}
	})

	t.Run("GetItemByID", func(t *testing.T) {
		tests := []struct {
			name           string
			itemID         string
			mockFunc       func(ctx context.Context, itemID int) (*models.ShoppingItem, error)
			expectedStatus int
			expectedBody   string
		}{
			{
				name:   "Success",
				itemID: "1",
				mockFunc: func(ctx context.Context, itemID int) (*models.ShoppingItem, error) {
					require.Equal(t, 1, itemID)
					return validItem, nil
				},
				expectedStatus: http.StatusOK,
				expectedBody:   "Milk",
			},
			{
				name:           "Invalid Item ID",
				itemID:         "invalid",
				mockFunc:       nil,
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "invalid item ID",
			},
			{
				name:   "Service Error",
				itemID: "1",
				mockFunc: func(ctx context.Context, itemID int) (*models.ShoppingItem, error) {
					return nil, errors.New("service error")
				},
				expectedStatus: http.StatusInternalServerError,
				expectedBody:   "service error",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := &mockShoppingService{
					FindItemByIDFunc: tt.mockFunc,
				}

				h := setupShoppingHandler(svc)
				r := setupShoppingRouter(h)

				req := httptest.NewRequest(http.MethodGet, "/items/"+tt.itemID, nil)
				rr := httptest.NewRecorder()

				r.ServeHTTP(rr, req)

				assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
			})
		}
	})

	t.Run("GetItemsByCategoryID", func(t *testing.T) {
		tests := []struct {
			name           string
			categoryID     string
			mockFunc       func(ctx context.Context, categoryID int) ([]models.ShoppingItem, error)
			expectedStatus int
			expectedBody   string
		}{
			{
				name:       "Success",
				categoryID: "1",
				mockFunc: func(ctx context.Context, categoryID int) ([]models.ShoppingItem, error) {
					require.Equal(t, 1, categoryID)
					return []models.ShoppingItem{*validItem}, nil
				},
				expectedStatus: http.StatusOK,
				expectedBody:   "Milk",
			},
			{
				name:           "Invalid Category ID",
				categoryID:     "invalid",
				mockFunc:       nil,
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "invalid category ID",
			},
			{
				name:       "Service Error",
				categoryID: "1",
				mockFunc: func(ctx context.Context, categoryID int) ([]models.ShoppingItem, error) {
					return nil, errors.New("service error")
				},
				expectedStatus: http.StatusInternalServerError,
				expectedBody:   "service error",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := &mockShoppingService{
					FindItemsByCategoryIDFunc: tt.mockFunc,
				}

				h := setupShoppingHandler(svc)
				r := setupShoppingRouter(h)

				req := httptest.NewRequest(http.MethodGet, "/categories/"+tt.categoryID+"/items", nil)
				rr := httptest.NewRecorder()

				r.ServeHTTP(rr, req)

				assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
			})
		}
	})

	t.Run("EditItem", func(t *testing.T) {
		tests := []struct {
			name           string
			itemID         string
			body           interface{}
			mockFunc       func(ctx context.Context, itemID int, name, image, link *string, isBought *bool, boughtAt *time.Time) error
			expectedStatus int
			expectedBody   string
		}{
			{
				name:   "Success",
				itemID: "1",
				body:   validUpdateItemRequest,
				mockFunc: func(ctx context.Context, itemID int, name, image, link *string, isBought *bool, boughtAt *time.Time) error {
					assert.Equal(t, 1, itemID)
					assert.Equal(t, "Updated Item", *name)
					assert.Equal(t, "updated.jpg", *image)
					assert.True(t, *isBought)
					return nil
				},
				expectedStatus: http.StatusOK,
				expectedBody:   "Edited successfully",
			},
			{
				name:           "Invalid Item ID",
				itemID:         "invalid",
				body:           validUpdateItemRequest,
				mockFunc:       nil,
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "invalid item ID",
			},
			{
				name:           "Invalid JSON",
				itemID:         "1",
				body:           "{bad json}",
				mockFunc:       nil,
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "invalid data",
			},
			{
				name:   "Service Error",
				itemID: "1",
				body:   validUpdateItemRequest,
				mockFunc: func(ctx context.Context, itemID int, name, image, link *string, isBought *bool, boughtAt *time.Time) error {
					return errors.New("edit failed")
				},
				expectedStatus: http.StatusInternalServerError,
				expectedBody:   "edit failed",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := &mockShoppingService{
					EditItemFunc: tt.mockFunc,
				}

				h := setupShoppingHandler(svc)
				r := setupShoppingRouter(h)

				var req *http.Request
				if tt.name == "Invalid JSON" {
					req = httptest.NewRequest(http.MethodPut, "/items/"+tt.itemID,
						bytes.NewBufferString("{bad json}"))
				} else {
					req = makeJSONRequest(http.MethodPut, "/items/"+tt.itemID, tt.body)
				}

				rr := httptest.NewRecorder()
				r.ServeHTTP(rr, req)

				assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
			})
		}
	})

	t.Run("DeleteItem", func(t *testing.T) {
		tests := []struct {
			name           string
			itemID         string
			mockFunc       func(ctx context.Context, itemID int) error
			expectedStatus int
			expectedBody   string
		}{
			{
				name:   "Success",
				itemID: "1",
				mockFunc: func(ctx context.Context, itemID int) error {
					assert.Equal(t, 1, itemID)
					return nil
				},
				expectedStatus: http.StatusOK,
				expectedBody:   "Deleted successfully",
			},
			{
				name:           "Invalid Item ID",
				itemID:         "invalid",
				mockFunc:       nil,
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "invalid item ID",
			},
			{
				name:   "Service Error",
				itemID: "1",
				mockFunc: func(ctx context.Context, itemID int) error {
					return errors.New("delete failed")
				},
				expectedStatus: http.StatusInternalServerError,
				expectedBody:   "delete failed",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := &mockShoppingService{
					DeleteItemFunc: tt.mockFunc,
				}

				h := setupShoppingHandler(svc)
				r := setupShoppingRouter(h)

				req := httptest.NewRequest(http.MethodDelete, "/items/"+tt.itemID, nil)
				rr := httptest.NewRecorder()

				r.ServeHTTP(rr, req)

				assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
			})
		}
	})

	t.Run("MarkIsBought", func(t *testing.T) {
		tests := []struct {
			name           string
			itemID         string
			mockFunc       func(ctx context.Context, itemID int) error
			expectedStatus int
			expectedBody   string
		}{
			{
				name:   "Success",
				itemID: "1",
				mockFunc: func(ctx context.Context, itemID int) error {
					assert.Equal(t, 1, itemID)
					return nil
				},
				expectedStatus: http.StatusOK,
				expectedBody:   "Marked successfully",
			},
			{
				name:           "Invalid Item ID",
				itemID:         "invalid",
				mockFunc:       nil,
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "invalid item ID",
			},
			{
				name:   "Service Error",
				itemID: "1",
				mockFunc: func(ctx context.Context, itemID int) error {
					return errors.New("mark failed")
				},
				expectedStatus: http.StatusInternalServerError,
				expectedBody:   "mark failed",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := &mockShoppingService{
					MarkIsBoughtFunc: tt.mockFunc,
				}

				h := setupShoppingHandler(svc)
				r := setupShoppingRouter(h)

				req := httptest.NewRequest(http.MethodPut, "/items/"+tt.itemID+"/mark-bought", nil)
				rr := httptest.NewRecorder()

				r.ServeHTTP(rr, req)

				assertJSONResponse(t, rr, tt.expectedStatus, tt.expectedBody)
			})
		}
	})

}

