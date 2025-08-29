package models

import (
	"time"
)

// ToolDefinition represents a tool that can be executed
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Examples    []ToolExample          `json:"examples,omitempty"`
}

// ToolSchema provides detailed schema information for a tool
type ToolSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Required    []string               `json:"required"`
	Optional    []string               `json:"optional"`
	Examples    []ToolExample          `json:"examples"`
}

// ToolExample provides usage examples for a tool
type ToolExample struct {
	Description string                 `json:"description"`
	Request     map[string]interface{} `json:"request"`
	Response    interface{}            `json:"response"`
}

// ToolExecutionResult represents the result of a tool execution
type ToolExecutionResult struct {
	ToolName   string      `json:"tool_name"`
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
	Duration   int64       `json:"duration_ms"`
	Timestamp  time.Time   `json:"timestamp"`
}

// ToolExecutionRequest represents a request to execute a tool
type ToolExecutionRequest struct {
	ToolName   string                 `json:"tool_name" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
	Options    *ExecutionOptions      `json:"options,omitempty"`
}

// ExecutionOptions provides options for tool execution
type ExecutionOptions struct {
	Timeout        int  `json:"timeout_seconds,omitempty"`
	Streaming      bool `json:"streaming,omitempty"`
	BufferedOutput bool `json:"buffered_output,omitempty"`
}

// ToolExecutionResponse represents the response from a tool execution
type ToolExecutionResponse struct {
	Status   string               `json:"status"`
	Result   *ToolExecutionResult `json:"result,omitempty"`
	Error    *ErrorDetails        `json:"error,omitempty"`
	Metadata *ResponseMetadata    `json:"metadata,omitempty"`
}

// ErrorDetails provides detailed error information
type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ResponseMetadata provides metadata about the response
type ResponseMetadata struct {
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
	Duration  int64     `json:"duration_ms"`
}

// ToolListResponse represents the response for listing available tools
type ToolListResponse struct {
	Tools []ToolDefinition `json:"tools"`
	Count int              `json:"count"`
}

// ToolSchemaResponse represents the response for getting a tool's schema
type ToolSchemaResponse struct {
	Schema *ToolSchema `json:"schema"`
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details string `json:"details,omitempty"`
	} `json:"error"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}