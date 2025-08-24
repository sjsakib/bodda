package services

import (
	"context"
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"

	"bodda/internal/database"
	"bodda/internal/models"
)

// Custom error types for chat service
var (
	ErrInvalidSessionTitle   = errors.New("invalid session title")
	ErrInvalidMessageContent = errors.New("invalid message content")
	ErrMessageTooLong        = errors.New("message content is too long")
	ErrSessionNotFound       = errors.New("session not found")
	ErrUnauthorizedAccess    = errors.New("unauthorized access to session")
)

type ChatService interface {
	CreateSession(userID string) (*models.Session, error)
	CreateSessionWithTitle(userID, title string) (*models.Session, error)
	GetSessions(userID string) ([]*models.Session, error)
	GetSession(sessionID string) (*models.Session, error)
	UpdateSessionTitle(sessionID, title string) error
	DeleteSession(sessionID string) error
	SendMessage(sessionID, role, content string) (*models.Message, error)
	GetMessages(sessionID string) ([]*models.Message, error)
	GetMessagesWithPagination(sessionID string, limit, offset int) ([]*models.Message, error)
	GetMessageCount(sessionID string) (int, error)
	StreamResponse(sessionID string, response chan string) error
}

type chatService struct {
	repo *database.Repository
}

func NewChatService(repo *database.Repository) ChatService {
	return &chatService{
		repo: repo,
	}
}

func (s *chatService) CreateSession(userID string) (*models.Session, error) {
	return s.CreateSessionWithTitle(userID, "New Conversation")
}

func (s *chatService) CreateSessionWithTitle(userID, title string) (*models.Session, error) {
	ctx := context.Background()

	// Validate and sanitize inputs
	if err := s.validateUserID(userID); err != nil {
		return nil, err
	}

	sanitizedTitle, err := s.validateAndSanitizeTitle(title)
	if err != nil {
		return nil, err
	}

	session := &models.Session{
		UserID: userID,
		Title:  sanitizedTitle,
	}

	err = s.repo.Session.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (s *chatService) GetSessions(userID string) ([]*models.Session, error) {
	ctx := context.Background()

	sessions, err := s.repo.Session.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	return sessions, nil
}

func (s *chatService) GetSession(sessionID string) (*models.Session, error) {
	ctx := context.Background()

	session, err := s.repo.Session.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

func (s *chatService) UpdateSessionTitle(sessionID, title string) error {
	ctx := context.Background()

	// Validate inputs
	if err := s.validateSessionID(sessionID); err != nil {
		return err
	}

	sanitizedTitle, err := s.validateAndSanitizeTitle(title)
	if err != nil {
		return err
	}

	// First get the session to ensure it exists
	session, err := s.repo.Session.GetByID(ctx, sessionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrSessionNotFound
		}
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Update the title
	session.Title = sanitizedTitle
	err = s.repo.Session.Update(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to update session title: %w", err)
	}

	return nil
}

func (s *chatService) DeleteSession(sessionID string) error {
	ctx := context.Background()

	err := s.repo.Session.Delete(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func (s *chatService) SendMessage(sessionID, role, content string) (*models.Message, error) {
	ctx := context.Background()

	// Validate inputs
	if err := s.validateSessionID(sessionID); err != nil {
		return nil, err
	}

	if err := s.validateRole(role); err != nil {
		return nil, err
	}

	sanitizedContent, err := s.validateAndSanitizeContent(content)
	if err != nil {
		return nil, err
	}

	// Verify session exists
	_, err = s.repo.Session.GetByID(ctx, sessionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("session not found: %w", err)
	}

	message := &models.Message{
		SessionID: sessionID,
		Role:      role,
		Content:   sanitizedContent,
	}

	err = s.repo.Message.Create(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return message, nil
}

func (s *chatService) GetMessages(sessionID string) ([]*models.Message, error) {
	ctx := context.Background()

	messages, err := s.repo.Message.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

func (s *chatService) GetMessagesWithPagination(sessionID string, limit, offset int) ([]*models.Message, error) {
	ctx := context.Background()

	messages, err := s.repo.Message.GetBySessionIDWithPagination(ctx, sessionID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages with pagination: %w", err)
	}

	return messages, nil
}

func (s *chatService) GetMessageCount(sessionID string) (int, error) {
	ctx := context.Background()

	count, err := s.repo.Message.CountBySessionID(ctx, sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to get message count: %w", err)
	}

	return count, nil
}

func (s *chatService) StreamResponse(sessionID string, response chan string) error {
	// This is a placeholder for streaming functionality
	// The actual streaming will be implemented when integrating with AI service
	// For now, we just validate that the session exists
	ctx := context.Background()

	_, err := s.repo.Session.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Simulate streaming by sending a test message
	go func() {
		defer close(response)
		response <- "Streaming functionality will be implemented with AI integration"
	}()

	return nil
}

// Validation and sanitization helper methods

func (s *chatService) validateUserID(userID string) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	// Basic UUID format validation
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(userID) {
		return fmt.Errorf("invalid user ID format")
	}

	return nil
}

func (s *chatService) validateSessionID(sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	// Basic UUID format validation
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(sessionID) {
		return fmt.Errorf("invalid session ID format")
	}

	return nil
}

func (s *chatService) validateRole(role string) error {
	if role != "user" && role != "assistant" {
		return fmt.Errorf("invalid role: %s, must be 'user' or 'assistant'", role)
	}
	return nil
}

func (s *chatService) validateAndSanitizeTitle(title string) (string, error) {
	// Trim whitespace
	title = strings.TrimSpace(title)

	// Check length
	if len(title) == 0 {
		return "New Conversation", nil // Default title
	}

	if len(title) > 200 {
		return "", ErrInvalidSessionTitle
	}

	// Sanitize HTML
	title = html.EscapeString(title)

	// Remove control characters and normalize whitespace
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")
	title = regexp.MustCompile(`[^\x20-\x7E]`).ReplaceAllString(title, "")

	return title, nil
}

func (s *chatService) validateAndSanitizeContent(content string) (string, error) {
	// Trim whitespace
	content = strings.TrimSpace(content)

	// Check if empty
	if len(content) == 0 {
		return "", ErrInvalidMessageContent
	}

	// Check length (max 10KB)
	if len(content) > 10240 {
		return "", ErrMessageTooLong
	}

	// Basic sanitization - remove null bytes and control characters except newlines and tabs
	content = strings.ReplaceAll(content, "\x00", "")
	content = regexp.MustCompile(`[\x01-\x08\x0B\x0C\x0E-\x1F\x7F]`).ReplaceAllString(content, "")

	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	return content, nil
}
