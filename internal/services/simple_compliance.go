package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"
)

// SimpleComplianceService handles basic compliance requirements without complex consent storage
type SimpleComplianceService interface {
	// Audit logging for compliance tracking
	LogUserAction(ctx context.Context, userID string, action string, details interface{}) error
	
	// Data export for user rights
	ExportUserData(ctx context.Context, userID string) (map[string]interface{}, error)
	
	// Account deletion with full cleanup
	DeleteUserAccount(ctx context.Context, userID string) error
	
	// Strava disconnection
	DisconnectStrava(ctx context.Context, userID string) error
}

// AuditLog represents a simple audit log entry
type AuditLog struct {
	ID        string          `json:"id" db:"id"`
	UserID    string          `json:"user_id" db:"user_id"`
	Action    string          `json:"action" db:"action"`
	Details   json.RawMessage `json:"details" db:"details"`
	Timestamp time.Time       `json:"timestamp" db:"timestamp"`
}

// SimpleAuditRepository defines minimal database interface for audit logging
type SimpleAuditRepository interface {
	CreateAuditEntry(ctx context.Context, entry *AuditLog) error
	GetUserAuditLog(ctx context.Context, userID string, limit int) ([]*AuditLog, error)
}

type simpleComplianceService struct {
	config      *config.Config
	auditRepo   SimpleAuditRepository
	userRepo    UserRepositoryInterface
	stravaService StravaService
}

// NewSimpleComplianceService creates a minimal compliance service
func NewSimpleComplianceService(
	cfg *config.Config, 
	auditRepo SimpleAuditRepository, 
	userRepo UserRepositoryInterface,
	stravaService StravaService,
) SimpleComplianceService {
	return &simpleComplianceService{
		config:        cfg,
		auditRepo:     auditRepo,
		userRepo:      userRepo,
		stravaService: stravaService,
	}
}

// LogUserAction logs user actions for compliance audit trail
func (s *simpleComplianceService) LogUserAction(ctx context.Context, userID string, action string, details interface{}) error {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("failed to marshal audit details: %w", err)
	}

	entry := &AuditLog{
		UserID:    userID,
		Action:    action,
		Details:   detailsJSON,
		Timestamp: time.Now(),
	}

	if err := s.auditRepo.CreateAuditEntry(ctx, entry); err != nil {
		// Log the error but don't fail the main operation
		log.Printf("Failed to create audit entry for user %s, action %s: %v", userID, action, err)
	}

	return nil
}

// ExportUserData exports all user data for GDPR/compliance purposes
func (s *simpleComplianceService) ExportUserData(ctx context.Context, userID string) (map[string]interface{}, error) {
	// Get user profile
	user := &models.User{ID: userID}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to get user data: %w", err)
	}

	// Get audit log (last 1000 entries)
	auditLog, err := s.auditRepo.GetUserAuditLog(ctx, userID, 1000)
	if err != nil {
		log.Printf("Failed to get audit log for user %s: %v", userID, err)
		auditLog = []*AuditLog{} // Continue without audit log if it fails
	}

	// Prepare export data
	exportData := map[string]interface{}{
		"user_profile": map[string]interface{}{
			"id":           user.ID,
			"first_name":   user.FirstName,
			"last_name":    user.LastName,
			"created_at":   user.CreatedAt,
			"updated_at":   user.UpdatedAt,
			// Don't export sensitive data like tokens
		},
		"strava_connection": map[string]interface{}{
			"connected":    user.AccessToken != "",
			"connected_at": user.CreatedAt, // Approximate
		},
		"audit_log":         auditLog,
		"export_timestamp":  time.Now(),
		"export_version":    "1.0",
	}

	// Log the export action
	s.LogUserAction(ctx, userID, "data_exported", map[string]interface{}{
		"export_size": len(exportData),
	})

	return exportData, nil
}

// DeleteUserAccount permanently deletes user account and all associated data
func (s *simpleComplianceService) DeleteUserAccount(ctx context.Context, userID string) error {
	// Log the deletion action before deleting
	s.LogUserAction(ctx, userID, "account_deleted", map[string]interface{}{
		"deletion_timestamp": time.Now(),
		"deletion_type":      "user_requested",
	})

	// Get user to revoke Strava tokens first
	user := &models.User{ID: userID}
	if err := s.userRepo.Update(ctx, user); err != nil {
		log.Printf("Failed to get user for deletion %s: %v", userID, err)
	} else if user.RefreshToken != "" {
		// Try to revoke Strava tokens (best effort)
		if _, err := s.stravaService.RefreshToken(user.RefreshToken); err != nil {
			log.Printf("Failed to revoke Strava tokens for user %s: %v", userID, err)
		}
	}

	// Delete user from database (should cascade to delete related data)
	// This depends on your database schema having proper CASCADE constraints
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to delete user account: %w", err)
	}

	log.Printf("Successfully deleted account for user %s", userID)
	return nil
}

// DisconnectStrava disconnects user's Strava account
func (s *simpleComplianceService) DisconnectStrava(ctx context.Context, userID string) error {
	// Get user
	user := &models.User{ID: userID}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Try to revoke Strava tokens
	if user.RefreshToken != "" {
		if _, err := s.stravaService.RefreshToken(user.RefreshToken); err != nil {
			log.Printf("Failed to revoke Strava tokens for user %s: %v", userID, err)
		}
	}

	// Clear Strava tokens from user record
	user.AccessToken = ""
	user.RefreshToken = ""
	user.TokenExpiry = time.Time{}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Log the disconnection
	s.LogUserAction(ctx, userID, "strava_disconnected", map[string]interface{}{
		"disconnection_timestamp": time.Now(),
	})

	return nil
}

// Helper function to check if user has Strava connected (for blocking access)
func (s *simpleComplianceService) HasStravaConnected(ctx context.Context, userID string) (bool, error) {
	user := &models.User{ID: userID}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	return user.AccessToken != "" && user.TokenExpiry.After(time.Now()), nil
}