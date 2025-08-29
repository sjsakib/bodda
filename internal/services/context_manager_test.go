package services

import (
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestNewContextManager(t *testing.T) {
	cm := NewContextManager(true)
	if cm == nil {
		t.Fatal("NewContextManager returned nil")
	}

	// Test that it correctly identifies stream tools
	if !cm.ShouldRedact("get-activity-streams") {
		t.Error("Expected get-activity-streams to be identified as a stream tool")
	}

	// Test that it doesn't redact non-stream tools
	if cm.ShouldRedact("get-athlete-profile") {
		t.Error("Expected get-athlete-profile to not be identified as a stream tool")
	}
}

func TestShouldRedact(t *testing.T) {
	cm := NewContextManager(true)

	tests := []struct {
		toolName string
		expected bool
	}{
		{"get-activity-streams", true},
		{"get-athlete-profile", false},
		{"get-recent-activities", false},
		{"get-activity-details", false},
		{"update-athlete-logbook", false},
		{"unknown-tool", false},
	}

	for _, test := range tests {
		result := cm.ShouldRedact(test.toolName)
		if result != test.expected {
			t.Errorf("ShouldRedact(%s) = %v, expected %v", test.toolName, result, test.expected)
		}
	}
}

func TestRedactPreviousStreamOutputs_DisabledRedaction(t *testing.T) {
	cm := NewContextManager(false) // Redaction disabled

	originalMessages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "Show me my stream data",
		},
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
			Content:    "📊 Stream Data with lots of detailed information...",
			ToolCallID: "call_123",
		},
	}

	result := cm.RedactPreviousStreamOutputs(originalMessages)

	// Should return unchanged messages when redaction is disabled
	if len(result) != len(originalMessages) {
		t.Errorf("Expected %d messages, got %d", len(originalMessages), len(result))
	}

	// Content should be unchanged
	if result[2].Content != originalMessages[2].Content {
		t.Error("Content was modified when redaction was disabled")
	}
}

func TestRedactPreviousStreamOutputs_StreamToolRedaction(t *testing.T) {
	cm := NewContextManager(true) // Redaction enabled

	originalMessages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "Show me my stream data",
		},
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
			Content:    "📊 Stream Data\n\nHeart rate: 150-180 bpm\nPower: 200-300W\nDetailed analysis with many lines...",
			ToolCallID: "call_123",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "Now show me another activity",
		},
	}

	result := cm.RedactPreviousStreamOutputs(originalMessages)

	// Should have same number of messages
	if len(result) != len(originalMessages) {
		t.Errorf("Expected %d messages, got %d", len(originalMessages), len(result))
	}

	// User and assistant messages should be unchanged
	if result[0].Content != originalMessages[0].Content {
		t.Error("User message was modified")
	}

	if len(result[1].ToolCalls) != len(originalMessages[1].ToolCalls) {
		t.Error("Assistant tool calls were modified")
	}

	// Tool result should be redacted
	toolResult := result[2]
	if toolResult.Role != openai.ChatMessageRoleTool {
		t.Error("Tool result role was changed")
	}

	if toolResult.ToolCallID != "call_123" {
		t.Error("Tool call ID was changed")
	}

	// Content should be redacted but preserve structure
	if !strings.Contains(toolResult.Content, "[Previous Stream Analysis - Redacted") {
		t.Error("Content was not properly redacted")
	}

	if strings.Contains(toolResult.Content, "Heart rate: 150-180 bpm") {
		t.Error("Original detailed content was not removed")
	}

	// Last user message should be unchanged
	if result[3].Content != originalMessages[3].Content {
		t.Error("Subsequent user message was modified")
	}
}

func TestRedactPreviousStreamOutputs_NonStreamToolPreserved(t *testing.T) {
	cm := NewContextManager(true) // Redaction enabled

	originalMessages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleAssistant,
			ToolCalls: []openai.ToolCall{
				{
					ID:   "call_456",
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
			ToolCallID: "call_456",
		},
	}

	result := cm.RedactPreviousStreamOutputs(originalMessages)

	// Non-stream tool results should not be redacted
	if result[1].Content != originalMessages[1].Content {
		t.Error("Non-stream tool result was incorrectly redacted")
	}
}

func TestRedactPreviousStreamOutputs_MixedTools(t *testing.T) {
	cm := NewContextManager(true) // Redaction enabled

	originalMessages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleAssistant,
			ToolCalls: []openai.ToolCall{
				{
					ID:   "call_stream",
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
			Content:    "📊 Detailed stream data with heart rate and power...",
			ToolCallID: "call_stream",
		},
		{
			Role:       openai.ChatMessageRoleTool,
			Content:    "Athlete Profile: John Doe, cyclist...",
			ToolCallID: "call_profile",
		},
	}

	result := cm.RedactPreviousStreamOutputs(originalMessages)

	// Stream tool result should NOT be redacted because it's not followed by non-tool call messages
	streamResult := result[1]
	if streamResult.Content != originalMessages[1].Content {
		t.Error("Stream tool result was incorrectly redacted when not followed by non-tool call messages")
	}

	// Profile tool result should not be redacted
	profileResult := result[2]
	if profileResult.Content != originalMessages[2].Content {
		t.Error("Non-stream tool result was incorrectly redacted")
	}
}

func TestDetectContentType(t *testing.T) {
	cm := NewContextManager(true).(*contextManager)

	tests := []struct {
		content  string
		expected string
	}{
		{
			"📊 Derived Features Analysis\nStatistical summary of heart rate...",
			"derived features and statistics",
		},
		{
			"🤖 AI-Generated Summary\nThis workout shows...",
			"AI-generated summary",
		},
		{
			"📊 Stream Data (Page 1 of 5)\nTime: 0-3600 seconds...",
			"paginated stream data",
		},
		{
			"⚠️ Output too large\nProcessing Mode Options:\n- raw\n- derived",
			"processing mode options",
		},
		{
			"📊 Stream Data\nHeart Rate: 150 bpm\nPower: 250W",
			"raw stream data",
		},
		{
			"Some other content without specific markers",
			"stream analysis",
		},
	}

	for _, test := range tests {
		result := cm.detectContentType(test.content)
		if result != test.expected {
			t.Errorf("detectContentType(%q) = %q, expected %q", test.content, result, test.expected)
		}
	}
}

func TestRedactContent(t *testing.T) {
	cm := NewContextManager(true).(*contextManager)

	originalContent := `📊 Stream Data

**Total Data Points:** 3600

**Available Streams:**
- time
- heartrate
- watts

**Stream Data:**
- **Heart Rate:** 3600 data points (120-180 bpm)
- **Power:** 3600 data points (150-350 watts)

Detailed analysis continues for many more lines...`

	result := cm.redactContent(originalContent, "call_123")

	// Should contain redaction marker
	if !strings.Contains(result, "[Previous Stream Analysis - Redacted") {
		t.Error("Redacted content missing redaction marker")
	}

	// Should mention line count (the test content has 14 lines)
	if !strings.Contains(result, "14 lines") {
		t.Errorf("Redacted content should mention original line count, got: %s", result)
	}

	// Should not contain original detailed data
	if strings.Contains(result, "120-180 bpm") {
		t.Error("Redacted content still contains original detailed data")
	}

	// Should contain instruction for accessing current data
	if !strings.Contains(result, "get-activity-streams tool") {
		t.Error("Redacted content should contain instruction for accessing current data")
	}
}

func TestIsNonToolCallMessage(t *testing.T) {
	cm := NewContextManager(true).(*contextManager)

	tests := []struct {
		name     string
		message  openai.ChatCompletionMessage
		expected bool
	}{
		{
			name: "Assistant message with tool calls",
			message: openai.ChatCompletionMessage{
				Role: openai.ChatMessageRoleAssistant,
				ToolCalls: []openai.ToolCall{
					{ID: "call_123", Type: "function"},
				},
			},
			expected: false,
		},
		{
			name: "Assistant message without tool calls",
			message: openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "Here's my analysis...",
			},
			expected: true,
		},
		{
			name: "User message",
			message: openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: "Show me my data",
			},
			expected: true,
		},
		{
			name: "System message",
			message: openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a helpful assistant",
			},
			expected: true,
		},
		{
			name: "Tool result message",
			message: openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    "Tool result data",
				ToolCallID: "call_123",
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := cm.isNonToolCallMessage(test.message)
			if result != test.expected {
				t.Errorf("isNonToolCallMessage() = %v, expected %v", result, test.expected)
			}
		})
	}
}

func TestHasSubsequentNonToolCallMessages(t *testing.T) {
	cm := NewContextManager(true).(*contextManager)

	tests := []struct {
		name            string
		messages        []openai.ChatCompletionMessage
		toolResultIndex int
		expected        bool
	}{
		{
			name: "Tool result followed by assistant message without tool calls",
			messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleTool, Content: "Tool result", ToolCallID: "call_123"},
				{Role: openai.ChatMessageRoleAssistant, Content: "Based on the data..."},
			},
			toolResultIndex: 0,
			expected:        true,
		},
		{
			name: "Tool result followed by user message",
			messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleTool, Content: "Tool result", ToolCallID: "call_123"},
				{Role: openai.ChatMessageRoleUser, Content: "What does this mean?"},
			},
			toolResultIndex: 0,
			expected:        true,
		},
		{
			name: "Tool result followed only by other tool calls",
			messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleTool, Content: "Tool result", ToolCallID: "call_123"},
				{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{{ID: "call_456"}}},
				{Role: openai.ChatMessageRoleTool, Content: "Another tool result", ToolCallID: "call_456"},
			},
			toolResultIndex: 0,
			expected:        false,
		},
		{
			name: "Tool result is last message",
			messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleTool, Content: "Tool result", ToolCallID: "call_123"},
			},
			toolResultIndex: 0,
			expected:        false,
		},
		{
			name: "Tool result followed by tool calls then non-tool call message",
			messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleTool, Content: "Tool result", ToolCallID: "call_123"},
				{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{{ID: "call_456"}}},
				{Role: openai.ChatMessageRoleTool, Content: "Another tool result", ToolCallID: "call_456"},
				{Role: openai.ChatMessageRoleAssistant, Content: "Here's my analysis..."},
			},
			toolResultIndex: 0,
			expected:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := cm.hasSubsequentNonToolCallMessages(test.messages, test.toolResultIndex)
			if result != test.expected {
				t.Errorf("hasSubsequentNonToolCallMessages() = %v, expected %v", result, test.expected)
			}
		})
	}
}

// TestMessageSequenceAnalysis provides comprehensive tests for message sequence analysis
// covering all scenarios mentioned in requirements 1.1, 1.2, 1.3, 2.1, 2.2
func TestMessageSequenceAnalysis(t *testing.T) {
	cm := NewContextManager(true).(*contextManager)

	t.Run("DetectingNonToolCallMessagesFollowingToolResults", func(t *testing.T) {
		tests := []struct {
			name            string
			messages        []openai.ChatCompletionMessage
			toolResultIndex int
			expected        bool
			description     string
		}{
			{
				name: "Assistant explanation after tool result",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleTool, Content: "Stream data: HR 150bpm", ToolCallID: "call_123"},
					{Role: openai.ChatMessageRoleAssistant, Content: "Based on your heart rate data, I can see..."},
				},
				toolResultIndex: 0,
				expected:        true,
				description:     "Tool result followed by assistant explanation should be detected",
			},
			{
				name: "User question after tool result",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleTool, Content: "Activity streams loaded", ToolCallID: "call_456"},
					{Role: openai.ChatMessageRoleUser, Content: "What does this data tell us?"},
				},
				toolResultIndex: 0,
				expected:        true,
				description:     "Tool result followed by user question should be detected",
			},
			{
				name: "System message after tool result",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleTool, Content: "Processing complete", ToolCallID: "call_789"},
					{Role: openai.ChatMessageRoleSystem, Content: "Context updated"},
				},
				toolResultIndex: 0,
				expected:        true,
				description:     "Tool result followed by system message should be detected",
			},
			{
				name: "Multiple non-tool messages after tool result",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleTool, Content: "Data retrieved", ToolCallID: "call_abc"},
					{Role: openai.ChatMessageRoleAssistant, Content: "Let me analyze this..."},
					{Role: openai.ChatMessageRoleUser, Content: "Looks interesting"},
					{Role: openai.ChatMessageRoleAssistant, Content: "Indeed, the patterns show..."},
				},
				toolResultIndex: 0,
				expected:        true,
				description:     "Tool result followed by multiple non-tool messages should be detected",
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				result := cm.hasSubsequentNonToolCallMessages(test.messages, test.toolResultIndex)
				if result != test.expected {
					t.Errorf("%s: hasSubsequentNonToolCallMessages() = %v, expected %v", 
						test.description, result, test.expected)
				}
			})
		}
	})

	t.Run("ToolResultsFollowedOnlyByOtherToolCalls", func(t *testing.T) {
		tests := []struct {
			name            string
			messages        []openai.ChatCompletionMessage
			toolResultIndex int
			expected        bool
			description     string
		}{
			{
				name: "Single tool call chain",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleTool, Content: "First tool result", ToolCallID: "call_1"},
					{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{{ID: "call_2", Type: "function"}}},
					{Role: openai.ChatMessageRoleTool, Content: "Second tool result", ToolCallID: "call_2"},
				},
				toolResultIndex: 0,
				expected:        false,
				description:     "Tool result followed only by another tool call should not trigger redaction",
			},
			{
				name: "Multiple chained tool calls",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleTool, Content: "Stream data", ToolCallID: "call_stream"},
					{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{{ID: "call_profile", Type: "function"}}},
					{Role: openai.ChatMessageRoleTool, Content: "Profile data", ToolCallID: "call_profile"},
					{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{{ID: "call_activities", Type: "function"}}},
					{Role: openai.ChatMessageRoleTool, Content: "Activities data", ToolCallID: "call_activities"},
				},
				toolResultIndex: 0,
				expected:        false,
				description:     "Tool result in chain of tool calls should not trigger redaction",
			},
			{
				name: "Parallel tool calls",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{
						{ID: "call_1", Type: "function"},
						{ID: "call_2", Type: "function"},
					}},
					{Role: openai.ChatMessageRoleTool, Content: "First result", ToolCallID: "call_1"},
					{Role: openai.ChatMessageRoleTool, Content: "Second result", ToolCallID: "call_2"},
				},
				toolResultIndex: 1,
				expected:        false,
				description:     "Tool result from parallel calls should not trigger redaction when no non-tool messages follow",
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				result := cm.hasSubsequentNonToolCallMessages(test.messages, test.toolResultIndex)
				if result != test.expected {
					t.Errorf("%s: hasSubsequentNonToolCallMessages() = %v, expected %v", 
						test.description, result, test.expected)
				}
			})
		}
	})

	t.Run("EndOfConversationScenarios", func(t *testing.T) {
		tests := []struct {
			name            string
			messages        []openai.ChatCompletionMessage
			toolResultIndex int
			expected        bool
			description     string
		}{
			{
				name: "Tool result is final message",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: "Get my stream data"},
					{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{{ID: "call_final", Type: "function"}}},
					{Role: openai.ChatMessageRoleTool, Content: "Final stream data result", ToolCallID: "call_final"},
				},
				toolResultIndex: 2,
				expected:        false,
				description:     "Final tool result in conversation should not trigger redaction",
			},
			{
				name: "Multiple tool results ending conversation",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: "Get comprehensive data"},
					{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{
						{ID: "call_1", Type: "function"},
						{ID: "call_2", Type: "function"},
					}},
					{Role: openai.ChatMessageRoleTool, Content: "First data set", ToolCallID: "call_1"},
					{Role: openai.ChatMessageRoleTool, Content: "Second data set", ToolCallID: "call_2"},
				},
				toolResultIndex: 2,
				expected:        false,
				description:     "First of final tool results should not trigger redaction",
			},
			{
				name: "Last tool result in parallel execution",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{
						{ID: "call_a", Type: "function"},
						{ID: "call_b", Type: "function"},
					}},
					{Role: openai.ChatMessageRoleTool, Content: "Result A", ToolCallID: "call_a"},
					{Role: openai.ChatMessageRoleTool, Content: "Result B", ToolCallID: "call_b"},
				},
				toolResultIndex: 2,
				expected:        false,
				description:     "Last tool result in conversation should not trigger redaction",
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				result := cm.hasSubsequentNonToolCallMessages(test.messages, test.toolResultIndex)
				if result != test.expected {
					t.Errorf("%s: hasSubsequentNonToolCallMessages() = %v, expected %v", 
						test.description, result, test.expected)
				}
			})
		}
	})

	t.Run("MixedSequencesWithToolCallsAndNonToolCallMessages", func(t *testing.T) {
		tests := []struct {
			name            string
			messages        []openai.ChatCompletionMessage
			toolResultIndex int
			expected        bool
			description     string
		}{
			{
				name: "Tool result followed by tool calls then explanation",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleTool, Content: "Initial stream data", ToolCallID: "call_1"},
					{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{{ID: "call_2", Type: "function"}}},
					{Role: openai.ChatMessageRoleTool, Content: "Additional data", ToolCallID: "call_2"},
					{Role: openai.ChatMessageRoleAssistant, Content: "Now I can provide a complete analysis..."},
				},
				toolResultIndex: 0,
				expected:        true,
				description:     "Tool result eventually followed by explanation should trigger redaction",
			},
			{
				name: "Complex conversation flow",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: "Analyze my workout"},
					{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{{ID: "call_streams", Type: "function"}}},
					{Role: openai.ChatMessageRoleTool, Content: "Stream data loaded", ToolCallID: "call_streams"},
					{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{{ID: "call_profile", Type: "function"}}},
					{Role: openai.ChatMessageRoleTool, Content: "Profile loaded", ToolCallID: "call_profile"},
					{Role: openai.ChatMessageRoleUser, Content: "What do you see?"},
					{Role: openai.ChatMessageRoleAssistant, Content: "Based on your data..."},
				},
				toolResultIndex: 2,
				expected:        true,
				description:     "Tool result in complex flow with eventual user interaction should trigger redaction",
			},
			{
				name: "Tool result followed by mixed messages ending with tool call",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleTool, Content: "Data retrieved", ToolCallID: "call_1"},
					{Role: openai.ChatMessageRoleAssistant, Content: "Let me get more data..."},
					{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{{ID: "call_2", Type: "function"}}},
					{Role: openai.ChatMessageRoleTool, Content: "More data", ToolCallID: "call_2"},
				},
				toolResultIndex: 0,
				expected:        true,
				description:     "Tool result followed by explanation then more tool calls should trigger redaction",
			},
			{
				name: "Interleaved user messages and tool calls",
				messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleTool, Content: "Stream analysis", ToolCallID: "call_analysis"},
					{Role: openai.ChatMessageRoleUser, Content: "Interesting, can you get more details?"},
					{Role: openai.ChatMessageRoleAssistant, ToolCalls: []openai.ToolCall{{ID: "call_details", Type: "function"}}},
					{Role: openai.ChatMessageRoleTool, Content: "Detailed analysis", ToolCallID: "call_details"},
					{Role: openai.ChatMessageRoleAssistant, Content: "Here's what the detailed analysis shows..."},
				},
				toolResultIndex: 0,
				expected:        true,
				description:     "Tool result with interleaved user messages should trigger redaction",
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				result := cm.hasSubsequentNonToolCallMessages(test.messages, test.toolResultIndex)
				if result != test.expected {
					t.Errorf("%s: hasSubsequentNonToolCallMessages() = %v, expected %v", 
						test.description, result, test.expected)
				}
			})
		}
	})
}

// TestConditionalRedactionBehavior tests the full redaction logic integration
// covering requirements 1.1, 1.2, 1.3, 1.4, 4.1, 4.2, 4.3, 4.4
func TestConditionalRedactionBehavior(t *testing.T) {
	t.Run("ToolResultsFollowedByNonToolCallMessagesAreRedacted", func(t *testing.T) {
		cm := NewContextManager(true) // Redaction enabled

		// Test case 1: Stream tool result followed by assistant explanation
		t.Run("StreamToolFollowedByAssistantExplanation", func(t *testing.T) {
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
					Content: "Based on your stream data, I can see that your heart rate...",
				},
			}

			result := cm.RedactPreviousStreamOutputs(messages)

			// Stream tool result should be redacted because it's followed by assistant explanation
			toolResult := result[2]
			if !strings.Contains(toolResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Stream tool result should be redacted when followed by non-tool call message")
			}
			if strings.Contains(toolResult.Content, "Heart rate: 150-180 bpm") {
				t.Error("Original stream data should be redacted")
			}
		})

		// Test case 2: Stream tool result followed by user message
		t.Run("StreamToolFollowedByUserMessage", func(t *testing.T) {
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
					Content:    "📊 Stream Data\n\nPower: 250W average\nCadence: 90 rpm\nDetailed metrics...",
					ToolCallID: "call_streams",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "What does this data tell us about my performance?",
				},
			}

			result := cm.RedactPreviousStreamOutputs(messages)

			// Stream tool result should be redacted because it's followed by user message
			toolResult := result[1]
			if !strings.Contains(toolResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Stream tool result should be redacted when followed by user message")
			}
			if strings.Contains(toolResult.Content, "Power: 250W average") {
				t.Error("Original stream data should be redacted")
			}
		})

		// Test case 3: Stream tool result followed by system message
		t.Run("StreamToolFollowedBySystemMessage", func(t *testing.T) {
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
					Content:    "📊 Stream Data\n\nElevation: 500m gain\nSpeed: 25 km/h average\nComplete analysis...",
					ToolCallID: "call_streams",
				},
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "Context has been updated with new activity data",
				},
			}

			result := cm.RedactPreviousStreamOutputs(messages)

			// Stream tool result should be redacted because it's followed by system message
			toolResult := result[1]
			if !strings.Contains(toolResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Stream tool result should be redacted when followed by system message")
			}
			if strings.Contains(toolResult.Content, "Elevation: 500m gain") {
				t.Error("Original stream data should be redacted")
			}
		})

		// Test case 4: Multiple stream tool results with mixed following messages
		t.Run("MultipleStreamToolsWithMixedFollowingMessages", func(t *testing.T) {
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
					Content:    "📊 First Stream Data\n\nHeart rate analysis for activity 1...",
					ToolCallID: "call_streams_1",
				},
				{
					Role:       openai.ChatMessageRoleTool,
					Content:    "📊 Second Stream Data\n\nHeart rate analysis for activity 2...",
					ToolCallID: "call_streams_2",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "Comparing both activities, I can see...",
				},
			}

			result := cm.RedactPreviousStreamOutputs(messages)

			// Both stream tool results should be redacted because they're followed by assistant explanation
			firstResult := result[1]
			if !strings.Contains(firstResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("First stream tool result should be redacted when followed by non-tool call message")
			}

			secondResult := result[2]
			if !strings.Contains(secondResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Second stream tool result should be redacted when followed by non-tool call message")
			}
		})
	})

	t.Run("ToolResultsFollowedOnlyByToolCallsAreNotRedacted", func(t *testing.T) {
		cm := NewContextManager(true) // Redaction enabled

		// Test case 1: Stream tool followed by another tool call
		t.Run("StreamToolFollowedByAnotherToolCall", func(t *testing.T) {
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
					Content:    "📊 Stream Data\n\nHeart rate: 150-180 bpm\nPower: 200-300W",
					ToolCallID: "call_streams",
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
					Content:    "Athlete Profile: John Doe, cyclist",
					ToolCallID: "call_profile",
				},
			}

			result := cm.RedactPreviousStreamOutputs(messages)

			// Stream tool result should NOT be redacted because it's only followed by other tool calls
			streamResult := result[1]
			if streamResult.Content != messages[1].Content {
				t.Error("Stream tool result should not be redacted when followed only by tool calls")
			}
			if strings.Contains(streamResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Stream tool result should not be redacted when not followed by non-tool call messages")
			}
		})

		// Test case 2: Chain of multiple tool calls
		t.Run("ChainOfMultipleToolCalls", func(t *testing.T) {
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
					Content:    "📊 Stream Data\n\nDetailed activity streams loaded...",
					ToolCallID: "call_streams",
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
					Content:    "Profile data loaded",
					ToolCallID: "call_profile",
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
					Content:    "Recent activities loaded",
					ToolCallID: "call_activities",
				},
			}

			result := cm.RedactPreviousStreamOutputs(messages)

			// Stream tool result should NOT be redacted because it's in a chain of tool calls
			streamResult := result[1]
			if streamResult.Content != messages[1].Content {
				t.Error("Stream tool result should not be redacted when in chain of tool calls")
			}
			if strings.Contains(streamResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Stream tool result in tool call chain should preserve original content")
			}
		})

		// Test case 3: Parallel tool calls
		t.Run("ParallelToolCalls", func(t *testing.T) {
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
					Content:    "📊 Stream Data\n\nParallel stream analysis complete...",
					ToolCallID: "call_streams",
				},
				{
					Role:       openai.ChatMessageRoleTool,
					Content:    "Profile data retrieved",
					ToolCallID: "call_profile",
				},
			}

			result := cm.RedactPreviousStreamOutputs(messages)

			// Stream tool result should NOT be redacted because it's only followed by another tool result
			streamResult := result[1]
			if streamResult.Content != messages[1].Content {
				t.Error("Stream tool result should not be redacted in parallel tool call scenario")
			}
			if strings.Contains(streamResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Stream tool result from parallel calls should preserve original content")
			}
		})
	})

	t.Run("FinalToolResultsInConversationAreNotRedacted", func(t *testing.T) {
		cm := NewContextManager(true) // Redaction enabled

		// Test case 1: Single final stream tool result
		t.Run("SingleFinalStreamToolResult", func(t *testing.T) {
			messages := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Get my latest activity streams",
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
					Content:    "📊 Final Stream Data\n\nHeart rate: 140-170 bpm\nPower: 180-280W\nComplete analysis...",
					ToolCallID: "call_final_streams",
				},
			}

			result := cm.RedactPreviousStreamOutputs(messages)

			// Final stream tool result should NOT be redacted
			finalResult := result[2]
			if finalResult.Content != messages[2].Content {
				t.Error("Final stream tool result should not be redacted")
			}
			if strings.Contains(finalResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Final stream tool result should preserve original content")
			}
		})

		// Test case 2: Multiple final tool results
		t.Run("MultipleFinalToolResults", func(t *testing.T) {
			messages := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Get comprehensive data",
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
					Content:    "📊 Final Stream Data\n\nComprehensive stream analysis...",
					ToolCallID: "call_streams",
				},
				{
					Role:       openai.ChatMessageRoleTool,
					Content:    "Final Profile Data\n\nComplete athlete profile...",
					ToolCallID: "call_profile",
				},
			}

			result := cm.RedactPreviousStreamOutputs(messages)

			// Both final tool results should NOT be redacted
			streamResult := result[2]
			if streamResult.Content != messages[2].Content {
				t.Error("Final stream tool result should not be redacted")
			}
			if strings.Contains(streamResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Final stream tool result should preserve original content")
			}

			profileResult := result[3]
			if profileResult.Content != messages[3].Content {
				t.Error("Final profile tool result should not be redacted")
			}
		})

		// Test case 3: Stream tool result as last message after other interactions
		t.Run("StreamToolResultAsLastMessageAfterInteractions", func(t *testing.T) {
			messages := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Show me my workout data",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "I'll get your workout data for you.",
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
					Content:    "📊 Stream Data\n\nFinal workout analysis with all metrics...",
					ToolCallID: "call_final_streams",
				},
			}

			result := cm.RedactPreviousStreamOutputs(messages)

			// Final stream tool result should NOT be redacted even after previous interactions
			finalResult := result[3]
			if finalResult.Content != messages[3].Content {
				t.Error("Final stream tool result should not be redacted even after previous interactions")
			}
			if strings.Contains(finalResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Final stream tool result should preserve original content")
			}
		})
	})

	t.Run("EnvironmentVariableControlsGlobalRedactionBehavior", func(t *testing.T) {
		// Test case 1: Redaction disabled globally
		t.Run("RedactionDisabledGlobally", func(t *testing.T) {
			cm := NewContextManager(false) // Redaction disabled

			messages := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Show me my stream data",
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
					Content:    "📊 Stream Data\n\nHeart rate: 150-180 bpm\nPower: 200-300W\nDetailed analysis...",
					ToolCallID: "call_streams",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "Based on your stream data, I can analyze...",
				},
			}

			result := cm.RedactPreviousStreamOutputs(messages)

			// Stream tool result should NOT be redacted when redaction is globally disabled
			toolResult := result[2]
			if toolResult.Content != messages[2].Content {
				t.Error("Stream tool result should not be redacted when redaction is globally disabled")
			}
			if strings.Contains(toolResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("No redaction should occur when globally disabled")
			}
		})

		// Test case 2: Redaction enabled globally
		t.Run("RedactionEnabledGlobally", func(t *testing.T) {
			cm := NewContextManager(true) // Redaction enabled

			messages := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Show me my stream data",
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
					Content:    "📊 Stream Data\n\nHeart rate: 150-180 bpm\nPower: 200-300W\nDetailed analysis...",
					ToolCallID: "call_streams",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "Based on your stream data, I can analyze...",
				},
			}

			result := cm.RedactPreviousStreamOutputs(messages)

			// Stream tool result should be redacted when redaction is globally enabled and followed by non-tool call
			toolResult := result[2]
			if !strings.Contains(toolResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Stream tool result should be redacted when redaction is globally enabled and followed by non-tool call message")
			}
			if strings.Contains(toolResult.Content, "Heart rate: 150-180 bpm") {
				t.Error("Original stream data should be redacted when redaction is enabled")
			}
		})

		// Test case 3: Environment variable overrides sequence analysis
		t.Run("EnvironmentVariableOverridesSequenceAnalysis", func(t *testing.T) {
			cm := NewContextManager(false) // Redaction disabled

			// Even with a scenario that would normally trigger redaction, it should be disabled
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
					Content:    "📊 Stream Data\n\nDetailed stream analysis that would normally be redacted...",
					ToolCallID: "call_streams",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "What does this data show?",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "The data shows interesting patterns...",
				},
			}

			result := cm.RedactPreviousStreamOutputs(messages)

			// Stream tool result should NOT be redacted because environment variable disables redaction
			toolResult := result[1]
			if toolResult.Content != messages[1].Content {
				t.Error("Stream tool result should not be redacted when environment variable disables redaction")
			}
			if strings.Contains(toolResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Environment variable should override sequence analysis and disable redaction")
			}
		})

		// Test case 4: Verify redaction behavior with different environment settings
		t.Run("VerifyRedactionBehaviorWithDifferentEnvironmentSettings", func(t *testing.T) {
			// Same message sequence, different environment settings
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
					Content:    "📊 Stream Data\n\nTest data for environment variable behavior...",
					ToolCallID: "call_streams",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "Analysis of the stream data...",
				},
			}

			// Test with redaction enabled
			cmEnabled := NewContextManager(true)
			resultEnabled := cmEnabled.RedactPreviousStreamOutputs(messages)
			toolResultEnabled := resultEnabled[1]

			// Test with redaction disabled
			cmDisabled := NewContextManager(false)
			resultDisabled := cmDisabled.RedactPreviousStreamOutputs(messages)
			toolResultDisabled := resultDisabled[1]

			// Verify different behavior based on environment variable
			if !strings.Contains(toolResultEnabled.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Tool result should be redacted when environment variable enables redaction")
			}

			if toolResultDisabled.Content != messages[1].Content {
				t.Error("Tool result should not be modified when environment variable disables redaction")
			}

			if strings.Contains(toolResultDisabled.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Tool result should not be redacted when environment variable disables redaction")
			}
		})
	})

	t.Run("ComplexScenariosTesting", func(t *testing.T) {
		cm := NewContextManager(true) // Redaction enabled

		// Test case 1: Mixed sequence with eventual non-tool call message
		t.Run("MixedSequenceWithEventualNonToolCallMessage", func(t *testing.T) {
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
					Content:    "📊 Stream Data\n\nDetailed heart rate and power analysis...",
					ToolCallID: "call_streams",
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
					Content:    "Athlete Profile: Jane Smith",
					ToolCallID: "call_profile",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "What insights can you provide?",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "Based on your stream data and profile, I can see...",
				},
			}

			result := cm.RedactPreviousStreamOutputs(messages)

			// Stream tool result should be redacted because conversation eventually has non-tool call messages
			streamResult := result[1]
			if !strings.Contains(streamResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Stream tool result should be redacted when eventually followed by non-tool call messages")
			}

			// Non-stream tool result should not be redacted
			profileResult := result[3]
			if profileResult.Content != messages[3].Content {
				t.Error("Non-stream tool result should not be redacted")
			}
		})

		// Test case 2: Multiple stream tools with different outcomes
		t.Run("MultipleStreamToolsWithDifferentOutcomes", func(t *testing.T) {
			messages := []openai.ChatCompletionMessage{
				// First stream tool call - will be followed by non-tool call message
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
					Content:    "📊 First Stream Data\n\nActivity 1 analysis...",
					ToolCallID: "call_streams_1",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "Let me get more data...",
				},
				// Second stream tool call - will be final message
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
					Content:    "📊 Second Stream Data\n\nActivity 2 analysis...",
					ToolCallID: "call_streams_2",
				},
			}

			result := cm.RedactPreviousStreamOutputs(messages)

			// First stream tool result should be redacted (followed by non-tool call message)
			firstStreamResult := result[1]
			if !strings.Contains(firstStreamResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("First stream tool result should be redacted when followed by non-tool call message")
			}

			// Second stream tool result should NOT be redacted (final message)
			secondStreamResult := result[4]
			if secondStreamResult.Content != messages[4].Content {
				t.Error("Second stream tool result should not be redacted when it's the final message")
			}
			if strings.Contains(secondStreamResult.Content, "[Previous Stream Analysis - Redacted") {
				t.Error("Final stream tool result should preserve original content")
			}
		})
	})
}

// TestEdgeCasesInMessageSequenceAnalysis tests edge cases and boundary conditions
func TestEdgeCasesInMessageSequenceAnalysis(t *testing.T) {
	cm := NewContextManager(true).(*contextManager)

	t.Run("EmptyMessageSequence", func(t *testing.T) {
		messages := []openai.ChatCompletionMessage{}
		result := cm.hasSubsequentNonToolCallMessages(messages, 0)
		if result != false {
			t.Error("Empty message sequence should return false")
		}
	})

	t.Run("SingleMessageSequence", func(t *testing.T) {
		messages := []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleTool, Content: "Only message", ToolCallID: "call_only"},
		}
		result := cm.hasSubsequentNonToolCallMessages(messages, 0)
		if result != false {
			t.Error("Single message sequence should return false")
		}
	})

	t.Run("IndexOutOfBounds", func(t *testing.T) {
		messages := []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleTool, Content: "Tool result", ToolCallID: "call_123"},
		}
		// Test with index beyond array bounds
		result := cm.hasSubsequentNonToolCallMessages(messages, 5)
		if result != false {
			t.Error("Out of bounds index should return false")
		}
	})

	t.Run("AssistantMessageWithEmptyToolCalls", func(t *testing.T) {
		messages := []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleTool, Content: "Tool result", ToolCallID: "call_123"},
			{Role: openai.ChatMessageRoleAssistant, Content: "Analysis", ToolCalls: []openai.ToolCall{}},
		}
		result := cm.hasSubsequentNonToolCallMessages(messages, 0)
		if result != true {
			t.Error("Assistant message with empty tool calls should be considered non-tool call message")
		}
	})

	t.Run("AssistantMessageWithNilToolCalls", func(t *testing.T) {
		messages := []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleTool, Content: "Tool result", ToolCallID: "call_123"},
			{Role: openai.ChatMessageRoleAssistant, Content: "Analysis", ToolCalls: nil},
		}
		result := cm.hasSubsequentNonToolCallMessages(messages, 0)
		if result != true {
			t.Error("Assistant message with nil tool calls should be considered non-tool call message")
		}
	})

	t.Run("UnknownMessageRole", func(t *testing.T) {
		messages := []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleTool, Content: "Tool result", ToolCallID: "call_123"},
			{Role: "unknown_role", Content: "Unknown message"},
		}
		result := cm.hasSubsequentNonToolCallMessages(messages, 0)
		if result != true {
			t.Error("Unknown message role should be treated as non-tool call message for safety")
		}
	})
}