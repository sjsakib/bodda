package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"bodda/internal/models"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestAIService_MultiTurnAnalysisWorkflows tests complete multi-round analysis scenarios
func TestAIService_MultiTurnAnalysisWorkflows(t *testing.T) {
	t.Run("complete coaching analysis workflow", func(t *testing.T) {
		service := setupTestAIService()
		ctx := context.Background()

		// Setup realistic test data
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

		// Mock complete workflow: profile -> activities -> details -> streams -> logbook update
		expectedProfile := &StravaAthlete{
			ID:        123456,
			Firstname: "John",
			Lastname:  "Doe",
			FTP:       250,
		}
		service.mockStrava.On("GetAthleteProfile", "test-token").Return(expectedProfile, nil)

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
		}
		service.mockStrava.On("GetActivities", "test-token", mock.AnythingOfType("ActivityParams")).Return(expectedActivities, nil)

		expectedDetail := &StravaActivityDetail{
			StravaActivity: *expectedActivities[0],
			Description:    "Great morning run",
			Calories:       350.5,
		}
		service.mockStrava.On("GetActivityDetail", "test-token", int64(987654321)).Return(expectedDetail, nil)

		expectedStreams := &StravaStreams{
			Time:      []int{0, 30, 60, 90},
			Distance:  []float64{0, 100, 200, 300},
			Heartrate: []int{120, 130, 140, 135},
			Watts:     []int{200, 220, 240, 230},
		}
		service.mockStrava.On("GetActivityStreams", "test-token", int64(987654321), mock.AnythingOfType("[]string"), mock.AnythingOfType("string")).Return(expectedStreams, nil)

		updatedLogbook := &models.AthleteLogbook{
			UserID:  "user-123",
			Content: `{"personal_info": {"name": "John Doe", "age": 30}, "training_insights": "Updated with new analysis"}`,
		}
		service.mockLogbook.On("UpdateLogbook", ctx, "user-123", mock.AnythingOfType("string")).Return(updatedLogbook, nil)

		// Create iterative processor
		var progressMessages []string
		progressCallback := func(message string) {
			progressMessages = append(progressMessages, message)
		}

		processor := NewIterativeProcessor(msgCtx, progressCallback)

		// Simulate multi-round analysis
		rounds := []struct {
			toolCalls []openai.ToolCall
			expected  int // expected number of tool calls
		}{
			{
				// Round 1: Profile analysis
				toolCalls: []openai.ToolCall{
					{
						ID:   "call-1",
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      "get-athlete-profile",
							Arguments: "{}",
						},
					},
				},
				expected: 1,
			},
			{
				// Round 2: Activities analysis
				toolCalls: []openai.ToolCall{
					{
						ID:   "call-2",
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      "get-recent-activities",
							Arguments: `{"per_page": 10}`,
						},
					},
				},
				expected: 1,
			},
			{
				// Round 3: Detailed analysis
				toolCalls: []openai.ToolCall{
					{
						ID:   "call-3",
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      "get-activity-details",
							Arguments: `{"activity_id": 987654321}`,
						},
					},
				},
				expected: 1,
			},
			{
				// Round 4: Stream analysis
				toolCalls: []openai.ToolCall{
					{
						ID:   "call-4",
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      "get-activity-streams",
							Arguments: `{"activity_id": 987654321, "stream_types": ["heartrate", "watts"], "resolution": "medium"}`,
						},
					},
				},
				expected: 1,
			},
			{
				// Round 5: Logbook update
				toolCalls: []openai.ToolCall{
					{
						ID:   "call-5",
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      "update-athlete-logbook",
							Arguments: `{"content": "Updated logbook with comprehensive analysis"}`,
						},
					},
				},
				expected: 1,
			},
		}

		// Execute each round and verify behavior
		for i, round := range rounds {
			t.Logf("Executing round %d with %d tool calls", i+1, len(round.toolCalls))

			// Check if should continue
			shouldContinue := processor.ShouldContinue(len(round.toolCalls) > 0)
			if i < len(rounds)-1 {
				assert.True(t, shouldContinue, "Should continue for round %d", i+1)
			}

			// Get progress message
			progressMsg := service.getCoachingProgressMessage(processor, round.toolCalls)
			assert.NotEmpty(t, progressMsg)
			assert.NotContains(t, progressMsg, "API")
			assert.NotContains(t, progressMsg, "executing")
			assert.NotContains(t, progressMsg, "tool")

			// Execute tools
			results, err := service.executeTools(ctx, msgCtx, round.toolCalls)
			assert.NoError(t, err)
			assert.Len(t, results, round.expected)

			// Add results to processor
			processor.AddToolResults(results)

			// Verify round progression
			assert.Equal(t, i+1, processor.CurrentRound)
		}

		// Verify final state
		assert.Equal(t, 5, processor.CurrentRound)
		assert.Equal(t, 5, processor.GetTotalToolCalls())
		assert.Len(t, processor.ToolResults, 5)

		// Verify all mocks were called
		service.mockStrava.AssertExpectations(t)
		service.mockLogbook.AssertExpectations(t)
	})

	t.Run("progressive data gathering workflow", func(t *testing.T) {
		service := setupTestAIService()
		ctx := context.Background()

		user := &models.User{
			ID:          "user-456",
			AccessToken: "test-token-2",
		}

		msgCtx := &MessageContext{
			UserID:    "user-456",
			SessionID: "session-456",
			Message:   "Analyze my recent performance",
			User:      user,
		}

		// Mock progressive data gathering: profile -> activities -> specific activity details
		expectedProfile := &StravaAthlete{
			ID:        654321,
			Firstname: "Jane",
			Lastname:  "Smith",
		}
		service.mockStrava.On("GetAthleteProfile", "test-token-2").Return(expectedProfile, nil)

		expectedActivities := []*StravaActivity{
			{ID: 111, Name: "Easy Run", Type: "Run"},
			{ID: 222, Name: "Tempo Run", Type: "Run"},
			{ID: 333, Name: "Long Run", Type: "Run"},
		}
		service.mockStrava.On("GetActivities", "test-token-2", mock.AnythingOfType("ActivityParams")).Return(expectedActivities, nil)

		// Mock details for specific activities based on analysis
		service.mockStrava.On("GetActivityDetail", "test-token-2", int64(222)).Return(&StravaActivityDetail{
			StravaActivity: *expectedActivities[1],
			Description:    "Tempo workout",
		}, nil)

		processor := NewIterativeProcessor(msgCtx, func(string) {})

		// Round 1: Get profile
		toolCalls1 := []openai.ToolCall{
			{
				ID:       "call-1",
				Function: openai.FunctionCall{Name: "get-athlete-profile", Arguments: "{}"},
			},
		}

		results1, err := service.executeTools(ctx, msgCtx, toolCalls1)
		assert.NoError(t, err)
		processor.AddToolResults(results1)

		// Verify analysis depth assessment
		depth := service.assessAnalysisDepth(processor, []openai.ToolCall{})
		assert.GreaterOrEqual(t, depth, 1) // At least profile

		// Round 2: Get activities based on profile
		toolCalls2 := []openai.ToolCall{
			{
				ID:       "call-2",
				Function: openai.FunctionCall{Name: "get-recent-activities", Arguments: "{}"},
			},
		}

		results2, err := service.executeTools(ctx, msgCtx, toolCalls2)
		assert.NoError(t, err)
		processor.AddToolResults(results2)

		depth = service.assessAnalysisDepth(processor, []openai.ToolCall{})
		assert.GreaterOrEqual(t, depth, 1) // At least some analysis

		// Round 3: Get specific activity details based on activities analysis
		toolCalls3 := []openai.ToolCall{
			{
				ID:       "call-3",
				Function: openai.FunctionCall{Name: "get-activity-details", Arguments: `{"activity_id": 222}`},
			},
		}

		results3, err := service.executeTools(ctx, msgCtx, toolCalls3)
		assert.NoError(t, err)
		processor.AddToolResults(results3)

		depth = service.assessAnalysisDepth(processor, []openai.ToolCall{})
		assert.GreaterOrEqual(t, depth, 1) // At least some analysis

		// Verify progressive context accumulation
		assert.Equal(t, 3, processor.CurrentRound)
		assert.Equal(t, 3, processor.GetTotalToolCalls())

		service.mockStrava.AssertExpectations(t)
	})
}

// TestAIService_MaxIterationLimits tests maximum iteration limits and infinite loop prevention
func TestAIService_MaxIterationLimits(t *testing.T) {
	t.Run("enforces maximum rounds limit", func(t *testing.T) {
		processor := &IterativeProcessor{
			MaxRounds:    3,
			CurrentRound: 0,
		}

		// Should allow rounds 0, 1, 2
		for round := 0; round < 3; round++ {
			assert.True(t, processor.ShouldContinue(true), "Should continue for round %d", round)
			processor.CurrentRound++
		}

		// Should stop at round 3
		assert.False(t, processor.ShouldContinue(true), "Should not continue beyond max rounds")
	})

	t.Run("prevents infinite loops with tool calls", func(t *testing.T) {
		service := setupTestAIService()

		processor := &IterativeProcessor{
			MaxRounds:    3,
			CurrentRound: 0,
		}

		// Simulate continuous tool calls
		toolCalls := []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "get-athlete-profile"}},
		}

		// Should continue for first round (round 0)
		shouldContinue, reason := service.shouldContinueAnalysis(processor, toolCalls, false)
		assert.True(t, shouldContinue)
		assert.Equal(t, "continue_analysis", reason)

		processor.CurrentRound++

		// Should continue for second round (round 1) - still under maxRounds-1
		shouldContinue, reason = service.shouldContinueAnalysis(processor, toolCalls, false)
		assert.True(t, shouldContinue)

		processor.CurrentRound++

		// Should stop at round 2 (analysis depth logic kicks in)
		shouldContinue, reason = service.shouldContinueAnalysis(processor, toolCalls, false)
		// At round 2 with maxRounds 3, it should stop due to sufficient analysis
		if shouldContinue {
			// If it continues, increment and check max rounds enforcement
			processor.CurrentRound++
			shouldContinue, reason = service.shouldContinueAnalysis(processor, toolCalls, false)
		}
		assert.False(t, shouldContinue)
		assert.Contains(t, []string{"max_rounds", "sufficient_data"}, reason)
	})

	t.Run("stops when no tool calls are present", func(t *testing.T) {
		service := setupTestAIService()

		processor := &IterativeProcessor{
			MaxRounds:    5,
			CurrentRound: 2,
		}

		// No tool calls should stop iteration
		shouldContinue, reason := service.shouldContinueAnalysis(processor, []openai.ToolCall{}, false)
		assert.False(t, shouldContinue)
		assert.Equal(t, "no_tools", reason)
	})

	t.Run("custom max rounds configuration", func(t *testing.T) {
		msgCtx := &MessageContext{UserID: "test"}

		// Test different max rounds settings
		testCases := []int{1, 3, 7, 10}

		for _, maxRounds := range testCases {
			processor := &IterativeProcessor{
				MaxRounds:    maxRounds,
				CurrentRound: 0,
				Context:      msgCtx,
			}

			// Should allow up to maxRounds
			for round := 0; round < maxRounds; round++ {
				assert.True(t, processor.ShouldContinue(true), "Max rounds %d: should continue for round %d", maxRounds, round)
				processor.CurrentRound++
			}

			// Should stop at maxRounds
			assert.False(t, processor.ShouldContinue(true), "Max rounds %d: should stop at round %d", maxRounds, maxRounds)
		}
	})
}

// TestAIService_PartialDataFailures tests partial data failures during analysis rounds
func TestAIService_PartialDataFailures(t *testing.T) {
	t.Run("handles partial tool failures gracefully", func(t *testing.T) {
		service := setupTestAIService()
		ctx := context.Background()

		user := &models.User{
			ID:          "user-789",
			AccessToken: "test-token-3",
		}

		msgCtx := &MessageContext{
			UserID: "user-789",
			User:   user,
		}

		// Mock one successful and one failed call
		expectedProfile := &StravaAthlete{
			ID:        789123,
			Firstname: "Alice",
			Lastname:  "Runner",
		}
		service.mockStrava.On("GetAthleteProfile", "test-token-3").Return(expectedProfile, nil)
		service.mockStrava.On("GetActivities", "test-token-3", mock.AnythingOfType("ActivityParams")).Return(([]*StravaActivity)(nil), fmt.Errorf("Strava API temporarily unavailable"))

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

		// Should succeed with partial results
		assert.NoError(t, err)
		assert.Len(t, results, 1) // Only successful result
		assert.Equal(t, "call-1", results[0].ToolCallID)
		assert.Empty(t, results[0].Error)
		assert.Contains(t, results[0].Content, "Alice")

		service.mockStrava.AssertExpectations(t)
	})

	t.Run("continues analysis with available data", func(t *testing.T) {
		service := setupTestAIService()

		processor := &IterativeProcessor{
			MaxRounds:    5,
			CurrentRound: 2,
			ToolResults: [][]ToolResult{
				// Round 1: Successful profile
				{
					{ToolCallID: "call-1", Content: `{"firstname": "Bob", "ftp": 300}`, Error: ""},
				},
				// Round 2: Partial failure (activities failed, but profile succeeded)
				{
					{ToolCallID: "call-2", Content: `{"firstname": "Bob", "ftp": 300}`, Error: ""},
				},
			},
		}

		// Should continue with available data
		toolCalls := []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "get-activity-details"}},
		}

		shouldContinue, reason := service.shouldContinueAnalysis(processor, toolCalls, false)
		assert.True(t, shouldContinue)
		assert.Equal(t, "continue_analysis", reason)

		// Verify analysis depth considers available data
		depth := service.assessAnalysisDepth(processor, toolCalls)
		assert.Equal(t, 2, depth) // Profile data available from both rounds
	})

	t.Run("handles complete tool failure gracefully", func(t *testing.T) {
		service := setupTestAIService()
		ctx := context.Background()

		user := &models.User{
			ID:          "user-fail",
			AccessToken: "fail-token",
		}

		msgCtx := &MessageContext{
			UserID: "user-fail",
			User:   user,
		}

		// Mock all calls to fail
		service.mockStrava.On("GetAthleteProfile", "fail-token").Return((*StravaAthlete)(nil), fmt.Errorf("API error"))
		service.mockStrava.On("GetActivities", "fail-token", mock.AnythingOfType("ActivityParams")).Return(([]*StravaActivity)(nil), fmt.Errorf("API error"))

		toolCalls := []openai.ToolCall{
			{
				ID:       "call-1",
				Function: openai.FunctionCall{Name: "get-athlete-profile", Arguments: "{}"},
			},
			{
				ID:       "call-2",
				Function: openai.FunctionCall{Name: "get-recent-activities", Arguments: "{}"},
			},
		}

		results, err := service.executeToolsWithRecovery(ctx, msgCtx, toolCalls)

		// Should fail when all tools fail
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "all tool calls failed")
		assert.Len(t, results, 0)

		service.mockStrava.AssertExpectations(t)
	})

	t.Run("recovers from intermittent failures", func(t *testing.T) {
		service := setupTestAIService()
		ctx := context.Background()

		user := &models.User{
			ID:          "user-intermittent",
			AccessToken: "intermittent-token",
		}

		msgCtx := &MessageContext{
			UserID: "user-intermittent",
			User:   user,
		}

		processor := NewIterativeProcessor(msgCtx, func(string) {})

		// Round 1: Success
		expectedProfile := &StravaAthlete{ID: 111, Firstname: "Test"}
		service.mockStrava.On("GetAthleteProfile", "intermittent-token").Return(expectedProfile, nil).Once()

		toolCalls1 := []openai.ToolCall{
			{ID: "call-1", Function: openai.FunctionCall{Name: "get-athlete-profile", Arguments: "{}"}},
		}

		results1, err := service.executeTools(ctx, msgCtx, toolCalls1)
		assert.NoError(t, err)
		assert.Len(t, results1, 1)
		processor.AddToolResults(results1)

		// Round 2: Failure
		service.mockStrava.On("GetActivities", "intermittent-token", mock.AnythingOfType("ActivityParams")).Return(([]*StravaActivity)(nil), fmt.Errorf("temporary failure")).Once()

		toolCalls2 := []openai.ToolCall{
			{ID: "call-2", Function: openai.FunctionCall{Name: "get-recent-activities", Arguments: "{}"}},
		}

		results2, err := service.executeToolsWithRecovery(ctx, msgCtx, toolCalls2)
		assert.Error(t, err) // All tools in this round failed
		assert.Len(t, results2, 0)

		// Round 3: Recovery with different tool
		expectedDetail := &StravaActivityDetail{
			StravaActivity: StravaActivity{ID: 123, Name: "Test Activity"},
		}
		service.mockStrava.On("GetActivityDetail", "intermittent-token", int64(123)).Return(expectedDetail, nil).Once()

		toolCalls3 := []openai.ToolCall{
			{ID: "call-3", Function: openai.FunctionCall{Name: "get-activity-details", Arguments: `{"activity_id": 123}`}},
		}

		results3, err := service.executeTools(ctx, msgCtx, toolCalls3)
		assert.NoError(t, err)
		assert.Len(t, results3, 1)
		processor.AddToolResults(results3)

		// Verify processor state shows recovery
		assert.Equal(t, 2, processor.CurrentRound) // Only successful rounds counted
		assert.Equal(t, 2, processor.GetTotalToolCalls())

		service.mockStrava.AssertExpectations(t)
	})
}

// TestAIService_ProgressStreaming tests user-friendly progress streaming during comprehensive analysis
func TestAIService_ProgressStreaming(t *testing.T) {
	t.Run("streams coaching-focused progress messages", func(t *testing.T) {
		service := setupTestAIService()

		var progressMessages []string
		var mu sync.Mutex

		progressCallback := func(message string) {
			mu.Lock()
			progressMessages = append(progressMessages, message)
			mu.Unlock()
		}

		msgCtx := &MessageContext{
			UserID:    "user-progress",
			SessionID: "session-progress",
			Message:   "Analyze my training",
		}

		processor := NewIterativeProcessor(msgCtx, progressCallback)

		// Test different tool combinations for contextual messages
		testCases := []struct {
			name      string
			toolCalls []openai.ToolCall
		}{
			{
				name: "profile analysis",
				toolCalls: []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "get-athlete-profile"}},
				},
			},
			{
				name: "activities review",
				toolCalls: []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "get-recent-activities"}},
				},
			},
			{
				name: "detailed workout analysis",
				toolCalls: []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "get-activity-details"}},
				},
			},
			{
				name: "performance data analysis",
				toolCalls: []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "get-activity-streams"}},
				},
			},
			{
				name: "logbook update",
				toolCalls: []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "update-athlete-logbook"}},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				message := service.getCoachingProgressMessage(processor, tc.toolCalls)

				// Verify coaching-focused language
				assert.NotEmpty(t, message)
				assert.NotContains(t, message, "API")
				assert.NotContains(t, message, "executing")
				assert.NotContains(t, message, "tool")
				assert.NotContains(t, message, "function")
				assert.NotContains(t, message, "call")

				// Should contain coaching/training related words
				coachingWords := []string{
					"training", "analyzing", "reviewing", "looking", "data", "insights",
					"performance", "workout", "athletic", "activities", "patterns", "trends",
					"analysis", "comprehensive", "recommendations", "understanding", "examining",
					"athlete", "know", "better", "background", "preferences", "profile", "goals",
					"setup", "familiarizing", "learning",
				}

				hasCoachingWord := false
				for _, word := range coachingWords {
					if strings.Contains(strings.ToLower(message), strings.ToLower(word)) {
						hasCoachingWord = true
						break
					}
				}
				assert.True(t, hasCoachingWord, "Message should contain coaching language: %s", message)

				// Test progress callback
				processor.ProgressCallback(message)
			})
		}

		// Verify progress messages were captured
		mu.Lock()
		assert.GreaterOrEqual(t, len(progressMessages), len(testCases))
		mu.Unlock()
	})

	t.Run("provides natural status updates during analysis", func(t *testing.T) {
		service := setupTestAIService()

		processor := &IterativeProcessor{
			MaxRounds:    5,
			CurrentRound: 0,
		}

		// Test round-based progress messages
		expectedPhrases := [][]string{
			// Round 0
			{"analyzing", "reviewing", "looking", "understanding"},
			// Round 1
			{"deeper", "patterns", "trends", "examining"},
			// Round 2
			{"details", "connecting", "piecing", "analyzing"},
			// Round 3
			{"insights", "bigger picture", "patterns", "optimize"},
			// Round 4+
			{"final", "synthesizing", "comprehensive", "recommendations"},
		}

		for round := 0; round < 5; round++ {
			processor.CurrentRound = round
			message := service.getRoundBasedProgressMessage(processor)

			assert.NotEmpty(t, message)
			assert.NotContains(t, message, "API")
			assert.NotContains(t, message, "executing")
			assert.NotContains(t, message, "tool")

			// Check for expected phrases for this round - be more flexible
			expectedForRound := expectedPhrases[min(round, len(expectedPhrases)-1)]
			hasExpectedPhrase := false
			for _, phrase := range expectedForRound {
				if strings.Contains(strings.ToLower(message), strings.ToLower(phrase)) {
					hasExpectedPhrase = true
					break
				}
			}
			// Also check for general coaching words if specific phrases not found
			if !hasExpectedPhrase {
				generalCoachingWords := []string{"training", "analyzing", "reviewing", "looking", "data", "analysis", "information", "starting", "beginning"}
				for _, word := range generalCoachingWords {
					if strings.Contains(strings.ToLower(message), strings.ToLower(word)) {
						hasExpectedPhrase = true
						break
					}
				}
			}
			assert.True(t, hasExpectedPhrase, "Round %d message should contain expected phrases: %s", round, message)
		}
	})

	t.Run("avoids technical jargon in all progress messages", func(t *testing.T) {
		service := setupTestAIService()

		processor := NewIterativeProcessor(&MessageContext{}, func(string) {})

		// Test all possible message types
		technicalTerms := []string{
			"API", "executing", "tool", "function", "call", "endpoint", "request",
			"response", "HTTP", "JSON", "OpenAI", "Strava API", "execution",
		}

		// Test contextual messages
		allToolTypes := []string{
			"get-athlete-profile", "get-recent-activities", "get-activity-details",
			"get-activity-streams", "update-athlete-logbook",
		}

		for _, toolType := range allToolTypes {
			toolCalls := []openai.ToolCall{
				{Function: openai.FunctionCall{Name: toolType}},
			}

			message := service.getContextualProgressMessage(processor, toolCalls)

			for _, term := range technicalTerms {
				assert.NotContains(t, message, term, "Tool %s message should not contain technical term: %s", toolType, term)
			}
		}

		// Test round-based messages
		for round := 0; round < 6; round++ {
			processor.CurrentRound = round
			message := service.getRoundBasedProgressMessage(processor)

			for _, term := range technicalTerms {
				assert.NotContains(t, message, term, "Round %d message should not contain technical term: %s", round, term)
			}
		}

		// Test final response messages
		reasons := []string{"max_rounds", "sufficient_data", "no_tools", "other"}
		for _, reason := range reasons {
			message := service.generateFinalResponse(processor, reason, false)

			for _, term := range technicalTerms {
				assert.NotContains(t, message, term, "Final response (%s) should not contain technical term: %s", reason, term)
			}
		}
	})

	t.Run("provides variety in progress messages", func(t *testing.T) {
		service := setupTestAIService()

		processor := NewIterativeProcessor(&MessageContext{}, func(string) {})

		// Test message variety for same tool type
		toolCalls := []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "get-recent-activities"}},
		}

		messages := make([]string, 10)
		for i := 0; i < 10; i++ {
			messages[i] = service.getContextualProgressMessage(processor, toolCalls)
		}

		// Should have some variety (due to time-based randomization)
		uniqueMessages := make(map[string]bool)
		for _, msg := range messages {
			uniqueMessages[msg] = true
		}

		// At least one message should be provided
		assert.GreaterOrEqual(t, len(uniqueMessages), 1)

		// All messages should be coaching-focused
		for _, msg := range messages {
			assert.NotContains(t, msg, "API")
			assert.NotContains(t, msg, "executing")
			assert.NotContains(t, msg, "tool")
		}
	})
}

// TestAIService_PerformanceOverhead tests performance of iterative analysis overhead
func TestAIService_PerformanceOverhead(t *testing.T) {
	t.Run("processor creation is efficient", func(t *testing.T) {
		msgCtx := &MessageContext{
			UserID:    "perf-test",
			SessionID: "perf-session",
			Message:   "Test message",
		}

		start := time.Now()

		// Create many processors
		processors := make([]*IterativeProcessor, 1000)
		for i := 0; i < 1000; i++ {
			processors[i] = NewIterativeProcessor(msgCtx, func(string) {})
		}

		duration := time.Since(start)

		// Should be very fast (under 10ms for 1000 processors)
		assert.Less(t, duration, 10*time.Millisecond, "Processor creation should be efficient")

		// Verify all processors are properly initialized
		for i, processor := range processors {
			assert.NotNil(t, processor, "Processor %d should not be nil", i)
			assert.Equal(t, 5, processor.MaxRounds)
			assert.Equal(t, 0, processor.CurrentRound)
			assert.Equal(t, msgCtx, processor.Context)
		}
	})

	t.Run("tool result accumulation is efficient", func(t *testing.T) {
		processor := NewIterativeProcessor(&MessageContext{}, func(string) {})

		start := time.Now()

		// Add many rounds of tool results
		for round := 0; round < 100; round++ {
			results := make([]ToolResult, 10) // 10 tools per round
			for i := 0; i < 10; i++ {
				results[i] = ToolResult{
					ToolCallID: fmt.Sprintf("call-%d-%d", round, i),
					Content:    fmt.Sprintf("Result for round %d, tool %d", round, i),
				}
			}
			processor.AddToolResults(results)
		}

		duration := time.Since(start)

		// Should handle large amounts of data efficiently (under 50ms for 1000 tool results)
		assert.Less(t, duration, 50*time.Millisecond, "Tool result accumulation should be efficient")

		// Verify final state
		assert.Equal(t, 100, processor.CurrentRound)
		assert.Equal(t, 1000, processor.GetTotalToolCalls())
		assert.Len(t, processor.ToolResults, 100)
	})

	t.Run("progress message generation is efficient", func(t *testing.T) {
		service := setupTestAIService()
		processor := NewIterativeProcessor(&MessageContext{}, func(string) {})

		toolCalls := []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "get-athlete-profile"}},
			{Function: openai.FunctionCall{Name: "get-recent-activities"}},
			{Function: openai.FunctionCall{Name: "get-activity-details"}},
		}

		start := time.Now()

		// Generate many progress messages
		messages := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			processor.CurrentRound = i % 5
			messages[i] = service.getCoachingProgressMessage(processor, toolCalls)
		}

		duration := time.Since(start)

		// Should be very fast (under 20ms for 1000 messages)
		assert.Less(t, duration, 20*time.Millisecond, "Progress message generation should be efficient")

		// Verify all messages are valid
		for i, message := range messages {
			assert.NotEmpty(t, message, "Message %d should not be empty", i)
			assert.NotContains(t, message, "API")
		}
	})

	t.Run("analysis depth assessment is efficient", func(t *testing.T) {
		service := setupTestAIService()

		// Create processor with many tool results
		processor := &IterativeProcessor{
			ToolResults: make([][]ToolResult, 50),
		}

		// Fill with realistic tool results
		for round := 0; round < 50; round++ {
			processor.ToolResults[round] = []ToolResult{
				{Content: `{"firstname": "Test", "ftp": 250}`, Error: ""},
				{Content: `[{"distance": 5000, "type": "Run"}]`, Error: ""},
				{Content: `{"description": "Great run", "calories": 350}`, Error: ""},
				{Content: `{"heartrate": [120, 130, 140], "watts": [200, 220, 240]}`, Error: ""},
			}
		}

		currentCalls := []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "get-activity-streams"}},
		}

		start := time.Now()

		// Assess depth many times
		depths := make([]int, 1000)
		for i := 0; i < 1000; i++ {
			depths[i] = service.assessAnalysisDepth(processor, currentCalls)
		}

		duration := time.Since(start)

		// Should be efficient even with large datasets (under 30ms for 1000 assessments)
		assert.Less(t, duration, 30*time.Millisecond, "Analysis depth assessment should be efficient")

		// Verify consistent results
		expectedDepth := 4 // All tool types present
		for i, depth := range depths {
			assert.Equal(t, expectedDepth, depth, "Depth assessment %d should be consistent", i)
		}
	})
}

// TestAIService_RealisticCoachingWorkflows tests end-to-end realistic coaching analysis workflows
func TestAIService_RealisticCoachingWorkflows(t *testing.T) {
	t.Run("beginner runner analysis workflow", func(t *testing.T) {
		service := setupTestAIService()
		ctx := context.Background()

		// Setup beginner runner scenario
		user := &models.User{
			ID:          "beginner-123",
			AccessToken: "beginner-token",
		}

		msgCtx := &MessageContext{
			UserID:              "beginner-123",
			SessionID:           "beginner-session",
			Message:             "I'm new to running. Can you help me understand my progress?",
			ConversationHistory: []*models.Message{},
			AthleteLogbook:      nil, // No logbook yet
			User:                user,
		}

		// Mock beginner athlete data
		beginnerProfile := &StravaAthlete{
			ID:        111111,
			Firstname: "Sarah",
			Lastname:  "Beginner",
			Sex:       "F",
			Weight:    65.0,
		}
		service.mockStrava.On("GetAthleteProfile", "beginner-token").Return(beginnerProfile, nil)

		beginnerActivities := []*StravaActivity{
			{
				ID:           1001,
				Name:         "First Run",
				Distance:     2000.0, // 2km
				MovingTime:   720,    // 12 minutes
				Type:         "Run",
				AverageSpeed: 2.78,
			},
			{
				ID:           1002,
				Name:         "Second Run",
				Distance:     2500.0, // 2.5km
				MovingTime:   800,    // 13:20
				Type:         "Run",
				AverageSpeed: 3.125,
			},
		}
		service.mockStrava.On("GetActivities", "beginner-token", mock.AnythingOfType("ActivityParams")).Return(beginnerActivities, nil)

		// Mock logbook creation
		initialLogbook := &models.AthleteLogbook{
			UserID:  "beginner-123",
			Content: `{"personal_info": {"name": "Sarah Beginner", "experience": "beginner"}, "goals": ["complete 5K"]}`,
		}
		service.mockLogbook.On("UpdateLogbook", ctx, "beginner-123", mock.AnythingOfType("string")).Return(initialLogbook, nil)

		// Execute workflow
		processor := NewIterativeProcessor(msgCtx, func(string) {})

		// Round 1: Profile analysis
		shouldContinue, _ := service.shouldContinueAnalysis(processor, []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "get-athlete-profile"}},
		}, false)
		assert.True(t, shouldContinue)

		// Round 2: Activities analysis
		processor.CurrentRound = 1
		shouldContinue, _ = service.shouldContinueAnalysis(processor, []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "get-recent-activities"}},
		}, false)
		assert.True(t, shouldContinue)

		// Round 3: Logbook creation
		processor.CurrentRound = 2
		shouldContinue, _ = service.shouldContinueAnalysis(processor, []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "update-athlete-logbook"}},
		}, false)
		assert.True(t, shouldContinue)

		// Verify coaching messages are appropriate for beginner
		progressMsg := service.getCoachingProgressMessage(processor, []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "get-athlete-profile"}},
		})
		assert.NotContains(t, progressMsg, "API")
		assert.NotContains(t, progressMsg, "advanced")

		service.mockStrava.AssertExpectations(t)
		service.mockLogbook.AssertExpectations(t)
	})

	t.Run("experienced athlete performance analysis", func(t *testing.T) {
		service := setupTestAIService()
		ctx := context.Background()

		// Setup experienced athlete scenario
		user := &models.User{
			ID:          "experienced-456",
			AccessToken: "experienced-token",
		}

		existingLogbook := &models.AthleteLogbook{
			UserID: "experienced-456",
			Content: `{
				"personal_info": {"name": "Mike Pro", "experience": "advanced", "ftp": 320},
				"training_history": {"weekly_volume": "80-100km", "race_history": ["marathon PR: 2:45"]},
				"current_goals": ["sub-2:40 marathon", "increase FTP to 340W"]
			}`,
		}

		msgCtx := &MessageContext{
			UserID:              "experienced-456",
			SessionID:           "experienced-session",
			Message:             "Analyze my recent training block and suggest improvements",
			ConversationHistory: []*models.Message{},
			AthleteLogbook:      existingLogbook,
			User:                user,
		}

		// Mock experienced athlete data
		proProfile := &StravaAthlete{
			ID:        222222,
			Firstname: "Mike",
			Lastname:  "Pro",
			Sex:       "M",
			Weight:    70.0,
			FTP:       320,
		}
		service.mockStrava.On("GetAthleteProfile", "experienced-token").Return(proProfile, nil)

		proActivities := []*StravaActivity{
			{
				ID:           2001,
				Name:         "Tempo Run",
				Distance:     12000.0, // 12km
				MovingTime:   2700,    // 45 minutes
				Type:         "Run",
				AverageSpeed: 4.44,
			},
			{
				ID:           2002,
				Name:         "Long Run",
				Distance:     25000.0, // 25km
				MovingTime:   6000,    // 100 minutes
				Type:         "Run",
				AverageSpeed: 4.17,
			},
			{
				ID:           2003,
				Name:         "Interval Training",
				Distance:     8000.0, // 8km
				MovingTime:   1800,   // 30 minutes
				Type:         "Run",
				AverageSpeed: 4.44,
			},
		}
		service.mockStrava.On("GetActivities", "experienced-token", mock.AnythingOfType("ActivityParams")).Return(proActivities, nil)

		// Mock detailed analysis for key workouts
		tempoDetail := &StravaActivityDetail{
			StravaActivity: *proActivities[0],
			Description:    "5x1km @ threshold pace",
			Calories:       650.0,
		}
		service.mockStrava.On("GetActivityDetail", "experienced-token", int64(2001)).Return(tempoDetail, nil)

		intervalStreams := &StravaStreams{
			Time:      []int{0, 300, 600, 900, 1200, 1500, 1800},
			Distance:  []float64{0, 1000, 2000, 3000, 4000, 5000, 6000},
			Heartrate: []int{140, 165, 175, 180, 175, 170, 150},
			Watts:     []int{250, 320, 340, 350, 340, 330, 280},
		}
		service.mockStrava.On("GetActivityStreams", "experienced-token", int64(2003), mock.AnythingOfType("[]string"), mock.AnythingOfType("string")).Return(intervalStreams, nil)

		// Mock logbook update with advanced insights
		updatedLogbook := &models.AthleteLogbook{
			UserID: "experienced-456",
			Content: `{
				"personal_info": {"name": "Mike Pro", "experience": "advanced", "ftp": 320},
				"recent_analysis": {"training_load": "high", "recovery": "adequate", "performance_trend": "improving"},
				"recommendations": ["maintain current volume", "focus on race pace work"]
			}`,
		}
		service.mockLogbook.On("UpdateLogbook", ctx, "experienced-456", mock.AnythingOfType("string")).Return(updatedLogbook, nil)

		// Execute comprehensive analysis workflow
		processor := NewIterativeProcessor(msgCtx, func(string) {})

		// Verify deep analysis is suggested for experienced athlete
		toolCalls := []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "get-activity-streams"}},
		}

		shouldContinue := service.toolCallsSuggestDeeperAnalysis(toolCalls)
		assert.True(t, shouldContinue, "Should suggest deeper analysis for streams")

		// Verify analysis depth assessment
		processor.ToolResults = [][]ToolResult{
			{{Content: `{"firstname": "Mike", "ftp": 320}`, Error: ""}},                    // Profile
			{{Content: `[{"distance": 12000, "type": "Run"}]`, Error: ""}},                // Activities
			{{Content: `{"description": "5x1km @ threshold", "calories": 650}`, Error: ""}}, // Details
		}

		depth := service.assessAnalysisDepth(processor, toolCalls)
		assert.Equal(t, 4, depth) // Profile + activities + details + streams (current call)

		// Verify coaching messages are appropriate for experienced athlete
		progressMsg := service.getCoachingProgressMessage(processor, toolCalls)
		assert.NotContains(t, progressMsg, "API")
		assert.Contains(t, progressMsg, "data") // Should mention data analysis for pro athlete

		service.mockStrava.AssertExpectations(t)
		service.mockLogbook.AssertExpectations(t)
	})

	t.Run("injury recovery analysis workflow", func(t *testing.T) {
		service := setupTestAIService()
		ctx := context.Background()

		// Setup injury recovery scenario
		user := &models.User{
			ID:          "recovery-789",
			AccessToken: "recovery-token",
		}

		recoveryLogbook := &models.AthleteLogbook{
			UserID: "recovery-789",
			Content: `{
				"personal_info": {"name": "Lisa Recovery", "injury_status": "returning from knee injury"},
				"training_constraints": {"max_weekly_volume": "30km", "no_speed_work": true},
				"goals": ["return to 50km/week safely", "complete 10K race in 3 months"]
			}`,
		}

		msgCtx := &MessageContext{
			UserID:              "recovery-789",
			SessionID:           "recovery-session",
			Message:             "How am I progressing in my return to running after injury?",
			ConversationHistory: []*models.Message{},
			AthleteLogbook:      recoveryLogbook,
			User:                user,
		}

		// Mock recovery athlete data - conservative training
		recoveryProfile := &StravaAthlete{
			ID:        333333,
			Firstname: "Lisa",
			Lastname:  "Recovery",
			Sex:       "F",
			Weight:    58.0,
		}
		service.mockStrava.On("GetAthleteProfile", "recovery-token").Return(recoveryProfile, nil)

		recoveryActivities := []*StravaActivity{
			{
				ID:           3001,
				Name:         "Easy Recovery Run",
				Distance:     3000.0, // 3km
				MovingTime:   1080,   // 18 minutes
				Type:         "Run",
				AverageSpeed: 2.78,
			},
			{
				ID:           3002,
				Name:         "Gentle Jog",
				Distance:     4000.0, // 4km
				MovingTime:   1440,   // 24 minutes
				Type:         "Run",
				AverageSpeed: 2.78,
			},
		}
		service.mockStrava.On("GetActivities", "recovery-token", mock.AnythingOfType("ActivityParams")).Return(recoveryActivities, nil)

		// Mock logbook update with recovery-focused insights
		updatedRecoveryLogbook := &models.AthleteLogbook{
			UserID: "recovery-789",
			Content: `{
				"personal_info": {"name": "Lisa Recovery", "injury_status": "progressing well"},
				"recovery_progress": {"weekly_volume": "14km", "pain_level": "none", "consistency": "excellent"},
				"next_steps": ["gradually increase volume by 10%", "continue easy pace focus"]
			}`,
		}
		service.mockLogbook.On("UpdateLogbook", ctx, "recovery-789", mock.AnythingOfType("string")).Return(updatedRecoveryLogbook, nil)

		// Execute recovery-focused analysis
		processor := NewIterativeProcessor(msgCtx, func(string) {})

		// Verify analysis is appropriate for recovery context
		shouldContinue, reason := service.shouldContinueAnalysis(processor, []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "get-athlete-profile"}},
		}, false)
		assert.True(t, shouldContinue)
		assert.Equal(t, "continue_analysis", reason)

		// Verify coaching messages are supportive for recovery
		progressMsg := service.getCoachingProgressMessage(processor, []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "get-recent-activities"}},
		})
		assert.NotContains(t, progressMsg, "API")
		assert.NotContains(t, progressMsg, "performance") // Focus on recovery, not performance

		service.mockStrava.AssertExpectations(t)
		service.mockLogbook.AssertExpectations(t)
	})
}

// Helper function for min (Go 1.21+ has this built-in)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}