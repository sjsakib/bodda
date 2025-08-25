package server

import (
	"bodda/internal/config"
	"bodda/internal/database"
	"bodda/internal/models"
	"bodda/internal/services"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"
	"strings"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	config         *config.Config
	db             *pgxpool.Pool
	router         *gin.Engine
	authService    services.AuthService
	chatService    services.ChatService
	aiService      services.AIService
	stravaService  services.StravaService
	logbookService services.LogbookService
	repo           *database.Repository
}

func New(cfg *config.Config, db *pgxpool.Pool) *Server {
	// Initialize repositories
	repo := database.NewRepository(db)
	
	// Initialize services
	authService := services.NewAuthService(cfg, repo.User)
	stravaService := services.NewStravaService(cfg, repo.User)
	logbookService := services.NewLogbookService(repo.Logbook)
	chatService := services.NewChatService(repo)
	aiService := services.NewAIService(cfg, stravaService, logbookService)

	s := &Server{
		config:         cfg,
		db:             db,
		router:         gin.Default(),
		authService:    authService,
		chatService:    chatService,
		aiService:      aiService,
		stravaService:  stravaService,
		logbookService: logbookService,
		repo:           repo,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Request logging middleware
	s.router.Use(s.requestLoggingMiddleware())

	// Error recovery middleware
	s.router.Use(s.errorRecoveryMiddleware())

	// CORS middleware
	s.router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", s.config.FrontendURL)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes
	auth := s.router.Group("/auth")
	{
		auth.GET("/strava", s.handleStravaOAuth)
		auth.GET("/callback", s.handleStravaCallback)
		auth.POST("/logout", s.handleLogout)
	}

	// API auth routes (require authentication)
	apiAuth := s.router.Group("/api/auth")
	apiAuth.Use(s.authMiddleware())
	{
		apiAuth.GET("/check", s.handleAuthCheck)
	}

	// API routes
	api := s.router.Group("/api")
	api.Use(s.authMiddleware())
	{
		api.GET("/sessions", s.getSessions)
		api.POST("/sessions", s.createSession)
		api.GET("/sessions/:id/messages", s.getMessages)
		api.POST("/sessions/:id/messages", s.sendMessage)
		api.GET("/sessions/:id/stream", s.streamResponse)
	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

// Authentication handlers
func (s *Server) handleStravaOAuth(c *gin.Context) {
	state := "random-state-string" // In production, use a secure random state
	url := s.authService.GetStravaOAuthURL(state)
	c.Redirect(302, url)
}

func (s *Server) handleStravaCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(400, gin.H{"error": "authorization code not provided"})
		return
	}

	// Handle OAuth callback
	user, err := s.authService.HandleStravaOAuth(code)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to authenticate with Strava", "details": err.Error()})
		return
	}

	// Generate JWT token
	token, err := s.authService.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate session token"})
		return
	}

	// Set JWT as HTTP-only cookie
	c.SetCookie("auth_token", token, 86400, "/", "", false, true) // 24 hours

	// Redirect to frontend
	c.Redirect(302, s.config.FrontendURL+"/chat")
}

func (s *Server) handleLogout(c *gin.Context) {
	// Clear the auth cookie
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	c.JSON(200, gin.H{"message": "logged out successfully"})
}

func (s *Server) handleAuthCheck(c *gin.Context) {
	// Get user from context (set by auth middleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "user not found in context"})
		return
	}

	// Return user information (excluding sensitive data)
	userModel := user.(*models.User)
	c.JSON(200, gin.H{
		"authenticated": true,
		"user": gin.H{
			"id":         userModel.ID,
			"strava_id":  userModel.StravaID,
			"first_name": userModel.FirstName,
			"last_name":  userModel.LastName,
		},
	})
}

// Request logging middleware
func (s *Server) requestLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// Error recovery middleware
func (s *Server) errorRecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("Panic recovered: %v", recovered)
		
		// Check if client disconnected
		if c.Writer.Written() {
			return
		}
		
		c.JSON(500, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
	})
}

// Authentication middleware
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get token from cookie first
		token, err := c.Cookie("auth_token")
		if err != nil {
			// Fallback to Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(401, gin.H{"error": "authentication required"})
				c.Abort()
				return
			}

			// Extract token from "Bearer <token>" format
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token = authHeader[7:]
			} else {
				c.JSON(401, gin.H{"error": "invalid authorization header format"})
				c.Abort()
				return
			}
		}

		// Validate token
		user, err := s.authService.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Store user in context
		c.Set("user", user)
		c.Next()
	}
}

// Session management handlers

// getSessions retrieves all sessions for the authenticated user
func (s *Server) getSessions(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{
			"error": "Authentication required",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	userModel := user.(*models.User)
	sessions, err := s.chatService.GetSessions(userModel.ID)
	if err != nil {
		log.Printf("Error getting sessions for user %s: %v", userModel.ID, err)
		
		// Handle specific error types
		if strings.Contains(err.Error(), "connection") {
			c.JSON(503, gin.H{
				"error": "Database temporarily unavailable",
				"code":  "DATABASE_UNAVAILABLE",
			})
			return
		}
		
		c.JSON(500, gin.H{
			"error": "Failed to retrieve sessions",
			"code":  "SESSION_RETRIEVAL_ERROR",
		})
		return
	}

	c.JSON(200, gin.H{"sessions": sessions})
}

// createSession creates a new chat session for the authenticated user
func (s *Server) createSession(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{
			"error": "Authentication required",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	var req struct {
		Title string `json:"title"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	userModel := user.(*models.User)
	
	var session *models.Session
	var err error
	
	if req.Title != "" {
		session, err = s.chatService.CreateSessionWithTitle(userModel.ID, req.Title)
	} else {
		session, err = s.chatService.CreateSession(userModel.ID)
	}

	if err != nil {
		log.Printf("Error creating session for user %s: %v", userModel.ID, err)
		
		// Handle specific error types
		if errors.Is(err, services.ErrInvalidSessionTitle) {
			c.JSON(400, gin.H{
				"error": "Invalid session title",
				"code":  "INVALID_TITLE",
			})
			return
		}
		
		if strings.Contains(err.Error(), "connection") {
			c.JSON(503, gin.H{
				"error": "Database temporarily unavailable",
				"code":  "DATABASE_UNAVAILABLE",
			})
			return
		}
		
		c.JSON(500, gin.H{
			"error": "Failed to create session",
			"code":  "SESSION_CREATION_ERROR",
		})
		return
	}

	c.JSON(201, gin.H{"session": session})
}

// getMessages retrieves all messages for a specific session
func (s *Server) getMessages(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(400, gin.H{"error": "session ID is required"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "user not found in context"})
		return
	}

	userModel := user.(*models.User)

	// Verify session belongs to user
	session, err := s.chatService.GetSession(sessionID)
	if err != nil {
		log.Printf("Error getting session %s: %v", sessionID, err)
		c.JSON(404, gin.H{"error": "session not found"})
		return
	}

	if session.UserID != userModel.ID {
		c.JSON(403, gin.H{"error": "access denied"})
		return
	}

	// Get pagination parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var messages []*models.Message
	if limit == 50 && offset == 0 {
		// Use default method for common case
		messages, err = s.chatService.GetMessages(sessionID)
	} else {
		// Use pagination method
		messages, err = s.chatService.GetMessagesWithPagination(sessionID, limit, offset)
	}

	if err != nil {
		log.Printf("Error getting messages for session %s: %v", sessionID, err)
		c.JSON(500, gin.H{"error": "failed to retrieve messages"})
		return
	}

	c.JSON(200, gin.H{"messages": messages})
}

// sendMessage sends a message in a chat session and processes it with AI
func (s *Server) sendMessage(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(400, gin.H{"error": "session ID is required"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "user not found in context"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}
	
	// Validate message content
	if strings.TrimSpace(req.Content) == "" {
		c.JSON(400, gin.H{
			"error": "Message content cannot be empty",
			"code":  "EMPTY_MESSAGE",
		})
		return
	}

	userModel := user.(*models.User)

	// Verify session belongs to user
	session, err := s.chatService.GetSession(sessionID)
	if err != nil {
		log.Printf("Error getting session %s: %v", sessionID, err)
		c.JSON(404, gin.H{"error": "session not found"})
		return
	}

	if session.UserID != userModel.ID {
		c.JSON(403, gin.H{"error": "access denied"})
		return
	}

	// Save user message
	userMessage, err := s.chatService.SendMessage(sessionID, "user", req.Content)
	if err != nil {
		log.Printf("Error saving user message: %v", err)
		
		// Handle specific error types
		if errors.Is(err, services.ErrMessageTooLong) {
			c.JSON(400, gin.H{
				"error": "Message is too long",
				"code":  "MESSAGE_TOO_LONG",
			})
			return
		}
		
		if errors.Is(err, services.ErrInvalidMessageContent) {
			c.JSON(400, gin.H{
				"error": "Invalid message content",
				"code":  "INVALID_CONTENT",
			})
			return
		}
		
		if errors.Is(err, services.ErrSessionNotFound) {
			c.JSON(404, gin.H{
				"error": "Session not found",
				"code":  "SESSION_NOT_FOUND",
			})
			return
		}
		
		c.JSON(500, gin.H{
			"error": "Failed to save message",
			"code":  "MESSAGE_SAVE_ERROR",
		})
		return
	}

	// Get conversation history for AI context
	messages, err := s.chatService.GetMessages(sessionID)
	if err != nil {
		log.Printf("Error getting conversation history: %v", err)
		c.JSON(500, gin.H{"error": "failed to get conversation history"})
		return
	}

	// Process message with AI
	ctx := context.Background()
	
	// Get athlete logbook for AI context
	logbook, err := s.logbookService.GetLogbook(ctx, userModel.ID)
	if err != nil {
		// Logbook might not exist yet, that's okay
		log.Printf("No logbook found for user %s: %v", userModel.ID, err)
	}

	msgCtx := &services.MessageContext{
		UserID:              userModel.ID,
		SessionID:           sessionID,
		Message:             req.Content,
		ConversationHistory: messages[:len(messages)-1], // Exclude the just-added user message
		AthleteLogbook:      logbook,
		User:                userModel,
	}

	// Get AI response synchronously for this endpoint
	aiResponse, err := s.aiService.ProcessMessageSync(ctx, msgCtx)
	if err != nil {
		log.Printf("Error processing AI message: %v", err)
		
		// Handle specific AI error types
		if errors.Is(err, services.ErrOpenAIUnavailable) {
			c.JSON(503, gin.H{
				"error": "AI service temporarily unavailable",
				"code":  "AI_UNAVAILABLE",
			})
			return
		}
		
		if errors.Is(err, services.ErrOpenAIRateLimit) {
			c.JSON(429, gin.H{
				"error": "AI service rate limit exceeded",
				"code":  "AI_RATE_LIMIT",
			})
			return
		}
		
		if errors.Is(err, services.ErrContextTooLong) {
			c.JSON(400, gin.H{
				"error": "Conversation is too long",
				"code":  "CONTEXT_TOO_LONG",
			})
			return
		}
		
		c.JSON(500, gin.H{
			"error": "Failed to process message",
			"code":  "AI_PROCESSING_ERROR",
		})
		return
	}

	// Save AI response
	assistantMessage, err := s.chatService.SendMessage(sessionID, "assistant", aiResponse)
	if err != nil {
		log.Printf("Error saving AI response: %v", err)
		c.JSON(500, gin.H{
			"error": "Failed to save AI response",
			"code":  "RESPONSE_SAVE_ERROR",
		})
		return
	}

	c.JSON(200, gin.H{
		"user_message":      userMessage,
		"assistant_message": assistantMessage,
	})
}

// streamResponse provides Server-Sent Events streaming for AI responses
func (s *Server) streamResponse(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(400, gin.H{"error": "session ID is required"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "user not found in context"})
		return
	}

	// Get message content from query parameter first
	message := c.Query("message")
	if message == "" {
		c.JSON(400, gin.H{"error": "message parameter is required"})
		return
	}

	userModel := user.(*models.User)

	// Verify session belongs to user
	session, err := s.chatService.GetSession(sessionID)
	if err != nil {
		log.Printf("Error getting session %s: %v", sessionID, err)
		c.JSON(404, gin.H{"error": "session not found"})
		return
	}

	if session.UserID != userModel.ID {
		c.JSON(403, gin.H{"error": "access denied"})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Save user message first
	userMessage, err := s.chatService.SendMessage(sessionID, "user", message)
	if err != nil {
		log.Printf("Error saving user message: %v", err)
		c.SSEvent("error", "failed to save message")
		return
	}

	// Send user message event
	userMessageEvent := map[string]interface{}{
		"type":    "user_message",
		"message": userMessage,
	}
	c.SSEvent("message", userMessageEvent)
	c.Writer.Flush()

	// Get conversation history for AI context
	messages, err := s.chatService.GetMessages(sessionID)
	if err != nil {
		log.Printf("Error getting conversation history: %v", err)
		errorEvent := map[string]interface{}{
			"type":    "error",
			"message": "failed to get conversation history",
		}
		c.SSEvent("message", errorEvent)
		return
	}

	// Process message with AI streaming
	ctx := context.Background()
	
	// Get athlete logbook for AI context
	logbook, err := s.logbookService.GetLogbook(ctx, userModel.ID)
	if err != nil {
		// Logbook might not exist yet, that's okay
		log.Printf("No logbook found for user %s: %v", userModel.ID, err)
	}

	msgCtx := &services.MessageContext{
		UserID:              userModel.ID,
		SessionID:           sessionID,
		Message:             message,
		ConversationHistory: messages[:len(messages)-1], // Exclude the just-added user message
		AthleteLogbook:      logbook,
		User:                userModel,
	}

	responseChan, err := s.aiService.ProcessMessage(ctx, msgCtx)
	if err != nil {
		log.Printf("Error processing AI message: %v", err)
		errorEvent := map[string]interface{}{
			"type":    "error",
			"message": "failed to process message with AI",
		}
		c.SSEvent("message", errorEvent)
		return
	}

	// Stream AI response
	var fullResponse string
	for chunk := range responseChan {
		fullResponse += chunk
		chunkEvent := map[string]interface{}{
			"type":    "chunk",
			"content": chunk,
		}
		c.SSEvent("message", chunkEvent)
		c.Writer.Flush()
	}

	// Save complete AI response
	assistantMessage, err := s.chatService.SendMessage(sessionID, "assistant", fullResponse)
	if err != nil {
		log.Printf("Error saving AI response: %v", err)
		errorEvent := map[string]interface{}{
			"type":    "error",
			"message": "failed to save AI response",
		}
		c.SSEvent("message", errorEvent)
		return
	}

	// Send completion event
	completeEvent := map[string]interface{}{
		"type":    "complete",
		"message": assistantMessage,
	}
	c.SSEvent("message", completeEvent)
}