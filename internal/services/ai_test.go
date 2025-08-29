package services

import (
	"testing"

	"bodda/internal/config"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestIterativeProcessor_Creation(t *testing.T) {
	msgCtx := &MessageContext{
		UserID:    "test-user",
		SessionID: "test-session",
		Message:   "Test message",
	}

	processor := NewIterativeProcessor(msgCtx, nil)

	assert.NotNil(t, processor)
	assert.Equal(t, 10, processor.MaxRounds)
	assert.Equal(t, 0, processor.CurrentRound)
	assert.Equal(t, msgCtx, processor.Context)
	assert.NotNil(t, processor.ToolResults)
	assert.NotNil(t, processor.Messages)
}

func TestIterativeProcessor_AddToolResults(t *testing.T) {
	processor := NewIterativeProcessor(&MessageContext{}, nil)

	results := []ToolResult{
		{ToolCallID: "call1", Content: "result1"},
		{ToolCallID: "call2", Content: "result2"},
	}

	processor.AddToolResults(results)

	assert.Equal(t, 1, processor.CurrentRound)
	assert.Len(t, processor.ToolResults, 1)
	assert.Len(t, processor.ToolResults[0], 2)
	assert.Equal(t, 2, processor.GetTotalToolCalls())
}

func TestIterativeProcessor_GetTotalToolCalls(t *testing.T) {
	processor := NewIterativeProcessor(&MessageContext{}, nil)

	// Add first round
	processor.AddToolResults([]ToolResult{
		{ToolCallID: "call1", Content: "result1"},
		{ToolCallID: "call2", Content: "result2"},
	})

	// Add second round
	processor.AddToolResults([]ToolResult{
		{ToolCallID: "call3", Content: "result3"},
	})

	assert.Equal(t, 3, processor.GetTotalToolCalls())
}

func TestAIService_ValidateMessageContext(t *testing.T) {
	service := &aiService{}

	t.Run("valid context", func(t *testing.T) {
		msgCtx := &MessageContext{
			UserID:    "test-user",
			SessionID: "test-session",
			Message:   "Test message",
		}

		err := service.validateMessageContext(msgCtx)
		assert.NoError(t, err)
	})

	t.Run("nil context", func(t *testing.T) {
		err := service.validateMessageContext(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid input provided")
	})

	t.Run("empty user ID", func(t *testing.T) {
		msgCtx := &MessageContext{
			SessionID: "test-session",
			Message:   "Test message",
		}

		err := service.validateMessageContext(msgCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID is required")
	})

	t.Run("empty session ID", func(t *testing.T) {
		msgCtx := &MessageContext{
			UserID:  "test-user",
			Message: "Test message",
		}

		err := service.validateMessageContext(msgCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session ID is required")
	})

	t.Run("empty message", func(t *testing.T) {
		msgCtx := &MessageContext{
			UserID:    "test-user",
			SessionID: "test-session",
			Message:   "",
		}

		err := service.validateMessageContext(msgCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message content cannot be empty")
	})
}

func TestAIService_ContextRedactionIntegration(t *testing.T) {
	// Create a mock AI service with redaction enabled
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			RedactionEnabled:  true,
			MaxContextTokens:  15000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1000,
			MaxPageSize:       5000,
		},
	}

	// Create AI service (this will initialize the context manager)
	service := NewAIService(cfg, nil, nil)
	aiSvc := service.(*aiService)

	// Verify context manager was created
	assert.NotNil(t, aiSvc.contextManager)

	// Test that the context manager correctly identifies stream tools
	assert.True(t, aiSvc.contextManager.ShouldRedact("get-activity-streams"))
	assert.False(t, aiSvc.contextManager.ShouldRedact("get-athlete-profile"))
}

func TestAIService_ContextRedactionDisabled(t *testing.T) {
	// Create a mock AI service with redaction disabled
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			RedactionEnabled:  false,
			MaxContextTokens:  15000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1000,
			MaxPageSize:       5000,
		},
	}

	// Create AI service
	service := NewAIService(cfg, nil, nil)
	aiSvc := service.(*aiService)

	// Verify context manager was created but redaction is disabled
	assert.NotNil(t, aiSvc.contextManager)

	// Test redaction behavior when disabled
	originalMessages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleAssistant,
			ToolCalls: []openai.ToolCall{
				{
					ID:   "call_123",
					Type: "function",
					Function: openai.FunctionCall{
						Name: "get-activity-streams",
					},
				},
			},
		},
		{
			Role:       openai.ChatMessageRoleTool,
			Content:    "📊 Stream Data with detailed information...",
			ToolCallID: "call_123",
		},
	}

	result := aiSvc.contextManager.RedactPreviousStreamOutputs(originalMessages)

	// Should return unchanged messages when redaction is disabled
	assert.Equal(t, len(originalMessages), len(result))
	assert.Equal(t, originalMessages[1].Content, result[1].Content)
}

func TestAIService_MultipleStreamToolCallsRedaction(t *testing.T) {
	// Create a mock AI service with redaction enabled
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

	// Simulate a conversation with multiple stream tool calls
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "Show me stream data for activity 1",
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
			},
		},
		{
			Role:       openai.ChatMessageRoleTool,
			Content:    "📊 First Stream Data\nHeart rate: 150-180 bpm\nPower: 200-300W\nDetailed analysis with many lines of data...",
			ToolCallID: "call_stream_1",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "Now show me stream data for activity 2",
		},
		{
			Role: openai.ChatMessageRoleAssistant,
			ToolCalls: []openai.ToolCall{
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
			Content:    "📊 Second Stream Data\nHeart rate: 140-170 bpm\nPower: 180-280W\nAnother detailed analysis...",
			ToolCallID: "call_stream_2",
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
			Content:    "Athlete Profile: John Doe, 35 years old, cyclist...",
			ToolCallID: "call_profile",
		},
	}

	// Apply redaction
	result := aiSvc.contextManager.RedactPreviousStreamOutputs(messages)

	// Should have same number of messages
	assert.Equal(t, len(messages), len(result))

	// User messages should be unchanged
	assert.Equal(t, messages[0].Content, result[0].Content)
	assert.Equal(t, messages[3].Content, result[3].Content)

	// Assistant messages should be unchanged
	assert.Equal(t, len(messages[1].ToolCalls), len(result[1].ToolCalls))
	assert.Equal(t, len(messages[4].ToolCalls), len(result[4].ToolCalls))
	assert.Equal(t, len(messages[6].ToolCalls), len(result[6].ToolCalls))

	// First stream tool result should be redacted (followed by user message)
	assert.Contains(t, result[2].Content, "[Previous Stream Analysis - Redacted")
	assert.NotContains(t, result[2].Content, "Heart rate: 150-180 bpm")
	
	// Second stream tool result should NOT be redacted (only followed by tool calls)
	assert.Equal(t, messages[5].Content, result[5].Content)
	assert.Contains(t, result[5].Content, "Heart rate: 140-170 bpm")

	// Non-stream tool result should not be redacted
	assert.Equal(t, messages[7].Content, result[7].Content)
	assert.Contains(t, result[7].Content, "Athlete Profile: John Doe")
}