package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolRegistry_Comprehensive(t *testing.T) {
	registry := NewToolRegistry()

	t.Run("GetAvailableTools_ReturnsAllExpectedTools", func(t *testing.T) {
		tools := registry.GetAvailableTools()
		
		assert.NotEmpty(t, tools, "Should return at least one tool")
		
		// Verify all expected tools are present
		expectedTools := map[string]bool{
			"get-athlete-profile":     false,
			"get-recent-activities":   false,
			"get-activity-details":    false,
			"get-activity-streams":    false,
			"update-athlete-logbook":  false,
		}
		
		for _, tool := range tools {
			if _, exists := expectedTools[tool.Name]; exists {
				expectedTools[tool.Name] = true
			}
		}
		
		for toolName, found := range expectedTools {
			assert.True(t, found, "Expected tool '%s' should be available", toolName)
		}
	})

	t.Run("GetAvailableTools_ToolsHaveRequiredFields", func(t *testing.T) {
		tools := registry.GetAvailableTools()
		
		for _, tool := range tools {
			assert.NotEmpty(t, tool.Name, "Tool name should not be empty")
			assert.NotEmpty(t, tool.Description, "Tool description should not be empty")
			assert.NotNil(t, tool.Parameters, "Tool parameters should not be nil")
			
			// Verify parameters have proper structure
			if params, ok := tool.Parameters["properties"]; ok {
				assert.IsType(t, map[string]interface{}{}, params, "Parameters properties should be a map")
			}
			
			// Verify examples are properly structured
			for i, example := range tool.Examples {
				assert.NotEmpty(t, example.Description, "Example %d should have description", i)
				assert.NotNil(t, example.Request, "Example %d should have request", i)
				assert.NotNil(t, example.Response, "Example %d should have response", i)
			}
		}
	})

	t.Run("GetToolSchema_ValidTool_ReturnsCompleteSchema", func(t *testing.T) {
		schema, err := registry.GetToolSchema("get-athlete-profile")
		
		require.NoError(t, err)
		require.NotNil(t, schema)
		
		assert.Equal(t, "get-athlete-profile", schema.Name)
		assert.NotEmpty(t, schema.Description)
		assert.NotNil(t, schema.Parameters)
		assert.NotNil(t, schema.Required)
		assert.NotNil(t, schema.Optional)
		assert.NotEmpty(t, schema.Examples)
	})

	t.Run("GetToolSchema_ToolWithRequiredParams_CorrectlyIdentifiesRequired", func(t *testing.T) {
		schema, err := registry.GetToolSchema("get-activity-details")
		
		require.NoError(t, err)
		require.NotNil(t, schema)
		
		assert.Contains(t, schema.Required, "activity_id", "activity_id should be required")
		assert.NotContains(t, schema.Optional, "activity_id", "activity_id should not be in optional")
	})

	t.Run("GetToolSchema_ToolWithOptionalParams_CorrectlyIdentifiesOptional", func(t *testing.T) {
		schema, err := registry.GetToolSchema("get-recent-activities")
		
		require.NoError(t, err)
		require.NotNil(t, schema)
		
		assert.Contains(t, schema.Optional, "per_page", "per_page should be optional")
		assert.NotContains(t, schema.Required, "per_page", "per_page should not be required")
	})

	t.Run("GetToolSchema_InvalidTool_ReturnsError", func(t *testing.T) {
		schema, err := registry.GetToolSchema("nonexistent-tool")
		
		assert.Error(t, err)
		assert.Nil(t, schema)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("ValidateToolCall_ValidParameters_NoError", func(t *testing.T) {
		// Test tool with no required parameters
		err := registry.ValidateToolCall("get-athlete-profile", map[string]interface{}{})
		assert.NoError(t, err)
		
		// Test tool with required parameters provided
		err = registry.ValidateToolCall("get-activity-details", map[string]interface{}{
			"activity_id": int64(123456),
		})
		assert.NoError(t, err)
		
		// Test tool with optional parameters
		err = registry.ValidateToolCall("get-recent-activities", map[string]interface{}{
			"per_page": 10,
		})
		assert.NoError(t, err)
	})

	t.Run("ValidateToolCall_MissingRequiredParameters_ReturnsError", func(t *testing.T) {
		err := registry.ValidateToolCall("get-activity-details", map[string]interface{}{})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required parameter")
		assert.Contains(t, err.Error(), "activity_id")
	})

	t.Run("ValidateToolCall_InvalidTool_ReturnsError", func(t *testing.T) {
		err := registry.ValidateToolCall("invalid-tool", map[string]interface{}{})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("IsToolAvailable_ValidTools_ReturnsTrue", func(t *testing.T) {
		validTools := []string{
			"get-athlete-profile",
			"get-recent-activities",
			"get-activity-details",
			"get-activity-streams",
			"update-athlete-logbook",
		}
		
		for _, toolName := range validTools {
			assert.True(t, registry.IsToolAvailable(toolName), "Tool '%s' should be available", toolName)
		}
	})

	t.Run("IsToolAvailable_InvalidTool_ReturnsFalse", func(t *testing.T) {
		invalidTools := []string{
			"nonexistent-tool",
			"invalid-tool",
			"",
			"get-invalid-tool",
		}
		
		for _, toolName := range invalidTools {
			assert.False(t, registry.IsToolAvailable(toolName), "Tool '%s' should not be available", toolName)
		}
	})

	t.Run("ToolExamples_HaveProperStructure", func(t *testing.T) {
		schema, err := registry.GetToolSchema("get-activity-streams")
		require.NoError(t, err)
		
		assert.NotEmpty(t, schema.Examples, "get-activity-streams should have examples")
		
		for i, example := range schema.Examples {
			assert.NotEmpty(t, example.Description, "Example %d should have description", i)
			assert.NotNil(t, example.Request, "Example %d should have request", i)
			assert.NotNil(t, example.Response, "Example %d should have response", i)
			
			// Verify request has required parameters
			if activityID, exists := example.Request["activity_id"]; exists {
				assert.IsType(t, int(0), activityID, "activity_id should be integer in example %d", i)
			}
		}
	})

	t.Run("ComplexParameterValidation_GetActivityStreams", func(t *testing.T) {
		// Test with all valid parameters
		validParams := map[string]interface{}{
			"activity_id":     int64(123456),
			"stream_types":    []string{"time", "heartrate", "watts"},
			"resolution":      "medium",
			"processing_mode": "auto",
			"page_number":     1,
			"page_size":       1000,
		}
		
		err := registry.ValidateToolCall("get-activity-streams", validParams)
		assert.NoError(t, err)
		
		// Test with only required parameters
		minimalParams := map[string]interface{}{
			"activity_id": int64(123456),
		}
		
		err = registry.ValidateToolCall("get-activity-streams", minimalParams)
		assert.NoError(t, err)
		
		// Test with missing required parameter
		invalidParams := map[string]interface{}{
			"stream_types": []string{"time", "heartrate"},
		}
		
		err = registry.ValidateToolCall("get-activity-streams", invalidParams)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "activity_id")
	})

	t.Run("ParameterSchemaConsistency", func(t *testing.T) {
		tools := registry.GetAvailableTools()
		
		for _, tool := range tools {
			schema, err := registry.GetToolSchema(tool.Name)
			require.NoError(t, err, "Should be able to get schema for tool %s", tool.Name)
			
			// Verify schema parameters match tool definition parameters
			assert.Equal(t, tool.Parameters, schema.Parameters, 
				"Schema parameters should match tool definition for %s", tool.Name)
			
			// Verify required/optional fields are properly categorized
			if properties, ok := schema.Parameters["properties"].(map[string]interface{}); ok {
				requiredFields := make(map[string]bool)
				if reqArray, ok := schema.Parameters["required"].([]interface{}); ok {
					for _, req := range reqArray {
						if reqStr, ok := req.(string); ok {
							requiredFields[reqStr] = true
						}
					}
				}
				
				// Check that all required fields are in properties
				for _, req := range schema.Required {
					assert.Contains(t, properties, req, 
						"Required field %s should be in properties for tool %s", req, tool.Name)
				}
				
				// Check that optional fields are in properties but not required
				for _, opt := range schema.Optional {
					assert.Contains(t, properties, opt, 
						"Optional field %s should be in properties for tool %s", opt, tool.Name)
					assert.False(t, requiredFields[opt], 
						"Optional field %s should not be in required list for tool %s", opt, tool.Name)
				}
			}
		}
	})
}

func TestToolRegistryWithAIService_Integration(t *testing.T) {
	// Create a mock AI service for integration testing
	mockAIService := &mockAIServiceForRegistry{}
	registry := NewToolRegistryWithAIService(mockAIService)

	t.Run("IntegrationWithAIService_HasExpectedTools", func(t *testing.T) {
		tools := registry.GetAvailableTools()
		
		expectedTools := []string{
			"get-athlete-profile",
			"get-recent-activities",
			"get-activity-details", 
			"get-activity-streams",
			"update-athlete-logbook",
		}
		
		toolNames := make(map[string]bool)
		for _, tool := range tools {
			toolNames[tool.Name] = true
		}
		
		for _, expected := range expectedTools {
			assert.True(t, toolNames[expected], "Expected tool '%s' should be available", expected)
		}
	})

	t.Run("ToolDefinitionsMatchAIServiceCapabilities", func(t *testing.T) {
		// Verify that tool definitions match what the AI service can actually execute
		schema, err := registry.GetToolSchema("get-activity-streams")
		require.NoError(t, err)
		
		// Check that complex parameters are properly defined
		properties, ok := schema.Parameters["properties"].(map[string]interface{})
		require.True(t, ok, "Parameters should have properties")
		
		// Verify stream_types parameter
		if streamTypes, exists := properties["stream_types"]; exists {
			streamTypesMap, ok := streamTypes.(map[string]interface{})
			require.True(t, ok, "stream_types should be a map")
			assert.Equal(t, "array", streamTypesMap["type"])
		}
		
		// Verify processing_mode parameter
		if processingMode, exists := properties["processing_mode"]; exists {
			processingModeMap, ok := processingMode.(map[string]interface{})
			require.True(t, ok, "processing_mode should be a map")
			assert.Equal(t, "string", processingModeMap["type"])
		}
	})
}

// Mock AI service for registry testing
type mockAIServiceForRegistry struct{}

func (m *mockAIServiceForRegistry) ProcessMessage(ctx context.Context, msgCtx *MessageContext) (<-chan string, error) {
	ch := make(chan string, 1)
	ch <- "mock response"
	close(ch)
	return ch, nil
}

func (m *mockAIServiceForRegistry) ProcessMessageSync(ctx context.Context, msgCtx *MessageContext) (string, error) {
	return "mock sync response", nil
}

func (m *mockAIServiceForRegistry) ExecuteGetAthleteProfile(ctx context.Context, msgCtx *MessageContext) (string, error) {
	return "mock profile", nil
}

func (m *mockAIServiceForRegistry) ExecuteGetRecentActivities(ctx context.Context, msgCtx *MessageContext, perPage int) (string, error) {
	return "mock activities", nil
}

func (m *mockAIServiceForRegistry) ExecuteGetActivityDetails(ctx context.Context, msgCtx *MessageContext, activityID int64) (string, error) {
	return "mock activity details", nil
}

func (m *mockAIServiceForRegistry) ExecuteGetActivityStreams(ctx context.Context, msgCtx *MessageContext, activityID int64, streamTypes []string, resolution string, processingMode string, pageNumber int, pageSize int, summaryPrompt string) (string, error) {
	return "mock activity streams", nil
}

func (m *mockAIServiceForRegistry) ExecuteUpdateAthleteLogbook(ctx context.Context, msgCtx *MessageContext, content string) (string, error) {
	return "mock logbook update", nil
}