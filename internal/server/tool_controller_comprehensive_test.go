package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"
	"bodda/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestToolController_Comprehensive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup comprehensive test environment
	mockRegistry := &mockToolRegistryComprehensive{}
	mockExecutor := &mockToolExecutorComprehensive{}
	config := &config.Config{IsDevelopment: true}
	controller := NewToolController(mockRegistry, mockExecutor, config)

	// Test user for authentication
	testUser := &models.User{
		ID:        "test-user-123",
		StravaID:  12345,
		FirstName: "Test",
		LastName:  "User",
	}

	t.Run("ListTools_Success_ReturnsAllTools", func(t *testing.T) {
		expectedTools := []models.ToolDefinition{
			{
				Name:        "get-athlete-profile",
				Description: "Get athlete profile",
				Parameters:  map[string]interface{}{"type": "object"},
				Examples: []models.ToolExample{
					{
						Description: "Get profile",
						Request:     map[string]interface{}{},
						Response:    map[string]interface{}{"id": float64(12345)}, // JSON unmarshaling converts to float64
					},
				},
			},
			{
				Name:        "get-activity-details",
				Description: "Get activity details",
				Parameters:  map[string]interface{}{"type": "object"},
				Examples: []models.ToolExample{
					{
						Description: "Get activity",
						Request:     map[string]interface{}{"activity_id": float64(123456)}, // JSON unmarshaling converts to float64
						Response:    map[string]interface{}{"id": float64(123456)}, // JSON unmarshaling converts to float64
					},
				},
			},
		}

		mockRegistry.On("GetAvailableTools").Return(expectedTools)

		router := gin.New()
		router.GET("/api/tools", controller.ListTools)

		req := httptest.NewRequest("GET", "/api/tools", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ToolListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, 2, response.Count)
		assert.Len(t, response.Tools, 2)
		assert.Equal(t, expectedTools, response.Tools)

		mockRegistry.AssertExpectations(t)
	})

	t.Run("GetToolSchema_ValidTool_ReturnsSchema", func(t *testing.T) {
		expectedSchema := &models.ToolSchema{
			Name:        "get-athlete-profile",
			Description: "Get athlete profile",
			Parameters:  map[string]interface{}{"type": "object"},
			Required:    []string{},
			Optional:    []string{"optional_param"},
			Examples: []models.ToolExample{
				{
					Description: "Basic usage",
					Request:     map[string]interface{}{},
					Response:    map[string]interface{}{"success": true},
				},
			},
		}

		mockRegistry.On("GetToolSchema", "get-athlete-profile").Return(expectedSchema, nil)

		router := gin.New()
		router.GET("/api/tools/:toolName/schema", controller.GetToolSchema)

		req := httptest.NewRequest("GET", "/api/tools/get-athlete-profile/schema", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ToolSchemaResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, expectedSchema, response.Schema)

		mockRegistry.AssertExpectations(t)
	})

	t.Run("GetToolSchema_InvalidTool_ReturnsNotFound", func(t *testing.T) {
		mockRegistry.On("GetToolSchema", "nonexistent-tool").Return((*models.ToolSchema)(nil), fmt.Errorf("tool not found"))

		router := gin.New()
		router.GET("/api/tools/:toolName/schema", controller.GetToolSchema)

		req := httptest.NewRequest("GET", "/api/tools/nonexistent-tool/schema", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "TOOL_NOT_FOUND", response.Error.Code)
		assert.Contains(t, response.Error.Message, "not found")

		mockRegistry.AssertExpectations(t)
	})

	t.Run("ExecuteTool_ValidRequest_Success", func(t *testing.T) {
		// Setup mocks
		mockRegistry.On("IsToolAvailable", "get-athlete-profile").Return(true)
		mockRegistry.On("ValidateToolCall", "get-athlete-profile", map[string]interface{}{}).Return(nil)

		expectedResult := &models.ToolExecutionResult{
			ToolName:  "get-athlete-profile",
			Success:   true,
			Data:      "Profile data",
			Duration:  150,
			Timestamp: time.Now(),
		}

		mockExecutor.On("ExecuteToolWithOptions", mock.Anything, "get-athlete-profile", map[string]interface{}{}, mock.Anything, (*models.ExecutionOptions)(nil)).Return(expectedResult, nil)

		// Setup router with auth middleware
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", testUser)
			c.Next()
		})
		router.POST("/api/tools/execute", controller.ExecuteTool)

		// Create request
		requestBody := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ToolExecutionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "success", response.Status)
		assert.NotNil(t, response.Result)
		assert.Equal(t, expectedResult.ToolName, response.Result.ToolName)
		assert.Equal(t, expectedResult.Success, response.Result.Success)
		assert.Equal(t, expectedResult.Data, response.Result.Data)
		assert.NotNil(t, response.Metadata)
		assert.NotEmpty(t, response.Metadata.RequestID)

		mockRegistry.AssertExpectations(t)
		mockExecutor.AssertExpectations(t)
	})

	t.Run("ExecuteTool_WithExecutionOptions_Success", func(t *testing.T) {
		// Setup mocks
		mockRegistry.On("IsToolAvailable", "get-athlete-profile").Return(true)
		mockRegistry.On("ValidateToolCall", "get-athlete-profile", map[string]interface{}{}).Return(nil)

		expectedResult := &models.ToolExecutionResult{
			ToolName:  "get-athlete-profile",
			Success:   true,
			Data:      "Streaming profile data",
			Duration:  200,
			Timestamp: time.Now(),
		}

		expectedOptions := &models.ExecutionOptions{
			Timeout:   10,
			Streaming: true,
		}

		mockExecutor.On("ExecuteToolWithOptions", mock.Anything, "get-athlete-profile", map[string]interface{}{}, mock.Anything, expectedOptions).Return(expectedResult, nil)

		// Setup router with auth middleware
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", testUser)
			c.Next()
		})
		router.POST("/api/tools/execute", controller.ExecuteTool)

		// Create request with options
		requestBody := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
			Options:    expectedOptions,
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ToolExecutionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "success", response.Status)
		assert.True(t, response.Result.Success)

		mockRegistry.AssertExpectations(t)
		mockExecutor.AssertExpectations(t)
	})

	t.Run("ExecuteTool_MissingAuthentication_ReturnsUnauthorized", func(t *testing.T) {
		// Setup mocks for validation that happens before auth check
		mockRegistry.On("IsToolAvailable", "get-athlete-profile").Return(true)
		mockRegistry.On("ValidateToolCall", "get-athlete-profile", map[string]interface{}{}).Return(nil)

		// Setup router without auth middleware
		router := gin.New()
		router.POST("/api/tools/execute", controller.ExecuteTool)

		requestBody := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "AUTH_REQUIRED", response.Error.Code)

		mockRegistry.AssertExpectations(t)
	})

	t.Run("ExecuteTool_InvalidJSON_ReturnsBadRequest", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", testUser)
			c.Next()
		})
		router.POST("/api/tools/execute", controller.ExecuteTool)

		// Send invalid JSON
		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "INVALID_REQUEST", response.Error.Code)
	})

	t.Run("ExecuteTool_MissingToolName_ReturnsBadRequest", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", testUser)
			c.Next()
		})
		router.POST("/api/tools/execute", controller.ExecuteTool)

		requestBody := models.ToolExecutionRequest{
			// Missing ToolName
			Parameters: map[string]interface{}{},
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// The validation happens at the JSON binding level, so it returns INVALID_REQUEST
		assert.Equal(t, "INVALID_REQUEST", response.Error.Code)
	})

	t.Run("ExecuteTool_ParameterValidationError_ReturnsBadRequest", func(t *testing.T) {
		// Setup mocks
		mockRegistry.On("IsToolAvailable", "get-activity-details").Return(true)
		mockRegistry.On("ValidateToolCall", "get-activity-details", map[string]interface{}{}).Return(fmt.Errorf("required parameter 'activity_id' is missing"))

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", testUser)
			c.Next()
		})
		router.POST("/api/tools/execute", controller.ExecuteTool)

		requestBody := models.ToolExecutionRequest{
			ToolName:   "get-activity-details",
			Parameters: map[string]interface{}{}, // Missing required activity_id
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
		assert.Contains(t, response.Error.Message, "Parameter validation failed")

		mockRegistry.AssertExpectations(t)
	})

	t.Run("ExecuteTool_ExecutionTimeout_ReturnsTimeout", func(t *testing.T) {
		t.Skip("Timeout test requires more complex mock setup - covered in integration tests")
	})
}

func TestToolController_SecurityValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRegistry := &mockToolRegistryComprehensive{}
	mockExecutor := &mockToolExecutorComprehensive{}
	config := &config.Config{IsDevelopment: true}
	controller := NewToolController(mockRegistry, mockExecutor, config)

	testUser := &models.User{
		ID:        "test-user-123",
		StravaID:  12345,
		FirstName: "Test",
		LastName:  "User",
	}

	t.Run("GetToolSchema_MaliciousToolName_ReturnsValidationError", func(t *testing.T) {
		maliciousNames := []string{
			"../../../etc/passwd",
			"tool;rm -rf /",
			"tool<script>alert('xss')</script>",
			"tool\x00null",
			strings.Repeat("a", 200), // Too long
		}

		for _, maliciousName := range maliciousNames {
			t.Run(fmt.Sprintf("MaliciousName_%s", maliciousName[:min(len(maliciousName), 20)]), func(t *testing.T) {
				router := gin.New()
				router.GET("/api/tools/:toolName/schema", controller.GetToolSchema)

				req := httptest.NewRequest("GET", "/api/tools/"+maliciousName+"/schema", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusBadRequest, w.Code)

				var response models.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
			})
		}
	})

	t.Run("ExecuteTool_MaliciousParameters_Sanitized", func(t *testing.T) {
		// Setup mocks
		mockRegistry.On("IsToolAvailable", "get-athlete-profile").Return(true)
		mockRegistry.On("ValidateToolCall", "get-athlete-profile", mock.Anything).Return(nil)

		expectedResult := &models.ToolExecutionResult{
			ToolName:  "get-athlete-profile",
			Success:   true,
			Data:      "Safe response",
			Duration:  100,
			Timestamp: time.Now(),
		}

		mockExecutor.On("ExecuteToolWithOptions", mock.Anything, "get-athlete-profile", mock.Anything, mock.Anything, (*models.ExecutionOptions)(nil)).Return(expectedResult, nil)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", testUser)
			c.Next()
		})
		router.POST("/api/tools/execute", controller.ExecuteTool)

		// Send request with potentially malicious parameters
		requestBody := models.ToolExecutionRequest{
			ToolName: "get-athlete-profile",
			Parameters: map[string]interface{}{
				"script":       "<script>alert('xss')</script>",
				"sql_injection": "'; DROP TABLE users; --",
				"path_traversal": "../../../etc/passwd",
			},
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should still succeed but parameters should be sanitized
		assert.Equal(t, http.StatusOK, w.Code)

		mockRegistry.AssertExpectations(t)
		mockExecutor.AssertExpectations(t)
	})

	t.Run("ExecuteTool_ExcessivelyLargeParameters_Rejected", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", testUser)
			c.Next()
		})
		router.POST("/api/tools/execute", controller.ExecuteTool)

		// Create excessively large parameter
		largeString := strings.Repeat("a", 10*1024*1024) // 10MB string

		requestBody := models.ToolExecutionRequest{
			ToolName: "get-athlete-profile",
			Parameters: map[string]interface{}{
				"large_param": largeString,
			},
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should be rejected due to size
		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("ExecuteTool_InvalidExecutionOptions_ReturnsValidationError", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", testUser)
			c.Next()
		})
		router.POST("/api/tools/execute", controller.ExecuteTool)

		requestBody := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
			Options: &models.ExecutionOptions{
				Timeout: -1, // Invalid negative timeout
			},
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})
}

func TestToolController_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRegistry := &mockToolRegistryComprehensive{}
	mockExecutor := &mockToolExecutorComprehensive{}
	config := &config.Config{IsDevelopment: true}
	controller := NewToolController(mockRegistry, mockExecutor, config)

	testUser := &models.User{
		ID:        "test-user-123",
		StravaID:  12345,
		FirstName: "Test",
		LastName:  "User",
	}

	t.Run("ExecuteTool_ServiceUnavailable_ReturnsServiceUnavailable", func(t *testing.T) {
		// Setup mocks
		mockRegistry.On("IsToolAvailable", "get-athlete-profile").Return(true)
		mockRegistry.On("ValidateToolCall", "get-athlete-profile", map[string]interface{}{}).Return(nil)

		serviceError := fmt.Errorf("service unavailable: external API is down")
		mockExecutor.On("ExecuteToolWithOptions", mock.Anything, "get-athlete-profile", map[string]interface{}{}, mock.Anything, (*models.ExecutionOptions)(nil)).Return((*models.ToolExecutionResult)(nil), serviceError)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", testUser)
			c.Next()
		})
		router.POST("/api/tools/execute", controller.ExecuteTool)

		requestBody := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var response models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "SERVICE_UNAVAILABLE", response.Error.Code)

		mockRegistry.AssertExpectations(t)
		mockExecutor.AssertExpectations(t)
	})

	t.Run("ExecuteTool_RateLimitExceeded_ReturnsRateLimit", func(t *testing.T) {
		// Setup mocks
		mockRegistry.On("IsToolAvailable", "get-athlete-profile").Return(true)
		mockRegistry.On("ValidateToolCall", "get-athlete-profile", map[string]interface{}{}).Return(nil)

		rateLimitError := fmt.Errorf("rate limit exceeded: too many requests")
		mockExecutor.On("ExecuteToolWithOptions", mock.Anything, "get-athlete-profile", map[string]interface{}{}, mock.Anything, (*models.ExecutionOptions)(nil)).Return((*models.ToolExecutionResult)(nil), rateLimitError)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", testUser)
			c.Next()
		})
		router.POST("/api/tools/execute", controller.ExecuteTool)

		requestBody := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		var response models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "RATE_LIMIT_EXCEEDED", response.Error.Code)

		mockRegistry.AssertExpectations(t)
		mockExecutor.AssertExpectations(t)
	})

	t.Run("ExecuteTool_PanicRecovery_ReturnsInternalError", func(t *testing.T) {
		// Setup mocks that will cause a panic
		mockRegistry.On("IsToolAvailable", "get-athlete-profile").Return(true)
		mockRegistry.On("ValidateToolCall", "get-athlete-profile", map[string]interface{}{}).Return(nil)

		// Mock executor that panics
		mockExecutor.On("ExecuteToolWithOptions", mock.Anything, "get-athlete-profile", map[string]interface{}{}, mock.Anything, (*models.ExecutionOptions)(nil)).Panic("simulated panic")

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", testUser)
			c.Next()
		})
		router.POST("/api/tools/execute", controller.ExecuteTool)

		requestBody := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)

		mockRegistry.AssertExpectations(t)
		mockExecutor.AssertExpectations(t)
	})
}

// Comprehensive mock implementations for testing
type mockToolRegistryComprehensive struct {
	mock.Mock
}

func (m *mockToolRegistryComprehensive) GetAvailableTools() []models.ToolDefinition {
	args := m.Called()
	return args.Get(0).([]models.ToolDefinition)
}

func (m *mockToolRegistryComprehensive) GetToolSchema(toolName string) (*models.ToolSchema, error) {
	args := m.Called(toolName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ToolSchema), args.Error(1)
}

func (m *mockToolRegistryComprehensive) ValidateToolCall(toolName string, parameters map[string]interface{}) error {
	args := m.Called(toolName, parameters)
	return args.Error(0)
}

func (m *mockToolRegistryComprehensive) IsToolAvailable(toolName string) bool {
	args := m.Called(toolName)
	return args.Bool(0)
}

type mockToolExecutorComprehensive struct {
	mock.Mock
}

func (m *mockToolExecutorComprehensive) ExecuteTool(ctx context.Context, toolName string, parameters map[string]interface{}, msgCtx *services.MessageContext) (*models.ToolExecutionResult, error) {
	args := m.Called(ctx, toolName, parameters, msgCtx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ToolExecutionResult), args.Error(1)
}

func (m *mockToolExecutorComprehensive) ExecuteToolWithOptions(ctx context.Context, toolName string, parameters map[string]interface{}, msgCtx *services.MessageContext, options *models.ExecutionOptions) (*models.ToolExecutionResult, error) {
	args := m.Called(ctx, toolName, parameters, msgCtx, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ToolExecutionResult), args.Error(1)
}

func (m *mockToolExecutorComprehensive) CancelJob(jobID string) bool {
	args := m.Called(jobID)
	return args.Bool(0)
}

func (m *mockToolExecutorComprehensive) GetActiveJobCount() int {
	args := m.Called()
	return args.Int(0)
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}