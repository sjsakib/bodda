package services

import (
	"fmt"
	"log"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// ContextManager interface defines context optimization functionality
type ContextManager interface {
	RedactPreviousStreamOutputs(messages []openai.ChatCompletionMessage) []openai.ChatCompletionMessage
	ShouldRedact(toolCallName string) bool
}

// contextManager implements the ContextManager interface
type contextManager struct {
	redactionEnabled bool
	streamToolNames  map[string]bool
}

// NewContextManager creates a new context manager with configuration
func NewContextManager(redactionEnabled bool) ContextManager {
	// Define which tool calls should be considered "stream tools" for redaction
	streamToolNames := map[string]bool{
		"get-activity-streams": true,
		// Add other stream-related tools here if needed in the future
	}

	return &contextManager{
		redactionEnabled: redactionEnabled,
		streamToolNames:  streamToolNames,
	}
}

// RedactPreviousStreamOutputs redacts content from previous stream tool outputs while preserving structure
// Only redacts tool results that are followed by non-tool call messages
func (cm *contextManager) RedactPreviousStreamOutputs(messages []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	if !cm.redactionEnabled {
		return messages
	}

	redactedMessages := make([]openai.ChatCompletionMessage, len(messages))
	copy(redactedMessages, messages)

	streamToolCallIDs := make(map[string]bool)
	redactionCount := 0

	// First pass: identify stream tool calls
	for _, message := range redactedMessages {
		if message.Role == openai.ChatMessageRoleAssistant && len(message.ToolCalls) > 0 {
			for _, toolCall := range message.ToolCalls {
				if cm.ShouldRedact(toolCall.Function.Name) {
					streamToolCallIDs[toolCall.ID] = true
				}
			}
		}
	}

	// Second pass: redact tool result messages for stream tools based on sequence analysis
	for i := range redactedMessages {
		message := &redactedMessages[i]
		
		// Check if this is a tool result message that corresponds to a stream tool call
		if message.Role == openai.ChatMessageRoleTool && streamToolCallIDs[message.ToolCallID] {
			// Apply conditional redaction based on subsequent message analysis
			if cm.hasSubsequentNonToolCallMessages(redactedMessages, i) {
				originalLength := len(message.Content)
				message.Content = cm.redactContent(message.Content, message.ToolCallID)
				redactionCount++
				
				// Enhanced logging for redaction decision
				cm.logRedactionDecision(message.ToolCallID, true, "followed by non-tool call messages", originalLength, len(message.Content), i, len(redactedMessages))
			} else {
				// Enhanced logging for preservation decision
				cm.logRedactionDecision(message.ToolCallID, false, cm.getPreservationReason(redactedMessages, i), len(message.Content), len(message.Content), i, len(redactedMessages))
			}
		}
	}

	if redactionCount > 0 {
		log.Printf("Context optimization: redacted %d previous stream tool outputs based on sequence analysis", redactionCount)
	}

	return redactedMessages
}

// ShouldRedact determines if a tool call should be redacted based on its name
func (cm *contextManager) ShouldRedact(toolCallName string) bool {
	return cm.streamToolNames[toolCallName]
}

// redactContent replaces the content of stream tool outputs with a summary placeholder
func (cm *contextManager) redactContent(originalContent, toolCallID string) string {
	// Simple redaction that preserves structure but removes detailed content
	lines := strings.Split(originalContent, "\n")
	
	// Count original lines and estimate content type
	originalLines := len(lines)
	contentType := cm.detectContentType(originalContent)
	
	// Create redacted placeholder that maintains context structure
	var redactedContent strings.Builder
	
	redactedContent.WriteString("📊 **[Previous Stream Analysis - Redacted for Context Optimization]**\n\n")
	redactedContent.WriteString(fmt.Sprintf("*This stream analysis contained %d lines of %s data and has been redacted to optimize context usage.*\n\n", 
		originalLines, contentType))
	
	// Preserve any processing mode information if present
	if strings.Contains(originalContent, "Processing Mode:") || strings.Contains(originalContent, "processing_mode") {
		redactedContent.WriteString("*Processing mode and options were provided in the original output.*\n\n")
	}
	
	// Preserve any pagination information if present
	if strings.Contains(originalContent, "Page") && strings.Contains(originalContent, "of") {
		redactedContent.WriteString("*Pagination information was included in the original output.*\n\n")
	}
	
	redactedContent.WriteString("*Use get-activity-streams tool with appropriate processing mode to access current stream data.*")
	
	return redactedContent.String()
}

// detectContentType analyzes content to determine what type of stream data it contains
func (cm *contextManager) detectContentType(content string) string {
	content = strings.ToLower(content)
	
	// Check for different types of stream content
	if strings.Contains(content, "derived features") || strings.Contains(content, "statistical") {
		return "derived features and statistics"
	}
	
	if strings.Contains(content, "ai-generated") || strings.Contains(content, "summary") {
		return "AI-generated summary"
	}
	
	if strings.Contains(content, "page") && strings.Contains(content, "of") {
		return "paginated stream data"
	}
	
	if strings.Contains(content, "processing mode") || strings.Contains(content, "options") {
		return "processing mode options"
	}
	
	if strings.Contains(content, "heart rate") || strings.Contains(content, "power") || strings.Contains(content, "stream") {
		return "raw stream data"
	}
	
	return "stream analysis"
}

// hasSubsequentNonToolCallMessages determines if a tool result has subsequent non-tool call messages
// This is the core logic for conditional redaction based on message sequence analysis
func (cm *contextManager) hasSubsequentNonToolCallMessages(messages []openai.ChatCompletionMessage, toolResultIndex int) bool {
	// Analyze all messages that come after the current tool result
	for i := toolResultIndex + 1; i < len(messages); i++ {
		message := messages[i]
		
		// If we find a non-tool call message, redaction should be applied
		if cm.isNonToolCallMessage(message) {
			return true
		}
	}
	
	// No subsequent non-tool call messages found
	return false
}

// isNonToolCallMessage determines if a message is a non-tool call message
// Non-tool call messages are assistant messages without tool calls, user messages, or system messages
func (cm *contextManager) isNonToolCallMessage(message openai.ChatCompletionMessage) bool {
	switch message.Role {
	case openai.ChatMessageRoleAssistant:
		// Assistant message without tool calls is considered a non-tool call message
		return len(message.ToolCalls) == 0
	case openai.ChatMessageRoleUser:
		// User messages are non-tool call messages
		return true
	case openai.ChatMessageRoleSystem:
		// System messages are non-tool call messages
		return true
	case openai.ChatMessageRoleTool:
		// Tool result messages are not non-tool call messages
		return false
	default:
		// Unknown role, treat as non-tool call message for safety
		return true
	}
}

// logRedactionDecision logs enhanced information about redaction decisions
// Includes tool call ID, decision rationale, and context without exposing sensitive content
func (cm *contextManager) logRedactionDecision(toolCallID string, wasRedacted bool, reason string, originalLength, finalLength, messageIndex, totalMessages int) {
	if wasRedacted {
		log.Printf("REDACTION_DECISION: tool_call_id=%s, action=REDACTED, reason=%s, original_length=%d, final_length=%d, position=%d/%d", 
			toolCallID, reason, originalLength, finalLength, messageIndex+1, totalMessages)
	} else {
		log.Printf("REDACTION_DECISION: tool_call_id=%s, action=PRESERVED, reason=%s, content_length=%d, position=%d/%d", 
			toolCallID, reason, originalLength, messageIndex+1, totalMessages)
	}
}

// getPreservationReason determines the specific reason why a tool result was preserved
// Provides detailed rationale for logging without exposing sensitive content
func (cm *contextManager) getPreservationReason(messages []openai.ChatCompletionMessage, toolResultIndex int) string {
	// Check if this is the last message
	if toolResultIndex == len(messages)-1 {
		return "final message in conversation"
	}
	
	// Analyze what follows to provide specific reason
	hasSubsequentToolCalls := false
	hasSubsequentToolResults := false
	subsequentMessageCount := len(messages) - toolResultIndex - 1
	
	for i := toolResultIndex + 1; i < len(messages); i++ {
		message := messages[i]
		
		switch message.Role {
		case openai.ChatMessageRoleAssistant:
			if len(message.ToolCalls) > 0 {
				hasSubsequentToolCalls = true
			}
		case openai.ChatMessageRoleTool:
			hasSubsequentToolResults = true
		}
	}
	
	if hasSubsequentToolCalls && hasSubsequentToolResults {
		return fmt.Sprintf("followed only by %d tool call(s) and result(s)", subsequentMessageCount)
	} else if hasSubsequentToolCalls {
		return fmt.Sprintf("followed only by %d tool call(s)", subsequentMessageCount)
	} else if hasSubsequentToolResults {
		return fmt.Sprintf("followed only by %d tool result(s)", subsequentMessageCount)
	} else if subsequentMessageCount == 0 {
		return "no subsequent messages"
	} else {
		return fmt.Sprintf("followed by %d message(s) with no non-tool call content", subsequentMessageCount)
	}
}