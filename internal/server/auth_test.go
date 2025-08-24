package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bodda/internal/config"
	"bodda/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) HandleStravaOAuth(code string) (*models.User, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) ValidateToken(token string) (*models.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) RefreshStravaToken(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockAuthService) GenerateJWT(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GetStravaOAuthURL(state string) string {
	args := m.Called(state)
	return args.String(0)
}

func TestServer_handleStravaOAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := &MockAuthService{}
	server := &Server{
		config: &config.Config{
			FrontendURL: "http://localhost:3000",
		},
		authService: mockAuthService,
		router:      gin.New(),
	}

	expectedURL := "https://www.strava.com/oauth/authorize?client_id=test&state=random-state-string"
	mockAuthService.On("GetStravaOAuthURL", "random-state-string").Return(expectedURL)

	req, _ := http.NewRequest("GET", "/auth/strava", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	server.handleStravaOAuth(c)

	assert.Equal(t, http.StatusFound, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestServer_handleStravaCallback_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := &MockAuthService{}
	server := &Server{
		config: &config.Config{
			FrontendURL: "http://localhost:3000",
		},
		authService: mockAuthService,
		router:      gin.New(),
	}

	user := &models.User{
		ID:        "test-user-id",
		StravaID:  12345,
		FirstName: "Test",
		LastName:  "User",
	}

	mockAuthService.On("HandleStravaOAuth", "test-code").Return(user, nil)
	mockAuthService.On("GenerateJWT", "test-user-id").Return("test-jwt-token", nil)

	req, _ := http.NewRequest("GET", "/auth/callback?code=test-code", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	server.handleStravaCallback(c)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "http://localhost:3000/chat")
	
	// Check that auth cookie was set
	cookies := w.Result().Cookies()
	var authCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "auth_token" {
			authCookie = cookie
			break
		}
	}
	assert.NotNil(t, authCookie)
	assert.Equal(t, "test-jwt-token", authCookie.Value)
	assert.True(t, authCookie.HttpOnly)

	mockAuthService.AssertExpectations(t)
}

func TestServer_handleStravaCallback_MissingCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := &MockAuthService{}
	server := &Server{
		authService: mockAuthService,
		router:      gin.New(),
	}

	req, _ := http.NewRequest("GET", "/auth/callback", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	server.handleStravaCallback(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "authorization code not provided")
}

func TestServer_handleLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := &Server{
		router: gin.New(),
	}

	req, _ := http.NewRequest("POST", "/auth/logout", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	server.handleLogout(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "logged out successfully")

	// Check that auth cookie was cleared
	cookies := w.Result().Cookies()
	var authCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "auth_token" {
			authCookie = cookie
			break
		}
	}
	assert.NotNil(t, authCookie)
	assert.Equal(t, "", authCookie.Value)
	assert.Equal(t, -1, authCookie.MaxAge)
}

func TestServer_authMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := &MockAuthService{}
	server := &Server{
		authService: mockAuthService,
		router:      gin.New(),
	}

	user := &models.User{
		ID:        "test-user-id",
		StravaID:  12345,
		FirstName: "Test",
		LastName:  "User",
	}

	mockAuthService.On("ValidateToken", "valid-token").Return(user, nil)

	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	middleware := server.authMiddleware()
	middleware(c)

	assert.False(t, c.IsAborted())
	
	// Check that user was set in context
	contextUser, exists := c.Get("user")
	assert.True(t, exists)
	assert.Equal(t, user, contextUser)

	mockAuthService.AssertExpectations(t)
}

func TestServer_authMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := &MockAuthService{}
	server := &Server{
		authService: mockAuthService,
		router:      gin.New(),
	}

	mockAuthService.On("ValidateToken", "invalid-token").Return(nil, assert.AnError)

	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	middleware := server.authMiddleware()
	middleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")

	mockAuthService.AssertExpectations(t)
}

func TestServer_authMiddleware_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := &Server{
		router: gin.New(),
	}

	req, _ := http.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	middleware := server.authMiddleware()
	middleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authentication required")
}

func TestServer_authMiddleware_CookieAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := &MockAuthService{}
	server := &Server{
		authService: mockAuthService,
		router:      gin.New(),
	}

	user := &models.User{
		ID:        "test-user-id",
		StravaID:  12345,
		FirstName: "Test",
		LastName:  "User",
	}

	mockAuthService.On("ValidateToken", "cookie-token").Return(user, nil)

	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: "cookie-token",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	middleware := server.authMiddleware()
	middleware(c)

	assert.False(t, c.IsAborted())
	
	// Check that user was set in context
	contextUser, exists := c.Get("user")
	assert.True(t, exists)
	assert.Equal(t, user, contextUser)

	mockAuthService.AssertExpectations(t)
}

func TestServer_handleAuthCheck_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := &Server{
		router: gin.New(),
	}

	user := &models.User{
		ID:        "test-user-id",
		StravaID:  12345,
		FirstName: "Test",
		LastName:  "User",
	}

	req, _ := http.NewRequest("GET", "/api/auth/check", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user", user) // Simulate auth middleware setting user

	server.handleAuthCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"authenticated":true`)
	assert.Contains(t, w.Body.String(), `"id":"test-user-id"`)
	assert.Contains(t, w.Body.String(), `"strava_id":12345`)
	assert.Contains(t, w.Body.String(), `"first_name":"Test"`)
	assert.Contains(t, w.Body.String(), `"last_name":"User"`)
}

func TestServer_handleAuthCheck_UserNotInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := &Server{
		router: gin.New(),
	}

	req, _ := http.NewRequest("GET", "/api/auth/check", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user in context to simulate missing user

	server.handleAuthCheck(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "user not found in context")
}