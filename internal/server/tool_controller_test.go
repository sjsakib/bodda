package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"
	"bodda/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing
type mockToolRegistry struct {
	mock.Mock
}

func (m *mockToolRegistry) GetAvailableTools() []models.ToolDefinition {
	args := m.Called()
	return args.Get(0).([]models.ToolDefinition)
}

func (m *mockToolRegistry) GetToolSchema(toolName string) (*models.ToolSchema, error) {
	args := m.Called(toolName)
	return args.Get(0).(*models.ToolSchema), args.Error(1)
}

func (m *mockToolRegistry) ValidateToolCall(toolName string, parameters map[string]interface{}) error {
	args := m.Called(toolName, parameters)
	return args.Error(0)
}

func (m *mockToolRegistry) IsToolAvailable(toolName string) bool {
	args := m.Called(toolName)
	return args.Bool(0)
}

type mockToolExecutor struct {
	mock.Mock
}

func (m *mockToolExecutor) ExecuteTool(ctx context.Context, toolName string, parameters map[string]interface{}, msgCtx *services.MessageContext) (*models.ToolExecutionResult, error) {
	args := m.Called(ctx, toolName, parameters, msgCtx)
	return args.Get(0).(*models.ToolExecutionResult), args.Error(1)
}

func (m *mockToolExecutor) ExecuteToolWithOptions(ctx context.Context, toolName string, parameters map[string]interface{}, msgCtx *services.MessageContext, options *models.ExecutionOptions) (*models.ToolExecutionResult, error) {
	args := m.Called(ctx, toolName, parameters, msgCtx, options)
	return args.Get(0).(*models.ToolExecutionResult), args.Error(1)
}

func (m *mockToolExecutor) CancelJob(jobID string) bool {
	args := m.Called(jobID)
	return args.Bool(0)
}

func (m *mockToolExecutor) GetActiveJobCount() int {
	args := m.Called()
	return args.Int(0)
}

func TestToolController_ListTools(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mocks
	mockRegistry := &mockToolRegistry{}
	mockExecutor := &mockToolExecutor{}
	config := &config.Config{IsDevelopment: true}

	expectedTools := []models.ToolDefinition{
		{
			Name:        "test-tool",
			Description: "A test tool",
			Parameters:  map[string]interface{}{"type": "object"},
		},
	}

	mockRegistry.On("GetAvailableTools").Return(expectedTools)

	// Create controller
	controller := NewToolController(mockRegistry, mockExecutor, config)

	// Setup router
	router := gin.New()
	router.GET("/api/tools", controller.ListTools)

	// Make request
	req := httptest.NewRequest("GET", "/api/tools", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ToolListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.Count)
	assert.Equal(t, expectedTools, response.Tools)

	mockRegistry.AssertExpectations(t)
}

func TestToolController_GetToolSchema(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mocks
	mockRegistry := &mockToolRegistry{}
	mockExecutor := &mockToolExecutor{}
	config := &config.Config{IsDevelopment: true}

	expectedSchema := &models.ToolSchema{
		Name:        "test-tool",
		Description: "A test tool",
		Parameters:  map[string]interface{}{"type": "object"},
		Required:    []string{"param1"},
		Optional:    []string{"param2"},
	}

	mockRegistry.On("GetToolSchema", "test-tool").Return(expectedSchema, nil)

	// Create controller
	controller := NewToolController(mockRegistry, mockExecutor, config)

	// Setup router
	router := gin.New()
	router.GET("/api/tools/:toolName/schema", controller.GetToolSchema)

	// Make request
	req := httptest.NewRequest("GET", "/api/tools/test-tool/schema", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ToolSchemaResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedSchema, response.Schema)

	mockRegistry.AssertExpectations(t)
}

func TestToolController_GetToolSchema_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mocks
	mockRegistry := &mockToolRegistry{}
	mockExecutor := &mockToolExecutor{}
	config := &config.Config{IsDevelopment: true}

	mockRegistry.On("GetToolSchema", "nonexistent-tool").Return((*models.ToolSchema)(nil), assert.AnError)

	// Create controller
	controller := NewToolController(mockRegistry, mockExecutor, config)

	// Setup router
	router := gin.New()
	router.GET("/api/tools/:toolName/schema", controller.GetToolSchema)

	// Make request
	req := httptest.NewRequest("GET", "/api/tools/nonexistent-tool/schema", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "TOOL_NOT_FOUND", response.Error.Code)

	mockRegistry.AssertExpectations(t)
}

func TestToolController_ExecuteTool(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mocks
	mockRegistry := &mockToolRegistry{}
	mockExecutor := &mockToolExecutor{}
	config := &config.Config{IsDevelopment: true}

	// Mock user for authentication
	user := &models.User{
		ID:        "user123",
		StravaID:  12345,
		FirstName: "Test",
		LastName:  "User",
	}

	// Setup expectations
	mockRegistry.On("IsToolAvailable", "test-tool").Return(true)
	mockRegistry.On("ValidateToolCall", "test-tool", map[string]interface{}{"param1": "value1"}).Return(nil)

	expectedResult := &models.ToolExecutionResult{
		ToolName:  "test-tool",
		Success:   true,
		Data:      "Tool executed successfully",
		Duration:  100,
		Timestamp: time.Now(),
	}

	mockExecutor.On("ExecuteToolWithOptions", mock.Anything, "test-tool", map[string]interface{}{"param1": "value1"}, mock.Anything, (*models.ExecutionOptions)(nil)).Return(expectedResult, nil)

	// Create controller
	controller := NewToolController(mockRegistry, mockExecutor, config)

	// Setup router with auth middleware mock
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	})
	router.POST("/api/tools/execute", controller.ExecuteTool)

	// Create request
	requestBody := models.ToolExecutionRequest{
		ToolName:   "test-tool",
		Parameters: map[string]interface{}{"param1": "value1"},
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ToolExecutionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)
	assert.NotNil(t, response.Result)
	assert.Equal(t, expectedResult.ToolName, response.Result.ToolName)
	assert.Equal(t, expectedResult.Success, response.Result.Success)
	assert.Equal(t, expectedResult.Data, response.Result.Data)
	assert.NotNil(t, response.Metadata)

	mockRegistry.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

func TestToolController_ExecuteTool_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mocks
	mockRegistry := &mockToolRegistry{}
	mockExecutor := &mockToolExecutor{}
	config := &config.Config{IsDevelopment: true}

	// Mock user for authentication
	user := &models.User{
		ID:        "user123",
		StravaID:  12345,
		FirstName: "Test",
		LastName:  "User",
	}

	// Setup expectations
	mockRegistry.On("IsToolAvailable", "test-tool").Return(true)
	mockRegistry.On("ValidateToolCall", "test-tool", map[string]interface{}{"invalid": "param"}).Return(assert.AnError)

	// Create controller
	controller := NewToolController(mockRegistry, mockExecutor, config)

	// Setup router with auth middleware mock
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	})
	router.POST("/api/tools/execute", controller.ExecuteTool)

	// Create request with invalid parameters
	requestBody := models.ToolExecutionRequest{
		ToolName:   "test-tool",
		Parameters: map[string]interface{}{"invalid": "param"},
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)

	mockRegistry.AssertExpectations(t)
}

func TestToolController_ExecuteTool_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mocks
	mockRegistry := &mockToolRegistry{}
	mockExecutor := &mockToolExecutor{}
	config := &config.Config{IsDevelopment: true}

	// Setup expectations for the calls that happen before auth check
	mockRegistry.On("IsToolAvailable", "test-tool").Return(true)
	mockRegistry.On("ValidateToolCall", "test-tool", map[string]interface{}{"param1": "value1"}).Return(nil)

	// Create controller
	controller := NewToolController(mockRegistry, mockExecutor, config)

	// Setup router without auth middleware (no user in context)
	router := gin.New()
	router.POST("/api/tools/execute", controller.ExecuteTool)

	// Create request
	requestBody := models.ToolExecutionRequest{
		ToolName:   "test-tool",
		Parameters: map[string]interface{}{"param1": "value1"},
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/tools/execute", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "AUTH_REQUIRED", response.Error.Code)

	mockRegistry.AssertExpectations(t)
}