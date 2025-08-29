package services

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
)

// TestEnhancedLoggingForRedactionDecisions tests the enhanced logging functionality
// covering requirement 2.4: Log when redaction decisions are made based on subsequent message analysis
func TestEnhancedLoggingForRedactionDecisions(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer func() {
		log.SetOutput(os.Stderr) // Restore default log output
	}()

	t.Run("LogRedactionDecisionWithRationale", func(t *testing.T) {
		logBuffer.Reset()
		cm := NewContextManager(true) // Redaction enabled

		messages := []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleAssistant,
				ToolCalls: []openai.ToolCall{
					{
						ID:   "call_streams_123",
						Type: "function",
						Function: openai.FunctionCall{
							Name: "get-activity-streams",
						},
					},
				},
			},
			{
				Role:       openai.ChatMessageRoleTool,
				Content:    "📊 Stream Data\n\nHeart rate: 150-180 bpm\nPower: 200-300W\nDetailed analysis with sensitive data...",
				ToolCallID: "call_streams_123",
			},
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "Based on your stream data, I can see that your performance shows...",
			},
		}

		cm.RedactPreviousStreamOutputs(messages)

		logOutput := logBuffer.String()

		// Verify enhanced logging format
		if !strings.Contains(logOutput, "REDACTION_DECISION:") {
			t.Error("Log should contain REDACTION_DECISION marker")
		}

		// Verify tool call ID is logged
		if !strings.Contains(logOutput, "tool_call_id=call_streams_123") {
			t.Error("Log should contain tool call ID")
		}

		// Verify action is logged
		if !strings.Contains(logOutput, "action=REDACTED") {
			t.Error("Log should contain redaction action")
		}

		// Verify reason is logged
		if !strings.Contains(logOutput, "reason=followed by non-tool call messages") {
			t.Error("Log should contain decision rationale")
		}

		// Verify content lengths are logged
		if !strings.Contains(logOutput, "original_length=") {
			t.Error("Log should contain original content length")
		}

		if !strings.Contains(logOutput, "final_length=") {
			t.Error("Log should contain final content length")
		}

		// Verify position information is logged
		if !strings.Contains(logOutput, "position=2/3") {
			t.Error("Log should contain message position information")
		}

		// Verify sensitive content is NOT exposed in logs
		if strings.Contains(logOutput, "Heart rate: 150-180 bpm") {
			t.Error("Log should not expose sensitive tool call content")
		}

		if strings.Contains(logOutput, "Power: 200-300W") {
			t.Error("Log should not expose sensitive tool call content")
		}

		if strings.Contains(logOutput, "Detailed analysis with sensitive data") {
			t.Error("Log should not expose sensitive tool call content")
		}
	})

	t.Run("LogPreservationDecisionWithRationale", func(t *testing.T) {
		logBuffer.Reset()
		cm := NewContextManager(true) // Redaction enabled

		messages := []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleAssistant,
				ToolCalls: []openai.ToolCall{
					{
						ID:   "call_streams_456",
						Type: "function",
						Function: openai.FunctionCall{
							Name: "get-activity-streams",
						},
					},
				},
			},
			{
				Role:       openai.ChatMessageRoleTool,
				Content:    "📊 Stream Data\n\nSensitive performance metrics and analysis...",
				ToolCallID: "call_streams_456",
			},
			{
				Role: openai.ChatMessageRoleAssistant,
				ToolCalls: []openai.ToolCall{
					{
						ID:   "call_profile_789",
						Type: "function",
						Function: openai.FunctionCall{
							Name: "get-athlete-profile",
						},
					},
				},
			},
			{
				Role:       openai.ChatMessageRoleTool,
				Content:    "Athlete profile data...",
				ToolCallID: "call_profile_789",
			},
		}

		cm.RedactPreviousStreamOutputs(messages)

		logOutput := logBuffer.String()

		// Verify preservation decision is logged
		if !strings.Contains(logOutput, "action=PRESERVED") {
			t.Error("Log should contain preservation action")
		}

		// Verify tool call ID is logged
		if !strings.Contains(logOutput, "tool_call_id=call_streams_456") {
			t.Error("Log should contain tool call ID")
		}

		// Verify specific preservation reason
		if !strings.Contains(logOutput, "reason=followed only by") {
			t.Error("Log should contain specific preservation reason")
		}

		// Verify content length is logged for preserved content
		if !strings.Contains(logOutput, "content_length=") {
			t.Error("Log should contain content length for preserved messages")
		}

		// Verify position information is logged
		if !strings.Contains(logOutput, "position=2/4") {
			t.Error("Log should contain message position information")
		}

		// Verify sensitive content is NOT exposed in logs
		if strings.Contains(logOutput, "Sensitive performance metrics") {
			t.Error("Log should not expose sensitive tool call content")
		}
	})

	t.Run("LogFinalMessagePreservation", func(t *testing.T) {
		logBuffer.Reset()
		cm := NewContextManager(true) // Redaction enabled

		messages := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Get my stream data",
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
				Content:    "📊 Final Stream Data\n\nConfidential performance analysis...",
				ToolCallID: "call_final_streams",
			},
		}

		cm.RedactPreviousStreamOutputs(messages)

		logOutput := logBuffer.String()

		// Verify final message preservation is logged with specific reason
		if !strings.Contains(logOutput, "reason=final message in conversation") {
			t.Error("Log should contain specific reason for final message preservation")
		}

		// Verify action is preservation
		if !strings.Contains(logOutput, "action=PRESERVED") {
			t.Error("Log should show preservation action for final message")
		}

		// Verify position shows it's the last message
		if !strings.Contains(logOutput, "position=3/3") {
			t.Error("Log should show correct position for final message")
		}

		// Verify sensitive content is NOT exposed
		if strings.Contains(logOutput, "Confidential performance analysis") {
			t.Error("Log should not expose sensitive content from final message")
		}
	})

	t.Run("LogMultipleDecisionsInSameConversation", func(t *testing.T) {
		logBuffer.Reset()
		cm := NewContextManager(true) // Redaction enabled

		messages := []openai.ChatCompletionMessage{
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
				Content:    "First stream data with sensitive metrics...",
				ToolCallID: "call_streams_1",
			},
			{
				Role:       openai.ChatMessageRoleTool,
				Content:    "Second stream data with confidential analysis...",
				ToolCallID: "call_streams_2",
			},
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "Comparing both activities, I can provide insights...",
			},
		}

		cm.RedactPreviousStreamOutputs(messages)

		logOutput := logBuffer.String()

		// Verify both tool calls are logged
		if !strings.Contains(logOutput, "tool_call_id=call_streams_1") {
			t.Error("Log should contain first tool call ID")
		}

		if !strings.Contains(logOutput, "tool_call_id=call_streams_2") {
			t.Error("Log should contain second tool call ID")
		}

		// Count the number of redaction decisions logged
		decisionCount := strings.Count(logOutput, "REDACTION_DECISION:")
		if decisionCount != 2 {
			t.Errorf("Expected 2 redaction decisions to be logged, got %d", decisionCount)
		}

		// Verify both are redacted with same reason
		redactedCount := strings.Count(logOutput, "action=REDACTED")
		if redactedCount != 2 {
			t.Errorf("Expected 2 redaction actions to be logged, got %d", redactedCount)
		}

		// Verify sensitive content from both messages is NOT exposed
		if strings.Contains(logOutput, "sensitive metrics") {
			t.Error("Log should not expose sensitive content from first message")
		}

		if strings.Contains(logOutput, "confidential analysis") {
			t.Error("Log should not expose sensitive content from second message")
		}
	})

	t.Run("LogPreservationReasonsVariety", func(t *testing.T) {
		logBuffer.Reset()
		cm := NewContextManager(true).(*contextManager)

		// Test different preservation reasons
		tests := []struct {
			name             string
			messages         []openai.ChatCompletionMessage
			expectedReason   string
			toolResultIndex  int
		}{
			{
				name: "Final message",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleTool, Content: "Data", ToolCallID: "call_1"},
				},
				expectedReason:  "final message in conversation",
				toolResultIndex: 0,
			},
			{
				name: "Followed by tool calls only",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleTool, Content: "Data", ToolCallID: "call_1"},
					{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{{ID: "call_2"}}},
					{Role: openai.ChatMessageRoleTool, Content: "More data", ToolCallID: "call_2"},
				},
				expectedReason:  "followed only by 2 tool call(s) and result(s)",
				toolResultIndex: 0,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				reason := cm.getPreservationReason(test.messages, test.toolResultIndex)
				if reason != test.expectedReason {
					t.Errorf("Expected reason %q, got %q", test.expectedReason, reason)
				}
			})
		}
	})
}

// TestLoggingDoesNotExposeContent verifies that logging never exposes sensitive tool call content
func TestLoggingDoesNotExposeContent(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer func() {
		log.SetOutput(os.Stderr) // Restore default log output
	}()

	cm := NewContextManager(true) // Redaction enabled

	// Create messages with various types of sensitive content
	sensitiveContents := []string{
		"Personal health data: HR 180bpm, medical condition XYZ",
		"API keys and tokens: sk-1234567890abcdef",
		"Private location data: GPS coordinates 40.7128, -74.0060",
		"Financial information: Account balance $50,000",
		"Personal identifiers: SSN 123-45-6789, Phone +1-555-0123",
	}

	for i, sensitiveContent := range sensitiveContents {
		logBuffer.Reset()
		toolCallID := fmt.Sprintf("call_sensitive_%d", i)

		messages := []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleAssistant,
				ToolCalls: []openai.ToolCall{
					{
						ID:   toolCallID,
						Type: "function",
						Function: openai.FunctionCall{
							Name: "get-activity-streams",
						},
					},
				},
			},
			{
				Role:       openai.ChatMessageRoleTool,
				Content:    sensitiveContent,
				ToolCallID: toolCallID,
			},
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "Analysis complete",
			},
		}

		cm.RedactPreviousStreamOutputs(messages)

		logOutput := logBuffer.String()

		// Verify the tool call ID is logged (this is safe to log)
		if !strings.Contains(logOutput, fmt.Sprintf("tool_call_id=%s", toolCallID)) {
			t.Errorf("Log should contain tool call ID %s", toolCallID)
		}

		// Verify sensitive content is NOT exposed in any form
		sensitiveWords := []string{
			"HR 180bpm", "medical condition", "sk-1234567890abcdef", "API keys",
			"GPS coordinates", "40.7128", "-74.0060", "Account balance", "$50,000",
			"SSN 123-45-6789", "Phone +1-555-0123", "Personal identifiers",
		}

		for _, sensitiveWord := range sensitiveWords {
			if strings.Contains(logOutput, sensitiveWord) {
				t.Errorf("Log should not expose sensitive content: %s", sensitiveWord)
			}
		}

		// Verify that only safe metadata is logged
		expectedSafeElements := []string{
			"REDACTION_DECISION:",
			"action=REDACTED",
			"reason=followed by non-tool call messages",
			"original_length=",
			"final_length=",
			"position=",
		}

		for _, safeElement := range expectedSafeElements {
			if !strings.Contains(logOutput, safeElement) {
				t.Errorf("Log should contain safe metadata element: %s", safeElement)
			}
		}
	}
}