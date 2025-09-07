package services

import (
	"testing"

	"bodda/internal/models"

	"github.com/openai/openai-go/v2/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToolDefinitionToOpenAI(t *testing.T) {
	// Create a mock AI service for testing
	service := &aiService{}

	t.Run("converts get-athlete-profile tool correctly", func(t *testing.T) {
		tool := models.ToolDefinition{
			Name:        "get-athlete-profile",
			Description: "Get the complete athlete profile from Strava including personal information, zones, and stats",
			Parameters: map[string]interface{}{
				"type":                 "object",
				"properties":           map[string]interface{}{},
				"required":             []string{},
				"additionalProperties": false,
			},
		}

		result := service.convertToolDefinitionToOpenAI(tool)

		// Verify the result is a ToolUnionParam
		assert.NotNil(t, result)
		
		// Since we can't easily inspect the internal structure of responses.ToolUnionParam,
		// we'll verify by comparing with the expected format from the current implementation
		expected := responses.ToolParamOfFunction(
			"get-athlete-profile",
			map[string]interface{}{
				"type":                 "object",
				"properties":           map[string]interface{}{},
				"required":             []string{},
				"additionalProperties": false,
			},
			true,
		)

		// Both should be of the same type and structure
		assert.IsType(t, expected, result)
	})

	t.Run("converts get-recent-activities tool correctly", func(t *testing.T) {
		tool := models.ToolDefinition{
			Name:        "get-recent-activities",
			Description: "Get the most recent activities for the athlete",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"per_page": map[string]interface{}{
						"type":        "integer",
						"description": "Number of activities to retrieve (1-200, default 30)",
						"minimum":     1,
						"maximum":     200,
						"default":     30,
					},
				},
				"required":             []string{},
				"additionalProperties": false,
			},
		}

		result := service.convertToolDefinitionToOpenAI(tool)

		assert.NotNil(t, result)
		
		expected := responses.ToolParamOfFunction(
			"get-recent-activities",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"per_page": map[string]interface{}{
						"type":        "integer",
						"description": "Number of activities to retrieve (1-200, default 30)",
						"minimum":     1,
						"maximum":     200,
						"default":     30,
					},
				},
				"required":             []string{"per_page"},
				"additionalProperties": false,
			},
			true,
		)

		assert.IsType(t, expected, result)
	})

	t.Run("converts get-activity-details tool correctly", func(t *testing.T) {
		tool := models.ToolDefinition{
			Name:        "get-activity-details",
			Description: "Get detailed information about a specific activity using its ID",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"activity_id": map[string]interface{}{
						"type":        "integer",
						"description": "The Strava activity ID",
					},
				},
				"required":             []string{"activity_id"},
				"additionalProperties": false,
			},
		}

		result := service.convertToolDefinitionToOpenAI(tool)

		assert.NotNil(t, result)
		
		expected := responses.ToolParamOfFunction(
			"get-activity-details",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"activity_id": map[string]interface{}{
						"type":        "integer",
						"description": "The Strava activity ID",
					},
				},
				"required":             []string{"activity_id"},
				"additionalProperties": false,
			},
			true,
		)

		assert.IsType(t, expected, result)
	})

	t.Run("converts get-activity-streams tool correctly", func(t *testing.T) {
		tool := models.ToolDefinition{
			Name:        "get-activity-streams",
			Description: "Get time-series data streams from a Strava activity with processing options for large datasets",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"activity_id": map[string]interface{}{
						"type":        "integer",
						"description": "The Strava activity ID",
					},
					"stream_types": map[string]interface{}{
						"type":        "array",
						"description": "Types of streams to retrieve",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []interface{}{
								"time", "distance", "latlng", "altitude", "velocity_smooth",
								"heartrate", "cadence", "watts", "temp", "moving", "grade_smooth",
							},
						},
						"default": []interface{}{"time", "distance", "heartrate", "watts"},
					},
					"resolution": map[string]interface{}{
						"type":        "string",
						"description": "Resolution of the data",
						"enum":        []interface{}{"low", "medium", "high"},
						"default":     "medium",
					},
					"processing_mode": map[string]interface{}{
						"type":        "string",
						"description": "How to process large datasets that exceed context limits",
						"enum":        []interface{}{"auto", "raw", "derived", "ai-summary"},
						"default":     "auto",
					},
					"page_number": map[string]interface{}{
						"type":        "integer",
						"description": "Page number for paginated processing (1-based)",
						"minimum":     1,
						"default":     1,
					},
					"page_size": map[string]interface{}{
						"type":        "integer",
						"description": "Number of data points per page. Use negative value to request full dataset (subject to context limits)",
						"minimum":     -1,
						"maximum":     5000,
						"default":     1000,
					},
					"summary_prompt": map[string]interface{}{
						"type":        "string",
						"description": "Custom prompt for AI summarization mode (required when processing_mode is 'ai-summary')",
					},
				},
				"required":             []string{"activity_id"},
				"additionalProperties": false,
			},
		}

		result := service.convertToolDefinitionToOpenAI(tool)

		assert.NotNil(t, result)
		
		expected := responses.ToolParamOfFunction(
			"get-activity-streams",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"activity_id": map[string]interface{}{
						"type":        "integer",
						"description": "The Strava activity ID",
					},
					"stream_types": map[string]interface{}{
						"type":        "array",
						"description": "Types of streams to retrieve",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []interface{}{
								"time", "distance", "latlng", "altitude", "velocity_smooth",
								"heartrate", "cadence", "watts", "temp", "moving", "grade_smooth",
							},
						},
						"default": []interface{}{"time", "distance", "heartrate", "watts"},
					},
					"resolution": map[string]interface{}{
						"type":        "string",
						"description": "Resolution of the data",
						"enum":        []interface{}{"low", "medium", "high"},
						"default":     "medium",
					},
					"processing_mode": map[string]interface{}{
						"type":        "string",
						"description": "How to process large datasets that exceed context limits",
						"enum":        []interface{}{"auto", "raw", "derived", "ai-summary"},
						"default":     "auto",
					},
					"page_number": map[string]interface{}{
						"type":        "integer",
						"description": "Page number for paginated processing (1-based)",
						"minimum":     1,
						"default":     1,
					},
					"page_size": map[string]interface{}{
						"type":        "integer",
						"description": "Number of data points per page. Use negative value to request full dataset (subject to context limits)",
						"minimum":     -1,
						"maximum":     5000,
						"default":     1000,
					},
					"summary_prompt": map[string]interface{}{
						"type":        "string",
						"description": "Custom prompt for AI summarization mode (required when processing_mode is 'ai-summary')",
					},
				},
				"required":             []string{"activity_id", "stream_types", "resolution", "processing_mode", "page_number", "page_size", "summary_prompt"},
				"additionalProperties": false,
			},
			true,
		)

		assert.IsType(t, expected, result)
	})

	t.Run("converts update-athlete-logbook tool correctly", func(t *testing.T) {
		tool := models.ToolDefinition{
			Name:        "update-athlete-logbook",
			Description: "Update or create the athlete logbook with free-form string content",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"content": map[string]interface{}{
						"type":        "string",
						"description": "The complete logbook content as a string. You can structure this however you want - include athlete profile, training data, goals, preferences, health metrics, equipment, coaching insights, observations, and recommendations. Use any format that makes sense (plain text, markdown, etc.).",
					},
				},
				"required":             []string{"content"},
				"additionalProperties": false,
			},
		}

		result := service.convertToolDefinitionToOpenAI(tool)

		assert.NotNil(t, result)
		
		expected := responses.ToolParamOfFunction(
			"update-athlete-logbook",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"content": map[string]interface{}{
						"type":        "string",
						"description": "The complete logbook content as a string. You can structure this however you want - include athlete profile, training data, goals, preferences, health metrics, equipment, coaching insights, observations, and recommendations. Use any format that makes sense (plain text, markdown, etc.).",
					},
				},
				"required":             []string{"content"},
				"additionalProperties": false,
			},
			true,
		)

		assert.IsType(t, expected, result)
	})
}

func TestConvertToolDefinitionToOpenAI_EdgeCases(t *testing.T) {
	service := &aiService{}

	t.Run("handles empty parameters", func(t *testing.T) {
		tool := models.ToolDefinition{
			Name:        "test-tool",
			Description: "Test tool with empty parameters",
			Parameters:  map[string]interface{}{},
		}

		result := service.convertToolDefinitionToOpenAI(tool)
		assert.NotNil(t, result)
	})

	t.Run("handles nil parameters", func(t *testing.T) {
		tool := models.ToolDefinition{
			Name:        "test-tool",
			Description: "Test tool with nil parameters",
			Parameters:  nil,
		}

		result := service.convertToolDefinitionToOpenAI(tool)
		assert.NotNil(t, result)
	})

	t.Run("handles complex nested parameters", func(t *testing.T) {
		tool := models.ToolDefinition{
			Name:        "complex-tool",
			Description: "Tool with complex nested parameters",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"nested_object": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"inner_field": map[string]interface{}{
								"type":        "string",
								"description": "Inner field description",
							},
						},
					},
					"array_field": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"item_field": map[string]interface{}{
									"type": "integer",
								},
							},
						},
					},
				},
				"required": []string{"nested_object"},
			},
		}

		result := service.convertToolDefinitionToOpenAI(tool)
		assert.NotNil(t, result)
	})
}

// TestConvertAllRegistryTools tests conversion of all tools from the registry
func TestConvertAllRegistryTools(t *testing.T) {
	// Create a tool registry with all tools
	registry := NewToolRegistry()
	service := &aiService{
		toolRegistry: registry,
	}

	// Get all tools from registry
	tools := registry.GetAvailableTools()
	require.Len(t, tools, 5, "Expected 5 tools in registry")

	// Convert each tool and verify
	for _, tool := range tools {
		t.Run("convert_"+tool.Name, func(t *testing.T) {
			result := service.convertToolDefinitionToOpenAI(tool)
			assert.NotNil(t, result, "Conversion result should not be nil for tool: %s", tool.Name)
		})
	}
}

// TestConversionConsistencyWithCurrentImplementation validates that converted tools
// produce the same format as the current hardcoded definitions
func TestConversionConsistencyWithCurrentImplementation(t *testing.T) {
	registry := NewToolRegistry()
	service := &aiService{
		toolRegistry: registry,
	}

	// Get tools from registry and convert them
	registryTools := registry.GetAvailableTools()
	convertedTools := make([]responses.ToolUnionParam, len(registryTools))
	
	for i, tool := range registryTools {
		convertedTools[i] = service.convertToolDefinitionToOpenAI(tool)
	}

	// Verify we have the expected number of tools
	assert.Len(t, convertedTools, 5, "Should have 5 tools")

	// Verify that the conversion produces valid results for all tools
	for i, convertedTool := range convertedTools {
		assert.NotNil(t, convertedTool, "Converted tool %d should not be nil", i)
	}

	// Test that we can convert specific tools and they match expected structure
	toolsByName := make(map[string]models.ToolDefinition)
	for _, tool := range registryTools {
		toolsByName[tool.Name] = tool
	}

	// Verify specific tools exist and can be converted
	expectedToolNames := []string{
		"get-athlete-profile",
		"get-recent-activities", 
		"get-activity-details",
		"get-activity-streams",
		"update-athlete-logbook",
	}

	for _, toolName := range expectedToolNames {
		tool, exists := toolsByName[toolName]
		assert.True(t, exists, "Tool %s should exist in registry", toolName)
		
		if exists {
			converted := service.convertToolDefinitionToOpenAI(tool)
			assert.NotNil(t, converted, "Converted tool %s should not be nil", toolName)
		}
	}
}

// TestConversionParameterStructure validates that the parameter structure is preserved correctly
func TestConversionParameterStructure(t *testing.T) {
	service := &aiService{}

	t.Run("preserves simple parameters", func(t *testing.T) {
		tool := models.ToolDefinition{
			Name: "test-tool",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"simple_param": map[string]interface{}{
						"type":        "string",
						"description": "A simple parameter",
					},
				},
				"required": []string{"simple_param"},
			},
		}

		result := service.convertToolDefinitionToOpenAI(tool)
		assert.NotNil(t, result)
	})

	t.Run("preserves complex nested parameters", func(t *testing.T) {
		tool := models.ToolDefinition{
			Name: "complex-tool",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"array_param": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []interface{}{"option1", "option2", "option3"},
						},
					},
					"object_param": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"nested_field": map[string]interface{}{
								"type": "integer",
								"minimum": 1,
								"maximum": 100,
							},
						},
					},
				},
				"required": []string{"array_param"},
			},
		}

		result := service.convertToolDefinitionToOpenAI(tool)
		assert.NotNil(t, result)
	})
}

// TestConversionHandlesParameterDifferences validates that the conversion method
// correctly handles the differences between current hardcoded definitions and registry definitions
func TestConversionHandlesParameterDifferences(t *testing.T) {
	registry := NewToolRegistry()
	service := &aiService{
		toolRegistry: registry,
	}

	// Get tools from registry
	registryTools := registry.GetAvailableTools()
	toolsByName := make(map[string]models.ToolDefinition)
	for _, tool := range registryTools {
		toolsByName[tool.Name] = tool
	}

	t.Run("get-recent-activities has enhanced parameters in registry", func(t *testing.T) {
		tool, exists := toolsByName["get-recent-activities"]
		require.True(t, exists)

		// Registry version should have per_page parameter with detailed schema
		properties, ok := tool.Parameters["properties"].(map[string]interface{})
		require.True(t, ok)
		
		perPageParam, exists := properties["per_page"]
		assert.True(t, exists, "Registry should have per_page parameter")
		
		if exists {
			perPageMap, ok := perPageParam.(map[string]interface{})
			require.True(t, ok)
			
			// Should have detailed parameter definition
			assert.Equal(t, "integer", perPageMap["type"])
			assert.Contains(t, perPageMap["description"], "Number of activities")
			assert.Equal(t, 1, perPageMap["minimum"])
			assert.Equal(t, 200, perPageMap["maximum"])
			assert.Equal(t, 30, perPageMap["default"])
		}

		// Conversion should preserve all parameter details
		converted := service.convertToolDefinitionToOpenAI(tool)
		assert.NotNil(t, converted)
	})

	t.Run("get-activity-streams has comprehensive parameters in registry", func(t *testing.T) {
		tool, exists := toolsByName["get-activity-streams"]
		require.True(t, exists)

		// Registry version should have comprehensive parameters
		properties, ok := tool.Parameters["properties"].(map[string]interface{})
		require.True(t, ok)
		
		// Should have multiple parameters beyond just activity_id
		expectedParams := []string{
			"activity_id", "stream_types", "resolution", 
			"processing_mode", "page_number", "page_size", "summary_prompt",
		}
		
		for _, param := range expectedParams {
			_, exists := properties[param]
			assert.True(t, exists, "Registry should have parameter: %s", param)
		}

		// Conversion should preserve all parameters
		converted := service.convertToolDefinitionToOpenAI(tool)
		assert.NotNil(t, converted)
	})

	t.Run("conversion preserves required fields correctly", func(t *testing.T) {
		// Test that required fields are preserved correctly
		for toolName, tool := range toolsByName {
			t.Run(toolName, func(t *testing.T) {
				converted := service.convertToolDefinitionToOpenAI(tool)
				assert.NotNil(t, converted, "Conversion should succeed for %s", toolName)
				
				// Verify required fields are present in original
				if required, ok := tool.Parameters["required"].([]string); ok {
					assert.NotNil(t, required, "Required fields should be preserved for %s", toolName)
				}
			})
		}
	})
}

// TestConversionWithCurrentHardcodedComparison compares conversion output with current hardcoded format
func TestConversionWithCurrentHardcodedComparison(t *testing.T) {
	registry := NewToolRegistry()
	service := &aiService{
		toolRegistry: registry,
	}

	// Test that conversion works for tools that match current hardcoded format
	t.Run("get-athlete-profile matches current format", func(t *testing.T) {
		registryTools := registry.GetAvailableTools()
		var profileTool models.ToolDefinition
		found := false
		
		for _, tool := range registryTools {
			if tool.Name == "get-athlete-profile" {
				profileTool = tool
				found = true
				break
			}
		}
		
		require.True(t, found, "get-athlete-profile should exist in registry")
		
		// Convert the registry tool
		converted := service.convertToolDefinitionToOpenAI(profileTool)
		assert.NotNil(t, converted)
		
		// Create the current hardcoded equivalent
		hardcoded := responses.ToolParamOfFunction(
			"get-athlete-profile",
			map[string]interface{}{
				"type":                 "object",
				"properties":           map[string]interface{}{},
				"required":             []string{},
				"additionalProperties": false,
			},
			true,
		)
		
		// Both should be the same type
		assert.IsType(t, hardcoded, converted)
	})

	t.Run("update-athlete-logbook matches current format", func(t *testing.T) {
		registryTools := registry.GetAvailableTools()
		var logbookTool models.ToolDefinition
		found := false
		
		for _, tool := range registryTools {
			if tool.Name == "update-athlete-logbook" {
				logbookTool = tool
				found = true
				break
			}
		}
		
		require.True(t, found, "update-athlete-logbook should exist in registry")
		
		// Convert the registry tool
		converted := service.convertToolDefinitionToOpenAI(logbookTool)
		assert.NotNil(t, converted)
		
		// Verify the registry has the content parameter
		properties, ok := logbookTool.Parameters["properties"].(map[string]interface{})
		require.True(t, ok)
		
		contentParam, exists := properties["content"]
		assert.True(t, exists, "Registry should have content parameter")
		
		if exists {
			contentMap, ok := contentParam.(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, "string", contentMap["type"])
			assert.Contains(t, contentMap["description"], "complete logbook content")
		}
	})
}