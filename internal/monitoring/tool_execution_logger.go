package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ToolExecutionLog represents a log entry for tool execution
type ToolExecutionLog struct {
	RequestID    string                 `json:"request_id"`
	Timestamp    time.Time              `json:"timestamp"`
	UserID       string                 `json:"user_id,omitempty"`
	ToolName     string                 `json:"tool_name"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Duration     int64                  `json:"duration_ms"`
	Success      bool                   `json:"success"`
	ErrorDetails string                 `json:"error_details,omitempty"`
	StackTrace   string                 `json:"stack_trace,omitempty"`
	ResponseSize int64                  `json:"response_size_bytes,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ToolExecutionLogger handles logging for tool executions
type ToolExecutionLogger struct {
	logger                *Logger
	enableParameterLogging bool
	sensitiveParams       map[string]bool
}

// NewToolExecutionLogger creates a new tool execution logger
func NewToolExecutionLogger(logger *Logger, enableParameterLogging bool) *ToolExecutionLogger {
	// Define sensitive parameters that should be redacted
	sensitiveParams := map[string]bool{
		"password":     true,
		"token":        true,
		"api_key":      true,
		"secret":       true,
		"private_key":  true,
		"access_token": true,
		"refresh_token": true,
		"auth":         true,
		"authorization": true,
	}

	return &ToolExecutionLogger{
		logger:                logger,
		enableParameterLogging: enableParameterLogging,
		sensitiveParams:       sensitiveParams,
	}
}

// LogToolExecution logs a successful tool execution
func (tel *ToolExecutionLogger) LogToolExecution(ctx context.Context, toolName string, parameters map[string]interface{}, duration time.Duration, responseSize int64, userID string) {
	requestID := tel.getRequestID(ctx)
	
	logEntry := ToolExecutionLog{
		RequestID:    requestID,
		Timestamp:    time.Now(),
		UserID:       userID,
		ToolName:     toolName,
		Duration:     duration.Milliseconds(),
		Success:      true,
		ResponseSize: responseSize,
	}

	// Add parameters if logging is enabled
	if tel.enableParameterLogging {
		logEntry.Parameters = tel.sanitizeParameters(parameters)
	}

	tel.logger.WithContext(ctx).Info("Tool execution completed",
		"request_id", logEntry.RequestID,
		"tool_name", logEntry.ToolName,
		"user_id", logEntry.UserID,
		"duration_ms", logEntry.Duration,
		"response_size_bytes", logEntry.ResponseSize,
		"success", logEntry.Success,
	)

	// Log detailed entry as JSON for structured logging
	if logData, err := json.Marshal(logEntry); err == nil {
		tel.logger.WithContext(ctx).Debug("Tool execution details", "log_data", string(logData))
	}
}

// LogToolExecutionError logs a failed tool execution with detailed error information
func (tel *ToolExecutionLogger) LogToolExecutionError(ctx context.Context, toolName string, parameters map[string]interface{}, duration time.Duration, err error, userID string) {
	requestID := tel.getRequestID(ctx)
	
	logEntry := ToolExecutionLog{
		RequestID:    requestID,
		Timestamp:    time.Now(),
		UserID:       userID,
		ToolName:     toolName,
		Duration:     duration.Milliseconds(),
		Success:      false,
		ErrorDetails: err.Error(),
		StackTrace:   tel.getStackTrace(),
	}

	// Add parameters if logging is enabled
	if tel.enableParameterLogging {
		logEntry.Parameters = tel.sanitizeParameters(parameters)
	}

	tel.logger.WithContext(ctx).Error("Tool execution failed",
		"request_id", logEntry.RequestID,
		"tool_name", logEntry.ToolName,
		"user_id", logEntry.UserID,
		"duration_ms", logEntry.Duration,
		"error", err.Error(),
		"success", logEntry.Success,
	)

	// Log detailed entry as JSON for structured logging
	if logData, err := json.Marshal(logEntry); err == nil {
		tel.logger.WithContext(ctx).Error("Tool execution error details", "log_data", string(logData))
	}
}

// LogToolExecutionTimeout logs a tool execution that timed out
func (tel *ToolExecutionLogger) LogToolExecutionTimeout(ctx context.Context, toolName string, parameters map[string]interface{}, timeout time.Duration, userID string) {
	requestID := tel.getRequestID(ctx)
	
	logEntry := ToolExecutionLog{
		RequestID:    requestID,
		Timestamp:    time.Now(),
		UserID:       userID,
		ToolName:     toolName,
		Duration:     timeout.Milliseconds(),
		Success:      false,
		ErrorDetails: fmt.Sprintf("Tool execution timed out after %v", timeout),
		Metadata: map[string]interface{}{
			"timeout_duration_ms": timeout.Milliseconds(),
			"timeout_reason":      "execution_timeout",
		},
	}

	// Add parameters if logging is enabled
	if tel.enableParameterLogging {
		logEntry.Parameters = tel.sanitizeParameters(parameters)
	}

	tel.logger.WithContext(ctx).Warn("Tool execution timeout",
		"request_id", logEntry.RequestID,
		"tool_name", logEntry.ToolName,
		"user_id", logEntry.UserID,
		"timeout_ms", timeout.Milliseconds(),
		"success", logEntry.Success,
	)

	// Log detailed entry as JSON for structured logging
	if logData, err := json.Marshal(logEntry); err == nil {
		tel.logger.WithContext(ctx).Warn("Tool execution timeout details", "log_data", string(logData))
	}
}

// sanitizeParameters removes or redacts sensitive parameters
func (tel *ToolExecutionLogger) sanitizeParameters(parameters map[string]interface{}) map[string]interface{} {
	if parameters == nil {
		return nil
	}

	sanitized := make(map[string]interface{})
	for key, value := range parameters {
		lowerKey := strings.ToLower(key)
		
		// Check if this is a sensitive parameter
		if tel.sensitiveParams[lowerKey] {
			sanitized[key] = "[REDACTED]"
		} else {
			// Check for sensitive substrings
			isSensitive := false
			for sensitiveKey := range tel.sensitiveParams {
				if strings.Contains(lowerKey, sensitiveKey) {
					sanitized[key] = "[REDACTED]"
					isSensitive = true
					break
				}
			}
			if !isSensitive {
				sanitized[key] = value
			}
		}
	}
	
	return sanitized
}

// getRequestID extracts request ID from context
func (tel *ToolExecutionLogger) getRequestID(ctx context.Context) string {
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	
	// Try to get from Gin context
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if requestID := ginCtx.GetString("request_id"); requestID != "" {
			return requestID
		}
	}
	
	// Generate a fallback request ID
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// getStackTrace captures the current stack trace
func (tel *ToolExecutionLogger) getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// LogToolValidationError logs parameter validation errors
func (tel *ToolExecutionLogger) LogToolValidationError(ctx context.Context, toolName string, parameters map[string]interface{}, validationError error, userID string) {
	requestID := tel.getRequestID(ctx)
	
	logEntry := ToolExecutionLog{
		RequestID:    requestID,
		Timestamp:    time.Now(),
		UserID:       userID,
		ToolName:     toolName,
		Duration:     0, // No execution time for validation errors
		Success:      false,
		ErrorDetails: fmt.Sprintf("Parameter validation failed: %v", validationError),
		Metadata: map[string]interface{}{
			"error_type": "validation_error",
		},
	}

	// Add parameters if logging is enabled
	if tel.enableParameterLogging {
		logEntry.Parameters = tel.sanitizeParameters(parameters)
	}

	tel.logger.WithContext(ctx).Warn("Tool parameter validation failed",
		"request_id", logEntry.RequestID,
		"tool_name", logEntry.ToolName,
		"user_id", logEntry.UserID,
		"validation_error", validationError.Error(),
		"success", logEntry.Success,
	)

	// Log detailed entry as JSON for structured logging
	if logData, err := json.Marshal(logEntry); err == nil {
		tel.logger.WithContext(ctx).Debug("Tool validation error details", "log_data", string(logData))
	}
}

// LogToolDiscovery logs tool discovery operations (listing tools, getting schemas)
func (tel *ToolExecutionLogger) LogToolDiscovery(ctx context.Context, operation string, toolName string, duration time.Duration, userID string) {
	requestID := tel.getRequestID(ctx)
	
	tel.logger.WithContext(ctx).Info("Tool discovery operation",
		"request_id", requestID,
		"operation", operation,
		"tool_name", toolName,
		"user_id", userID,
		"duration_ms", duration.Milliseconds(),
	)
}