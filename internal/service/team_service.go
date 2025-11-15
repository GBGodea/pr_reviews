package service

import (
	"context"

	"pr-reviewer-service/internal/models"
	"pr-reviewer-service/internal/repository"
)

type TeamService struct {
	teamRepo *repository.TeamRepository
	userRepo *repository.UserRepository
}

func NewTeamService(teamRepo *repository.TeamRepository, userRepo *repository.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, team *models.Team) (*models.Team, error) {
	if err := s.teamRepo.CreateTeam(ctx, team); err != nil {
		return nil, err
	}

	for i := range team.Members {
		user := &models.User{
			UserID:   team.Members[i].UserID,
			Username: team.Members[i].Username,
			TeamName: team.TeamName,
			IsActive: team.Members[i].IsActive,
		}
		if err := s.userRepo.CreateOrUpdateUser(ctx, user); err != nil {
			return nil, err
		}
	}

	return team, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	return s.teamRepo.GetTeam(ctx, teamName)
}