package service

import (
	"context"

	"pr-reviewer-service/internal/models"
	"pr-reviewer-service/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	return s.userRepo.SetUserActive(ctx, userID, isActive)
}

func (s *UserService) GetUserReviews(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	return []models.PullRequestShort{}, nil
}
