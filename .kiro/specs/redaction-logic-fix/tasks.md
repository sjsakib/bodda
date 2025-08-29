# Implementation Plan

- [x] 1. Enhance context manager with message sequence analysis

  - Add new internal methods to analyze message sequences after tool call results
  - Implement logic to detect non-tool call messages vs tool call messages
  - Add method to determine if tool result has subsequent non-tool call messages
  - _Requirements: 2.1, 2.2_

- [x] 2. Implement conditional redaction logic in RedactPreviousStreamOutputs method

  - Modify the existing redaction method to analyze each tool result's position in conversation
  - Apply redaction only when tool results are followed by non-tool call messages
  - Preserve existing behavior when redaction is globally disabled via environment variable
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 4.1, 4.2_

- [x] 3. Add comprehensive unit tests for message sequence analysis

  - Write tests for detecting non-tool call messages following tool results
  - Test scenarios where tool results are followed only by other tool calls
  - Test end-of-conversation scenarios where tool result is the final message
  - Test mixed sequences with both tool calls and non-tool call messages
  - _Requirements: 1.1, 1.2, 1.3, 2.1, 2.2_

- [x] 4. Add unit tests for conditional redaction behavior

  - Test that tool results followed by non-tool call messages are redacted
  - Test that tool results followed only by tool calls are not redacted
  - Test that final tool results in conversation are not redacted
  - Test that environment variable still controls global redaction behavior
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 4.1, 4.2, 4.3, 4.4_

- [x] 5. Add integration tests for AI service redaction behavior

  - Test redaction logic within full AI service chat flow context
  - Test multiple conversation rounds with various tool call patterns
  - Test streaming response handling with new redaction logic
  - Verify existing AI service tests continue to pass with enhanced logic
  - _Requirements: 2.3, 3.1, 3.2, 3.3, 3.4_

- [x] 6. Add enhanced logging for redaction decisions

  - Log when redaction decisions are made based on subsequent message analysis
  - Include tool call ID and decision rationale in log messages
  - Ensure logging doesn't expose sensitive tool call content
  - _Requirements: 2.4_

- [x] 7. Update existing tests to work with enhanced redaction logic
  - Review and update existing context manager tests that may expect different behavior
  - Ensure all existing AI service tests pass with new conditional redaction
  - Update test expectations where appropriate for new behavior
  - _Requirements: 1.1, 1.2, 1.3, 1.4_
