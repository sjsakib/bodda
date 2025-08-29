package server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateToolName(t *testing.T) {
	controller := &ToolController{}
	
	tests := []struct {
		name        string
		toolName    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid tool name",
			toolName:    "get-athlete-profile",
			expectError: false,
		},
		{
			name:        "valid tool name with underscores",
			toolName:    "get_recent_activities",
			expectError: false,
		},
		{
			name:        "valid tool name with numbers",
			toolName:    "tool123",
			expectError: false,
		},
		{
			name:        "empty tool name",
			toolName:    "",
			expectError: true,
			errorMsg:    "tool name cannot be empty",
		},
		{
			name:        "tool name too long",
			toolName:    string(make([]byte, 101)),
			expectError: true,
			errorMsg:    "tool name too long",
		},
		{
			name:        "tool name with invalid characters",
			toolName:    "get-athlete/profile",
			expectError: true,
			errorMsg:    "tool name contains invalid characters",
		},
		{
			name:        "tool name with suspicious patterns",
			toolName:    "get-athlete-script",
			expectError: true,
			errorMsg:    "tool name contains suspicious patterns",
		},
		{
			name:        "tool name with path traversal",
			toolName:    "get-athlete../profile",
			expectError: true,
			errorMsg:    "tool name contains invalid characters",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := controller.validateToolName(tt.toolName)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateParameterKey(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid key",
			key:         "activity_id",
			expectError: false,
		},
		{
			name:        "valid key with hyphens",
			key:         "stream-types",
			expectError: false,
		},
		{
			name:        "empty key",
			key:         "",
			expectError: true,
			errorMsg:    "parameter key cannot be empty",
		},
		{
			name:        "key too long",
			key:         strings.Repeat("a", 101),
			expectError: true,
			errorMsg:    "parameter key too long",
		},
		{
			name:        "key with invalid characters",
			key:         "activity/id",
			expectError: true,
			errorMsg:    "parameter key contains invalid characters",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParameterKey(tt.key)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStringParameter(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid string",
			key:         "description",
			value:       "This is a valid description",
			expectError: false,
		},
		{
			name:        "valid file path",
			key:         "file_path",
			value:       "data/activities.json",
			expectError: false,
		},
		{
			name:        "string too long",
			key:         "description",
			value:       string(make([]byte, 10001)),
			expectError: true,
			errorMsg:    "parameter value too long",
		},
		{
			name:        "SQL injection attempt",
			key:         "query",
			value:       "'; DROP TABLE users; --",
			expectError: true,
			errorMsg:    "SQL injection",
		},
		{
			name:        "XSS attempt",
			key:         "content",
			value:       "<script>alert('xss')</script>",
			expectError: true,
			errorMsg:    "XSS attempt",
		},
		{
			name:        "path traversal in file path",
			key:         "file_path",
			value:       "../../../etc/passwd",
			expectError: true,
			errorMsg:    "path traversal",
		},
		{
			name:        "command injection",
			key:         "command",
			value:       "exec('rm -rf /')",
			expectError: true,
			errorMsg:    "command injection",
		},
		{
			name:        "template injection",
			key:         "template",
			value:       "${7*7}",
			expectError: true,
			errorMsg:    "template injection",
		},
		{
			name:        "null byte injection",
			key:         "filename",
			value:       "file.txt\x00.exe",
			expectError: true,
			errorMsg:    "suspicious characters",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStringParameter(tt.key, tt.value)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid relative path",
			path:        "data/activities.json",
			expectError: false,
		},
		{
			name:        "empty path",
			path:        "",
			expectError: false, // Empty paths are allowed
		},
		{
			name:        "path traversal with dots",
			path:        "../../../etc/passwd",
			expectError: true,
			errorMsg:    "path traversal detected",
		},
		{
			name:        "path traversal with encoded dots",
			path:        "%2e%2e/etc/passwd",
			expectError: true,
			errorMsg:    "path traversal detected",
		},
		{
			name:        "absolute path",
			path:        "/etc/passwd",
			expectError: true,
			errorMsg:    "absolute paths not allowed",
		},
		{
			name:        "access to git directory",
			path:        ".git/config",
			expectError: true,
			errorMsg:    "git repository",
		},
		{
			name:        "access to environment file",
			path:        ".env",
			expectError: true,
			errorMsg:    "environment files",
		},
		{
			name:        "access to node_modules",
			path:        "node_modules/package/index.js",
			expectError: true,
			errorMsg:    "node modules",
		},
		{
			name:        "executable file",
			path:        "malware.exe",
			expectError: true,
			errorMsg:    "potentially dangerous file extension",
		},
		{
			name:        "shell script",
			path:        "script.sh",
			expectError: true,
			errorMsg:    "potentially dangerous file extension",
		},
		{
			name:        "package.json access",
			path:        "package.json",
			expectError: true,
			errorMsg:    "package configuration",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFilePath(tt.path)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateParameters(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid parameters",
			params: map[string]interface{}{
				"activity_id": 12345,
				"stream_types": []interface{}{"time", "distance"},
				"resolution": "medium",
			},
			expectError: false,
		},
		{
			name: "invalid parameter key",
			params: map[string]interface{}{
				"activity/id": 12345,
			},
			expectError: true,
			errorMsg:    "parameter key 'activity/id'",
		},
		{
			name: "malicious parameter value",
			params: map[string]interface{}{
				"query": "'; DROP TABLE users; --",
			},
			expectError: true,
			errorMsg:    "parameter 'query'",
		},
		{
			name: "nested object with invalid value",
			params: map[string]interface{}{
				"config": map[string]interface{}{
					"script": "<script>alert('xss')</script>",
				},
			},
			expectError: true,
			errorMsg:    "parameter 'script'",
		},
		{
			name: "array with invalid values",
			params: map[string]interface{}{
				"commands": []interface{}{"ls", "exec('rm -rf /')"},
			},
			expectError: true,
			errorMsg:    "parameter 'commands'",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParameters(tt.params)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsFilePathParameter(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "file path parameter",
			key:      "file_path",
			expected: true,
		},
		{
			name:     "filename parameter",
			key:      "filename",
			expected: true,
		},
		{
			name:     "directory parameter",
			key:      "directory",
			expected: true,
		},
		{
			name:     "input file parameter",
			key:      "input_file",
			expected: true,
		},
		{
			name:     "regular parameter",
			key:      "activity_id",
			expected: false,
		},
		{
			name:     "description parameter",
			key:      "description",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFilePathParameter(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}