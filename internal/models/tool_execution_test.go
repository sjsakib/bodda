package models

import (
	"testing"
	"time"
)

func TestToolDefinition(t *testing.T) {
	tool := ToolDefinition{
		Name:        "test-tool",
		Description: "A test tool",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param1": map[string]interface{}{
					"type":        "string",
					"description": "Test parameter",
				},
			},
			"required": []string{"param1"},
		},
		Examples: []ToolExample{
			{
				Description: "Test example",
				Request:     map[string]interface{}{"param1": "test"},
				Response:    "Test response",
			},
		},
	}

	if tool.Name != "test-tool" {
		t.Errorf("Expected tool name 'test-tool', got '%s'", tool.Name)
	}

	if len(tool.Examples) != 1 {
		t.Errorf("Expected 1 example, got %d", len(tool.Examples))
	}
}

func TestToolExecutionResult(t *testing.T) {
	now := time.Now()
	result := ToolExecutionResult{
		ToolName:  "test-tool",
		Success:   true,
		Data:      "test data",
		Duration:  100,
		Timestamp: now,
	}

	if result.ToolName != "test-tool" {
		t.Errorf("Expected tool name 'test-tool', got '%s'", result.ToolName)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Duration != 100 {
		t.Errorf("Expected duration 100, got %d", result.Duration)
	}
}

func TestToolExecutionRequest(t *testing.T) {
	request := ToolExecutionRequest{
		ToolName: "test-tool",
		Parameters: map[string]interface{}{
			"param1": "value1",
		},
		Options: &ExecutionOptions{
			Timeout:        30,
			Streaming:      false,
			BufferedOutput: true,
		},
	}

	if request.ToolName != "test-tool" {
		t.Errorf("Expected tool name 'test-tool', got '%s'", request.ToolName)
	}

	if request.Options.Timeout != 30 {
		t.Errorf("Expected timeout 30, got %d", request.Options.Timeout)
	}
}

func TestToolExecutionResponse(t *testing.T) {
	now := time.Now()
	response := ToolExecutionResponse{
		Status: "success",
		Result: &ToolExecutionResult{
			ToolName:  "test-tool",
			Success:   true,
			Data:      "test data",
			Duration:  100,
			Timestamp: now,
		},
		Metadata: &ResponseMetadata{
			RequestID: "req-123",
			Timestamp: now,
			Duration:  100,
		},
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}

	if response.Result.ToolName != "test-tool" {
		t.Errorf("Expected tool name 'test-tool', got '%s'", response.Result.ToolName)
	}

	if response.Metadata.RequestID != "req-123" {
		t.Errorf("Expected request ID 'req-123', got '%s'", response.Metadata.RequestID)
	}
}

func TestErrorResponse(t *testing.T) {
	now := time.Now()
	errorResp := ErrorResponse{
		Error: struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Details string `json:"details,omitempty"`
		}{
			Code:    "TOOL_NOT_FOUND",
			Message: "Tool not found",
			Details: "The specified tool does not exist",
		},
		RequestID: "req-123",
		Timestamp: now,
	}

	if errorResp.Error.Code != "TOOL_NOT_FOUND" {
		t.Errorf("Expected error code 'TOOL_NOT_FOUND', got '%s'", errorResp.Error.Code)
	}

	if errorResp.RequestID != "req-123" {
		t.Errorf("Expected request ID 'req-123', got '%s'", errorResp.RequestID)
	}
}