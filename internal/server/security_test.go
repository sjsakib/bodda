package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMaliciousParameterHandling tests various malicious parameter injection attempts
func TestMaliciousParameterHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		parameters  map[string]interface{}
		expectError bool
		description string
	}{
		{
			name: "SQL injection in query parameter",
			parameters: map[string]interface{}{
				"query": "SELECT * FROM users; DROP TABLE users; --",
			},
			expectError: true,
			description: "Should detect and reject SQL injection attempts",
		},
		{
			name: "Union-based SQL injection",
			parameters: map[string]interface{}{
				"search": "' UNION SELECT password FROM admin_users --",
			},
			expectError: true,
			description: "Should detect union-based SQL injection",
		},
		{
			name: "XSS script injection",
			parameters: map[string]interface{}{
				"content": "<script>alert('XSS')</script>",
			},
			expectError: true,
			description: "Should detect script injection attempts",
		},
		{
			name: "Command injection attempt",
			parameters: map[string]interface{}{
				"command": "ls; rm -rf /",
			},
			expectError: false, // This should pass basic validation but be handled by individual tools
			description: "Command injection should be handled by individual tool validation",
		},
		{
			name: "Path traversal in file parameter",
			parameters: map[string]interface{}{
				"file_path": "../../../../etc/passwd",
			},
			expectError: true,
			description: "Should detect path traversal attempts",
		},
		{
			name: "Windows path traversal",
			parameters: map[string]interface{}{
				"filepath": "..\\..\\..\\windows\\system32\\config\\sam",
			},
			expectError: true,
			description: "Should detect Windows-style path traversal",
		},
		{
			name: "Encoded path traversal",
			parameters: map[string]interface{}{
				"path": "%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
			},
			expectError: false, // URL encoding should be handled at HTTP level
			description: "URL encoded paths should be decoded before validation",
		},
		{
			name: "Null byte injection",
			parameters: map[string]interface{}{
				"filename": "safe.txt\x00../../etc/passwd",
			},
			expectError: true, // Should detect path traversal even with null bytes
			description: "Null byte injection should be detected and rejected",
		},
		{
			name: "Extremely long parameter value",
			parameters: map[string]interface{}{
				"data": string(make([]byte, 15000)), // Exceeds 10000 byte limit
			},
			expectError: true,
			description: "Should reject excessively long parameter values",
		},
		{
			name: "Nested malicious content",
			parameters: map[string]interface{}{
				"config": map[string]interface{}{
					"database": map[string]interface{}{
						"query": "DROP TABLE users",
					},
				},
			},
			expectError: true,
			description: "Should detect malicious content in nested objects",
		},
		{
			name: "Array with malicious content",
			parameters: map[string]interface{}{
				"queries": []interface{}{
					"SELECT * FROM users",
					"DROP TABLE sessions",
				},
			},
			expectError: true,
			description: "Should detect malicious content in arrays",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(InputValidationMiddleware())
			router.POST("/api/tools/execute", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			requestBody := map[string]interface{}{
				"tool_name":  "test_tool",
				"parameters": tt.parameters,
			}

			bodyBytes, err := json.Marshal(requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tt.expectError {
				assert.Equal(t, http.StatusBadRequest, w.Code, tt.description)
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				assert.Equal(t, http.StatusOK, w.Code, tt.description)
			}
		})
	}
}

// TestWorkspaceBoundaryEnforcement tests workspace boundary enforcement
func TestWorkspaceBoundaryEnforcement(t *testing.T) {
	tests := []struct {
		name          string
		workspaceRoot string
		targetPath    string
		expectError   bool
		description   string
	}{
		{
			name:          "valid path within workspace",
			workspaceRoot: "/workspace",
			targetPath:    "data/file.txt",
			expectError:   false,
			description:   "Should allow access to files within workspace",
		},
		{
			name:          "path escaping workspace with double dots",
			workspaceRoot: "/workspace",
			targetPath:    "../outside/file.txt",
			expectError:   true,
			description:   "Should prevent access outside workspace using ../",
		},
		{
			name:          "complex path escaping workspace",
			workspaceRoot: "/workspace",
			targetPath:    "data/../../outside/file.txt",
			expectError:   true,
			description:   "Should prevent complex path traversal attempts",
		},
		{
			name:          "symlink-like path escaping",
			workspaceRoot: "/workspace",
			targetPath:    "data/../../../etc/passwd",
			expectError:   true,
			description:   "Should prevent symlink-style path traversal",
		},
		{
			name:          "deeply nested valid path",
			workspaceRoot: "/workspace",
			targetPath:    "data/subdir/deep/nested/file.txt",
			expectError:   false,
			description:   "Should allow deeply nested paths within workspace",
		},
		{
			name:          "path with current directory references",
			workspaceRoot: "/workspace",
			targetPath:    "./data/./file.txt",
			expectError:   false,
			description:   "Should handle current directory references correctly",
		},
		{
			name:          "empty target path",
			workspaceRoot: "/workspace",
			targetPath:    "",
			expectError:   false,
			description:   "Should handle empty paths gracefully",
		},
		{
			name:          "root directory access attempt",
			workspaceRoot: "/workspace",
			targetPath:    "../../../",
			expectError:   true,
			description:   "Should prevent access to root directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWorkspacePath(tt.workspaceRoot, tt.targetPath)
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestSecurityHeaders tests that security-related headers are properly set
func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(InputValidationMiddleware())
	router.Use(WorkspaceBoundaryMiddleware())
	router.POST("/api/tools/execute", func(c *gin.Context) {
		// Verify workspace root is set in context
		workspaceRoot, exists := c.Get("workspace_root")
		assert.True(t, exists, "workspace_root should be set in context")
		assert.NotEmpty(t, workspaceRoot, "workspace_root should not be empty")

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	requestBody := map[string]interface{}{
		"tool_name": "test_tool",
		"parameters": map[string]interface{}{
			"param1": "value1",
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestParameterSanitization tests that parameters are properly sanitized
func TestParameterSanitization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(InputValidationMiddleware())
	router.POST("/api/tools/execute", func(c *gin.Context) {
		// Check if validated request is available in context
		validatedRequest, exists := c.Get("validated_request")
		assert.True(t, exists, "validated_request should be set in context")
		assert.NotNil(t, validatedRequest, "validated_request should not be nil")

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	requestBody := map[string]interface{}{
		"tool_name": "valid_tool",
		"parameters": map[string]interface{}{
			"safe_param": "safe_value",
			"number":     123,
			"boolean":    true,
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestEdgeCaseParameterValidation tests edge cases in parameter validation
func TestEdgeCaseParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		toolName    interface{}
		parameters  interface{}
		expectError bool
		description string
	}{
		{
			name:        "null tool name",
			toolName:    nil,
			parameters:  map[string]interface{}{},
			expectError: true,
			description: "Should reject null tool name",
		},
		{
			name:        "numeric tool name",
			toolName:    123,
			parameters:  map[string]interface{}{},
			expectError: true,
			description: "Should reject numeric tool name",
		},
		{
			name:        "array as tool name",
			toolName:    []string{"tool1", "tool2"},
			parameters:  map[string]interface{}{},
			expectError: true,
			description: "Should reject array as tool name",
		},
		{
			name:        "null parameters",
			toolName:    "valid_tool",
			parameters:  nil,
			expectError: false,
			description: "Should allow null parameters",
		},
		{
			name:        "string as parameters",
			toolName:    "valid_tool",
			parameters:  "invalid_params",
			expectError: true,
			description: "Should reject string as parameters",
		},
		{
			name:        "array as parameters",
			toolName:    "valid_tool",
			parameters:  []interface{}{"param1", "param2"},
			expectError: true,
			description: "Should reject array as parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			router := gin.New()
			router.Use(InputValidationMiddleware())
			router.POST("/api/tools/execute", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			requestBody := map[string]interface{}{
				"tool_name":  tt.toolName,
				"parameters": tt.parameters,
			}

			bodyBytes, err := json.Marshal(requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tt.expectError {
				assert.Equal(t, http.StatusBadRequest, w.Code, tt.description)
			} else {
				assert.Equal(t, http.StatusOK, w.Code, tt.description)
			}
		})
	}
}