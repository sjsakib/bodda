package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"
	"bodda/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ToolController handles HTTP requests for tool execution endpoints
type ToolController struct {
	registry  services.ToolRegistry
	executor  services.ToolExecutor
	config    *config.Config
}

// NewToolController creates a new tool controller
func NewToolController(registry services.ToolRegistry, executor services.ToolExecutor, config *config.Config) *ToolController {
	return &ToolController{
		registry: registry,
		executor: executor,
		config:   config,
	}
}

// ListTools handles GET /api/tools - returns all available tools
func (tc *ToolController) ListTools(c *gin.Context) {
	log.Printf("Listing available tools")
	
	tools := tc.registry.GetAvailableTools()
	
	response := models.ToolListResponse{
		Tools: tools,
		Count: len(tools),
	}
	
	log.Printf("Returning %d available tools", len(tools))
	c.JSON(http.StatusOK, response)
}

// GetToolSchema handles GET /api/tools/{toolName}/schema - returns schema for a specific tool
func (tc *ToolController) GetToolSchema(c *gin.Context) {
	requestID := tc.generateRequestID()
	startTime := time.Now()
	
	toolName := c.Param("toolName")
	if toolName == "" {
		toolErr := NewToolExecutionError(ErrorCodeMissingToolName, "Tool name is required").
			WithRequestID(requestID).
			WithDuration(time.Since(startTime))
		tc.sendToolError(c, toolErr)
		return
	}
	
	// Validate tool name format
	if err := tc.validateToolName(toolName); err != nil {
		toolErr := NewToolExecutionError(ErrorCodeValidationError, "Invalid tool name format").
			WithDetails(err.Error()).
			WithToolName(toolName).
			WithRequestID(requestID).
			WithDuration(time.Since(startTime))
		tc.sendToolError(c, toolErr)
		return
	}
	
	log.Printf("Getting schema for tool: %s (request: %s)", toolName, requestID)
	
	schema, err := tc.registry.GetToolSchema(toolName)
	if err != nil {
		log.Printf("Tool schema not found: %s, error: %v (request: %s)", toolName, err, requestID)
		
		toolErr := NewToolExecutionError(ErrorCodeToolNotFound, fmt.Sprintf("Tool '%s' not found", toolName)).
			WithDetails(err.Error()).
			WithToolName(toolName).
			WithRequestID(requestID).
			WithDuration(time.Since(startTime)).
			WithCause(err)
		tc.sendToolError(c, toolErr)
		return
	}
	
	response := models.ToolSchemaResponse{
		Schema: schema,
	}
	
	log.Printf("Returning schema for tool: %s (request: %s, duration: %dms)", 
		toolName, requestID, time.Since(startTime).Milliseconds())
	c.JSON(http.StatusOK, response)
}

// ExecuteTool handles POST /api/tools/execute - executes a tool with provided parameters
func (tc *ToolController) ExecuteTool(c *gin.Context) {
	requestID := tc.generateRequestID()
	startTime := time.Now()
	
	log.Printf("Starting tool execution request: %s", requestID)
	
	// Defer panic recovery with detailed error reporting
	defer func() {
		if r := recover(); r != nil {
			stackTrace := string(debug.Stack())
			log.Printf("Panic in tool execution (request: %s): %v\nStack trace: %s", requestID, r, stackTrace)
			
			toolErr := NewToolExecutionError(ErrorCodeInternalError, "Internal server error during tool execution").
				WithDetails(fmt.Sprintf("Panic: %v", r)).
				WithRequestID(requestID).
				WithDuration(time.Since(startTime)).
				WithStackTrace(stackTrace)
			tc.sendToolError(c, toolErr)
		}
	}()
	
	// Parse and validate request
	var req models.ToolExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request format for request %s: %v", requestID, err)
		
		toolErr := NewToolExecutionError(ErrorCodeInvalidRequest, "Invalid request format").
			WithDetails(err.Error()).
			WithRequestID(requestID).
			WithDuration(time.Since(startTime)).
			WithCause(err)
		tc.sendToolError(c, toolErr)
		return
	}
	
	// Comprehensive request validation
	if validationErr := tc.validateExecutionRequest(&req, requestID, startTime); validationErr != nil {
		tc.sendToolError(c, validationErr)
		return
	}
	
	// Get and validate user context
	userModel, authErr := tc.validateUserContext(c, requestID, startTime)
	if authErr != nil {
		tc.sendToolError(c, authErr)
		return
	}
	
	log.Printf("Executing tool %s for user %s (request %s)", req.ToolName, userModel.ID, requestID)
	
	// Create message context for tool execution
	msgCtx := &services.MessageContext{
		UserID:    userModel.ID,
		SessionID: "", // Tool execution doesn't require a session
		Message:   fmt.Sprintf("Tool execution: %s", req.ToolName),
		User:      userModel,
	}
	
	// Execute tool with comprehensive error handling
	result, execErr := tc.executeToolWithErrorHandling(c.Request.Context(), &req, msgCtx, requestID, startTime)
	if execErr != nil {
		tc.sendToolError(c, execErr)
		return
	}
	
	// Create successful response
	response := models.ToolExecutionResponse{
		Status: "success",
		Result: result,
		Metadata: &models.ResponseMetadata{
			RequestID: requestID,
			Timestamp: time.Now(),
			Duration:  time.Since(startTime).Milliseconds(),
		},
	}
	
	log.Printf("Tool execution completed successfully for request %s: tool=%s, duration=%dms", 
		requestID, req.ToolName, response.Metadata.Duration)
	
	c.JSON(http.StatusOK, response)
}

// getExecutionTimeout determines the execution timeout from options or defaults
func (tc *ToolController) getExecutionTimeout(options *models.ExecutionOptions) time.Duration {
	if options != nil && options.Timeout > 0 {
		// Cap at 5 minutes for safety
		if options.Timeout > 300 {
			return 300 * time.Second
		}
		return time.Duration(options.Timeout) * time.Second
	}
	// Default timeout of 30 seconds
	return 30 * time.Second
}

// generateRequestID generates a unique request ID for tracking
func (tc *ToolController) generateRequestID() string {
	return uuid.New().String()
}

// sendErrorResponse sends a standardized error response
func (tc *ToolController) sendErrorResponse(c *gin.Context, statusCode int, code, message, details string) {
	response := models.ErrorResponse{
		RequestID: tc.generateRequestID(),
		Timestamp: time.Now(),
	}
	response.Error.Code = code
	response.Error.Message = message
	response.Error.Details = details
	
	c.JSON(statusCode, response)
}

// sendErrorResponseWithMetadata sends a standardized error response with request metadata
func (tc *ToolController) sendErrorResponseWithMetadata(c *gin.Context, statusCode int, code, message, details, requestID string, startTime time.Time) {
	response := models.ErrorResponse{
		RequestID: requestID,
		Timestamp: time.Now(),
	}
	response.Error.Code = code
	response.Error.Message = message
	response.Error.Details = details
	
	log.Printf("Sending error response for request %s: code=%s, message=%s, duration=%dms", 
		requestID, code, message, time.Since(startTime).Milliseconds())
	
	c.JSON(statusCode, response)
}

// sendToolError sends a ToolExecutionError as an HTTP response
func (tc *ToolController) sendToolError(c *gin.Context, toolErr *ToolExecutionError) {
	statusCode := GetHTTPStatusCode(toolErr)
	response := toolErr.ToErrorResponse()
	
	// Log error with appropriate level based on severity
	if statusCode >= 500 {
		log.Printf("Server error for request %s: %s (duration: %dms)", 
			toolErr.RequestID, toolErr.Error(), toolErr.Duration)
		if toolErr.StackTrace != "" {
			log.Printf("Stack trace for request %s: %s", toolErr.RequestID, toolErr.StackTrace)
		}
	} else {
		log.Printf("Client error for request %s: %s (duration: %dms)", 
			toolErr.RequestID, toolErr.Error(), toolErr.Duration)
	}
	
	c.JSON(statusCode, response)
}

// validateToolName validates tool name format and security
func (tc *ToolController) validateToolName(toolName string) error {
	if toolName == "" {
		return errors.New("tool name cannot be empty")
	}
	
	if len(toolName) > 100 {
		return errors.New("tool name too long (max 100 characters)")
	}
	
	// Check for valid characters (alphanumeric, hyphens, underscores)
	for _, char := range toolName {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '-' || char == '_') {
			return errors.New("tool name contains invalid characters")
		}
	}
	
	// Prevent suspicious patterns
	suspiciousPatterns := []string{
		"..", "/", "\\", "script", "eval", "exec", "system", "cmd",
	}
	
	toolNameLower := strings.ToLower(toolName)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(toolNameLower, pattern) {
			return errors.New("tool name contains suspicious patterns")
		}
	}
	
	return nil
}

// validateExecutionRequest performs comprehensive validation of the execution request
func (tc *ToolController) validateExecutionRequest(req *models.ToolExecutionRequest, requestID string, startTime time.Time) *ToolExecutionError {
	// Validate tool name
	if req.ToolName == "" {
		return NewToolExecutionError(ErrorCodeMissingToolName, "Tool name is required").
			WithRequestID(requestID).
			WithDuration(time.Since(startTime))
	}
	
	// Validate tool name format
	if err := tc.validateToolName(req.ToolName); err != nil {
		return NewToolExecutionError(ErrorCodeValidationError, "Invalid tool name format").
			WithDetails(err.Error()).
			WithToolName(req.ToolName).
			WithRequestID(requestID).
			WithDuration(time.Since(startTime)).
			WithCause(err)
	}
	
	// Check if tool exists
	if !tc.registry.IsToolAvailable(req.ToolName) {
		return NewToolExecutionError(ErrorCodeToolNotFound, fmt.Sprintf("Tool '%s' not found", req.ToolName)).
			WithToolName(req.ToolName).
			WithRequestID(requestID).
			WithDuration(time.Since(startTime))
	}
	
	// Initialize parameters if nil
	if req.Parameters == nil {
		req.Parameters = make(map[string]interface{})
	}
	
	// Validate parameters against tool schema
	if err := tc.registry.ValidateToolCall(req.ToolName, req.Parameters); err != nil {
		validationErr := NewValidationError("Parameter validation failed")
		validationErr.ToolExecutionError = validationErr.ToolExecutionError.
			WithToolName(req.ToolName).
			WithParameters(req.Parameters).
			WithRequestID(requestID).
			WithDuration(time.Since(startTime)).
			WithCause(err)
		
		// Try to extract specific parameter errors
		if strings.Contains(err.Error(), "required") {
			validationErr.AddRequiredField("unknown")
		}
		if strings.Contains(err.Error(), "invalid") {
			validationErr.AddInvalidField("unknown")
		}
		
		return validationErr.ToolExecutionError
	}
	
	// Validate execution options
	if req.Options != nil {
		if err := tc.validateExecutionOptions(req.Options); err != nil {
			return NewToolExecutionError(ErrorCodeValidationError, "Invalid execution options").
				WithDetails(err.Error()).
				WithToolName(req.ToolName).
				WithRequestID(requestID).
				WithDuration(time.Since(startTime)).
				WithCause(err)
		}
	}
	
	return nil
}

// validateExecutionOptions validates execution options
func (tc *ToolController) validateExecutionOptions(options *models.ExecutionOptions) error {
	if options.Timeout < 0 {
		return errors.New("timeout cannot be negative")
	}
	
	if options.Timeout > 300 { // 5 minutes max
		return errors.New("timeout cannot exceed 300 seconds")
	}
	
	return nil
}

// validateUserContext validates and extracts user context
func (tc *ToolController) validateUserContext(c *gin.Context, requestID string, startTime time.Time) (*models.User, *ToolExecutionError) {
	user, exists := c.Get("user")
	if !exists {
		return nil, NewToolExecutionError(ErrorCodeAuthRequired, "Authentication required").
			WithRequestID(requestID).
			WithDuration(time.Since(startTime))
	}
	
	userModel, ok := user.(*models.User)
	if !ok {
		return nil, NewToolExecutionError(ErrorCodeInternalError, "Invalid user context").
			WithDetails("User context is not of expected type").
			WithRequestID(requestID).
			WithDuration(time.Since(startTime))
	}
	
	if userModel.ID == "" {
		return nil, NewToolExecutionError(ErrorCodeAuthRequired, "Invalid user ID").
			WithRequestID(requestID).
			WithDuration(time.Since(startTime))
	}
	
	return userModel, nil
}

// executeToolWithErrorHandling executes a tool with comprehensive error handling
func (tc *ToolController) executeToolWithErrorHandling(ctx context.Context, req *models.ToolExecutionRequest, msgCtx *services.MessageContext, requestID string, startTime time.Time) (*models.ToolExecutionResult, *ToolExecutionError) {
	// Create timeout context
	timeout := tc.getExecutionTimeout(req.Options)
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Execute tool
	result, err := tc.executor.ExecuteToolWithOptions(execCtx, req.ToolName, req.Parameters, msgCtx, req.Options)
	if err != nil {
		log.Printf("Tool execution failed for request %s: %v", requestID, err)
		
		// Handle specific error types with detailed context
		if execCtx.Err() == context.DeadlineExceeded {
			timeoutErr := NewTimeoutError("Tool execution timed out", timeout).
				WithExecutionPhase("tool_execution")
			timeoutErr.ToolExecutionError = timeoutErr.ToolExecutionError.
				WithToolName(req.ToolName).
				WithParameters(req.Parameters).
				WithRequestID(requestID).
				WithDuration(time.Since(startTime)).
				WithCause(err)
			return nil, timeoutErr.ToolExecutionError
		}
		
		if execCtx.Err() == context.Canceled {
			return nil, NewToolExecutionError(ErrorCodeExecutionCancelled, "Tool execution was cancelled").
				WithToolName(req.ToolName).
				WithParameters(req.Parameters).
				WithRequestID(requestID).
				WithDuration(time.Since(startTime)).
				WithCause(err)
		}
		
		// Check for service-specific errors
		if strings.Contains(err.Error(), "rate limit") {
			return nil, NewToolExecutionError(ErrorCodeRateLimitExceeded, "Rate limit exceeded").
				WithDetails(err.Error()).
				WithToolName(req.ToolName).
				WithRequestID(requestID).
				WithDuration(time.Since(startTime)).
				WithCause(err)
		}
		
		if strings.Contains(err.Error(), "unavailable") || strings.Contains(err.Error(), "service") {
			return nil, NewToolExecutionError(ErrorCodeServiceUnavailable, "Service temporarily unavailable").
				WithDetails(err.Error()).
				WithToolName(req.ToolName).
				WithRequestID(requestID).
				WithDuration(time.Since(startTime)).
				WithCause(err)
		}
		
		if strings.Contains(err.Error(), "resource") || strings.Contains(err.Error(), "memory") {
			return nil, NewToolExecutionError(ErrorCodeResourceExhausted, "Resource exhausted").
				WithDetails(err.Error()).
				WithToolName(req.ToolName).
				WithRequestID(requestID).
				WithDuration(time.Since(startTime)).
				WithCause(err)
		}
		
		// Generic execution error
		return nil, NewToolExecutionError(ErrorCodeExecutionError, "Tool execution failed").
			WithDetails(err.Error()).
			WithToolName(req.ToolName).
			WithParameters(req.Parameters).
			WithRequestID(requestID).
			WithDuration(time.Since(startTime)).
			WithCause(err)
	}
	
	// Validate result
	if result == nil {
		return nil, NewToolExecutionError(ErrorCodeInternalError, "Tool execution returned null result").
			WithToolName(req.ToolName).
			WithRequestID(requestID).
			WithDuration(time.Since(startTime))
	}
	
	return result, nil
}