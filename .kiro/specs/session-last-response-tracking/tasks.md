# Implementation Plan

- [x] 1. Add database migration for last_response_id column

  - Add migration SQL to add last_response_id column to sessions table
  - Update migrations.go to include the new migration
  - Write tests to verify migration execution and rollback
  - _Requirements: 2.1, 2.4_

- [x] 2. Update Session model with LastResponseID field

  - Add LastResponseID field to Session struct in internal/models/user.go
  - Update JSON and database tags for proper serialization
  - Write unit tests for Session model with new field
  - _Requirements: 1.1, 2.2_

- [x] 3. Enhance SessionRepository with last_response_id support

  - Update Create method to handle NULL last_response_id initialization
  - Update GetByID method to include last_response_id in SELECT query
  - Update GetByUserID method to include last_response_id in SELECT query
  - Add UpdateLastResponseID method for atomic updates
  - Write comprehensive tests for all repository methods
  - _Requirements: 1.1, 2.2, 2.3_

- [x] 4. Integrate last_response_id tracking in AI service

  - Modify chat service to populate MessageContext.LastResponseID from session
  - Update AI service to store new response ID back to session after successful generation
  - Ensure atomic updates between response creation and session update
  - Write integration tests for the complete flow
  - _Requirements: 1.2, 1.3_

- [ ] 5. Add comprehensive test coverage
  - Write unit tests for all modified components
  - Add integration tests for multi-turn conversation scenarios
  - Test backward compatibility with existing sessions (NULL last_response_id)
  - Test error scenarios and rollback behavior
  - _Requirements: 1.1, 1.2, 1.3, 1.4_
