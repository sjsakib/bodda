# Design Document

## Overview

This design outlines the migration from OpenAI's completion API (`CreateChatCompletionStream`) to the responses API pattern. The OpenAI Go SDK v1.41.1 supports both patterns, but the responses API provides better error handling, more consistent streaming behavior, and access to newer OpenAI features.

The migration will focus on updating the AI service implementation while maintaining the existing interface and functionality for seamless integration with the rest of the application.

## Architecture

### Current Architecture

The current implementation uses:
- `openai.Client.CreateChatCompletionStream()` for streaming responses
- `openai.ChatCompletionRequest` for request configuration
- Manual stream processing with `stream.Recv()` loops
- Custom error handling for streaming errors

### Target Architecture

The new implementation will use:
- OpenAI responses API pattern with structured response handling
- Improved streaming with better error recovery
- Enhanced tool call processing with responses API
- Standardized error handling patterns

## Components and Interfaces

### AIService Interface

The `AIService` interface will remain unchanged to maintain compatibility:

```go
type AIService interface {
    ProcessMessage(ctx context.Context, msgCtx *MessageContext) (<-chan string, error)
    ProcessMessageSync(ctx context.Context, msgCtx *MessageContext) (string, error)
    // ... existing tool execution methods
}
```

### Core Components to Update

#### 1. AI Service Implementation (`aiService`)

**Current Structure:**
```go
type aiService struct {
    client *openai.Client
    // ... other fields
}
```

**Updated Structure:**
- Same struct, but methods will use responses API patterns
- Enhanced error handling with responses API error types
- Improved streaming logic with responses API

#### 2. Streaming Response Handler

**Current Pattern:**
```go
stream, err := s.client.CreateChatCompletionStream(ctx, req)
for {
    response, err := stream.Recv()
    // Manual processing
}
```

**New Pattern:**
```go
// Use responses API with structured response handling
// Enhanced streaming with better error recovery
// Improved tool call processing
```

#### 3. Tool Call Processing

**Current Implementation:**
- Manual parsing of tool calls from streaming deltas
- Custom tool call accumulation logic
- Basic error recovery

**Enhanced Implementation:**
- Structured tool call handling with responses API
- Better error handling and recovery
- Better tool call state management

## Data Models

### Request/Response Models

The migration will update how we interact with OpenAI models while maintaining the same data flow:

#### Current Models Used:
- `openai.ChatCompletionRequest`
- `openai.ChatCompletionMessage`
- `openai.ToolCall`
- `openai.Tool`

#### Updated Usage:
- Same models but used with responses API patterns
- Enhanced error handling structures
- Improved streaming response types

### Internal Data Models

No changes to internal models:
- `MessageContext` - remains unchanged
- `ToolResult` - remains unchanged  
- `IterativeProcessor` - enhanced but interface unchanged

## Error Handling

### Current Error Handling

```go
func (s *aiService) handleStreamingError(err error, processor *IterativeProcessor, responseChan chan<- string) error {
    // Basic error categorization
    // Simple retry logic
}
```

### Enhanced Error Handling

The responses API provides better error categorization and handling:

1. **Structured Error Types**: Use responses API error types for better error classification
2. **Better Logging**: Structured error logging with context
3. **Graceful Degradation**: Better fallback mechanisms for partial failures

### Error Categories

1. **Network Errors**: Connection issues, timeouts
2. **API Errors**: Rate limits, quota exceeded, invalid requests
3. **Streaming Errors**: Stream interruption, parsing errors
4. **Tool Execution Errors**: Tool call failures, timeout errors

## Testing Strategy

### Unit Tests

1. **AI Service Tests**: Update existing tests to work with responses API
2. **Streaming Tests**: Enhanced streaming behavior validation
3. **Tool Call Tests**: Improved tool execution testing
4. **Error Handling Tests**: Comprehensive error scenario coverage

### Integration Tests

1. **End-to-End Flow**: Complete message processing with responses API
2. **Tool Execution Flow**: Multi-round tool calling validation
3. **Error Recovery**: Error handling and recovery testing
4. **Performance**: Streaming performance validation

### Test Structure

```go
func TestAIService_ProcessMessage_ResponsesAPI(t *testing.T) {
    // Test streaming with responses API
}

func TestAIService_ToolExecution_ResponsesAPI(t *testing.T) {
    // Test tool calling with responses API
}

func TestAIService_ErrorHandling_ResponsesAPI(t *testing.T) {
    // Test error scenarios with responses API
}
```

## Implementation Phases

### Phase 1: Core Migration
- Update streaming logic to use responses API
- Maintain existing functionality and interface
- Basic error handling migration

### Phase 2: Enhanced Features
- Improved error handling and recovery
- Enhanced tool call processing
- Better streaming performance

### Phase 3: Optimization
- Performance optimizations
- Advanced error recovery
- Monitoring and observability improvements

## Migration Strategy

### Approach

1. **In-Place Migration**: Update existing `aiService` implementation
2. **Interface Preservation**: Keep existing interfaces unchanged
3. **Gradual Enhancement**: Implement improvements incrementally
4. **Comprehensive Testing**: Validate each change thoroughly

### Rollback Plan

- Keep current implementation patterns as reference
- Maintain existing error handling as fallback
- Comprehensive test coverage to catch regressions
- Monitoring to detect issues early

## Performance Considerations

### Streaming Performance
- Responses API may have different streaming characteristics
- Monitor latency and throughput during migration
- Optimize buffer sizes and processing patterns

### Memory Usage
- Responses API may have different memory patterns
- Monitor memory usage during streaming
- Optimize tool call accumulation and processing

### Error Recovery
- Improved error recovery may reduce failed requests
- Better error handling should improve success rates
- Monitor error rates and recovery effectiveness

## Security Considerations

### API Key Management
- Same OpenAI API key configuration
- No changes to authentication patterns
- Maintain existing security practices

### Data Handling
- Same data flow and processing patterns
- No changes to data retention or logging
- Maintain existing privacy protections

## Monitoring and Observability

### Metrics to Track
- Response latency and throughput
- Error rates by category
- Tool execution success rates
- Streaming performance metrics

### Logging Enhancements
- Structured logging with responses API context
- Better error categorization in logs
- Enhanced debugging information

### Alerting
- Monitor for increased error rates
- Alert on streaming performance degradation
- Track tool execution failures