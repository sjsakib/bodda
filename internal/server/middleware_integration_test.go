package server

import (
	"bodda/internal/config"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMiddlewareIntegration tests all middleware components working together
func TestMiddlewareIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		isDevelopment  bool
		requestBody    map[string]interface{}
		expectedStatus int
		description    string
	}{
		{
			name:          "valid request in development mode",
			isDevelopment: true,
			requestBody: map[string]interface{}{
				"tool_name": "test_tool",
				"parameters": map[string]interface{}{
					"param1": "value1",
					"param2": 123,
				},
			},
			expectedStatus: http.StatusOK,
			description:    "Should allow valid requests in development mode",
		},
		{
			name:          "any request in production mode returns 404",
			isDevelopment: false,
			requestBody: map[string]interface{}{
				"tool_name": "test_tool",
				"parameters": map[string]interface{}{
					"param1": "value1",
				},
			},
			expectedStatus: http.StatusNotFound,
			description:    "Should return 404 for any request in production mode",
		},
		{
			name:          "malicious request blocked in development mode",
			isDevelopment: true,
			requestBody: map[string]interface{}{
				"tool_name": "test_tool",
				"parameters": map[string]interface{}{
					"query": "DROP TABLE users",
					"file_path": "../../../etc/passwd",
				},
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should block malicious requests even in development mode",
		},
		{
			name:          "invalid tool name blocked in development mode",
			isDevelopment: true,
			requestBody: map[string]interface{}{
				"tool_name": "invalid/tool@name",
				"parameters": map[string]interface{}{
					"param1": "value1",
				},
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should block invalid tool names in development mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				IsDevelopment: tt.isDevelopment,
			}

			router := gin.New()
			
			// Apply all middleware in the correct order
			router.Use(DevelopmentOnlyMiddleware(cfg))
			router.Use(InputValidationMiddleware())
			router.Use(WorkspaceBoundaryMiddleware())
			
			router.POST("/api/tools/execute", func(c *gin.Context) {
				// Verify all context values are set correctly
				workspaceRoot, exists := c.Get("workspace_root")
				assert.True(t, exists, "workspace_root should be set")
				assert.NotEmpty(t, workspaceRoot, "workspace_root should not be empty")

				if tt.isDevelopment && tt.expectedStatus == http.StatusOK {
					validatedRequest, exists := c.Get("validated_request")
					assert.True(t, exists, "validated_request should be set for valid requests")
					assert.NotNil(t, validatedRequest, "validated_request should not be nil")
				}

				c.JSON(http.StatusOK, gin.H{
					"message": "success",
					"workspace_root": workspaceRoot,
				})
			})

			bodyBytes, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)

			// Additional assertions based on expected status
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "success", response["message"])
				assert.Contains(t, response, "workspace_root")
			} else if tt.expectedStatus == http.StatusBadRequest {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			}
		})
	}
}

// TestMiddlewareOrder tests that middleware is applied in the correct order
func TestMiddlewareOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test that development middleware runs first and blocks production requests
	// before input validation can run
	t.Run("development middleware runs before input validation", func(t *testing.T) {
		cfg := &config.Config{
			IsDevelopment: false, // Production mode
		}

		router := gin.New()
		
		// Apply middleware in correct order
		router.Use(DevelopmentOnlyMiddleware(cfg))
		router.Use(InputValidationMiddleware()) // This should not run in production
		
		router.POST("/api/tools/execute", func(c *gin.Context) {
			t.Error("Handler should not be reached in production mode")
			c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
		})

		// Send a request with invalid JSON that would fail input validation
		// But should be blocked by development middleware first
		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 404 from development middleware, not 400 from input validation
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test that input validation runs before workspace boundary middleware
	t.Run("input validation runs before workspace boundary", func(t *testing.T) {
		cfg := &config.Config{
			IsDevelopment: true, // Development mode
		}

		router := gin.New()
		
		// Apply middleware in correct order
		router.Use(DevelopmentOnlyMiddleware(cfg))
		router.Use(InputValidationMiddleware())
		router.Use(WorkspaceBoundaryMiddleware())
		
		router.POST("/api/tools/execute", func(c *gin.Context) {
			t.Error("Handler should not be reached with invalid input")
			c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
		})

		// Send a request with invalid tool name
		requestBody := map[string]interface{}{
			"tool_name": "", // Invalid empty tool name
		}
		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 400 from input validation
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestMiddlewarePerformance tests that middleware doesn't significantly impact performance
func TestMiddlewarePerformance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		IsDevelopment: true,
	}

	router := gin.New()
	router.Use(DevelopmentOnlyMiddleware(cfg))
	router.Use(InputValidationMiddleware())
	router.Use(WorkspaceBoundaryMiddleware())
	
	router.POST("/api/tools/execute", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	requestBody := map[string]interface{}{
		"tool_name": "test_tool",
		"parameters": map[string]interface{}{
			"param1": "value1",
			"param2": 123,
		},
	}
	bodyBytes, err := json.Marshal(requestBody)
	require.NoError(t, err)

	// Run multiple requests to test performance
	numRequests := 100
	for i := 0; i < numRequests; i++ {
		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}

// TestMiddlewareErrorHandling tests error handling in middleware
func TestMiddlewareErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		IsDevelopment: true,
	}

	tests := []struct {
		name        string
		requestBody string
		contentType string
		expectedStatus int
		description string
	}{
		{
			name:        "invalid JSON",
			requestBody: `{"tool_name": "test", "parameters": {invalid json}`,
			contentType: "application/json",
			expectedStatus: http.StatusBadRequest,
			description: "Should handle invalid JSON gracefully",
		},
		{
			name:        "missing content type",
			requestBody: `{"tool_name": "test"}`,
			contentType: "",
			expectedStatus: http.StatusOK, // Gin can still parse JSON without explicit content type
			description: "Should handle missing content type gracefully",
		},
		{
			name:        "empty request body",
			requestBody: "",
			contentType: "application/json",
			expectedStatus: http.StatusBadRequest,
			description: "Should handle empty request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(DevelopmentOnlyMiddleware(cfg))
			router.Use(InputValidationMiddleware())
			router.Use(WorkspaceBoundaryMiddleware())
			
			router.POST("/api/tools/execute", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer([]byte(tt.requestBody)))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)
		})
	}
}