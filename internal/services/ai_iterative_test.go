package services

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"bodda/internal/models"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewIterativeProcessor(t *testing.T) {
	msgCtx := &MessageContext{
		UserID:    "test-user",
		SessionID: "test-session",
		Message:   "Test message",
	}

	var progressMessages []string
	progressCallback := func(message string) {
		progressMessages = append(progressMessages, message)
	}

	processor := NewIterativeProcessor(msgCtx, progressCallback)

	assert.NotNil(t, processor)
	assert.Equal(t, 5, processor.MaxRounds)
	assert.Equal(t, 0, processor.CurrentRound)
	assert.Equal(t, msgCtx, processor.Context)
	assert.NotNil(t, processor.ProgressCallback)
	assert.Empty(t, processor.ToolResults)
	assert.Empty(t, processor.Messages)
}

func TestIterativeProcessor_GetProgressMessage(t *testing.T) {
	processor := &IterativeProcessor{CurrentRound: 0}

	// Test that messages are provided for each round and are coaching-focused
	for round := 0; round <= 10; round++ {
		processor.CurrentRound = round
		result := processor.GetProgressMessage()

		// Ensure message is not empty
		assert.NotEmpty(t, result, "Round %d should provide a message", round)

		// Ensure no technical jargon
		assert.NotContains(t, result, "API", "Round %d should not contain API", round)
		assert.NotContains(t, result, "executing", "Round %d should not contain executing", round)
		assert.NotContains(t, result, "tool", "Round %d should not contain tool", round)
		assert.NotContains(t, result, "function", "Round %d should not contain function", round)

		// Ensure coaching language is present
		coachingWords := []string{"training", "analyzing", "reviewing", "looking", "data", "insights", "performance", "workout", "athletic", "activities", "patterns", "trends", "analysis", "final", "comprehensive", "putting", "together", "preparing", "recommendations"}
		hasCoachingWord := false
		for _, word := range coachingWords {
			if strings.Contains(strings.ToLower(result), strings.ToLower(word)) {
				hasCoachingWord = true
				break
			}
		}
		assert.True(t, hasCoachingWord, "Round %d message should contain coaching language: %s", round, result)
	}
}

func TestIterativeProcessor_ShouldContinue(t *testing.T) {
	tests := []struct {
		name         string
		currentRound int
		maxRounds    int
		hasToolCalls bool
		expected     bool
	}{
		{
			name:         "should continue with tool calls and rounds remaining",
			currentRound: 2,
			maxRounds:    5,
			hasToolCalls: true,
			expected:     true,
		},
		{
			name:         "should not continue without tool calls",
			currentRound: 2,
			maxRounds:    5,
			hasToolCalls: false,
			expected:     false,
		},
		{
			name:         "should not continue when max rounds reached",
			currentRound: 5,
			maxRounds:    5,
			hasToolCalls: true,
			expected:     false,
		},
		{
			name:         "should not continue when max rounds exceeded",
			currentRound: 6,
			maxRounds:    5,
			hasToolCalls: true,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := &IterativeProcessor{
				CurrentRound: tt.currentRound,
				MaxRounds:    tt.maxRounds,
			}

			result := processor.ShouldContinue(tt.hasToolCalls)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIterativeProcessor_AddToolResults(t *testing.T) {
	processor := &IterativeProcessor{
		CurrentRound: 0,
		ToolResults:  make([][]ToolResult, 0),
	}

	results1 := []ToolResult{
		{ToolCallID: "call1", Content: "result1"},
		{ToolCallID: "call2", Content: "result2"},
	}

	results2 := []ToolResult{
		{ToolCallID: "call3", Content: "result3"},
	}

	// Add first round of results
	processor.AddToolResults(results1)
	assert.Equal(t, 1, processor.CurrentRound)
	assert.Len(t, processor.ToolResults, 1)
	assert.Equal(t, results1, processor.ToolResults[0])

	// Add second round of results
	processor.AddToolResults(results2)
	assert.Equal(t, 2, processor.CurrentRound)
	assert.Len(t, processor.ToolResults, 2)
	assert.Equal(t, results2, processor.ToolResults[1])
}

func TestIterativeProcessor_GetTotalToolCalls(t *testing.T) {
	processor := &IterativeProcessor{
		ToolResults: [][]ToolResult{
			{
				{ToolCallID: "call1", Content: "result1"},
				{ToolCallID: "call2", Content: "result2"},
			},
			{
				{ToolCallID: "call3", Content: "result3"},
			},
			{
				{ToolCallID: "call4", Content: "result4"},
				{ToolCallID: "call5", Content: "result5"},
				{ToolCallID: "call6", Content: "result6"},
			},
		},
	}

	total := processor.GetTotalToolCalls()
	assert.Equal(t, 6, total)
}

func TestIterativeProcessor_GetTotalToolCalls_Empty(t *testing.T) {
	processor := &IterativeProcessor{
		ToolResults: make([][]ToolResult, 0),
	}

	total := processor.GetTotalToolCalls()
	assert.Equal(t, 0, total)
}

// Note: Mock services are defined in error_handling_test.go to avoid duplication

func TestIterativeProcessor_ProgressCallback(t *testing.T) {
	var progressMessages []string
	progressCallback := func(message string) {
		progressMessages = append(progressMessages, message)
	}

	msgCtx := &MessageContext{
		UserID:    "test-user",
		SessionID: "test-session",
		Message:   "Test message",
	}

	processor := NewIterativeProcessor(msgCtx, progressCallback)

	// Test that progress callback works
	processor.ProgressCallback("Test progress message")
	assert.Len(t, progressMessages, 1)
	assert.Equal(t, "Test progress message", progressMessages[0])

	// Test multiple messages
	processor.ProgressCallback("Second message")
	assert.Len(t, progressMessages, 2)
	assert.Equal(t, "Second message", progressMessages[1])
}

func TestIterativeProcessor_MaxRoundsEnforcement(t *testing.T) {
	processor := &IterativeProcessor{
		MaxRounds:    3,
		CurrentRound: 0,
	}

	// Should continue for rounds 0, 1, 2
	assert.True(t, processor.ShouldContinue(true))
	processor.CurrentRound++

	assert.True(t, processor.ShouldContinue(true))
	processor.CurrentRound++

	assert.True(t, processor.ShouldContinue(true))
	processor.CurrentRound++

	// Should not continue for round 3 (reached max)
	assert.False(t, processor.ShouldContinue(true))
}

func TestIterativeProcessor_ContextAccumulation(t *testing.T) {
	msgCtx := &MessageContext{
		UserID:              "test-user",
		SessionID:           "test-session",
		Message:             "Test message",
		ConversationHistory: []*models.Message{},
		AthleteLogbook:      &models.AthleteLogbook{Content: "Test logbook"},
		User: &models.User{
			ID:          "test-user",
			AccessToken: "test-token",
		},
	}

	processor := NewIterativeProcessor(msgCtx, func(string) {})

	// Simulate adding tool results across multiple rounds
	round1Results := []ToolResult{
		{ToolCallID: "call1", Content: "Profile data"},
	}
	processor.AddToolResults(round1Results)

	round2Results := []ToolResult{
		{ToolCallID: "call2", Content: "Activity data"},
		{ToolCallID: "call3", Content: "Stream data"},
	}
	processor.AddToolResults(round2Results)

	// Verify context accumulation
	assert.Equal(t, 2, processor.CurrentRound)
	assert.Len(t, processor.ToolResults, 2)
	assert.Equal(t, 3, processor.GetTotalToolCalls())

	// Verify original context is preserved
	assert.Equal(t, msgCtx, processor.Context)
	assert.Equal(t, "test-user", processor.Context.UserID)
	assert.Equal(t, "Test logbook", processor.Context.AthleteLogbook.Content)
}

func TestIterativeProcessor_SafeguardsPrevention(t *testing.T) {
	processor := &IterativeProcessor{
		MaxRounds:    2,
		CurrentRound: 0,
	}

	// Simulate reaching max rounds
	for i := 0; i < 5; i++ {
		if processor.ShouldContinue(true) {
			processor.CurrentRound++
		} else {
			break
		}
	}

	// Should stop at max rounds (2)
	assert.Equal(t, 2, processor.CurrentRound)
	assert.False(t, processor.ShouldContinue(true))
}

// Test enhanced multi-round analysis scenarios

func TestAIService_ShouldContinueAnalysis(t *testing.T) {
	service := setupTestAIService()

	tests := []struct {
		name             string
		currentRound     int
		maxRounds        int
		toolCalls        []openai.ToolCall
		hasContent       bool
		expectedContinue bool
		expectedReason   string
	}{
		{
			name:         "continue with profile analysis",
			currentRound: 0,
			maxRounds:    5,
			toolCalls: []openai.ToolCall{
				{Function: openai.FunctionCall{Name: "get-athlete-profile"}},
			},
			hasContent:       false,
			expectedContinue: true,
			expectedReason:   "continue_analysis",
		},
		{
			name:         "continue with deeper analysis",
			currentRound: 2,
			maxRounds:    5,
			toolCalls: []openai.ToolCall{
				{Function: openai.FunctionCall{Name: "get-activity-streams"}},
			},
			hasContent:       false,
			expectedContinue: true,
			expectedReason:   "continue_analysis",
		},
		{
			name:         "stop at max rounds",
			currentRound: 5,
			maxRounds:    5,
			toolCalls: []openai.ToolCall{
				{Function: openai.FunctionCall{Name: "get-recent-activities"}},
			},
			hasContent:       false,
			expectedContinue: false,
			expectedReason:   "max_rounds",
		},
		{
			name:             "stop with no tool calls",
			currentRound:     2,
			maxRounds:        5,
			toolCalls:        []openai.ToolCall{},
			hasContent:       true,
			expectedContinue: false,
			expectedReason:   "no_tools",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := &IterativeProcessor{
				CurrentRound: tt.currentRound,
				MaxRounds:    tt.maxRounds,
			}

			shouldContinue, reason := service.shouldContinueAnalysis(processor, tt.toolCalls, tt.hasContent)
			assert.Equal(t, tt.expectedContinue, shouldContinue)
			assert.Equal(t, tt.expectedReason, reason)
		})
	}
}

func TestAIService_AssessAnalysisDepth(t *testing.T) {
	service := setupTestAIService()

	tests := []struct {
		name          string
		toolResults   [][]ToolResult
		currentCalls  []openai.ToolCall
		expectedDepth int
	}{
		{
			name: "basic profile analysis",
			toolResults: [][]ToolResult{
				{
					{Content: `{"firstname": "John", "ftp": 250}`, Error: ""},
				},
			},
			currentCalls: []openai.ToolCall{
				{Function: openai.FunctionCall{Name: "get-recent-activities"}},
			},
			expectedDepth: 2, // profile + activities
		},
		{
			name: "comprehensive analysis",
			toolResults: [][]ToolResult{
				{
					{Content: `{"firstname": "John", "ftp": 250}`, Error: ""},
				},
				{
					{Content: `[{"distance": 5000, "type": "Run"}]`, Error: ""},
				},
				{
					{Content: `{"description": "Great run", "calories": 350}`, Error: ""},
				},
			},
			currentCalls: []openai.ToolCall{
				{Function: openai.FunctionCall{Name: "get-activity-streams"}},
			},
			expectedDepth: 3, // profile + activities + details (streams not yet executed)
		},
		{
			name:          "no analysis yet",
			toolResults:   [][]ToolResult{},
			currentCalls:  []openai.ToolCall{},
			expectedDepth: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := &IterativeProcessor{
				ToolResults: tt.toolResults,
			}

			depth := service.assessAnalysisDepth(processor, tt.currentCalls)
			assert.Equal(t, tt.expectedDepth, depth)
		})
	}
}

func TestAIService_GetCoachingProgressMessage(t *testing.T) {
	service := setupTestAIService()

	tests := []struct {
		name          string
		currentRound  int
		toolCalls     []openai.ToolCall
		expectedWords []string // Words that should be present in coaching messages
	}{
		{
			name:         "athlete profile analysis",
			currentRound: 0,
			toolCalls: []openai.ToolCall{
				{Function: openai.FunctionCall{Name: "get-athlete-profile"}},
			},
			expectedWords: []string{"athlete", "profile", "know"},
		},
		{
			name:         "recent activities analysis",
			currentRound: 1,
			toolCalls: []openai.ToolCall{
				{Function: openai.FunctionCall{Name: "get-recent-activities"}},
			},
			expectedWords: []string{"activities", "training", "recent"},
		},
		{
			name:         "detailed activity analysis",
			currentRound: 2,
			toolCalls: []openai.ToolCall{
				{Function: openai.FunctionCall{Name: "get-activity-details"}},
			},
			expectedWords: []string{"workout", "closer", "specific"},
		},
		{
			name:         "streams analysis",
			currentRound: 3,
			toolCalls: []openai.ToolCall{
				{Function: openai.FunctionCall{Name: "get-activity-streams"}},
			},
			expectedWords: []string{"data", "performance", "detailed"},
		},
		{
			name:          "fallback message",
			currentRound:  10,
			toolCalls:     []openai.ToolCall{},
			expectedWords: []string{"analysis", "final", "comprehensive"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := &IterativeProcessor{
				CurrentRound: tt.currentRound,
			}

			message := service.getCoachingProgressMessage(processor, tt.toolCalls)

			// Ensure message is not empty and coaching-focused
			assert.NotEmpty(t, message)
			assert.NotContains(t, message, "API")
			assert.NotContains(t, message, "executing")
			assert.NotContains(t, message, "tool")

			// Check that at least one expected word is present
			hasExpectedWord := false
			for _, word := range tt.expectedWords {
				if strings.Contains(strings.ToLower(message), strings.ToLower(word)) {
					hasExpectedWord = true
					break
				}
			}
			assert.True(t, hasExpectedWord, "Message should contain at least one expected word: %s", message)
		})
	}
}

func TestAIService_GenerateFinalResponse(t *testing.T) {
	service := setupTestAIService()

	tests := []struct {
		name          string
		reason        string
		hasContent    bool
		shouldBeEmpty bool
		expectedWords []string
	}{
		{
			name:          "max rounds reached",
			reason:        "max_rounds",
			hasContent:    false,
			shouldBeEmpty: false,
			expectedWords: []string{"training", "analysis", "recommendations", "comprehensive"},
		},
		{
			name:          "sufficient data gathered",
			reason:        "sufficient_data",
			hasContent:    false,
			shouldBeEmpty: false,
			expectedWords: []string{"analysis", "training", "recommend", "suggest", "coaching", "insights", "clear", "picture", "excellent"},
		},
		{
			name:          "already has content",
			reason:        "max_rounds",
			hasContent:    true,
			shouldBeEmpty: true,
			expectedWords: []string{},
		},
		{
			name:          "default case",
			reason:        "other",
			hasContent:    false,
			shouldBeEmpty: false,
			expectedWords: []string{"insights", "training", "share", "learned", "analysis", "coaching", "advice", "based"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := &IterativeProcessor{}

			response := service.generateFinalResponse(processor, tt.reason, tt.hasContent)

			if tt.shouldBeEmpty {
				assert.Empty(t, response)
			} else {
				assert.NotEmpty(t, response)
				assert.NotContains(t, response, "API")
				assert.NotContains(t, response, "executing")
				assert.NotContains(t, response, "tool")

				// Check that at least one expected word is present
				hasExpectedWord := false
				for _, word := range tt.expectedWords {
					if strings.Contains(strings.ToLower(response), strings.ToLower(word)) {
						hasExpectedWord = true
						break
					}
				}
				assert.True(t, hasExpectedWord, "Response should contain at least one expected word: %s", response)
			}
		})
	}
}

func TestAIService_ExecuteToolsWithRecovery(t *testing.T) {
	service := setupTestAIService()
	ctx := context.Background()

	user := &models.User{
		AccessToken: "test-token",
	}

	msgCtx := &MessageContext{
		UserID: "user-123",
		User:   user,
	}

	t.Run("partial failure recovery", func(t *testing.T) {
		// Mock one successful and one failed call
		expectedProfile := &StravaAthlete{
			ID:        123456,
			Firstname: "John",
			Lastname:  "Doe",
		}
		service.mockStrava.On("GetAthleteProfile", "test-token").Return(expectedProfile, nil)
		service.mockStrava.On("GetActivities", "test-token", mock.AnythingOfType("ActivityParams")).Return(([]*StravaActivity)(nil), fmt.Errorf("API error"))

		toolCalls := []openai.ToolCall{
			{
				ID:   "call-1",
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      "get-athlete-profile",
					Arguments: "{}",
				},
			},
			{
				ID:   "call-2",
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      "get-recent-activities",
					Arguments: `{"per_page": 10}`,
				},
			},
		}

		results, err := service.executeToolsWithRecovery(ctx, msgCtx, toolCalls)

		assert.NoError(t, err)
		assert.Len(t, results, 1) // Only successful result
		assert.Equal(t, "call-1", results[0].ToolCallID)
		assert.Empty(t, results[0].Error)

		service.mockStrava.AssertExpectations(t)
	})

	t.Run("all tools fail", func(t *testing.T) {
		// Create a fresh service for this test to avoid mock conflicts
		freshService := setupTestAIService()
		freshService.mockStrava.On("GetAthleteProfile", "test-token").Return((*StravaAthlete)(nil), fmt.Errorf("API error"))

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

		results, err := freshService.executeToolsWithRecovery(ctx, msgCtx, toolCalls)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "all tool calls failed")
		assert.Len(t, results, 0)

		freshService.mockStrava.AssertExpectations(t)
	})
}

func TestAIService_BuildEnhancedSystemPrompt(t *testing.T) {
	service := setupTestAIService()

	t.Run("with logbook and iterative guidance", func(t *testing.T) {
		logbook := &models.AthleteLogbook{
			Content: `{"personal_info": {"name": "John Doe"}}`,
		}

		msgCtx := &MessageContext{
			AthleteLogbook: logbook,
		}

		prompt := service.buildEnhancedSystemPrompt(msgCtx)

		assert.Contains(t, prompt, "Bodda")
		assert.Contains(t, prompt, "comprehensive analysis")
		assert.Contains(t, prompt, "multiple rounds of tool calls")
		assert.Contains(t, prompt, "Build insights progressively")
		assert.Contains(t, prompt, "Current Athlete Logbook:")
		assert.Contains(t, prompt, logbook.Content)
	})

	t.Run("without logbook", func(t *testing.T) {
		msgCtx := &MessageContext{}

		prompt := service.buildEnhancedSystemPrompt(msgCtx)

		assert.Contains(t, prompt, "Bodda")
		assert.Contains(t, prompt, "iterative analysis")
		assert.Contains(t, prompt, "No athlete logbook exists yet")
		assert.Contains(t, prompt, "update-athlete-logbook tool")
	})
}

func TestAIService_AccumulateAnalysisContext(t *testing.T) {
	service := setupTestAIService()

	processor := &IterativeProcessor{
		CurrentRound: 1,
		ToolResults:  [][]ToolResult{},
		Messages:     []openai.ChatCompletionMessage{},
	}

	toolCalls := []openai.ToolCall{
		{
			ID:   "call-1",
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionCall{
				Name: "get-athlete-profile",
			},
		},
	}

	toolResults := []ToolResult{
		{
			ToolCallID: "call-1",
			Content:    `{"firstname": "John", "lastname": "Doe"}`,
		},
	}

	responseContent := "Let me analyze your profile data..."

	updatedProcessor := service.accumulateAnalysisContext(processor, toolCalls, toolResults, responseContent)

	// Verify tool results were added
	assert.Equal(t, 2, updatedProcessor.CurrentRound)
	assert.Len(t, updatedProcessor.ToolResults, 1)
	assert.Equal(t, toolResults, updatedProcessor.ToolResults[0])

	// Verify messages were added
	assert.Len(t, updatedProcessor.Messages, 2)
	assert.Equal(t, openai.ChatMessageRoleAssistant, updatedProcessor.Messages[0].Role)
	assert.Equal(t, responseContent, updatedProcessor.Messages[0].Content)
	assert.Equal(t, toolCalls, updatedProcessor.Messages[0].ToolCalls)

	assert.Equal(t, openai.ChatMessageRoleTool, updatedProcessor.Messages[1].Role)
	assert.Equal(t, toolResults[0].Content, updatedProcessor.Messages[1].Content)
	assert.Equal(t, toolResults[0].ToolCallID, updatedProcessor.Messages[1].ToolCallID)
}

// Integration test for complete multi-round analysis workflow
func TestAIService_MultiRoundAnalysisWorkflow(t *testing.T) {
	service := setupTestAIService()

	// Create realistic test data
	user := &models.User{
		ID:          "user-123",
		AccessToken: "test-token",
	}

	logbook := &models.AthleteLogbook{
		UserID:  "user-123",
		Content: `{"personal_info": {"name": "John Doe", "age": 30}}`,
	}

	msgCtx := &MessageContext{
		UserID:              "user-123",
		SessionID:           "session-123",
		Message:             "How is my training progressing?",
		ConversationHistory: []*models.Message{},
		AthleteLogbook:      logbook,
		User:                user,
	}

	// Test context building
	messages := service.buildConversationContext(msgCtx)

	assert.Len(t, messages, 2) // system + current message
	assert.Equal(t, openai.ChatMessageRoleSystem, messages[0].Role)
	assert.Contains(t, messages[0].Content, "comprehensive analysis")
	assert.Contains(t, messages[0].Content, "John Doe")

	assert.Equal(t, openai.ChatMessageRoleUser, messages[1].Role)
	assert.Equal(t, "How is my training progressing?", messages[1].Content)

	// Test iterative processor workflow
	processor := NewIterativeProcessor(msgCtx, func(string) {})

	// Simulate first round - profile analysis
	toolCalls1 := []openai.ToolCall{
		{Function: openai.FunctionCall{Name: "get-athlete-profile"}},
	}

	shouldContinue, reason := service.shouldContinueAnalysis(processor, toolCalls1, false)
	assert.True(t, shouldContinue)
	assert.Equal(t, "continue_analysis", reason)

	progressMsg := service.getCoachingProgressMessage(processor, toolCalls1)
	assert.NotEmpty(t, progressMsg)
	assert.NotContains(t, progressMsg, "API")
	assert.NotContains(t, progressMsg, "executing")
	// Should contain athlete-related words
	hasAthleteWord := strings.Contains(strings.ToLower(progressMsg), "athlete") ||
		strings.Contains(strings.ToLower(progressMsg), "profile") ||
		strings.Contains(strings.ToLower(progressMsg), "know")
	assert.True(t, hasAthleteWord, "Progress message should be athlete-focused: %s", progressMsg)

	// Simulate second round - activities analysis
	processor.CurrentRound = 1
	toolCalls2 := []openai.ToolCall{
		{Function: openai.FunctionCall{Name: "get-recent-activities"}},
	}

	shouldContinue, reason = service.shouldContinueAnalysis(processor, toolCalls2, false)
	assert.True(t, shouldContinue)

	progressMsg = service.getCoachingProgressMessage(processor, toolCalls2)
	assert.NotEmpty(t, progressMsg)
	assert.NotContains(t, progressMsg, "API")
	// Should contain training/activities-related words
	hasTrainingWord := strings.Contains(strings.ToLower(progressMsg), "training") ||
		strings.Contains(strings.ToLower(progressMsg), "activities") ||
		strings.Contains(strings.ToLower(progressMsg), "recent")
	assert.True(t, hasTrainingWord, "Progress message should be training-focused: %s", progressMsg)

	// Simulate reaching sufficient analysis depth
	processor.CurrentRound = 4
	toolCalls3 := []openai.ToolCall{
		{Function: openai.FunctionCall{Name: "update-athlete-logbook"}},
	}

	shouldContinue, reason = service.shouldContinueAnalysis(processor, toolCalls3, false)
	assert.False(t, shouldContinue)
	assert.Equal(t, "sufficient_data", reason)

	finalResponse := service.generateFinalResponse(processor, reason, false)
	assert.NotEmpty(t, finalResponse)
	assert.NotContains(t, finalResponse, "API")
	// Should contain coaching-related words
	hasCoachingWord := strings.Contains(strings.ToLower(finalResponse), "analysis") ||
		strings.Contains(strings.ToLower(finalResponse), "recommend") ||
		strings.Contains(strings.ToLower(finalResponse), "suggest") ||
		strings.Contains(strings.ToLower(finalResponse), "training")
	assert.True(t, hasCoachingWord, "Final response should be coaching-focused: %s", finalResponse)
}
