package server

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewToolExecutionError(t *testing.T) {
	code := ErrorCodeValidationError
	message := "Test error message"
	
	err := NewToolExecutionError(code, message)
	
	assert.Equal(t, code, err.Code)
	assert.Equal(t, message, err.Message)
	assert.False(t, err.Timestamp.IsZero())
}

func TestToolExecutionError_WithMethods(t *testing.T) {
	err := NewToolExecutionError(ErrorCodeExecutionError, "Test error")
	
	// Test method chaining
	err = err.WithDetails("Test details").
		WithToolName("test-tool").
		WithRequestID("test-request-id").
		WithDuration(100 * time.Millisecond).
		WithCause(errors.New("underlying error"))
	
	assert.Equal(t, "Test details", err.Details)
	assert.Equal(t, "test-tool", err.ToolName)
	assert.Equal(t, "test-request-id", err.RequestID)
	assert.Equal(t, int64(100), err.Duration)
	assert.NotNil(t, err.Cause)
	assert.Equal(t, "underlying error", err.Cause.Error())
}

func TestToolExecutionError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ToolExecutionError
		expected string
	}{
		{
			name: "error without details",
			err:  NewToolExecutionError(ErrorCodeValidationError, "Test message"),
			expected: "VALIDATION_ERROR: Test message",
		},
		{
			name: "error with details",
			err:  NewToolExecutionError(ErrorCodeValidationError, "Test message").WithDetails("Test details"),
			expected: "VALIDATION_ERROR: Test message (Test details)",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestToolExecutionError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewToolExecutionError(ErrorCodeExecutionError, "Test error").WithCause(cause)
	
	assert.Equal(t, cause, err.Unwrap())
	assert.True(t, errors.Is(err, cause))
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("Validation failed")
	
	assert.Equal(t, ErrorCodeValidationError, err.Code)
	assert.Equal(t, "Validation failed", err.Message)
	assert.NotNil(t, err.ParameterErrors)
	assert.Empty(t, err.ParameterErrors)
}

func TestValidationError_Methods(t *testing.T) {
	err := NewValidationError("Validation failed")
	
	err = err.AddParameterError("param1", "Invalid value").
		AddRequiredField("required_field").
		AddInvalidField("invalid_field")
	
	assert.Equal(t, "Invalid value", err.ParameterErrors["param1"])
	assert.Contains(t, err.RequiredFields, "required_field")
	assert.Contains(t, err.InvalidFields, "invalid_field")
}

func TestNewTimeoutError(t *testing.T) {
	timeout := 30 * time.Second
	err := NewTimeoutError("Execution timed out", timeout)
	
	assert.Equal(t, ErrorCodeExecutionTimeout, err.Code)
	assert.Equal(t, "Execution timed out", err.Message)
	assert.Equal(t, timeout, err.TimeoutDuration)
}

func TestTimeoutError_WithExecutionPhase(t *testing.T) {
	err := NewTimeoutError("Timeout", 30*time.Second).WithExecutionPhase("initialization")
	
	assert.Equal(t, "initialization", err.ExecutionPhase)
}

func TestSanitizeParametersForLogging(t *testing.T) {
	params := map[string]interface{}{
		"username":     "testuser",
		"password":     "secret123",
		"access_token": "token123",
		"data":         "some data",
		"nested": map[string]interface{}{
			"api_key": "key123",
			"value":   "normal value",
		},
		"long_string": string(make([]byte, 1500)),
	}
	
	sanitized := sanitizeParametersForLogging(params)
	
	assert.Equal(t, "testuser", sanitized["username"])
	assert.Equal(t, "[REDACTED]", sanitized["password"])
	assert.Equal(t, "[REDACTED]", sanitized["access_token"])
	assert.Equal(t, "some data", sanitized["data"])
	
	nested := sanitized["nested"].(map[string]interface{})
	assert.Equal(t, "[REDACTED]", nested["api_key"])
	assert.Equal(t, "normal value", nested["value"])
	
	longString := sanitized["long_string"].(string)
	assert.True(t, len(longString) <= 1014, "Expected length <= 1014, got %d", len(longString)) // 1000 + "...[TRUNCATED]"
	assert.Contains(t, longString, "...[TRUNCATED]")
}

func TestGetHTTPStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "validation error",
			err:      NewToolExecutionError(ErrorCodeValidationError, "Validation failed"),
			expected: 400,
		},
		{
			name:     "auth required",
			err:      NewToolExecutionError(ErrorCodeAuthRequired, "Auth required"),
			expected: 401,
		},
		{
			name:     "insufficient permissions",
			err:      NewToolExecutionError(ErrorCodeInsufficientPerms, "Insufficient permissions"),
			expected: 403,
		},
		{
			name:     "timeout error",
			err:      NewToolExecutionError(ErrorCodeExecutionTimeout, "Timeout"),
			expected: 408,
		},
		{
			name:     "rate limit",
			err:      NewToolExecutionError(ErrorCodeRateLimitExceeded, "Rate limit"),
			expected: 429,
		},
		{
			name:     "internal error",
			err:      NewToolExecutionError(ErrorCodeInternalError, "Internal error"),
			expected: 500,
		},
		{
			name:     "service unavailable",
			err:      NewToolExecutionError(ErrorCodeServiceUnavailable, "Service unavailable"),
			expected: 503,
		},
		{
			name:     "standard error",
			err:      ErrToolNotFound,
			expected: 400,
		},
		{
			name:     "auth error",
			err:      ErrAuthenticationRequired,
			expected: 401,
		},
		{
			name:     "timeout standard error",
			err:      ErrExecutionTimeout,
			expected: 408,
		},
		{
			name:     "unknown error",
			err:      errors.New("unknown error"),
			expected: 500,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode := GetHTTPStatusCode(tt.err)
			assert.Equal(t, tt.expected, statusCode)
		})
	}
}

func TestToolExecutionError_ToErrorResponse(t *testing.T) {
	err := NewToolExecutionError(ErrorCodeValidationError, "Validation failed").
		WithDetails("Parameter 'name' is required").
		WithRequestID("test-request-123")
	
	response := err.ToErrorResponse()
	
	assert.Equal(t, "test-request-123", response.RequestID)
	assert.Equal(t, ErrorCodeValidationError, response.Error.Code)
	assert.Equal(t, "Validation failed", response.Error.Message)
	assert.Equal(t, "Parameter 'name' is required", response.Error.Details)
	assert.False(t, response.Timestamp.IsZero())
}