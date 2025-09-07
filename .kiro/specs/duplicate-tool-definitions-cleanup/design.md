# Design Document

## Overview

This design consolidates duplicate tool definitions by making the tool registry the single source of truth for all tool metadata and schemas. The AI service will be refactored to retrieve tool definitions from the registry instead of maintaining its own duplicate definitions.

## Architecture

### Current State
- **AI Service (`ai.go`)**: Contains tool definitions in `getTools()` method using OpenAI's `responses.ToolUnionParam` format
- **Tool Registry (`tool_registry_impl.go`)**: Contains duplicate tool definitions in `models.ToolDefinition` format
- **Duplication**: Same 5 tools defined in both places with slightly different formats

### Target State
- **Tool Registry**: Single source of truth for all tool definitions
- **AI Service**: Retrieves tool definitions from registry and converts to OpenAI format as needed
- **Consistency**: All components use the same underlying tool definitions

## Components and Interfaces

### 1. Enhanced Tool Registry Interface

The existing `ToolRegistry` interface already provides the necessary methods:
- `GetAvailableTools() []models.ToolDefinition`
- `GetToolSchema(toolName string) (*models.ToolSchema, error)`
- `IsToolAvailable(toolName string) bool`

### 2. AI Service Refactoring

**Modified Methods:**
- `getTools()`: Will call `registry.GetAvailableTools()` and convert to OpenAI format
- Add new method: `convertToolDefinitionToOpenAI(tool models.ToolDefinition) responses.ToolUnionParam`

**Dependencies:**
- Add `ToolRegistry` as a dependency to `aiService` struct
- Update constructor to accept registry parameter

### 3. Tool Definition Conversion

**Conversion Logic:**
```go
func (s *aiService) convertToolDefinitionToOpenAI(tool models.ToolDefinition) responses.ToolUnionParam {
    return responses.ToolParamOfFunction(
        tool.Name,
        tool.Description,
        tool.Parameters,
    )
}
```

## Data Models

### Existing Models (No Changes Required)

**`models.ToolDefinition`** - Already contains all necessary fields:
- `Name string`
- `Description string` 
- `Parameters map[string]interface{}`
- `Examples []models.ToolExample`

**`models.ToolSchema`** - Used for detailed schema information
**`responses.ToolUnionParam`** - OpenAI format for API calls

## Error Handling

### Registry Dependency Errors
- **Scenario**: Tool registry fails to initialize or return tools
- **Handling**: AI service falls back to empty tool list and logs error
- **Recovery**: System continues to function but without tool capabilities

### Conversion Errors
- **Scenario**: Tool definition cannot be converted to OpenAI format
- **Handling**: Skip problematic tool, log warning, continue with remaining tools
- **Recovery**: Partial tool functionality maintained

### Backward Compatibility
- **Scenario**: Existing code expects old tool format
- **Handling**: Maintain existing public interfaces, only change internal implementation
- **Recovery**: No breaking changes to external APIs

## Testing Strategy

### Unit Tests
1. **Tool Registry Tests**: Verify all tools are properly defined with correct schemas
2. **AI Service Tests**: Test tool retrieval and conversion to OpenAI format
3. **Conversion Tests**: Validate `convertToolDefinitionToOpenAI` method
4. **Error Handling Tests**: Test fallback behavior when registry is unavailable

### Integration Tests
1. **End-to-End Tool Execution**: Verify tools work through both AI service and direct execution
2. **API Compatibility**: Ensure all existing endpoints continue to work
3. **Schema Consistency**: Validate tool schemas are identical across components

### Regression Tests
1. **Existing Test Suite**: All current tests must continue to pass
2. **Tool Execution Tests**: Verify all 5 tools execute correctly
3. **Parameter Validation**: Ensure parameter validation works consistently

## Implementation Phases

### Phase 1: Preparation
- Add registry dependency to AI service constructor
- Create tool definition conversion method
- Add comprehensive unit tests for conversion logic

### Phase 2: Core Refactoring  
- Replace hardcoded tool definitions in `getTools()` with registry calls
- Update AI service initialization to accept registry parameter
- Modify dependency injection in main application

### Phase 3: Cleanup and Validation
- Remove duplicate tool definitions from AI service
- Run comprehensive test suite
- Validate all tool execution paths work correctly
- Performance testing to ensure no regression

## Migration Strategy

### Backward Compatibility
- All existing public APIs remain unchanged
- Tool execution behavior remains identical
- No configuration changes required

### Rollback Plan
- Keep original `getTools()` method commented out during initial deployment
- Monitor for any issues in tool execution
- Quick rollback available by reverting to hardcoded definitions if needed

### Validation Criteria
- All existing tests pass
- Tool execution latency remains within acceptable bounds
- No errors in tool schema validation
- Consistent tool behavior across all execution paths