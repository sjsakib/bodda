package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserModel(t *testing.T) {
	user := &User{
		ID:           "test-id",
		StravaID:     12345,
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		TokenExpiry:  time.Now(),
		FirstName:    "John",
		LastName:     "Doe",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Test JSON marshaling (sensitive fields should be omitted)
	jsonData, err := json.Marshal(user)
	assert.NoError(t, err)
	
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, "John")
	assert.Contains(t, jsonStr, "Doe")
	assert.Contains(t, jsonStr, "12345")
	assert.NotContains(t, jsonStr, "access-token") // Should be omitted with json:"-"
	assert.NotContains(t, jsonStr, "refresh-token") // Should be omitted with json:"-"
}

func TestSessionModel(t *testing.T) {
	session := &Session{
		ID:        "session-id",
		UserID:    "user-id",
		Title:     "Test Session",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(session)
	assert.NoError(t, err)
	
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, "session-id")
	assert.Contains(t, jsonStr, "user-id")
	assert.Contains(t, jsonStr, "Test Session")
}

func TestSessionModelWithLastResponseID(t *testing.T) {
	responseID := "response-123"
	session := &Session{
		ID:             "session-id",
		UserID:         "user-id",
		Title:          "Test Session",
		LastResponseID: &responseID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Test JSON marshaling with LastResponseID
	jsonData, err := json.Marshal(session)
	assert.NoError(t, err)
	
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, "session-id")
	assert.Contains(t, jsonStr, "user-id")
	assert.Contains(t, jsonStr, "Test Session")
	assert.Contains(t, jsonStr, "response-123")
	assert.Contains(t, jsonStr, "last_response_id")
}

func TestSessionModelWithNullLastResponseID(t *testing.T) {
	session := &Session{
		ID:             "session-id",
		UserID:         "user-id",
		Title:          "Test Session",
		LastResponseID: nil, // Explicitly set to nil
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Test JSON marshaling with nil LastResponseID
	jsonData, err := json.Marshal(session)
	assert.NoError(t, err)
	
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, "session-id")
	assert.Contains(t, jsonStr, "user-id")
	assert.Contains(t, jsonStr, "Test Session")
	
	// With omitempty tag, nil LastResponseID should not appear in JSON
	assert.NotContains(t, jsonStr, "last_response_id")
}

func TestSessionModelJSONUnmarshaling(t *testing.T) {
	// Test unmarshaling JSON with LastResponseID
	jsonWithResponseID := `{
		"id": "session-id",
		"user_id": "user-id",
		"title": "Test Session",
		"last_response_id": "response-456",
		"created_at": "2023-01-01T00:00:00Z",
		"updated_at": "2023-01-01T00:00:00Z"
	}`
	
	var session Session
	err := json.Unmarshal([]byte(jsonWithResponseID), &session)
	assert.NoError(t, err)
	
	assert.Equal(t, "session-id", session.ID)
	assert.Equal(t, "user-id", session.UserID)
	assert.Equal(t, "Test Session", session.Title)
	assert.NotNil(t, session.LastResponseID)
	assert.Equal(t, "response-456", *session.LastResponseID)
	
	// Test unmarshaling JSON without LastResponseID
	jsonWithoutResponseID := `{
		"id": "session-id-2",
		"user_id": "user-id-2",
		"title": "Test Session 2",
		"created_at": "2023-01-01T00:00:00Z",
		"updated_at": "2023-01-01T00:00:00Z"
	}`
	
	var session2 Session
	err = json.Unmarshal([]byte(jsonWithoutResponseID), &session2)
	assert.NoError(t, err)
	
	assert.Equal(t, "session-id-2", session2.ID)
	assert.Equal(t, "user-id-2", session2.UserID)
	assert.Equal(t, "Test Session 2", session2.Title)
	assert.Nil(t, session2.LastResponseID)
}

func TestMessageModel(t *testing.T) {
	message := &Message{
		ID:        "message-id",
		SessionID: "session-id",
		Role:      "user",
		Content:   "Hello, this is a test message",
		CreatedAt: time.Now(),
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(message)
	assert.NoError(t, err)
	
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, "message-id")
	assert.Contains(t, jsonStr, "session-id")
	assert.Contains(t, jsonStr, "user")
	assert.Contains(t, jsonStr, "Hello, this is a test message")
}

func TestAthleteLogbookModel(t *testing.T) {
	logbook := &AthleteLogbook{
		ID:        "logbook-id",
		UserID:    "user-id",
		Content:   "Athlete profile and training data",
		UpdatedAt: time.Now(),
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(logbook)
	assert.NoError(t, err)
	
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, "logbook-id")
	assert.Contains(t, jsonStr, "user-id")
	assert.Contains(t, jsonStr, "Athlete profile and training data")
}

func TestMessageRoleValidation(t *testing.T) {
	// Test valid roles
	validRoles := []string{"user", "assistant"}
	
	for _, role := range validRoles {
		message := &Message{
			ID:        "test-id",
			SessionID: "session-id",
			Role:      role,
			Content:   "Test content",
			CreatedAt: time.Now(),
		}
		
		// Should be able to create message with valid role
		assert.Equal(t, role, message.Role)
	}
}

func TestUserTokenSecurity(t *testing.T) {
	user := &User{
		ID:           "test-id",
		StravaID:     12345,
		AccessToken:  "secret-access-token",
		RefreshToken: "secret-refresh-token",
		TokenExpiry:  time.Now(),
		FirstName:    "John",
		LastName:     "Doe",
	}

	// Test that sensitive fields are not included in JSON
	jsonData, err := json.Marshal(user)
	assert.NoError(t, err)
	
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	
	// These fields should not be present in JSON due to json:"-" tags
	_, hasAccessToken := unmarshaled["access_token"]
	_, hasRefreshToken := unmarshaled["refresh_token"]
	_, hasTokenExpiry := unmarshaled["token_expiry"]
	
	assert.False(t, hasAccessToken, "access_token should not be in JSON")
	assert.False(t, hasRefreshToken, "refresh_token should not be in JSON")
	assert.False(t, hasTokenExpiry, "token_expiry should not be in JSON")
	
	// These fields should be present
	assert.Equal(t, "test-id", unmarshaled["id"])
	assert.Equal(t, float64(12345), unmarshaled["strava_id"]) // JSON numbers are float64
	assert.Equal(t, "John", unmarshaled["first_name"])
	assert.Equal(t, "Doe", unmarshaled["last_name"])
}