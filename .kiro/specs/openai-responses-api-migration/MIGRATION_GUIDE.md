# OpenAI Responses API Migration Guide

## Overview

This document provides a comprehensive guide for the migration from OpenAI's Chat Completions API to the Responses API, including the transition from the third-party SDK (`github.com/sashabaranov/go-openai`) to OpenAI's official SDK (`github.com/openai/openai-go`).

## Table of Contents

1. [Migration Process Overview](#migration-process-overview)
2. [SDK Differences](#sdk-differences)
3. [API Architectural Differences](#api-architectural-differences)
4. [Implementation Changes](#implementation-changes)
5. [Troubleshooting Guide](#troubleshooting-guide)
6. [Legacy Removal Process](#legacy-removal-process)
7. [Final System Architecture](#final-system-architecture)
8. [Performance Considerations](#performance-considerations)
9. [Best Practices](#best-practices)

## Migration Process Overview

The migration was completed in four main phases:

### Phase 1: Parallel SDK Addition
- Added official OpenAI SDK alongside existing third-party SDK
- Maintained existing functionality during transition
- Implemented feature flags for controlled rollout

### Phase 2: Responses API Implementation
- Implemented new methods using Responses API patterns
- Created event-based streaming processing
- Enhanced tool call handling and error management

### Phase 3: Testing and Validation
- Comprehensive testing of both implementations
- Performance comparison and validation
- Feature parity verification with gradual rollout

### Phase 4: Migration Completion
- Switched to Responses API as default
- Removed third-party SDK dependency
- Cleaned up legacy implementation code

## SDK Differences

### Third-Party SDK vs Official SDK

| Aspect | Third-Party SDK | Official SDK |
|--------|----------------|--------------|
| **Package** | `github.com/sashabaranov/go-openai` | `github.com/openai/openai-go` |
| **Maintenance** | Community maintained | OpenAI maintained |
| **API Coverage** | Limited to available endpoints | Full API coverage including latest features |
| **Error Handling** | Basic HTTP error handling | Structured error types and enhanced handling |
| **Type Safety** | Manual type definitions | Auto-generated from OpenAI specs |
| **Future Support** | Dependent on community updates | Direct OpenAI support and updates |

### Key Import Changes

**Before (Third-Party SDK):**
```go
import "github.com/sashabaranov/go-openai"

client := openai.NewClient(apiKey)
```

**After (Official SDK):**
```go
import "github.com/openai/openai-go"

client := openai.NewClient(
    option.WithAPIKey(apiKey),
)
```

## API Architectural Differences

### Chat Completions API vs Responses API

| Feature | Chat Completions API | Responses API |
|---------|---------------------|---------------|
| **Endpoint** | `/v1/chat/completions` | `/v1/responses` |
| **Streaming Pattern** | Delta-based streaming | Event-based streaming |
| **Request Structure** | `ChatCompletionRequest` | `ResponseNewParams` |
| **Response Processing** | Manual delta accumulation | Structured event handling |
| **Tool Calls** | Delta-based tool call parsing | Event-driven tool call processing |
| **Error Recovery** | Basic retry mechanisms | Enhanced error categorization |
| **Advanced Features** | Limited access | Full access to reasoning, computer use, etc. |

### Request Structure Comparison

**Chat Completions API Request:**
```go
req := openai.ChatCompletionRequest{
    Model: openai.GPT4,
    Messages: []openai.ChatCompletionMessage{
        {
            Role:    openai.ChatMessageRoleUser,
            Content: "Hello, world!",
        },
    },
    Tools: []openai.Tool{
        {
            Type: openai.ToolTypeFunction,
            Function: &openai.FunctionDefinition{
                Name:        "get_weather",
                Description: "Get weather information",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "location": map[string]interface{}{
                            "type": "string",
                        },
                    },
                },
            },
        },
    },
    Stream: true,
}
```

**Responses API Request:**
```go
params := responses.ResponseNewParams{
    Model: responses.ResponsesModelO1Pro,
    Input: []responses.ResponseInputItemUnionParam{
        responses.ResponseInputItemParamOfMessage(
            "Hello, world!",
            responses.EasyInputMessageRoleUser,
        ),
    },
    Tools: []responses.ToolUnionParam{
        responses.ToolUnionParamOfFunction(responses.FunctionToolParam{
            Name:        "get_weather",
            Description: "Get weather information",
            Parameters: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "location": map[string]interface{}{
                        "type": "string",
                    },
                },
            },
        }),
    },
}
```

### Streaming Processing Comparison

**Chat Completions API Streaming:**
```go
stream, err := client.CreateChatCompletionStream(ctx, req)
if err != nil {
    return err
}
defer stream.Close()

for {
    response, err := stream.Recv()
    if err != nil {
        if err == io.EOF {
            break
        }
        return err
    }
    
    delta := response.Choices[0].Delta
    if delta.Content != "" {
        responseChan <- delta.Content
    }
    
    // Manual tool call delta processing
    if len(delta.ToolCalls) > 0 {
        // Accumulate tool call deltas manually
    }
}
```

**Responses API Streaming:**
```go
stream := client.Responses.NewStreaming(ctx, params)

for stream.Next() {
    event := stream.Current()
    
    switch event := event.(type) {
    case responses.ResponseTextDeltaEvent:
        responseChan <- event.Delta
        
    case responses.ResponseFunctionCallArgumentsDeltaEvent:
        // Structured tool call event handling
        
    case responses.ResponseCompletedEvent:
        // Handle completion
        break
    }
}

if err := stream.Err(); err != nil {
    return err
}
```

## Implementation Changes

### Core Service Changes

#### AIService Structure Evolution

**Before:**
```go
type aiService struct {
    client          *openai.Client
    config          *config.Config
    toolRegistry    services.ToolRegistry
    toolExecutor    services.ToolExecutor
    summaryProcessor *SummaryProcessor
}
```

**After:**
```go
type aiService struct {
    client          *openai.Client  // Now official SDK
    config          *config.Config
    toolRegistry    services.ToolRegistry
    toolExecutor    services.ToolExecutor
    summaryProcessor *SummaryProcessor
}
```

#### Method Signature Changes

The public interface remained unchanged, but internal implementations were completely rewritten:

```go
// Interface unchanged - maintains backward compatibility
func (s *aiService) ProcessMessage(ctx context.Context, msgCtx *MessageContext) (<-chan string, error)
```

### Tool Call Processing Changes

**Before (Delta-based):**
```go
func (s *aiService) parseToolCallsFromDelta(delta openai.ChatCompletionStreamChoiceDelta) []ToolCall {
    // Manual delta accumulation
    // Complex state management
    // Error-prone parsing
}
```

**After (Event-based):**
```go
func (s *aiService) parseToolCallsFromEvents(events []responses.ResponseStreamEventUnion) []ToolCall {
    // Structured event processing
    // Built-in state management
    // Robust error handling
}
```

### Error Handling Improvements

**Before:**
```go
func (s *aiService) handleStreamingError(err error) error {
    // Basic error categorization
    if strings.Contains(err.Error(), "rate limit") {
        return ErrRateLimit
    }
    return err
}
```

**After:**
```go
func (s *aiService) handleResponsesAPIError(err error) error {
    // Structured error handling with official SDK types
    var apiErr *openai.Error
    if errors.As(err, &apiErr) {
        switch apiErr.Code {
        case "rate_limit_exceeded":
            return ErrRateLimit
        case "insufficient_quota":
            return ErrQuotaExceeded
        default:
            return fmt.Errorf("API error: %s", apiErr.Message)
        }
    }
    return err
}
```

## Troubleshooting Guide

### Common Issues and Solutions

#### 1. Import Errors After Migration

**Problem:** Import statements not found after removing third-party SDK.

**Solution:**
```bash
# Clean module cache
go clean -modcache
go mod tidy
go mod download
```

#### 2. Type Conversion Issues

**Problem:** Type mismatches between old and new SDK types.

**Solution:** Use the conversion helpers:
```go
// Convert message format
func convertToResponseInput(msg ChatMessage) responses.ResponseInputItemUnionParam {
    return responses.ResponseInputItemParamOfMessage(
        msg.Content,
        responses.EasyInputMessageRole(msg.Role),
    )
}
```

#### 3. Streaming Connection Issues

**Problem:** Stream connections failing or hanging.

**Symptoms:**
- Timeouts during streaming
- Incomplete responses
- Connection drops

**Solution:**
```go
// Add proper context timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Check for stream errors
for stream.Next() {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Process event
    }
}
```

#### 4. Tool Call Processing Errors

**Problem:** Tool calls not being parsed correctly.

**Symptoms:**
- Missing tool call arguments
- Incomplete tool call data
- Tool execution failures

**Solution:**
```go
// Ensure proper event type checking
switch event := event.(type) {
case responses.ResponseFunctionCallArgumentsDeltaEvent:
    if event.Name != "" && event.Arguments != "" {
        // Process complete tool call
    }
case responses.ResponseFunctionCallCompletedEvent:
    // Handle tool call completion
}
```

#### 5. Authentication Issues

**Problem:** API key not working with official SDK.

**Solution:**
```go
// Ensure proper client initialization
client := openai.NewClient(
    option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
    option.WithBaseURL("https://api.openai.com/v1"), // Optional: explicit base URL
)
```

### Debugging Tips

#### Enable Debug Logging

```go
// Add debug logging for streaming events
for stream.Next() {
    event := stream.Current()
    log.Printf("Received event type: %T", event)
    
    switch event := event.(type) {
    case responses.ResponseTextDeltaEvent:
        log.Printf("Text delta: %s", event.Delta)
    }
}
```

#### Monitor API Usage

```go
// Add metrics collection
func (s *aiService) ProcessMessage(ctx context.Context, msgCtx *MessageContext) (<-chan string, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        log.Printf("ProcessMessage took %v", duration)
    }()
    
    // Implementation...
}
```

### Performance Troubleshooting

#### Slow Streaming Response

1. **Check network latency:** Use `ping api.openai.com`
2. **Monitor buffer sizes:** Adjust channel buffer sizes if needed
3. **Profile memory usage:** Use `go tool pprof` to identify bottlenecks

#### High Memory Usage

1. **Check for goroutine leaks:** Use `go tool pprof http://localhost:6060/debug/pprof/goroutine`
2. **Monitor channel usage:** Ensure channels are properly closed
3. **Review event accumulation:** Avoid storing unnecessary event history

## Legacy Removal Process

### Step-by-Step Removal Guide

#### 1. Verify Responses API Functionality

Before removing legacy code, ensure:
- All tests pass with Responses API
- Production monitoring shows stable performance
- No error rate increases observed

#### 2. Remove Feature Flags

```go
// Remove configuration options
type Config struct {
    // Remove: UseResponsesAPI bool
    OpenAIAPIKey string
    // ... other config
}
```

#### 3. Remove Third-Party SDK Dependency

```bash
# Remove from go.mod
go mod edit -droprequire github.com/sashabaranov/go-openai
go mod tidy
```

#### 4. Clean Up Import Statements

```go
// Remove old imports
// import sashabaranov "github.com/sashabaranov/go-openai"

// Keep only official SDK
import "github.com/openai/openai-go"
```

#### 5. Remove Legacy Methods

Remove all methods with "Legacy" or "ChatCompletion" in their names:
- `processMessageWithChatCompletion`
- `handleChatCompletionError`
- `parseToolCallsFromDelta`

#### 6. Update Configuration Files

Remove legacy configuration options from:
- Environment variable documentation
- Configuration struct definitions
- Default configuration values

#### 7. Clean Up Test Files

Remove or update tests that reference legacy implementations:
- Remove mock objects for third-party SDK
- Update test assertions for new response formats
- Remove feature flag test cases

### Verification Checklist

- [ ] All imports reference official SDK only
- [ ] No references to `sashabaranov` package
- [ ] All tests pass without legacy dependencies
- [ ] Configuration files updated
- [ ] Documentation reflects new implementation
- [ ] No dead code or unused methods remain

## Final System Architecture

### Current Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     AI Service Layer                        │
├─────────────────────────────────────────────────────────────┤
│  ProcessMessage() │ ProcessMessageSync() │ Tool Execution   │
├─────────────────────────────────────────────────────────────┤
│                OpenAI Official SDK Client                   │
├─────────────────────────────────────────────────────────────┤
│                   Responses API Layer                       │
├─────────────────────────────────────────────────────────────┤
│  Event Processing │ Stream Handling │ Error Management     │
├─────────────────────────────────────────────────────────────┤
│                     OpenAI API                             │
│                 /v1/responses endpoint                      │
└─────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

#### AIService
- **Purpose:** Main interface for AI operations
- **Responsibilities:**
  - Message processing coordination
  - Tool execution orchestration
  - Error handling and recovery
  - Response streaming management

#### Official SDK Client
- **Purpose:** OpenAI API communication
- **Responsibilities:**
  - API request/response handling
  - Authentication management
  - Connection pooling and retries
  - Type-safe API interactions

#### Responses API Layer
- **Purpose:** Event-based response processing
- **Responsibilities:**
  - Stream event handling
  - Tool call event processing
  - Response completion detection
  - Error event management

### Data Flow

```
User Request
    ↓
MessageContext Creation
    ↓
AIService.ProcessMessage()
    ↓
Request Conversion (ResponseNewParams)
    ↓
Official SDK Client
    ↓
Responses API (/v1/responses)
    ↓
Event Stream Processing
    ↓
Tool Call Detection & Execution
    ↓
Response Channel Output
    ↓
User Response
```

### Integration Points

#### Database Layer
- **Connection:** Unchanged - uses existing repository pattern
- **Data Models:** No changes to internal data structures
- **Transactions:** Same transaction handling patterns

#### Tool Registry
- **Interface:** Unchanged - maintains existing tool definitions
- **Execution:** Enhanced error handling and logging
- **Results:** Same result format and processing

#### Configuration
- **Structure:** Simplified - removed feature flags and legacy options
- **Environment:** Same environment variable patterns
- **Validation:** Enhanced validation for official SDK requirements

## Performance Considerations

### Streaming Performance

#### Latency Improvements
- **Event-based processing:** Reduced parsing overhead
- **Structured responses:** Faster type conversion
- **Better buffering:** Optimized channel usage

#### Throughput Enhancements
- **Connection pooling:** Official SDK handles connection reuse
- **Request batching:** Better request optimization
- **Error recovery:** Faster failure recovery

### Memory Usage

#### Optimizations
- **Event streaming:** Lower memory footprint per request
- **Garbage collection:** Reduced object allocation
- **Buffer management:** More efficient buffer usage

#### Monitoring
```go
// Memory usage tracking
var memStats runtime.MemStats
runtime.ReadMemStats(&memStats)
log.Printf("Memory usage: %d KB", memStats.Alloc/1024)
```

### Error Recovery Performance

#### Improvements
- **Structured errors:** Faster error categorization
- **Retry logic:** More intelligent retry strategies
- **Circuit breaker:** Better failure isolation

## Best Practices

### Code Organization

#### Service Layer
```go
// Keep interfaces clean and focused
type AIService interface {
    ProcessMessage(ctx context.Context, msgCtx *MessageContext) (<-chan string, error)
    ProcessMessageSync(ctx context.Context, msgCtx *MessageContext) (string, error)
}

// Implementation should be internal
type aiService struct {
    client *openai.Client
    // ... other dependencies
}
```

#### Error Handling
```go
// Use structured error handling
func (s *aiService) handleError(err error) error {
    var apiErr *openai.Error
    if errors.As(err, &apiErr) {
        // Handle specific API errors
        return s.categorizeAPIError(apiErr)
    }
    
    // Handle other error types
    return fmt.Errorf("unexpected error: %w", err)
}
```

### Testing Strategies

#### Unit Testing
```go
// Mock the official SDK client
type mockOpenAIClient struct {
    responses []responses.ResponseStreamEventUnion
}

func (m *mockOpenAIClient) NewStreaming(ctx context.Context, params responses.ResponseNewParams) *responses.Stream {
    // Return mock stream
}
```

#### Integration Testing
```go
// Test with real API calls in integration tests
func TestAIService_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Use real client with test API key
    client := openai.NewClient(option.WithAPIKey(testAPIKey))
    service := NewAIService(client, config)
    
    // Test real API interaction
}
```

### Monitoring and Observability

#### Metrics Collection
```go
// Track key metrics
type Metrics struct {
    RequestCount    prometheus.Counter
    ResponseLatency prometheus.Histogram
    ErrorRate       prometheus.Counter
}

func (s *aiService) recordMetrics(duration time.Duration, err error) {
    s.metrics.RequestCount.Inc()
    s.metrics.ResponseLatency.Observe(duration.Seconds())
    
    if err != nil {
        s.metrics.ErrorRate.Inc()
    }
}
```

#### Logging Best Practices
```go
// Use structured logging
log.WithFields(log.Fields{
    "request_id": requestID,
    "model":      params.Model,
    "duration":   duration,
}).Info("Request completed")
```

### Security Considerations

#### API Key Management
```go
// Use secure configuration
client := openai.NewClient(
    option.WithAPIKey(config.OpenAIAPIKey), // From secure config
)
```

#### Request Validation
```go
// Validate inputs before API calls
func (s *aiService) validateRequest(msgCtx *MessageContext) error {
    if msgCtx == nil {
        return errors.New("message context is required")
    }
    
    if len(msgCtx.Messages) == 0 {
        return errors.New("at least one message is required")
    }
    
    return nil
}
```

---

This migration guide provides comprehensive documentation for understanding and maintaining the OpenAI Responses API implementation. For additional support or questions, refer to the OpenAI official documentation or the internal development team.