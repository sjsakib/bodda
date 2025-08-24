package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"bodda/internal/database"
	"bodda/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ChatServiceTestSuite struct {
	suite.Suite
	service  ChatService
	repo     *database.Repository
	db       *database.TestDB
	testUser *models.User
}

func (suite *ChatServiceTestSuite) SetupSuite() {
	suite.db = database.NewTestDB(suite.T())
	suite.repo = database.NewRepository(suite.db.Pool)
	suite.service = NewChatService(suite.repo)
}

func (suite *ChatServiceTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *ChatServiceTestSuite) SetupTest() {
	suite.db.CleanTables()
	
	// Create a test user for chat tests
	suite.testUser = &models.User{
		StravaID:     12345,
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenExpiry:  time.Now().Add(time.Hour),
		FirstName:    "John",
		LastName:     "Doe",
	}
	err := suite.repo.User.Create(context.Background(), suite.testUser)
	assert.NoError(suite.T(), err)
}

func (suite *ChatServiceTestSuite) TestCreateSession() {
	session, err := suite.service.CreateSession(suite.testUser.ID)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), session)
	assert.NotEmpty(suite.T(), session.ID)
	assert.Equal(suite.T(), suite.testUser.ID, session.UserID)
	assert.Equal(suite.T(), "New Conversation", session.Title)
	assert.NotZero(suite.T(), session.CreatedAt)
	assert.NotZero(suite.T(), session.UpdatedAt)
}

func (suite *ChatServiceTestSuite) TestCreateSessionWithTitle() {
	title := "Custom Session Title"
	session, err := suite.service.CreateSessionWithTitle(suite.testUser.ID, title)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), session)
	assert.Equal(suite.T(), title, session.Title)
	assert.Equal(suite.T(), suite.testUser.ID, session.UserID)
}

func (suite *ChatServiceTestSuite) TestGetSessions() {
	// Create multiple sessions
	session1, err := suite.service.CreateSessionWithTitle(suite.testUser.ID, "Session 1")
	assert.NoError(suite.T(), err)
	
	session2, err := suite.service.CreateSessionWithTitle(suite.testUser.ID, "Session 2")
	assert.NoError(suite.T(), err)
	
	// Get all sessions
	sessions, err := suite.service.GetSessions(suite.testUser.ID)
	
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 2)
	
	// Should be ordered by updated_at DESC (most recent first)
	sessionIDs := []string{sessions[0].ID, sessions[1].ID}
	assert.Contains(suite.T(), sessionIDs, session1.ID)
	assert.Contains(suite.T(), sessionIDs, session2.ID)
}

func (suite *ChatServiceTestSuite) TestGetSession() {
	// Create a session
	session, err := suite.service.CreateSession(suite.testUser.ID)
	assert.NoError(suite.T(), err)
	
	// Get the session
	retrievedSession, err := suite.service.GetSession(session.ID)
	
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), session.ID, retrievedSession.ID)
	assert.Equal(suite.T(), session.UserID, retrievedSession.UserID)
	assert.Equal(suite.T(), session.Title, retrievedSession.Title)
}

func (suite *ChatServiceTestSuite) TestGetNonExistentSession() {
	_, err := suite.service.GetSession("non-existent-id")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to get session")
}

func (suite *ChatServiceTestSuite) TestUpdateSessionTitle() {
	// Create a session
	session, err := suite.service.CreateSession(suite.testUser.ID)
	assert.NoError(suite.T(), err)
	
	// Update the title
	newTitle := "Updated Session Title"
	err = suite.service.UpdateSessionTitle(session.ID, newTitle)
	assert.NoError(suite.T(), err)
	
	// Verify the update
	updatedSession, err := suite.service.GetSession(session.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newTitle, updatedSession.Title)
}

func (suite *ChatServiceTestSuite) TestUpdateNonExistentSessionTitle() {
	err := suite.service.UpdateSessionTitle("non-existent-id", "New Title")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to get session")
}

func (suite *ChatServiceTestSuite) TestDeleteSession() {
	// Create a session
	session, err := suite.service.CreateSession(suite.testUser.ID)
	assert.NoError(suite.T(), err)
	
	// Delete the session
	err = suite.service.DeleteSession(session.ID)
	assert.NoError(suite.T(), err)
	
	// Verify the session is deleted
	_, err = suite.service.GetSession(session.ID)
	assert.Error(suite.T(), err)
}

func (suite *ChatServiceTestSuite) TestDeleteNonExistentSession() {
	err := suite.service.DeleteSession("non-existent-id")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to delete session")
}

func (suite *ChatServiceTestSuite) TestSendMessage() {
	// Create a session
	session, err := suite.service.CreateSession(suite.testUser.ID)
	assert.NoError(suite.T(), err)
	
	// Send a user message
	content := "Hello, this is a test message"
	message, err := suite.service.SendMessage(session.ID, "user", content)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), message)
	assert.NotEmpty(suite.T(), message.ID)
	assert.Equal(suite.T(), session.ID, message.SessionID)
	assert.Equal(suite.T(), "user", message.Role)
	assert.Equal(suite.T(), content, message.Content)
	assert.NotZero(suite.T(), message.CreatedAt)
}

func (suite *ChatServiceTestSuite) TestSendAssistantMessage() {
	// Create a session
	session, err := suite.service.CreateSession(suite.testUser.ID)
	assert.NoError(suite.T(), err)
	
	// Send an assistant message
	content := "Hello! I'm your AI coach."
	message, err := suite.service.SendMessage(session.ID, "assistant", content)
	
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "assistant", message.Role)
	assert.Equal(suite.T(), content, message.Content)
}

func (suite *ChatServiceTestSuite) TestSendMessageInvalidRole() {
	// Create a session
	session, err := suite.service.CreateSession(suite.testUser.ID)
	assert.NoError(suite.T(), err)
	
	// Try to send a message with invalid role
	_, err = suite.service.SendMessage(session.ID, "invalid", "test content")
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid role")
}

func (suite *ChatServiceTestSuite) TestSendMessageToNonExistentSession() {
	_, err := suite.service.SendMessage("non-existent-id", "user", "test content")
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "session not found")
}

func (suite *ChatServiceTestSuite) TestGetMessages() {
	// Create a session
	session, err := suite.service.CreateSession(suite.testUser.ID)
	assert.NoError(suite.T(), err)
	
	// Send multiple messages
	msg1, err := suite.service.SendMessage(session.ID, "user", "First message")
	assert.NoError(suite.T(), err)
	
	msg2, err := suite.service.SendMessage(session.ID, "assistant", "Assistant response")
	assert.NoError(suite.T(), err)
	
	msg3, err := suite.service.SendMessage(session.ID, "user", "Second user message")
	assert.NoError(suite.T(), err)
	
	// Get all messages
	messages, err := suite.service.GetMessages(session.ID)
	
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), messages, 3)
	
	// Should be in chronological order
	assert.Equal(suite.T(), msg1.ID, messages[0].ID)
	assert.Equal(suite.T(), msg2.ID, messages[1].ID)
	assert.Equal(suite.T(), msg3.ID, messages[2].ID)
}

func (suite *ChatServiceTestSuite) TestGetMessagesWithPagination() {
	// Create a session
	session, err := suite.service.CreateSession(suite.testUser.ID)
	assert.NoError(suite.T(), err)
	
	// Send 5 messages
	var sentMessages []*models.Message
	for i := 0; i < 5; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		msg, err := suite.service.SendMessage(session.ID, role, fmt.Sprintf("Message %d", i+1))
		assert.NoError(suite.T(), err)
		sentMessages = append(sentMessages, msg)
	}
	
	// Get first 2 messages
	messages, err := suite.service.GetMessagesWithPagination(session.ID, 2, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), messages, 2)
	assert.Equal(suite.T(), sentMessages[0].ID, messages[0].ID)
	assert.Equal(suite.T(), sentMessages[1].ID, messages[1].ID)
	
	// Get next 2 messages
	messages, err = suite.service.GetMessagesWithPagination(session.ID, 2, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), messages, 2)
	assert.Equal(suite.T(), sentMessages[2].ID, messages[0].ID)
	assert.Equal(suite.T(), sentMessages[3].ID, messages[1].ID)
	
	// Get last message
	messages, err = suite.service.GetMessagesWithPagination(session.ID, 2, 4)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), messages, 1)
	assert.Equal(suite.T(), sentMessages[4].ID, messages[0].ID)
}

func (suite *ChatServiceTestSuite) TestGetMessageCount() {
	// Create a session
	session, err := suite.service.CreateSession(suite.testUser.ID)
	assert.NoError(suite.T(), err)
	
	// Initially should have 0 messages
	count, err := suite.service.GetMessageCount(session.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)
	
	// Send 3 messages
	for i := 0; i < 3; i++ {
		_, err := suite.service.SendMessage(session.ID, "user", fmt.Sprintf("Message %d", i+1))
		assert.NoError(suite.T(), err)
	}
	
	// Should now have 3 messages
	count, err = suite.service.GetMessageCount(session.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, count)
}

func (suite *ChatServiceTestSuite) TestStreamResponse() {
	// Create a session
	session, err := suite.service.CreateSession(suite.testUser.ID)
	assert.NoError(suite.T(), err)
	
	// Test streaming (placeholder implementation)
	response := make(chan string, 1)
	err = suite.service.StreamResponse(session.ID, response)
	assert.NoError(suite.T(), err)
	
	// Should receive the placeholder message
	msg := <-response
	assert.Equal(suite.T(), "Streaming functionality will be implemented with AI integration", msg)
}

func (suite *ChatServiceTestSuite) TestStreamResponseNonExistentSession() {
	response := make(chan string, 1)
	err := suite.service.StreamResponse("non-existent-id", response)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "session not found")
}

func (suite *ChatServiceTestSuite) TestCompleteWorkflow() {
	// Test a complete chat workflow
	
	// 1. Create a session
	session, err := suite.service.CreateSession(suite.testUser.ID)
	assert.NoError(suite.T(), err)
	
	// 2. Send user message
	userMsg, err := suite.service.SendMessage(session.ID, "user", "Hello, I need help with my training")
	assert.NoError(suite.T(), err)
	
	// 3. Send assistant response
	assistantMsg, err := suite.service.SendMessage(session.ID, "assistant", "Hello! I'd be happy to help you with your training. What specific area would you like to focus on?")
	assert.NoError(suite.T(), err)
	
	// 4. Update session title based on conversation
	err = suite.service.UpdateSessionTitle(session.ID, "Training Help Session")
	assert.NoError(suite.T(), err)
	
	// 5. Verify the complete conversation
	messages, err := suite.service.GetMessages(session.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), messages, 2)
	assert.Equal(suite.T(), userMsg.ID, messages[0].ID)
	assert.Equal(suite.T(), assistantMsg.ID, messages[1].ID)
	
	// 6. Verify session title was updated
	updatedSession, err := suite.service.GetSession(session.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Training Help Session", updatedSession.Title)
	
	// 7. Verify message count
	count, err := suite.service.GetMessageCount(session.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, count)
}

func TestChatServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ChatServiceTestSuite))
}