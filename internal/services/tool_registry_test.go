package services

import (
	"testing"
)

func TestNewToolRegistry(t *testing.T) {
	registry := NewToolRegistry()
	
	if registry == nil {
		t.Fatal("Expected registry to be created, got nil")
	}
}

func TestToolRegistry_GetAvailableTools(t *testing.T) {
	registry := NewToolRegistry()
	tools := registry.GetAvailableTools()
	
	if len(tools) == 0 {
		t.Error("Expected at least one tool to be available")
	}
	
	// Check that expected tools are present
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
		if !toolNames[expected] {
			t.Errorf("Expected tool '%s' to be available", expected)
		}
	}
}

func TestToolRegistry_GetToolSchema(t *testing.T) {
	registry := NewToolRegistry()
	
	// Test valid tool
	schema, err := registry.GetToolSchema("get-athlete-profile")
	if err != nil {
		t.Errorf("Expected no error for valid tool, got: %v", err)
	}
	
	if schema == nil {
		t.Fatal("Expected schema to be returned")
	}
	
	if schema.Name != "get-athlete-profile" {
		t.Errorf("Expected schema name 'get-athlete-profile', got '%s'", schema.Name)
	}
	
	// Test invalid tool
	_, err = registry.GetToolSchema("invalid-tool")
	if err == nil {
		t.Error("Expected error for invalid tool")
	}
}

func TestToolRegistry_ValidateToolCall(t *testing.T) {
	registry := NewToolRegistry()
	
	// Test valid tool call with required parameters
	err := registry.ValidateToolCall("get-activity-details", map[string]interface{}{
		"activity_id": int64(12345),
	})
	if err != nil {
		t.Errorf("Expected no error for valid tool call, got: %v", err)
	}
	
	// Test tool call missing required parameters
	err = registry.ValidateToolCall("get-activity-details", map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for missing required parameters")
	}
	
	// Test invalid tool
	err = registry.ValidateToolCall("invalid-tool", map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for invalid tool")
	}
}

func TestToolRegistry_IsToolAvailable(t *testing.T) {
	registry := NewToolRegistry()
	
	// Test valid tool
	if !registry.IsToolAvailable("get-athlete-profile") {
		t.Error("Expected 'get-athlete-profile' to be available")
	}
	
	// Test invalid tool
	if registry.IsToolAvailable("invalid-tool") {
		t.Error("Expected 'invalid-tool' to not be available")
	}
}

func TestToolSchema_RequiredOptionalParameters(t *testing.T) {
	registry := NewToolRegistry()
	
	// Test tool with required parameters
	schema, err := registry.GetToolSchema("get-activity-details")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if len(schema.Required) == 0 {
		t.Error("Expected at least one required parameter")
	}
	
	// Check that activity_id is required
	found := false
	for _, req := range schema.Required {
		if req == "activity_id" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'activity_id' to be a required parameter")
	}
	
	// Test tool with optional parameters
	schema, err = registry.GetToolSchema("get-recent-activities")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// per_page should be optional
	found = false
	for _, opt := range schema.Optional {
		if opt == "per_page" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'per_page' to be an optional parameter")
	}
}

func TestToolRegistry_EnhancedExamples(t *testing.T) {
	registry := NewToolRegistry()
	
	// Test that tools have comprehensive examples
	schema, err := registry.GetToolSchema("get-activity-streams")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if len(schema.Examples) == 0 {
		t.Error("Expected at least one example for get-activity-streams")
	}
	
	// Check that examples have proper structure
	for i, example := range schema.Examples {
		if example.Description == "" {
			t.Errorf("Example %d missing description", i)
		}
		if example.Request == nil {
			t.Errorf("Example %d missing request", i)
		}
		if example.Response == nil {
			t.Errorf("Example %d missing response", i)
		}
	}
}

func TestToolRegistry_IntegrationWithAIService(t *testing.T) {
	registry := NewToolRegistry()
	
	// Test that all tools match the AI service implementation
	expectedTools := map[string]bool{
		"get-athlete-profile":     true,
		"get-recent-activities":   true,
		"get-activity-details":    true,
		"get-activity-streams":    true,
		"update-athlete-logbook":  true,
	}
	
	tools := registry.GetAvailableTools()
	
	for _, tool := range tools {
		if !expectedTools[tool.Name] {
			t.Errorf("Unexpected tool found: %s", tool.Name)
		}
		delete(expectedTools, tool.Name)
	}
	
	// Check that all expected tools were found
	for toolName := range expectedTools {
		t.Errorf("Expected tool not found: %s", toolName)
	}
}

func TestToolRegistry_ParameterValidation(t *testing.T) {
	registry := NewToolRegistry()
	
	// Test validation for get-activity-streams with complex parameters
	validParams := map[string]interface{}{
		"activity_id":     int64(123456),
		"stream_types":    []string{"time", "heartrate"},
		"resolution":      "medium",
		"processing_mode": "auto",
		"page_number":     1,
		"page_size":       1000,
	}
	
	err := registry.ValidateToolCall("get-activity-streams", validParams)
	if err != nil {
		t.Errorf("Expected no error for valid parameters, got: %v", err)
	}
	
	// Test validation with missing required parameter
	invalidParams := map[string]interface{}{
		"stream_types": []string{"time", "heartrate"},
		"resolution":   "medium",
	}
	
	err = registry.ValidateToolCall("get-activity-streams", invalidParams)
	if err == nil {
		t.Error("Expected error for missing required parameter 'activity_id'")
	}
}