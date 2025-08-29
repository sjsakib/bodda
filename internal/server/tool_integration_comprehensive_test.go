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
	"github.com/stretchr/testify/require"
)

func TestToolExecution_EndToEnd_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create real implementations for integration testing
	mockAIService := &mockAIServiceIntegration{}
	registry := services.NewToolRegistryWithAIService(mockAIService)
	executor := services.NewToolExecutor(mockAIService, registry)
	config := &config.Config{IsDevelopment: true}
	controller := NewToolController(registry, executor, config)

	// Test user
	testUser := &models.User{
		ID:          "integration-test-user",
		StravaID:    12345,
		AccessToken: "test-access-token",
		FirstName:   "Integration",
		LastName:    "Test",
	}

	t.Run("EndToEnd_GetAthleteProfile_Success", func(t *testing.T) {
		router := setupIntegrationRouter(controller, testUser)

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
		assert.True(t, response.Result.Success)
		assert.Equal(t, "get-athlete-profile", response.Result.ToolName)
		assert.Contains(t, response.Result.Data, "athlete_profile")
		assert.Greater(t, response.Result.Duration, int64(0))
		assert.NotNil(t, response.Metadata)
		assert.NotEmpty(t, response.Metadata.RequestID)
	})

	t.Run("EndToEnd_GetRecentActivities_WithParameters_Success", func(t *testing.T) {
		router := setupIntegrationRouter(controller, testUser)

		requestBody := models.ToolExecutionRequest{
			ToolName: "get-recent-activities",
			Parameters: map[string]interface{}{
				"per_page": 10,
			},
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
		assert.Equal(t, "get-recent-activities", response.Result.ToolName)
		assert.Contains(t, response.Result.Data, "recent_activities")
	})

	t.Run("EndToEnd_GetActivityDetails_WithRequiredParameters_Success", func(t *testing.T) {
		router := setupIntegrationRouter(controller, testUser)

		requestBody := models.ToolExecutionRequest{
			ToolName: "get-activity-details",
			Parameters: map[string]interface{}{
				"activity_id": int64(123456789),
			},
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
		assert.Equal(t, "get-activity-details", response.Result.ToolName)
		assert.Contains(t, response.Result.Data, "activity_details")
	})

	t.Run("EndToEnd_GetActivityStreams_ComplexParameters_Success", func(t *testing.T) {
		router := setupIntegrationRouter(controller, testUser)

		requestBody := models.ToolExecutionRequest{
			ToolName: "get-activity-streams",
			Parameters: map[string]interface{}{
				"activity_id":     int64(123456789),
				"stream_types":    []string{"time", "heartrate", "watts"},
				"resolution":      "medium",
				"processing_mode": "auto",
				"page_number":     1,
				"page_size":       1000,
			},
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
		assert.Equal(t, "get-activity-streams", response.Result.ToolName)
		assert.Contains(t, response.Result.Data, "activity_streams")
	})

	t.Run("EndToEnd_UpdateAthleteLogbook_Success", func(t *testing.T) {
		router := setupIntegrationRouter(controller, testUser)

		requestBody := models.ToolExecutionRequest{
			ToolName: "update-athlete-logbook",
			Parameters: map[string]interface{}{
				"content": "Integration test logbook update with comprehensive athlete data and training insights.",
			},
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
		assert.Equal(t, "update-athlete-logbook", response.Result.ToolName)
		assert.Contains(t, response.Result.Data, "logbook_updated")
	})

	t.Run("EndToEnd_StreamingExecution_Success", func(t *testing.T) {
		router := setupIntegrationRouter(controller, testUser)

		requestBody := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
			Options: &models.ExecutionOptions{
				Streaming: true,
				Timeout:   10,
			},
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
		assert.Contains(t, response.Result.Data, "athlete_profile")
	})

	t.Run("EndToEnd_BufferedExecution_Success", func(t *testing.T) {
		router := setupIntegrationRouter(controller, testUser)

		requestBody := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
			Options: &models.ExecutionOptions{
				BufferedOutput: true,
				Timeout:        10,
			},
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
	})

	t.Run("EndToEnd_ToolDiscovery_ListAndSchema", func(t *testing.T) {
		router := setupIntegrationRouter(controller, testUser)

		// First, list all available tools
		req := httptest.NewRequest("GET", "/api/tools", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var listResponse models.ToolListResponse
		err := json.Unmarshal(w.Body.Bytes(), &listResponse)
		require.NoError(t, err)

		assert.Greater(t, listResponse.Count, 0)
		assert.NotEmpty(t, listResponse.Tools)

		// Then, get schema for each tool
		for _, tool := range listResponse.Tools {
			t.Run(fmt.Sprintf("Schema_%s", tool.Name), func(t *testing.T) {
				req := httptest.NewRequest("GET", "/api/tools/"+tool.Name+"/schema", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)

				var schemaResponse models.ToolSchemaResponse
				err := json.Unmarshal(w.Body.Bytes(), &schemaResponse)
				require.NoError(t, err)

				assert.Equal(t, tool.Name, schemaResponse.Schema.Name)
				assert.NotEmpty(t, schemaResponse.Schema.Description)
				assert.NotNil(t, schemaResponse.Schema.Parameters)
				assert.NotNil(t, schemaResponse.Schema.Required)
				assert.NotNil(t, schemaResponse.Schema.Optional)
				assert.NotEmpty(t, schemaResponse.Schema.Examples)
			})
		}
	})

	t.Run("EndToEnd_ErrorScenarios_ProperHandling", func(t *testing.T) {
		router := setupIntegrationRouter(controller, testUser)

		// Test missing required parameter
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
	})

	t.Run("EndToEnd_ResponseFormatConsistency", func(t *testing.T) {
		router := setupIntegrationRouter(controller, testUser)

		tools := []string{
			"get-athlete-profile",
			"get-recent-activities",
		}

		for _, toolName := range tools {
			t.Run(fmt.Sprintf("ResponseFormat_%s", toolName), func(t *testing.T) {
				requestBody := models.ToolExecutionRequest{
					ToolName:   toolName,
					Parameters: getValidParametersForIntegrationTool(toolName),
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

				// Verify consistent response structure
				assert.Equal(t, "success", response.Status)
				assert.NotNil(t, response.Result)
				assert.NotNil(t, response.Metadata)

				// Verify result structure
				assert.Equal(t, toolName, response.Result.ToolName)
				assert.True(t, response.Result.Success)
				assert.NotNil(t, response.Result.Data)
				assert.Greater(t, response.Result.Duration, int64(0))
				assert.False(t, response.Result.Timestamp.IsZero())

				// Verify metadata structure
				assert.NotEmpty(t, response.Metadata.RequestID)
				assert.False(t, response.Metadata.Timestamp.IsZero())
				assert.Greater(t, response.Metadata.Duration, int64(0))
			})
		}
	})
}

func TestToolExecution_SecurityBoundaryEnforcement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create service that simulates security boundary enforcement
	mockAIService := &mockAIServiceWithSecurity{}
	registry := services.NewToolRegistryWithAIService(mockAIService)
	executor := services.NewToolExecutor(mockAIService, registry)
	config := &config.Config{IsDevelopment: true}
	controller := NewToolController(registry, executor, config)

	testUser := &models.User{
		ID:          "security-test-user",
		StravaID:    12345,
		AccessToken: "test-access-token",
		FirstName:   "Security",
		LastName:    "Test",
	}

	t.Run("Security_MaliciousParameterHandling", func(t *testing.T) {
		router := setupIntegrationRouter(controller, testUser)

		maliciousParameters := map[string]interface{}{
			"script_injection":  "<script>alert('xss')</script>",
			"sql_injection":     "'; DROP TABLE users; --",
			"path_traversal":    "../../../etc/passwd",
			"command_injection": "; rm -rf /",
			"null_byte":         "test\x00.txt",
		}

		requestBody := models.ToolExecutionRequest{
			ToolName:   "update-athlete-logbook",
			Parameters: map[string]interface{}{
				"content": maliciousParameters,
			},
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should succeed but with sanitized parameters
		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ToolExecutionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "success", response.Status)
		assert.True(t, response.Result.Success)
		
		// Verify that malicious content was handled safely
		responseData, ok := response.Result.Data.(string)
		assert.True(t, ok)
		assert.Contains(t, responseData, "sanitized")
	})

	t.Run("Security_WorkspaceBoundaryEnforcement", func(t *testing.T) {
		router := setupIntegrationRouter(controller, testUser)

		// Attempt to access files outside workspace
		requestBody := models.ToolExecutionRequest{
			ToolName: "update-athlete-logbook",
			Parameters: map[string]interface{}{
				"content": "Attempt to access: ../../../etc/passwd",
			},
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

		// Should succeed but with workspace boundary enforcement
		assert.True(t, response.Result.Success)
		responseData, ok := response.Result.Data.(string)
		assert.True(t, ok)
		assert.Contains(t, responseData, "workspace boundary enforced")
	})

	t.Run("Security_InputSanitization", func(t *testing.T) {
		router := setupIntegrationRouter(controller, testUser)

		// Test various input sanitization scenarios
		testCases := []struct {
			name  string
			input string
		}{
			{"HTMLTags", "<div>test</div>"},
			{"JavaScriptCode", "javascript:alert('test')"},
			{"SQLInjection", "1' OR '1'='1"},
			{"ShellCommands", "$(rm -rf /)"},
			{"UnicodeExploits", "\u0000\u0001\u0002"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				requestBody := models.ToolExecutionRequest{
					ToolName: "update-athlete-logbook",
					Parameters: map[string]interface{}{
						"content": tc.input,
					},
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

				assert.True(t, response.Result.Success)
				// Verify input was sanitized
				responseData, ok := response.Result.Data.(string)
				assert.True(t, ok)
				assert.Contains(t, responseData, "input sanitized")
			})
		}
	})
}

// Helper functions and mock implementations for integration testing

func setupIntegrationRouter(controller *ToolController, user *models.User) *gin.Engine {
	router := gin.New()
	
	// Add auth middleware that sets the test user
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	})
	
	// Add routes
	router.GET("/api/tools", controller.ListTools)
	router.GET("/api/tools/:toolName/schema", controller.GetToolSchema)
	router.POST("/api/tools/execute", controller.ExecuteTool)
	
	return router
}

func getValidParametersForIntegrationTool(toolName string) map[string]interface{} {
	switch toolName {
	case "get-athlete-profile":
		return map[string]interface{}{}
	case "get-recent-activities":
		return map[string]interface{}{
			"per_page": 5,
		}
	case "get-activity-details":
		return map[string]interface{}{
			"activity_id": int64(123456789),
		}
	case "get-activity-streams":
		return map[string]interface{}{
			"activity_id":  int64(123456789),
			"stream_types": []string{"time", "heartrate"},
		}
	case "update-athlete-logbook":
		return map[string]interface{}{
			"content": "Integration test logbook content",
		}
	default:
		return map[string]interface{}{}
	}
}

// Mock AI service for integration testing
type mockAIServiceIntegration struct{}

func (m *mockAIServiceIntegration) ProcessMessage(ctx context.Context, msgCtx *services.MessageContext) (<-chan string, error) {
	ch := make(chan string, 1)
	ch <- "mock integration response"
	close(ch)
	return ch, nil
}

func (m *mockAIServiceIntegration) ProcessMessageSync(ctx context.Context, msgCtx *services.MessageContext) (string, error) {
	return "mock integration sync response", nil
}

func (m *mockAIServiceIntegration) ExecuteGetAthleteProfile(ctx context.Context, msgCtx *services.MessageContext) (string, error) {
	// Simulate realistic response time
	time.Sleep(50 * time.Millisecond)
	return `{"athlete_profile": {"id": 12345, "name": "Integration Test Athlete", "city": "Test City"}}`, nil
}

func (m *mockAIServiceIntegration) ExecuteGetRecentActivities(ctx context.Context, msgCtx *services.MessageContext, perPage int) (string, error) {
	time.Sleep(75 * time.Millisecond)
	return fmt.Sprintf(`{"recent_activities": {"count": %d, "activities": [{"id": 1, "name": "Test Run"}]}}`, perPage), nil
}

func (m *mockAIServiceIntegration) ExecuteGetActivityDetails(ctx context.Context, msgCtx *services.MessageContext, activityID int64) (string, error) {
	time.Sleep(60 * time.Millisecond)
	return fmt.Sprintf(`{"activity_details": {"id": %d, "name": "Test Activity", "distance": 5000}}`, activityID), nil
}

func (m *mockAIServiceIntegration) ExecuteGetActivityStreams(ctx context.Context, msgCtx *services.MessageContext, activityID int64, streamTypes []string, resolution string, processingMode string, pageNumber int, pageSize int, summaryPrompt string) (string, error) {
	time.Sleep(100 * time.Millisecond)
	return fmt.Sprintf(`{"activity_streams": {"activity_id": %d, "streams": %v, "resolution": "%s"}}`, activityID, streamTypes, resolution), nil
}

func (m *mockAIServiceIntegration) ExecuteUpdateAthleteLogbook(ctx context.Context, msgCtx *services.MessageContext, content string) (string, error) {
	time.Sleep(80 * time.Millisecond)
	return fmt.Sprintf(`{"logbook_updated": true, "content_length": %d}`, len(content)), nil
}

// Mock AI service with security features for security testing
type mockAIServiceWithSecurity struct{}

func (m *mockAIServiceWithSecurity) ProcessMessage(ctx context.Context, msgCtx *services.MessageContext) (<-chan string, error) {
	ch := make(chan string, 1)
	ch <- "mock security response"
	close(ch)
	return ch, nil
}

func (m *mockAIServiceWithSecurity) ProcessMessageSync(ctx context.Context, msgCtx *services.MessageContext) (string, error) {
	return "mock security sync response", nil
}

func (m *mockAIServiceWithSecurity) ExecuteGetAthleteProfile(ctx context.Context, msgCtx *services.MessageContext) (string, error) {
	return `{"athlete_profile": "sanitized profile data"}`, nil
}

func (m *mockAIServiceWithSecurity) ExecuteGetRecentActivities(ctx context.Context, msgCtx *services.MessageContext, perPage int) (string, error) {
	return `{"recent_activities": "sanitized activities data"}`, nil
}

func (m *mockAIServiceWithSecurity) ExecuteGetActivityDetails(ctx context.Context, msgCtx *services.MessageContext, activityID int64) (string, error) {
	return `{"activity_details": "sanitized activity data"}`, nil
}

func (m *mockAIServiceWithSecurity) ExecuteGetActivityStreams(ctx context.Context, msgCtx *services.MessageContext, activityID int64, streamTypes []string, resolution string, processingMode string, pageNumber int, pageSize int, summaryPrompt string) (string, error) {
	return `{"activity_streams": "sanitized streams data"}`, nil
}

func (m *mockAIServiceWithSecurity) ExecuteUpdateAthleteLogbook(ctx context.Context, msgCtx *services.MessageContext, content string) (string, error) {
	// Simulate security checks and sanitization
	if containsMaliciousContent(content) {
		return `{"logbook_updated": true, "message": "input sanitized and workspace boundary enforced"}`, nil
	}
	return `{"logbook_updated": true, "message": "content processed safely"}`, nil
}

func containsMaliciousContent(content string) bool {
	maliciousPatterns := []string{
		"<script>", "javascript:", "'; DROP", "$(", "../", "\x00",
	}
	
	contentStr := fmt.Sprintf("%v", content)
	for _, pattern := range maliciousPatterns {
		if strings.Contains(contentStr, pattern) {
			return true
		}
	}
	return false
}