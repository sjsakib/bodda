package database

import (
	"context"
	"fmt"

	"bodda/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (strava_id, access_token, refresh_token, token_expiry, first_name, last_name)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		user.StravaID,
		user.AccessToken,
		user.RefreshToken,
		user.TokenExpiry,
		user.FirstName,
		user.LastName,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, strava_id, access_token, refresh_token, token_expiry, 
		       first_name, last_name, created_at, updated_at
		FROM users WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.StravaID,
		&user.AccessToken,
		&user.RefreshToken,
		&user.TokenExpiry,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByStravaID(ctx context.Context, stravaID int64) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, strava_id, access_token, refresh_token, token_expiry, 
		       first_name, last_name, created_at, updated_at
		FROM users WHERE strava_id = $1`

	err := r.db.QueryRow(ctx, query, stravaID).Scan(
		&user.ID,
		&user.StravaID,
		&user.AccessToken,
		&user.RefreshToken,
		&user.TokenExpiry,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET access_token = $2, refresh_token = $3, token_expiry = $4, 
		    first_name = $5, last_name = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRow(ctx, query,
		user.ID,
		user.AccessToken,
		user.RefreshToken,
		user.TokenExpiry,
		user.FirstName,
		user.LastName,
	).Scan(&user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}