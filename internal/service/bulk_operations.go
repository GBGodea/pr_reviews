package service

import (
	"context"
	"fmt"
	"sync"

	"pr-reviewer-service/internal/models"
)

func (s *PRService) BulkDeactivateTeamUsers(ctx context.Context, teamName string, userIDs []string) error {
	userRepo := s.userRepo
	prRepo := s.prRepo

	userIDMap := make(map[string]bool)
	for _, id := range userIDs {
		userIDMap[id] = true
	}

	members, err := userRepo.GetTeamMembers(ctx, teamName)
	if err != nil {
		return err
	}

	var candidates []models.User
	for _, member := range members {
		if !userIDMap[member.UserID] && member.IsActive {
			candidates = append(candidates, member)
		}
	}

	if len(candidates) == 0 {
		return fmt.Errorf("no active candidates for replacement")
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(userIDs))
	sem := make(chan struct{}, 3)

	for _, userID := range userIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			prs, err := prRepo.GetUserReviews(ctx, id)
			if err != nil {
				errChan <- fmt.Errorf("failed to get user reviews for %s: %w", id, err)
				return
			}

			for _, pr := range prs {
				fullPR, err := prRepo.GetPR(ctx, pr.ID)
				if err != nil || fullPR.Status != "OPEN" {
					continue
				}

				var replacement *models.User
				for _, cand := range candidates {
					isAssigned := false
					for _, rev := range fullPR.AssignedReviewers {
						if rev == cand.UserID {
							isAssigned = true
							break
						}
					}
					if !isAssigned {
						replacement = &cand
						break
					}
				}

				if replacement == nil {
				}

				if err := prRepo.RemoveReviewer(ctx, pr.ID, id); err != nil {
					errChan <- err
					return
				}

				if err := prRepo.AssignReviewers(ctx, pr.ID, []string{replacement.UserID}); err != nil {
					errChan <- err
					return
				}
			}
		}(userID)
	}

	wg.Wait()
	close(errChan)

	var lastErr error
	for err := range errChan {
		if err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		return lastErr
	}

	for _, userID := range userIDs {
		if _, err := userRepo.SetUserActive(ctx, userID, false); err != nil {
			return err
		}
	}

	return nil
}