package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"bodda/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MessageRepositoryTestSuite struct {
	suite.Suite
	repo        *MessageRepository
	userRepo    *UserRepository
	sessionRepo *SessionRepository
	db          *TestDB
	testUser    *models.User
	testSession *models.Session
}

func (suite *MessageRepositoryTestSuite) SetupSuite() {
	suite.db = NewTestDB(suite.T())
	suite.repo = NewMessageRepository(suite.db.Pool)
	suite.userRepo = NewUserRepository(suite.db.Pool)
	suite.sessionRepo = NewSessionRepository(suite.db.Pool)
}

func (suite *MessageRepositoryTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *MessageRepositoryTestSuite) SetupTest() {
	suite.db.CleanTables()
	
	// Create a test user
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

	// Create a test session
	suite.testSession = &models.Session{
		UserID: suite.testUser.ID,
		Title:  "Test Session",
	}
	err = suite.sessionRepo.Create(context.Background(), suite.testSession)
	assert.NoError(suite.T(), err)
}

func (suite *MessageRepositoryTestSuite) TestCreateMessage() {
	message := &models.Message{
		SessionID: suite.testSession.ID,
		Role:      "user",
		Content:   "Hello, this is a test message",
	}

	err := suite.repo.Create(context.Background(), message)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), message.ID)
	assert.NotZero(suite.T(), message.CreatedAt)
}

func (suite *MessageRepositoryTestSuite) TestGetMessageByID() {
	// Create a message first
	message := &models.Message{
		SessionID: suite.testSession.ID,
		Role:      "user",
		Content:   "Hello, this is a test message",
	}
	err := suite.repo.Create(context.Background(), message)
	assert.NoError(suite.T(), err)

	// Get the message by ID
	retrievedMessage, err := suite.repo.GetByID(context.Background(), message.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), message.ID, retrievedMessage.ID)
	assert.Equal(suite.T(), message.SessionID, retrievedMessage.SessionID)
	assert.Equal(suite.T(), message.Role, retrievedMessage.Role)
	assert.Equal(suite.T(), message.Content, retrievedMessage.Content)
}

func (suite *MessageRepositoryTestSuite) TestGetMessagesBySessionID() {
	// Create multiple messages
	message1 := &models.Message{
		SessionID: suite.testSession.ID,
		Role:      "user",
		Content:   "First message",
	}
	message2 := &models.Message{
		SessionID: suite.testSession.ID,
		Role:      "assistant",
		Content:   "Second message",
	}
	message3 := &models.Message{
		SessionID: suite.testSession.ID,
		Role:      "user",
		Content:   "Third message",
	}

	err := suite.repo.Create(context.Background(), message1)
	assert.NoError(suite.T(), err)
	
	// Add small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	err = suite.repo.Create(context.Background(), message2)
	assert.NoError(suite.T(), err)
	
	time.Sleep(time.Millisecond)
	err = suite.repo.Create(context.Background(), message3)
	assert.NoError(suite.T(), err)

	// Get messages by session ID
	messages, err := suite.repo.GetBySessionID(context.Background(), suite.testSession.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), messages, 3)
	
	// Should be ordered by created_at ASC (chronological order)
	assert.Equal(suite.T(), "First message", messages[0].Content)
	assert.Equal(suite.T(), "Second message", messages[1].Content)
	assert.Equal(suite.T(), "Third message", messages[2].Content)
}

func (suite *MessageRepositoryTestSuite) TestGetMessagesBySessionIDWithLimit() {
	// Create multiple messages
	messages := []*models.Message{
		{SessionID: suite.testSession.ID, Role: "user", Content: "Message 1"},
		{SessionID: suite.testSession.ID, Role: "assistant", Content: "Message 2"},
		{SessionID: suite.testSession.ID, Role: "user", Content: "Message 3"},
		{SessionID: suite.testSession.ID, Role: "assistant", Content: "Message 4"},
		{SessionID: suite.testSession.ID, Role: "user", Content: "Message 5"},
	}

	for i, msg := range messages {
		err := suite.repo.Create(context.Background(), msg)
		assert.NoError(suite.T(), err)
		if i < len(messages)-1 {
			time.Sleep(time.Millisecond) // Ensure different timestamps
		}
	}

	// Get last 3 messages
	recentMessages, err := suite.repo.GetBySessionIDWithLimit(context.Background(), suite.testSession.ID, 3)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), recentMessages, 3)
	
	// Should return the last 3 messages in chronological order
	assert.Equal(suite.T(), "Message 3", recentMessages[0].Content)
	assert.Equal(suite.T(), "Message 4", recentMessages[1].Content)
	assert.Equal(suite.T(), "Message 5", recentMessages[2].Content)
}

func (suite *MessageRepositoryTestSuite) TestDeleteMessage() {
	// Create a message first
	message := &models.Message{
		SessionID: suite.testSession.ID,
		Role:      "user",
		Content:   "Test message",
	}
	err := suite.repo.Create(context.Background(), message)
	assert.NoError(suite.T(), err)

	// Delete the message
	err = suite.repo.Delete(context.Background(), message.ID)
	assert.NoError(suite.T(), err)

	// Verify the message is deleted
	_, err = suite.repo.GetByID(context.Background(), message.ID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "message not found")
}

func (suite *MessageRepositoryTestSuite) TestGetNonExistentMessage() {
	_, err := suite.repo.GetByID(context.Background(), "non-existent-id")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "message not found")
}

func (suite *MessageRepositoryTestSuite) TestGetMessagesBySessionIDWithPagination() {
	// Create 5 messages
	messages := []*models.Message{
		{SessionID: suite.testSession.ID, Role: "user", Content: "Message 1"},
		{SessionID: suite.testSession.ID, Role: "assistant", Content: "Message 2"},
		{SessionID: suite.testSession.ID, Role: "user", Content: "Message 3"},
		{SessionID: suite.testSession.ID, Role: "assistant", Content: "Message 4"},
		{SessionID: suite.testSession.ID, Role: "user", Content: "Message 5"},
	}

	for i, msg := range messages {
		err := suite.repo.Create(context.Background(), msg)
		assert.NoError(suite.T(), err)
		if i < len(messages)-1 {
			time.Sleep(time.Millisecond) // Ensure different timestamps
		}
	}

	// Test pagination - first page (limit 2, offset 0)
	page1, err := suite.repo.GetBySessionIDWithPagination(context.Background(), suite.testSession.ID, 2, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), page1, 2)
	assert.Equal(suite.T(), "Message 1", page1[0].Content)
	assert.Equal(suite.T(), "Message 2", page1[1].Content)

	// Test pagination - second page (limit 2, offset 2)
	page2, err := suite.repo.GetBySessionIDWithPagination(context.Background(), suite.testSession.ID, 2, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), page2, 2)
	assert.Equal(suite.T(), "Message 3", page2[0].Content)
	assert.Equal(suite.T(), "Message 4", page2[1].Content)

	// Test pagination - third page (limit 2, offset 4)
	page3, err := suite.repo.GetBySessionIDWithPagination(context.Background(), suite.testSession.ID, 2, 4)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), page3, 1)
	assert.Equal(suite.T(), "Message 5", page3[0].Content)

	// Test pagination beyond available data
	page4, err := suite.repo.GetBySessionIDWithPagination(context.Background(), suite.testSession.ID, 2, 6)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), page4, 0)
}

func (suite *MessageRepositoryTestSuite) TestCountBySessionID() {
	// Initially should have 0 messages
	count, err := suite.repo.CountBySessionID(context.Background(), suite.testSession.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)

	// Create 3 messages
	for i := 1; i <= 3; i++ {
		message := &models.Message{
			SessionID: suite.testSession.ID,
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i),
		}
		err := suite.repo.Create(context.Background(), message)
		assert.NoError(suite.T(), err)
	}

	// Should now have 3 messages
	count, err = suite.repo.CountBySessionID(context.Background(), suite.testSession.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, count)
}

func (suite *MessageRepositoryTestSuite) TestMessageRoleValidation() {
	// Test valid roles
	validRoles := []string{"user", "assistant"}
	for _, role := range validRoles {
		message := &models.Message{
			SessionID: suite.testSession.ID,
			Role:      role,
			Content:   "Test message with role " + role,
		}
		err := suite.repo.Create(context.Background(), message)
		assert.NoError(suite.T(), err, "Should accept valid role: %s", role)
	}
}

func TestMessageRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MessageRepositoryTestSuite))
}