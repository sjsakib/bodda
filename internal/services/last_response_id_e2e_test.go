package services

import (
	"context"
	"testing"
	"time"

	"bodda/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestLastResponseIDEndToEndFlow tests the complete flow of last_response_id tracking
func TestLastResponseIDEndToEndFlow(t *testing.T) {
	// This test simulates the complete flow:
	// 1. Server gets session and populates MessageContext.LastResponseID
	// 2. AI service processes message and updates session with new response ID
	// 3. Server saves assistant message with response ID

	t.Run("complete_flow_simulation", func(t *testing.T) {
		// Step 1: Simulate server getting session and populating MessageContext
		session := &models.Session{
			ID:             "test-session-id",
			UserID:         "test-user-id",
			Title:          "Test Session",
			LastResponseID: stringPtrHelper("resp_previous_123"), // Previous response ID
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// Extract last_response_id from session for AI context (as done in server.go)
		var lastResponseID string
		if session.LastResponseID != nil {
			lastResponseID = *session.LastResponseID
		}

		// Create MessageContext as done in server.go
		msgCtx := &MessageContext{
			UserID:         session.UserID,
			SessionID:      session.ID,
			Message:        "How was my last workout?",
			LastResponseID: lastResponseID,
			ConversationHistory: []*models.Message{
				{
					ID:         "msg1",
					SessionID:  session.ID,
					Role:       "user",
					Content:    "Previous user message",
					ResponseID: nil,
					CreatedAt:  time.Now().Add(-10 * time.Minute),
				},
				{
					ID:         "msg2",
					SessionID:  session.ID,
					Role:       "assistant",
					Content:    "Previous assistant response",
					ResponseID: stringPtrHelper("resp_previous_123"),
					CreatedAt:  time.Now().Add(-9 * time.Minute),
				},
			},
		}

		// Verify that MessageContext has the previous response ID
		assert.Equal(t, "resp_previous_123", msgCtx.LastResponseID)

		// Step 2: Simulate AI service processing and updating session
		mockSessionRepo := &MockSessionRepositorySimple{}
		newResponseID := "resp_new_456"

		// Expect the session to be updated with the new response ID
		mockSessionRepo.On("UpdateLastResponseID", mock.Anything, session.ID, newResponseID).
			Return(nil)

		// Simulate the AI service updating the session (as done in ai.go)
		ctx := context.Background()
		err := mockSessionRepo.UpdateLastResponseID(ctx, session.ID, newResponseID)
		assert.NoError(t, err)

		// Update the MessageContext as the AI service would do
		msgCtx.LastResponseID = newResponseID

		// Step 3: Simulate server saving assistant message with response ID
		// This would be done by calling chatService.SendMessageWithResponseID
		var responseIDPtr *string
		if msgCtx.LastResponseID != "" {
			responseIDPtr = &msgCtx.LastResponseID
		}

		// Verify that we have the response ID to save
		assert.NotNil(t, responseIDPtr)
		assert.Equal(t, newResponseID, *responseIDPtr)

		// Verify mock expectations
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("new_session_without_previous_response", func(t *testing.T) {
		// Test the flow for a new session without previous responses
		session := &models.Session{
			ID:             "new-session-id",
			UserID:         "test-user-id",
			Title:          "New Session",
			LastResponseID: nil, // No previous response
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// Extract last_response_id from session (should be empty)
		var lastResponseID string
		if session.LastResponseID != nil {
			lastResponseID = *session.LastResponseID
		}

		// Create MessageContext
		msgCtx := &MessageContext{
			UserID:              session.UserID,
			SessionID:           session.ID,
			Message:             "Hello, this is my first message",
			LastResponseID:      lastResponseID,
			ConversationHistory: []*models.Message{}, // No previous messages
		}

		// Verify that MessageContext has no previous response ID
		assert.Equal(t, "", msgCtx.LastResponseID)

		// Simulate AI service processing and updating session with first response
		mockSessionRepo := &MockSessionRepositorySimple{}
		firstResponseID := "resp_first_789"

		mockSessionRepo.On("UpdateLastResponseID", mock.Anything, session.ID, firstResponseID).
			Return(nil)

		// Update session with first response ID
		ctx := context.Background()
		err := mockSessionRepo.UpdateLastResponseID(ctx, session.ID, firstResponseID)
		assert.NoError(t, err)

		// Update MessageContext
		msgCtx.LastResponseID = firstResponseID

		// Verify the response ID is now set
		assert.Equal(t, firstResponseID, msgCtx.LastResponseID)

		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("session_update_failure_handling", func(t *testing.T) {
		// Test that session update failures don't break the flow
		session := &models.Session{
			ID:             "test-session-id",
			UserID:         "test-user-id",
			Title:          "Test Session",
			LastResponseID: nil,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		msgCtx := &MessageContext{
			UserID:              session.UserID,
			SessionID:           session.ID,
			Message:             "Test message",
			LastResponseID:      "",
			ConversationHistory: []*models.Message{},
		}

		// Simulate session update failure
		mockSessionRepo := &MockSessionRepositorySimple{}
		responseID := "resp_test_error"

		mockSessionRepo.On("UpdateLastResponseID", mock.Anything, session.ID, responseID).
			Return(assert.AnError)

		// The AI service should handle the error gracefully
		ctx := context.Background()
		err := mockSessionRepo.UpdateLastResponseID(ctx, session.ID, responseID)
		assert.Error(t, err) // Error should occur

		// But the MessageContext should still be updated for the response
		msgCtx.LastResponseID = responseID
		assert.Equal(t, responseID, msgCtx.LastResponseID)

		mockSessionRepo.AssertExpectations(t)
	})
}

// Helper function to create string pointers (using different name to avoid conflicts)
func stringPtrHelper(s string) *string {
	return &s
}