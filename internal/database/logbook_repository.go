package database

import (
	"context"
	"fmt"

	"bodda/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LogbookRepository struct {
	db *pgxpool.Pool
}

// Ensure LogbookRepository implements LogbookRepositoryInterface
var _ LogbookRepositoryInterface = (*LogbookRepository)(nil)

func NewLogbookRepository(db *pgxpool.Pool) *LogbookRepository {
	return &LogbookRepository{db: db}
}

func (r *LogbookRepository) Create(ctx context.Context, logbook *models.AthleteLogbook) error {
	query := `
		INSERT INTO athlete_logbooks (user_id, content)
		VALUES ($1, $2)
		RETURNING id, updated_at`

	err := r.db.QueryRow(ctx, query,
		logbook.UserID,
		logbook.Content,
	).Scan(&logbook.ID, &logbook.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create logbook: %w", err)
	}

	return nil
}

func (r *LogbookRepository) GetByID(ctx context.Context, id string) (*models.AthleteLogbook, error) {
	logbook := &models.AthleteLogbook{}
	query := `
		SELECT id, user_id, content, updated_at
		FROM athlete_logbooks WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&logbook.ID,
		&logbook.UserID,
		&logbook.Content,
		&logbook.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("logbook not found")
		}
		return nil, fmt.Errorf("failed to get logbook: %w", err)
	}

	return logbook, nil
}

func (r *LogbookRepository) GetByUserID(ctx context.Context, userID string) (*models.AthleteLogbook, error) {
	logbook := &models.AthleteLogbook{}
	query := `
		SELECT id, user_id, content, updated_at
		FROM athlete_logbooks WHERE user_id = $1`

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&logbook.ID,
		&logbook.UserID,
		&logbook.Content,
		&logbook.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("logbook not found")
		}
		return nil, fmt.Errorf("failed to get logbook: %w", err)
	}

	return logbook, nil
}

func (r *LogbookRepository) Update(ctx context.Context, logbook *models.AthleteLogbook) error {
	query := `
		UPDATE athlete_logbooks 
		SET content = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRow(ctx, query,
		logbook.ID,
		logbook.Content,
	).Scan(&logbook.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update logbook: %w", err)
	}

	return nil
}

func (r *LogbookRepository) Upsert(ctx context.Context, logbook *models.AthleteLogbook) error {
	query := `
		INSERT INTO athlete_logbooks (user_id, content)
		VALUES ($1, $2)
		ON CONFLICT (user_id) 
		DO UPDATE SET content = EXCLUDED.content, updated_at = NOW()
		RETURNING id, updated_at`

	err := r.db.QueryRow(ctx, query,
		logbook.UserID,
		logbook.Content,
	).Scan(&logbook.ID, &logbook.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert logbook: %w", err)
	}

	return nil
}

func (r *LogbookRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM athlete_logbooks WHERE id = $1`
	
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete logbook: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("logbook not found")
	}

	return nil
}