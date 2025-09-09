# Requirements Document

## Introduction

This feature adds the capability to track and store the last response ID associated with each session. The primary purpose is to enable the system to reference the previous AI response when a user sends a new message to an existing session, supporting conversation continuity and context management.

## Requirements

### Requirement 1

**User Story:** As a user, I want the system to track the last response ID for my session, so that when I send a new message, the system can reference the previous response for context continuity.

#### Acceptance Criteria

1. WHEN a new session is created THEN the system SHALL initialize the last_response_id field as null
2. WHEN an AI response is generated for a session THEN the system SHALL update the session's last_response_id with the new response ID
3. WHEN a user sends a new message to an existing session THEN the system SHALL have access to the previous response ID
4. WHEN a session has no previous responses THEN the last_response_id SHALL remain null

### Requirement 2

**User Story:** As a developer, I want the last response ID to be persisted in the database, so that the information survives application restarts and can be reliably accessed when processing new messages.

#### Acceptance Criteria

1. WHEN the database schema is updated THEN the sessions table SHALL include a last_response_id column
2. WHEN a session is saved to the database THEN the last_response_id SHALL be persisted
3. WHEN a session is loaded from the database THEN the last_response_id SHALL be retrieved accurately
4. WHEN migrating existing sessions THEN the last_response_id SHALL default to null for backward compatibility