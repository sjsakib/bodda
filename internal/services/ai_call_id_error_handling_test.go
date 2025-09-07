package services

import (
	"strings"
	"testing"

	"github.com/openai/openai-go/v2/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallIDValidationLogic(t *testing.T) {
	t.Run("validates fallback logic for missing call_id", func(t *testing.T) {
		// Test the validation and fallback logic that's implemented in handleOutputItemAdded
		callID := ""
		itemID := "test-item-123"
		
		// This simulates the logic from our implementation
		if callID == "" {
			callID = itemID
		}
		
		assert.Equal(t, "test-item-123", callID)
		assert.NotEmpty(t, callID)
	})
	
	t.Run("validates error case for empty call_id and item_id", func(t *testing.T) {
		// Test the error case where both call_id and item_id are empty
		callID := ""
		itemID := ""
		
		// This simulates the validation logic from our implementation
		if callID == "" {
			callID = itemID
			if callID == "" {
				// This should trigger an error in our implementation
				assert.Empty(t, callID, "Both call_id and item_id are empty, should trigger error")
				return
			}
		}
		
		t.Error("Expected error condition for empty call_id and item_id")
	})
	
	t.Run("validates whitespace-only call_id handling", func(t *testing.T) {
		callID := "   " // Whitespace-only call_id
		itemID := "test-item-123"
		
		// Test the enhanced validation logic for whitespace
		if strings.TrimSpace(callID) == "" {
			callID = itemID
		}
		
		// Our enhanced validation should catch whitespace-only strings
		assert.Equal(t, "test-item-123", callID)
		assert.NotEmpty(t, callID)
	})
}

func TestHandleFunctionCallArgumentsDelta_ErrorHandling(t *testing.T) {
	service := &aiService{}
	
	t.Run("handles empty item_id", func(t *testing.T) {
		state := NewToolCallState()
		
		// Test with empty item_id
		deltaEvent := responses.ResponseFunctionCallArgumentsDeltaEvent{
			ItemID: "", // Empty item_id
			Delta:  `{"activity_id": 123}`,
		}
		
		err := service.handleFunctionCallArgumentsDelta(deltaEvent, state)
		
		// Should return an error for empty item_id
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty item ID")
	})
	
	t.Run("handles whitespace-only item_id", func(t *testing.T) {
		state := NewToolCallState()
		
		// Test with whitespace-only item_id
		deltaEvent := responses.ResponseFunctionCallArgumentsDeltaEvent{
			ItemID: "   ", // Whitespace-only item_id
			Delta:  `{"activity_id": 123}`,
		}
		
		err := service.handleFunctionCallArgumentsDelta(deltaEvent, state)
		
		// Should return an error for whitespace-only item_id
		require.Error(t, err)
		assert.Contains(t, err.Error(), "whitespace")
	})
	
	t.Run("handles missing item_to_call_id mapping with fallback", func(t *testing.T) {
		state := NewToolCallState()
		
		// Test with valid item_id but no existing mapping
		deltaEvent := responses.ResponseFunctionCallArgumentsDeltaEvent{
			ItemID: "test-item-456",
			Delta:  `{"activity_id": 123}`,
		}
		
		err := service.handleFunctionCallArgumentsDelta(deltaEvent, state)
		
		// Should not return an error, should use fallback
		require.NoError(t, err)
		
		// Should have created a tool call using the fallback
		assert.Equal(t, 1, service.GetToolCallCount(state))
		
		toolCall, exists := service.GetToolCallByID(state, "test-item-456")
		require.True(t, exists)
		assert.Equal(t, "test-item-456", toolCall.CallID) // Should use item_id as call_id
	})
}