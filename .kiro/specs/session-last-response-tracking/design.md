# Design Document

## Overview

This design implements session-level tracking of the last response ID to enable proper multi-turn conversation context. The system will store the most recent AI response ID with each session, allowing the AI service to reference previous responses when processing new messages from users.

## Architecture

The implementation follows the existing layered architecture:

- **Database Layer**: Add `last_response_id` column to sessions table
- **Repository Layer**: Update SessionRepository to handle the new field
- **Service Layer**: Modify AI service to update session's last response ID when generating responses
- **Model Layer**: Update Session model to include LastResponseID field

## Components and Interfaces

### Database Schema Changes

```sql
ALTER TABLE sessions 
ADD COLUMN IF NOT EXISTS last_response_id TEXT;
```

The `last_response_id` column will:
- Store OpenAI response IDs as TEXT (nullable)
- Default to NULL for new sessions
- Be updated atomically with response creation

### Model Updates

```go
type Session struct {
    ID             string    `json:"id" db:"id"`
    UserID         string    `json:"user_id" db:"user_id"`
    Title          string    `json:"title" db:"title"`
    LastResponseID *string   `json:"last_response_id,omitempty" db:"last_response_id"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}
```

### Repository Interface Updates

The SessionRepository will be extended with:

```go
// UpdateLastResponseID updates only the last_response_id field for a session
func (r *SessionRepository) UpdateLastResponseID(ctx context.Context, sessionID string, responseID string) error

// GetByID will be updated to include last_response_id in SELECT query
```

### Service Integration

The AI service will be modified to:

1. **Retrieve last response ID**: When processing a message, get the session's current last_response_id
2. **Use in API calls**: Pass the last_response_id to OpenAI's Responses API for multi-turn context
3. **Update after response**: Store the new response ID back to the session after successful response generation

## Data Models

### Session Model Enhancement

```go
type Session struct {
    ID             string    `json:"id" db:"id"`
    UserID         string    `json:"user_id" db:"user_id"`
    Title          string    `json:"title" db:"title"`
    LastResponseID *string   `json:"last_response_id,omitempty" db:"last_response_id"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}
```

### MessageContext Enhancement

The existing MessageContext already includes LastResponseID field:

```go
type MessageContext struct {
    UserID              string
    SessionID           string
    Message             string
    ConversationHistory []*models.Message
    AthleteLogbook      *models.AthleteLogbook
    User                *models.User
    LastResponseID      string // This field already exists
}
```

## Error Handling

### Database Errors
- Migration failures will be logged and handled gracefully
- NULL values for last_response_id will be handled properly in queries
- Transaction rollbacks will revert last_response_id updates if response creation fails

### Service Errors
- If session update fails after response generation, log the error but don't fail the user request
- Handle cases where session doesn't exist when trying to update last_response_id
- Gracefully handle NULL last_response_id values when building MessageContext

### Backward Compatibility
- Existing sessions will have NULL last_response_id initially
- The system will work normally for sessions without a last_response_id
- Migration will not break existing functionality

## Testing Strategy

### Unit Tests
- Test Session model serialization/deserialization with new field
- Test SessionRepository CRUD operations with last_response_id
- Test AI service integration with last_response_id updates
- Test migration script execution

### Integration Tests
- Test end-to-end flow: message → response → session update
- Test multi-turn conversation scenarios
- Test error scenarios (failed updates, missing sessions)
- Test backward compatibility with existing sessions

### Database Tests
- Test migration execution and rollback
- Test NULL handling in queries
- Test concurrent updates to last_response_id

## Implementation Flow

1. **Database Migration**: Add last_response_id column to sessions table
2. **Model Update**: Add LastResponseID field to Session struct
3. **Repository Update**: Modify SessionRepository methods to handle new field
4. **Service Integration**: Update AI service to populate and use last_response_id
5. **Testing**: Add comprehensive tests for all components

## Integration Points

### AI Service Integration
The AI service already captures response IDs in the `processResponsesAPIStreamWithID` method. The integration point will be:

1. **Before processing**: Retrieve session's last_response_id and populate MessageContext
2. **After response**: Update session's last_response_id with the new response ID

### Chat Service Integration
The chat service will need to:
1. Load session with last_response_id when building MessageContext
2. Ensure session updates are called after successful AI responses

### API Endpoints
No new API endpoints are required. Existing session endpoints will automatically include the new field in responses.