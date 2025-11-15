package repository

import (
	"context"
	"errors"

	"pr-reviewer-service/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) CreateOrUpdateUser(ctx context.Context, user *models.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (user_id, username, team_name, is_active) 
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id) DO UPDATE 
		 SET username = $2, team_name = $3, is_active = $4`,
		user.UserID, user.Username, user.TeamName, user.IsActive)
	return err
}

func (r *UserRepository) GetUser(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(ctx,
		"SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1",
		userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetTeamMembers(ctx context.Context, teamName string) ([]models.User, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT user_id, username, team_name, is_active FROM users WHERE team_name = $1 ORDER BY user_id",
		teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (r *UserRepository) SetUserActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	user := &models.User{}
	err := r.pool.QueryRow(ctx,
		"UPDATE users SET is_active = $1 WHERE user_id = $2 RETURNING user_id, username, team_name, is_active",
		isActive, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}