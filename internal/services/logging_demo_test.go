package services

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
)

// TestEnhancedLoggingDemo demonstrates the enhanced logging functionality
// This test shows the actual log output format for documentation purposes
func TestEnhancedLoggingDemo(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer func() {
		log.SetOutput(os.Stderr) // Restore default log output
	}()

	cm := NewContextManager(true) // Redaction enabled

	// Scenario 1: Tool result followed by explanation (should be redacted)
	fmt.Println("=== Scenario 1: Tool result followed by explanation (REDACTED) ===")
	logBuffer.Reset()

	messages1 := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleAssistant,
			ToolCalls: []openai.ToolCall{
				{
					ID:   "call_demo_redacted",
					Type: "function",
					Function: openai.FunctionCall{
						Name: "get-activity-streams",
					},
				},
			},
		},
		{
			Role:       openai.ChatMessageRoleTool,
			Content:    "📊 Stream Data\n\nHeart rate: 150-180 bpm\nPower: 200-300W\nDetailed performance analysis...",
			ToolCallID: "call_demo_redacted",
		},
		{
			Role:    openai.ChatMessageRoleAssistant,
			Content: "Based on your stream data, I can see excellent performance patterns...",
		},
	}

	cm.RedactPreviousStreamOutputs(messages1)
	fmt.Println("Log output:", logBuffer.String())

	// Scenario 2: Tool result as final message (should be preserved)
	fmt.Println("\n=== Scenario 2: Tool result as final message (PRESERVED) ===")
	logBuffer.Reset()

	messages2 := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "Get my latest activity streams",
		},
		{
			Role: openai.ChatMessageRoleAssistant,
			ToolCalls: []openai.ToolCall{
				{
					ID:   "call_demo_preserved",
					Type: "function",
					Function: openai.FunctionCall{
						Name: "get-activity-streams",
					},
				},
			},
		},
		{
			Role:       openai.ChatMessageRoleTool,
			Content:    "📊 Final Stream Analysis\n\nComplete performance metrics and insights...",
			ToolCallID: "call_demo_preserved",
		},
	}

	cm.RedactPreviousStreamOutputs(messages2)
	fmt.Println("Log output:", logBuffer.String())

	// Scenario 3: Tool result followed by more tool calls (should be preserved)
	fmt.Println("\n=== Scenario 3: Tool result followed by more tool calls (PRESERVED) ===")
	logBuffer.Reset()

	messages3 := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleAssistant,
			ToolCalls: []openai.ToolCall{
				{
					ID:   "call_demo_chained",
					Type: "function",
					Function: openai.FunctionCall{
						Name: "get-activity-streams",
					},
				},
			},
		},
		{
			Role:       openai.ChatMessageRoleTool,
			Content:    "📊 Stream Data for further analysis\n\nInitial metrics loaded...",
			ToolCallID: "call_demo_chained",
		},
		{
			Role: openai.ChatMessageRoleAssistant,
			ToolCalls: []openai.ToolCall{
				{
					ID:   "call_demo_followup",
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
			ToolCallID: "call_demo_followup",
		},
	}

	cm.RedactPreviousStreamOutputs(messages3)
	fmt.Println("Log output:", logBuffer.String())

	// This test always passes - it's for demonstration purposes
	t.Log("Enhanced logging demonstration completed")
}