package services

import (
	"testing"
	
	"github.com/openai/openai-go/v2/responses"
)

// TestCallIDLoggingHelpers tests the new call_id logging helper methods
func TestCallIDLoggingHelpers(t *testing.T) {
	// Create a mock AI service
	service := &aiService{}
	
	// Create a test tool call state
	state := NewToolCallState()
	
	// Add some test tool calls
	state.toolCalls["call_123"] = &responses.ResponseFunctionToolCall{
		CallID: "call_123",
		Name:   "get-athlete-profile",
	}
	state.toolCalls["call_456"] = &responses.ResponseFunctionToolCall{
		CallID: "call_456",
		Name:   "get-recent-activities",
	}
	
	// Mark one as completed
	state.completed["call_123"] = true
	
	// Test getAllCallIDs
	allCallIDs := service.getAllCallIDs(state)
	if len(allCallIDs) != 2 {
		t.Errorf("Expected 2 call IDs, got %d", len(allCallIDs))
	}
	
	// Test getCompletedCallIDs
	completedCallIDs := service.getCompletedCallIDs(state)
	if len(completedCallIDs) != 1 {
		t.Errorf("Expected 1 completed call ID, got %d", len(completedCallIDs))
	}
	
	// Test getPendingCallIDs
	pendingCallIDs := service.getPendingCallIDs(state)
	if len(pendingCallIDs) != 1 {
		t.Errorf("Expected 1 pending call ID, got %d", len(pendingCallIDs))
	}
	
	// Test getActiveCallIDs
	activeCallIDs := service.getActiveCallIDs(state)
	if len(activeCallIDs) != 2 {
		t.Errorf("Expected 2 active call IDs, got %d", len(activeCallIDs))
	}
}

// TestContentPreview tests the content preview helper method
func TestContentPreview(t *testing.T) {
	service := &aiService{}
	
	// Test short content
	shortContent := "Short content"
	preview := service.getContentPreview(shortContent)
	if preview != shortContent {
		t.Errorf("Expected short content to be unchanged, got %s", preview)
	}
	
	// Test long content
	longContent := "This is a very long content that should be truncated because it exceeds the maximum preview length limit"
	preview = service.getContentPreview(longContent)
	if len(preview) > 103 { // 100 chars + "..."
		t.Errorf("Expected content to be truncated, got length %d", len(preview))
	}
	if preview[len(preview)-3:] != "..." {
		t.Errorf("Expected content to end with '...', got %s", preview[len(preview)-3:])
	}
}