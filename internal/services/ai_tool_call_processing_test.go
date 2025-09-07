package services

import (
	"testing"

	"github.com/openai/openai-go/v2/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolCallState(t *testing.T) {
	service := &aiService{}
	
	t.Run("NewToolCallState creates empty state", func(t *testing.T) {
		state := NewToolCallState()
		
		assert.NotNil(t, state)
		assert.Equal(t, 0, service.GetToolCallCount(state))
		assert.Equal(t, 0, service.GetCompletedToolCallCount(state))
		assert.False(t, service.HasPendingToolCalls(state))
	})
}

func TestParseToolCallsFromEvents(t *testing.T) {
	service := &aiService{}
	
	t.Run("handles function call arguments delta event", func(t *testing.T) {
		state := NewToolCallState()
		
		// Test the handleFunctionCallArgumentsDelta method directly
		deltaEvent := responses.ResponseFunctionCallArgumentsDeltaEvent{
			ItemID: "test-tool-call-1",
			Delta:  `{"activity_id": 123`,
		}
		
		err := service.handleFunctionCallArgumentsDelta(deltaEvent, state)
		require.NoError(t, err)
		
		assert.Equal(t, 1, service.GetToolCallCount(state))
		
		toolCall, exists := service.GetToolCallByID(state, "test-tool-call-1")
		require.True(t, exists)
		assert.Equal(t, "test-tool-call-1", toolCall.ID)
		assert.Equal(t, `{"activity_id": 123`, toolCall.Arguments)
	})
	
	t.Run("accumulates arguments across multiple delta events", func(t *testing.T) {
		state := NewToolCallState()
		
		// First delta
		deltaEvent1 := responses.ResponseFunctionCallArgumentsDeltaEvent{
			ItemID: "test-tool-call-1",
			Delta:  `{"activity_id": 123`,
		}
		
		err := service.handleFunctionCallArgumentsDelta(deltaEvent1, state)
		require.NoError(t, err)
		
		// Second delta
		deltaEvent2 := responses.ResponseFunctionCallArgumentsDeltaEvent{
			ItemID: "test-tool-call-1",
			Delta:  `, "stream_types": ["time", "heartrate"]}`,
		}
		
		err = service.handleFunctionCallArgumentsDelta(deltaEvent2, state)
		require.NoError(t, err)
		
		assert.Equal(t, 1, service.GetToolCallCount(state))
		
		toolCall, exists := service.GetToolCallByID(state, "test-tool-call-1")
		require.True(t, exists)
		assert.Equal(t, `{"activity_id": 123, "stream_types": ["time", "heartrate"]}`, toolCall.Arguments)
	})
}

func TestExtractFunctionNameFromArguments(t *testing.T) {
	service := &aiService{}
	
	testCases := []struct {
		name     string
		args     string
		expected string
	}{
		{
			name:     "empty arguments defaults to profile",
			args:     "",
			expected: "get-athlete-profile",
		},
		{
			name:     "empty JSON object defaults to profile",
			args:     "{}",
			expected: "get-athlete-profile",
		},
		{
			name:     "activity_id with stream_types infers streams",
			args:     `{"activity_id": 123, "stream_types": ["time", "heartrate"]}`,
			expected: "get-activity-streams",
		},
		{
			name:     "activity_id with resolution infers streams",
			args:     `{"activity_id": 123, "resolution": "medium"}`,
			expected: "get-activity-streams",
		},
		{
			name:     "activity_id alone infers activity details",
			args:     `{"activity_id": 123}`,
			expected: "get-activity-details",
		},
		{
			name:     "per_page infers recent activities",
			args:     `{"per_page": 30}`,
			expected: "get-recent-activities",
		},
		{
			name:     "content infers logbook update",
			args:     `{"content": "Training notes"}`,
			expected: "update-athlete-logbook",
		},
		{
			name:     "malformed JSON uses string heuristics",
			args:     `{"activity_id": 123, "stream_types"`,
			expected: "get-activity-streams",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.extractFunctionNameFromArguments(tc.args)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidateToolCall(t *testing.T) {
	service := &aiService{}
	
	t.Run("valid tool call passes validation", func(t *testing.T) {
		toolCall := &responses.ResponseFunctionToolCall{
			CallID:    "test-call-id",
			Name:      "get-athlete-profile",
			Arguments: "{}",
		}
		
		assert.True(t, service.validateToolCall(toolCall))
	})
	
	t.Run("missing CallID fails validation", func(t *testing.T) {
		toolCall := &responses.ResponseFunctionToolCall{
			CallID:    "",
			Name:      "get-athlete-profile",
			Arguments: "{}",
		}
		
		assert.False(t, service.validateToolCall(toolCall))
	})
	
	t.Run("missing function name fails validation", func(t *testing.T) {
		toolCall := &responses.ResponseFunctionToolCall{
			CallID:    "test-call-id",
			Name:      "",
			Arguments: "{}",
		}
		
		assert.False(t, service.validateToolCall(toolCall))
	})
	
	t.Run("unknown function name fails validation", func(t *testing.T) {
		toolCall := &responses.ResponseFunctionToolCall{
			CallID:    "test-call-id",
			Name:      "unknown-function",
			Arguments: "{}",
		}
		
		assert.False(t, service.validateToolCall(toolCall))
	})
	
	t.Run("invalid JSON arguments fails validation", func(t *testing.T) {
		toolCall := &responses.ResponseFunctionToolCall{
			CallID:    "test-call-id",
			Name:      "get-athlete-profile",
			Arguments: `{"invalid": json}`,
		}
		
		assert.False(t, service.validateToolCall(toolCall))
	})
}

func TestSanitizeToolCall(t *testing.T) {
	service := &aiService{}
	
	t.Run("sanitizes whitespace", func(t *testing.T) {
		toolCall := responses.ResponseFunctionToolCall{
			CallID:    "  test-call-id  ",
			Name:      "  get-athlete-profile  ",
			Arguments: "  {}  ",
		}
		
		sanitized := service.sanitizeToolCall(toolCall)
		
		assert.Equal(t, "test-call-id", sanitized.CallID)
		assert.Equal(t, "get-athlete-profile", sanitized.Name)
		assert.Equal(t, "{}", sanitized.Arguments)
	})
	
	t.Run("fixes empty arguments", func(t *testing.T) {
		toolCall := responses.ResponseFunctionToolCall{
			CallID:    "test-call-id",
			Name:      "get-athlete-profile",
			Arguments: "",
		}
		
		sanitized := service.sanitizeToolCall(toolCall)
		
		assert.Equal(t, "{}", sanitized.Arguments)
	})
	
	t.Run("fixes invalid JSON arguments", func(t *testing.T) {
		toolCall := responses.ResponseFunctionToolCall{
			CallID:    "test-call-id",
			Name:      "get-athlete-profile",
			Arguments: `{"invalid": json}`,
		}
		
		sanitized := service.sanitizeToolCall(toolCall)
		
		assert.Equal(t, "{}", sanitized.Arguments)
	})
}

func TestGetCompletedToolCalls(t *testing.T) {
	service := &aiService{}
	
	t.Run("returns valid tool calls with inferred function names", func(t *testing.T) {
		state := NewToolCallState()
		
		// Add a tool call with arguments but no function name
		state.toolCalls["test-1"] = &responses.ResponseFunctionToolCall{
			ID:        "test-1",
			CallID:    "test-1",
			Name:      "", // Will be inferred
			Arguments: `{"activity_id": 123}`,
		}
		
		// Add a tool call with function name
		state.toolCalls["test-2"] = &responses.ResponseFunctionToolCall{
			ID:        "test-2",
			CallID:    "test-2",
			Name:      "get-recent-activities",
			Arguments: `{"per_page": 30}`,
		}
		
		completed := service.GetCompletedToolCalls(state)
		
		assert.Len(t, completed, 2)
		
		// Find the tool calls by ID
		var toolCall1, toolCall2 *responses.ResponseFunctionToolCall
		for i := range completed {
			if completed[i].ID == "test-1" {
				toolCall1 = &completed[i]
			} else if completed[i].ID == "test-2" {
				toolCall2 = &completed[i]
			}
		}
		
		require.NotNil(t, toolCall1)
		assert.Equal(t, "get-activity-details", toolCall1.Name) // Inferred from arguments
		
		require.NotNil(t, toolCall2)
		assert.Equal(t, "get-recent-activities", toolCall2.Name) // Already set
	})
	
	t.Run("filters out invalid tool calls", func(t *testing.T) {
		state := NewToolCallState()
		
		// Add a valid tool call
		state.toolCalls["valid"] = &responses.ResponseFunctionToolCall{
			CallID:    "valid",
			Name:      "get-athlete-profile",
			Arguments: "{}",
		}
		
		// Add an invalid tool call (missing CallID)
		state.toolCalls["invalid"] = &responses.ResponseFunctionToolCall{
			CallID:    "", // Invalid
			Name:      "get-athlete-profile",
			Arguments: "{}",
		}
		
		completed := service.GetCompletedToolCalls(state)
		
		assert.Len(t, completed, 1)
		assert.Equal(t, "valid", completed[0].CallID)
	})
}