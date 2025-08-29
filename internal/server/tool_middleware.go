package server

import (
	"bodda/internal/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DevelopmentOnlyMiddleware ensures tool execution endpoints are only available in development mode
func DevelopmentOnlyMiddleware(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.IsDevelopment {
			// Return 404 as if endpoint doesn't exist in production
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Next()
	}
}

// InputValidationMiddleware provides input validation and sanitization for tool execution
func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate and sanitize request body for tool execution
		if c.Request.Method == "POST" && strings.Contains(c.Request.URL.Path, "/tools/execute") {
			if err := validateToolExecutionRequest(c); err != nil {
				requestID := generateRequestID()
				
				// Create detailed validation error response
				toolErr := NewValidationError("Request validation failed")
				toolErr.ToolExecutionError = toolErr.ToolExecutionError.
					WithRequestID(requestID).
					WithCause(err)
				
				// Try to extract specific validation details
				if strings.Contains(err.Error(), "malicious") {
					toolErr.ToolExecutionError.Code = ErrorCodeMaliciousInput
					toolErr.ToolExecutionError.Message = "Malicious input detected"
				} else if strings.Contains(err.Error(), "too long") {
					toolErr.ToolExecutionError.Code = ErrorCodeParameterTooLarge
					toolErr.ToolExecutionError.Message = "Parameter value too large"
				}
				
				statusCode := GetHTTPStatusCode(toolErr.ToolExecutionError)
				response := toolErr.ToolExecutionError.ToErrorResponse()
				
				log.Printf("Input validation failed for request %s: %s", requestID, err.Error())
				c.JSON(statusCode, response)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// generateRequestID generates a unique request ID for tracking
func generateRequestID() string {
	return uuid.New().String()
}

// WorkspaceBoundaryMiddleware enforces workspace boundaries for file system operations
func WorkspaceBoundaryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store workspace root in context for later validation
		// This will be used by tool executors to validate file paths
		c.Set("workspace_root", getWorkspaceRoot())
		c.Next()
	}
}

// validateToolExecutionRequest validates and sanitizes tool execution requests
func validateToolExecutionRequest(c *gin.Context) error {
	// Read the body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	
	// Restore the body for the controller to read
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	
	// Parse the JSON for validation
	var request map[string]interface{}
	if err := json.Unmarshal(body, &request); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	// Validate tool name
	toolName, exists := request["tool_name"]
	if !exists {
		return fmt.Errorf("tool_name is required")
	}

	toolNameStr, ok := toolName.(string)
	if !ok {
		return fmt.Errorf("tool_name must be a string")
	}

	if err := validateToolName(toolNameStr); err != nil {
		return fmt.Errorf("invalid tool_name: %w", err)
	}

	// Validate parameters if present
	if params, exists := request["parameters"]; exists && params != nil {
		if paramsMap, ok := params.(map[string]interface{}); ok {
			if err := validateParameters(paramsMap); err != nil {
				return fmt.Errorf("invalid parameters: %w", err)
			}
		} else {
			return fmt.Errorf("parameters must be an object")
		}
	}

	// Store validated request back in context
	c.Set("validated_request", request)
	return nil
}

// validateToolName ensures tool name is safe and valid
func validateToolName(toolName string) error {
	if toolName == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	// Allow only alphanumeric characters, underscores, and hyphens
	validToolName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validToolName.MatchString(toolName) {
		return fmt.Errorf("tool name contains invalid characters")
	}

	// Prevent excessively long tool names
	if len(toolName) > 100 {
		return fmt.Errorf("tool name too long")
	}

	return nil
}

// validateParameters validates and sanitizes tool parameters with detailed error feedback
func validateParameters(params map[string]interface{}) error {
	var validationErrors []string
	
	for key, value := range params {
		// Validate parameter keys
		if err := validateParameterKey(key); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("parameter key '%s': %s", key, err.Error()))
			continue
		}

		// Validate parameter values
		if err := validateParameterValue(key, value); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("parameter '%s': %s", key, err.Error()))
		}
	}
	
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(validationErrors, "; "))
	}
	
	return nil
}

// validateParameterKey ensures parameter keys are safe
func validateParameterKey(key string) error {
	if key == "" {
		return fmt.Errorf("parameter key cannot be empty")
	}

	// Allow only alphanumeric characters, underscores, and hyphens
	validKey := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validKey.MatchString(key) {
		return fmt.Errorf("parameter key contains invalid characters")
	}

	// Prevent excessively long keys
	if len(key) > 100 {
		return fmt.Errorf("parameter key too long")
	}

	return nil
}

// validateParameterValue validates parameter values and detects malicious content
func validateParameterValue(key string, value interface{}) error {
	switch v := value.(type) {
	case string:
		return validateStringParameter(key, v)
	case map[string]interface{}:
		// Recursively validate nested objects
		return validateParameters(v)
	case []interface{}:
		// Validate array elements
		for i, item := range v {
			if err := validateParameterValue(fmt.Sprintf("%s[%d]", key, i), item); err != nil {
				return err
			}
		}
	case nil, bool, float64, int:
		// These types are generally safe
		return nil
	default:
		return fmt.Errorf("unsupported parameter type: %T", value)
	}
	return nil
}

// validateStringParameter validates string parameters and detects malicious content
func validateStringParameter(key, value string) error {
	// Check for excessively long strings with different limits based on parameter type
	maxLength := 10000
	if isFilePathParameter(key) {
		maxLength = 1000 // File paths should be shorter
	} else if strings.Contains(strings.ToLower(key), "content") {
		maxLength = 50000 // Content parameters can be longer
	}
	
	if len(value) > maxLength {
		return fmt.Errorf("parameter value too long (max %d characters)", maxLength)
	}

	// Detect potential injection patterns with more comprehensive coverage
	maliciousPatterns := []struct {
		pattern string
		description string
	}{
		{`(?i)(union\s+select)`, "SQL injection"},
		{`(?i)(drop\s+table)`, "SQL injection"},
		{`(?i)(delete\s+from)`, "SQL injection"},
		{`(?i)(insert\s+into)`, "SQL injection"},
		{`(?i)(update\s+.+set)`, "SQL injection"},
		{`(?i)(exec\s*\()`, "command injection"},
		{`(?i)(system\s*\()`, "command injection"},
		{`(?i)(eval\s*\()`, "code injection"},
		{`(?i)(<\s*script)`, "XSS attempt"},
		{`(?i)(javascript\s*:)`, "XSS attempt"},
		{`(?i)(on\w+\s*=)`, "XSS attempt"},
		{`(?i)(\.\./)`, "path traversal"},
		{`(?i)(\.\.\\)`, "path traversal"},
		{`(?i)(\$\{)`, "template injection"},
		{`(?i)(<%.*%>)`, "template injection"},
		{`(?i)(__import__)`, "Python injection"},
		{`(?i)(require\s*\()`, "Node.js injection"},
	}

	for _, malicious := range maliciousPatterns {
		if matched, _ := regexp.MatchString(malicious.pattern, value); matched {
			return fmt.Errorf("potentially malicious content detected: %s", malicious.description)
		}
	}

	// Check for suspicious character sequences
	suspiciousChars := []string{
		"\x00", "\x01", "\x02", "\x03", "\x04", "\x05", // null bytes and control chars
		"\r\n\r\n", // HTTP header injection
		"<?php", "<?=", // PHP injection
	}
	
	for _, suspicious := range suspiciousChars {
		if strings.Contains(value, suspicious) {
			return fmt.Errorf("potentially malicious content detected: suspicious characters")
		}
	}

	// For file path parameters, validate against path traversal
	if isFilePathParameter(key) {
		if err := validateFilePath(value); err != nil {
			return fmt.Errorf("invalid file path: %w", err)
		}
	}

	return nil
}

// isFilePathParameter determines if a parameter likely contains a file path
func isFilePathParameter(key string) bool {
	filePathKeys := []string{
		"path", "file", "filename", "filepath", "directory", "dir",
		"input_file", "output_file", "config_file", "log_file",
	}

	keyLower := strings.ToLower(key)
	for _, pathKey := range filePathKeys {
		if strings.Contains(keyLower, pathKey) {
			return true
		}
	}
	return false
}

// validateFilePath validates file paths and prevents path traversal attacks
func validateFilePath(path string) error {
	if path == "" {
		return nil // Empty paths are handled by individual tools
	}

	// Prevent path traversal attempts with multiple encodings
	traversalPatterns := []string{
		"..", "..\\", "../", "..\\\\",
		"%2e%2e", "%2e%2e%2f", "%2e%2e%5c",
		"..%2f", "..%5c", "%2e%2e/", "%2e%2e\\",
		"....//", "....\\\\",
	}
	
	pathLower := strings.ToLower(path)
	for _, pattern := range traversalPatterns {
		if strings.Contains(pathLower, pattern) {
			return fmt.Errorf("path traversal detected: %s", pattern)
		}
	}

	// Prevent absolute paths outside workspace
	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute paths not allowed")
	}

	// Clean the path and check for suspicious modifications
	cleanPath := filepath.Clean(path)
	if cleanPath != path && cleanPath != "./"+path {
		return fmt.Errorf("path contains suspicious elements (cleaned: %s)", cleanPath)
	}

	// Prevent access to sensitive directories and files
	sensitivePatterns := []struct {
		pattern string
		description string
	}{
		{"/etc/", "system configuration"},
		{"/proc/", "process information"},
		{"/sys/", "system information"},
		{"/dev/", "device files"},
		{"/.git/", "git repository"},
		{"/.env", "environment files"},
		{"/node_modules/", "node modules"},
		{"/tmp/", "temporary files"},
		{"/var/", "variable data"},
		{"/usr/", "user programs"},
		{"/bin/", "system binaries"},
		{"/sbin/", "system binaries"},
		{".git/", "git repository"},
		{".env", "environment files"},
		{"node_modules/", "node modules"},
		{"package.json", "package configuration"},
		{"package-lock.json", "package lock"},
		{"go.mod", "go module"},
		{"go.sum", "go dependencies"},
		{"Dockerfile", "docker configuration"},
		{"docker-compose", "docker compose"},
		{".ssh/", "SSH keys"},
		{".aws/", "AWS credentials"},
		{".kube/", "Kubernetes config"},
	}

	for _, sensitive := range sensitivePatterns {
		if strings.Contains(pathLower, sensitive.pattern) {
			return fmt.Errorf("access to %s not allowed", sensitive.description)
		}
	}

	// Check for suspicious file extensions
	suspiciousExtensions := []string{
		".exe", ".bat", ".cmd", ".com", ".scr", ".pif",
		".sh", ".bash", ".zsh", ".fish",
		".ps1", ".psm1", ".psd1",
		".vbs", ".vbe", ".js", ".jse",
		".jar", ".class",
	}
	
	ext := strings.ToLower(filepath.Ext(path))
	for _, suspiciousExt := range suspiciousExtensions {
		if ext == suspiciousExt {
			return fmt.Errorf("potentially dangerous file extension: %s", ext)
		}
	}

	return nil
}

// getWorkspaceRoot returns the current workspace root directory
func getWorkspaceRoot() string {
	// In a real implementation, this would be configurable
	// For now, assume current working directory is the workspace root
	return "."
}

// ValidateWorkspacePath validates that a file path is within the workspace boundaries
func ValidateWorkspacePath(workspaceRoot, targetPath string) error {
	if workspaceRoot == "" {
		return fmt.Errorf("workspace root not set")
	}

	// Convert to absolute paths for comparison
	absWorkspace, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return fmt.Errorf("invalid workspace root: %w", err)
	}

	absTarget, err := filepath.Abs(filepath.Join(workspaceRoot, targetPath))
	if err != nil {
		return fmt.Errorf("invalid target path: %w", err)
	}

	// Ensure target path is within workspace
	relPath, err := filepath.Rel(absWorkspace, absTarget)
	if err != nil {
		return fmt.Errorf("cannot determine relative path: %w", err)
	}

	// Check if path escapes workspace
	if strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("path escapes workspace boundaries")
	}

	return nil
}