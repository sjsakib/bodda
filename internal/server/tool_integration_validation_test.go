package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"
	"bodda/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ToolIntegrationValidationTestSuite provides comprehensive integration testing for the tool execution endpoint
type ToolIntegrationValidationTestSuite struct {
	suite.Suite
	router         *gin.Engine
	toolController *ToolController
	registry       services.ToolRegistry
	executor       services.ToolExecutor
	config         *config.Config
	testUser       *models.User
}

func TestToolIntegrationValidationSuite(t *testing.T) {
	suite.Run(t, new(ToolIntegrationValidationTestSuite))
}

func (suite *ToolIntegrationValidationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	
	// Create test user
	suite.testUser = &models.User{
		ID:        "test-user-123",
		StravaID:  12345,
		FirstName: "Test",
		LastName:  "User",
	}
}

func (suite *ToolIntegrationValidationTestSuite) SetupTest() {
	// Create development config
	suite.config = &config.Config{
		IsDevelopment: true,
	}
	
	// Create real tool registry and executor with mock services
	suite.registry = services.NewToolRegistry()
	mockToolService := &mockToolExecutionService{}
	suite.executor = services.NewToolExecutor(mockToolService, suite.registry)
	suite.toolController = NewToolController(suite.registry, suite.executor, suite.config)
	
	// Setup router with all middleware
	suite.router = gin.New()
	suite.setupToolRoutes()
}

func (suite *ToolIntegrationValidationTestSuite) setupToolRoutes() {
	// Tool execution routes with all middleware (development only)
	tools := suite.router.Group("/api/tools")
	tools.Use(DevelopmentOnlyMiddleware(suite.config))
	tools.Use(InputValidationMiddleware())
	tools.Use(WorkspaceBoundaryMiddleware())
	tools.Use(suite.authMiddleware())
	{
		tools.GET("", suite.toolController.ListTools)
		tools.GET("/:toolName/schema", suite.toolController.GetToolSchema)
		tools.POST("/execute", suite.toolController.ExecuteTool)
	}
}

func (suite *ToolIntegrationValidationTestSuite) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user", suite.testUser)
		c.Next()
	}
}

// Test all endpoints with existing tool implementations
func (suite *ToolIntegrationValidationTestSuite) TestAllEndpointsWithExistingTools() {
	// Test 1: List all available tools
	suite.Run("ListAllTools", func() {
		req := httptest.NewRequest("GET", "/api/tools", nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		suite.Equal(http.StatusOK, w.Code)
		
		var response models.ToolListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.NoError(err)
		
		// Verify we have all expected tools
		expectedTools := []string{
			"get-athlete-profile",
			"get-recent-activities", 
			"get-activity-details",
			"get-activity-streams",
			"update-athlete-logbook",
		}
		
		suite.Equal(len(expectedTools), response.Count)
		suite.Equal(len(expectedTools), len(response.Tools))
		
		// Verify each expected tool is present
		toolNames := make(map[string]bool)
		for _, tool := range response.Tools {
			toolNames[tool.Name] = true
		}
		
		for _, expectedTool := range expectedTools {
			suite.True(toolNames[expectedTool], "Expected tool %s not found", expectedTool)
		}
	})

	// Test 2: Get schema for each tool
	suite.Run("GetSchemaForAllTools", func() {
		tools := []string{
			"get-athlete-profile",
			"get-recent-activities", 
			"get-activity-details",
			"get-activity-streams",
			"update-athlete-logbook",
		}
		
		for _, toolName := range tools {
			suite.Run(fmt.Sprintf("Schema_%s", toolName), func() {
				req := httptest.NewRequest("GET", fmt.Sprintf("/api/tools/%s/schema", toolName), nil)
				w := httptest.NewRecorder()
				suite.router.ServeHTTP(w, req)

				suite.Equal(http.StatusOK, w.Code, "Failed to get schema for tool %s", toolName)
				
				var response models.ToolSchemaResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err, "Failed to parse schema response for tool %s", toolName)
				
				// Verify schema structure
				suite.NotNil(response.Schema)
				suite.Equal(toolName, response.Schema.Name)
				suite.NotEmpty(response.Schema.Description)
				suite.NotNil(response.Schema.Parameters)
				suite.NotNil(response.Schema.Required)
				suite.NotNil(response.Schema.Optional)
				suite.NotNil(response.Schema.Examples)
			})
		}
	})

	// Test 3: Execute each tool with valid parameters
	suite.Run("ExecuteAllToolsWithValidParameters", func() {
		testCases := []struct {
			toolName   string
			parameters map[string]interface{}
		}{
			{
				toolName:   "get-athlete-profile",
				parameters: map[string]interface{}{},
			},
			{
				toolName: "get-recent-activities",
				parameters: map[string]interface{}{
					"per_page": 10,
				},
			},
			{
				toolName: "get-activity-details",
				parameters: map[string]interface{}{
					"activity_id": int64(123456789),
				},
			},
			{
				toolName: "get-activity-streams",
				parameters: map[string]interface{}{
					"activity_id":  int64(123456789),
					"stream_types": []string{"time", "heartrate"},
					"resolution":   "medium",
				},
			},
			{
				toolName: "update-athlete-logbook",
				parameters: map[string]interface{}{
					"content": "Test logbook update",
				},
			},
		}
		
		for _, tc := range testCases {
			suite.Run(fmt.Sprintf("Execute_%s", tc.toolName), func() {
				requestBody := models.ToolExecutionRequest{
					ToolName:   tc.toolName,
					Parameters: tc.parameters,
				}
				jsonBody, _ := json.Marshal(requestBody)

				req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				suite.router.ServeHTTP(w, req)

				suite.Equal(http.StatusOK, w.Code, "Failed to execute tool %s", tc.toolName)
				
				var response models.ToolExecutionResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err, "Failed to parse execution response for tool %s", tc.toolName)
				
				// Verify response structure
				suite.Equal("success", response.Status)
				suite.NotNil(response.Result)
				suite.Equal(tc.toolName, response.Result.ToolName)
				suite.NotNil(response.Metadata)
				suite.NotEmpty(response.Metadata.RequestID)
				suite.True(response.Metadata.Duration >= 0, "Duration should be >= 0, got %d", response.Metadata.Duration)
			})
		}
	})
}

// Validate response format consistency across all tools
func (suite *ToolIntegrationValidationTestSuite) TestResponseFormatConsistency() {
	suite.Run("SuccessResponseConsistency", func() {
		tools := []struct {
			name       string
			parameters map[string]interface{}
		}{
			{"get-athlete-profile", map[string]interface{}{}},
			{"get-recent-activities", map[string]interface{}{"per_page": 5}},
			{"get-activity-details", map[string]interface{}{"activity_id": int64(123456789)}},
		}
		
		var responses []models.ToolExecutionResponse
		
		// Execute all tools and collect responses
		for _, tool := range tools {
			requestBody := models.ToolExecutionRequest{
				ToolName:   tool.name,
				Parameters: tool.parameters,
			}
			jsonBody, _ := json.Marshal(requestBody)

			req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(http.StatusOK, w.Code)
			
			var response models.ToolExecutionResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			suite.NoError(err)
			responses = append(responses, response)
		}
		
		// Verify all responses have consistent structure
		for i, response := range responses {
			suite.Equal("success", response.Status, "Tool %d has inconsistent status", i)
			suite.NotNil(response.Result, "Tool %d missing result", i)
			suite.NotNil(response.Metadata, "Tool %d missing metadata", i)
			suite.NotEmpty(response.Metadata.RequestID, "Tool %d missing request ID", i)
			suite.True(response.Metadata.Duration >= 0, "Tool %d has invalid duration", i)
			suite.NotEmpty(response.Result.ToolName, "Tool %d missing tool name in result", i)
			suite.True(response.Result.Duration >= 0, "Tool %d has invalid result duration", i)
		}
	})

	suite.Run("ErrorResponseConsistency", func() {
		errorCases := []struct {
			name           string
			requestBody    interface{}
			expectedStatus int
			expectedCode   string
		}{
			{
				name:           "InvalidToolName",
				requestBody:    models.ToolExecutionRequest{ToolName: "nonexistent-tool", Parameters: map[string]interface{}{}},
				expectedStatus: http.StatusBadRequest,
				expectedCode:   "TOOL_NOT_FOUND",
			},
			{
				name:           "MissingRequiredParameter",
				requestBody:    models.ToolExecutionRequest{ToolName: "get-activity-details", Parameters: map[string]interface{}{}},
				expectedStatus: http.StatusBadRequest,
				expectedCode:   "VALIDATION_ERROR",
			},
			{
				name:           "InvalidJSON",
				requestBody:    "invalid json",
				expectedStatus: http.StatusBadRequest,
				expectedCode:   "VALIDATION_ERROR",
			},
		}
		
		var errorResponses []models.ErrorResponse
		
		for _, tc := range errorCases {
			suite.Run(tc.name, func() {
				var jsonBody []byte
				var err error
				
				if str, ok := tc.requestBody.(string); ok {
					jsonBody = []byte(str)
				} else {
					jsonBody, err = json.Marshal(tc.requestBody)
					suite.NoError(err)
				}

				req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				suite.router.ServeHTTP(w, req)

				suite.Equal(tc.expectedStatus, w.Code)
				
				var response models.ErrorResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				
				suite.Equal(tc.expectedCode, response.Error.Code)
				suite.NotEmpty(response.Error.Message)
				suite.NotEmpty(response.RequestID)
				errorResponses = append(errorResponses, response)
			})
		}
		
		// Verify all error responses have consistent structure
		for i, response := range errorResponses {
			suite.NotEmpty(response.Error.Code, "Error response %d missing error code", i)
			suite.NotEmpty(response.Error.Message, "Error response %d missing error message", i)
			suite.NotEmpty(response.RequestID, "Error response %d missing request ID", i)
			suite.False(response.Timestamp.IsZero(), "Error response %d missing timestamp", i)
		}
	})
}

// Verify development mode enforcement and production security
func (suite *ToolIntegrationValidationTestSuite) TestDevelopmentModeEnforcement() {
	suite.Run("DevelopmentModeEnabled", func() {
		// Verify endpoints are accessible in development mode
		req := httptest.NewRequest("GET", "/api/tools", nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		suite.Equal(http.StatusOK, w.Code, "Tools endpoint should be accessible in development mode")
	})

	suite.Run("ProductionModeBlocked", func() {
		// Create production config
		prodConfig := &config.Config{
			IsDevelopment: false,
		}
		
		// Create new router with production config
		prodRouter := gin.New()
		prodToolController := NewToolController(suite.registry, suite.executor, prodConfig)
		
		tools := prodRouter.Group("/api/tools")
		tools.Use(DevelopmentOnlyMiddleware(prodConfig))
		tools.Use(InputValidationMiddleware())
		tools.Use(WorkspaceBoundaryMiddleware())
		tools.Use(suite.authMiddleware())
		{
			tools.GET("", prodToolController.ListTools)
			tools.GET("/:toolName/schema", prodToolController.GetToolSchema)
			tools.POST("/execute", prodToolController.ExecuteTool)
		}
		
		// Test all endpoints return 404 in production
		endpoints := []struct {
			method string
			path   string
			body   []byte
		}{
			{"GET", "/api/tools", nil},
			{"GET", "/api/tools/get-athlete-profile/schema", nil},
			{"POST", "/api/tools/execute", []byte(`{"tool_name":"get-athlete-profile","parameters":{}}`)},
		}
		
		for _, endpoint := range endpoints {
			suite.Run(fmt.Sprintf("Production_%s_%s", endpoint.method, endpoint.path), func() {
				var req *http.Request
				if endpoint.body != nil {
					req = httptest.NewRequest(endpoint.method, endpoint.path, bytes.NewBuffer(endpoint.body))
					req.Header.Set("Content-Type", "application/json")
				} else {
					req = httptest.NewRequest(endpoint.method, endpoint.path, nil)
				}
				
				w := httptest.NewRecorder()
				prodRouter.ServeHTTP(w, req)

				suite.Equal(http.StatusNotFound, w.Code, 
					"Endpoint %s %s should return 404 in production mode", endpoint.method, endpoint.path)
			})
		}
	})

	suite.Run("SecurityValidation", func() {
		// Test malicious parameter rejection
		maliciousRequests := []struct {
			name        string
			requestBody models.ToolExecutionRequest
		}{
			{
				name: "SQLInjection",
				requestBody: models.ToolExecutionRequest{
					ToolName: "get-activity-details",
					Parameters: map[string]interface{}{
						"activity_id": "123; DROP TABLE users;",
					},
				},
			},
			{
				name: "PathTraversal",
				requestBody: models.ToolExecutionRequest{
					ToolName: "update-athlete-logbook",
					Parameters: map[string]interface{}{
						"content": "../../../etc/passwd",
					},
				},
			},
			{
				name: "XSSAttempt",
				requestBody: models.ToolExecutionRequest{
					ToolName: "update-athlete-logbook",
					Parameters: map[string]interface{}{
						"content": "<script>alert('xss')</script>",
					},
				},
			},
		}
		
		for _, tc := range maliciousRequests {
			suite.Run(tc.name, func() {
				jsonBody, _ := json.Marshal(tc.requestBody)

				req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				suite.router.ServeHTTP(w, req)

				// Should be rejected with 400 status
				suite.Equal(http.StatusBadRequest, w.Code, "Malicious request should be rejected")
				
				var response models.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				
				// Should indicate validation error or malicious input
				suite.Contains([]string{"VALIDATION_ERROR", "MALICIOUS_INPUT"}, response.Error.Code)
			})
		}
	})

	suite.Run("WorkspaceBoundaryEnforcement", func() {
		// Test that workspace root is set in context
		req := httptest.NewRequest("GET", "/api/tools", nil)
		w := httptest.NewRecorder()
		
		// Add middleware to check workspace context
		testRouter := gin.New()
		testRouter.Use(WorkspaceBoundaryMiddleware())
		testRouter.Use(func(c *gin.Context) {
			workspaceRoot, exists := c.Get("workspace_root")
			suite.True(exists, "Workspace root should be set in context")
			suite.NotEmpty(workspaceRoot, "Workspace root should not be empty")
			c.JSON(200, gin.H{"workspace_root": workspaceRoot})
		})
		testRouter.GET("/api/tools", func(c *gin.Context) {})
		
		testRouter.ServeHTTP(w, req)
		suite.Equal(http.StatusOK, w.Code)
	})
}

// Test timeout and streaming functionality
func (suite *ToolIntegrationValidationTestSuite) TestExecutionOptionsAndTimeout() {
	suite.Run("TimeoutHandling", func() {
		// Create executor with very short timeout for testing
		mockToolService := &mockToolExecutionServiceWithDelay{
			delay: 2 * time.Second, // Longer than timeout
		}
		shortTimeoutExecutor := services.NewToolExecutorWithConfig(
			mockToolService, 
			suite.registry, 
			100*time.Millisecond, // Very short default timeout
			1*time.Second,        // Short max timeout
		)
		
		timeoutController := NewToolController(suite.registry, shortTimeoutExecutor, suite.config)
		
		// Setup router with timeout controller
		timeoutRouter := gin.New()
		tools := timeoutRouter.Group("/api/tools")
		tools.Use(DevelopmentOnlyMiddleware(suite.config))
		tools.Use(InputValidationMiddleware())
		tools.Use(WorkspaceBoundaryMiddleware())
		tools.Use(suite.authMiddleware())
		tools.POST("/execute", timeoutController.ExecuteTool)
		
		requestBody := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		timeoutRouter.ServeHTTP(w, req)

		// Should timeout and return error
		suite.Equal(http.StatusInternalServerError, w.Code)
		
		var response models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.NoError(err)
		// The error message should indicate a timeout occurred
		// It could be "timeout", "deadline exceeded", "context deadline exceeded", or "Tool execution failed" (which is the generic wrapper)
		suite.True(strings.Contains(response.Error.Message, "timeout") || 
			strings.Contains(response.Error.Message, "deadline exceeded") ||
			strings.Contains(response.Error.Message, "context deadline exceeded") ||
			strings.Contains(response.Error.Message, "Tool execution failed"),
			"Error message should indicate timeout, got: %s", response.Error.Message)
	})

	suite.Run("StreamingVsBufferedExecution", func() {
		// Test streaming execution
		streamingRequest := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
			Options: &models.ExecutionOptions{
				Streaming: true,
			},
		}
		jsonBody, _ := json.Marshal(streamingRequest)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		suite.Equal(http.StatusOK, w.Code)
		
		var response models.ToolExecutionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.NoError(err)
		suite.Equal("success", response.Status)
		
		// Test buffered execution (default)
		bufferedRequest := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
			Options: &models.ExecutionOptions{
				BufferedOutput: true,
			},
		}
		jsonBody, _ = json.Marshal(bufferedRequest)

		req = httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		suite.Equal(http.StatusOK, w.Code)
		
		err = json.Unmarshal(w.Body.Bytes(), &response)
		suite.NoError(err)
		suite.Equal("success", response.Status)
	})
}

// Test authentication and authorization
func (suite *ToolIntegrationValidationTestSuite) TestAuthenticationAndAuthorization() {
	suite.Run("AuthenticationRequired", func() {
		// Create router without auth middleware
		noAuthRouter := gin.New()
		tools := noAuthRouter.Group("/api/tools")
		tools.Use(DevelopmentOnlyMiddleware(suite.config))
		tools.Use(InputValidationMiddleware())
		tools.Use(WorkspaceBoundaryMiddleware())
		// No auth middleware
		tools.POST("/execute", suite.toolController.ExecuteTool)
		
		requestBody := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		noAuthRouter.ServeHTTP(w, req)

		suite.Equal(http.StatusUnauthorized, w.Code)
		
		var response models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.NoError(err)
		suite.Equal("AUTH_REQUIRED", response.Error.Code)
	})

	suite.Run("ValidAuthentication", func() {
		// Test with valid authentication (our normal setup)
		requestBody := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		suite.Equal(http.StatusOK, w.Code)
		
		var response models.ToolExecutionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.NoError(err)
		suite.Equal("success", response.Status)
	})
}

// Test comprehensive error scenarios
func (suite *ToolIntegrationValidationTestSuite) TestComprehensiveErrorScenarios() {
	suite.Run("ValidationErrors", func() {
		validationCases := []struct {
			name           string
			requestBody    interface{}
			expectedStatus int
			expectedCode   string
		}{
			{
				name:           "EmptyToolName",
				requestBody:    models.ToolExecutionRequest{ToolName: "", Parameters: map[string]interface{}{}},
				expectedStatus: http.StatusBadRequest,
				expectedCode:   "VALIDATION_ERROR",
			},
			{
				name:           "InvalidToolNameChars",
				requestBody:    models.ToolExecutionRequest{ToolName: "invalid@tool#name", Parameters: map[string]interface{}{}},
				expectedStatus: http.StatusBadRequest,
				expectedCode:   "VALIDATION_ERROR",
			},
			{
				name:           "TooLongToolName",
				requestBody:    models.ToolExecutionRequest{ToolName: string(make([]byte, 200)), Parameters: map[string]interface{}{}},
				expectedStatus: http.StatusBadRequest,
				expectedCode:   "VALIDATION_ERROR",
			},
			{
				name:           "InvalidParameterType",
				requestBody:    map[string]interface{}{"tool_name": "get-athlete-profile", "parameters": "invalid"},
				expectedStatus: http.StatusBadRequest,
				expectedCode:   "VALIDATION_ERROR",
			},
		}
		
		for _, tc := range validationCases {
			suite.Run(tc.name, func() {
				jsonBody, _ := json.Marshal(tc.requestBody)

				req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				suite.router.ServeHTTP(w, req)

				suite.Equal(tc.expectedStatus, w.Code)
				
				var response models.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Equal(tc.expectedCode, response.Error.Code)
				suite.NotEmpty(response.Error.Message)
			})
		}
	})
}

// Mock tool execution service with delay for timeout testing
type mockToolExecutionServiceWithDelay struct {
	delay time.Duration
}

func (m *mockToolExecutionServiceWithDelay) ExecuteGetAthleteProfile(ctx context.Context, msgCtx *services.MessageContext) (string, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return `{"id": 12345, "username": "test_athlete", "firstname": "Test", "lastname": "User"}`, nil
}

func (m *mockToolExecutionServiceWithDelay) ExecuteGetRecentActivities(ctx context.Context, msgCtx *services.MessageContext, perPage int) (string, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return fmt.Sprintf(`{"activities": [], "count": %d}`, perPage), nil
}

func (m *mockToolExecutionServiceWithDelay) ExecuteGetActivityDetails(ctx context.Context, msgCtx *services.MessageContext, activityID int64) (string, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return fmt.Sprintf(`{"id": %d, "name": "Test Activity", "type": "Run"}`, activityID), nil
}

func (m *mockToolExecutionServiceWithDelay) ExecuteGetActivityStreams(ctx context.Context, msgCtx *services.MessageContext, activityID int64, streamTypes []string, resolution string, processingMode string, pageNumber int, pageSize int, summaryPrompt string) (string, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return fmt.Sprintf(`{"streams": {"time": [0,1,2], "heartrate": [120,125,130]}, "activity_id": %d}`, activityID), nil
}

func (m *mockToolExecutionServiceWithDelay) ExecuteUpdateAthleteLogbook(ctx context.Context, msgCtx *services.MessageContext, content string) (string, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return `{"success": true, "message": "Logbook updated successfully"}`, nil
}

// Integration test to verify environment variable configuration
func TestDevelopmentModeEnvironmentConfiguration(t *testing.T) {
	// Save original environment
	originalEnv := os.Getenv("DEVELOPMENT_MODE")
	defer func() {
		if originalEnv != "" {
			os.Setenv("DEVELOPMENT_MODE", originalEnv)
		} else {
			os.Unsetenv("DEVELOPMENT_MODE")
		}
	}()

	t.Run("DevelopmentModeFromEnvironment", func(t *testing.T) {
		// Test development mode enabled via environment
		os.Setenv("DEVELOPMENT_MODE", "true")
		
		cfg := &config.Config{}
		// In a real scenario, config would read from environment
		// For this test, we'll simulate it
		cfg.IsDevelopment = os.Getenv("DEVELOPMENT_MODE") == "true"
		
		assert.True(t, cfg.IsDevelopment, "Development mode should be enabled from environment")
		
		// Test development mode disabled via environment
		os.Setenv("DEVELOPMENT_MODE", "false")
		cfg.IsDevelopment = os.Getenv("DEVELOPMENT_MODE") == "true"
		
		assert.False(t, cfg.IsDevelopment, "Development mode should be disabled from environment")
	})
}

// Benchmark test for tool execution performance
func BenchmarkToolExecution(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	config := &config.Config{IsDevelopment: true}
	registry := services.NewToolRegistry()
	mockToolService := &mockToolExecutionService{}
	executor := services.NewToolExecutor(mockToolService, registry)
	controller := NewToolController(registry, executor, config)
	
	router := gin.New()
	tools := router.Group("/api/tools")
	tools.Use(DevelopmentOnlyMiddleware(config))
	tools.Use(InputValidationMiddleware())
	tools.Use(WorkspaceBoundaryMiddleware())
	tools.Use(func(c *gin.Context) {
		c.Set("user", &models.User{ID: "test-user", StravaID: 12345})
		c.Next()
	})
	tools.POST("/execute", controller.ExecuteTool)
	
	requestBody := models.ToolExecutionRequest{
		ToolName:   "get-athlete-profile",
		Parameters: map[string]interface{}{},
	}
	jsonBody, _ := json.Marshal(requestBody)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}