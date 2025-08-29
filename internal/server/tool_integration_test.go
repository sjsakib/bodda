package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bodda/internal/config"
	"bodda/internal/models"
	"bodda/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestToolEndpoints_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test with development mode enabled
	t.Run("development mode - endpoints accessible", func(t *testing.T) {
		config := &config.Config{IsDevelopment: true}
		
		// Create real services for integration test
		toolRegistry := services.NewToolRegistry()
		toolExecutionService := &mockToolExecutionService{}
		toolExecutor := services.NewToolExecutor(toolExecutionService, toolRegistry)
		toolController := NewToolController(toolRegistry, toolExecutor, config)

		// Setup router with all middleware
		router := gin.New()
		tools := router.Group("/api/tools")
		tools.Use(DevelopmentOnlyMiddleware(config))
		tools.Use(InputValidationMiddleware())
		tools.Use(WorkspaceBoundaryMiddleware())
		tools.Use(func(c *gin.Context) {
			// Mock auth middleware
			user := &models.User{
				ID:        "test-user",
				StravaID:  12345,
				FirstName: "Test",
				LastName:  "User",
			}
			c.Set("user", user)
			c.Next()
		})
		{
			tools.GET("", toolController.ListTools)
			tools.GET("/:toolName/schema", toolController.GetToolSchema)
			tools.POST("/execute", toolController.ExecuteTool)
		}

		// Test GET /api/tools
		req := httptest.NewRequest("GET", "/api/tools", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var toolsResponse models.ToolListResponse
		err := json.Unmarshal(w.Body.Bytes(), &toolsResponse)
		assert.NoError(t, err)
		assert.Greater(t, toolsResponse.Count, 0)

		// Test GET /api/tools/{toolName}/schema
		req = httptest.NewRequest("GET", "/api/tools/get-athlete-profile/schema", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var schemaResponse models.ToolSchemaResponse
		err = json.Unmarshal(w.Body.Bytes(), &schemaResponse)
		assert.NoError(t, err)
		assert.Equal(t, "get-athlete-profile", schemaResponse.Schema.Name)

		// Test POST /api/tools/execute
		executeRequest := models.ToolExecutionRequest{
			ToolName:   "get-athlete-profile",
			Parameters: map[string]interface{}{},
		}
		jsonBody, _ := json.Marshal(executeRequest)
		req = httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// The request should succeed (200) or fail with validation error (400)
		// Both are acceptable since we're testing the routing works
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest, 
			"Expected 200 or 400, got %d. Body: %s", w.Code, w.Body.String())
		
		if w.Code == http.StatusOK {
			var executeResponse models.ToolExecutionResponse
			err = json.Unmarshal(w.Body.Bytes(), &executeResponse)
			assert.NoError(t, err)
			assert.Equal(t, "success", executeResponse.Status)
		}
	})

	// Test with production mode - endpoints should return 404
	t.Run("production mode - endpoints return 404", func(t *testing.T) {
		config := &config.Config{IsDevelopment: false}
		
		toolRegistry := services.NewToolRegistry()
		toolExecutionService := &mockToolExecutionService{}
		toolExecutor := services.NewToolExecutor(toolExecutionService, toolRegistry)
		toolController := NewToolController(toolRegistry, toolExecutor, config)

		// Setup router with development middleware
		router := gin.New()
		tools := router.Group("/api/tools")
		tools.Use(DevelopmentOnlyMiddleware(config))
		{
			tools.GET("", toolController.ListTools)
			tools.GET("/:toolName/schema", toolController.GetToolSchema)
			tools.POST("/execute", toolController.ExecuteTool)
		}

		// All endpoints should return 404 in production
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
			var req *http.Request
			if endpoint.body != nil {
				req = httptest.NewRequest(endpoint.method, endpoint.path, bytes.NewBuffer(endpoint.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(endpoint.method, endpoint.path, nil)
			}
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code, "Expected 404 for %s %s in production mode", endpoint.method, endpoint.path)
		}
	})
}

// Mock tool execution service for integration tests
type mockToolExecutionService struct{}

func (m *mockToolExecutionService) ExecuteGetAthleteProfile(ctx context.Context, msgCtx *services.MessageContext) (string, error) {
	return "Mock athlete profile data", nil
}

func (m *mockToolExecutionService) ExecuteGetRecentActivities(ctx context.Context, msgCtx *services.MessageContext, perPage int) (string, error) {
	return "Mock recent activities data", nil
}

func (m *mockToolExecutionService) ExecuteGetActivityDetails(ctx context.Context, msgCtx *services.MessageContext, activityID int64) (string, error) {
	return "Mock activity details data", nil
}

func (m *mockToolExecutionService) ExecuteGetActivityStreams(ctx context.Context, msgCtx *services.MessageContext, activityID int64, streamTypes []string, resolution string, processingMode string, pageNumber int, pageSize int, summaryPrompt string) (string, error) {
	return "Mock activity streams data", nil
}

func (m *mockToolExecutionService) ExecuteUpdateAthleteLogbook(ctx context.Context, msgCtx *services.MessageContext, content string) (string, error) {
	return "Mock logbook update response", nil
}