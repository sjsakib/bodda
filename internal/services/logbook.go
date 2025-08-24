package services

import (
	"context"
	"fmt"
	"strings"

	"bodda/internal/database"
	"bodda/internal/models"
)



type LogbookService interface {
	GetLogbook(ctx context.Context, userID string) (*models.AthleteLogbook, error)
	CreateInitialLogbook(ctx context.Context, userID string, stravaProfile *StravaAthlete) (*models.AthleteLogbook, error)
	UpdateLogbook(ctx context.Context, userID string, content string) (*models.AthleteLogbook, error)
	UpsertLogbook(ctx context.Context, userID string, content string) (*models.AthleteLogbook, error)
}

type logbookService struct {
	repo database.LogbookRepositoryInterface
}

func NewLogbookService(repo database.LogbookRepositoryInterface) LogbookService {
	return &logbookService{
		repo: repo,
	}
}

func (s *logbookService) GetLogbook(ctx context.Context, userID string) (*models.AthleteLogbook, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	logbook, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("logbook not found for user %s", userID)
		}
		return nil, fmt.Errorf("failed to retrieve logbook: %w", err)
	}

	return logbook, nil
}

func (s *logbookService) CreateInitialLogbook(ctx context.Context, userID string, stravaProfile *StravaAthlete) (*models.AthleteLogbook, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	if stravaProfile == nil {
		return nil, fmt.Errorf("strava profile cannot be nil")
	}

	// Create initial logbook content as a simple string
	initialContent := fmt.Sprintf(`Athlete Profile:
Name: %s %s
Gender: %s
Location: %s
Weight: %.1f kg
FTP: %d watts
Strava Member Since: %s

Initial logbook created from Strava profile data. This logbook will be updated with training insights, goals, preferences, and coaching observations as we learn more about the athlete.`,
		stravaProfile.Firstname,
		stravaProfile.Lastname,
		stravaProfile.Sex,
		formatLocation(stravaProfile.City, stravaProfile.State, stravaProfile.Country),
		stravaProfile.Weight,
		stravaProfile.FTP,
		stravaProfile.CreatedAt,
	)

	logbook := &models.AthleteLogbook{
		UserID:  userID,
		Content: initialContent,
	}

	if err := s.repo.Create(ctx, logbook); err != nil {
		return nil, fmt.Errorf("failed to create initial logbook: %w", err)
	}

	return logbook, nil
}

func (s *logbookService) UpdateLogbook(ctx context.Context, userID string, content string) (*models.AthleteLogbook, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	if content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	existingLogbook, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("logbook not found for user %s", userID)
		}
		return nil, fmt.Errorf("failed to retrieve existing logbook: %w", err)
	}

	existingLogbook.Content = content
	if err := s.repo.Update(ctx, existingLogbook); err != nil {
		return nil, fmt.Errorf("failed to update logbook: %w", err)
	}

	return existingLogbook, nil
}

func (s *logbookService) UpsertLogbook(ctx context.Context, userID string, content string) (*models.AthleteLogbook, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	if content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	logbook := &models.AthleteLogbook{
		UserID:  userID,
		Content: content,
	}

	if err := s.repo.Upsert(ctx, logbook); err != nil {
		return nil, fmt.Errorf("failed to upsert logbook: %w", err)
	}

	return logbook, nil
}





func formatLocation(city, state, country string) string {
	parts := []string{}
	if city != "" {
		parts = append(parts, city)
	}
	if state != "" {
		parts = append(parts, state)
	}
	if country != "" {
		parts = append(parts, country)
	}
	return strings.Join(parts, ", ")
}

