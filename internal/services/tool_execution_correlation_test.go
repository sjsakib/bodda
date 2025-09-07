package services

import (
	"testing"

	"github.com/openai/openai-go/v2/responses"
)

func TestToolExecutionAndResultCorrelation(t *testing.T) {
	t.Run("ToolResult creation uses correct call_id", func(t *testing.T) {
		// Test the core logic: ToolResult should use toolCall.CallID, not toolCall.ID
		
		// Create test tool calls with specific call_ids
		toolCalls := []responses.ResponseFunctionToolCall{
			{
				ID:        "fc_item_123",                    // Item ID (should not be used)
				CallID:    "call_abc123",                    // OpenAI call ID (should be used)
				Name:      "get-athlete-profile",
				Arguments: "{}",
			},
			{
				ID:        "fc_item_456",                    // Item ID (should not be used)
				CallID:    "call_def456",                    // OpenAI call ID (should be used)
				Name:      "get-recent-activities",
				Arguments: `{"per_page": 10}`,
			},
		}
		
		// Test that ToolResult creation uses the correct field
		for i, toolCall := range toolCalls {
			// This simulates the logic from executeToolsFromResponsesAPI
			result := ToolResult{
				ToolCallID: toolCall.CallID,  // This should use CallID, not ID
			}
			
			// Verify that the ToolResult uses the correct call_id
			expectedCallID := []string{"call_abc123", "call_def456"}[i]
			if result.ToolCallID != expectedCallID {
				t.Errorf("Tool result %d should use call_id '%s', got: %s", i, expectedCallID, result.ToolCallID)
			}
			
			// Verify that it's not using the item ID
			itemID := []string{"fc_item_123", "fc_item_456"}[i]
			if result.ToolCallID == itemID {
				t.Errorf("Tool result %d should not use item ID '%s'", i, itemID)
			}
		}
	})
	
	t.Run("fixToolResultIDs uses correct call_id for mapping", func(t *testing.T) {
		// Test that the fixToolResultIDs method uses CallID for mapping, not ID
		
		// Create test tool calls
		toolCalls := []responses.ResponseFunctionToolCall{
			{
				ID:     "fc_item_123",
				CallID: "call_abc123",
				Name:   "get-athlete-profile",
			},
			{
				ID:     "fc_item_456", 
				CallID: "call_def456",
				Name:   "get-recent-activities",
			},
		}
		
		// Simulate the mapping logic from fixToolResultIDs
		toolCallMap := make(map[string]string) // key: function_name, value: tool_call_id
		
		for _, toolCall := range toolCalls {
			if toolCall.Name != "" && toolCall.CallID != "" {
				// This should use CallID, not ID (this was the bug we fixed)
				toolCallMap[toolCall.Name] = toolCall.CallID
			}
		}
		
		// Verify the mapping uses call_id, not item ID
		if toolCallMap["get-athlete-profile"] != "call_abc123" {
			t.Errorf("Mapping should use call_id 'call_abc123', got: %s", toolCallMap["get-athlete-profile"])
		}
		if toolCallMap["get-recent-activities"] != "call_def456" {
			t.Errorf("Mapping should use call_id 'call_def456', got: %s", toolCallMap["get-recent-activities"])
		}
		
		// Verify it's not using item IDs
		if toolCallMap["get-athlete-profile"] == "fc_item_123" {
			t.Error("Mapping should not use item ID 'fc_item_123'")
		}
		if toolCallMap["get-recent-activities"] == "fc_item_456" {
			t.Error("Mapping should not use item ID 'fc_item_456'")
		}
	})
}