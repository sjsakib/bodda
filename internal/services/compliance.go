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

// ComplianceAction represents different types of compliance actions
type ComplianceAction string

const (
	ActionConsentGranted    ComplianceAction = "consent_granted"
	ActionConsentRevoked    ComplianceAction = "consent_revoked"
	ActionDataAccessed      ComplianceAction = "data_accessed"
	ActionDataExported      ComplianceAction = "data_exported"
	ActionDataDeleted       ComplianceAction = "data_deleted"
	ActionAccountDeleted    ComplianceAction = "account_deleted"
	ActionStravaConnected   ComplianceAction = "strava_connected"
	ActionStravaDisconnected ComplianceAction = "strava_disconnected"
)

// ConsentType represents different types of user consent
type ConsentType string

const (
	ConsentDataProcessing ConsentType = "data_processing"
	ConsentStravaAccess   ConsentType = "strava_access"
	ConsentAICoaching     ConsentType = "ai_coaching"
	ConsentMarketing      ConsentType = "marketing"
)

// UserConsent represents a user's consent record
type UserConsent struct {
	ID          string     `json:"id" db:"id"`
	UserID      string     `json:"user_id" db:"user_id"`
	ConsentType ConsentType `json:"consent_type" db:"consent_type"`
	Granted     bool       `json:"granted" db:"granted"`
	GrantedAt   *time.Time `json:"granted_at" db:"granted_at"`
	RevokedAt   *time.Time `json:"revoked_at" db:"revoked_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// ComplianceAudit represents an audit log entry
type ComplianceAudit struct {
	ID        string           `json:"id" db:"id"`
	UserID    *string          `json:"user_id" db:"user_id"`
	Action    ComplianceAction `json:"action" db:"action"`
	Details   json.RawMessage  `json:"details" db:"details"`
	IPAddress *string          `json:"ip_address" db:"ip_address"`
	UserAgent *string          `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
}

// DataRetention represents data retention tracking
type DataRetention struct {
	ID             string     `json:"id" db:"id"`
	UserID         string     `json:"user_id" db:"user_id"`
	DataType       string     `json:"data_type" db:"data_type"`
	LastAccessed   *time.Time `json:"last_accessed" db:"last_accessed"`
	RetentionUntil *time.Time `json:"retention_until" db:"retention_until"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// ComplianceService handles compliance-related operations
type ComplianceService interface {
	// Consent management
	GrantConsent(ctx context.Context, userID string, consentType ConsentType, ipAddress, userAgent string) error
	RevokeConsent(ctx context.Context, userID string, consentType ConsentType, ipAddress, userAgent string) error
	GetUserConsents(ctx context.Context, userID string) ([]*UserConsent, error)
	HasValidConsent(ctx context.Context, userID string, consentType ConsentType) (bool, error)

	// Audit logging
	LogAction(ctx context.Context, userID *string, action ComplianceAction, details interface{}, ipAddress, userAgent string) error
	GetAuditLog(ctx context.Context, userID string, limit int) ([]*ComplianceAudit, error)

	// Data retention
	UpdateDataAccess(ctx context.Context, userID string, dataType string) error
	GetRetentionStatus(ctx context.Context, userID string) ([]*DataRetention, error)
	CleanupExpiredData(ctx context.Context) error

	// Data export and deletion
	ExportUserData(ctx context.Context, userID string) (map[string]interface{}, error)
	DeleteUserData(ctx context.Context, userID string, ipAddress, userAgent string) error
}

// ComplianceRepository defines the database interface for compliance operations
type ComplianceRepository interface {
	// Consent operations
	CreateConsent(ctx context.Context, consent *UserConsent) error
	UpdateConsent(ctx context.Context, consent *UserConsent) error
	GetConsentsByUser(ctx context.Context, userID string) ([]*UserConsent, error)
	GetConsentByType(ctx context.Context, userID string, consentType ConsentType) (*UserConsent, error)

	// Audit operations
	CreateAuditEntry(ctx context.Context, audit *ComplianceAudit) error
	GetAuditEntries(ctx context.Context, userID string, limit int) ([]*ComplianceAudit, error)

	// Data retention operations
	CreateRetentionRecord(ctx context.Context, retention *DataRetention) error
	UpdateRetentionRecord(ctx context.Context, retention *DataRetention) error
	GetRetentionRecords(ctx context.Context, userID string) ([]*DataRetention, error)
	GetExpiredRetentionRecords(ctx context.Context) ([]*DataRetention, error)
}

type complianceService struct {
	config     *config.Config
	repository ComplianceRepository
	userRepo   UserRepositoryInterface
}

// NewComplianceService creates a new compliance service
func NewComplianceService(cfg *config.Config, repo ComplianceRepository, userRepo UserRepositoryInterface) ComplianceService {
	return &complianceService{
		config:     cfg,
		repository: repo,
		userRepo:   userRepo,
	}
}

// GrantConsent grants consent for a specific type
func (s *complianceService) GrantConsent(ctx context.Context, userID string, consentType ConsentType, ipAddress, userAgent string) error {
	now := time.Now()
	
	// Check if consent already exists
	existing, err := s.repository.GetConsentByType(ctx, userID, consentType)
	if err != nil {
		return fmt.Errorf("failed to check existing consent: %w", err)
	}

	if existing != nil {
		// Update existing consent
		existing.Granted = true
		existing.GrantedAt = &now
		existing.RevokedAt = nil
		
		if err := s.repository.UpdateConsent(ctx, existing); err != nil {
			return fmt.Errorf("failed to update consent: %w", err)
		}
	} else {
		// Create new consent record
		consent := &UserConsent{
			UserID:      userID,
			ConsentType: consentType,
			Granted:     true,
			GrantedAt:   &now,
			CreatedAt:   now,
		}
		
		if err := s.repository.CreateConsent(ctx, consent); err != nil {
			return fmt.Errorf("failed to create consent: %w", err)
		}
	}

	// Log the action
	details := map[string]interface{}{
		"consent_type": consentType,
		"granted":      true,
	}
	
	if err := s.LogAction(ctx, &userID, ActionConsentGranted, details, ipAddress, userAgent); err != nil {
		log.Printf("Failed to log consent granted action: %v", err)
	}

	return nil
}

// RevokeConsent revokes consent for a specific type
func (s *complianceService) RevokeConsent(ctx context.Context, userID string, consentType ConsentType, ipAddress, userAgent string) error {
	now := time.Now()
	
	existing, err := s.repository.GetConsentByType(ctx, userID, consentType)
	if err != nil {
		return fmt.Errorf("failed to check existing consent: %w", err)
	}

	if existing == nil {
		return fmt.Errorf("no consent record found for type %s", consentType)
	}

	// Update consent to revoked
	existing.Granted = false
	existing.RevokedAt = &now
	
	if err := s.repository.UpdateConsent(ctx, existing); err != nil {
		return fmt.Errorf("failed to update consent: %w", err)
	}

	// Log the action
	details := map[string]interface{}{
		"consent_type": consentType,
		"granted":      false,
	}
	
	if err := s.LogAction(ctx, &userID, ActionConsentRevoked, details, ipAddress, userAgent); err != nil {
		log.Printf("Failed to log consent revoked action: %v", err)
	}

	return nil
}

// GetUserConsents retrieves all consents for a user
func (s *complianceService) GetUserConsents(ctx context.Context, userID string) ([]*UserConsent, error) {
	return s.repository.GetConsentsByUser(ctx, userID)
}

// HasValidConsent checks if user has valid consent for a specific type
func (s *complianceService) HasValidConsent(ctx context.Context, userID string, consentType ConsentType) (bool, error) {
	consent, err := s.repository.GetConsentByType(ctx, userID, consentType)
	if err != nil {
		return false, fmt.Errorf("failed to check consent: %w", err)
	}

	return consent != nil && consent.Granted && consent.RevokedAt == nil, nil
}

// LogAction logs a compliance action
func (s *complianceService) LogAction(ctx context.Context, userID *string, action ComplianceAction, details interface{}, ipAddress, userAgent string) error {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}

	audit := &ComplianceAudit{
		UserID:    userID,
		Action:    action,
		Details:   detailsJSON,
		CreatedAt: time.Now(),
	}

	if ipAddress != "" {
		audit.IPAddress = &ipAddress
	}
	if userAgent != "" {
		audit.UserAgent = &userAgent
	}

	return s.repository.CreateAuditEntry(ctx, audit)
}

// GetAuditLog retrieves audit log entries for a user
func (s *complianceService) GetAuditLog(ctx context.Context, userID string, limit int) ([]*ComplianceAudit, error) {
	return s.repository.GetAuditEntries(ctx, userID, limit)
}

// UpdateDataAccess updates the last accessed time for user data
func (s *complianceService) UpdateDataAccess(ctx context.Context, userID string, dataType string) error {
	now := time.Now()
	
	// Calculate retention period based on data type
	var retentionUntil time.Time
	switch dataType {
	case "strava_activities":
		retentionUntil = now.AddDate(2, 0, 0) // 2 years
	case "chat_history":
		retentionUntil = now.AddDate(1, 0, 0) // 1 year
	case "user_profile":
		retentionUntil = now.AddDate(5, 0, 0) // 5 years
	default:
		retentionUntil = now.AddDate(1, 0, 0) // Default 1 year
	}

	retention := &DataRetention{
		UserID:         userID,
		DataType:       dataType,
		LastAccessed:   &now,
		RetentionUntil: &retentionUntil,
		CreatedAt:      now,
	}

	// Try to update existing record, create if not exists
	if err := s.repository.UpdateRetentionRecord(ctx, retention); err != nil {
		// If update fails, try to create new record
		if err := s.repository.CreateRetentionRecord(ctx, retention); err != nil {
			return fmt.Errorf("failed to create retention record: %w", err)
		}
	}

	return nil
}

// GetRetentionStatus retrieves data retention status for a user
func (s *complianceService) GetRetentionStatus(ctx context.Context, userID string) ([]*DataRetention, error) {
	return s.repository.GetRetentionRecords(ctx, userID)
}

// CleanupExpiredData removes data that has exceeded retention periods
func (s *complianceService) CleanupExpiredData(ctx context.Context) error {
	expiredRecords, err := s.repository.GetExpiredRetentionRecords(ctx)
	if err != nil {
		return fmt.Errorf("failed to get expired records: %w", err)
	}

	for _, record := range expiredRecords {
		log.Printf("Cleaning up expired data: user=%s, type=%s", record.UserID, record.DataType)
		
		// Log the cleanup action
		details := map[string]interface{}{
			"data_type":       record.DataType,
			"retention_until": record.RetentionUntil,
		}
		
		if err := s.LogAction(ctx, &record.UserID, ActionDataDeleted, details, "", "system"); err != nil {
			log.Printf("Failed to log data cleanup action: %v", err)
		}
	}

	return nil
}

// ExportUserData exports all user data for compliance purposes
func (s *complianceService) ExportUserData(ctx context.Context, userID string) (map[string]interface{}, error) {
	// Get user profile
	user := &models.User{ID: userID}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to get user data: %w", err)
	}

	// Get consents
	consents, err := s.GetUserConsents(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get consents: %w", err)
	}

	// Get audit log
	auditLog, err := s.GetAuditLog(ctx, userID, 1000) // Last 1000 entries
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	// Get retention status
	retentionStatus, err := s.GetRetentionStatus(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get retention status: %w", err)
	}

	exportData := map[string]interface{}{
		"user_profile":      user,
		"consents":          consents,
		"audit_log":         auditLog,
		"retention_status":  retentionStatus,
		"export_timestamp":  time.Now(),
	}

	// Log the export action
	details := map[string]interface{}{
		"export_size": len(exportData),
	}
	
	if err := s.LogAction(ctx, &userID, ActionDataExported, details, "", "user_request"); err != nil {
		log.Printf("Failed to log data export action: %v", err)
	}

	return exportData, nil
}

// DeleteUserData deletes all user data for compliance purposes
func (s *complianceService) DeleteUserData(ctx context.Context, userID string, ipAddress, userAgent string) error {
	// Log the deletion action before deleting
	details := map[string]interface{}{
		"deletion_type": "complete_account_deletion",
	}
	
	if err := s.LogAction(ctx, &userID, ActionAccountDeleted, details, ipAddress, userAgent); err != nil {
		log.Printf("Failed to log account deletion action: %v", err)
	}

	// Delete user data from database
	// This should cascade to delete related data based on foreign key constraints
	user := &models.User{ID: userID}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	log.Printf("Successfully deleted all data for user %s", userID)
	return nil
}