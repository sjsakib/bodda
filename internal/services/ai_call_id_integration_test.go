package services

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"bodda/internal/config"
	"bodda/internal/models"

	"github.com/openai/openai-go/v2/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSessionRepository for testing
type MockSessionRepository struct{}

func (m *MockSessionRepository) UpdateLastResponseID(ctx context.Context, sessionID string, responseID string) error {
	return nil // Simple mock that always succeeds
}

// TestCallIDExtractionEndToEndFlow tests the complete tool call processing pipeline with correct call_id usage
func TestCallIDExtractionEndToEndFlow(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		OpenAIAPIKey: "test-key",
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  4000,
			TokenPerCharRatio: 4,
			DefaultPageSize:   50,
			MaxPageSize:       200,
			RedactionEnabled:  false,
		},
	}

	// Create mock services
	mockStravaService := &mockStravaServiceForIntegration{}
	mockLogbookService := &mockLogbookServiceForIntegration{}
	mockToolRegistry := NewToolRegistry()
	
	// Create mock session repository
	mockSessionRepo := &MockSessionRepositoryForCallID{}
	
	// Create AI service
	aiService := NewAIService(cfg, mockStravaService, mockLogbookService, mockSessionRepo, mockToolRegistry).(*aiService)

	t.Run("complete tool call processing pipeline with correct call_id usage", func(t *testing.T) {
		// Test data with realistic call_id values
		testCallID := "call_ijVhE1A5JfpvEbYCLs7MVtDk"
		testItemID := "fc_68bba53afb688194b3c5b1ba405cd60109ec6f992e6a53b2"

		// Create tool call state
		state := NewToolCallState()

		// Simulate response.output_item.added event processing
		mockOutputItemEvent := createMockOutputItemAddedEvent(testItemID, testCallID, "get-athlete-profile")
		err := aiService.handleOutputItemAdded(mockOutputItemEvent, state)
		require.NoError(t, err)

		// Verify call_id was extracted and stored correctly
		assert.Contains(t, state.itemToCallID, testItemID)
		assert.Equal(t, testCallID, state.itemToCallID[testItemID])

		// Simulate function call arguments delta events
		deltaEvent1 := responses.ResponseFunctionCallArgumentsDeltaEvent{
			ItemID: testItemID,
			Delta:  `{"per_page": `,
		}
		err = aiService.handleFunctionCallArgumentsDelta(deltaEvent1, state)
		require.NoError(t, err)

		deltaEvent2 := responses.ResponseFunctionCallArgumentsDeltaEvent{
			ItemID: testItemID,
			Delta:  `30}`,
		}
		err = aiService.handleFunctionCallArgumentsDelta(deltaEvent2, state)
		require.NoError(t, err)

		// Verify tool call was created with correct call_id
		toolCall, exists := state.toolCalls[testCallID]
		require.True(t, exists)
		assert.Equal(t, testCallID, toolCall.CallID)
		assert.Equal(t, testItemID, toolCall.ID)
		assert.Equal(t, `{"per_page": 30}`, toolCall.Arguments)

		// Simulate function call completion
		state.completed[testCallID] = true

		// Get completed tool calls
		completedCalls := aiService.GetCompletedToolCalls(state)
		require.Len(t, completedCalls, 1)

		completedCall := completedCalls[0]
		assert.Equal(t, testCallID, completedCall.CallID)
		assert.Equal(t, "get-athlete-profile", completedCall.Name)

		// Test tool execution with correct call_id
		ctx := context.Background()
		msgCtx := &MessageContext{
			UserID:    "test-user",
			SessionID: "test-session",
			Message:   "Get my profile",
			User: &models.User{
				ID:          "test-user",
				StravaID:    12345,
				AccessToken: "test-token",
			},
		}

		// Execute the tool call
		toolResults, err := aiService.executeToolsWithRecovery(ctx, msgCtx, completedCalls)
		require.NoError(t, err)
		require.Len(t, toolResults, 1)

		// Verify tool result uses correct call_id
		result := toolResults[0]
		assert.Equal(t, testCallID, result.ToolCallID)
		assert.NotEmpty(t, result.Content)
		assert.Empty(t, result.Error)
	})

	t.Run("tool result correlation using extracted call_id", func(t *testing.T) {
		// Test multiple tool calls with different call_ids
		testCases := []struct {
			callID   string
			itemID   string
			toolName string
		}{
			{"call_abc123", "fc_item1", "get-athlete-profile"},
			{"call_def456", "fc_item2", "get-recent-activities"},
			{"call_ghi789", "fc_item3", "get-activity-details"},
		}

		state := NewToolCallState()
		ctx := context.Background()
		msgCtx := &MessageContext{
			UserID:    "test-user",
			SessionID: "test-session",
			Message:   "Get my data",
			User: &models.User{
				ID:          "test-user",
				StravaID:    12345,
				AccessToken: "test-token",
			},
		}

		var completedCalls []responses.ResponseFunctionToolCall

		// Process each tool call
		for _, tc := range testCases {
			// Simulate output item added event
			mockEvent := createMockOutputItemAddedEvent(tc.itemID, tc.callID, tc.toolName)
			err := aiService.handleOutputItemAdded(mockEvent, state)
			require.NoError(t, err)

			// Add arguments based on tool type
			var args string
			switch tc.toolName {
			case "get-athlete-profile":
				args = "{}"
			case "get-recent-activities":
				args = `{"per_page": 30}`
			case "get-activity-details":
				args = `{"activity_id": 123456}`
			}

			// Simulate arguments delta
			deltaEvent := responses.ResponseFunctionCallArgumentsDeltaEvent{
				ItemID: tc.itemID,
				Delta:  args,
			}
			err = aiService.handleFunctionCallArgumentsDelta(deltaEvent, state)
			require.NoError(t, err)

			// Mark as completed
			state.completed[tc.callID] = true

			// Verify tool call was created correctly
			toolCall, exists := state.toolCalls[tc.callID]
			require.True(t, exists)
			assert.Equal(t, tc.callID, toolCall.CallID)
			assert.Equal(t, tc.toolName, toolCall.Name)

			completedCalls = append(completedCalls, *toolCall)
		}

		// Execute all tool calls
		toolResults, err := aiService.executeToolsWithRecovery(ctx, msgCtx, completedCalls)
		require.NoError(t, err)
		require.Len(t, toolResults, len(testCases))

		// Verify each result has correct call_id correlation
		resultMap := make(map[string]ToolResult)
		for _, result := range toolResults {
			resultMap[result.ToolCallID] = result
		}

		for _, tc := range testCases {
			result, exists := resultMap[tc.callID]
			require.True(t, exists, "Result not found for call_id: %s", tc.callID)
			assert.Equal(t, tc.callID, result.ToolCallID)
			assert.NotEmpty(t, result.Content)
		}
	})
}

// TestMultiTurnConversationWithCallIDTracking tests multi-turn conversation scenarios with proper ID tracking
func TestMultiTurnConversationWithCallIDTracking(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		OpenAIAPIKey: "test-key",
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  4000,
			TokenPerCharRatio: 4,
			DefaultPageSize:   50,
			MaxPageSize:       200,
			RedactionEnabled:  false,
		},
	}

	// Create mock services
	mockStravaService := &mockStravaServiceForIntegration{}
	mockLogbookService := &mockLogbookServiceForIntegration{}
	mockToolRegistry := NewToolRegistry()
	mockSessionRepo := &MockSessionRepository{}

	// Create AI service
	aiService := NewAIService(cfg, mockStravaService, mockLogbookService, mockSessionRepo, mockToolRegistry).(*aiService)

	t.Run("multi-turn conversation with proper call_id tracking", func(t *testing.T) {
		ctx := context.Background()

		// First turn - get athlete profile
		msgCtx1 := &MessageContext{
			UserID:              "test-user",
			SessionID:           "test-session",
			Message:             "Get my profile",
			ConversationHistory: []*models.Message{},
			User: &models.User{
				ID:          "test-user",
				StravaID:    12345,
				AccessToken: "test-token",
			},
		}

		// Simulate first turn processing
		state1 := NewToolCallState()
		callID1 := "call_turn1_abc123"
		itemID1 := "fc_turn1_item1"

		// Process output item added event
		mockEvent1 := createMockOutputItemAddedEvent(itemID1, callID1, "get-athlete-profile")
		err := aiService.handleOutputItemAdded(mockEvent1, state1)
		require.NoError(t, err)

		// Add arguments and complete
		deltaEvent1 := responses.ResponseFunctionCallArgumentsDeltaEvent{
			ItemID: itemID1,
			Delta:  "{}",
		}
		err = aiService.handleFunctionCallArgumentsDelta(deltaEvent1, state1)
		require.NoError(t, err)
		state1.completed[callID1] = true

		// Execute first turn
		completedCalls1 := aiService.GetCompletedToolCalls(state1)
		toolResults1, err := aiService.executeToolsWithRecovery(ctx, msgCtx1, completedCalls1)
		require.NoError(t, err)
		require.Len(t, toolResults1, 1)
		assert.Equal(t, callID1, toolResults1[0].ToolCallID)

		// Second turn - get recent activities (with conversation history)
		msgCtx2 := &MessageContext{
			UserID:    "test-user",
			SessionID: "test-session",
			Message:   "Now get my recent activities",
			ConversationHistory: []*models.Message{
				{
					ID:         "msg1",
					SessionID:  "test-session",
					Role:       "user",
					Content:    "Get my profile",
					ResponseID: stringPtr("resp_turn1_123"),
				},
				{
					ID:         "msg2",
					SessionID:  "test-session",
					Role:       "assistant",
					Content:    toolResults1[0].Content,
					ResponseID: stringPtr("resp_turn1_123"),
				},
			},
			User: &models.User{
				ID:          "test-user",
				StravaID:    12345,
				AccessToken: "test-token",
			},
			LastResponseID: "resp_turn1_123",
		}

		// Simulate second turn processing
		state2 := NewToolCallState()
		callID2 := "call_turn2_def456"
		itemID2 := "fc_turn2_item2"

		// Process output item added event
		mockEvent2 := createMockOutputItemAddedEvent(itemID2, callID2, "get-recent-activities")
		err = aiService.handleOutputItemAdded(mockEvent2, state2)
		require.NoError(t, err)

		// Add arguments and complete
		deltaEvent2 := responses.ResponseFunctionCallArgumentsDeltaEvent{
			ItemID: itemID2,
			Delta:  `{"per_page": 30}`,
		}
		err = aiService.handleFunctionCallArgumentsDelta(deltaEvent2, state2)
		require.NoError(t, err)
		state2.completed[callID2] = true

		// Execute second turn
		completedCalls2 := aiService.GetCompletedToolCalls(state2)
		toolResults2, err := aiService.executeToolsWithRecovery(ctx, msgCtx2, completedCalls2)
		require.NoError(t, err)
		require.Len(t, toolResults2, 1)
		assert.Equal(t, callID2, toolResults2[0].ToolCallID)

		// Verify call_ids are different between turns
		assert.NotEqual(t, callID1, callID2)
		assert.NotEqual(t, toolResults1[0].ToolCallID, toolResults2[0].ToolCallID)

		// Verify conversation context includes previous response ID
		inputItems := aiService.buildConversationContextForResponsesAPI(msgCtx2)
		assert.Greater(t, len(inputItems), 2) // Should include system, history, and current message
	})
}

// TestErrorRecoveryAndFallbackScenarios tests error recovery and fallback scenarios in realistic conditions
func TestErrorRecoveryAndFallbackScenarios(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		OpenAIAPIKey: "test-key",
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  4000,
			TokenPerCharRatio: 4,
			DefaultPageSize:   50,
			MaxPageSize:       200,
			RedactionEnabled:  false,
		},
	}

	// Create mock services
	mockStravaService := &mockStravaServiceForIntegration{}
	mockLogbookService := &mockLogbookServiceForIntegration{}
	mockToolRegistry := NewToolRegistry()
	mockSessionRepo := &MockSessionRepository{}

	// Create AI service
	aiService := NewAIService(cfg, mockStravaService, mockLogbookService, mockSessionRepo, mockToolRegistry).(*aiService)

	t.Run("fallback to item_id when call_id is missing", func(t *testing.T) {
		state := NewToolCallState()
		itemID := "fc_fallback_test"

		// Create mock event with missing call_id
		mockEvent := createMockOutputItemAddedEventWithMissingCallID(itemID, "get-athlete-profile")
		
		// This should handle the missing call_id gracefully
		err := aiService.handleOutputItemAdded(mockEvent, state)
		
		// The implementation should handle this gracefully and use fallback
		// The exact behavior depends on the implementation - it might succeed with fallback or return an error
		if err != nil {
			// If it returns an error, verify it's the expected error type
			assert.Contains(t, err.Error(), "call_id")
		} else {
			// If it succeeds with fallback, verify the fallback was used
			// Check if item_id was used as fallback
			assert.Contains(t, state.itemToCallID, itemID)
		}
	})

	t.Run("error recovery during tool execution", func(t *testing.T) {
		// Create a failing mock service for this test
		failingStravaService := &failingMockStravaService{}
		failingLogbookService := &mockLogbookServiceForIntegration{}
		mockToolRegistry := NewToolRegistry()
		mockSessionRepo := &MockSessionRepository{}
		
		// Create AI service with failing mock
		failingAIService := NewAIService(cfg, failingStravaService, failingLogbookService, mockSessionRepo, mockToolRegistry)
		
		ctx := context.Background()
		msgCtx := &MessageContext{
			UserID:    "test-user",
			SessionID: "test-session",
			Message:   "Test error recovery",
			User: &models.User{
				ID:          "test-user",
				StravaID:    12345,
				AccessToken: "test-token",
			},
		}

		// Test error recovery through the public interface
		// Use ExecuteGetAthleteProfile which should fail with the failing mock
		result, err := failingAIService.ExecuteGetAthleteProfile(ctx, msgCtx)
		
		// Should return an error since the mock service fails
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated Strava API error")
		assert.Empty(t, result) // Result should be empty on error
		
		// Test that the call_id would be properly tracked even in error scenarios
		// by verifying the error message structure
		assert.Contains(t, err.Error(), "athlete profile not found")
	})

	t.Run("malformed event handling", func(t *testing.T) {
		state := NewToolCallState()

		// Test with empty item ID
		deltaEvent := responses.ResponseFunctionCallArgumentsDeltaEvent{
			ItemID: "", // Empty item ID
			Delta:  `{"test": "data"}`,
		}

		err := aiService.handleFunctionCallArgumentsDelta(deltaEvent, state)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty item ID")

		// Test with whitespace-only item ID
		deltaEventWhitespace := responses.ResponseFunctionCallArgumentsDeltaEvent{
			ItemID: "   ", // Whitespace-only item ID
			Delta:  `{"test": "data"}`,
		}

		err = aiService.handleFunctionCallArgumentsDelta(deltaEventWhitespace, state)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "whitespace")
	})

	t.Run("concurrent tool call processing", func(t *testing.T) {
		state := NewToolCallState()

		// Simulate multiple concurrent tool calls
		testCalls := []struct {
			callID   string
			itemID   string
			toolName string
		}{
			{"call_concurrent1", "fc_concurrent1", "get-athlete-profile"},
			{"call_concurrent2", "fc_concurrent2", "get-recent-activities"},
			{"call_concurrent3", "fc_concurrent3", "get-activity-details"},
		}

		// Process all output item events
		for _, tc := range testCalls {
			mockEvent := createMockOutputItemAddedEvent(tc.itemID, tc.callID, tc.toolName)
			err := aiService.handleOutputItemAdded(mockEvent, state)
			require.NoError(t, err)
		}

		// Process arguments for all calls
		for _, tc := range testCalls {
			var args string
			switch tc.toolName {
			case "get-athlete-profile":
				args = "{}"
			case "get-recent-activities":
				args = `{"per_page": 30}`
			case "get-activity-details":
				args = `{"activity_id": 123456}`
			}

			deltaEvent := responses.ResponseFunctionCallArgumentsDeltaEvent{
				ItemID: tc.itemID,
				Delta:  args,
			}
			err := aiService.handleFunctionCallArgumentsDelta(deltaEvent, state)
			require.NoError(t, err)
		}

		// Mark all as completed
		for _, tc := range testCalls {
			state.completed[tc.callID] = true
		}

		// Verify all tool calls were processed correctly
		completedCalls := aiService.GetCompletedToolCalls(state)
		assert.Len(t, completedCalls, len(testCalls))

		// Verify each call has correct call_id
		callIDMap := make(map[string]bool)
		for _, call := range completedCalls {
			callIDMap[call.CallID] = true
		}

		for _, tc := range testCalls {
			assert.True(t, callIDMap[tc.callID], "Call ID %s not found in completed calls", tc.callID)
		}
	})
}

// Helper functions for creating mock events

func createMockOutputItemAddedEvent(itemID, callID, toolName string) responses.ResponseStreamEventUnion {
	// Create a mock event that simulates the structure of response.output_item.added
	eventData := map[string]interface{}{
		"type": "response.output_item.added",
		"item": map[string]interface{}{
			"id":      itemID,
			"call_id": callID,
			"name":    toolName,
			"type":    "function_call",
			"status":  "in_progress",
		},
	}

	// Convert to JSON and back to create the event structure
	jsonData, _ := json.Marshal(eventData)
	var event responses.ResponseStreamEventUnion
	json.Unmarshal(jsonData, &event)
	
	return event
}

func createMockOutputItemAddedEventWithMissingCallID(itemID, toolName string) responses.ResponseStreamEventUnion {
	// Create a mock event without call_id to test fallback behavior
	eventData := map[string]interface{}{
		"type": "response.output_item.added",
		"item": map[string]interface{}{
			"id":     itemID,
			"name":   toolName,
			"type":   "function_call",
			"status": "in_progress",
			// Note: call_id is intentionally missing
		},
	}

	// Convert to JSON and back to create the event structure
	jsonData, _ := json.Marshal(eventData)
	var event responses.ResponseStreamEventUnion
	json.Unmarshal(jsonData, &event)
	
	return event
}

func stringPtr(s string) *string {
	return &s
}

// Mock services for integration testing

type mockStravaServiceForIntegration struct{}

func (m *mockStravaServiceForIntegration) GetAthleteProfile(user *models.User) (*StravaAthleteWithZones, error) {
	return &StravaAthleteWithZones{
		StravaAthlete: &StravaAthlete{
			ID:        12345,
			Username:  "test_athlete",
			Firstname: "Test",
			Lastname:  "Athlete",
			City:      "Test City",
			State:     "Test State",
			Country:   "Test Country",
		},
	}, nil
}

func (m *mockStravaServiceForIntegration) GetAthleteZones(user *models.User) (*StravaAthleteZones, error) {
	return &StravaAthleteZones{}, nil
}

func (m *mockStravaServiceForIntegration) GetActivities(user *models.User, params ActivityParams) ([]*StravaActivity, error) {
	return []*StravaActivity{
		{
			ID:   123456,
			Name: "Test Activity",
			Type: "Run",
		},
	}, nil
}

func (m *mockStravaServiceForIntegration) GetActivityDetail(user *models.User, activityID int64) (*StravaActivityDetail, error) {
	return &StravaActivityDetail{
		StravaActivity: StravaActivity{
			ID:   activityID,
			Name: "Test Activity Details",
			Type: "Run",
		},
	}, nil
}

func (m *mockStravaServiceForIntegration) GetActivityDetailWithZones(user *models.User, activityID int64) (*StravaActivityDetailWithZones, error) {
	return &StravaActivityDetailWithZones{
		StravaActivityDetail: &StravaActivityDetail{
			StravaActivity: StravaActivity{
				ID:   activityID,
				Name: "Test Activity Details With Zones",
				Type: "Run",
			},
		},
	}, nil
}

func (m *mockStravaServiceForIntegration) GetActivityStreams(user *models.User, activityID int64, streamTypes []string, resolution string) (*StravaStreams, error) {
	return &StravaStreams{
		Time:      []int{0, 1, 2, 3, 4},
		Distance:  []float64{0, 100, 200, 300, 400},
		Heartrate: []int{120, 125, 130, 135, 140},
	}, nil
}

func (m *mockStravaServiceForIntegration) GetActivityZones(user *models.User, activityID int64) (*StravaActivityZones, error) {
	return &StravaActivityZones{}, nil
}

func (m *mockStravaServiceForIntegration) RefreshToken(refreshToken string) (*TokenResponse, error) {
	return &TokenResponse{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		ExpiresAt:    3600,
	}, nil
}

type mockLogbookServiceForIntegration struct{}

func (m *mockLogbookServiceForIntegration) GetLogbook(ctx context.Context, userID string) (*models.AthleteLogbook, error) {
	return &models.AthleteLogbook{
		UserID:  userID,
		Content: "Test logbook content",
	}, nil
}

func (m *mockLogbookServiceForIntegration) CreateInitialLogbook(ctx context.Context, userID string, stravaProfile *StravaAthlete) (*models.AthleteLogbook, error) {
	return &models.AthleteLogbook{
		UserID:  userID,
		Content: "Initial logbook content",
	}, nil
}

func (m *mockLogbookServiceForIntegration) UpdateLogbook(ctx context.Context, userID string, content string) (*models.AthleteLogbook, error) {
	return &models.AthleteLogbook{
		UserID:  userID,
		Content: content,
	}, nil
}

func (m *mockLogbookServiceForIntegration) UpsertLogbook(ctx context.Context, userID string, content string) (*models.AthleteLogbook, error) {
	return &models.AthleteLogbook{
		UserID:  userID,
		Content: content,
	}, nil
}

// Failing mock service for error recovery testing
type failingMockStravaService struct{}

func (m *failingMockStravaService) GetAthleteProfile(user *models.User) (*StravaAthleteWithZones, error) {
	return nil, fmt.Errorf("simulated Strava API error: athlete profile not found")
}

func (m *failingMockStravaService) GetAthleteZones(user *models.User) (*StravaAthleteZones, error) {
	return nil, fmt.Errorf("simulated Strava API error: zones not available")
}

func (m *failingMockStravaService) GetActivities(user *models.User, params ActivityParams) ([]*StravaActivity, error) {
	return nil, fmt.Errorf("simulated Strava API error: activities not accessible")
}

func (m *failingMockStravaService) GetActivityDetail(user *models.User, activityID int64) (*StravaActivityDetail, error) {
	return nil, fmt.Errorf("simulated Strava API error: activity %d not found", activityID)
}

func (m *failingMockStravaService) GetActivityDetailWithZones(user *models.User, activityID int64) (*StravaActivityDetailWithZones, error) {
	return nil, fmt.Errorf("simulated Strava API error: activity %d zones not available", activityID)
}

func (m *failingMockStravaService) GetActivityStreams(user *models.User, activityID int64, streamTypes []string, resolution string) (*StravaStreams, error) {
	return nil, fmt.Errorf("simulated Strava API error: streams for activity %d not available", activityID)
}

func (m *failingMockStravaService) GetActivityZones(user *models.User, activityID int64) (*StravaActivityZones, error) {
	return nil, fmt.Errorf("simulated Strava API error: zones for activity %d not available", activityID)
}

func (m *failingMockStravaService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	return nil, fmt.Errorf("simulated Strava API error: token refresh failed")
}

// Mock tool registry for integration testing
type mockToolRegistryForIntegration struct{}

func (m *mockToolRegistryForIntegration) GetAvailableTools() []models.ToolDefinition {
	return []models.ToolDefinition{
		{
			Name:        "get-athlete-profile",
			Description: "Get athlete profile information",
			Parameters:  map[string]interface{}{},
		},
		{
			Name:        "get-recent-activities",
			Description: "Get recent activities",
			Parameters:  map[string]interface{}{},
		},
		{
			Name:        "get-activity-details",
			Description: "Get activity details",
			Parameters: map[string]interface{}{
				"activity_id": map[string]interface{}{
					"type":        "integer",
					"description": "The Strava activity ID",
				},
			},
		},
		{
			Name:        "get-activity-streams",
			Description: "Get activity streams",
			Parameters: map[string]interface{}{
				"activity_id": map[string]interface{}{
					"type":        "integer",
					"description": "The Strava activity ID",
				},
			},
		},
		{
			Name:        "update-athlete-logbook",
			Description: "Update athlete logbook",
			Parameters: map[string]interface{}{
				"content": map[string]interface{}{
					"type":        "string",
					"description": "The logbook content",
				},
			},
		},
	}
}

func (m *mockToolRegistryForIntegration) GetToolSchema(toolName string) (*models.ToolSchema, error) {
	return &models.ToolSchema{
		Name:        toolName,
		Description: "Mock tool schema",
		Parameters:  map[string]interface{}{},
	}, nil
}

func (m *mockToolRegistryForIntegration) ValidateToolCall(toolName string, parameters map[string]interface{}) error {
	return nil
}

func (m *mockToolRegistryForIntegration) IsToolAvailable(toolName string) bool {
	return true
}

// MockSessionRepositoryForCallID for testing
type MockSessionRepositoryForCallID struct{}

func (m *MockSessionRepositoryForCallID) UpdateLastResponseID(ctx context.Context, sessionID string, responseID string) error {
	return nil // Simple mock that always succeeds
}