package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"
	"bodda/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock services for testing
type MockChatService struct {
	mock.Mock
}

func (m *MockChatService) CreateSession(userID string) (*models.Session, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockChatService) CreateSessionWithTitle(userID, title string) (*models.Session, error) {
	args := m.Called(userID, title)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockChatService) GetSessions(userID string) ([]*models.Session, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockChatService) GetSession(sessionID string) (*models.Session, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockChatService) UpdateSessionTitle(sessionID, title string) error {
	args := m.Called(sessionID, title)
	return args.Error(0)
}

func (m *MockChatService) DeleteSession(sessionID string) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *MockChatService) SendMessage(sessionID, role, content string) (*models.Message, error) {
	args := m.Called(sessionID, role, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockChatService) GetMessages(sessionID string) ([]*models.Message, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockChatService) GetMessagesWithPagination(sessionID string, limit, offset int) ([]*models.Message, error) {
	args := m.Called(sessionID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockChatService) GetMessageCount(sessionID string) (int, error) {
	args := m.Called(sessionID)
	return args.Int(0), args.Error(1)
}

func (m *MockChatService) StreamResponse(sessionID string, response chan string) error {
	args := m.Called(sessionID, response)
	return args.Error(0)
}

type MockAIService struct {
	mock.Mock
}

func (m *MockAIService) ProcessMessage(ctx context.Context, msgCtx *services.MessageContext) (<-chan string, error) {
	args := m.Called(ctx, msgCtx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan string), args.Error(1)
}

func (m *MockAIService) ProcessMessageSync(ctx context.Context, msgCtx *services.MessageContext) (string, error) {
	args := m.Called(ctx, msgCtx)
	return args.String(0), args.Error(1)
}

// Tool execution methods for the tool execution endpoint
func (m *MockAIService) ExecuteGetAthleteProfile(ctx context.Context, msgCtx *services.MessageContext) (string, error) {
	args := m.Called(ctx, msgCtx)
	return args.String(0), args.Error(1)
}

func (m *MockAIService) ExecuteGetRecentActivities(ctx context.Context, msgCtx *services.MessageContext, perPage int) (string, error) {
	args := m.Called(ctx, msgCtx, perPage)
	return args.String(0), args.Error(1)
}

func (m *MockAIService) ExecuteGetActivityDetails(ctx context.Context, msgCtx *services.MessageContext, activityID int64) (string, error) {
	args := m.Called(ctx, msgCtx, activityID)
	return args.String(0), args.Error(1)
}

func (m *MockAIService) ExecuteGetActivityStreams(ctx context.Context, msgCtx *services.MessageContext, activityID int64, streamTypes []string, resolution string, processingMode string, pageNumber int, pageSize int, summaryPrompt string) (string, error) {
	args := m.Called(ctx, msgCtx, activityID, streamTypes, resolution, processingMode, pageNumber, pageSize, summaryPrompt)
	return args.String(0), args.Error(1)
}

func (m *MockAIService) ExecuteUpdateAthleteLogbook(ctx context.Context, msgCtx *services.MessageContext, content string) (string, error) {
	args := m.Called(ctx, msgCtx, content)
	return args.String(0), args.Error(1)
}

type MockLogbookService struct {
	mock.Mock
}

func (m *MockLogbookService) GetLogbook(ctx context.Context, userID string) (*models.AthleteLogbook, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AthleteLogbook), args.Error(1)
}

func (m *MockLogbookService) CreateInitialLogbook(ctx context.Context, userID string, stravaProfile *services.StravaAthlete) (*models.AthleteLogbook, error) {
	args := m.Called(ctx, userID, stravaProfile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AthleteLogbook), args.Error(1)
}

func (m *MockLogbookService) UpdateLogbook(ctx context.Context, userID string, content string) (*models.AthleteLogbook, error) {
	args := m.Called(ctx, userID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AthleteLogbook), args.Error(1)
}

func (m *MockLogbookService) UpsertLogbook(ctx context.Context, userID string, content string) (*models.AthleteLogbook, error) {
	args := m.Called(ctx, userID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AthleteLogbook), args.Error(1)
}



// Helper function to create a test server with mocked services
func createTestServer() (*Server, *MockChatService, *MockAIService, *MockLogbookService) {
	gin.SetMode(gin.TestMode)

	mockChatService := &MockChatService{}
	mockAIService := &MockAIService{}
	mockLogbookService := &MockLogbookService{}

	server := &Server{
		config: &config.Config{
			FrontendURL: "http://localhost:3000",
		},
		router:         gin.New(),
		chatService:    mockChatService,
		aiService:      mockAIService,
		logbookService: mockLogbookService,
	}

	return server, mockChatService, mockAIService, mockLogbookService
}

// Helper function to create a test context with authenticated user
func createAuthenticatedContext(server *Server, method, path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	var req *http.Request
	if body != nil {
		req, _ = http.NewRequest(method, path, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Set authenticated user in context
	user := &models.User{
		ID:        "test-user-id",
		StravaID:  12345,
		FirstName: "Test",
		LastName:  "User",
	}
	c.Set("user", user)

	return c, w
}

func TestServer_getSessions_Success(t *testing.T) {
	server, mockChatService, _, _ := createTestServer()

	sessions := []*models.Session{
		{
			ID:        "session-1",
			UserID:    "test-user-id",
			Title:     "First Session",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "session-2",
			UserID:    "test-user-id",
			Title:     "Second Session",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	mockChatService.On("GetSessions", "test-user-id").Return(sessions, nil)

	c, w := createAuthenticatedContext(server, "GET", "/api/sessions", nil)
	server.getSessions(c)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "sessions")
	sessionList := response["sessions"].([]interface{})
	assert.Len(t, sessionList, 2)

	mockChatService.AssertExpectations(t)
}

func TestServer_getSessions_UserNotInContext(t *testing.T) {
	server, _, _, _ := createTestServer()

	req, _ := http.NewRequest("GET", "/api/sessions", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user in context

	server.getSessions(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "user not found in context")
}

func TestServer_createSession_Success(t *testing.T) {
	server, mockChatService, _, _ := createTestServer()

	session := &models.Session{
		ID:        "new-session-id",
		UserID:    "test-user-id",
		Title:     "Custom Title",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockChatService.On("CreateSessionWithTitle", "test-user-id", "Custom Title").Return(session, nil)

	requestBody := map[string]string{
		"title": "Custom Title",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	c, w := createAuthenticatedContext(server, "POST", "/api/sessions", bodyBytes)
	server.createSession(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "session")
	sessionData := response["session"].(map[string]interface{})
	assert.Equal(t, "new-session-id", sessionData["id"])
	assert.Equal(t, "Custom Title", sessionData["title"])

	mockChatService.AssertExpectations(t)
}

func TestServer_createSession_DefaultTitle(t *testing.T) {
	server, mockChatService, _, _ := createTestServer()

	session := &models.Session{
		ID:        "new-session-id",
		UserID:    "test-user-id",
		Title:     "New Conversation",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockChatService.On("CreateSession", "test-user-id").Return(session, nil)

	requestBody := map[string]string{}
	bodyBytes, _ := json.Marshal(requestBody)

	c, w := createAuthenticatedContext(server, "POST", "/api/sessions", bodyBytes)
	server.createSession(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "session")
	sessionData := response["session"].(map[string]interface{})
	assert.Equal(t, "new-session-id", sessionData["id"])
	assert.Equal(t, "New Conversation", sessionData["title"])

	mockChatService.AssertExpectations(t)
}

func TestServer_getMessages_Success(t *testing.T) {
	server, mockChatService, _, _ := createTestServer()

	session := &models.Session{
		ID:     "test-session-id",
		UserID: "test-user-id",
		Title:  "Test Session",
	}

	messages := []*models.Message{
		{
			ID:        "msg-1",
			SessionID: "test-session-id",
			Role:      "user",
			Content:   "Hello",
			CreatedAt: time.Now(),
		},
		{
			ID:        "msg-2",
			SessionID: "test-session-id",
			Role:      "assistant",
			Content:   "Hi there!",
			CreatedAt: time.Now(),
		},
	}

	mockChatService.On("GetSession", "test-session-id").Return(session, nil)
	mockChatService.On("GetMessages", "test-session-id").Return(messages, nil)

	c, w := createAuthenticatedContext(server, "GET", "/api/sessions/test-session-id/messages", nil)
	c.Params = []gin.Param{{Key: "id", Value: "test-session-id"}}
	server.getMessages(c)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "messages")
	messageList := response["messages"].([]interface{})
	assert.Len(t, messageList, 2)

	mockChatService.AssertExpectations(t)
}

func TestServer_getMessages_SessionNotFound(t *testing.T) {
	server, mockChatService, _, _ := createTestServer()

	mockChatService.On("GetSession", "nonexistent-session").Return(nil, assert.AnError)

	c, w := createAuthenticatedContext(server, "GET", "/api/sessions/nonexistent-session/messages", nil)
	c.Params = []gin.Param{{Key: "id", Value: "nonexistent-session"}}
	server.getMessages(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "session not found")

	mockChatService.AssertExpectations(t)
}

func TestServer_getMessages_AccessDenied(t *testing.T) {
	server, mockChatService, _, _ := createTestServer()

	session := &models.Session{
		ID:     "test-session-id",
		UserID: "different-user-id", // Different user ID
		Title:  "Test Session",
	}

	mockChatService.On("GetSession", "test-session-id").Return(session, nil)

	c, w := createAuthenticatedContext(server, "GET", "/api/sessions/test-session-id/messages", nil)
	c.Params = []gin.Param{{Key: "id", Value: "test-session-id"}}
	server.getMessages(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "access denied")

	mockChatService.AssertExpectations(t)
}

func TestServer_sendMessage_Success(t *testing.T) {
	server, mockChatService, mockAIService, mockLogbookService := createTestServer()

	session := &models.Session{
		ID:     "test-session-id",
		UserID: "test-user-id",
		Title:  "Test Session",
	}

	userMessage := &models.Message{
		ID:        "user-msg-id",
		SessionID: "test-session-id",
		Role:      "user",
		Content:   "Hello AI",
		CreatedAt: time.Now(),
	}

	assistantMessage := &models.Message{
		ID:        "assistant-msg-id",
		SessionID: "test-session-id",
		Role:      "assistant",
		Content:   "Hello! How can I help you?",
		CreatedAt: time.Now(),
	}

	messages := []*models.Message{userMessage}

	mockChatService.On("GetSession", "test-session-id").Return(session, nil)
	mockChatService.On("SendMessage", "test-session-id", "user", "Hello AI").Return(userMessage, nil)
	mockChatService.On("GetMessages", "test-session-id").Return(messages, nil)
	mockChatService.On("SendMessage", "test-session-id", "assistant", "Hello! How can I help you?").Return(assistantMessage, nil)

	mockLogbookService.On("GetLogbook", mock.Anything, "test-user-id").Return(nil, assert.AnError) // No logbook found

	mockAIService.On("ProcessMessageSync", mock.Anything, mock.AnythingOfType("*services.MessageContext")).Return("Hello! How can I help you?", nil)

	requestBody := map[string]string{
		"content": "Hello AI",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	c, w := createAuthenticatedContext(server, "POST", "/api/sessions/test-session-id/messages", bodyBytes)
	c.Params = []gin.Param{{Key: "id", Value: "test-session-id"}}
	server.sendMessage(c)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "user_message")
	assert.Contains(t, response, "assistant_message")

	mockChatService.AssertExpectations(t)
	mockAIService.AssertExpectations(t)
	mockLogbookService.AssertExpectations(t)
}

func TestServer_sendMessage_MissingContent(t *testing.T) {
	server, _, _, _ := createTestServer()

	requestBody := map[string]string{}
	bodyBytes, _ := json.Marshal(requestBody)

	c, w := createAuthenticatedContext(server, "POST", "/api/sessions/test-session-id/messages", bodyBytes)
	c.Params = []gin.Param{{Key: "id", Value: "test-session-id"}}
	server.sendMessage(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "message content is required")
}

func TestServer_streamResponse_MissingMessage(t *testing.T) {
	server, _, _, _ := createTestServer()

	c, w := createAuthenticatedContext(server, "GET", "/api/sessions/test-session-id/stream", nil)
	c.Params = []gin.Param{{Key: "id", Value: "test-session-id"}}
	server.streamResponse(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "message parameter is required")
}

func TestServer_streamResponse_SessionNotFound(t *testing.T) {
	server, mockChatService, _, _ := createTestServer()

	mockChatService.On("GetSession", "nonexistent-session").Return(nil, assert.AnError)

	c, w := createAuthenticatedContext(server, "GET", "/api/sessions/nonexistent-session/stream?message=test", nil)
	c.Params = []gin.Param{{Key: "id", Value: "nonexistent-session"}}
	c.Request.URL.RawQuery = "message=test"
	server.streamResponse(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockChatService.AssertExpectations(t)
}