package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock ShoppingRepository
type mockShoppingRepo struct {
	// Categories
	CreateCategoryFunc   func(ctx context.Context, c *models.ShoppingCategory) error
	FindAllCategoriesFunc func(ctx context.Context, homeID int) (*[]models.ShoppingCategory, error)
	FindCategoryByIDFunc func(ctx context.Context, id int) (*models.ShoppingCategory, error)
	DeleteCategoryFunc   func(ctx context.Context, id int) error
	EditCategoryFunc     func(ctx context.Context, category *models.ShoppingCategory, updates map[string]interface{}) error

	// Items
	CreateItemFunc           func(ctx context.Context, i *models.ShoppingItem) error
	FindItemsByCategoryIDFunc func(ctx context.Context, categoryID int) ([]models.ShoppingItem, error)
	FindItemByIDFunc         func(ctx context.Context, id int) (*models.ShoppingItem, error)
	DeleteItemFunc           func(ctx context.Context, id int) error
	MarkIsBoughtFunc         func(ctx context.Context, id int) error
	EditItemFunc             func(ctx context.Context, item *models.ShoppingItem, updates map[string]interface{}) error
}

func (m *mockShoppingRepo) CreateCategory(ctx context.Context, c *models.ShoppingCategory) error {
	if m.CreateCategoryFunc != nil {
		return m.CreateCategoryFunc(ctx, c)
	}
	return nil
}

func (m *mockShoppingRepo) FindAllCategories(ctx context.Context, homeID int) (*[]models.ShoppingCategory, error) {
	if m.FindAllCategoriesFunc != nil {
		return m.FindAllCategoriesFunc(ctx, homeID)
	}
	return &[]models.ShoppingCategory{}, nil
}

func (m *mockShoppingRepo) FindCategoryByID(ctx context.Context, id int) (*models.ShoppingCategory, error) {
	if m.FindCategoryByIDFunc != nil {
		return m.FindCategoryByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockShoppingRepo) DeleteCategory(ctx context.Context, id int) error {
	if m.DeleteCategoryFunc != nil {
		return m.DeleteCategoryFunc(ctx, id)
	}
	return nil
}

func (m *mockShoppingRepo) EditCategory(ctx context.Context, category *models.ShoppingCategory, updates map[string]interface{}) error {
	if m.EditCategoryFunc != nil {
		return m.EditCategoryFunc(ctx, category, updates)
	}
	return nil
}

func (m *mockShoppingRepo) CreateItem(ctx context.Context, i *models.ShoppingItem) error {
	if m.CreateItemFunc != nil {
		return m.CreateItemFunc(ctx, i)
	}
	return nil
}

func (m *mockShoppingRepo) FindItemsByCategoryID(ctx context.Context, categoryID int) ([]models.ShoppingItem, error) {
	if m.FindItemsByCategoryIDFunc != nil {
		return m.FindItemsByCategoryIDFunc(ctx, categoryID)
	}
	return []models.ShoppingItem{}, nil
}

func (m *mockShoppingRepo) FindItemByID(ctx context.Context, id int) (*models.ShoppingItem, error) {
	if m.FindItemByIDFunc != nil {
		return m.FindItemByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockShoppingRepo) DeleteItem(ctx context.Context, id int) error {
	if m.DeleteItemFunc != nil {
		return m.DeleteItemFunc(ctx, id)
	}
	return nil
}

func (m *mockShoppingRepo) MarkIsBought(ctx context.Context, id int) error {
	if m.MarkIsBoughtFunc != nil {
		return m.MarkIsBoughtFunc(ctx, id)
	}
	return nil
}

func (m *mockShoppingRepo) EditItem(ctx context.Context, item *models.ShoppingItem, updates map[string]interface{}) error {
	if m.EditItemFunc != nil {
		return m.EditItemFunc(ctx, item, updates)
	}
	return nil
}

// Test helpers
func setupShoppingService(t *testing.T, repo repository.ShoppingRepository) *services.ShoppingService {
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	return services.NewShoppingService(repo, redisClient)
}

// CreateCategory Tests
func TestShoppingService_CreateCategory_Success(t *testing.T) {
	icon := "🛒"

	repo := &mockShoppingRepo{
		CreateCategoryFunc: func(ctx context.Context, c *models.ShoppingCategory) error {
			require.Equal(t, "Groceries", c.Name)
			require.Equal(t, "red", c.Color)
			require.Equal(t, 1, c.HomeID)
			return nil
		},
	}

	svc := setupShoppingService(t, repo)
	err := svc.CreateCategory(context.Background(), "Groceries", &icon, "red", 1)

	assert.NoError(t, err)
}

func TestShoppingService_CreateCategory_RepositoryError(t *testing.T) {
	repo := &mockShoppingRepo{
		CreateCategoryFunc: func(ctx context.Context, c *models.ShoppingCategory) error {
			return errors.New("database error")
		},
	}

	svc := setupShoppingService(t, repo)
	err := svc.CreateCategory(context.Background(), "Groceries", nil, "red", 1)

	assert.Error(t, err)
}

// FindAllCategoriesForHome Tests
func TestShoppingService_FindAllCategoriesForHome_Success(t *testing.T) {
	expectedCategories := &[]models.ShoppingCategory{
		{ID: 1, Name: "Groceries", HomeID: 1},
		{ID: 2, Name: "Electronics", HomeID: 1},
	}

	repo := &mockShoppingRepo{
		FindAllCategoriesFunc: func(ctx context.Context, homeID int) (*[]models.ShoppingCategory, error) {
			require.Equal(t, 1, homeID)
			return expectedCategories, nil
		},
	}

	svc := setupShoppingService(t, repo)
	categories, err := svc.FindAllCategoriesForHome(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, *categories, 2)
	assert.Equal(t, "Groceries", (*categories)[0].Name)
}

func TestShoppingService_FindAllCategoriesForHome_Empty(t *testing.T) {
	repo := &mockShoppingRepo{
		FindAllCategoriesFunc: func(ctx context.Context, homeID int) (*[]models.ShoppingCategory, error) {
			return &[]models.ShoppingCategory{}, nil
		},
	}

	svc := setupShoppingService(t, repo)
	categories, err := svc.FindAllCategoriesForHome(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, *categories, 0)
}

// FindCategoryByID Tests
func TestShoppingService_FindCategoryByID_Success(t *testing.T) {
	expectedCategory := &models.ShoppingCategory{
		ID:     1,
		Name:   "Groceries",
		HomeID: 1,
	}

	repo := &mockShoppingRepo{
		FindCategoryByIDFunc: func(ctx context.Context, id int) (*models.ShoppingCategory, error) {
			require.Equal(t, 1, id)
			return expectedCategory, nil
		},
	}

	svc := setupShoppingService(t, repo)
	category, err := svc.FindCategoryByID(context.Background(), 1, 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedCategory.Name, category.Name)
}

func TestShoppingService_FindCategoryByID_NotFound(t *testing.T) {
	repo := &mockShoppingRepo{
		FindCategoryByIDFunc: func(ctx context.Context, id int) (*models.ShoppingCategory, error) {
			return nil, errors.New("category not found")
		},
	}

	svc := setupShoppingService(t, repo)
	_, err := svc.FindCategoryByID(context.Background(), 999, 1)

	assert.Error(t, err)
}

// DeleteCategory Tests
func TestShoppingService_DeleteCategory_Success(t *testing.T) {
	repo := &mockShoppingRepo{
		DeleteCategoryFunc: func(ctx context.Context, id int) error {
			require.Equal(t, 1, id)
			return nil
		},
	}

	svc := setupShoppingService(t, repo)
	err := svc.DeleteCategory(context.Background(), 1, 1)

	assert.NoError(t, err)
}

// EditCategory Tests
func TestShoppingService_EditCategory_Success(t *testing.T) {
	newName := "Updated Groceries"
	newIcon := "🛍️"
	newColor := "blue"

	category := &models.ShoppingCategory{ID: 1, Name: "Groceries", HomeID: 1}

	repo := &mockShoppingRepo{
		FindCategoryByIDFunc: func(ctx context.Context, id int) (*models.ShoppingCategory, error) {
			return category, nil
		},
		EditCategoryFunc: func(ctx context.Context, cat *models.ShoppingCategory, updates map[string]interface{}) error {
			require.Equal(t, category, cat)
			require.Equal(t, "Updated Groceries", updates["name"])
			require.Equal(t, "🛍️", updates["icon"])
			require.Equal(t, "blue", updates["color"])
			return nil
		},
	}

	svc := setupShoppingService(t, repo)
	err := svc.EditCategory(context.Background(), 1, 1, &newName, &newIcon, &newColor)

	assert.NoError(t, err)
}

// CreateItem Tests
func TestShoppingService_CreateItem_Success(t *testing.T) {
	image := "http://image.url"
	link := "http://product.link"

	repo := &mockShoppingRepo{
		CreateItemFunc: func(ctx context.Context, i *models.ShoppingItem) error {
			require.Equal(t, "Milk", i.Name)
			require.Equal(t, 1, i.CategoryID)
			require.Equal(t, 5, i.UploadedBy)
			require.False(t, i.IsBought)
			return nil
		},
	}

	svc := setupShoppingService(t, repo)
	err := svc.CreateItem(context.Background(), 1, 5, "Milk", &image, &link)

	assert.NoError(t, err)
}

func TestShoppingService_CreateItem_RepositoryError(t *testing.T) {
	repo := &mockShoppingRepo{
		CreateItemFunc: func(ctx context.Context, i *models.ShoppingItem) error {
			return errors.New("database error")
		},
	}

	svc := setupShoppingService(t, repo)
	err := svc.CreateItem(context.Background(), 1, 5, "Milk", nil, nil)

	assert.Error(t, err)
}

// FindItemsByCategoryID Tests
func TestShoppingService_FindItemsByCategoryID_Success(t *testing.T) {
	expectedItems := []models.ShoppingItem{
		{ID: 1, Name: "Milk", CategoryID: 1},
		{ID: 2, Name: "Bread", CategoryID: 1},
	}

	repo := &mockShoppingRepo{
		FindItemsByCategoryIDFunc: func(ctx context.Context, categoryID int) ([]models.ShoppingItem, error) {
			require.Equal(t, 1, categoryID)
			return expectedItems, nil
		},
	}

	svc := setupShoppingService(t, repo)
	items, err := svc.FindItemsByCategoryID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "Milk", items[0].Name)
}

// FindItemByID Tests
func TestShoppingService_FindItemByID_Success(t *testing.T) {
	expectedItem := &models.ShoppingItem{
		ID:         1,
		Name:       "Milk",
		CategoryID: 1,
	}

	repo := &mockShoppingRepo{
		FindItemByIDFunc: func(ctx context.Context, id int) (*models.ShoppingItem, error) {
			require.Equal(t, 1, id)
			return expectedItem, nil
		},
	}

	svc := setupShoppingService(t, repo)
	item, err := svc.FindItemByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedItem.Name, item.Name)
}

// DeleteItem Tests
func TestShoppingService_DeleteItem_Success(t *testing.T) {
	repo := &mockShoppingRepo{
		DeleteItemFunc: func(ctx context.Context, id int) error {
			require.Equal(t, 1, id)
			return nil
		},
	}

	svc := setupShoppingService(t, repo)
	err := svc.DeleteItem(context.Background(), 1)

	assert.NoError(t, err)
}

// MarkIsBought Tests
func TestShoppingService_MarkIsBought_Success(t *testing.T) {
	repo := &mockShoppingRepo{
		MarkIsBoughtFunc: func(ctx context.Context, id int) error {
			require.Equal(t, 1, id)
			return nil
		},
	}

	svc := setupShoppingService(t, repo)
	err := svc.MarkIsBought(context.Background(), 1)

	assert.NoError(t, err)
}
