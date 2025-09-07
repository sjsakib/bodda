# OpenAI Responses API Troubleshooting Reference

## Quick Reference Guide

This document provides quick solutions for common issues encountered with the OpenAI Responses API implementation.

## Common Error Patterns

### 1. Stream Connection Errors

#### Error: "stream connection timeout"
```
Error: context deadline exceeded while reading stream
```

**Cause:** Network timeout or slow API response

**Solution:**
```go
// Increase context timeout
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

// Add retry logic
for retries := 0; retries < 3; retries++ {
    stream := client.Responses.NewStreaming(ctx, params)
    if err := processStream(stream); err != nil {
        if retries < 2 {
            time.Sleep(time.Duration(retries+1) * time.Second)
            continue
        }
        return err
    }
    break
}
```

#### Error: "unexpected EOF"
```
Error: unexpected EOF while reading stream response
```

**Cause:** Stream interrupted or connection dropped

**Solution:**
```go
// Add proper error handling and recovery
for stream.Next() {
    event := stream.Current()
    
    // Process event with error checking
    if err := processEvent(event); err != nil {
        log.Printf("Event processing error: %v", err)
        // Continue processing other events
        continue
    }
}

// Check final stream error
if err := stream.Err(); err != nil {
    // Attempt recovery or return error
    return fmt.Errorf("stream error: %w", err)
}
```

### 2. Authentication Errors

#### Error: "invalid API key"
```
Error: 401 Unauthorized - Invalid API key provided
```

**Cause:** Missing or incorrect API key configuration

**Solution:**
```go
// Verify API key configuration
apiKey := os.Getenv("OPENAI_API_KEY")
if apiKey == "" {
    return errors.New("OPENAI_API_KEY environment variable not set")
}

// Initialize client with proper options
client := openai.NewClient(
    option.WithAPIKey(apiKey),
    option.WithBaseURL("https://api.openai.com/v1"),
)
```

#### Error: "quota exceeded"
```
Error: 429 Too Many Requests - You exceeded your current quota
```

**Cause:** API usage limits reached

**Solution:**
```go
// Implement exponential backoff
func (s *aiService) handleRateLimit(err error) error {
    var apiErr *openai.Error
    if errors.As(err, &apiErr) && apiErr.Code == "rate_limit_exceeded" {
        // Extract retry-after header if available
        retryAfter := time.Second * 60 // Default 1 minute
        
        log.Printf("Rate limit exceeded, waiting %v", retryAfter)
        time.Sleep(retryAfter)
        
        return ErrRetryable
    }
    return err
}
```

### 3. Tool Call Processing Errors

#### Error: "incomplete tool call arguments"
```
Error: tool call arguments incomplete or malformed
```

**Cause:** Tool call events not properly accumulated

**Solution:**
```go
// Proper tool call accumulation
type ToolCallAccumulator struct {
    calls map[string]*ToolCall
    mutex sync.RWMutex
}

func (acc *ToolCallAccumulator) ProcessEvent(event responses.ResponseStreamEventUnion) {
    acc.mutex.Lock()
    defer acc.mutex.Unlock()
    
    switch event := event.(type) {
    case responses.ResponseFunctionCallArgumentsDeltaEvent:
        if acc.calls[event.ID] == nil {
            acc.calls[event.ID] = &ToolCall{
                ID:   event.ID,
                Name: event.Name,
            }
        }
        
        // Accumulate arguments
        acc.calls[event.ID].Arguments += event.Arguments
        
    case responses.ResponseFunctionCallCompletedEvent:
        if call, exists := acc.calls[event.ID]; exists {
            call.Completed = true
        }
    }
}
```

#### Error: "tool execution timeout"
```
Error: tool execution exceeded maximum duration
```

**Cause:** Tool execution taking too long

**Solution:**
```go
// Add timeout to tool execution
func (s *aiService) executeToolWithTimeout(ctx context.Context, toolCall ToolCall) (string, error) {
    // Create timeout context
    toolCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Execute with timeout
    resultChan := make(chan ToolResult, 1)
    go func() {
        result, err := s.toolExecutor.Execute(toolCtx, toolCall)
        resultChan <- ToolResult{Result: result, Error: err}
    }()
    
    select {
    case result := <-resultChan:
        return result.Result, result.Error
    case <-toolCtx.Done():
        return "", fmt.Errorf("tool execution timeout: %s", toolCall.Name)
    }
}
```

### 4. Response Processing Errors

#### Error: "event type not handled"
```
Error: unknown event type in response stream
```

**Cause:** New event types not handled in switch statement

**Solution:**
```go
// Comprehensive event handling with fallback
func (s *aiService) processEvent(event responses.ResponseStreamEventUnion) error {
    switch event := event.(type) {
    case responses.ResponseTextDeltaEvent:
        return s.handleTextDelta(event)
        
    case responses.ResponseFunctionCallArgumentsDeltaEvent:
        return s.handleToolCallDelta(event)
        
    case responses.ResponseFunctionCallCompletedEvent:
        return s.handleToolCallCompleted(event)
        
    case responses.ResponseCompletedEvent:
        return s.handleResponseCompleted(event)
        
    default:
        // Log unknown event type but don't fail
        log.Printf("Unknown event type: %T", event)
        return nil
    }
}
```

#### Error: "response channel closed"
```
Error: send on closed channel
```

**Cause:** Attempting to send to closed response channel

**Solution:**
```go
// Safe channel operations
func (s *aiService) safeChannelSend(ch chan<- string, msg string) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from channel send panic: %v", r)
        }
    }()
    
    select {
    case ch <- msg:
        // Successfully sent
    default:
        // Channel full or closed, log but don't block
        log.Printf("Could not send message to channel: %s", msg)
    }
}
```

## Performance Issues

### 1. Slow Response Times

#### Symptom: High latency in streaming responses

**Diagnosis:**
```go
// Add timing measurements
start := time.Now()
stream := client.Responses.NewStreaming(ctx, params)

eventCount := 0
for stream.Next() {
    eventCount++
    event := stream.Current()
    
    eventStart := time.Now()
    processEvent(event)
    eventDuration := time.Since(eventStart)
    
    if eventDuration > 100*time.Millisecond {
        log.Printf("Slow event processing: %v for event type %T", eventDuration, event)
    }
}

totalDuration := time.Since(start)
log.Printf("Stream processing completed: %d events in %v", eventCount, totalDuration)
```

**Solutions:**
1. **Optimize event processing:**
```go
// Use buffered channels for async processing
eventChan := make(chan responses.ResponseStreamEventUnion, 100)

go func() {
    for event := range eventChan {
        processEventAsync(event)
    }
}()

for stream.Next() {
    select {
    case eventChan <- stream.Current():
    default:
        // Channel full, process synchronously
        processEvent(stream.Current())
    }
}
```

2. **Reduce processing overhead:**
```go
// Minimize allocations in hot path
var stringBuilder strings.Builder
stringBuilder.Grow(1024) // Pre-allocate capacity

for stream.Next() {
    event := stream.Current()
    if textEvent, ok := event.(responses.ResponseTextDeltaEvent); ok {
        stringBuilder.WriteString(textEvent.Delta)
    }
}
```

### 2. High Memory Usage

#### Symptom: Memory usage growing during streaming

**Diagnosis:**
```go
// Monitor memory usage
func (s *aiService) monitorMemory() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    log.Printf("Memory: Alloc=%d KB, TotalAlloc=%d KB, Sys=%d KB, NumGC=%d",
        m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC)
}

// Call periodically during processing
ticker := time.NewTicker(5 * time.Second)
defer ticker.Stop()

go func() {
    for range ticker.C {
        s.monitorMemory()
    }
}()
```

**Solutions:**
1. **Limit event history:**
```go
// Don't store all events, only what's needed
type StreamProcessor struct {
    currentToolCalls map[string]*ToolCall
    responseBuilder  strings.Builder
    // Don't store: eventHistory []Event
}
```

2. **Force garbage collection:**
```go
// Trigger GC after large operations
defer func() {
    runtime.GC()
    debug.FreeOSMemory()
}()
```

## Configuration Issues

### 1. Environment Variables

#### Issue: Configuration not loading properly

**Check list:**
```bash
# Verify environment variables
echo $OPENAI_API_KEY
echo $OPENAI_MODEL
echo $OPENAI_BASE_URL

# Check .env file loading
cat .env | grep OPENAI
```

**Solution:**
```go
// Robust configuration loading
func loadConfig() (*Config, error) {
    // Load from .env file
    if err := godotenv.Load(); err != nil {
        log.Printf("Warning: .env file not found: %v", err)
    }
    
    config := &Config{
        OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
        Model:        getEnvWithDefault("OPENAI_MODEL", "gpt-4"),
        BaseURL:      getEnvWithDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
    }
    
    // Validate required fields
    if config.OpenAIAPIKey == "" {
        return nil, errors.New("OPENAI_API_KEY is required")
    }
    
    return config, nil
}

func getEnvWithDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### 2. Model Configuration

#### Issue: Model not supported by Responses API

**Error:**
```
Error: model 'gpt-3.5-turbo' is not supported by the responses API
```

**Solution:**
```go
// Use supported models
var supportedModels = map[string]responses.ResponsesModel{
    "o1-pro":     responses.ResponsesModelO1Pro,
    "o1":         responses.ResponsesModelO1,
    "o1-mini":    responses.ResponsesModelO1Mini,
    "gpt-4o":     responses.ResponsesModelGPT4o,
    "gpt-4o-mini": responses.ResponsesModelGPT4oMini,
}

func validateAndConvertModel(modelName string) (responses.ResponsesModel, error) {
    if model, exists := supportedModels[modelName]; exists {
        return model, nil
    }
    
    // Provide helpful error message
    var supported []string
    for name := range supportedModels {
        supported = append(supported, name)
    }
    
    return "", fmt.Errorf("model '%s' not supported by Responses API. Supported models: %v", 
        modelName, supported)
}
```

## Testing Issues

### 1. Mock Setup Problems

#### Issue: Mocks not working with official SDK

**Solution:**
```go
// Create proper mocks for official SDK
type MockResponsesService struct {
    events []responses.ResponseStreamEventUnion
    err    error
}

func (m *MockResponsesService) NewStreaming(ctx context.Context, params responses.ResponseNewParams) *MockStream {
    return &MockStream{
        events: m.events,
        err:    m.err,
        index:  0,
    }
}

type MockStream struct {
    events []responses.ResponseStreamEventUnion
    err    error
    index  int
}

func (s *MockStream) Next() bool {
    return s.index < len(s.events)
}

func (s *MockStream) Current() responses.ResponseStreamEventUnion {
    if s.index < len(s.events) {
        event := s.events[s.index]
        s.index++
        return event
    }
    return nil
}

func (s *MockStream) Err() error {
    return s.err
}
```

### 2. Integration Test Failures

#### Issue: Tests failing with real API

**Solution:**
```go
// Robust integration test setup
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    apiKey := os.Getenv("OPENAI_API_KEY_TEST")
    if apiKey == "" {
        t.Skip("OPENAI_API_KEY_TEST not set")
    }
    
    // Use test-specific configuration
    client := openai.NewClient(
        option.WithAPIKey(apiKey),
        option.WithBaseURL("https://api.openai.com/v1"),
    )
    
    // Add timeout for integration tests
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Test with simple request
    params := responses.ResponseNewParams{
        Model: responses.ResponsesModelGPT4oMini, // Use cheaper model for tests
        Input: []responses.ResponseInputItemUnionParam{
            responses.ResponseInputItemParamOfMessage("Hello", responses.EasyInputMessageRoleUser),
        },
    }
    
    stream := client.Responses.NewStreaming(ctx, params)
    
    var responseText strings.Builder
    for stream.Next() {
        event := stream.Current()
        if textEvent, ok := event.(responses.ResponseTextDeltaEvent); ok {
            responseText.WriteString(textEvent.Delta)
        }
    }
    
    if err := stream.Err(); err != nil {
        t.Fatalf("Stream error: %v", err)
    }
    
    if responseText.Len() == 0 {
        t.Error("Expected non-empty response")
    }
}
```

## Debugging Tools

### 1. Event Logging

```go
// Comprehensive event logging
func logEvent(event responses.ResponseStreamEventUnion) {
    switch event := event.(type) {
    case responses.ResponseTextDeltaEvent:
        log.Printf("[TEXT] Delta: %q", event.Delta)
        
    case responses.ResponseFunctionCallArgumentsDeltaEvent:
        log.Printf("[TOOL] ID: %s, Name: %s, Args: %q", 
            event.ID, event.Name, event.Arguments)
            
    case responses.ResponseFunctionCallCompletedEvent:
        log.Printf("[TOOL_COMPLETE] ID: %s", event.ID)
        
    case responses.ResponseCompletedEvent:
        log.Printf("[COMPLETE] Finish reason: %s", event.FinishReason)
        
    default:
        log.Printf("[UNKNOWN] Event type: %T", event)
    }
}
```

### 2. Performance Profiling

```go
// Add profiling endpoints for debugging
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Use profiling during development:
// go tool pprof http://localhost:6060/debug/pprof/profile
// go tool pprof http://localhost:6060/debug/pprof/heap
```

### 3. Request/Response Logging

```go
// Log full request/response for debugging
func (s *aiService) debugLogRequest(params responses.ResponseNewParams) {
    if s.config.DebugMode {
        jsonData, _ := json.MarshalIndent(params, "", "  ")
        log.Printf("Request: %s", jsonData)
    }
}

func (s *aiService) debugLogResponse(events []responses.ResponseStreamEventUnion) {
    if s.config.DebugMode {
        for i, event := range events {
            log.Printf("Event %d: %T", i, event)
        }
    }
}
```

## Emergency Procedures

### 1. Rollback Plan

If critical issues arise:

1. **Immediate rollback:**
```bash
# Revert to previous version
git revert <migration-commit-hash>
git push origin main
```

2. **Temporary fix:**
```go
// Add feature flag for emergency rollback
if os.Getenv("EMERGENCY_ROLLBACK") == "true" {
    return s.processWithLegacyAPI(ctx, msgCtx)
}
```

### 2. Circuit Breaker

```go
// Implement circuit breaker for API calls
type CircuitBreaker struct {
    failures    int
    lastFailure time.Time
    threshold   int
    timeout     time.Duration
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.failures >= cb.threshold {
        if time.Since(cb.lastFailure) < cb.timeout {
            return errors.New("circuit breaker open")
        }
        // Reset after timeout
        cb.failures = 0
    }
    
    if err := fn(); err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        return err
    }
    
    cb.failures = 0
    return nil
}
```

---

For additional support, check the main [Migration Guide](./MIGRATION_GUIDE.md) or contact the development team.