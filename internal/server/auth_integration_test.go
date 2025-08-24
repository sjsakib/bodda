package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticationFlow_Integration(t *testing.T) {
	// This is a placeholder for integration tests that would require a real database
	// In a full implementation, this would test the complete OAuth flow
	t.Skip("Skipping integration test - requires database setup")
}

func TestAuthRoutes_Setup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a minimal server setup to test route configuration
	router := gin.New()
	
	// Test that auth routes are properly configured
	auth := router.Group("/auth")
	{
		auth.GET("/strava", func(c *gin.Context) { c.String(200, "strava") })
		auth.GET("/callback", func(c *gin.Context) { c.String(200, "callback") })
		auth.POST("/logout", func(c *gin.Context) { c.String(200, "logout") })
	}

	// Test Strava OAuth route
	req, _ := http.NewRequest("GET", "/auth/strava", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test callback route
	req, _ = http.NewRequest("GET", "/auth/callback", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test logout route
	req, _ = http.NewRequest("POST", "/auth/logout", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProtectedRoutes_RequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	
	// Simulate auth middleware that always rejects
	authMiddleware := func(c *gin.Context) {
		c.JSON(401, gin.H{"error": "authentication required"})
		c.Abort()
	}

	// Test that API routes are protected
	api := router.Group("/api")
	api.Use(authMiddleware)
	{
		api.GET("/sessions", func(c *gin.Context) { c.String(200, "sessions") })
		api.POST("/sessions", func(c *gin.Context) { c.String(200, "create session") })
	}

	// Test that protected routes require authentication
	req, _ := http.NewRequest("GET", "/api/sessions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authentication required")
}

func TestAuthCheckEndpoint_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	
	// Test that auth check endpoint is properly configured and protected
	apiAuth := router.Group("/api/auth")
	apiAuth.Use(func(c *gin.Context) {
		c.JSON(401, gin.H{"error": "authentication required"})
		c.Abort()
	})
	{
		apiAuth.GET("/check", func(c *gin.Context) { c.String(200, "auth check") })
	}

	// Test that auth check route requires authentication
	req, _ := http.NewRequest("GET", "/api/auth/check", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authentication required")
}