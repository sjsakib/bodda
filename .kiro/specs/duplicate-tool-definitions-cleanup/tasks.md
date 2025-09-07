# Implementation Plan

- [x] 1. Add tool registry dependency to AI service

  - Modify `aiService` struct to include `ToolRegistry` field
  - Update `NewAIService` constructor to accept registry parameter
  - Update dependency injection in main application setup
  - _Requirements: 2.1, 2.2_

- [x] 2. Create tool definition conversion method

  - Implement `convertToolDefinitionToOpenAI` method in AI service
  - Add unit tests for conversion method with all 5 existing tools
  - Validate conversion produces identical OpenAI format as current hardcoded definitions
  - _Requirements: 2.1, 2.2, 4.1_

- [x] 3. Refactor getTools method to use registry

  - Replace hardcoded tool definitions in `getTools()` with registry calls
  - Use conversion method to transform registry tools to OpenAI format
  - Add error handling for registry failures with fallback behavior
  - _Requirements: 1.1, 1.2, 2.1_

- [x] 4. Update AI service initialization and dependency injection

  - Modify main application to pass tool registry to AI service constructor
  - Update any other places where AI service is instantiated
  - Ensure proper initialization order (registry before AI service)
  - _Requirements: 2.2, 3.3_

- [ ] 5. Add comprehensive unit tests for refactored functionality

  - Test AI service tool retrieval from registry
  - Test conversion of all tool types to OpenAI format
  - Test error handling when registry is unavailable
  - Test that converted tools match original hardcoded format exactly
  - _Requirements: 4.1, 4.3_

- [ ] 6. Run integration tests to validate tool execution

  - Test tool execution through AI service chat interface
  - Test direct tool execution through tool controller endpoints
  - Verify all 5 tools execute correctly with same parameters and results
  - Test tool schema validation consistency
  - _Requirements: 3.1, 3.2, 4.2_

- [-] 7. Remove duplicate tool definitions from AI service

  - Delete hardcoded tool definitions from `getTools()` method
  - Remove any unused tool-related constants or helper methods
  - Clean up imports if no longer needed
  - _Requirements: 1.1, 1.2_

- [ ] 8. Validate all existing tests pass
  - Run complete test suite to ensure no regressions
  - Fix any tests that depend on old hardcoded tool definitions
  - Verify tool execution tests work with registry-based approach
  - _Requirements: 3.1, 3.2, 4.1_
