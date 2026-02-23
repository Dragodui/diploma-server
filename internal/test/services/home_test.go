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

// Mock HomeRepository
type mockHomeRepo struct {
	CreateFunc                   func(ctx context.Context, h *models.Home) error
	RegenerateCodeFunc           func(ctx context.Context, code string, id int) error
	FindByIDFunc                 func(ctx context.Context, id int) (*models.Home, error)
	FindByInviteCodeFunc         func(ctx context.Context, inviteCode string) (*models.Home, error)
	DeleteFunc                   func(ctx context.Context, id int) error
	IsAdminFunc                  func(ctx context.Context, id int, userID int) (bool, error)
	AddMemberFunc                func(ctx context.Context, id int, userID int, role string) error
	IsMemberFunc                 func(ctx context.Context, id int, userID int) (bool, error)
	DeleteMemberFunc             func(ctx context.Context, id int, userID int) error
	GenerateUniqueInviteCodeFunc func(ctx context.Context) (string, error)
	GetUserHomeFunc              func(ctx context.Context, userID int) (*models.Home, error)
}

func (m *mockHomeRepo) Create(ctx context.Context, h *models.Home) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, h)
	}
	return nil
}

func (m *mockHomeRepo) RegenerateCode(ctx context.Context, code string, id int) error {
	if m.RegenerateCodeFunc != nil {
		return m.RegenerateCodeFunc(ctx, code, id)
	}
	return nil
}

func (m *mockHomeRepo) FindByID(ctx context.Context, id int) (*models.Home, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockHomeRepo) FindByInviteCode(ctx context.Context, inviteCode string) (*models.Home, error) {
	if m.FindByInviteCodeFunc != nil {
		return m.FindByInviteCodeFunc(ctx, inviteCode)
	}
	return nil, nil
}

func (m *mockHomeRepo) Delete(ctx context.Context, id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *mockHomeRepo) IsAdmin(ctx context.Context, id int, userID int) (bool, error) {
	if m.IsAdminFunc != nil {
		return m.IsAdminFunc(ctx, id, userID)
	}
	return false, nil
}

func (m *mockHomeRepo) AddMember(ctx context.Context, id int, userID int, role string) error {
	if m.AddMemberFunc != nil {
		return m.AddMemberFunc(ctx, id, userID, role)
	}
	return nil
}

func (m *mockHomeRepo) IsMember(ctx context.Context, id int, userID int) (bool, error) {
	if m.IsMemberFunc != nil {
		return m.IsMemberFunc(ctx, id, userID)
	}
	return false, nil
}

func (m *mockHomeRepo) DeleteMember(ctx context.Context, id int, userID int) error {
	if m.DeleteMemberFunc != nil {
		return m.DeleteMemberFunc(ctx, id, userID)
	}
	return nil
}

func (m *mockHomeRepo) GenerateUniqueInviteCode(ctx context.Context) (string, error) {
	if m.GenerateUniqueInviteCodeFunc != nil {
		return m.GenerateUniqueInviteCodeFunc(ctx)
	}
	return "", nil
}

func (m *mockHomeRepo) GetUserHome(ctx context.Context, userID int) (*models.Home, error) {
	if m.GetUserHomeFunc != nil {
		return m.GetUserHomeFunc(ctx, userID)
	}
	return nil, nil
}

// Test helpers
func setupHomeService(t *testing.T, repo repository.HomeRepository) *services.HomeService {
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	return services.NewHomeService(repo, redisClient)
}

// CreateHome Tests
func TestHomeService_CreateHome_Success(t *testing.T) {
	repo := &mockHomeRepo{
		GenerateUniqueInviteCodeFunc: func(ctx context.Context) (string, error) {
			return "ABC123", nil
		},
		CreateFunc: func(ctx context.Context, h *models.Home) error {
			require.Equal(t, "My Home", h.Name)
			require.Equal(t, "ABC123", h.InviteCode)
			h.ID = 1
			return nil
		},
		AddMemberFunc: func(ctx context.Context, id int, userID int, role string) error {
			require.Equal(t, 1, id)
			require.Equal(t, 5, userID)
			require.Equal(t, "admin", role)
			return nil
		},
	}

	svc := setupHomeService(t, repo)
	err := svc.CreateHome(context.Background(), "My Home", 5)

	assert.NoError(t, err)
}

func TestHomeService_CreateHome_InviteCodeGenerationFails(t *testing.T) {
	repo := &mockHomeRepo{
		GenerateUniqueInviteCodeFunc: func(ctx context.Context) (string, error) {
			return "", errors.New("failed to generate code")
		},
	}

	svc := setupHomeService(t, repo)
	err := svc.CreateHome(context.Background(), "My Home", 5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate code")
}

func TestHomeService_CreateHome_RepositoryError(t *testing.T) {
	repo := &mockHomeRepo{
		GenerateUniqueInviteCodeFunc: func(ctx context.Context) (string, error) {
			return "ABC123", nil
		},
		CreateFunc: func(ctx context.Context, h *models.Home) error {
			return errors.New("database error")
		},
	}

	svc := setupHomeService(t, repo)
	err := svc.CreateHome(context.Background(), "My Home", 5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

// JoinHome Tests
func TestHomeService_JoinHome_Success(t *testing.T) {
	home := &models.Home{
		ID:         1,
		Name:       "Test Home",
		InviteCode: "ABC123",
	}

	repo := &mockHomeRepo{
		FindByInviteCodeFunc: func(ctx context.Context, inviteCode string) (*models.Home, error) {
			require.Equal(t, "ABC123", inviteCode)
			return home, nil
		},
		AddMemberFunc: func(ctx context.Context, id int, userID int, role string) error {
			require.Equal(t, 1, id)
			require.Equal(t, 5, userID)
			require.Equal(t, "member", role)
			return nil
		},
	}

	svc := setupHomeService(t, repo)
	err := svc.JoinHomeByCode(context.Background(), "ABC123", 5)

	assert.NoError(t, err)
}

func TestHomeService_JoinHome_InvalidInviteCode(t *testing.T) {
	repo := &mockHomeRepo{
		FindByInviteCodeFunc: func(ctx context.Context, inviteCode string) (*models.Home, error) {
			return nil, errors.New("home not found")
		},
	}

	svc := setupHomeService(t, repo)
	err := svc.JoinHomeByCode(context.Background(), "INVALID", 5)

	assert.Error(t, err)
}

// GetHomeByID Tests
func TestHomeService_GetHomeByID_Success(t *testing.T) {
	expectedHome := &models.Home{
		ID:         1,
		Name:       "Test Home",
		InviteCode: "ABC123",
	}

	repo := &mockHomeRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*models.Home, error) {
			require.Equal(t, 1, id)
			return expectedHome, nil
		},
	}

	svc := setupHomeService(t, repo)
	home, err := svc.GetHomeByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedHome.ID, home.ID)
	assert.Equal(t, expectedHome.Name, home.Name)
}

func TestHomeService_GetHomeByID_NotFound(t *testing.T) {
	repo := &mockHomeRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*models.Home, error) {
			return nil, errors.New("home not found")
		},
	}

	svc := setupHomeService(t, repo)
	_, err := svc.GetHomeByID(context.Background(), 999)

	assert.Error(t, err)
}

// GetUserHome Tests
func TestHomeService_GetUserHome_Success(t *testing.T) {
	expectedHome := &models.Home{
		ID:   1,
		Name: "User's Home",
	}

	repo := &mockHomeRepo{
		GetUserHomeFunc: func(ctx context.Context, userID int) (*models.Home, error) {
			require.Equal(t, 5, userID)
			return expectedHome, nil
		},
	}

	svc := setupHomeService(t, repo)
	home, err := svc.GetUserHome(context.Background(), 5)

	assert.NoError(t, err)
	assert.Equal(t, expectedHome.ID, home.ID)
	assert.Equal(t, expectedHome.Name, home.Name)
}

func TestHomeService_GetUserHome_NotFound(t *testing.T) {
	repo := &mockHomeRepo{
		GetUserHomeFunc: func(ctx context.Context, userID int) (*models.Home, error) {
			return nil, errors.New("user has no home")
		},
	}

	svc := setupHomeService(t, repo)
	_, err := svc.GetUserHome(context.Background(), 5)

	assert.Error(t, err)
}

// DeleteHome Tests
func TestHomeService_DeleteHome_Success(t *testing.T) {
	repo := &mockHomeRepo{
		DeleteFunc: func(ctx context.Context, id int) error {
			require.Equal(t, 1, id)
			return nil
		},
	}

	svc := setupHomeService(t, repo)
	err := svc.DeleteHome(context.Background(), 1)

	assert.NoError(t, err)
}

func TestHomeService_DeleteHome_NotFound(t *testing.T) {
	repo := &mockHomeRepo{
		DeleteFunc: func(ctx context.Context, id int) error {
			return errors.New("home not found")
		},
	}

	svc := setupHomeService(t, repo)
	err := svc.DeleteHome(context.Background(), 999)

	assert.Error(t, err)
}

// LeaveHome Tests
func TestHomeService_LeaveHome_Success(t *testing.T) {
	repo := &mockHomeRepo{
		DeleteMemberFunc: func(ctx context.Context, id int, userID int) error {
			require.Equal(t, 1, id)
			require.Equal(t, 5, userID)
			return nil
		},
	}

	svc := setupHomeService(t, repo)
	err := svc.LeaveHome(context.Background(), 1, 5)

	assert.NoError(t, err)
}

// RemoveMember Tests
func TestHomeService_RemoveMember_Success(t *testing.T) {
	repo := &mockHomeRepo{
		DeleteMemberFunc: func(ctx context.Context, id int, userID int) error {
			require.Equal(t, 1, id)
			require.Equal(t, 10, userID)
			return nil
		},
	}

	svc := setupHomeService(t, repo)
	err := svc.RemoveMember(context.Background(), 1, 10, 1)

	assert.NoError(t, err)
}

// RegenerateInviteCode Tests
func TestHomeService_RegenerateInviteCode_Success(t *testing.T) {
	repo := &mockHomeRepo{
		GenerateUniqueInviteCodeFunc: func(ctx context.Context) (string, error) {
			return "XYZ789", nil
		},
		RegenerateCodeFunc: func(ctx context.Context, code string, id int) error {
			require.Equal(t, "XYZ789", code)
			require.Equal(t, 1, id)
			return nil
		},
	}

	svc := setupHomeService(t, repo)
	err := svc.RegenerateInviteCode(context.Background(), 1)

	assert.NoError(t, err)
}

func TestHomeService_RegenerateInviteCode_CodeGenerationFails(t *testing.T) {
	repo := &mockHomeRepo{
		GenerateUniqueInviteCodeFunc: func(ctx context.Context) (string, error) {
			return "", errors.New("failed to generate code")
		},
	}

	svc := setupHomeService(t, repo)
	err := svc.RegenerateInviteCode(context.Background(), 1)

	assert.Error(t, err)
}
