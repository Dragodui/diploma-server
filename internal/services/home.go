package services

import (
	"context"
	"errors"

	"github.com/Dragodui/diploma-server/internal/event"
	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/metrics"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/redis/go-redis/v9"
)

type HomeService struct {
	repo     repository.HomeRepository
	cache    *redis.Client
	notifSvc INotificationService
}

type IHomeService interface {
	CreateHome(ctx context.Context, name string, userID int) error
	RegenerateInviteCode(ctx context.Context, homeID int) error
	JoinHomeByCode(ctx context.Context, code string, userID int) error
	GetHomeByID(ctx context.Context, id int) (*models.Home, error)
	DeleteHome(ctx context.Context, id int) error
	LeaveHome(ctx context.Context, homeID int, userID int) error
	RemoveMember(ctx context.Context, homeID int, userID int, currentUserID int) error
	GetMembers(ctx context.Context, homeID int) ([]models.HomeMembership, error)
	GetUserHome(ctx context.Context, userID int) (*models.Home, error)
	GetUserHomes(ctx context.Context, userID int) ([]models.Home, error)
	ApproveMember(ctx context.Context, homeID int, userID int) error
	RejectMember(ctx context.Context, homeID int, userID int) error
	GetPendingMembers(ctx context.Context, homeID int) ([]models.HomeMembership, error)
	UpdateMemberRole(ctx context.Context, homeID int, userID int, role string) error
}

func NewHomeService(repo repository.HomeRepository, cache *redis.Client, notifSvc INotificationService) *HomeService {
	return &HomeService{repo: repo, cache: cache, notifSvc: notifSvc}
}

func (s *HomeService) CreateHome(ctx context.Context, name string, userID int) error {
	inviteCode, err := s.repo.GenerateUniqueInviteCode(ctx)
	if err != nil {
		return err
	}

	home := &models.Home{
		Name:       name,
		InviteCode: inviteCode,
	}

	if err := s.repo.Create(ctx, home); err != nil {
		return err
	}

	if err := s.repo.AddMember(ctx, home.ID, userID, "admin", "approved"); err != nil {
		return err
	}

	// invalidate user homes list cache
	if err := utils.DeleteFromCache(ctx, utils.GetUserHomesKey(userID), s.cache); err != nil {
		logger.Info.Printf("Failed to delete user homes cache: %v", err)
	}

	metrics.HomesTotal.Inc()
	metrics.HomeOperationsTotal.WithLabelValues("create").Inc()

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleHome,
		Action: event.ActionCreated,
		Data:   home,
	})

	return nil
}

func (s *HomeService) RegenerateInviteCode(ctx context.Context, homeID int) error {
	inviteCode, err := s.repo.GenerateUniqueInviteCode(ctx)
	if err != nil {
		return err
	}

	key := utils.GetHomeCacheKey(homeID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	if err := s.repo.RegenerateCode(ctx, inviteCode, homeID); err != nil {
		return err
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleHome,
		Action: event.ActionUpdated,
		Data:   map[string]int{"homeID": homeID},
	})

	return nil
}

func (s *HomeService) JoinHomeByCode(ctx context.Context, code string, userID int) error {
	home, err := s.repo.FindByInviteCode(ctx, code)
	if err != nil || home == nil {
		return errors.New("invalid invite code")
	}

	already, err := s.repo.IsMember(ctx, home.ID, userID)
	if err != nil {
		return err
	}
	if already {
		return errors.New("user already in this home")
	}

	pending, err := s.repo.IsPendingMember(ctx, home.ID, userID)
	if err != nil {
		return err
	}
	if pending {
		return errors.New("join request already pending")
	}

	key := utils.GetHomeCacheKey(home.ID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	if err := s.repo.AddMember(ctx, home.ID, userID, "member", "pending"); err != nil {
		return err
	}

	metrics.HomeOperationsTotal.WithLabelValues("join_request").Inc()

	// Notify home that a user has requested to join
	fromID := userID
	_ = s.notifSvc.CreateHomeNotification(ctx, &fromID, home.ID, "A user has requested to join the home")

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleHome,
		Action: event.ActionMemberJoined,
		Data:   map[string]int{"homeID": home.ID, "userID": userID},
	})

	return nil
}

func (s *HomeService) GetHomeByID(ctx context.Context, id int) (*models.Home, error) {
	key := utils.GetHomeCacheKey(id)
	// if in cache => returns from cache
	cached, err := utils.GetFromCache[models.Home](ctx, key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	// if not in cache => returns from db
	home, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if home == nil {
		return nil, errors.New("home not found")
	}

	// saves to cache
	if err := utils.WriteToCache(ctx, key, home, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return home, nil
}

func (s *HomeService) DeleteHome(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	metrics.HomesTotal.Dec()
	metrics.HomeOperationsTotal.WithLabelValues("delete").Inc()

	// remove from cache
	key := utils.GetHomeCacheKey(id)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleHome,
		Action: event.ActionDeleted,
		Data:   map[string]int{"id": id},
	})

	return nil
}

func (s *HomeService) LeaveHome(ctx context.Context, homeID int, userID int) error {
	key := utils.GetHomeCacheKey(homeID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	if err := s.repo.DeleteMember(ctx, homeID, userID); err != nil {
		return err
	}

	metrics.HomeOperationsTotal.WithLabelValues("leave").Inc()

	// invalidate user homes list cache
	if err := utils.DeleteFromCache(ctx, utils.GetUserHomesKey(userID), s.cache); err != nil {
		logger.Info.Printf("Failed to delete user homes cache: %v", err)
	}

	// Notify home that a member left
	fromID := userID
	_ = s.notifSvc.CreateHomeNotification(ctx, &fromID, homeID, "A member has left the home")

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleHome,
		Action: event.ActionMemberLeft,
		Data:   map[string]int{"homeID": homeID, "userID": userID},
	})

	return nil
}

func (s *HomeService) RemoveMember(ctx context.Context, homeID int, userID int, currentUserID int) error {
	if userID == currentUserID {
		return errors.New("you cannot remove yourself")
	}

	key := utils.GetHomeCacheKey(homeID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	if err := s.repo.DeleteMember(ctx, homeID, userID); err != nil {
		return err
	}

	metrics.HomeOperationsTotal.WithLabelValues("remove_member").Inc()

	// invalidate removed user's homes list cache
	if err := utils.DeleteFromCache(ctx, utils.GetUserHomesKey(userID), s.cache); err != nil {
		logger.Info.Printf("Failed to delete user homes cache: %v", err)
	}

	// Notify the removed user
	fromID := currentUserID
	_ = s.notifSvc.Create(ctx, &fromID, userID, "You have been removed from a home")
	// Notify home that a member was removed
	_ = s.notifSvc.CreateHomeNotification(ctx, &fromID, homeID, "A member has been removed from the home")

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleHome,
		Action: event.ActionMemberRemoved,
		Data:   map[string]int{"homeID": homeID, "userID": userID},
	})

	return nil
}

func (s *HomeService) GetMembers(ctx context.Context, homeID int) ([]models.HomeMembership, error) {
	return s.repo.GetMembers(ctx, homeID)
}

func (s *HomeService) GetUserHome(ctx context.Context, userID int) (*models.Home, error) {
	key := utils.GetUserHomeKey(userID)
	cached, err := utils.GetFromCache[models.Home](ctx, key, s.cache)
	if cached != nil && err == nil {
		return cached, nil
	}

	home, err := s.repo.GetUserHome(ctx, userID)
	if err != nil {
		return nil, err
	}
	if home == nil {
		return nil, errors.New("home not found")
	}

	if err := utils.WriteToCache(ctx, key, home, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	return home, nil
}

func (s *HomeService) GetUserHomes(ctx context.Context, userID int) ([]models.Home, error) {
	key := utils.GetUserHomesKey(userID)
	cached, err := utils.GetFromCache[[]models.Home](ctx, key, s.cache)
	if cached != nil && err == nil {
		return *cached, nil
	}

	homes, err := s.repo.GetUserHomes(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := utils.WriteToCache(ctx, key, homes, s.cache); err != nil {
		logger.Info.Printf("Failed to write to cache [%s]: %v", key, err)
	}

	return homes, nil
}

func (s *HomeService) ApproveMember(ctx context.Context, homeID int, userID int) error {
	if err := s.repo.ApproveMember(ctx, homeID, userID); err != nil {
		return err
	}

	key := utils.GetHomeCacheKey(homeID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	if err := utils.DeleteFromCache(ctx, utils.GetUserHomesKey(userID), s.cache); err != nil {
		logger.Info.Printf("Failed to delete user homes cache: %v", err)
	}

	if err := utils.DeleteFromCache(ctx, utils.GetUserHomeKey(userID), s.cache); err != nil {
		logger.Info.Printf("Failed to delete user home cache: %v", err)
	}

	metrics.HomeOperationsTotal.WithLabelValues("approve_member").Inc()

	_ = s.notifSvc.Create(ctx, nil, userID, "Your request to join the home has been approved")
	_ = s.notifSvc.CreateHomeNotification(ctx, nil, homeID, "A new member has been approved")

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleHome,
		Action: event.ActionMemberJoined,
		Data:   map[string]int{"homeID": homeID, "userID": userID},
	})

	return nil
}

func (s *HomeService) RejectMember(ctx context.Context, homeID int, userID int) error {
	if err := s.repo.RejectMember(ctx, homeID, userID); err != nil {
		return err
	}

	key := utils.GetHomeCacheKey(homeID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	metrics.HomeOperationsTotal.WithLabelValues("reject_member").Inc()

	_ = s.notifSvc.Create(ctx, nil, userID, "Your request to join the home has been rejected")

	return nil
}

func (s *HomeService) GetPendingMembers(ctx context.Context, homeID int) ([]models.HomeMembership, error) {
	return s.repo.GetPendingMembers(ctx, homeID)
}

func (s *HomeService) UpdateMemberRole(ctx context.Context, homeID int, userID int, role string) error {
	if err := s.repo.UpdateMemberRole(ctx, homeID, userID, role); err != nil {
		return err
	}

	key := utils.GetHomeCacheKey(homeID)
	if err := utils.DeleteFromCache(ctx, key, s.cache); err != nil {
		logger.Info.Printf("Failed to delete redis cache for key %s: %v", key, err)
	}

	metrics.HomeOperationsTotal.WithLabelValues("update_role").Inc()

	_ = s.notifSvc.Create(ctx, nil, userID, "Your role has been updated to "+role)

	event.SendEvent(ctx, s.cache, "updates", &event.RealTimeEvent{
		Module: event.ModuleHome,
		Action: event.ActionUpdated,
		Data:   map[string]interface{}{"homeID": homeID, "userID": userID, "role": role},
	})

	return nil
}

