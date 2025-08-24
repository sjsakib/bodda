package database

import (
	"context"
	"bodda/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LogbookRepositoryInterface defines the interface for logbook repository operations
type LogbookRepositoryInterface interface {
	Create(ctx context.Context, logbook *models.AthleteLogbook) error
	GetByID(ctx context.Context, id string) (*models.AthleteLogbook, error)
	GetByUserID(ctx context.Context, userID string) (*models.AthleteLogbook, error)
	Update(ctx context.Context, logbook *models.AthleteLogbook) error
	Upsert(ctx context.Context, logbook *models.AthleteLogbook) error
	Delete(ctx context.Context, id string) error
}

// Repository provides access to all database repositories
type Repository struct {
	User     *UserRepository
	Session  *SessionRepository
	Message  *MessageRepository
	Logbook  *LogbookRepository
}

// NewRepository creates a new repository instance with all sub-repositories
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		User:     NewUserRepository(db),
		Session:  NewSessionRepository(db),
		Message:  NewMessageRepository(db),
		Logbook:  NewLogbookRepository(db),
	}
}