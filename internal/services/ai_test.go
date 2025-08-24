package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock OpenAI Client
type MockOpenAIClient struct {
	mock.Mock
}

func (m *MockOpenAIClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(openai.ChatCompletionResponse), args.Error(1)
}

func (m *MockOpenAIClient) CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*openai.ChatCompletionStream), args.Error(1)
}

// Mock Strava Service
type MockStravaService struct {
	mock.Mock
}

func (m *MockStravaService) GetAthleteProfile(accessToken string) (*StravaAthlete, error) {
	args := m.Called(accessToken)
	return args.Get(0).(*StravaAthlete), args.Error(1)
}

func (m *MockStravaService) GetActivities(accessToken string, params ActivityParams) ([]*StravaActivity, error) {
	args := m.Called(accessToken, params)
	return args.Get(0).([]*StravaActivity), args.Error(1)
}

func (m *MockStravaService) GetActivityDetail(accessToken string, activityID int64) (*StravaActivityDetail, error) {
	args := m.Called(accessToken, activityID)
	return args.Get(0).(*StravaActivityDetail), args.Error(1)
}

func (m *MockStravaService) GetActivityStreams(accessToken string, activityID int64, streamTypes []string, resolution string) (*StravaStreams, error) {
	args := m.Called(accessToken, activityID, streamTypes, resolution)
	return args.Get(0).(*StravaStreams), args.Error(1)
}

func (m *MockStravaService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	args := m.Called(refreshToken)
	return args.Get(0).(*TokenResponse), args.Error(1)
}

// Mock Logbook Service
type MockLogbookService struct {
	mock.Mock
}

func (m *MockLogbookService) GetLogbook(ctx context.Context, userID string) (*models.AthleteLogbook, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.AthleteLogbook), args.Error(1)
}

func (m *MockLogbookService) CreateInitialLogbook(ctx context.Context, userID string, stravaProfile *StravaAthlete) (*models.AthleteLogbook, error) {
	args := m.Called(ctx, userID, stravaProfile)
	return args.Get(0).(*models.AthleteLogbook), args.Error(1)
}

func (m *MockLogbookService) UpdateLogbook(ctx context.Context, userID string, content string) (*models.AthleteLogbook, error) {
	args := m.Called(ctx, userID, content)
	return args.Get(0).(*models.AthleteLogbook), args.Error(1)
}

func (m *MockLogbookService) UpsertLogbook(ctx context.Context, userID string, content string) (*models.AthleteLogbook, error) {
	args := m.Called(ctx, userID, content)
	return args.Get(0).(*models.AthleteLogbook), args.Error(1)
}

// Test AI Service with mocked dependencies
type testAIService struct {
	*aiService
	mockStrava  *MockStravaService
	mockLogbook *MockLogbookService
}

func setupTestAIService() *testAIService {
	cfg := &config.Config{
		OpenAIAPIKey: "test-key",
	}
	
	mockStrava := &MockStravaService{}
	mockLogbook := &MockLogbookService{}
	
	service := &aiService{
		stravaService:  mockStrava,
		logbookService: mockLogbook,
		config:         cfg,
	}
	
	return &testAIService{
		aiService:   service,
		mockStrava:  mockStrava,
		mockLogbook: mockLogbook,
	}
}

func TestAIService_PrepareMessages(t *testing.T) {
	service := setupTestAIService()
	
	// Test data
	user := &models.User{
		ID:          "user-123",
		FirstName:   "John",
		LastName:    "Doe",
		AccessToken: "token-123",
	}
	
	logbook := &models.AthleteLogbook{
		UserID:  "user-123",
		Content: `{"personal_info": {"name": "John Doe", "age": 30}}`,
	}
	
	conversationHistory := []*models.Message{
		{
			ID:        "msg-1",
			SessionID: "session-123",
			Role:      "user",
			Content:   "Hello, I'm new to running",
			CreatedAt: time.Now(),
		},
		{
			ID:        "msg-2",
			SessionID: "session-123",
			Role:      "assistant",
			Content:   "Welcome! I'd be happy to help you get started with running.",
			CreatedAt: time.Now(),
		},
	}
	
	msgCtx := &MessageContext{
		UserID:              "user-123",
		SessionID:           "session-123",
		Message:             "What should I focus on as a beginner?",
		ConversationHistory: conversationHistory,
		AthleteLogbook:      logbook,
		User:                user,
	}
	
	messages := service.prepareMessages(msgCtx)
	
	// Verify message structure
	assert.Len(t, messages, 4) // system + 2 history + current
	assert.Equal(t, openai.ChatMessageRoleSystem, messages[0].Role)
	assert.Contains(t, messages[0].Content, "Bodda")
	assert.Contains(t, messages[0].Content, logbook.Content)
	
	// Verify conversation history
	assert.Equal(t, openai.ChatMessageRoleUser, messages[1].Role)
	assert.Equal(t, "Hello, I'm new to running", messages[1].Content)
	assert.Equal(t, openai.ChatMessageRoleAssistant, messages[2].Role)
	assert.Equal(t, "Welcome! I'd be happy to help you get started with running.", messages[2].Content)
	
	// Verify current message
	assert.Equal(t, openai.ChatMessageRoleUser, messages[3].Role)
	assert.Equal(t, "What should I focus on as a beginner?", messages[3].Content)
}

func TestAIService_BuildSystemPrompt(t *testing.T) {
	service := setupTestAIService()
	
	t.Run("with logbook", func(t *testing.T) {
		logbook := &models.AthleteLogbook{
			Content: `{"personal_info": {"name": "John Doe"}}`,
		}
		
		msgCtx := &MessageContext{
			AthleteLogbook: logbook,
		}
		
		prompt := service.buildSystemPrompt(msgCtx)
		
		assert.Contains(t, prompt, "Bodda")
		assert.Contains(t, prompt, "AI-powered running and cycling coach")
		assert.Contains(t, prompt, "Current Athlete Logbook:")
		assert.Contains(t, prompt, logbook.Content)
	})
	
	t.Run("without logbook", func(t *testing.T) {
		msgCtx := &MessageContext{}
		
		prompt := service.buildSystemPrompt(msgCtx)
		
		assert.Contains(t, prompt, "Bodda")
		assert.Contains(t, prompt, "No athlete logbook exists yet")
		assert.Contains(t, prompt, "update-athlete-logbook tool")
	})
}

func TestAIService_GetAvailableTools(t *testing.T) {
	service := setupTestAIService()
	
	tools := service.getAvailableTools()
	
	assert.Len(t, tools, 5)
	
	toolNames := make([]string, len(tools))
	for i, tool := range tools {
		toolNames[i] = tool.Function.Name
	}
	
	expectedTools := []string{
		"get-athlete-profile",
		"get-recent-activities",
		"get-activity-details",
		"get-activity-streams",
		"update-athlete-logbook",
	}
	
	for _, expected := range expectedTools {
		assert.Contains(t, toolNames, expected)
	}
}

func TestAIService_ExecuteGetAthleteProfile(t *testing.T) {
	service := setupTestAIService()
	ctx := context.Background()
	
	expectedProfile := &StravaAthlete{
		ID:        123456,
		Firstname: "John",
		Lastname:  "Doe",
		City:      "San Francisco",
		State:     "CA",
		Country:   "USA",
		Sex:       "M",
		Weight:    70.5,
		FTP:       250,
	}
	
	user := &models.User{
		AccessToken: "test-token",
	}
	
	msgCtx := &MessageContext{
		User: user,
	}
	
	service.mockStrava.On("GetAthleteProfile", "test-token").Return(expectedProfile, nil)
	
	result, err := service.executeGetAthleteProfile(ctx, msgCtx)
	
	assert.NoError(t, err)
	assert.Equal(t, expectedProfile, result)
	service.mockStrava.AssertExpectations(t)
}

func TestAIService_ExecuteGetRecentActivities(t *testing.T) {
	service := setupTestAIService()
	ctx := context.Background()
	
	expectedActivities := []*StravaActivity{
		{
			ID:           987654321,
			Name:         "Morning Run",
			Distance:     5000.0,
			MovingTime:   1800,
			Type:         "Run",
			StartDate:    "2024-01-15T08:00:00Z",
			AverageSpeed: 2.78,
		},
		{
			ID:           987654322,
			Name:         "Evening Bike Ride",
			Distance:     20000.0,
			MovingTime:   3600,
			Type:         "Ride",
			StartDate:    "2024-01-14T18:00:00Z",
			AverageSpeed: 5.56,
		},
	}
	
	user := &models.User{
		AccessToken: "test-token",
	}
	
	msgCtx := &MessageContext{
		User: user,
	}
	
	expectedParams := ActivityParams{
		PerPage: 30,
	}
	
	service.mockStrava.On("GetActivities", "test-token", expectedParams).Return(expectedActivities, nil)
	
	result, err := service.executeGetRecentActivities(ctx, msgCtx, 30)
	
	assert.NoError(t, err)
	assert.Equal(t, expectedActivities, result)
	service.mockStrava.AssertExpectations(t)
}

func TestAIService_ExecuteGetActivityDetails(t *testing.T) {
	service := setupTestAIService()
	ctx := context.Background()
	
	expectedDetail := &StravaActivityDetail{
		StravaActivity: StravaActivity{
			ID:           987654321,
			Name:         "Morning Run",
			Distance:     5000.0,
			MovingTime:   1800,
			Type:         "Run",
			AverageSpeed: 2.78,
		},
		Description: "Great morning run in the park",
		Calories:    350.5,
	}
	
	user := &models.User{
		AccessToken: "test-token",
	}
	
	msgCtx := &MessageContext{
		User: user,
	}
	
	service.mockStrava.On("GetActivityDetail", "test-token", int64(987654321)).Return(expectedDetail, nil)
	
	result, err := service.executeGetActivityDetails(ctx, msgCtx, 987654321)
	
	assert.NoError(t, err)
	assert.Equal(t, expectedDetail, result)
	service.mockStrava.AssertExpectations(t)
}

func TestAIService_ExecuteGetActivityStreams(t *testing.T) {
	service := setupTestAIService()
	ctx := context.Background()
	
	expectedStreams := &StravaStreams{
		Time:      []int{0, 30, 60, 90},
		Distance:  []float64{0, 100, 200, 300},
		Heartrate: []int{120, 130, 140, 135},
		Watts:     []int{200, 220, 240, 230},
	}
	
	user := &models.User{
		AccessToken: "test-token",
	}
	
	msgCtx := &MessageContext{
		User: user,
	}
	
	streamTypes := []string{"time", "distance", "heartrate", "watts"}
	resolution := "medium"
	
	service.mockStrava.On("GetActivityStreams", "test-token", int64(987654321), streamTypes, resolution).Return(expectedStreams, nil)
	
	result, err := service.executeGetActivityStreams(ctx, msgCtx, 987654321, streamTypes, resolution)
	
	assert.NoError(t, err)
	assert.Equal(t, expectedStreams, result)
	service.mockStrava.AssertExpectations(t)
}

func TestAIService_ExecuteUpdateAthleteLogbook(t *testing.T) {
	service := setupTestAIService()
	ctx := context.Background()
	
	expectedLogbook := &models.AthleteLogbook{
		ID:      "logbook-123",
		UserID:  "user-123",
		Content: `{"personal_info":{"name":"John Doe","age":30,"weight":70.5},"training_data":{"ftp":250,"max_heart_rate":190,"weekly_volume":"30-40 km per week"}}`,
	}
	
	msgCtx := &MessageContext{
		UserID: "user-123",
	}
	
	// Mock the UpdateLogbook call
	service.mockLogbook.On("UpdateLogbook", ctx, "user-123", mock.AnythingOfType("string")).Return(expectedLogbook, nil)
	
	result, err := service.executeUpdateAthleteLogbook(ctx, msgCtx, `{"personal_info":{"name":"John Doe","age":30}}`)
	
	assert.NoError(t, err)
	assert.Equal(t, expectedLogbook, result)
	service.mockLogbook.AssertExpectations(t)
}

func TestAIService_ExecuteUpdateAthleteLogbook_CreateNew(t *testing.T) {
	service := setupTestAIService()
	ctx := context.Background()
	
	expectedLogbook := &models.AthleteLogbook{
		ID:      "logbook-456",
		UserID:  "user-456",
		Content: `{"personal_info":{"name":"Jane Doe","age":25}}`,
	}
	
	msgCtx := &MessageContext{
		UserID: "user-456",
	}
	
	// Mock UpdateLogbook to return "not found" error, then UpsertLogbook to succeed
	service.mockLogbook.On("UpdateLogbook", ctx, "user-456", mock.AnythingOfType("string")).Return((*models.AthleteLogbook)(nil), fmt.Errorf("logbook not found for user user-456"))
	service.mockLogbook.On("UpsertLogbook", ctx, "user-456", mock.AnythingOfType("string")).Return(expectedLogbook, nil)
	
	result, err := service.executeUpdateAthleteLogbook(ctx, msgCtx, `{"personal_info":{"name":"Jane Doe","age":25}}`)
	
	assert.NoError(t, err)
	assert.Equal(t, expectedLogbook, result)
	service.mockLogbook.AssertExpectations(t)
}

func TestAIService_ExecuteTools(t *testing.T) {
	service := setupTestAIService()
	ctx := context.Background()
	
	user := &models.User{
		AccessToken: "test-token",
	}
	
	msgCtx := &MessageContext{
		UserID: "user-123",
		User:   user,
	}
	
	// Mock Strava service calls
	expectedProfile := &StravaAthlete{
		ID:        123456,
		Firstname: "John",
		Lastname:  "Doe",
	}
	service.mockStrava.On("GetAthleteProfile", "test-token").Return(expectedProfile, nil)
	
	toolCalls := []openai.ToolCall{
		{
			ID:   "call-1",
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionCall{
				Name:      "get-athlete-profile",
				Arguments: "{}",
			},
		},
	}
	
	results, err := service.executeTools(ctx, msgCtx, toolCalls)
	
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "call-1", results[0].ToolCallID)
	assert.Empty(t, results[0].Error)
	assert.Contains(t, results[0].Content, "John")
	assert.Contains(t, results[0].Content, "Doe")
	
	service.mockStrava.AssertExpectations(t)
}

func TestAIService_ExecuteTools_InvalidTool(t *testing.T) {
	service := setupTestAIService()
	ctx := context.Background()
	
	msgCtx := &MessageContext{
		UserID: "user-123",
	}
	
	toolCalls := []openai.ToolCall{
		{
			ID:   "call-1",
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionCall{
				Name:      "invalid-tool",
				Arguments: "{}",
			},
		},
	}
	
	results, err := service.executeTools(ctx, msgCtx, toolCalls)
	
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "call-1", results[0].ToolCallID)
	assert.Equal(t, "unknown tool", results[0].Error)
	assert.Contains(t, results[0].Content, "Unknown tool: invalid-tool")
}

func TestAIService_ExecuteTools_InvalidArguments(t *testing.T) {
	service := setupTestAIService()
	ctx := context.Background()
	
	msgCtx := &MessageContext{
		UserID: "user-123",
	}
	
	toolCalls := []openai.ToolCall{
		{
			ID:   "call-1",
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionCall{
				Name:      "get-activity-details",
				Arguments: "invalid json",
			},
		},
	}
	
	results, err := service.executeTools(ctx, msgCtx, toolCalls)
	
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "call-1", results[0].ToolCallID)
	assert.NotEmpty(t, results[0].Error)
	assert.Contains(t, results[0].Content, "Error parsing arguments")
}

// Integration test for message context preparation
func TestAIService_MessageContextIntegration(t *testing.T) {
	service := setupTestAIService()
	
	// Create a realistic message context
	user := &models.User{
		ID:          "user-123",
		FirstName:   "Alice",
		LastName:    "Runner",
		AccessToken: "strava-token-123",
	}
	
	logbook := &models.AthleteLogbook{
		UserID: "user-123",
		Content: `{
			"personal_info": {
				"name": "Alice Runner",
				"age": 28,
				"gender": "F",
				"weight": 60.0
			},
			"training_data": {
				"ftp": 220,
				"max_heart_rate": 185,
				"weekly_volume": "40-50 km per week"
			},
			"goals": {
				"short_term": ["Complete a 10K race", "Improve 5K time"],
				"long_term": ["Run a marathon"]
			}
		}`,
	}
	
	conversationHistory := []*models.Message{
		{
			Role:    "user",
			Content: "Hi, I want to improve my running performance",
		},
		{
			Role:    "assistant",
			Content: "Great! I can help you with that. Let me look at your recent activities to understand your current training.",
		},
	}
	
	msgCtx := &MessageContext{
		UserID:              "user-123",
		SessionID:           "session-456",
		Message:             "What should I focus on this week?",
		ConversationHistory: conversationHistory,
		AthleteLogbook:      logbook,
		User:                user,
	}
	
	// Test message preparation
	messages := service.prepareMessages(msgCtx)
	
	assert.Len(t, messages, 4) // system + 2 history + current
	
	// Verify system message contains logbook data
	systemMsg := messages[0]
	assert.Equal(t, openai.ChatMessageRoleSystem, systemMsg.Role)
	assert.Contains(t, systemMsg.Content, "Alice Runner")
	assert.Contains(t, systemMsg.Content, "marathon")
	assert.Contains(t, systemMsg.Content, "get-athlete-profile")
	
	// Verify conversation flow
	assert.Equal(t, "Hi, I want to improve my running performance", messages[1].Content)
	assert.Equal(t, "Great! I can help you with that. Let me look at your recent activities to understand your current training.", messages[2].Content)
	assert.Equal(t, "What should I focus on this week?", messages[3].Content)
}

// Test error handling in tool execution
func TestAIService_ToolExecutionErrorHandling(t *testing.T) {
	service := setupTestAIService()
	ctx := context.Background()
	
	user := &models.User{
		AccessToken: "test-token",
	}
	
	msgCtx := &MessageContext{
		UserID: "user-123",
		User:   user,
	}
	
	// Mock Strava service to return an error
	service.mockStrava.On("GetAthleteProfile", "test-token").Return((*StravaAthlete)(nil), fmt.Errorf("Strava API error"))
	
	toolCalls := []openai.ToolCall{
		{
			ID:   "call-1",
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionCall{
				Name:      "get-athlete-profile",
				Arguments: "{}",
			},
		},
	}
	
	results, err := service.executeTools(ctx, msgCtx, toolCalls)
	
	assert.NoError(t, err) // executeTools should not return error, but populate result.Error
	assert.Len(t, results, 1)
	assert.Equal(t, "call-1", results[0].ToolCallID)
	assert.Contains(t, results[0].Error, "Strava API error")
	assert.Contains(t, results[0].Content, "Error getting athlete profile")
	
	service.mockStrava.AssertExpectations(t)
}