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
}

func (suite *SessionRepositoryTestSuite) TestGetSessionsByUserID() {
	// Create multiple sessions
	session1 := &models.Session{
		UserID: suite.testUser.ID,
		Title:  "Session 1",
	}
	session2 := &models.Session{
		UserID: suite.testUser.ID,
		Title:  "Session 2",
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

func (suite *SessionRepositoryTestSuite) TestGetNonExistentSession() {
	_, err := suite.repo.GetByID(context.Background(), "non-existent-id")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "session not found")
}

func TestSessionRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(SessionRepositoryTestSuite))
}