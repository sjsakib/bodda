# Implementation Plan

- [x] 1. Research and understand OpenAI responses API patterns

  - Study the OpenAI Go SDK v1.41.1 responses API documentation
  - Identify the specific methods and patterns to replace `CreateChatCompletionStream`
  - Document the key differences between completion API and responses API
  - _Requirements: 1.1, 5.1, 5.2_

- [x] 2. Add official OpenAI SDK alongside existing SDK (temporary)

  - Add `github.com/openai/openai-go` to go.mod (keep existing `github.com/sashabaranov/go-openai` temporarily)
  - Update import statements to use aliases for both SDKs during migration
  - Add official SDK client initialization in `NewAIService` alongside existing client
  - Add feature flag configuration to control which implementation to use during validation
  - Ensure compilation continues to work with both SDKs during migration period
  - _Requirements: 1.1, 5.1_

- [x] 3. Implement Responses API methods alongside existing methods (temporary dual implementation)

  - Create new `processMessageWithResponsesAPI` method using `Responses.NewStreaming`
  - Implement request structure conversion to `ResponseNewParams`
  - Convert message format to `ResponseInputItemUnionParam` structure
  - Update tool definitions to use `ToolUnionParam` format
  - Keep existing `processIterativeToolCalls` method temporarily for comparison
  - _Requirements: 1.1, 1.2_

- [x] 4. Implement event-based streaming processing

  - Create new `processResponsesAPIStream` method with event-based processing
  - Handle `ResponseStreamEventUnion` events in new method
  - Implement streaming loop using `stream.Next()` and `stream.Current()` pattern
  - Process different event types: `ResponseTextDeltaEvent`, `ResponseFunctionCallArgumentsDeltaEvent`, etc.
  - Keep existing delta-based processing method for comparison
  - _Requirements: 1.2, 1.3_

- [x] 5. Implement tool call processing for Responses API

  - Create new `parseToolCallsFromEvents` method to work with `ResponseFunctionCallArgumentsDeltaEvent`
  - Implement tool call accumulation logic for event-based processing
  - Handle tool call completion events and state management
  - Keep existing `parseToolCallsFromDelta` method for legacy implementation
  - Ensure tool call execution flow remains unchanged for consumers
  - _Requirements: 1.3_

- [x] 6. Implement error handling for official SDK

  - Create new `handleResponsesAPIError` method for official SDK error patterns
  - Implement error handling for official SDK error types
  - Create new error categorization and recovery logic
  - Keep existing error handling methods for legacy implementation
  - Add error handling routing based on feature flag
  - _Requirements: 1.4_

- [x] 7. Implement summary processor with official SDK

  - Create new `ProcessStreamDataWithResponsesAPI` method in summary processor
  - Implement Responses API usage alongside existing Chat Completions API
  - Convert request creation to use `ResponseNewParams` in new method
  - Add feature flag support to summary processor
  - Ensure summary processing maintains same functionality
  - _Requirements: 1.1_

- [x] 8. Implement temporary feature flag routing in AIService

  - Add temporary configuration option to enable/disable Responses API usage during validation
  - Implement routing logic in `ProcessMessage` to choose implementation during migration
  - Route to legacy or new implementation based on feature flag for testing
  - Ensure seamless switching between implementations during validation period
  - Add logging to track which implementation is being used for monitoring
  - _Requirements: 2.1, 5.2_

- [x] 9. Create comprehensive tests for both implementations

  - Create tests for both legacy and Responses API implementations
  - Mock both `ChatCompletionStreamResponse` and `ResponseStreamEventUnion` events
  - Ensure all existing test cases pass with both implementations
  - Add tests for feature flag routing logic
  - _Requirements: 3.1_

- [x] 10. Test streaming functionality for both APIs

  - Create comprehensive tests for both Chat Completions and Responses API streaming
  - Test event-based streaming response processing and error handling
  - Validate tool call processing with both delta and event structures
  - Test different event types and streaming scenarios
  - Compare behavior between implementations
  - _Requirements: 3.2_

- [x] 11. Test tool execution with both implementations

  - Test iterative tool calling with both Chat Completions and Responses API
  - Validate tool call parsing and execution flow with both structures
  - Ensure multi-round analysis works correctly with both SDKs
  - Test tool call state management and completion handling
  - Verify identical behavior between implementations
  - _Requirements: 3.3_

- [x] 12. Test error handling for both implementations

  - Test all error scenarios with both third-party and official SDK error types
  - Validate error recovery and fallback mechanisms for both APIs
  - Ensure proper error categorization and logging with both SDKs
  - Test enhanced error handling capabilities
  - Compare error handling behavior
  - _Requirements: 3.4_

- [x] 13. Validate end-to-end functionality parity

  - Test complete message processing flow with both implementations
  - Validate streaming behavior matches between implementations
  - Ensure user experience is identical regardless of implementation
  - Test all tool execution and multi-round analysis scenarios
  - Verify feature flag switching works seamlessly
  - _Requirements: 2.1, 2.2, 2.3, 2.4_

- [x] 14. Performance validation and comparison

  - Measure streaming performance with both Chat Completions and Responses API
  - Compare latency and throughput between implementations
  - Identify any performance differences or regressions
  - Validate memory usage and resource consumption for both
  - Document performance characteristics
  - _Requirements: 2.2_

- [x] 15. Switch to Responses API as default and validate production readiness

  - Update configuration to use Responses API as default implementation
  - Monitor production usage and performance for stability
  - Validate that all functionality works correctly in production environment
  - Ensure no regressions or issues with the new implementation
  - Document any performance improvements or changes observed
  - _Requirements: 2.1, 5.2_

- [x] 16. Complete legacy API removal

  - Remove third-party SDK dependency (`github.com/sashabaranov/go-openai`) from go.mod
  - Remove all legacy Chat Completions API implementation code and methods
  - Remove feature flag configuration and routing logic completely
  - Update all imports to use only the official OpenAI SDK
  - Clean up any remaining references to the old implementation
  - _Requirements: 4.1, 4.2, 4.3_

- [x] 17. Final cleanup and validation

  - Verify no legacy code or dependencies remain in the codebase
  - Run full test suite to ensure all functionality works with responses API only
  - Update configuration files to remove legacy API options
  - Validate that the application builds and runs without any legacy dependencies
  - _Requirements: 4.1, 4.4_

- [x] 18. Create comprehensive migration documentation

  - Document the complete migration process from completion API to responses API
  - Explain key differences between third-party SDK and official SDK usage
  - Document Chat Completions API vs Responses API architectural differences
  - Provide troubleshooting guide for the new responses API implementation
  - Document the legacy removal process and final system architecture
  - _Requirements: 5.1, 5.2, 5.3, 5.4_
