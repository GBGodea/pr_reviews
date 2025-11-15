package repository

import (
	"context"
	"errors"
	"fmt"

	"pr-reviewer-service/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PRRepository struct {
	pool *pgxpool.Pool
}

func NewPRRepository(pool *pgxpool.Pool) *PRRepository {
	return &PRRepository{pool: pool}
}

func (r *PRRepository) CreatePR(ctx context.Context, pr *models.PullRequest) error {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)", pr.ID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("PR already exists")
	}

	_, err = r.pool.Exec(ctx,
		"INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status) VALUES ($1, $2, $3, $4)",
		pr.ID, pr.Name, pr.AuthorID, "OPEN")
	
	return err
}

func (r *PRRepository) AssignReviewers(ctx context.Context, prID string, reviewerIDs []string) error {
	for _, reviewerID := range reviewerIDs {
		_, err := r.pool.Exec(ctx,
			"INSERT INTO pr_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)",
			prID, reviewerID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PRRepository) GetPR(ctx context.Context, prID string) (*models.PullRequest, error) {
	pr := &models.PullRequest{}

	err := r.pool.QueryRow(ctx,
		"SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at FROM pull_requests WHERE pull_request_id = $1",
		prID).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("PR not found")
		}
		return nil, err
	}

	rows, err := r.pool.Query(ctx,
		"SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id = $1 ORDER BY reviewer_id",
		prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pr.AssignedReviewers = []string{}
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	return pr, rows.Err()
}

func (r *PRRepository) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {
	pr, err := r.GetPR(ctx, prID)
	if err != nil {
		return nil, err
	}

	if pr.Status == "MERGED" {
		return pr, nil
	}

	_, err = r.pool.Exec(ctx,
		"UPDATE pull_requests SET status = 'MERGED', merged_at = CURRENT_TIMESTAMP WHERE pull_request_id = $1",
		prID)

	if err != nil {
		return nil, err
	}

	return r.GetPR(ctx, prID)
}

func (r *PRRepository) GetUserReviews(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status 
		 FROM pull_requests pr
		 INNER JOIN pr_reviewers rev ON pr.pull_request_id = rev.pull_request_id
		 WHERE rev.reviewer_id = $1
		 ORDER BY pr.pull_request_id`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []models.PullRequestShort
	for rows.Next() {
		var pr models.PullRequestShort
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	return prs, rows.Err()
}

func (r *PRRepository) RemoveReviewer(ctx context.Context, prID string, reviewerID string) error {
	_, err := r.pool.Exec(ctx,
		"DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2",
		prID, reviewerID)
	return err
}

func (r *PRRepository) IsReviewerAssigned(ctx context.Context, prID string, reviewerID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2)",
		prID, reviewerID).Scan(&exists)
	return exists, err
}