package services

import (
	"context"
	"testing"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"
)

// MockSessionRepositoryForToolExecutor for testing
type MockSessionRepositoryForToolExecutor struct{}

func (m *MockSessionRepositoryForToolExecutor) UpdateLastResponseID(ctx context.Context, sessionID string, responseID string) error {
	return nil // Simple mock that always succeeds
}

func TestToolExecutorIntegrationWithAIService(t *testing.T) {
	// Skip if no OpenAI API key is available
	cfg := &config.Config{
		OpenAIAPIKey: "test-key", // Mock key for testing
	}

	// Create mock services for testing
	mockStravaService := &mockStravaServiceForToolExecutor{}
	mockLogbookService := &mockLogbookServiceForToolExecutor{}

	// Create tool registry first
	mockToolRegistry := NewToolRegistry()
	mockSessionRepo := &MockSessionRepositoryForToolExecutor{}
	
	// Create AI service (this will use mock services)
	aiService := NewAIService(cfg, mockStravaService, mockLogbookService, mockSessionRepo, mockToolRegistry)

	// Create tool registry and executor
	registry := NewToolRegistryWithAIService(aiService)
	executor := NewToolExecutor(aiService, registry)

	// Test context
	ctx := context.Background()
	msgCtx := &MessageContext{
		UserID:    "test-user",
		SessionID: "test-session",
		Message:   "test message",
		User: &models.User{
			ID:          "test-user",
			StravaID:    12345,
			AccessToken: "test-token",
		},
	}

	// Test that the executor can list available tools
	availableTools := registry.GetAvailableTools()
	if len(availableTools) == 0 {
		t.Error("Expected at least one available tool")
	}

	// Verify expected tools are available
	expectedTools := []string{
		"get-athlete-profile",
		"get-recent-activities", 
		"get-activity-details",
		"get-activity-streams",
		"update-athlete-logbook",
	}

	toolMap := make(map[string]bool)
	for _, tool := range availableTools {
		toolMap[tool.Name] = true
	}

	for _, expectedTool := range expectedTools {
		if !toolMap[expectedTool] {
			t.Errorf("Expected tool '%s' not found in available tools", expectedTool)
		}
	}

	// Test tool schema retrieval
	schema, err := registry.GetToolSchema("get-athlete-profile")
	if err != nil {
		t.Errorf("Failed to get tool schema: %v", err)
	}

	if schema.Name != "get-athlete-profile" {
		t.Errorf("Expected schema name 'get-athlete-profile', got '%s'", schema.Name)
	}

	// Test parameter validation
	err = registry.ValidateToolCall("get-activity-details", map[string]interface{}{})
	if err == nil {
		t.Error("Expected validation error for missing required parameter")
	}

	err = registry.ValidateToolCall("get-activity-details", map[string]interface{}{
		"activity_id": int64(123456),
	})
	if err != nil {
		t.Errorf("Expected no validation error, got: %v", err)
	}

	// Test timeout configuration
	options := &models.ExecutionOptions{
		Timeout: 1, // 1 second timeout
	}

	// This should work quickly with mock services
	result, err := executor.ExecuteToolWithOptions(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx, options)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got error: %s", result.Error)
	}

	if result.Duration < 0 {
		t.Error("Expected non-negative execution duration")
	}

	// Test streaming mode
	streamingOptions := &models.ExecutionOptions{
		Streaming: true,
		Timeout:   5,
	}

	result, err = executor.ExecuteToolWithOptions(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx, streamingOptions)
	if err != nil {
		t.Errorf("Expected no error in streaming mode, got: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful streaming execution, got error: %s", result.Error)
	}
}

// Mock services for integration testing (with unique names to avoid conflicts)
type mockStravaServiceForToolExecutor struct{}

func (m *mockStravaServiceForToolExecutor) GetAthleteProfile(user *models.User) (*StravaAthleteWithZones, error) {
	return &StravaAthleteWithZones{
		StravaAthlete: &StravaAthlete{
			ID:        12345,
			Username:  "test-athlete",
			Firstname: "Test",
			Lastname:  "Athlete",
		},
	}, nil
}

func (m *mockStravaServiceForToolExecutor) GetAthleteZones(user *models.User) (*StravaAthleteZones, error) {
	return &StravaAthleteZones{}, nil
}

func (m *mockStravaServiceForToolExecutor) GetActivities(user *models.User, params ActivityParams) ([]*StravaActivity, error) {
	return []*StravaActivity{
		{
			ID:   123456,
			Name: "Test Activity",
			Type: "Run",
		},
	}, nil
}

func (m *mockStravaServiceForToolExecutor) GetActivityDetail(user *models.User, activityID int64) (*StravaActivityDetail, error) {
	return &StravaActivityDetail{
		StravaActivity: StravaActivity{
			ID:   activityID,
			Name: "Test Activity Detail",
			Type: "Run",
		},
	}, nil
}

func (m *mockStravaServiceForToolExecutor) GetActivityDetailWithZones(user *models.User, activityID int64) (*StravaActivityDetailWithZones, error) {
	return &StravaActivityDetailWithZones{
		StravaActivityDetail: &StravaActivityDetail{
			StravaActivity: StravaActivity{
				ID:   activityID,
				Name: "Test Activity Detail with Zones",
				Type: "Run",
			},
		},
		Zones: &StravaActivityZones{
			HeartRate: &StravaZoneDistribution{
				Type: "heartrate",
				Zones: []StravaZoneData{
					{Min: 0, Max: 150, Time: 1800},
					{Min: 150, Max: 170, Time: 600},
				},
			},
		},
	}, nil
}

func (m *mockStravaServiceForToolExecutor) GetActivityStreams(user *models.User, activityID int64, streamTypes []string, resolution string) (*StravaStreams, error) {
	return &StravaStreams{
		Time:      []int{0, 1, 2, 3, 4},
		Heartrate: []int{120, 125, 130, 135, 140},
	}, nil
}

func (m *mockStravaServiceForToolExecutor) GetActivityZones(user *models.User, activityID int64) (*StravaActivityZones, error) {
	return &StravaActivityZones{
		HeartRate: &StravaZoneDistribution{
			Type: "heartrate",
			Zones: []StravaZoneData{
				{Min: 0, Max: 150, Time: 1800},
				{Min: 150, Max: 170, Time: 600},
			},
		},
	}, nil
}

func (m *mockStravaServiceForToolExecutor) RefreshToken(refreshToken string) (*TokenResponse, error) {
	return &TokenResponse{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		ExpiresAt:    time.Now().Add(6 * time.Hour).Unix(),
	}, nil
}

type mockLogbookServiceForToolExecutor struct{}

func (m *mockLogbookServiceForToolExecutor) GetLogbook(ctx context.Context, userID string) (*models.AthleteLogbook, error) {
	return &models.AthleteLogbook{
		ID:      "test-logbook",
		UserID:  userID,
		Content: "Test logbook content",
	}, nil
}

func (m *mockLogbookServiceForToolExecutor) CreateInitialLogbook(ctx context.Context, userID string, stravaProfile *StravaAthlete) (*models.AthleteLogbook, error) {
	return &models.AthleteLogbook{
		ID:        "test-logbook",
		UserID:    userID,
		Content:   "Initial logbook content",
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockLogbookServiceForToolExecutor) UpdateLogbook(ctx context.Context, userID, content string) (*models.AthleteLogbook, error) {
	return &models.AthleteLogbook{
		ID:        "test-logbook",
		UserID:    userID,
		Content:   content,
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockLogbookServiceForToolExecutor) UpsertLogbook(ctx context.Context, userID, content string) (*models.AthleteLogbook, error) {
	return &models.AthleteLogbook{
		ID:        "test-logbook",
		UserID:    userID,
		Content:   content,
		UpdatedAt: time.Now(),
	}, nil
}