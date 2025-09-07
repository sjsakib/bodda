# Design Document

## Overview

This design outlines the migration from OpenAI's Chat Completions API to the Responses API by replacing the third-party SDK (`github.com/sashabaranov/go-openai`) with OpenAI's official SDK (`github.com/openai/openai-go`). The Responses API provides enhanced error handling, better streaming capabilities, access to advanced features like reasoning and computer use, and future-proof architecture.

The migration involves both SDK replacement and API pattern changes while maintaining the existing interface and functionality for seamless integration with the rest of the application.

## Architecture

### Current Architecture

The current implementation uses:

- **Third-party SDK**: `github.com/sashabaranov/go-openai v1.41.1`
- **API**: Chat Completions API (`/v1/chat/completions`)
- **Streaming**: `openai.Client.CreateChatCompletionStream()`
- **Request Type**: `openai.ChatCompletionRequest`
- **Response Processing**: Manual delta processing with `stream.Recv()` loops
- **Error Handling**: Basic HTTP error handling

### Target Architecture

The new implementation will use:

- **Official SDK**: `github.com/openai/openai-go v1.12.0`
- **API**: Responses API (`/v1/responses`)
- **Streaming**: `client.Responses.NewStreaming()`
- **Request Type**: `responses.ResponseNewParams`
- **Response Processing**: Event-based processing with `ResponseStreamEventUnion`
- **Error Handling**: Enhanced error categorization and recovery
- **Advanced Features**: Access to reasoning, computer use, and future capabilities

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
import "github.com/sashabaranov/go-openai"

type aiService struct {
    client *openai.Client  // Third-party SDK client
    // ... other fields
}
```

**Intermediate Structure (Parallel Implementation):**

```go
import (
    sashabaranov "github.com/sashabaranov/go-openai"
    openai "github.com/openai/openai-go"
)

type aiService struct {
    legacyClient   *sashabaranov.Client  // Existing third-party SDK client
    officialClient *openai.Client       // New official SDK client
    useResponsesAPI bool                 // Feature flag
    // ... other fields (unchanged)
}
```

**Final Structure:**

```go
import "github.com/openai/openai-go"

type aiService struct {
    client *openai.Client  // Official OpenAI SDK client only
    // ... other fields (unchanged)
}
```

**Key Changes:**

- Add official SDK alongside existing SDK (parallel implementation)
- Implement feature flag to switch between implementations
- Gradually migrate methods to use Responses API
- Remove third-party SDK only after full validation

#### 2. Streaming Response Handler

**Current Pattern (Chat Completions API):**

```go
// Third-party SDK with Chat Completions API
req := openai.ChatCompletionRequest{
    Model: openai.GPT4,
    Messages: messages,
    Tools: tools,
    Stream: true,
}
stream, err := s.client.CreateChatCompletionStream(ctx, req)
for {
    response, err := stream.Recv()
    if err != nil {
        if err == io.EOF {
            break
        }
        return err
    }
    delta := response.Choices[0].Delta
    // Process delta content and tool calls
}
```

**New Pattern (Responses API):**

```go
// Official SDK with Responses API
params := responses.ResponseNewParams{
    Model: responses.ResponsesModelO1Pro,
    Input: []responses.ResponseInputItemUnionParam{
        responses.ResponseInputItemParamOfMessage(content, responses.EasyInputMessageRoleUser),
    },
    Tools: tools,
}
stream := s.client.Responses.NewStreaming(ctx, params)
for stream.Next() {
    event := stream.Current()
    switch event := event.(type) {
    case responses.ResponseTextDeltaEvent:
        // Handle text content
        responseChan <- event.Delta
    case responses.ResponseFunctionCallArgumentsDeltaEvent:
        // Handle tool call arguments
    case responses.ResponseCompletedEvent:
        // Handle completion
        break
    }
}
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

The migration will replace third-party SDK models with official SDK models:

#### Current Models (Third-party SDK):

```go
// github.com/sashabaranov/go-openai
openai.ChatCompletionRequest
openai.ChatCompletionMessage
openai.ChatCompletionStreamResponse
openai.ToolCall
openai.Tool
```

#### New Models (Official SDK):

```go
// github.com/openai/openai-go
responses.ResponseNewParams
responses.ResponseInputItemUnionParam
responses.ResponseStreamEventUnion
responses.ResponseFunctionToolCall
responses.ToolUnionParam
```

#### Model Mapping:

| Current (Third-party)          | New (Official)                | Purpose                  |
| ------------------------------ | ----------------------------- | ------------------------ |
| `ChatCompletionRequest`        | `ResponseNewParams`           | Request configuration    |
| `ChatCompletionMessage`        | `ResponseInputItemUnionParam` | Input messages           |
| `ChatCompletionStreamResponse` | `ResponseStreamEventUnion`    | Streaming events         |
| `ToolCall`                     | `ResponseFunctionToolCall`    | Tool call representation |
| `Tool`                         | `ToolUnionParam`              | Tool definitions         |

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

### Phase 1: Parallel SDK Addition

- Add `github.com/openai/openai-go` alongside existing `github.com/sashabaranov/go-openai`
- Create new client initialization for official SDK
- Maintain existing functionality and interface
- Ensure compilation continues to work

### Phase 2: Responses API Implementation

- Implement new methods using Responses API alongside existing Chat Completions methods
- Create event-based streaming processing in parallel to delta-based processing
- Implement new tool call handling for event structure
- Add feature flag to switch between implementations

### Phase 3: Testing and Validation

- Comprehensive testing of both implementations side-by-side
- Performance comparison and validation
- Feature parity verification
- Gradual rollout with monitoring

### Phase 4: Migration Completion

- Switch default implementation to Responses API
- Remove third-party SDK dependency
- Clean up old implementation code
- Documentation and knowledge transfer

## Migration Strategy

### Approach

1. **Parallel Implementation**: Add official SDK alongside existing third-party SDK
2. **Gradual Migration**: Implement Responses API methods while keeping existing ones
3. **Interface Preservation**: Keep existing `AIService` interface unchanged
4. **Feature Flag**: Use configuration to switch between implementations
5. **Comprehensive Testing**: Validate functionality parity before switching
6. **Clean Removal**: Remove third-party SDK only after full validation

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
