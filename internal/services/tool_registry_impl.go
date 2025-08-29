package services

import (
	"fmt"

	"bodda/internal/models"
)

// toolRegistry implements the ToolRegistry interface
type toolRegistry struct {
	tools map[string]models.ToolDefinition
}

// NewToolRegistry creates a new tool registry with predefined tools
func NewToolRegistry() ToolRegistry {
	registry := &toolRegistry{
		tools: make(map[string]models.ToolDefinition),
	}
	
	// Initialize with existing tools from AI service
	registry.initializeTools()
	
	return registry
}

// NewToolRegistryWithAIService creates a tool registry that integrates with the AI service
func NewToolRegistryWithAIService(aiService AIService) ToolRegistry {
	registry := &toolRegistry{
		tools: make(map[string]models.ToolDefinition),
	}
	
	// Initialize with tools that match the AI service implementation
	registry.initializeToolsFromAIService()
	
	return registry
}

// initializeTools populates the registry with available tools
func (tr *toolRegistry) initializeTools() {
	tr.initializeToolsFromAIService()
}

// initializeToolsFromAIService populates the registry with tools that match the AI service implementation
func (tr *toolRegistry) initializeToolsFromAIService() {
	// Define get-athlete-profile tool
	tr.tools["get-athlete-profile"] = models.ToolDefinition{
		Name:        "get-athlete-profile",
		Description: "Get the complete athlete profile from Strava including personal information, zones, and stats",
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []interface{}{},
		},
		Examples: []models.ToolExample{
			{
				Description: "Get current athlete profile with all available information",
				Request:     map[string]interface{}{},
				Response: map[string]interface{}{
					"id":         12345,
					"username":   "athlete_username",
					"firstname":  "John",
					"lastname":   "Doe",
					"city":       "San Francisco",
					"state":      "CA",
					"country":    "United States",
					"sex":        "M",
					"premium":    true,
					"created_at": "2020-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z",
					"follower_count": 150,
					"friend_count":   75,
					"athlete_type":   1,
					"date_preference": "%m/%d/%Y",
					"measurement_preference": "feet",
					"clubs": []interface{}{},
					"ftp": 250,
					"weight": 70.5,
				},
			},
		},
	}

	// Define get-recent-activities tool
	tr.tools["get-recent-activities"] = models.ToolDefinition{
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
			"required": []interface{}{},
		},
		Examples: []models.ToolExample{
			{
				Description: "Get default number of recent activities (30)",
				Request:     map[string]interface{}{},
				Response: map[string]interface{}{
					"activities": []interface{}{
						map[string]interface{}{
							"id":               123456789,
							"name":             "Morning Run",
							"distance":         5000.0,
							"moving_time":      1800,
							"elapsed_time":     1900,
							"total_elevation_gain": 50.0,
							"type":             "Run",
							"start_date":       "2024-01-15T07:00:00Z",
							"average_speed":    2.78,
							"max_speed":        4.2,
							"average_heartrate": 150.5,
							"max_heartrate":    175,
						},
					},
					"count": 30,
				},
			},
			{
				Description: "Get last 10 activities",
				Request: map[string]interface{}{
					"per_page": 10,
				},
				Response: map[string]interface{}{
					"activities": "Array of 10 most recent activities",
					"count":      10,
				},
			},
		},
	}

	// Define get-activity-details tool
	tr.tools["get-activity-details"] = models.ToolDefinition{
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
			"required": []interface{}{"activity_id"},
		},
		Examples: []models.ToolExample{
			{
				Description: "Get detailed information for a specific activity",
				Request: map[string]interface{}{
					"activity_id": 123456789,
				},
				Response: map[string]interface{}{
					"id":                   123456789,
					"name":                 "Morning Run",
					"distance":             5000.0,
					"moving_time":          1800,
					"elapsed_time":         1900,
					"total_elevation_gain": 50.0,
					"type":                 "Run",
					"start_date":           "2024-01-15T07:00:00Z",
					"start_latlng":         []float64{37.7749, -122.4194},
					"end_latlng":           []float64{37.7849, -122.4094},
					"average_speed":        2.78,
					"max_speed":            4.2,
					"average_heartrate":    150.5,
					"max_heartrate":        175,
					"average_cadence":      85.2,
					"average_watts":        200.5,
					"kilojoules":           360.9,
					"device_watts":         true,
					"has_heartrate":        true,
					"description":          "Great morning run in the city",
					"calories":             350.0,
					"gear_id":              "b12345",
					"splits_metric":        []interface{}{},
					"splits_standard":      []interface{}{},
					"segment_efforts":      []interface{}{},
				},
			},
		},
	}

	// Define get-activity-streams tool
	tr.tools["get-activity-streams"] = models.ToolDefinition{
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
			"required": []interface{}{"activity_id"},
		},
		Examples: []models.ToolExample{
			{
				Description: "Get default streams with automatic processing",
				Request: map[string]interface{}{
					"activity_id": 123456789,
				},
				Response: map[string]interface{}{
					"streams": map[string]interface{}{
						"time":      []int{0, 1, 2, 3, 4},
						"distance":  []float64{0.0, 2.8, 5.6, 8.4, 11.2},
						"heartrate": []int{120, 125, 130, 135, 140},
						"watts":     []int{180, 190, 200, 210, 220},
					},
					"processing_info": map[string]interface{}{
						"mode":         "auto",
						"total_points": 1800,
						"page_size":    1000,
						"page_number":  1,
					},
				},
			},
			{
				Description: "Get specific streams with high resolution",
				Request: map[string]interface{}{
					"activity_id":  123456789,
					"stream_types": []string{"time", "heartrate", "watts"},
					"resolution":   "high",
				},
				Response: map[string]interface{}{
					"streams": "High resolution time series data for heartrate and watts",
					"processing_info": "Processing metadata",
				},
			},
			{
				Description: "Get paginated stream data",
				Request: map[string]interface{}{
					"activity_id":     123456789,
					"processing_mode": "raw",
					"page_number":     2,
					"page_size":       500,
				},
				Response: map[string]interface{}{
					"streams": "Second page of raw stream data (500 points)",
					"processing_info": "Pagination metadata",
				},
			},
		},
	}

	// Define update-athlete-logbook tool
	tr.tools["update-athlete-logbook"] = models.ToolDefinition{
		Name:        "update-athlete-logbook",
		Description: "Update or create the athlete logbook with free-form string content. You can structure the content however you want - as plain text, markdown, or any format that makes sense for organizing athlete information, training insights, goals, and coaching observations.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"content": map[string]interface{}{
					"type":        "string",
					"description": "The complete logbook content as a string. You can structure this however you want - include athlete profile, training data, goals, preferences, health metrics, equipment, coaching insights, observations, and recommendations. Use any format that makes sense (plain text, markdown, etc.).",
				},
			},
			"required": []interface{}{"content"},
		},
		Examples: []models.ToolExample{
			{
				Description: "Create initial athlete logbook with profile and goals",
				Request: map[string]interface{}{
					"content": `# Athlete Logbook - John Doe

## Profile
- Strava ID: 12345
- Location: San Francisco, CA
- Experience: Intermediate runner, 3 years
- Current FTP: 250W
- Weight: 70.5kg

## Goals
- Complete first marathon in under 3:30
- Improve 5K time to under 20 minutes
- Build consistent training routine

## Training Preferences
- Morning workouts preferred
- Enjoys trail running
- Has access to power meter and heart rate monitor

## Recent Observations
- Good aerobic base but needs speed work
- Consistent with easy runs
- Recovery could be improved

Last updated: 2024-01-15`,
				},
				Response: map[string]interface{}{
					"success": true,
					"message": "Athlete logbook updated successfully",
					"logbook_id": "logbook_12345",
					"updated_at": "2024-01-15T10:30:00Z",
				},
			},
			{
				Description: "Add training analysis to existing logbook",
				Request: map[string]interface{}{
					"content": "Updated logbook with analysis from today's interval session. Athlete showed good power consistency but heart rate recovery needs work.",
				},
				Response: map[string]interface{}{
					"success": true,
					"message": "Logbook updated with new training insights",
				},
			},
		},
	}
}

// GetAvailableTools returns all available tools
func (tr *toolRegistry) GetAvailableTools() []models.ToolDefinition {
	tools := make([]models.ToolDefinition, 0, len(tr.tools))
	for _, tool := range tr.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetToolSchema returns detailed schema for a specific tool
func (tr *toolRegistry) GetToolSchema(toolName string) (*models.ToolSchema, error) {
	tool, exists := tr.tools[toolName]
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", toolName)
	}

	// Extract required and optional parameters from the tool definition
	required := []string{}
	optional := []string{}
	
	if params, ok := tool.Parameters["properties"].(map[string]interface{}); ok {
		requiredFields := []string{}
		if reqArray, ok := tool.Parameters["required"].([]interface{}); ok {
			for _, req := range reqArray {
				if reqStr, ok := req.(string); ok {
					requiredFields = append(requiredFields, reqStr)
				}
			}
		}
		
		for paramName := range params {
			isRequired := false
			for _, req := range requiredFields {
				if req == paramName {
					isRequired = true
					break
				}
			}
			
			if isRequired {
				required = append(required, paramName)
			} else {
				optional = append(optional, paramName)
			}
		}
	}

	schema := &models.ToolSchema{
		Name:        tool.Name,
		Description: tool.Description,
		Parameters:  tool.Parameters,
		Required:    required,
		Optional:    optional,
		Examples:    tool.Examples,
	}

	return schema, nil
}

// ValidateToolCall validates that a tool call has the correct parameters
func (tr *toolRegistry) ValidateToolCall(toolName string, parameters map[string]interface{}) error {
	schema, err := tr.GetToolSchema(toolName)
	if err != nil {
		return err
	}

	// Check required parameters
	for _, required := range schema.Required {
		if _, exists := parameters[required]; !exists {
			return fmt.Errorf("required parameter '%s' is missing", required)
		}
	}

	// Additional validation could be added here for parameter types, ranges, etc.
	
	return nil
}

// IsToolAvailable checks if a tool with the given name exists
func (tr *toolRegistry) IsToolAvailable(toolName string) bool {
	_, exists := tr.tools[toolName]
	return exists
}