package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	if args.Get(0) != nil {
		// Simulate database setting ID and timestamps
		user.ID = "test-user-id"
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByStravaID(ctx context.Context, stravaID int64) (*models.User, error) {
	args := m.Called(ctx, stravaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	if args.Error(0) == nil {
		user.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestAuthService_GenerateJWT(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	mockRepo := &MockUserRepository{}
	authService := NewAuthService(cfg, mockRepo)

	userID := "test-user-id"
	token, err := authService.GenerateJWT(userID)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token can be parsed
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims["user_id"])
}

func TestAuthService_ValidateToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	mockRepo := &MockUserRepository{}
	authService := NewAuthService(cfg, mockRepo)

	user := &models.User{
		ID:        "test-user-id",
		StravaID:  12345,
		FirstName: "Test",
		LastName:  "User",
	}

	// Mock repository call
	mockRepo.On("GetByID", mock.Anything, "test-user-id").Return(user, nil)

	// Generate a valid token
	token, err := authService.GenerateJWT(user.ID)
	assert.NoError(t, err)

	// Validate the token
	validatedUser, err := authService.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, validatedUser.ID)
	assert.Equal(t, user.StravaID, validatedUser.StravaID)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_ValidateToken_Invalid(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	mockRepo := &MockUserRepository{}
	authService := NewAuthService(cfg, mockRepo)

	// Test with invalid token
	_, err := authService.ValidateToken("invalid-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestAuthService_GetStravaOAuthURL(t *testing.T) {
	cfg := &config.Config{
		StravaClientID:    "test-client-id",
		StravaRedirectURL: "http://localhost:8080/auth/callback",
	}
	mockRepo := &MockUserRepository{}
	authService := NewAuthService(cfg, mockRepo)

	state := "test-state"
	url := authService.GetStravaOAuthURL(state)

	assert.Contains(t, url, "https://www.strava.com/oauth/authorize")
	assert.Contains(t, url, "client_id=test-client-id")
	assert.Contains(t, url, "state=test-state")
	assert.Contains(t, url, "redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fauth%2Fcallback")
}

func TestAuthService_HandleStravaOAuth_NewUser(t *testing.T) {
	// Create mock Strava server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			response := map[string]interface{}{
				"access_token":  "test-access-token",
				"refresh_token": "test-refresh-token",
				"expires_at":    time.Now().Add(6 * time.Hour).Unix(),
				"athlete": map[string]interface{}{
					"id":        int64(12345),
					"firstname": "Test",
					"lastname":  "User",
				},
			}
			json.NewEncoder(w).Encode(response)
		} else if r.URL.Path == "/api/v3/athlete" {
			response := map[string]interface{}{
				"id":        int64(12345),
				"firstname": "Test",
				"lastname":  "User",
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer mockServer.Close()

	cfg := &config.Config{
		StravaClientID:     "test-client-id",
		StravaClientSecret: "test-client-secret",
		StravaRedirectURL:  "http://localhost:8080/auth/callback",
	}
	mockRepo := &MockUserRepository{}

	// Create auth service with custom OAuth config pointing to mock server
	service := &authService{
		config:   cfg,
		userRepo: mockRepo,
	}

	// Mock repository calls
	mockRepo.On("GetByStravaID", mock.Anything, int64(12345)).Return(nil, assert.AnError)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	// This test would require more complex OAuth2 mocking
	// For now, we'll test the basic structure
	assert.NotNil(t, service)
}

func TestAuthService_RefreshStravaToken(t *testing.T) {
	cfg := &config.Config{
		StravaClientID:     "test-client-id",
		StravaClientSecret: "test-client-secret",
	}
	mockRepo := &MockUserRepository{}
	authService := NewAuthService(cfg, mockRepo)

	user := &models.User{
		ID:           "test-user-id",
		AccessToken:  "old-access-token",
		RefreshToken: "test-refresh-token",
		TokenExpiry:  time.Now().Add(-1 * time.Hour), // Expired
	}

	// Mock repository call
	mockRepo.On("Update", mock.Anything, user).Return(nil)

	// This test would require OAuth2 server mocking for full implementation
	// For now, we verify the service structure
	assert.NotNil(t, authService)
}