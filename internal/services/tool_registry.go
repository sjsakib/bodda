package services

import (
	"context"

	"bodda/internal/models"
)

// ToolRegistry defines the interface for managing and discovering tools
type ToolRegistry interface {
	// GetAvailableTools returns all available tools with their basic information
	GetAvailableTools() []models.ToolDefinition

	// GetToolSchema returns detailed schema information for a specific tool
	GetToolSchema(toolName string) (*models.ToolSchema, error)

	// ValidateToolCall validates that a tool call has the correct parameters
	ValidateToolCall(toolName string, parameters map[string]interface{}) error

	// IsToolAvailable checks if a tool with the given name exists
	IsToolAvailable(toolName string) bool
}

// ToolExecutor defines the interface for executing tools with timeout and streaming support
type ToolExecutor interface {
	// ExecuteTool executes a tool with the given parameters and context
	ExecuteTool(ctx context.Context, toolName string, parameters map[string]interface{}, msgCtx *MessageContext) (*models.ToolExecutionResult, error)

	// ExecuteToolWithOptions executes a tool with additional execution options including timeout and streaming
	ExecuteToolWithOptions(ctx context.Context, toolName string, parameters map[string]interface{}, msgCtx *MessageContext, options *models.ExecutionOptions) (*models.ToolExecutionResult, error)

	// CancelJob cancels an active tool execution job by ID
	CancelJob(jobID string) bool

	// GetActiveJobCount returns the number of currently active jobs
	GetActiveJobCount() int
}