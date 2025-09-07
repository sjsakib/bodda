package services

import (
	"log"

	"github.com/openai/openai-go/v2/responses"
)

// ContextManager interface defines context optimization functionality
type ContextManager interface {
	RedactPreviousStreamOutputs(messages []responses.ResponseInputItemUnionParam) []responses.ResponseInputItemUnionParam
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
// Simplified implementation for responses API
func (cm *contextManager) RedactPreviousStreamOutputs(messages []responses.ResponseInputItemUnionParam) []responses.ResponseInputItemUnionParam {
	if !cm.redactionEnabled {
		return messages
	}

	// For now, return messages as-is since redaction logic is complex
	// This can be enhanced later if needed
	log.Printf("Context redaction is enabled but simplified for responses API")
	return messages
}

// ShouldRedact determines if a tool call should be redacted based on its name
func (cm *contextManager) ShouldRedact(toolCallName string) bool {
	return cm.streamToolNames[toolCallName]
}