package database

import (
	"context"
	"testing"
	"time"

	"bodda/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SessionRepositoryTestSuite struct {
	suite.Suite
	repo     *SessionRepository
	userRepo *UserRepository
	db       *TestDB
	testUser *models.User
}

func (suite *SessionRepositoryTestSuite) SetupSuite() {
	suite.db = NewTestDB(suite.T())
	suite.repo = NewSessionRepository(suite.db.Pool)
	suite.userRepo = NewUserRepository(suite.db.Pool)
}

func (suite *SessionRepositoryTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *SessionRepositoryTestSuite) SetupTest() {
	suite.db.CleanTables()
	
	// Create a test user for session tests
	suite.testUser = &models.User{
		StravaID:     12345,
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenExpiry:  time.Now().Add(time.Hour),
		FirstName:    "John",
		LastName:     "Doe",
	}
	err := suite.userRepo.Create(context.Background(), suite.testUser)
	assert.NoError(suite.T(), err)
}

func (suite *SessionRepositoryTestSuite) TestCreateSession() {
	session := &models.Session{
		UserID: suite.testUser.ID,
		Title:  "Test Session",
	}

	err := suite.repo.Create(context.Background(), session)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), session.ID)
	assert.NotZero(suite.T(), session.CreatedAt)
	assert.NotZero(suite.T(), session.UpdatedAt)
	assert.Nil(suite.T(), session.LastResponseID)
}

func (suite *SessionRepositoryTestSuite) TestCreateSessionWithLastResponseID() {
	responseID := "response_123"
	session := &models.Session{
		UserID:         suite.testUser.ID,
		Title:          "Test Session",
		LastResponseID: &responseID,
	}

	err := suite.repo.Create(context.Background(), session)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), session.ID)
	assert.NotZero(suite.T(), session.CreatedAt)
	assert.NotZero(suite.T(), session.UpdatedAt)
	assert.NotNil(suite.T(), session.LastResponseID)
	assert.Equal(suite.T(), responseID, *session.LastResponseID)
}

func (suite *SessionRepositoryTestSuite) TestGetSessionByID() {
	// Create a session first
	session := &models.Session{
		UserID: suite.testUser.ID,
		Title:  "Test Session",
	}
	err := suite.repo.Create(context.Background(), session)
	assert.NoError(suite.T(), err)

	// Get the session by ID
	retrievedSession, err := suite.repo.GetByID(context.Background(), session.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), session.ID, retrievedSession.ID)
	assert.Equal(suite.T(), session.UserID, retrievedSession.UserID)
	assert.Equal(suite.T(), session.Title, retrievedSession.Title)
	assert.Nil(suite.T(), retrievedSession.LastResponseID)
}

func (suite *SessionRepositoryTestSuite) TestGetSessionByIDWithLastResponseID() {
	// Create a session with last_response_id
	responseID := "response_456"
	session := &models.Session{
		UserID:         suite.testUser.ID,
		Title:          "Test Session",
		LastResponseID: &responseID,
	}
	err := suite.repo.Create(context.Background(), session)
	assert.NoError(suite.T(), err)

	// Get the session by ID
	retrievedSession, err := suite.repo.GetByID(context.Background(), session.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), session.ID, retrievedSession.ID)
	assert.Equal(suite.T(), session.UserID, retrievedSession.UserID)
	assert.Equal(suite.T(), session.Title, retrievedSession.Title)
	assert.NotNil(suite.T(), retrievedSession.LastResponseID)
	assert.Equal(suite.T(), responseID, *retrievedSession.LastResponseID)
}

func (suite *SessionRepositoryTestSuite) TestGetSessionsByUserID() {
	// Create multiple sessions with different last_response_id values
	responseID1 := "response_1"
	session1 := &models.Session{
		UserID:         suite.testUser.ID,
		Title:          "Session 1",
		LastResponseID: &responseID1,
	}
	session2 := &models.Session{
		UserID: suite.testUser.ID,
		Title:  "Session 2",
		// LastResponseID is nil
	}

	err := suite.repo.Create(context.Background(), session1)
	assert.NoError(suite.T(), err)
	
	err = suite.repo.Create(context.Background(), session2)
	assert.NoError(suite.T(), err)

	// Get sessions by user ID
	sessions, err := suite.repo.GetByUserID(context.Background(), suite.testUser.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 2)
	
	// Should be ordered by updated_at DESC (most recent first)
	assert.True(suite.T(), sessions[0].UpdatedAt.After(sessions[1].UpdatedAt) || 
		sessions[0].UpdatedAt.Equal(sessions[1].UpdatedAt))
	
	// Verify last_response_id is properly retrieved
	for _, session := range sessions {
		if session.Title == "Session 1" {
			assert.NotNil(suite.T(), session.LastResponseID)
			assert.Equal(suite.T(), responseID1, *session.LastResponseID)
		} else if session.Title == "Session 2" {
			assert.Nil(suite.T(), session.LastResponseID)
		}
	}
}

func (suite *SessionRepositoryTestSuite) TestUpdateSession() {
	// Create a session first
	session := &models.Session{
		UserID: suite.testUser.ID,
		Title:  "Original Title",
	}
	err := suite.repo.Create(context.Background(), session)
	assert.NoError(suite.T(), err)

	// Update the session
	session.Title = "Updated Title"
	originalUpdatedAt := session.UpdatedAt

	err = suite.repo.Update(context.Background(), session)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), session.UpdatedAt.After(originalUpdatedAt))

	// Verify the update
	retrievedSession, err := suite.repo.GetByID(context.Background(), session.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Title", retrievedSession.Title)
}

func (suite *SessionRepositoryTestSuite) TestDeleteSession() {
	// Create a session first
	session := &models.Session{
		UserID: suite.testUser.ID,
		Title:  "Test Session",
	}
	err := suite.repo.Create(context.Background(), session)
	assert.NoError(suite.T(), err)

	// Delete the session
	err = suite.repo.Delete(context.Background(), session.ID)
	assert.NoError(suite.T(), err)

	// Verify the session is deleted
	_, err = suite.repo.GetByID(context.Background(), session.ID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "session not found")
}

func (suite *SessionRepositoryTestSuite) TestUpdateLastResponseID() {
	// Create a session first
	session := &models.Session{
		UserID: suite.testUser.ID,
		Title:  "Test Session",
	}
	err := suite.repo.Create(context.Background(), session)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), session.LastResponseID)

	// Update the last_response_id
	responseID := "new_response_123"
	err = suite.repo.UpdateLastResponseID(context.Background(), session.ID, responseID)
	assert.NoError(suite.T(), err)

	// Verify the update
	retrievedSession, err := suite.repo.GetByID(context.Background(), session.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedSession.LastResponseID)
	assert.Equal(suite.T(), responseID, *retrievedSession.LastResponseID)
	assert.True(suite.T(), retrievedSession.UpdatedAt.After(session.UpdatedAt))
}

func (suite *SessionRepositoryTestSuite) TestUpdateLastResponseIDOverwrite() {
	// Create a session with an existing last_response_id
	initialResponseID := "initial_response"
	session := &models.Session{
		UserID:         suite.testUser.ID,
		Title:          "Test Session",
		LastResponseID: &initialResponseID,
	}
	err := suite.repo.Create(context.Background(), session)
	assert.NoError(suite.T(), err)

	// Update with a new last_response_id
	newResponseID := "new_response_456"
	err = suite.repo.UpdateLastResponseID(context.Background(), session.ID, newResponseID)
	assert.NoError(suite.T(), err)

	// Verify the update overwrote the previous value
	retrievedSession, err := suite.repo.GetByID(context.Background(), session.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedSession.LastResponseID)
	assert.Equal(suite.T(), newResponseID, *retrievedSession.LastResponseID)
	assert.NotEqual(suite.T(), initialResponseID, *retrievedSession.LastResponseID)
}

func (suite *SessionRepositoryTestSuite) TestUpdateLastResponseIDNonExistentSession() {
	// Use a valid UUID format that doesn't exist in the database
	nonExistentUUID := "00000000-0000-0000-0000-000000000000"
	err := suite.repo.UpdateLastResponseID(context.Background(), nonExistentUUID, "response_123")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "session not found")
}

func (suite *SessionRepositoryTestSuite) TestBackwardCompatibilityNullLastResponseID() {
	// This test ensures that sessions with NULL last_response_id work properly
	// Create a session without last_response_id (simulating existing data)
	session := &models.Session{
		UserID: suite.testUser.ID,
		Title:  "Legacy Session",
	}
	err := suite.repo.Create(context.Background(), session)
	assert.NoError(suite.T(), err)

	// Verify it can be retrieved properly
	retrievedSession, err := suite.repo.GetByID(context.Background(), session.ID)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), retrievedSession.LastResponseID)

	// Verify it appears in user sessions list
	sessions, err := suite.repo.GetByUserID(context.Background(), suite.testUser.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 1)
	assert.Nil(suite.T(), sessions[0].LastResponseID)

	// Verify we can update it later
	responseID := "first_response"
	err = suite.repo.UpdateLastResponseID(context.Background(), session.ID, responseID)
	assert.NoError(suite.T(), err)

	// Verify the update worked
	updatedSession, err := suite.repo.GetByID(context.Background(), session.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedSession.LastResponseID)
	assert.Equal(suite.T(), responseID, *updatedSession.LastResponseID)
}

func (suite *SessionRepositoryTestSuite) TestGetNonExistentSession() {
	// Use a valid UUID format that doesn't exist in the database
	nonExistentUUID := "00000000-0000-0000-0000-000000000000"
	_, err := suite.repo.GetByID(context.Background(), nonExistentUUID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "session not found")
}

func TestSessionRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(SessionRepositoryTestSuite))
}