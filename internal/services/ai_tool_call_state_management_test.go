package services

import (
	"testing"

	"github.com/openai/openai-go/v2/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToolCallCreationWithCallID tests tool call creation with correct call_id as primary key
// Requirements: 2.1, 2.2
func TestToolCallCreationWithCallID(t *testing.T) {
	service := &aiService{}

	t.Run("creates tool call with call_id as primary key", func(t *testing.T) {
		state := NewToolCallState()
		
		// Create tool call with call_id as primary identifier
		toolCall := &responses.ResponseFunctionToolCall{
			ID:        "fc_item123",
			CallID:    "call_primary456",
			Name:      "get-athlete-profile",
			Arguments: "{}",
		}
		
		// Add tool call using call_id as primary key
		state.toolCalls[toolCall.CallID] = toolCall
		
		// Verify tool call was created with correct call_id as key
		retrieved, exists := service.GetToolCallByID(state, "call_primary456")
		require.True(t, exists, "Tool call should exist with call_id as key")
		assert.Equal(t, "call_primary456", retrieved.CallID)
		assert.Equal(t, "fc_item123", retrieved.ID)
		assert.Equal(t, "get-athlete-profile", retrieved.Name)
		
		// Verify cannot retrieve by item_id (should not be primary key)
		_, exists = service.GetToolCallByID(state, "fc_item123")
		assert.False(t, exists, "Tool call should not be retrievable by item_id")
	})

	t.Run("creates multiple tool calls with unique call_ids", func(t *testing.T) {
		state := NewToolCallState()
		
		toolCalls := []struct {
			itemID   string
			callID   string
			funcName string
		}{
			{"fc_1", "call_abc123", "get-athlete-profile"},
			{"fc_2", "call_def456", "get-recent-activities"},
			{"fc_3", "call_ghi789", "get-activity-details"},
		}
		
		// Create multiple tool calls using call_id as primary key
		for _, tc := range toolCalls {
			toolCall := &responses.ResponseFunctionToolCall{
				ID:        tc.itemID,
				CallID:    tc.callID,
				Name:      tc.funcName,
				Arguments: "{}",
			}
			state.toolCalls[tc.callID] = toolCall
		}
		
		// Verify all tool calls exist with call_id as primary key
		assert.Equal(t, 3, service.GetToolCallCount(state))
		
		for _, tc := range toolCalls {
			retrieved, exists := service.GetToolCallByID(state, tc.callID)
			require.True(t, exists, "Tool call with call_id %s should exist", tc.callID)
			assert.Equal(t, tc.callID, retrieved.CallID)
			assert.Equal(t, tc.itemID, retrieved.ID)
			assert.Equal(t, tc.funcName, retrieved.Name)
		}
	})

	t.Run("handles tool call creation with same item_id but different call_ids", func(t *testing.T) {
		state := NewToolCallState()
		
		// Create two tool calls with same item_id but different call_ids
		// This tests that call_id is truly the primary key
		toolCall1 := &responses.ResponseFunctionToolCall{
			ID:        "fc_same",
			CallID:    "call_first",
			Name:      "get-athlete-profile",
			Arguments: "{}",
		}
		
		toolCall2 := &responses.ResponseFunctionToolCall{
			ID:        "fc_same", // Same item_id
			CallID:    "call_second", // Different call_id
			Name:      "get-recent-activities",
			Arguments: `{"per_page": 30}`,
		}
		
		// Add both using call_id as primary key
		state.toolCalls[toolCall1.CallID] = toolCall1
		state.toolCalls[toolCall2.CallID] = toolCall2
		
		// Verify both exist as separate entries
		assert.Equal(t, 2, service.GetToolCallCount(state))
		
		retrieved1, exists1 := service.GetToolCallByID(state, "call_first")
		require.True(t, exists1)
		assert.Equal(t, "call_first", retrieved1.CallID)
		assert.Equal(t, "get-athlete-profile", retrieved1.Name)
		
		retrieved2, exists2 := service.GetToolCallByID(state, "call_second")
		require.True(t, exists2)
		assert.Equal(t, "call_second", retrieved2.CallID)
		assert.Equal(t, "get-recent-activities", retrieved2.Name)
	})

	t.Run("validates tool call creation with empty call_id fails", func(t *testing.T) {
		state := NewToolCallState()
		
		// Create tool call with empty call_id
		toolCall := &responses.ResponseFunctionToolCall{
			ID:        "fc_empty",
			CallID:    "", // Empty call_id
			Name:      "get-athlete-profile",
			Arguments: "{}",
		}
		
		// Validation should fail for empty call_id
		isValid := service.validateToolCall(toolCall)
		assert.False(t, isValid, "Tool call with empty call_id should not be valid")
		
		// If added to state, it should not be retrievable
		state.toolCalls[""] = toolCall
		_, exists := service.GetToolCallByID(state, "")
		assert.True(t, exists) // It exists in the map but is invalid
		
		// But validation should catch this
		assert.False(t, service.validateToolCall(toolCall))
	})
}

// TestToolCallCompletionTracking tests tool call completion tracking using call_id
// Requirements: 2.2, 2.3
func TestToolCallCompletionTracking(t *testing.T) {
	service := &aiService{}

	t.Run("tracks completion using call_id", func(t *testing.T) {
		state := NewToolCallState()
		
		// Create tool calls
		callIDs := []string{"call_track1", "call_track2", "call_track3"}
		for _, callID := range callIDs {
			toolCall := &responses.ResponseFunctionToolCall{
				ID:        "fc_" + callID,
				CallID:    callID,
				Name:      "get-athlete-profile",
				Arguments: "{}",
			}
			state.toolCalls[callID] = toolCall
		}
		
		// Initially no tool calls should be completed
		assert.Equal(t, 0, service.GetCompletedToolCallCount(state))
		assert.True(t, service.HasPendingToolCalls(state))
		
		// Mark first tool call as completed using call_id
		service.MarkToolCallCompleted(state, "call_track1")
		
		// Verify completion tracking
		assert.Equal(t, 1, service.GetCompletedToolCallCount(state))
		assert.True(t, service.IsToolCallComplete(state, "call_track1"))
		assert.False(t, service.IsToolCallComplete(state, "call_track2"))
		assert.True(t, service.HasPendingToolCalls(state))
		
		// Mark remaining tool calls as completed
		service.MarkToolCallCompleted(state, "call_track2")
		service.MarkToolCallCompleted(state, "call_track3")
		
		// Verify all are completed
		assert.Equal(t, 3, service.GetCompletedToolCallCount(state))
		assert.False(t, service.HasPendingToolCalls(state))
		
		for _, callID := range callIDs {
			assert.True(t, service.IsToolCallComplete(state, callID))
		}
	})

	t.Run("completion tracking with mixed states", func(t *testing.T) {
		state := NewToolCallState()
		
		// Create tool calls with different completion states
		completedCallIDs := []string{"call_done1", "call_done2"}
		pendingCallIDs := []string{"call_pending1", "call_pending2", "call_pending3"}
		
		// Add completed tool calls
		for _, callID := range completedCallIDs {
			toolCall := &responses.ResponseFunctionToolCall{
				ID:        "fc_" + callID,
				CallID:    callID,
				Name:      "get-athlete-profile",
				Arguments: "{}",
			}
			state.toolCalls[callID] = toolCall
			state.completed[callID] = true
		}
		
		// Add pending tool calls
		for _, callID := range pendingCallIDs {
			toolCall := &responses.ResponseFunctionToolCall{
				ID:        "fc_" + callID,
				CallID:    callID,
				Name:      "get-recent-activities",
				Arguments: "{}",
			}
			state.toolCalls[callID] = toolCall
			// Don't mark as completed
		}
		
		// Verify counts
		assert.Equal(t, 5, service.GetToolCallCount(state))
		assert.Equal(t, 2, service.GetCompletedToolCallCount(state))
		assert.True(t, service.HasPendingToolCalls(state))
		
		// Verify individual completion states
		for _, callID := range completedCallIDs {
			assert.True(t, service.IsToolCallComplete(state, callID))
		}
		for _, callID := range pendingCallIDs {
			assert.False(t, service.IsToolCallComplete(state, callID))
		}
		
		// Verify helper methods return correct call_ids
		completedIDs := service.getCompletedCallIDs(state)
		assert.Len(t, completedIDs, 2)
		for _, callID := range completedCallIDs {
			assert.Contains(t, completedIDs, callID)
		}
		
		pendingIDs := service.getPendingCallIDs(state)
		assert.Len(t, pendingIDs, 3)
		for _, callID := range pendingCallIDs {
			assert.Contains(t, pendingIDs, callID)
		}
	})

	t.Run("completion tracking edge cases", func(t *testing.T) {
		state := NewToolCallState()
		
		// Test marking non-existent tool call as completed
		service.MarkToolCallCompleted(state, "call_nonexistent")
		assert.False(t, service.IsToolCallComplete(state, "call_nonexistent"))
		assert.Equal(t, 0, service.GetCompletedToolCallCount(state))
		
		// Test with empty state
		assert.False(t, service.HasPendingToolCalls(state))
		assert.Equal(t, 0, service.GetToolCallCount(state))
		assert.Equal(t, 0, service.GetCompletedToolCallCount(state))
		
		// Add tool call and test completion
		toolCall := &responses.ResponseFunctionToolCall{
			ID:        "fc_edge",
			CallID:    "call_edge",
			Name:      "get-athlete-profile",
			Arguments: "{}",
		}
		state.toolCalls["call_edge"] = toolCall
		
		assert.True(t, service.HasPendingToolCalls(state))
		assert.Equal(t, 1, service.GetToolCallCount(state))
		assert.Equal(t, 0, service.GetCompletedToolCallCount(state))
		
		// Mark as completed
		service.MarkToolCallCompleted(state, "call_edge")
		assert.False(t, service.HasPendingToolCalls(state))
		assert.Equal(t, 1, service.GetCompletedToolCallCount(state))
	})
}

// TestToolCallIdentifierValidation tests validation of tool call identifiers
// Requirements: 2.1, 2.2, 2.3
func TestToolCallIdentifierValidation(t *testing.T) {
	service := &aiService{}

	t.Run("validates tool call identifiers", func(t *testing.T) {
		testCases := []struct {
			name        string
			toolCall    responses.ResponseFunctionToolCall
			expectValid bool
			reason      string
		}{
			{
				name: "valid tool call with call_id",
				toolCall: responses.ResponseFunctionToolCall{
					ID:        "fc_valid",
					CallID:    "call_valid123",
					Name:      "get-athlete-profile",
					Arguments: "{}",
				},
				expectValid: true,
				reason:      "has all required fields",
			},
			{
				name: "invalid tool call with empty call_id",
				toolCall: responses.ResponseFunctionToolCall{
					ID:        "fc_invalid",
					CallID:    "", // Empty call_id
					Name:      "get-athlete-profile",
					Arguments: "{}",
				},
				expectValid: false,
				reason:      "missing call_id",
			},
			{
				name: "invalid tool call with empty function name",
				toolCall: responses.ResponseFunctionToolCall{
					ID:        "fc_noname",
					CallID:    "call_noname",
					Name:      "", // Empty function name
					Arguments: "{}",
				},
				expectValid: false,
				reason:      "missing function name",
			},
			{
				name: "invalid tool call with unknown function name",
				toolCall: responses.ResponseFunctionToolCall{
					ID:        "fc_unknown",
					CallID:    "call_unknown",
					Name:      "unknown-function", // Unknown function
					Arguments: "{}",
				},
				expectValid: false,
				reason:      "unknown function name",
			},
			{
				name: "invalid tool call with malformed JSON arguments",
				toolCall: responses.ResponseFunctionToolCall{
					ID:        "fc_badjson",
					CallID:    "call_badjson",
					Name:      "get-athlete-profile",
					Arguments: `{"invalid": json}`, // Malformed JSON
				},
				expectValid: false,
				reason:      "malformed JSON arguments",
			},
			{
				name: "valid tool call with empty arguments",
				toolCall: responses.ResponseFunctionToolCall{
					ID:        "fc_empty_args",
					CallID:    "call_empty_args",
					Name:      "get-athlete-profile",
					Arguments: "", // Empty arguments (should default to {})
				},
				expectValid: true,
				reason:      "empty arguments are valid",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				isValid := service.validateToolCall(&tc.toolCall)
				assert.Equal(t, tc.expectValid, isValid, "Validation result should match expectation: %s", tc.reason)
			})
		}
	})

	t.Run("validates known function names", func(t *testing.T) {
		knownFunctions := []string{
			"get-athlete-profile",
			"get-recent-activities",
			"get-activity-details",
			"get-activity-streams",
			"update-athlete-logbook",
		}

		for _, funcName := range knownFunctions {
			toolCall := &responses.ResponseFunctionToolCall{
				ID:        "fc_known",
				CallID:    "call_known",
				Name:      funcName,
				Arguments: "{}",
			}
			
			isValid := service.validateToolCall(toolCall)
			assert.True(t, isValid, "Known function %s should be valid", funcName)
		}
	})

	t.Run("rejects unknown function names", func(t *testing.T) {
		unknownFunctions := []string{
			"unknown-function",
			"invalid-tool",
			"get-weather",
			"send-email",
			"",
		}

		for _, funcName := range unknownFunctions {
			toolCall := &responses.ResponseFunctionToolCall{
				ID:        "fc_unknown",
				CallID:    "call_unknown",
				Name:      funcName,
				Arguments: "{}",
			}
			
			isValid := service.validateToolCall(toolCall)
			assert.False(t, isValid, "Unknown function %s should be invalid", funcName)
		}
	})

	t.Run("validates JSON arguments", func(t *testing.T) {
		validJSONArgs := []string{
			"{}",
			`{"per_page": 30}`,
			`{"activity_id": 123}`,
			`{"content": "test logbook content"}`,
			`{"stream_types": ["time", "distance"], "resolution": "high"}`,
		}

		for _, args := range validJSONArgs {
			toolCall := &responses.ResponseFunctionToolCall{
				ID:        "fc_json",
				CallID:    "call_json",
				Name:      "get-athlete-profile",
				Arguments: args,
			}
			
			isValid := service.validateToolCall(toolCall)
			assert.True(t, isValid, "Valid JSON arguments should pass validation: %s", args)
		}

		invalidJSONArgs := []string{
			`{"invalid": json}`,
			`{missing_quotes: "value"}`,
			`{"unclosed": "string}`,
			`{trailing_comma: "value",}`,
			`not_json_at_all`,
		}

		for _, args := range invalidJSONArgs {
			toolCall := &responses.ResponseFunctionToolCall{
				ID:        "fc_badjson",
				CallID:    "call_badjson",
				Name:      "get-athlete-profile",
				Arguments: args,
			}
			
			isValid := service.validateToolCall(toolCall)
			assert.False(t, isValid, "Invalid JSON arguments should fail validation: %s", args)
		}
	})
}

// TestMultipleConcurrentToolCalls tests state management with multiple concurrent tool calls
// Requirements: 2.1, 2.2, 2.3
func TestMultipleConcurrentToolCalls(t *testing.T) {
	service := &aiService{}

	t.Run("manages multiple concurrent tool calls", func(t *testing.T) {
		state := NewToolCallState()
		
		// Create multiple concurrent tool calls
		concurrentCalls := []struct {
			itemID   string
			callID   string
			funcName string
			args     string
		}{
			{"fc_1", "call_profile", "get-athlete-profile", "{}"},
			{"fc_2", "call_activities", "get-recent-activities", `{"per_page": 30}`},
			{"fc_3", "call_details", "get-activity-details", `{"activity_id": 123}`},
			{"fc_4", "call_streams", "get-activity-streams", `{"activity_id": 456}`},
			{"fc_5", "call_logbook", "update-athlete-logbook", `{"content": "test"}`},
		}
		
		// Add all tool calls simultaneously (simulating concurrent processing)
		for _, tc := range concurrentCalls {
			toolCall := &responses.ResponseFunctionToolCall{
				ID:        tc.itemID,
				CallID:    tc.callID,
				Name:      tc.funcName,
				Arguments: tc.args,
			}
			state.toolCalls[tc.callID] = toolCall
			state.itemToCallID[tc.itemID] = tc.callID
		}
		
		// Verify all tool calls exist
		assert.Equal(t, 5, service.GetToolCallCount(state))
		assert.True(t, service.HasPendingToolCalls(state))
		assert.Equal(t, 0, service.GetCompletedToolCallCount(state))
		
		// Verify each tool call individually
		for _, tc := range concurrentCalls {
			retrieved, exists := service.GetToolCallByID(state, tc.callID)
			require.True(t, exists, "Tool call %s should exist", tc.callID)
			assert.Equal(t, tc.callID, retrieved.CallID)
			assert.Equal(t, tc.itemID, retrieved.ID)
			assert.Equal(t, tc.funcName, retrieved.Name)
			assert.Equal(t, tc.args, retrieved.Arguments)
			
			// Verify item_id to call_id mapping
			assert.Equal(t, tc.callID, state.itemToCallID[tc.itemID])
		}
		
		// Verify helper methods work correctly
		activeCallIDs := service.getActiveCallIDs(state)
		assert.Len(t, activeCallIDs, 5)
		for _, tc := range concurrentCalls {
			assert.Contains(t, activeCallIDs, tc.callID)
		}
		
		allCallIDs := service.getAllCallIDs(state)
		assert.Len(t, allCallIDs, 5)
		
		pendingCallIDs := service.getPendingCallIDs(state)
		assert.Len(t, pendingCallIDs, 5)
		
		completedCallIDs := service.getCompletedCallIDs(state)
		assert.Len(t, completedCallIDs, 0)
	})

	t.Run("handles partial completion of concurrent tool calls", func(t *testing.T) {
		state := NewToolCallState()
		
		// Create concurrent tool calls
		callIDs := []string{"call_c1", "call_c2", "call_c3", "call_c4", "call_c5"}
		for _, callID := range callIDs {
			toolCall := &responses.ResponseFunctionToolCall{
				ID:        "fc_" + callID,
				CallID:    callID,
				Name:      "get-athlete-profile",
				Arguments: "{}",
			}
			state.toolCalls[callID] = toolCall
		}
		
		// Complete some tool calls in random order
		completedOrder := []string{"call_c3", "call_c1", "call_c5"}
		for _, callID := range completedOrder {
			service.MarkToolCallCompleted(state, callID)
		}
		
		// Verify partial completion state
		assert.Equal(t, 5, service.GetToolCallCount(state))
		assert.Equal(t, 3, service.GetCompletedToolCallCount(state))
		assert.True(t, service.HasPendingToolCalls(state))
		
		// Verify individual completion states
		for _, callID := range completedOrder {
			assert.True(t, service.IsToolCallComplete(state, callID))
		}
		
		pendingCallIDs := []string{"call_c2", "call_c4"}
		for _, callID := range pendingCallIDs {
			assert.False(t, service.IsToolCallComplete(state, callID))
		}
		
		// Verify helper methods
		completedIDs := service.getCompletedCallIDs(state)
		assert.Len(t, completedIDs, 3)
		for _, callID := range completedOrder {
			assert.Contains(t, completedIDs, callID)
		}
		
		pendingIDs := service.getPendingCallIDs(state)
		assert.Len(t, pendingIDs, 2)
		for _, callID := range pendingCallIDs {
			assert.Contains(t, pendingIDs, callID)
		}
	})

	t.Run("handles concurrent tool calls with same function but different arguments", func(t *testing.T) {
		state := NewToolCallState()
		
		// Create multiple calls to the same function with different arguments
		sameFunctionCalls := []struct {
			callID string
			args   string
		}{
			{"call_details1", `{"activity_id": 123}`},
			{"call_details2", `{"activity_id": 456}`},
			{"call_details3", `{"activity_id": 789}`},
		}
		
		for i, tc := range sameFunctionCalls {
			toolCall := &responses.ResponseFunctionToolCall{
				ID:        "fc_details" + string(rune('1'+i)),
				CallID:    tc.callID,
				Name:      "get-activity-details", // Same function name
				Arguments: tc.args,
			}
			state.toolCalls[tc.callID] = toolCall
		}
		
		// Verify all tool calls exist as separate entities
		assert.Equal(t, 3, service.GetToolCallCount(state))
		
		for _, tc := range sameFunctionCalls {
			retrieved, exists := service.GetToolCallByID(state, tc.callID)
			require.True(t, exists)
			assert.Equal(t, tc.callID, retrieved.CallID)
			assert.Equal(t, "get-activity-details", retrieved.Name)
			assert.Equal(t, tc.args, retrieved.Arguments)
		}
		
		// Complete them in different order
		service.MarkToolCallCompleted(state, "call_details2")
		service.MarkToolCallCompleted(state, "call_details1")
		
		assert.Equal(t, 2, service.GetCompletedToolCallCount(state))
		assert.True(t, service.HasPendingToolCalls(state))
		assert.True(t, service.IsToolCallComplete(state, "call_details1"))
		assert.True(t, service.IsToolCallComplete(state, "call_details2"))
		assert.False(t, service.IsToolCallComplete(state, "call_details3"))
	})

	t.Run("handles large number of concurrent tool calls", func(t *testing.T) {
		state := NewToolCallState()
		
		// Create a large number of concurrent tool calls
		const numCalls = 100
		callIDs := make([]string, numCalls)
		
		for i := 0; i < numCalls; i++ {
			callID := "call_large_" + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10)) + string(rune('0'+(i/100)%10))
			callIDs[i] = callID
			
			toolCall := &responses.ResponseFunctionToolCall{
				ID:        "fc_" + callID,
				CallID:    callID,
				Name:      "get-athlete-profile",
				Arguments: "{}",
			}
			state.toolCalls[callID] = toolCall
		}
		
		// Verify all were created
		assert.Equal(t, numCalls, service.GetToolCallCount(state))
		assert.True(t, service.HasPendingToolCalls(state))
		assert.Equal(t, 0, service.GetCompletedToolCallCount(state))
		
		// Complete half of them
		for i := 0; i < numCalls/2; i++ {
			service.MarkToolCallCompleted(state, callIDs[i])
		}
		
		// Verify partial completion
		assert.Equal(t, numCalls/2, service.GetCompletedToolCallCount(state))
		assert.True(t, service.HasPendingToolCalls(state))
		
		// Verify helper methods work with large numbers
		allCallIDs := service.getAllCallIDs(state)
		assert.Len(t, allCallIDs, numCalls)
		
		completedCallIDs := service.getCompletedCallIDs(state)
		assert.Len(t, completedCallIDs, numCalls/2)
		
		pendingCallIDs := service.getPendingCallIDs(state)
		assert.Len(t, pendingCallIDs, numCalls/2)
	})
}

// TestToolCallStateEdgeCases tests edge cases in tool call state management
func TestToolCallStateEdgeCases(t *testing.T) {
	service := &aiService{}

	t.Run("handles empty state operations", func(t *testing.T) {
		state := NewToolCallState()
		
		// Test all operations on empty state
		assert.Equal(t, 0, service.GetToolCallCount(state))
		assert.Equal(t, 0, service.GetCompletedToolCallCount(state))
		assert.False(t, service.HasPendingToolCalls(state))
		
		// Test retrieval operations
		_, exists := service.GetToolCallByID(state, "nonexistent")
		assert.False(t, exists)
		
		assert.False(t, service.IsToolCallComplete(state, "nonexistent"))
		
		// Test helper methods
		assert.Empty(t, service.getActiveCallIDs(state))
		assert.Empty(t, service.getAllCallIDs(state))
		assert.Empty(t, service.getCompletedCallIDs(state))
		assert.Empty(t, service.getPendingCallIDs(state))
		
		// Test completion operations
		service.MarkToolCallCompleted(state, "nonexistent") // Should not panic
		assert.Equal(t, 0, service.GetCompletedToolCallCount(state))
	})

	t.Run("handles duplicate call_id scenarios", func(t *testing.T) {
		state := NewToolCallState()
		
		// Create first tool call
		toolCall1 := &responses.ResponseFunctionToolCall{
			ID:        "fc_first",
			CallID:    "call_duplicate",
			Name:      "get-athlete-profile",
			Arguments: "{}",
		}
		state.toolCalls["call_duplicate"] = toolCall1
		
		// Create second tool call with same call_id (simulating overwrite)
		toolCall2 := &responses.ResponseFunctionToolCall{
			ID:        "fc_second",
			CallID:    "call_duplicate", // Same call_id
			Name:      "get-recent-activities",
			Arguments: `{"per_page": 30}`,
		}
		state.toolCalls["call_duplicate"] = toolCall2 // Overwrites first
		
		// Verify only the second tool call exists
		assert.Equal(t, 1, service.GetToolCallCount(state))
		
		retrieved, exists := service.GetToolCallByID(state, "call_duplicate")
		require.True(t, exists)
		assert.Equal(t, "fc_second", retrieved.ID) // Should be the second one
		assert.Equal(t, "get-recent-activities", retrieved.Name)
	})

	t.Run("handles item_id to call_id mapping edge cases", func(t *testing.T) {
		state := NewToolCallState()
		
		// Test mapping with same item_id to different call_ids
		state.itemToCallID["fc_shared"] = "call_first"
		
		// Overwrite mapping (simulating delta event processing)
		state.itemToCallID["fc_shared"] = "call_second"
		
		// Verify only the latest mapping exists
		assert.Equal(t, "call_second", state.itemToCallID["fc_shared"])
		
		// Test retrieval of non-existent mapping
		assert.Equal(t, "", state.itemToCallID["fc_nonexistent"])
	})

	t.Run("handles completion state inconsistencies", func(t *testing.T) {
		state := NewToolCallState()
		
		// Mark tool call as completed before it exists in toolCalls
		state.completed["call_orphaned"] = true
		
		// Verify completion tracking
		assert.True(t, service.IsToolCallComplete(state, "call_orphaned"))
		assert.Equal(t, 1, service.GetCompletedToolCallCount(state))
		assert.Equal(t, 0, service.GetToolCallCount(state)) // No actual tool calls
		
		// Add the actual tool call later
		toolCall := &responses.ResponseFunctionToolCall{
			ID:        "fc_orphaned",
			CallID:    "call_orphaned",
			Name:      "get-athlete-profile",
			Arguments: "{}",
		}
		state.toolCalls["call_orphaned"] = toolCall
		
		// Verify state is now consistent
		assert.Equal(t, 1, service.GetToolCallCount(state))
		assert.Equal(t, 1, service.GetCompletedToolCallCount(state))
		assert.False(t, service.HasPendingToolCalls(state))
	})

	t.Run("handles special characters in call_ids", func(t *testing.T) {
		state := NewToolCallState()
		
		specialCallIDs := []string{
			"call_with-dashes",
			"call_with_underscores",
			"call.with.dots",
			"call123with456numbers",
			"call_with_CAPS",
		}
		
		// Create tool calls with special characters in call_ids
		for i, callID := range specialCallIDs {
			toolCall := &responses.ResponseFunctionToolCall{
				ID:        "fc_special_" + string(rune('1'+i)),
				CallID:    callID,
				Name:      "get-athlete-profile",
				Arguments: "{}",
			}
			state.toolCalls[callID] = toolCall
		}
		
		// Verify all tool calls exist and are retrievable
		assert.Equal(t, len(specialCallIDs), service.GetToolCallCount(state))
		
		for _, callID := range specialCallIDs {
			retrieved, exists := service.GetToolCallByID(state, callID)
			require.True(t, exists, "Tool call with special call_id should exist: %s", callID)
			assert.Equal(t, callID, retrieved.CallID)
		}
		
		// Test completion tracking with special characters
		service.MarkToolCallCompleted(state, "call_with-dashes")
		service.MarkToolCallCompleted(state, "call.with.dots")
		
		assert.Equal(t, 2, service.GetCompletedToolCallCount(state))
		assert.True(t, service.IsToolCallComplete(state, "call_with-dashes"))
		assert.True(t, service.IsToolCallComplete(state, "call.with.dots"))
	})
}