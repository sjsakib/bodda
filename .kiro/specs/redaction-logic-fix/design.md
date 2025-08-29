# Design Document

## Overview

This design addresses the issue in the current redaction logic where tool call results are being redacted unnecessarily. The current implementation redacts all previous stream tool outputs regardless of what follows them. The fix will modify the redaction logic to only redact tool call results when there are subsequent non-tool call messages in the conversation.

The solution maintains the existing architecture while enhancing the `ContextManager` interface and implementation to analyze message sequences and make intelligent redaction decisions based on what follows each tool call result.

## Architecture

The redaction logic fix integrates into the existing AI service architecture without requiring major structural changes:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   AI Service    │───▶│ Context Manager │───▶│ Enhanced Logic  │
│                 │    │                 │    │ • Sequence      │
│ • Chat Flow     │    │ • Redaction     │    │   Analysis      │
│ • Tool Calls    │    │   Control       │    │ • Message Type  │
│ • Streaming     │    │ • Environment   │    │   Detection     │
└─────────────────┘    │   Variable      │    │ • Conditional   │
                       └─────────────────┘    │   Redaction     │
                                              └─────────────────┘
```

### Key Components

- **Enhanced ContextManager**: Modified to analyze message sequences and apply conditional redaction
- **Message Sequence Analyzer**: New logic to determine what follows each tool call result
- **Environment Variable Control**: Existing `STREAM_REDACTION_ENABLED` continues to provide global on/off control
- **Backward Compatibility**: All existing interfaces remain unchanged

## Components and Interfaces

### Enhanced ContextManager Interface

The existing `ContextManager` interface remains unchanged to maintain backward compatibility:

```go
type ContextManager interface {
    RedactPreviousStreamOutputs(messages []openai.ChatCompletionMessage) []openai.ChatCompletionMessage
    ShouldRedact(toolCallName string) bool
}
```

### Internal Implementation Changes

The `contextManager` struct will be enhanced with new methods:

```go
type contextManager struct {
    redactionEnabled bool
    streamToolNames  map[string]bool
}

// New internal methods for sequence analysis
func (cm *contextManager) analyzeMessageSequence(messages []openai.ChatCompletionMessage, toolResultIndex int) bool
func (cm *contextManager) isNonToolCallMessage(message openai.ChatCompletionMessage) bool
func (cm *contextManager) hasSubsequentNonToolCallMessages(messages []openai.ChatCompletionMessage, startIndex int) bool
```

### Message Type Classification

Messages will be classified into categories for redaction decision-making:

1. **Tool Call Messages**: Assistant messages containing tool calls
2. **Tool Result Messages**: Tool role messages with tool call results
3. **Non-Tool Call Messages**: Assistant messages without tool calls (explanatory text, summaries, etc.)
4. **User Messages**: User role messages
5. **System Messages**: System role messages

## Data Models

### Message Sequence Analysis

The enhanced logic will analyze message sequences using this approach:

```go
type MessageAnalysis struct {
    ToolResultIndex     int
    ToolCallID         string
    HasSubsequentNonToolCalls bool
    IsLastMessage      bool
    FollowingMessageTypes []string
}
```

### Redaction Decision Matrix

| Scenario | Tool Result Followed By | Redaction Decision |
|----------|------------------------|-------------------|
| 1 | Non-tool call message | REDACT |
| 2 | Only other tool calls | DO NOT REDACT |
| 3 | End of conversation | DO NOT REDACT |
| 4 | User message then tool calls | DO NOT REDACT |
| 5 | Mixed: tool calls then non-tool call | REDACT |

## Error Handling

### Sequence Analysis Errors

- **Malformed Message Structure**: Log warning and default to current behavior (redact)
- **Missing Tool Call IDs**: Skip redaction for affected messages
- **Concurrent Tool Calls**: Analyze each tool call result independently

### Environment Variable Handling

- **Invalid Boolean Values**: Default to `true` (redaction enabled)
- **Missing Environment Variable**: Default to `true` (redaction enabled)
- **Runtime Changes**: Not supported - requires service restart

### Backward Compatibility

- **Existing Tests**: All current tests must continue to pass
- **API Compatibility**: No changes to public interfaces
- **Configuration**: Existing environment variables work unchanged

## Testing Strategy

### Unit Tests

1. **Message Sequence Analysis**
   - Test detection of non-tool call messages following tool results
   - Test handling of tool-call-only sequences
   - Test end-of-conversation scenarios
   - Test mixed message type sequences

2. **Redaction Decision Logic**
   - Test each scenario in the decision matrix
   - Test edge cases with empty messages
   - Test concurrent tool call handling

3. **Environment Variable Control**
   - Test redaction disabled behavior
   - Test redaction enabled behavior
   - Test default behavior when variable not set

### Integration Tests

1. **AI Service Integration**
   - Test redaction logic within full chat flow
   - Test streaming response handling
   - Test multiple conversation rounds

2. **Real Message Sequences**
   - Test with actual OpenAI message structures
   - Test with various tool call patterns
   - Test with mixed conversation flows

### Regression Tests

1. **Existing Functionality**
   - Ensure all current redaction tests pass
   - Verify no performance degradation
   - Confirm backward compatibility

2. **Edge Cases**
   - Empty conversation histories
   - Single message conversations
   - Tool calls without results

## Implementation Plan

### Phase 1: Core Logic Enhancement

1. **Enhance Message Analysis**
   - Add sequence analysis methods to `contextManager`
   - Implement message type classification
   - Add logging for redaction decisions

2. **Modify Redaction Logic**
   - Update `RedactPreviousStreamOutputs` method
   - Implement conditional redaction based on sequence analysis
   - Maintain existing behavior when redaction is disabled

### Phase 2: Testing and Validation

1. **Unit Test Implementation**
   - Create comprehensive test suite for new logic
   - Ensure all existing tests continue to pass
   - Add edge case coverage

2. **Integration Testing**
   - Test within AI service context
   - Validate with real conversation flows
   - Performance testing

### Phase 3: Documentation and Deployment

1. **Code Documentation**
   - Add comprehensive comments to new methods
   - Update existing documentation
   - Create troubleshooting guide

2. **Configuration Documentation**
   - Document environment variable behavior
   - Provide configuration examples
   - Update deployment guides

## Configuration

### Environment Variables

The existing environment variable continues to work:

```bash
# Enable/disable redaction globally
STREAM_REDACTION_ENABLED=true  # Default: true
```

### Configuration Examples

```bash
# Disable all redaction
STREAM_REDACTION_ENABLED=false

# Enable smart redaction (default)
STREAM_REDACTION_ENABLED=true
```

## Performance Considerations

### Computational Overhead

- **Message Analysis**: O(n) where n is the number of messages
- **Memory Usage**: Minimal additional memory for analysis state
- **Caching**: No caching needed as analysis is stateless

### Optimization Strategies

- **Early Termination**: Stop analysis once non-tool call message is found
- **Lazy Evaluation**: Only analyze sequences for tool results that might be redacted
- **Minimal Allocations**: Reuse analysis structures where possible

## Security Considerations

### Data Privacy

- **Redaction Content**: Ensure redacted content doesn't leak sensitive information
- **Logging**: Avoid logging sensitive tool call results during analysis
- **Error Messages**: Don't expose internal message content in error logs

### Access Control

- **Environment Variables**: Secure configuration management
- **Runtime Changes**: Prevent unauthorized redaction setting changes
- **Audit Trail**: Log redaction decisions for debugging

## Monitoring and Observability

### Logging

```go
// Enhanced logging for redaction decisions
log.Printf("Redaction analysis: tool_call_id=%s, has_subsequent_non_tool_calls=%v, decision=%s", 
    toolCallID, hasSubsequent, decision)
```

### Metrics

- **Redaction Rate**: Percentage of tool results redacted
- **Analysis Performance**: Time spent on sequence analysis
- **Decision Distribution**: Count of each redaction scenario

### Debugging

- **Verbose Mode**: Optional detailed logging of message analysis
- **Decision Tracing**: Track redaction decisions through conversation flow
- **Test Utilities**: Helper functions for testing redaction scenarios