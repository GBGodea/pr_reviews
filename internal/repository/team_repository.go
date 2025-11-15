package repository

import (
	"context"
	"errors"
	"fmt"

	"pr-reviewer-service/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepository struct {
	pool *pgxpool.Pool
}

func NewTeamRepository(pool *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{pool: pool}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, team *models.Team) error {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", team.TeamName).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("team already exists")
	}

	_, err = r.pool.Exec(ctx, "INSERT INTO teams (team_name) VALUES ($1)", team.TeamName)
	return err
}

func (r *TeamRepository) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", teamName).Scan(&exists)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("team not found")
	}

	rows, err := r.pool.Query(ctx, 
		"SELECT user_id, username, is_active FROM users WHERE team_name = $1 ORDER BY user_id", 
		teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	team := &models.Team{
		TeamName: teamName,
		Members:  []models.TeamMember{},
	}

	for rows.Next() {
		var member models.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, err
		}
		team.Members = append(team.Members, member)
	}

	return team, rows.Err()
}