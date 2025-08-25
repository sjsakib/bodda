package services

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"
)

func TestStravaServiceErrorHandling(t *testing.T) {
	cfg := &config.Config{
		StravaClientID:     "test_client_id",
		StravaClientSecret: "test_client_secret",
	}

	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedError  error
	}{
		{
			name:          "Rate limit exceeded",
			statusCode:    429,
			responseBody:  `{"message": "Rate Limit Exceeded"}`,
			expectedError: ErrRateLimitExceeded,
		},
		{
			name:          "Token expired",
			statusCode:    401,
			responseBody:  `{"message": "Authorization Error"}`,
			expectedError: ErrTokenExpired,
		},
		{
			name:          "Invalid token",
			statusCode:    403,
			responseBody:  `{"message": "Forbidden"}`,
			expectedError: ErrInvalidToken,
		},
		{
			name:          "Activity not found",
			statusCode:    404,
			responseBody:  `{"message": "Record Not Found"}`,
			expectedError: ErrActivityNotFound,
		},
		{
			name:          "Service unavailable",
			statusCode:    503,
			responseBody:  `{"message": "Service Unavailable"}`,
			expectedError: ErrServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server that returns the expected status code and body
			testServer := createTestServer(tt.statusCode, tt.responseBody)
			defer testServer.Close()

			mockUserRepo := &MockStravaUserRepository{}
			service := NewTestStravaService(cfg, testServer.URL, mockUserRepo)

			// Test GetAthleteProfile
			testUser := &models.User{
				ID:           "test-user-id",
				AccessToken:  "test_token",
				RefreshToken: "test_refresh_token",
				TokenExpiry:  time.Now().Add(time.Hour),
			}
			_, err := service.GetAthleteProfile(testUser)
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestInputValidation(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		sessionID   string
		message     string
		expectError bool
	}{
		{
			name:        "Empty user ID",
			userID:      "",
			sessionID:   "test-session-id",
			message:     "test message",
			expectError: true,
		},
		{
			name:        "Empty session ID",
			userID:      "test-user-id",
			sessionID:   "",
			message:     "test message",
			expectError: true,
		},
		{
			name:        "Empty message",
			userID:      "test-user-id",
			sessionID:   "test-session-id",
			message:     "",
			expectError: true,
		},
		{
			name:        "Message too long",
			userID:      "test-user-id",
			sessionID:   "test-session-id",
			message:     string(make([]byte, 9000)),
			expectError: true,
		},
		{
			name:        "Valid inputs",
			userID:      "test-user-id",
			sessionID:   "test-session-id",
			message:     "test message",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgCtx := &MessageContext{
				UserID:    tt.userID,
				SessionID: tt.sessionID,
				Message:   tt.message,
			}

			// Create a mock AI service to test validation
			cfg := &config.Config{OpenAIAPIKey: "test_key"}
			mockStravaService := &mockStravaService{}
			mockLogbookService := &mockLogbookService{}
			aiService := NewAIService(cfg, mockStravaService, mockLogbookService).(*aiService)

			err := aiService.validateMessageContext(msgCtx)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidationHelpers(t *testing.T) {
	// Create a mock chat service to test validation helpers
	service := &chatService{}

	tests := []struct {
		name        string
		operation   string
		input       string
		expectError bool
		errorType   error
	}{
		{
			name:        "Title too long",
			operation:   "validateTitle",
			input:       string(make([]byte, 250)),
			expectError: true,
			errorType:   ErrInvalidSessionTitle,
		},
		{
			name:        "Valid title",
			operation:   "validateTitle",
			input:       "Valid Title",
			expectError: false,
		},
		{
			name:        "Message too long",
			operation:   "validateContent",
			input:       string(make([]byte, 15000)),
			expectError: true,
			errorType:   ErrMessageTooLong,
		},
		{
			name:        "Empty message content",
			operation:   "validateContent",
			input:       "",
			expectError: true,
			errorType:   ErrInvalidMessageContent,
		},
		{
			name:        "Valid message content",
			operation:   "validateContent",
			input:       "Valid message",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			switch tt.operation {
			case "validateTitle":
				_, err = service.validateAndSanitizeTitle(tt.input)
			case "validateContent":
				_, err = service.validateAndSanitizeContent(tt.input)
			}

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorType != nil && !errors.Is(err, tt.errorType) {
					t.Errorf("Expected error type %v, got %v", tt.errorType, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestRateLimiterErrorHandling(t *testing.T) {
	// Test rate limiter functionality
	rateLimiter := NewRateLimiter(2, 1*time.Second)

	// First two requests should be allowed
	if !rateLimiter.Allow() {
		t.Error("First request should be allowed")
	}
	if !rateLimiter.Allow() {
		t.Error("Second request should be allowed")
	}

	// Third request should be denied
	if rateLimiter.Allow() {
		t.Error("Third request should be denied")
	}

	// Wait for window to reset
	time.Sleep(1100 * time.Millisecond)

	// Request should be allowed again
	if !rateLimiter.Allow() {
		t.Error("Request after window reset should be allowed")
	}
}

// Mock implementations for testing

type mockStravaService struct{}

func (m *mockStravaService) GetAthleteProfile(user *models.User) (*StravaAthlete, error) {
	return nil, ErrTokenExpired
}

func (m *mockStravaService) GetActivities(user *models.User, params ActivityParams) ([]*StravaActivity, error) {
	return nil, ErrRateLimitExceeded
}

func (m *mockStravaService) GetActivityDetail(user *models.User, activityID int64) (*StravaActivityDetail, error) {
	return nil, ErrActivityNotFound
}

func (m *mockStravaService) GetActivityStreams(user *models.User, activityID int64, streamTypes []string, resolution string) (*StravaStreams, error) {
	return nil, ErrServiceUnavailable
}

func (m *mockStravaService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	return nil, ErrInvalidToken
}

type mockLogbookService struct{}

func (m *mockLogbookService) GetLogbook(ctx context.Context, userID string) (*models.AthleteLogbook, error) {
	return nil, errors.New("logbook not found")
}

func (m *mockLogbookService) CreateInitialLogbook(ctx context.Context, userID string, athleteData *StravaAthlete) (*models.AthleteLogbook, error) {
	return nil, nil
}

func (m *mockLogbookService) UpdateLogbook(ctx context.Context, userID string, content string) (*models.AthleteLogbook, error) {
	return nil, nil
}

func (m *mockLogbookService) UpsertLogbook(ctx context.Context, userID, content string) (*models.AthleteLogbook, error) {
	return nil, nil
}

// createTestServer creates a test HTTP server that returns the specified status code and body
func createTestServer(statusCode int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}))
}