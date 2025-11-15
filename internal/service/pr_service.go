package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"pr-reviewer-service/internal/models"
	"pr-reviewer-service/internal/repository"
)

type PRService struct {
	prRepo   *repository.PRRepository
	userRepo *repository.UserRepository
	teamRepo *repository.TeamRepository
}

func NewPRService(prRepo *repository.PRRepository, userRepo *repository.UserRepository, teamRepo *repository.TeamRepository) *PRService {
	return &PRService{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}

func (s *PRService) CreatePR(ctx context.Context, prID string, prName string, authorID string) (*models.PullRequest, error) {
	author, err := s.userRepo.GetUser(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("author not found")
	}

	pr := &models.PullRequest{
		ID:                prID,
		Name:              prName,
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: []string{},
	}

	if err := s.prRepo.CreatePR(ctx, pr); err != nil {
		return nil, err
	}

	members, err := s.userRepo.GetTeamMembers(ctx, author.TeamName)
	if err != nil {
		return nil, err
	}

	var candidates []models.User
	for _, member := range members {
		if member.IsActive && member.UserID != authorID {
			candidates = append(candidates, member)
		}
	}

	numReviewers := 2
	if len(candidates) < 2 {
		numReviewers = len(candidates)
	}

	reviewerIDs := make([]string, 0, numReviewers)
	if numReviewers > 0 {
		for i := len(candidates) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			candidates[i], candidates[j] = candidates[j], candidates[i]
		}

		for i := 0; i < numReviewers; i++ {
			reviewerIDs = append(reviewerIDs, candidates[i].UserID)
		}
	}

	if len(reviewerIDs) > 0 {
		if err := s.prRepo.AssignReviewers(ctx, prID, reviewerIDs); err != nil {
			return nil, err
		}
	}

	pr.AssignedReviewers = reviewerIDs
	return pr, nil
}

func (s *PRService) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {
	return s.prRepo.MergePR(ctx, prID)
}

func (s *PRService) ReassignReviewer(ctx context.Context, prID string, oldReviewerID string) (*models.PullRequest, string, error) {
	pr, err := s.prRepo.GetPR(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	if pr.Status == "MERGED" {
		return nil, "", errors.New("PR is merged")
	}

	isAssigned, err := s.prRepo.IsReviewerAssigned(ctx, prID, oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	if !isAssigned {
		return nil, "", errors.New("reviewer is not assigned")
	}

	oldReviewer, err := s.userRepo.GetUser(ctx, oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	members, err := s.userRepo.GetTeamMembers(ctx, oldReviewer.TeamName)
	if err != nil {
		return nil, "", err
	}

	var candidates []models.User
	for _, member := range members {
		if member.IsActive && member.UserID != oldReviewerID {
			isDuplicate := false
			for _, rev := range pr.AssignedReviewers {
				if rev == member.UserID {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				candidates = append(candidates, member)
			}
		}
	}

	if len(candidates) == 0 {
		return nil, "", errors.New("no active replacement candidate in team")
	}

	newReviewerID := candidates[rand.Intn(len(candidates))].UserID

	if err := s.prRepo.RemoveReviewer(ctx, prID, oldReviewerID); err != nil {
		return nil, "", err
	}

	if err := s.prRepo.AssignReviewers(ctx, prID, []string{newReviewerID}); err != nil {
		return nil, "", err
	}

	updatedPR, err := s.prRepo.GetPR(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	return updatedPR, newReviewerID, nil
}

func (s *PRService) GetUserReviews(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	return s.prRepo.GetUserReviews(ctx, userID)
}