package server

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"bodda/internal/models"
)

// Tool execution error codes
const (
	// Request validation errors
	ErrorCodeInvalidRequest     = "INVALID_REQUEST"
	ErrorCodeMissingToolName    = "MISSING_TOOL_NAME"
	ErrorCodeToolNotFound       = "TOOL_NOT_FOUND"
	ErrorCodeValidationError    = "VALIDATION_ERROR"
	ErrorCodeMaliciousInput     = "MALICIOUS_INPUT"
	ErrorCodeParameterTooLarge  = "PARAMETER_TOO_LARGE"
	ErrorCodeInvalidParameterType = "INVALID_PARAMETER_TYPE"

	// Authentication and authorization errors
	ErrorCodeAuthRequired       = "AUTH_REQUIRED"
	ErrorCodeInsufficientPerms  = "INSUFFICIENT_PERMISSIONS"
	ErrorCodeInvalidToken       = "INVALID_TOKEN"

	// Execution errors
	ErrorCodeExecutionError     = "EXECUTION_ERROR"
	ErrorCodeExecutionTimeout   = "EXECUTION_TIMEOUT"
	ErrorCodeExecutionCancelled = "EXECUTION_CANCELLED"
	ErrorCodeResourceExhausted  = "RESOURCE_EXHAUSTED"
	ErrorCodeServiceUnavailable = "SERVICE_UNAVAILABLE"

	// System errors
	ErrorCodeInternalError      = "INTERNAL_ERROR"
	ErrorCodeConfigurationError = "CONFIGURATION_ERROR"
	ErrorCodeDependencyError    = "DEPENDENCY_ERROR"

	// Rate limiting errors
	ErrorCodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	ErrorCodeConcurrencyLimit   = "CONCURRENCY_LIMIT_EXCEEDED"
)

// Predefined error instances for common scenarios
var (
	ErrToolNotFound         = errors.New("tool not found")
	ErrInvalidToolName      = errors.New("invalid tool name")
	ErrMissingParameters    = errors.New("missing required parameters")
	ErrInvalidParameters    = errors.New("invalid parameters")
	ErrMaliciousInput       = errors.New("malicious input detected")
	ErrExecutionTimeout     = errors.New("tool execution timed out")
	ErrExecutionCancelled   = errors.New("tool execution was cancelled")
	ErrAuthenticationRequired = errors.New("authentication required")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")
	ErrConcurrencyLimitExceeded = errors.New("concurrency limit exceeded")
	ErrServiceUnavailable   = errors.New("service temporarily unavailable")
)

// ToolExecutionError represents a detailed error from tool execution
type ToolExecutionError struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     string                 `json:"details,omitempty"`
	ToolName    string                 `json:"tool_name,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Duration    int64                  `json:"duration_ms,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Cause       error                  `json:"-"`
}

// Error implements the error interface
func (e *ToolExecutionError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error
func (e *ToolExecutionError) Unwrap() error {
	return e.Cause
}

// NewToolExecutionError creates a new tool execution error
func NewToolExecutionError(code, message string) *ToolExecutionError {
	return &ToolExecutionError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// WithDetails adds details to the error
func (e *ToolExecutionError) WithDetails(details string) *ToolExecutionError {
	e.Details = details
	return e
}

// WithToolName adds tool name to the error
func (e *ToolExecutionError) WithToolName(toolName string) *ToolExecutionError {
	e.ToolName = toolName
	return e
}

// WithParameters adds parameters to the error (sanitized for logging)
func (e *ToolExecutionError) WithParameters(params map[string]interface{}) *ToolExecutionError {
	e.Parameters = sanitizeParametersForLogging(params)
	return e
}

// WithRequestID adds request ID to the error
func (e *ToolExecutionError) WithRequestID(requestID string) *ToolExecutionError {
	e.RequestID = requestID
	return e
}

// WithDuration adds execution duration to the error
func (e *ToolExecutionError) WithDuration(duration time.Duration) *ToolExecutionError {
	e.Duration = duration.Milliseconds()
	return e
}

// WithCause adds the underlying cause error
func (e *ToolExecutionError) WithCause(cause error) *ToolExecutionError {
	e.Cause = cause
	return e
}

// WithStackTrace adds stack trace information
func (e *ToolExecutionError) WithStackTrace(stackTrace string) *ToolExecutionError {
	e.StackTrace = stackTrace
	return e
}

// ToErrorResponse converts the error to a standardized error response
func (e *ToolExecutionError) ToErrorResponse() *models.ErrorResponse {
	response := &models.ErrorResponse{
		RequestID: e.RequestID,
		Timestamp: e.Timestamp,
	}
	response.Error.Code = e.Code
	response.Error.Message = e.Message
	response.Error.Details = e.Details
	return response
}

// ValidationError represents parameter validation errors with detailed feedback
type ValidationError struct {
	*ToolExecutionError
	ParameterErrors map[string]string `json:"parameter_errors"`
	RequiredFields  []string          `json:"required_fields,omitempty"`
	InvalidFields   []string          `json:"invalid_fields,omitempty"`
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		ToolExecutionError: NewToolExecutionError(ErrorCodeValidationError, message),
		ParameterErrors:    make(map[string]string),
	}
}

// AddParameterError adds a parameter-specific error
func (e *ValidationError) AddParameterError(param, message string) *ValidationError {
	e.ParameterErrors[param] = message
	return e
}

// AddRequiredField adds a required field that was missing
func (e *ValidationError) AddRequiredField(field string) *ValidationError {
	e.RequiredFields = append(e.RequiredFields, field)
	return e
}

// AddInvalidField adds an invalid field
func (e *ValidationError) AddInvalidField(field string) *ValidationError {
	e.InvalidFields = append(e.InvalidFields, field)
	return e
}

// TimeoutError represents timeout-specific errors with context
type TimeoutError struct {
	*ToolExecutionError
	TimeoutDuration time.Duration `json:"timeout_duration_ms"`
	ExecutionPhase  string        `json:"execution_phase,omitempty"`
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(message string, timeout time.Duration) *TimeoutError {
	return &TimeoutError{
		ToolExecutionError: NewToolExecutionError(ErrorCodeExecutionTimeout, message),
		TimeoutDuration:    timeout,
	}
}

// WithExecutionPhase adds the execution phase where timeout occurred
func (e *TimeoutError) WithExecutionPhase(phase string) *TimeoutError {
	e.ExecutionPhase = phase
	return e
}

// sanitizeParametersForLogging removes sensitive information from parameters for logging
func sanitizeParametersForLogging(params map[string]interface{}) map[string]interface{} {
	if params == nil {
		return nil
	}

	sanitized := make(map[string]interface{})
	sensitiveKeys := []string{
		"password", "token", "secret", "key", "auth", "credential",
		"access_token", "refresh_token", "api_key", "private_key",
	}

	for k, v := range params {
		keyLower := strings.ToLower(k)
		isSensitive := false
		
		for _, sensitive := range sensitiveKeys {
			if strings.Contains(keyLower, sensitive) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			sanitized[k] = "[REDACTED]"
		} else {
			// Recursively sanitize nested objects
			switch val := v.(type) {
			case map[string]interface{}:
				sanitized[k] = sanitizeParametersForLogging(val)
			case string:
				// Truncate very long strings
				if len(val) > 1000 {
					sanitized[k] = val[:1000] + "...[TRUNCATED]"
				} else {
					sanitized[k] = val
				}
			default:
				sanitized[k] = val
			}
		}
	}

	return sanitized
}

// GetHTTPStatusCode returns the appropriate HTTP status code for the error
func GetHTTPStatusCode(err error) int {
	var toolErr *ToolExecutionError
	if errors.As(err, &toolErr) {
		switch toolErr.Code {
		case ErrorCodeInvalidRequest, ErrorCodeMissingToolName, ErrorCodeToolNotFound,
			 ErrorCodeValidationError, ErrorCodeMaliciousInput, ErrorCodeParameterTooLarge,
			 ErrorCodeInvalidParameterType:
			return 400 // Bad Request
		case ErrorCodeAuthRequired, ErrorCodeInvalidToken:
			return 401 // Unauthorized
		case ErrorCodeInsufficientPerms:
			return 403 // Forbidden
		case ErrorCodeExecutionTimeout:
			return 408 // Request Timeout
		case ErrorCodeRateLimitExceeded, ErrorCodeConcurrencyLimit:
			return 429 // Too Many Requests
		case ErrorCodeInternalError, ErrorCodeExecutionError, ErrorCodeDependencyError:
			return 500 // Internal Server Error
		case ErrorCodeServiceUnavailable, ErrorCodeResourceExhausted:
			return 503 // Service Unavailable
		case ErrorCodeConfigurationError:
			return 500 // Internal Server Error
		default:
			return 500 // Default to Internal Server Error
		}
	}
	
	// Handle standard Go errors
	if errors.Is(err, ErrToolNotFound) || errors.Is(err, ErrInvalidToolName) ||
	   errors.Is(err, ErrMissingParameters) || errors.Is(err, ErrInvalidParameters) {
		return 400
	}
	if errors.Is(err, ErrAuthenticationRequired) {
		return 401
	}
	if errors.Is(err, ErrInsufficientPermissions) {
		return 403
	}
	if errors.Is(err, ErrExecutionTimeout) {
		return 408
	}
	if errors.Is(err, ErrRateLimitExceeded) || errors.Is(err, ErrConcurrencyLimitExceeded) {
		return 429
	}
	if errors.Is(err, ErrServiceUnavailable) {
		return 503
	}
	
	return 500 // Default to Internal Server Error
}