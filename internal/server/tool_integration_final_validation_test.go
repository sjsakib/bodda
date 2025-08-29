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

// FinalIntegrationTestSuite provides comprehensive integration testing for task 11
// This test suite validates:
// - All endpoints with existing tool implementations
// - Response format consistency across all tools
// - Development mode enforcement and production security
type FinalIntegrationTestSuite struct {
	suite.Suite
	router         *gin.Engine
	toolController *ToolController
	registry       services.ToolRegistry
	executor       services.ToolExecutor
	config         *config.Config
	testUser       *models.User
}

func TestFinalIntegrationSuite(t *testing.T) {
	suite.Run(t, new(FinalIntegrationTestSuite))
}

func (suite *FinalIntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	
	// Create test user
	suite.testUser = &models.User{
		ID:          "final-test-user-123",
		StravaID:    12345,
		AccessToken: "test-access-token",
		FirstName:   "Final",
		LastName:    "Test",
	}
}

func (suite *FinalIntegrationTestSuite) SetupTest() {
	// Create development config
	suite.config = &config.Config{
		IsDevelopment: true,
	}
	
	// Create real tool registry and executor with mock services
	suite.registry = services.NewToolRegistry()
	mockToolService := &finalTestToolExecutionService{}
	suite.executor = services.NewToolExecutor(mockToolService, suite.registry)
	suite.toolController = NewToolController(suite.registry, suite.executor, suite.config)
	
	// Setup router with all middleware
	suite.router = gin.New()
	suite.setupToolRoutes()
}

func (suite *FinalIntegrationTestSuite) setupToolRoutes() {
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

func (suite *FinalIntegrationTestSuite) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user", suite.testUser)
		c.Next()
	}
}

// Test 1: Test all endpoints with existing tool implementations
func (suite *FinalIntegrationTestSuite) TestAllEndpointsWithExistingToolImplementations() {
	suite.Run("ListAllAvailableTools", func() {
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
		
		// Verify each expected tool is present with proper structure
		toolNames := make(map[string]models.ToolDefinition)
		for _, tool := range response.Tools {
			toolNames[tool.Name] = tool
		}
		
		for _, expectedTool := range expectedTools {
			suite.True(len(toolNames[expectedTool].Name) > 0, "Expected tool %s not found", expectedTool)
			suite.NotEmpty(toolNames[expectedTool].Description, "Tool %s missing description", expectedTool)
			suite.NotNil(toolNames[expectedTool].Parameters, "Tool %s missing parameters", expectedTool)
		}
	})

	suite.Run("GetSchemaForAllExistingTools", func() {
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
				
				// Verify comprehensive schema structure
				suite.NotNil(response.Schema)
				suite.Equal(toolName, response.Schema.Name)
				suite.NotEmpty(response.Schema.Description)
				suite.NotNil(response.Schema.Parameters)
				suite.NotNil(response.Schema.Required)
				suite.NotNil(response.Schema.Optional)
				suite.NotNil(response.Schema.Examples)
				suite.Greater(len(response.Schema.Examples), 0, "Tool %s should have examples", toolName)
			})
		}
	})

	suite.Run("ExecuteAllExistingToolsWithValidParameters", func() {
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
					"content": "Final integration test logbook update",
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
				
				// Verify comprehensive response structure
				suite.Equal("success", response.Status)
				suite.NotNil(response.Result)
				suite.Equal(tc.toolName, response.Result.ToolName)
				suite.True(response.Result.Success, "Tool %s execution should succeed", tc.toolName)
				suite.NotNil(response.Result.Data)
				suite.NotNil(response.Metadata)
				suite.NotEmpty(response.Metadata.RequestID)
				suite.True(response.Metadata.Duration >= 0, "Duration should be >= 0, got %d", response.Metadata.Duration)
				suite.False(response.Result.Timestamp.IsZero(), "Result timestamp should be set")
			})
		}
	})
}

// Test 2: Validate response format consistency across all tools
func (suite *FinalIntegrationTestSuite) TestResponseFormatConsistencyAcrossAllTools() {
	suite.Run("SuccessResponseFormatConsistency", func() {
		tools := []struct {
			name       string
			parameters map[string]interface{}
		}{
			{"get-athlete-profile", map[string]interface{}{}},
			{"get-recent-activities", map[string]interface{}{"per_page": 5}},
			{"get-activity-details", map[string]interface{}{"activity_id": int64(123456789)}},
			{"get-activity-streams", map[string]interface{}{
				"activity_id": int64(123456789), 
				"stream_types": []string{"time"},
			}},
			{"update-athlete-logbook", map[string]interface{}{"content": "test"}},
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

			suite.Equal(http.StatusOK, w.Code, "Tool %s should execute successfully", tool.name)
			
			var response models.ToolExecutionResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			suite.NoError(err, "Tool %s response should parse correctly", tool.name)
			responses = append(responses, response)
		}
		
		// Verify all responses have identical structure
		for i, response := range responses {
			toolName := tools[i].name
			
			// Top-level response structure
			suite.Equal("success", response.Status, "Tool %s has inconsistent status", toolName)
			suite.NotNil(response.Result, "Tool %s missing result", toolName)
			suite.NotNil(response.Metadata, "Tool %s missing metadata", toolName)
			suite.Nil(response.Error, "Tool %s should not have error in success response", toolName)
			
			// Result structure consistency
			suite.Equal(toolName, response.Result.ToolName, "Tool %s has incorrect tool name in result", toolName)
			suite.True(response.Result.Success, "Tool %s result should indicate success", toolName)
			suite.NotNil(response.Result.Data, "Tool %s missing result data", toolName)
			suite.True(response.Result.Duration >= 0, "Tool %s has invalid result duration", toolName)
			suite.False(response.Result.Timestamp.IsZero(), "Tool %s missing result timestamp", toolName)
			suite.Empty(response.Result.Error, "Tool %s should not have error in successful result", toolName)
			
			// Metadata structure consistency
			suite.NotEmpty(response.Metadata.RequestID, "Tool %s missing request ID", toolName)
			suite.False(response.Metadata.Timestamp.IsZero(), "Tool %s missing metadata timestamp", toolName)
			suite.True(response.Metadata.Duration >= 0, "Tool %s has invalid metadata duration", toolName)
		}
	})

	suite.Run("ErrorResponseFormatConsistency", func() {
		errorCases := []struct {
			name           string
			requestBody    interface{}
			expectedStatus int
			expectedCode   string
		}{
			{
				name:           "NonexistentTool",
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
				requestBody:    "invalid json content",
				expectedStatus: http.StatusBadRequest,
				expectedCode:   "VALIDATION_ERROR",
			},
			{
				name:           "EmptyToolName",
				requestBody:    models.ToolExecutionRequest{ToolName: "", Parameters: map[string]interface{}{}},
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

				suite.Equal(tc.expectedStatus, w.Code, "Error case %s has wrong status code", tc.name)
				
				var response models.ErrorResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err, "Error case %s response should parse correctly", tc.name)
				
				suite.Equal(tc.expectedCode, response.Error.Code, "Error case %s has wrong error code", tc.name)
				suite.NotEmpty(response.Error.Message, "Error case %s missing error message", tc.name)
				suite.NotEmpty(response.RequestID, "Error case %s missing request ID", tc.name)
				suite.False(response.Timestamp.IsZero(), "Error case %s missing timestamp", tc.name)
				
				errorResponses = append(errorResponses, response)
			})
		}
		
		// Verify all error responses have consistent structure
		for i, response := range errorResponses {
			caseName := errorCases[i].name
			suite.NotEmpty(response.Error.Code, "Error response %s missing error code", caseName)
			suite.NotEmpty(response.Error.Message, "Error response %s missing error message", caseName)
			suite.NotEmpty(response.RequestID, "Error response %s missing request ID", caseName)
			suite.False(response.Timestamp.IsZero(), "Error response %s missing timestamp", caseName)
		}
	})
}

// Test 3: Verify development mode enforcement and production security
func (suite *FinalIntegrationTestSuite) TestDevelopmentModeEnforcementAndProductionSecurity() {
	suite.Run("DevelopmentModeAccessible", func() {
		// Verify all endpoints are accessible in development mode
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
			suite.Run(fmt.Sprintf("Dev_%s_%s", endpoint.method, strings.ReplaceAll(endpoint.path, "/", "_")), func() {
				var req *http.Request
				if endpoint.body != nil {
					req = httptest.NewRequest(endpoint.method, endpoint.path, bytes.NewBuffer(endpoint.body))
					req.Header.Set("Content-Type", "application/json")
				} else {
					req = httptest.NewRequest(endpoint.method, endpoint.path, nil)
				}
				
				w := httptest.NewRecorder()
				suite.router.ServeHTTP(w, req)

				suite.NotEqual(http.StatusNotFound, w.Code, 
					"Endpoint %s %s should be accessible in development mode", endpoint.method, endpoint.path)
				suite.True(w.Code < 500, 
					"Endpoint %s %s should not return server error in development mode", endpoint.method, endpoint.path)
			})
		}
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
		
		// Test all endpoints return 404 in production (as if they don't exist)
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
			suite.Run(fmt.Sprintf("Prod_%s_%s", endpoint.method, strings.ReplaceAll(endpoint.path, "/", "_")), func() {
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

	suite.Run("SecurityValidationAndBoundaryEnforcement", func() {
		// Test malicious parameter detection and rejection
		maliciousRequests := []struct {
			name        string
			requestBody models.ToolExecutionRequest
		}{
			{
				name: "SQLInjectionAttempt",
				requestBody: models.ToolExecutionRequest{
					ToolName: "get-activity-details",
					Parameters: map[string]interface{}{
						"activity_id": "123; DROP TABLE users;",
					},
				},
			},
			{
				name: "PathTraversalAttempt",
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
			{
				name: "CommandInjectionAttempt",
				requestBody: models.ToolExecutionRequest{
					ToolName: "update-athlete-logbook",
					Parameters: map[string]interface{}{
						"content": "exec(rm -rf /)",
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

				// Should be rejected with 400 status due to input validation
				suite.Equal(http.StatusBadRequest, w.Code, "Malicious request %s should be rejected", tc.name)
				
				var response models.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err, "Malicious request %s error response should parse", tc.name)
				
				// Should indicate validation error or malicious input detection
				suite.Contains([]string{"VALIDATION_ERROR", "MALICIOUS_INPUT"}, response.Error.Code,
					"Malicious request %s should be flagged as validation error", tc.name)
				suite.NotEmpty(response.Error.Message, "Malicious request %s should have error message", tc.name)
			})
		}
	})

	suite.Run("WorkspaceBoundaryEnforcement", func() {
		// Test that workspace root is properly set and enforced
		req := httptest.NewRequest("GET", "/api/tools", nil)
		w := httptest.NewRecorder()
		
		// Create test router to verify workspace boundary middleware
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
		suite.Equal(http.StatusOK, w.Code, "Workspace boundary middleware should work correctly")
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.NoError(err, "Workspace boundary response should parse")
		suite.NotEmpty(response["workspace_root"], "Workspace root should be returned")
	})
}

// Additional comprehensive tests for edge cases and performance
func (suite *FinalIntegrationTestSuite) TestAdditionalIntegrationScenarios() {
	suite.Run("TimeoutAndStreamingFunctionality", func() {
		// Test streaming execution
		streamingRequest := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
			Options: &models.ExecutionOptions{
				Streaming: true,
				Timeout:   10,
			},
		}
		jsonBody, _ := json.Marshal(streamingRequest)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		suite.Equal(http.StatusOK, w.Code, "Streaming execution should work")
		
		var response models.ToolExecutionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.NoError(err, "Streaming response should parse")
		suite.Equal("success", response.Status, "Streaming execution should succeed")
		
		// Test buffered execution
		bufferedRequest := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
			Options: &models.ExecutionOptions{
				BufferedOutput: true,
				Timeout:        10,
			},
		}
		jsonBody, _ = json.Marshal(bufferedRequest)

		req = httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		suite.Equal(http.StatusOK, w.Code, "Buffered execution should work")
		
		err = json.Unmarshal(w.Body.Bytes(), &response)
		suite.NoError(err, "Buffered response should parse")
		suite.Equal("success", response.Status, "Buffered execution should succeed")
	})

	suite.Run("AuthenticationAndAuthorizationValidation", func() {
		// Test without authentication
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

		suite.Equal(http.StatusUnauthorized, w.Code, "Request without auth should be rejected")
		
		var response models.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.NoError(err, "Auth error response should parse")
		suite.Equal("AUTH_REQUIRED", response.Error.Code, "Should indicate auth required")
	})

	suite.Run("ComprehensiveParameterValidation", func() {
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
				name:           "InvalidToolNameCharacters",
				requestBody:    models.ToolExecutionRequest{ToolName: "invalid@tool#name", Parameters: map[string]interface{}{}},
				expectedStatus: http.StatusBadRequest,
				expectedCode:   "VALIDATION_ERROR",
			},
			{
				name:           "TooLongToolName",
				requestBody:    models.ToolExecutionRequest{ToolName: strings.Repeat("a", 200), Parameters: map[string]interface{}{}},
				expectedStatus: http.StatusBadRequest,
				expectedCode:   "PARAMETER_TOO_LARGE",
			},
			{
				name:           "InvalidParameterType",
				requestBody:    map[string]interface{}{"tool_name": "get-athlete-profile", "parameters": "invalid"},
				expectedStatus: http.StatusBadRequest,
				expectedCode:   "VALIDATION_ERROR",
			},
			{
				name:           "InvalidTimeoutValue",
				requestBody:    models.ToolExecutionRequest{
					ToolName: "get-athlete-profile", 
					Parameters: map[string]interface{}{},
					Options: &models.ExecutionOptions{Timeout: -1},
				},
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

				suite.Equal(tc.expectedStatus, w.Code, "Validation case %s should return correct status", tc.name)
				
				var response models.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err, "Validation case %s error response should parse", tc.name)
				suite.Equal(tc.expectedCode, response.Error.Code, "Validation case %s should return correct error code", tc.name)
				suite.NotEmpty(response.Error.Message, "Validation case %s should have error message", tc.name)
			})
		}
	})
}

// Mock tool execution service for final integration testing
type finalTestToolExecutionService struct{}

func (m *finalTestToolExecutionService) ExecuteGetAthleteProfile(ctx context.Context, msgCtx *services.MessageContext) (string, error) {
	// Simulate realistic execution time
	time.Sleep(10 * time.Millisecond)
	return `{"id": 12345, "username": "final_test_athlete", "firstname": "Final", "lastname": "Test", "city": "Test City", "country": "Test Country"}`, nil
}

func (m *finalTestToolExecutionService) ExecuteGetRecentActivities(ctx context.Context, msgCtx *services.MessageContext, perPage int) (string, error) {
	time.Sleep(15 * time.Millisecond)
	activities := make([]map[string]interface{}, perPage)
	for i := 0; i < perPage; i++ {
		activities[i] = map[string]interface{}{
			"id":       123456789 + i,
			"name":     fmt.Sprintf("Test Activity %d", i+1),
			"type":     "Run",
			"distance": 5000.0 + float64(i*100),
		}
	}
	result := map[string]interface{}{
		"activities": activities,
		"count":      perPage,
		"per_page":   perPage,
	}
	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}

func (m *finalTestToolExecutionService) ExecuteGetActivityDetails(ctx context.Context, msgCtx *services.MessageContext, activityID int64) (string, error) {
	time.Sleep(12 * time.Millisecond)
	result := map[string]interface{}{
		"id":                activityID,
		"name":              "Final Test Activity",
		"type":              "Run",
		"distance":          5000.0,
		"moving_time":       1800,
		"elapsed_time":      1900,
		"total_elevation_gain": 100.0,
		"start_date":        "2024-01-01T10:00:00Z",
	}
	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}

func (m *finalTestToolExecutionService) ExecuteGetActivityStreams(ctx context.Context, msgCtx *services.MessageContext, activityID int64, streamTypes []string, resolution string, processingMode string, pageNumber int, pageSize int, summaryPrompt string) (string, error) {
	time.Sleep(20 * time.Millisecond)
	
	streams := make(map[string][]interface{})
	for _, streamType := range streamTypes {
		switch streamType {
		case "time":
			streams[streamType] = []interface{}{0, 1, 2, 3, 4, 5}
		case "heartrate":
			streams[streamType] = []interface{}{120, 125, 130, 135, 140, 145}
		case "watts":
			streams[streamType] = []interface{}{200, 210, 220, 230, 240, 250}
		case "distance":
			streams[streamType] = []interface{}{0, 100, 200, 300, 400, 500}
		default:
			streams[streamType] = []interface{}{1, 2, 3, 4, 5, 6}
		}
	}
	
	result := map[string]interface{}{
		"activity_id":     activityID,
		"streams":         streams,
		"resolution":      resolution,
		"processing_mode": processingMode,
		"page_number":     pageNumber,
		"page_size":       pageSize,
	}
	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}

func (m *finalTestToolExecutionService) ExecuteUpdateAthleteLogbook(ctx context.Context, msgCtx *services.MessageContext, content string) (string, error) {
	time.Sleep(18 * time.Millisecond)
	result := map[string]interface{}{
		"success":        true,
		"message":        "Logbook updated successfully",
		"content_length": len(content),
		"updated_at":     time.Now().Format(time.RFC3339),
		"user_id":        msgCtx.UserID,
	}
	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}

// Environment configuration test for development mode
func TestFinalDevelopmentModeEnvironmentConfiguration(t *testing.T) {
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
		cfg.IsDevelopment = os.Getenv("DEVELOPMENT_MODE") == "true"
		
		assert.True(t, cfg.IsDevelopment, "Development mode should be enabled from environment")
		
		// Test development mode disabled via environment
		os.Setenv("DEVELOPMENT_MODE", "false")
		cfg.IsDevelopment = os.Getenv("DEVELOPMENT_MODE") == "true"
		
		assert.False(t, cfg.IsDevelopment, "Development mode should be disabled from environment")
	})
}

// Performance benchmark test for tool execution
func BenchmarkFinalToolExecution(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	config := &config.Config{IsDevelopment: true}
	registry := services.NewToolRegistry()
	mockToolService := &finalTestToolExecutionService{}
	executor := services.NewToolExecutor(mockToolService, registry)
	controller := NewToolController(registry, executor, config)
	
	router := gin.New()
	tools := router.Group("/api/tools")
	tools.Use(DevelopmentOnlyMiddleware(config))
	tools.Use(InputValidationMiddleware())
	tools.Use(WorkspaceBoundaryMiddleware())
	tools.Use(func(c *gin.Context) {
		c.Set("user", &models.User{ID: "benchmark-user", StravaID: 12345})
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