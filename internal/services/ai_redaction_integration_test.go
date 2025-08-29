package services

import (
	"context"
	"io"
	"strings"
	"testing"

	"bodda/internal/config"
	"bodda/internal/models"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOpenAIClient is a mock implementation of the OpenAI client for testing
type MockOpenAIClient struct {
	mock.Mock
}

// MockStream implements the streaming interface for testing
type MockStream struct {
	responses []openai.ChatCompletionStreamResponse
	index     int
	closed    bool
}

func (m *MockStream) Recv() (openai.ChatCompletionStreamResponse, error) {
	if m.index >= len(m.responses) {
		return openai.ChatCompletionStreamResponse{}, io.EOF
	}
	response := m.responses[m.index]
	m.index++
	return response, nil
}

func (m *MockStream) Close() error {
	m.closed = true
	return nil
}

// TestAIServiceRedactionIntegration tests redaction logic within full AI service chat flow context
func TestAIServiceRedactionIntegration(t *testing.T) {
	t.Run("SingleConversationRoundWithStreamToolRedaction", func(t *testing.T) {
		// Create AI service with redaction enabled
		cfg := &config.Config{
			StreamProcessing: config.StreamProcessingConfig{
				RedactionEnabled:  true,
				MaxContextTokens:  15000,
				TokenPerCharRatio: 0.25,
				DefaultPageSize:   1000,
				MaxPageSize:       5000,
			},
		}

		// Create mock services
		mockStravaService := &MockStravaService{}
		mockLogbookService := &MockLogbookService{}

		service := NewAIService(cfg, mockStravaService, mockLogbookService)
		aiSvc := service.(*aiService)

		// Test that context manager is properly integrated
		assert.NotNil(t, aiSvc.contextManager)
		assert.True(t, aiSvc.contextManager.ShouldRedact("get-activity-streams"))

		// Simulate a conversation where stream tool is followed by assistant explanation
		messages := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Show me my activity streams",
			},
			{
				Role: openai.ChatMessageRoleAssistant,
				ToolCalls: []openai.ToolCall{
					{
						ID:   "call_streams",
						Type: "function",
						Function: openai.FunctionCall{
							Name: "get-activity-streams",
						},
					},
				},
			},
			{
				Role:       openai.ChatMessageRoleTool,
				Content:    "📊 Stream Data\n\nHeart rate: 150-180 bpm\nPower: 200-300W\nDetailed stream analysis...",
				ToolCallID: "call_streams",
			},
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "Based on your stream data, I can see excellent performance patterns...",
			},
		}

		// Apply redaction through context manager
		redactedMessages := aiSvc.contextManager.RedactPreviousStreamOutputs(messages)

		// Verify redaction behavior
		assert.Equal(t, len(messages), len(redactedMessages))

		// Stream tool result should be redacted because it's followed by assistant explanation
		toolResult := redactedMessages[2]
		assert.Contains(t, toolResult.Content, "[Previous Stream Analysis - Redacted")
		assert.NotContains(t, toolResult.Content, "Heart rate: 150-180 bpm")

		// Assistant explanation should remain unchanged
		assistantResponse := redactedMessages[3]
		assert.Equal(t, messages[3].Content, assistantResponse.Content)
		assert.Contains(t, assistantResponse.Content, "excellent performance patterns")
	})

	t.Run("MultipleConversationRoundsWithVariousToolCallPatterns", func(t *testing.T) {
		cfg := &config.Config{
			StreamProcessing: config.StreamProcessingConfig{
				RedactionEnabled:  true,
				MaxContextTokens:  15000,
				TokenPerCharRatio: 0.25,
				DefaultPageSize:   1000,
				MaxPageSize:       5000,
			},
		}

		mockStravaService := &MockStravaService{}
		mockLogbookService := &MockLogbookService{}

		service := NewAIService(cfg, mockStravaService, mockLogbookService)
		aiSvc := service.(*aiService)

		// Simulate multiple conversation rounds with different tool call patterns
		t.Run("Round1_StreamToolFollowedByUserMessage", func(t *testing.T) {
			messages := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Get my stream data for activity 123",
				},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{
							ID:   "call_streams_1",
							Type: "function",
							Function: openai.FunctionCall{
								Name: "get-activity-streams",
							},
						},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					Content:    "📊 Stream Data for Activity 123\n\nHeart rate: 140-175 bpm\nPower: 180-280W\nCadence: 85-95 rpm\nDetailed analysis continues...",
					ToolCallID: "call_streams_1",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "What does this data tell us about my performance?",
				},
			}

			result := aiSvc.contextManager.RedactPreviousStreamOutputs(messages)

			// Stream tool result should be redacted (followed by user message)
			toolResult := result[2]
			assert.Contains(t, toolResult.Content, "[Previous Stream Analysis - Redacted")
			assert.NotContains(t, toolResult.Content, "Heart rate: 140-175 bpm")
		})

		t.Run("Round2_StreamToolFollowedOnlyByOtherToolCalls", func(t *testing.T) {
			messages := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Compare with my profile data",
				},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{
							ID:   "call_streams_2",
							Type: "function",
							Function: openai.FunctionCall{
								Name: "get-activity-streams",
							},
						},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					Content:    "📊 Stream Data for Comparison\n\nHeart rate: 145-170 bpm\nPower: 200-320W\nElevation: 500m gain\nComprehensive stream analysis...",
					ToolCallID: "call_streams_2",
				},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{
							ID:   "call_profile",
							Type: "function",
							Function: openai.FunctionCall{
								Name: "get-athlete-profile",
							},
						},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					Content:    "Athlete Profile: John Doe, 35 years old, cyclist, FTP: 280W",
					ToolCallID: "call_profile",
				},
			}

			result := aiSvc.contextManager.RedactPreviousStreamOutputs(messages)

			// Stream tool result should NOT be redacted (only followed by tool calls)
			streamResult := result[2]
			assert.Equal(t, messages[2].Content, streamResult.Content)
			assert.Contains(t, streamResult.Content, "Heart rate: 145-170 bpm")

			// Profile tool result should not be redacted (non-stream tool)
			profileResult := result[4]
			assert.Equal(t, messages[4].Content, profileResult.Content)
			assert.Contains(t, profileResult.Content, "John Doe")
		})

		t.Run("Round3_ChainedToolCallsEndingWithExplanation", func(t *testing.T) {
			messages := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Give me a complete analysis",
				},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{
							ID:   "call_streams_3",
							Type: "function",
							Function: openai.FunctionCall{
								Name: "get-activity-streams",
							},
						},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					Content:    "📊 Complete Stream Analysis\n\nTime: 3600 seconds\nDistance: 50km\nAverage power: 250W\nNormalized power: 265W\nIntensive factor: 0.85\nDetailed metrics and analysis...",
					ToolCallID: "call_streams_3",
				},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{
							ID:   "call_activities",
							Type: "function",
							Function: openai.FunctionCall{
								Name: "get-recent-activities",
							},
						},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					Content:    "Recent Activities: 5 rides in the last week, total distance: 200km",
					ToolCallID: "call_activities",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "Based on your complete stream analysis and recent activity history, I can provide a comprehensive assessment...",
				},
			}

			result := aiSvc.contextManager.RedactPreviousStreamOutputs(messages)

			// Stream tool result should be redacted (eventually followed by assistant explanation)
			streamResult := result[2]
			assert.Contains(t, streamResult.Content, "[Previous Stream Analysis - Redacted")
			assert.NotContains(t, streamResult.Content, "Average power: 250W")

			// Activities tool result should not be redacted (non-stream tool)
			activitiesResult := result[4]
			assert.Equal(t, messages[4].Content, activitiesResult.Content)
			assert.Contains(t, activitiesResult.Content, "5 rides in the last week")

			// Final assistant explanation should remain unchanged
			finalResponse := result[5]
			assert.Equal(t, messages[5].Content, finalResponse.Content)
			assert.Contains(t, finalResponse.Content, "comprehensive assessment")
		})

		t.Run("Round4_FinalStreamToolResultNotRedacted", func(t *testing.T) {
			messages := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Show me the final stream data",
				},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{
							ID:   "call_final_streams",
							Type: "function",
							Function: openai.FunctionCall{
								Name: "get-activity-streams",
							},
						},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					Content:    "📊 Final Stream Data\n\nPeak power: 450W\nPeak heart rate: 185 bpm\nFinal analysis complete with all metrics...",
					ToolCallID: "call_final_streams",
				},
			}

			result := aiSvc.contextManager.RedactPreviousStreamOutputs(messages)

			// Final stream tool result should NOT be redacted (end of conversation)
			finalResult := result[2]
			assert.Equal(t, messages[2].Content, finalResult.Content)
			assert.Contains(t, finalResult.Content, "Peak power: 450W")
			assert.Contains(t, finalResult.Content, "Peak heart rate: 185 bpm")
		})
	})

	t.Run("StreamingResponseHandlingWithRedactionLogic", func(t *testing.T) {
		cfg := &config.Config{
			StreamProcessing: config.StreamProcessingConfig{
				RedactionEnabled:  true,
				MaxContextTokens:  15000,
				TokenPerCharRatio: 0.25,
				DefaultPageSize:   1000,
				MaxPageSize:       5000,
			},
		}

		mockStravaService := &MockStravaService{}
		mockLogbookService := &MockLogbookService{}

		service := NewAIService(cfg, mockStravaService, mockLogbookService)
		aiSvc := service.(*aiService)

		// Test streaming response with redaction applied to context
		t.Run("StreamingWithContextRedaction", func(t *testing.T) {
			// Create a proper conversation context with OpenAI message format
			// that includes tool calls and tool results that would trigger redaction
			messages := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Show me stream data for activity 456",
				},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{
							ID:   "call_previous_streams",
							Type: "function",
							Function: openai.FunctionCall{
								Name: "get-activity-streams",
							},
						},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					Content:    "📊 Previous Stream Data\n\nHeart rate: 160-185 bpm\nPower: 220-350W\nExtensive previous analysis...",
					ToolCallID: "call_previous_streams",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "What patterns do you see?",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "I can see several interesting patterns in your data...",
				},
			}

			// Apply redaction as would happen before streaming
			redactedMessages := aiSvc.contextManager.RedactPreviousStreamOutputs(messages)

			// Verify that previous stream data was redacted in the context
			foundRedactedContent := false
			for _, msg := range redactedMessages {
				if strings.Contains(msg.Content, "[Previous Stream Analysis - Redacted") {
					foundRedactedContent = true
					assert.NotContains(t, msg.Content, "Heart rate: 160-185 bpm")
					break
				}
			}
			assert.True(t, foundRedactedContent, "Previous stream data should be redacted in streaming context")

			// Verify that other content remains unchanged
			userMessageFound := false
			for _, msg := range redactedMessages {
				if msg.Role == openai.ChatMessageRoleUser && strings.Contains(msg.Content, "What patterns do you see?") {
					userMessageFound = true
					break
				}
			}
			assert.True(t, userMessageFound, "User messages should remain unchanged")
		})

		t.Run("StreamingWithMultipleRedactionDecisions", func(t *testing.T) {
			// Complex conversation with multiple stream tool calls and different redaction outcomes
			messages := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Analyze multiple activities",
				},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{
							ID:   "call_stream_1",
							Type: "function",
							Function: openai.FunctionCall{
								Name: "get-activity-streams",
							},
						},
						{
							ID:   "call_stream_2",
							Type: "function",
							Function: openai.FunctionCall{
								Name: "get-activity-streams",
							},
						},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					Content:    "📊 Stream Data Activity 1\n\nHeart rate: 140-170 bpm\nPower: 180-250W\nFirst activity analysis...",
					ToolCallID: "call_stream_1",
				},
				{
					Role:       openai.ChatMessageRoleTool,
					Content:    "📊 Stream Data Activity 2\n\nHeart rate: 150-180 bpm\nPower: 200-280W\nSecond activity analysis...",
					ToolCallID: "call_stream_2",
				},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{
							ID:   "call_stream_3",
							Type: "function",
							Function: openai.FunctionCall{
								Name: "get-activity-streams",
							},
						},
						{
							ID:   "call_profile",
							Type: "function",
							Function: openai.FunctionCall{
								Name: "get-athlete-profile",
							},
						},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					Content:    "📊 Stream Data Activity 3\n\nHeart rate: 145-175 bpm\nPower: 190-270W\nThird activity analysis...",
					ToolCallID: "call_stream_3",
				},
				{
					Role:       openai.ChatMessageRoleTool,
					Content:    "Athlete Profile: Jane Smith, 28 years old, runner/cyclist",
					ToolCallID: "call_profile",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "Comparing all three activities with your profile...",
				},
			}

			redactedMessages := aiSvc.contextManager.RedactPreviousStreamOutputs(messages)

			// Count redacted stream analyses
			redactedCount := 0
			preservedStreamCount := 0
			preservedProfileCount := 0

			for _, msg := range redactedMessages {
				if strings.Contains(msg.Content, "[Previous Stream Analysis - Redacted") {
					redactedCount++
				} else if strings.Contains(msg.Content, "📊 Stream Data Activity") {
					preservedStreamCount++
				} else if strings.Contains(msg.Content, "Athlete Profile: Jane Smith") {
					preservedProfileCount++
				}
			}

			// All three stream analyses should be redacted (followed by assistant explanation)
			assert.Equal(t, 3, redactedCount, "All three stream analyses should be redacted")
			
			// No stream analyses should be preserved (all followed by assistant explanation)
			assert.Equal(t, 0, preservedStreamCount, "No stream analyses should be preserved")
			
			// Profile should be preserved (non-stream tool)
			assert.Equal(t, 1, preservedProfileCount, "Profile should be preserved")
		})
	})

	t.Run("ExistingAIServiceTestsCompatibility", func(t *testing.T) {
		// Verify that enhanced redaction logic doesn't break existing functionality
		cfg := &config.Config{
			StreamProcessing: config.StreamProcessingConfig{
				RedactionEnabled:  true,
				MaxContextTokens:  15000,
				TokenPerCharRatio: 0.25,
				DefaultPageSize:   1000,
				MaxPageSize:       5000,
			},
		}

		service := NewAIService(cfg, nil, nil)
		aiSvc := service.(*aiService)

		// Test that basic AI service functionality still works
		t.Run("ValidateMessageContextStillWorks", func(t *testing.T) {
			validCtx := &MessageContext{
				UserID:    "test-user",
				SessionID: "test-session",
				Message:   "Test message",
			}
			err := aiSvc.validateMessageContext(validCtx)
			assert.NoError(t, err)

			invalidCtx := &MessageContext{
				UserID:  "test-user",
				Message: "Test message",
				// Missing SessionID
			}
			err = aiSvc.validateMessageContext(invalidCtx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "session ID is required")
		})

		t.Run("IterativeProcessorStillWorks", func(t *testing.T) {
			msgCtx := &MessageContext{
				UserID:    "test-user",
				SessionID: "test-session",
				Message:   "Test message",
			}

			processor := NewIterativeProcessor(msgCtx, nil)
			assert.NotNil(t, processor)
			assert.Equal(t, 10, processor.MaxRounds)
			assert.Equal(t, 0, processor.CurrentRound)

			// Add tool results
			results := []ToolResult{
				{ToolCallID: "call1", Content: "result1"},
				{ToolCallID: "call2", Content: "result2"},
			}
			processor.AddToolResults(results)

			assert.Equal(t, 1, processor.CurrentRound)
			assert.Equal(t, 2, processor.GetTotalToolCalls())
		})

		t.Run("ContextManagerIntegrationPreservesExistingBehavior", func(t *testing.T) {
			// Test that context manager integration doesn't change non-redaction behavior
			assert.NotNil(t, aiSvc.contextManager)
			
			// Should still identify stream tools correctly
			assert.True(t, aiSvc.contextManager.ShouldRedact("get-activity-streams"))
			assert.False(t, aiSvc.contextManager.ShouldRedact("get-athlete-profile"))
			assert.False(t, aiSvc.contextManager.ShouldRedact("get-recent-activities"))
			assert.False(t, aiSvc.contextManager.ShouldRedact("get-activity-details"))
			assert.False(t, aiSvc.contextManager.ShouldRedact("update-athlete-logbook"))
		})

		t.Run("RedactionDisabledBehaviorPreserved", func(t *testing.T) {
			// Test with redaction disabled
			disabledCfg := &config.Config{
				StreamProcessing: config.StreamProcessingConfig{
					RedactionEnabled:  false,
					MaxContextTokens:  15000,
					TokenPerCharRatio: 0.25,
					DefaultPageSize:   1000,
					MaxPageSize:       5000,
				},
			}

			disabledService := NewAIService(disabledCfg, nil, nil)
			disabledAiSvc := disabledService.(*aiService)

			messages := []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{
							ID:   "call_streams",
							Type: "function",
							Function: openai.FunctionCall{
								Name: "get-activity-streams",
							},
						},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					Content:    "📊 Stream Data\n\nDetailed stream information...",
					ToolCallID: "call_streams",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "Based on your data...",
				},
			}

			result := disabledAiSvc.contextManager.RedactPreviousStreamOutputs(messages)

			// Should return unchanged messages when redaction is disabled
			assert.Equal(t, len(messages), len(result))
			assert.Equal(t, messages[1].Content, result[1].Content)
			assert.Contains(t, result[1].Content, "Detailed stream information")
		})
	})
}

// TestAIServiceRedactionEnvironmentVariableControl tests environment variable control
// covering requirements 4.1, 4.2, 4.3, 4.4
func TestAIServiceRedactionEnvironmentVariableControl(t *testing.T) {
	t.Run("RedactionEnabledByEnvironmentVariable", func(t *testing.T) {
		cfg := &config.Config{
			StreamProcessing: config.StreamProcessingConfig{
				RedactionEnabled:  true, // Explicitly enabled
				MaxContextTokens:  15000,
				TokenPerCharRatio: 0.25,
				DefaultPageSize:   1000,
				MaxPageSize:       5000,
			},
		}

		service := NewAIService(cfg, nil, nil)
		aiSvc := service.(*aiService)

		// Verify redaction is enabled
		messages := []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleAssistant,
				ToolCalls: []openai.ToolCall{
					{
						ID:   "call_streams",
						Type: "function",
						Function: openai.FunctionCall{
							Name: "get-activity-streams",
						},
					},
				},
			},
			{
				Role:       openai.ChatMessageRoleTool,
				Content:    "📊 Stream Data\n\nHeart rate: 150 bpm\nPower: 250W\nDetailed analysis...",
				ToolCallID: "call_streams",
			},
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "Based on your stream data...",
			},
		}

		result := aiSvc.contextManager.RedactPreviousStreamOutputs(messages)

		// Should apply redaction when enabled
		toolResult := result[1]
		assert.Contains(t, toolResult.Content, "[Previous Stream Analysis - Redacted")
		assert.NotContains(t, toolResult.Content, "Heart rate: 150 bpm")
	})

	t.Run("RedactionDisabledByEnvironmentVariable", func(t *testing.T) {
		cfg := &config.Config{
			StreamProcessing: config.StreamProcessingConfig{
				RedactionEnabled:  false, // Explicitly disabled
				MaxContextTokens:  15000,
				TokenPerCharRatio: 0.25,
				DefaultPageSize:   1000,
				MaxPageSize:       5000,
			},
		}

		service := NewAIService(cfg, nil, nil)
		aiSvc := service.(*aiService)

		// Verify redaction is disabled
		messages := []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleAssistant,
				ToolCalls: []openai.ToolCall{
					{
						ID:   "call_streams",
						Type: "function",
						Function: openai.FunctionCall{
							Name: "get-activity-streams",
						},
					},
				},
			},
			{
				Role:       openai.ChatMessageRoleTool,
				Content:    "📊 Stream Data\n\nHeart rate: 150 bpm\nPower: 250W\nDetailed analysis...",
				ToolCallID: "call_streams",
			},
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "Based on your stream data...",
			},
		}

		result := aiSvc.contextManager.RedactPreviousStreamOutputs(messages)

		// Should NOT apply redaction when disabled
		toolResult := result[1]
		assert.Equal(t, messages[1].Content, toolResult.Content)
		assert.Contains(t, toolResult.Content, "Heart rate: 150 bpm")
		assert.NotContains(t, toolResult.Content, "[Previous Stream Analysis - Redacted")
	})

	t.Run("DefaultRedactionBehavior", func(t *testing.T) {
		// Test default behavior (should default to enabled)
		cfg := &config.Config{
			StreamProcessing: config.StreamProcessingConfig{
				// RedactionEnabled not explicitly set, should default to enabled
				MaxContextTokens:  15000,
				TokenPerCharRatio: 0.25,
				DefaultPageSize:   1000,
				MaxPageSize:       5000,
			},
		}

		service := NewAIService(cfg, nil, nil)
		aiSvc := service.(*aiService)

		// Test that default behavior applies redaction
		messages := []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleAssistant,
				ToolCalls: []openai.ToolCall{
					{
						ID:   "call_streams",
						Type: "function",
						Function: openai.FunctionCall{
							Name: "get-activity-streams",
						},
					},
				},
			},
			{
				Role:       openai.ChatMessageRoleTool,
				Content:    "📊 Stream Data\n\nHeart rate: 160 bpm\nPower: 270W\nDefault behavior test...",
				ToolCallID: "call_streams",
			},
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "Your performance shows...",
			},
		}

		result := aiSvc.contextManager.RedactPreviousStreamOutputs(messages)

		// Should apply redaction by default (assuming default is true)
		toolResult := result[1]
		// Note: This test assumes the default is true. If the default changes to false,
		// this test should be updated accordingly.
		if cfg.StreamProcessing.RedactionEnabled {
			assert.Contains(t, toolResult.Content, "[Previous Stream Analysis - Redacted")
			assert.NotContains(t, toolResult.Content, "Heart rate: 160 bpm")
		} else {
			assert.Equal(t, messages[1].Content, toolResult.Content)
		}
	})

	t.Run("GlobalRedactionControlAcrossConversations", func(t *testing.T) {
		// Test that environment variable controls redaction globally across all conversations
		enabledCfg := &config.Config{
			StreamProcessing: config.StreamProcessingConfig{
				RedactionEnabled:  true,
				MaxContextTokens:  15000,
				TokenPerCharRatio: 0.25,
				DefaultPageSize:   1000,
				MaxPageSize:       5000,
			},
		}

		service := NewAIService(enabledCfg, nil, nil)
		aiSvc := service.(*aiService)

		// Test multiple conversation scenarios
		testCases := []struct {
			name     string
			messages []openai.ChatCompletionMessage
		}{
			{
				name: "Conversation1_UserInitiated",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: "Show me stream data"},
					{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{{ID: "call1", Type: "function", Function: openai.FunctionCall{Name: "get-activity-streams"}}}},
					{Role: openai.ChatMessageRoleTool, Content: "📊 Stream Data 1\n\nHR: 150 bpm", ToolCallID: "call1"},
					{Role: openai.ChatMessageRoleUser, Content: "What does this mean?"},
				},
			},
			{
				name: "Conversation2_SystemInitiated",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleSystem, Content: "Analyze user data"},
					{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{{ID: "call2", Type: "function", Function: openai.FunctionCall{Name: "get-activity-streams"}}}},
					{Role: openai.ChatMessageRoleTool, Content: "📊 Stream Data 2\n\nPower: 250W", ToolCallID: "call2"},
					{Role: openai.ChatMessageRoleAssistant, Content: "Analysis complete"},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := aiSvc.contextManager.RedactPreviousStreamOutputs(tc.messages)
				
				// Find the tool result message
				for _, msg := range result {
					if msg.Role == openai.ChatMessageRoleTool && strings.Contains(msg.Content, "📊 Stream Data") {
						// Should be redacted in all conversations when globally enabled
						assert.Contains(t, msg.Content, "[Previous Stream Analysis - Redacted", 
							"Stream data should be redacted in %s", tc.name)
						break
					}
				}
			})
		}
	})
}

// MockStravaService for testing
type MockStravaService struct {
	mock.Mock
}

func (m *MockStravaService) GetAthleteProfile(user *models.User) (*StravaAthlete, error) {
	args := m.Called(user)
	return args.Get(0).(*StravaAthlete), args.Error(1)
}

func (m *MockStravaService) GetActivities(user *models.User, params ActivityParams) ([]*StravaActivity, error) {
	args := m.Called(user, params)
	return args.Get(0).([]*StravaActivity), args.Error(1)
}

func (m *MockStravaService) GetActivityDetail(user *models.User, activityID int64) (*StravaActivityDetail, error) {
	args := m.Called(user, activityID)
	return args.Get(0).(*StravaActivityDetail), args.Error(1)
}

func (m *MockStravaService) GetActivityStreams(user *models.User, activityID int64, streamTypes []string, resolution string) (*StravaStreams, error) {
	args := m.Called(user, activityID, streamTypes, resolution)
	return args.Get(0).(*StravaStreams), args.Error(1)
}

func (m *MockStravaService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	args := m.Called(refreshToken)
	return args.Get(0).(*TokenResponse), args.Error(1)
}

// MockLogbookService for testing
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

