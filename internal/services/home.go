package services

import (
	"errors"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
)

type HomeService struct {
	homes repository.HomeRepository
}

func NewHomeService(repo repository.HomeRepository) *HomeService {
	return &HomeService{homes: repo}
}

func (s *HomeService) CreateHome(name string) error {
	inviteCode, err := s.homes.GenerateUniqueInviteCode()
	if err != nil {
		return err
	}

	return s.homes.Create(&models.Home{
		Name:       name,
		InviteCode: inviteCode,
	})
}

func (s *HomeService) JoinHomeByCode(code string, userID int) error {
	home, err := s.homes.FindByInviteCode(code)
	if err != nil {
		return errors.New("invalid invite code")
	}

	already, err := s.homes.IsMember(home.ID, userID)
	if err != nil {
		return err
	}
	if already {
		return errors.New("user already in this home")
	}

	return s.homes.AddMember(home.ID, userID, "member")
}

func (s *HomeService) GetHomeByID(id int) (*models.Home, error) {
	home, err := s.homes.FindByID(id)
	if err != nil {
		return nil, err
	}
	return home, nil
}

func (s *HomeService) DeleteHome(id int) error {
	if err := s.homes.Delete(id); err != nil {
		return err
	}
	return nil
}

func (s *HomeService) LeaveHome(homeID int, userID int) error {
	return s.homes.DeleteMember(homeID, userID)
}

func (s *HomeService) RemoveMember(homeID int, userID int, currentUserID int) error {
	if userID == currentUserID {
		return errors.New("you cannot remove yourself")
	}
	return s.homes.DeleteMember(homeID, userID)
}
