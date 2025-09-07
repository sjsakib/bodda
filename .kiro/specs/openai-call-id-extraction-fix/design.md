# Design Document

## Overview

The OpenAI Responses API integration currently uses the wrong identifier when processing function call events. The system extracts `event.ItemID` (which corresponds to the `id` field in the event item) instead of the `call_id` field, which is the actual OpenAI tool call identifier needed for proper correlation.

The event structure shows:
```json
{
  "type": "response.output_item.added",
  "item": {
    "id": "fc_68bba53afb688194b3c5b1ba405cd60109ec6f992e6a53b2",
    "call_id": "call_ijVhE1A5JfpvEbYCLs7MVtDk",
    "name": "get-athlete-profile"
  }
}
```

The `call_id` field is the OpenAI tool call identifier that should be used throughout the tool execution pipeline, while the `id` field appears to be an internal item identifier.

## Architecture

The fix involves updating the event processing pipeline to properly extract and use the `call_id` field from nested event structures. The changes will be made in the AI service's event processing methods.

### Current Flow
1. Event received → Only handle specific function call events → Extract `event.ItemID` → Use as tool call identifier
2. Tool call state managed using incorrect ID
3. Tool results created with incorrect `ToolCallID`

### Updated Flow
1. Event received → Handle `response.output_item.added` events → Check if `item.type` is `function_call` → Extract `call_id` from event item → Add to tool call state with logging (Requirements 1.1, 1.4, 3.1)
2. Tool call state managed using correct `call_id` as primary key (Requirement 2.2)
3. Tool results created with correct `ToolCallID` for proper correlation (Requirement 2.1)
4. Function call arguments delta events processed using correct `call_id` (Requirement 1.2)
5. Tool execution receives correct `call_id` for result correlation (Requirement 1.3)

## Components and Interfaces

### Event Processing Components

#### 1. New Event Handler for response.output_item.added
Add a new case in the event processing switch to handle `response.output_item.added` events:

```go
case "response.output_item.added":
    // Handle output item added events - check for function calls
    return s.handleOutputItemAdded(event, state)
```

**Design Rationale**: This addresses Requirement 1.1 by ensuring the system properly handles the specific event type that contains function call items with `call_id` information.

#### 2. Output Item Added Handler
New handler to process function call items and extract call_id:

```go
func (s *aiService) handleOutputItemAdded(event responses.ResponseStreamEventUnion, state *ToolCallState) error {
    // Use event.AsResponseOutputItemAdded() to get the typed event
    // Extract call_id using event.Item.CallID
    // Check if item type is function_call
    // Add call_id to tool call state with comprehensive logging
    // Log extraction process for traceability (Requirement 3.1)
}
```

**Design Rationale**: This handler specifically addresses Requirements 1.1 and 1.4 by extracting the `call_id` from the nested item structure and providing comprehensive logging for traceability.

#### 3. Enhanced Tool Call State Management
Update `ToolCallState` to use `call_id` as the primary key:

```go
type ToolCallState struct {
    toolCalls map[string]*responses.ResponseFunctionToolCall // keyed by call_id
    completed map[string]bool                                // keyed by call_id
}
```

**Design Rationale**: This change ensures Requirements 2.2 and 2.3 are met by using `call_id` as the primary key for all tool call tracking and state management.

## Data Models

### Event Structure Handling
The system will handle the `response.output_item.added` event to extract `call_id`:

1. **response.output_item.added**: Use `event.AsResponseOutputItemAdded()` to get typed event, then extract `call_id` from `event.Item.CallID` (Requirement 1.1)
2. **Function call arguments delta events**: Process using the correct `call_id` for identification (Requirement 1.2)
3. **Existing events**: Continue to use existing logic but reference tool calls by the correct `call_id`

### Tool Call Identification
- **Primary Key**: `call_id` (e.g., "call_ijVhE1A5JfpvEbYCLs7MVtDk") - used for all tool call tracking (Requirement 2.2)
- **Secondary ID**: `id` (e.g., "fc_68bba53afb688194b3c5b1ba405cd60109ec6f992e6a53b2") - for internal tracking only
- **Function Name**: Extracted from event or inferred from arguments

### ToolResult Structure
Updated to ensure proper correlation:
```go
type ToolResult struct {
    ToolCallID string // Must use extracted call_id (Requirement 2.1)
    Content    string
    Error      error
}
```

**Design Rationale**: This ensures that ToolResult objects use the correct `call_id` for proper correlation between tool calls and their results, addressing Requirement 2.1.

## Error Handling

### Call ID Extraction Errors
1. **Missing call_id**: Log warning with detailed event structure and attempt fallback to `event.ItemID` (Requirement 2.4)
2. **Empty call_id**: Log warning with reason and use fallback identification method (Requirement 2.4)
3. **Invalid event structure**: Log detailed error information including full event structure for debugging (Requirement 3.2)

### Validation and Fallback Strategy
The `handleOutputItemAdded` function will:
1. Use `event.AsResponseOutputItemAdded()` to get the typed event
2. Check if the item type is function_call
3. Extract `call_id` using `event.Item.CallID` with comprehensive logging (Requirement 3.1)
4. Add the `call_id` to the tool call state for later correlation
5. Log warnings with detailed context if `call_id` is missing or invalid (Requirements 2.4, 3.2)
6. Use fallback identification methods when necessary and log the strategy used (Requirement 3.3)

### Error Recovery
- Continue processing other events if one call_id extraction fails
- Maintain tool call correlation using available identifiers with fallback logging (Requirement 3.3)
- Provide detailed logging for debugging with appropriate log levels (Requirement 3.2)
- Ensure all completed tool calls are logged with summary including `call_id` values (Requirement 3.4)

## Testing Strategy

### Unit Tests
1. **Call ID Extraction Tests**
   - Test extraction from different event types
   - Test fallback behavior when call_id is missing
   - Test error handling for malformed events

2. **Tool Call State Tests**
   - Test tool call creation with correct call_id
   - Test tool call completion tracking
   - Test validation of tool call identifiers

3. **Integration Tests**
   - Test end-to-end tool call processing with correct call_id
   - Test tool result correlation using call_id
   - Test multi-turn conversation with proper ID tracking

### Test Data
Create mock events with realistic OpenAI response structures:
```go
mockEvent := responses.ResponseStreamEventUnion{
    Type: "response.output_item.added",
    Item: map[string]interface{}{
        "id": "fc_68bba53afb688194b3c5b1ba405cd60109ec6f992e6a53b2",
        "call_id": "call_ijVhE1A5JfpvEbYCLs7MVtDk",
        "name": "get-athlete-profile",
        "type": "function_call",
        "status": "completed"
    }
}
```

### Logging and Monitoring
- Add structured logging for call_id extraction process with appropriate log levels (Requirement 3.1)
- Include call_id in all tool call related log messages for traceability (Requirement 1.4)
- Log detailed error information when extraction fails, including event structure (Requirement 3.2)
- Log fallback identification methods with reasons when used (Requirement 3.3)
- Log completion summaries including all processed call_id values (Requirement 3.4)
- Add metrics for successful vs failed call_id extractions
- Monitor tool call correlation accuracy

## Implementation Notes

### Backward Compatibility
- Existing event handlers continue to work unchanged
- New `response.output_item.added` handler adds call_id information without breaking existing flow
- Tool call state will have both the correct `call_id` and existing item ID for correlation

### Performance Considerations
- Minimal additional processing - just one new event case
- No additional JSON parsing beyond what the SDK already provides
- Simple map lookup to correlate call_ids with existing tool calls

### OpenAI SDK Compatibility
- Uses existing SDK event structures and methods
- No assumptions about internal SDK implementation details
- Handles the event structure as provided by the OpenAI SDK