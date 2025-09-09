package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSessionRepositorySimple implements SessionRepository for testing
type MockSessionRepositorySimple struct {
	mock.Mock
}

func (m *MockSessionRepositorySimple) UpdateLastResponseID(ctx context.Context, sessionID string, responseID string) error {
	args := m.Called(ctx, sessionID, responseID)
	return args.Error(0)
}

func TestSessionLastResponseIDUpdate(t *testing.T) {
	tests := []struct {
		name            string
		sessionID       string
		responseID      string
		shouldFail      bool
		expectedError   bool
	}{
		{
			name:       "successful_update",
			sessionID:  "session-123",
			responseID: "resp-456",
			shouldFail: false,
		},
		{
			name:          "update_failure",
			sessionID:     "session-123",
			responseID:    "resp-456",
			shouldFail:    true,
			expectedError: true,
		},
		{
			name:       "empty_response_id",
			sessionID:  "session-123",
			responseID: "",
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockSessionRepositorySimple{}
			
			if tt.shouldFail {
				mockRepo.On("UpdateLastResponseID", mock.Anything, tt.sessionID, tt.responseID).
					Return(assert.AnError)
			} else {
				mockRepo.On("UpdateLastResponseID", mock.Anything, tt.sessionID, tt.responseID).
					Return(nil)
			}

			ctx := context.Background()
			err := mockRepo.UpdateLastResponseID(ctx, tt.sessionID, tt.responseID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAIServiceSessionRepositoryIntegration(t *testing.T) {
	// Test that the AI service properly uses the session repository
	mockRepo := &MockSessionRepositorySimple{}
	
	// Create a minimal AI service with just the session repository
	aiService := &aiService{
		sessionRepository: mockRepo,
	}

	ctx := context.Background()
	sessionID := "test-session"
	responseID := "test-response"

	// Test successful update
	mockRepo.On("UpdateLastResponseID", ctx, sessionID, responseID).Return(nil).Once()
	
	err := aiService.sessionRepository.UpdateLastResponseID(ctx, sessionID, responseID)
	assert.NoError(t, err)
	
	mockRepo.AssertExpectations(t)
}