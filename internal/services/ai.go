package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"

	"github.com/sashabaranov/go-openai"
)

// MessageContext contains all the context needed for AI processing
type MessageContext struct {
	UserID              string
	SessionID           string
	Message             string
	ConversationHistory []*models.Message
	AthleteLogbook      *models.AthleteLogbook
	User                *models.User
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolCallID string      `json:"tool_call_id"`
	Content    string      `json:"content"`
	Error      string      `json:"error,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

// Custom error types for AI service
var (
	ErrOpenAIUnavailable   = errors.New("OpenAI service is currently unavailable")
	ErrOpenAIRateLimit     = errors.New("OpenAI API rate limit exceeded")
	ErrOpenAIQuotaExceeded = errors.New("OpenAI API quota exceeded")
	ErrInvalidInput        = errors.New("Invalid input provided")
	ErrContextTooLong      = errors.New("Conversation context is too long")
)

// AIService handles OpenAI integration and function calling
type AIService interface {
	ProcessMessage(ctx context.Context, msgCtx *MessageContext) (<-chan string, error)
	ProcessMessageSync(ctx context.Context, msgCtx *MessageContext) (string, error)
}

type aiService struct {
	client         *openai.Client
	stravaService  StravaService
	logbookService LogbookService
	config         *config.Config
}

// NewAIService creates a new AI service instance
func NewAIService(cfg *config.Config, stravaService StravaService, logbookService LogbookService) AIService {
	client := openai.NewClient(cfg.OpenAIAPIKey)

	return &aiService{
		client:         client,
		stravaService:  stravaService,
		logbookService: logbookService,
		config:         cfg,
	}
}

// ProcessMessage processes a user message and returns a streaming response channel
func (s *aiService) ProcessMessage(ctx context.Context, msgCtx *MessageContext) (<-chan string, error) {
	// Validate input
	if err := s.validateMessageContext(msgCtx); err != nil {
		return nil, err
	}

	messages := s.prepareMessages(msgCtx)
	tools := s.getAvailableTools()

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT4o,
		Messages: messages,
		Tools:    tools,
		Stream:   true,
		// Temperature: 0.7,
		// MaxTokens:   2000,
	}

	stream, err := s.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, s.handleOpenAIError(err)
	}

	responseChan := make(chan string, 100)

	go func() {
		defer close(responseChan)
		defer stream.Close()

		var toolCalls []openai.ToolCall
		var currentToolCall *openai.ToolCall
		var responseContent strings.Builder

		for {
			response, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}

				// Handle streaming errors gracefully
				aiErr := s.handleOpenAIError(err)
				if errors.Is(aiErr, ErrOpenAIUnavailable) {
					responseChan <- s.getFallbackResponse("I'm experiencing technical difficulties right now. Please try again in a moment.")
				} else {
					responseChan <- fmt.Sprintf("I encountered an error while processing your message: %v. Please try again.", aiErr)
				}
				return
			}

			if len(response.Choices) == 0 {
				continue
			}

			delta := response.Choices[0].Delta

			// Handle content streaming
			if delta.Content != "" {
				responseContent.WriteString(delta.Content)
				responseChan <- delta.Content
			}

			// Handle tool calls
			if len(delta.ToolCalls) > 0 {
				for _, toolCall := range delta.ToolCalls {
					if toolCall.Index != nil {
						// New tool call or existing one
						for len(toolCalls) <= *toolCall.Index {
							toolCalls = append(toolCalls, openai.ToolCall{})
						}
						currentToolCall = &toolCalls[*toolCall.Index]

						if toolCall.ID != "" {
							currentToolCall.ID = toolCall.ID
						}
						if toolCall.Type != "" {
							currentToolCall.Type = toolCall.Type
						}
						if toolCall.Function.Name != "" {
							currentToolCall.Function.Name = toolCall.Function.Name
						}
					}

					if currentToolCall != nil && toolCall.Function.Arguments != "" {
						currentToolCall.Function.Arguments += toolCall.Function.Arguments
					}
				}
			}
		}

		// Process tool calls if any
		if len(toolCalls) > 0 {
			toolResults, err := s.executeTools(ctx, msgCtx, toolCalls)
			if err != nil {
				responseChan <- fmt.Sprintf("Error executing tools: %v", err)
				return
			}

			// Create follow-up request with tool results
			followUpMessages := append(messages, openai.ChatCompletionMessage{
				Role:      openai.ChatMessageRoleAssistant,
				Content:   responseContent.String(),
				ToolCalls: toolCalls,
			})

			// Add tool results as messages
			for _, result := range toolResults {
				followUpMessages = append(followUpMessages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    result.Content,
					ToolCallID: result.ToolCallID,
				})
			}

			// Make follow-up request
			followUpReq := openai.ChatCompletionRequest{
				Model:    openai.GPT4o,
				Messages: followUpMessages,
				Stream:   true,
				// Temperature: 0.7,
				// MaxTokens:   2000,
			}

			followUpStream, err := s.client.CreateChatCompletionStream(ctx, followUpReq)
			if err != nil {
				aiErr := s.handleOpenAIError(err)
				responseChan <- fmt.Sprintf("I encountered an error while processing the tool results: %v. Please try again.", aiErr)
				return
			}
			defer followUpStream.Close()

			for {
				response, err := followUpStream.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					aiErr := s.handleOpenAIError(err)
					responseChan <- fmt.Sprintf("Error during response generation: %v", aiErr)
					return
				}

				if len(response.Choices) > 0 && response.Choices[0].Delta.Content != "" {
					responseChan <- response.Choices[0].Delta.Content
				}
			}
		}
	}()

	return responseChan, nil
}

// ProcessMessageSync processes a message synchronously and returns the complete response
func (s *aiService) ProcessMessageSync(ctx context.Context, msgCtx *MessageContext) (string, error) {
	// Validate input
	if err := s.validateMessageContext(msgCtx); err != nil {
		return "", err
	}

	messages := s.prepareMessages(msgCtx)
	tools := s.getAvailableTools()

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT4o,
		Messages: messages,
		Tools:    tools,
		// Temperature: 0.7,
		// MaxTokens:   2000,
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		aiErr := s.handleOpenAIError(err)
		if errors.Is(aiErr, ErrOpenAIUnavailable) {
			return s.getFallbackResponse("I'm experiencing technical difficulties right now. Please try again in a moment."), nil
		}
		return "", aiErr
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	choice := resp.Choices[0]

	// Handle tool calls
	if len(choice.Message.ToolCalls) > 0 {
		toolResults, err := s.executeTools(ctx, msgCtx, choice.Message.ToolCalls)
		if err != nil {
			return "", fmt.Errorf("failed to execute tools: %w", err)
		}

		// Create follow-up request with tool results
		followUpMessages := append(messages, choice.Message)

		// Add tool results as messages
		for _, result := range toolResults {
			followUpMessages = append(followUpMessages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result.Content,
				ToolCallID: result.ToolCallID,
			})
		}

		// Make follow-up request
		followUpReq := openai.ChatCompletionRequest{
			Model:       openai.GPT4o,
			Messages:    followUpMessages,
			Temperature: 0.7,
			MaxTokens:   2000,
		}

		followUpResp, err := s.client.CreateChatCompletion(ctx, followUpReq)
		if err != nil {
			aiErr := s.handleOpenAIError(err)
			if errors.Is(aiErr, ErrOpenAIUnavailable) {
				return s.getFallbackResponse("I processed your request but encountered an issue generating the final response. Please try again."), nil
			}
			return "", fmt.Errorf("failed to create follow-up completion: %w", aiErr)
		}

		if len(followUpResp.Choices) == 0 {
			return "", fmt.Errorf("no follow-up response choices returned")
		}

		return followUpResp.Choices[0].Message.Content, nil
	}

	return choice.Message.Content, nil
}

// prepareMessages converts the conversation history and current message into OpenAI format
func (s *aiService) prepareMessages(msgCtx *MessageContext) []openai.ChatCompletionMessage {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: s.buildSystemPrompt(msgCtx),
		},
	}

	// Add conversation history
	for _, msg := range msgCtx.ConversationHistory {
		role := openai.ChatMessageRoleUser
		if msg.Role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Add current message
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: msgCtx.Message,
	})

	return messages
}

// buildSystemPrompt creates the system prompt with athlete logbook context
func (s *aiService) buildSystemPrompt(msgCtx *MessageContext) string {
	basePrompt := `You are Bodda, an AI-powered running and/or cycling coach. You provide personalized coaching advice based on the athlete's Strava data and logbook information.

Your capabilities include:
- Analyzing Strava activity data (profile, recent activities, detailed activity information, and activity streams)
- Maintaining and updating athlete logbooks with training insights
- Providing personalized coaching recommendations
- Helping with training plans, performance analysis, and goal setting

Guidelines:
- Always be encouraging and supportive
- Base your advice on data when available
- Ask clarifying questions when you need more information
- Update the athlete logbook when you learn new information about the athlete
- Be specific in your recommendations and explain your reasoning
- Consider the athlete's goals, experience level, and current fitness when giving advice

Available tools:
- get-athlete-profile: Get complete Strava athlete profile
- get-recent-activities: Get recent activities (configurable count)
- get-activity-details: Get detailed information about a specific activity
- get-activity-streams: Get time-series data from an activity (heart rate, power, etc.)
- update-athlete-logbook: Update the athlete's logbook with new information`

	// Add athlete logbook context if available
	if msgCtx.AthleteLogbook != nil && msgCtx.AthleteLogbook.Content != "" {
		basePrompt += fmt.Sprintf("\n\nCurrent Athlete Logbook:\n%s", msgCtx.AthleteLogbook.Content)
	} else {
		basePrompt += "\n\nNo athlete logbook exists yet. You should create one using the update-athlete-logbook tool when you learn about the athlete."
	}

	return basePrompt
}

// getAvailableTools returns the OpenAI function definitions for available tools
func (s *aiService) getAvailableTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get-athlete-profile",
				Description: "Get the complete athlete profile from Strava including personal information, zones, and stats",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
					"required":   []string{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get-recent-activities",
				Description: "Get the most recent activities for the athlete",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"per_page": map[string]interface{}{
							"type":        "integer",
							"description": "Number of activities to retrieve (1-200, default 30)",
							"minimum":     1,
							"maximum":     200,
						},
					},
					"required": []string{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get-activity-details",
				Description: "Get detailed information about a specific activity using its ID",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"activity_id": map[string]interface{}{
							"type":        "integer",
							"description": "The Strava activity ID",
						},
					},
					"required": []string{"activity_id"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get-activity-streams",
				Description: "Get time-series data streams from a Strava activity (heart rate, power, cadence, etc.)",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"activity_id": map[string]interface{}{
							"type":        "integer",
							"description": "The Strava activity ID",
						},
						"stream_types": map[string]interface{}{
							"type":        "array",
							"description": "Types of streams to retrieve",
							"items": map[string]interface{}{
								"type": "string",
								"enum": []string{
									"time", "distance", "latlng", "altitude", "velocity_smooth",
									"heartrate", "cadence", "watts", "temp", "moving", "grade_smooth",
								},
							},
							"default": []string{"time", "distance", "heartrate", "watts"},
						},
						"resolution": map[string]interface{}{
							"type":        "string",
							"description": "Resolution of the data (low, medium, high)",
							"enum":        []string{"low", "medium", "high"},
							"default":     "medium",
						},
					},
					"required": []string{"activity_id"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "update-athlete-logbook",
				Description: "Update or create the athlete logbook with free-form string content. You can structure the content however you want - as plain text, markdown, or any format that makes sense for organizing athlete information, training insights, goals, and coaching observations.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"content": map[string]interface{}{
							"type":        "string",
							"description": "The complete logbook content as a string. You can structure this however you want - include athlete profile, training data, goals, preferences, health metrics, equipment, coaching insights, observations, and recommendations. Use any format that makes sense (plain text, markdown, etc.).",
						},
					},
					"required": []string{"content"},
				},
			},
		},
	}
}

// executeTools executes the tool calls and returns the results
func (s *aiService) executeTools(ctx context.Context, msgCtx *MessageContext, toolCalls []openai.ToolCall) ([]ToolResult, error) {
	var results []ToolResult

	for _, toolCall := range toolCalls {
		result := ToolResult{
			ToolCallID: toolCall.ID,
		}

		switch toolCall.Function.Name {
		case "get-athlete-profile":
			data, err := s.executeGetAthleteProfile(ctx, msgCtx)
			if err != nil {
				result.Error = err.Error()
				result.Content = fmt.Sprintf("Error getting athlete profile: %v", err)
			} else {
				result.Data = data
				jsonData, _ := json.Marshal(data)
				result.Content = string(jsonData)
			}

		case "get-recent-activities":
			var args struct {
				PerPage int `json:"per_page"`
			}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				result.Error = err.Error()
				result.Content = fmt.Sprintf("Error parsing arguments: %v", err)
			} else {
				if args.PerPage == 0 {
					args.PerPage = 30
				}
				data, err := s.executeGetRecentActivities(ctx, msgCtx, args.PerPage)
				if err != nil {
					result.Error = err.Error()
					result.Content = fmt.Sprintf("Error getting recent activities: %v", err)
				} else {
					result.Data = data
					jsonData, _ := json.Marshal(data)
					result.Content = string(jsonData)
				}
			}

		case "get-activity-details":
			var args struct {
				ActivityID int64 `json:"activity_id"`
			}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				result.Error = err.Error()
				result.Content = fmt.Sprintf("Error parsing arguments: %v", err)
			} else {
				data, err := s.executeGetActivityDetails(ctx, msgCtx, args.ActivityID)
				if err != nil {
					result.Error = err.Error()
					result.Content = fmt.Sprintf("Error getting activity details: %v", err)
				} else {
					result.Data = data
					jsonData, _ := json.Marshal(data)
					result.Content = string(jsonData)
				}
			}

		case "get-activity-streams":
			var args struct {
				ActivityID  int64    `json:"activity_id"`
				StreamTypes []string `json:"stream_types"`
				Resolution  string   `json:"resolution"`
			}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				result.Error = err.Error()
				result.Content = fmt.Sprintf("Error parsing arguments: %v", err)
			} else {
				if len(args.StreamTypes) == 0 {
					args.StreamTypes = []string{"time", "distance", "heartrate", "watts"}
				}
				if args.Resolution == "" {
					args.Resolution = "medium"
				}
				data, err := s.executeGetActivityStreams(ctx, msgCtx, args.ActivityID, args.StreamTypes, args.Resolution)
				if err != nil {
					result.Error = err.Error()
					result.Content = fmt.Sprintf("Error getting activity streams: %v", err)
				} else {
					result.Data = data
					jsonData, _ := json.Marshal(data)
					result.Content = string(jsonData)
				}
			}

		case "update-athlete-logbook":
			var args struct {
				Content string `json:"content"`
			}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				result.Error = err.Error()
				result.Content = fmt.Sprintf("Error parsing arguments: %v", err)
			} else {
				data, err := s.executeUpdateAthleteLogbook(ctx, msgCtx, args.Content)
				if err != nil {
					result.Error = err.Error()
					result.Content = fmt.Sprintf("Error updating athlete logbook: %v", err)
				} else {
					result.Data = data
					result.Content = "Athlete logbook updated successfully"
				}
			}

		default:
			result.Error = "unknown tool"
			result.Content = fmt.Sprintf("Unknown tool: %s", toolCall.Function.Name)
		}

		results = append(results, result)
	}

	return results, nil
}

// Tool execution methods

func (s *aiService) executeGetAthleteProfile(ctx context.Context, msgCtx *MessageContext) (*StravaAthlete, error) {
	if msgCtx.User == nil {
		return nil, fmt.Errorf("user context is required")
	}

	profile, err := s.stravaService.GetAthleteProfile(msgCtx.User.AccessToken)
	if err != nil {
		return nil, s.handleStravaError(err, "athlete profile")
	}

	return profile, nil
}

func (s *aiService) executeGetRecentActivities(ctx context.Context, msgCtx *MessageContext, perPage int) ([]*StravaActivity, error) {
	if msgCtx.User == nil {
		return nil, fmt.Errorf("user context is required")
	}

	params := ActivityParams{
		PerPage: perPage,
	}

	activities, err := s.stravaService.GetActivities(msgCtx.User.AccessToken, params)
	if err != nil {
		return nil, s.handleStravaError(err, "recent activities")
	}

	return activities, nil
}

func (s *aiService) executeGetActivityDetails(ctx context.Context, msgCtx *MessageContext, activityID int64) (*StravaActivityDetail, error) {
	if msgCtx.User == nil {
		return nil, fmt.Errorf("user context is required")
	}

	details, err := s.stravaService.GetActivityDetail(msgCtx.User.AccessToken, activityID)
	if err != nil {
		return nil, s.handleStravaError(err, "activity details")
	}

	return details, nil
}

func (s *aiService) executeGetActivityStreams(ctx context.Context, msgCtx *MessageContext, activityID int64, streamTypes []string, resolution string) (*StravaStreams, error) {
	if msgCtx.User == nil {
		return nil, fmt.Errorf("user context is required")
	}

	streams, err := s.stravaService.GetActivityStreams(msgCtx.User.AccessToken, activityID, streamTypes, resolution)
	if err != nil {
		return nil, s.handleStravaError(err, "activity streams")
	}

	return streams, nil
}

func (s *aiService) executeUpdateAthleteLogbook(ctx context.Context, msgCtx *MessageContext, content string) (*models.AthleteLogbook, error) {
	if msgCtx.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	if content == "" {
		return nil, fmt.Errorf("logbook content cannot be empty")
	}

	// Try to update existing logbook, or create if it doesn't exist
	logbook, err := s.logbookService.UpdateLogbook(ctx, msgCtx.UserID, content)
	if err != nil {
		// If logbook doesn't exist, try to create it using UpsertLogbook
		if strings.Contains(err.Error(), "not found") {
			logbook, err = s.logbookService.UpsertLogbook(ctx, msgCtx.UserID, content)
			if err != nil {
				return nil, fmt.Errorf("failed to create logbook: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to update logbook: %w", err)
		}
	}

	return logbook, nil
}

// validateMessageContext validates the message context before processing
func (s *aiService) validateMessageContext(msgCtx *MessageContext) error {
	if msgCtx == nil {
		return ErrInvalidInput
	}

	if msgCtx.UserID == "" {
		return fmt.Errorf("%w: user ID is required", ErrInvalidInput)
	}

	if msgCtx.SessionID == "" {
		return fmt.Errorf("%w: session ID is required", ErrInvalidInput)
	}

	if strings.TrimSpace(msgCtx.Message) == "" {
		return fmt.Errorf("%w: message content cannot be empty", ErrInvalidInput)
	}

	// Check if message is too long (approximate token count)
	if len(msgCtx.Message) > 8000 { // Rough estimate: 1 token ≈ 4 characters
		return fmt.Errorf("%w: message is too long", ErrInvalidInput)
	}

	// Check conversation history length
	if len(msgCtx.ConversationHistory) > 50 {
		return fmt.Errorf("%w: conversation history is too long", ErrContextTooLong)
	}

	return nil
}

// handleOpenAIError converts OpenAI errors to our custom error types
func (s *aiService) handleOpenAIError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	log.Printf("OpenAI API error: %v", err)

	// Check for specific OpenAI error types
	switch {
	case strings.Contains(errStr, "rate limit"):
		return ErrOpenAIRateLimit
	case strings.Contains(errStr, "quota"):
		return ErrOpenAIQuotaExceeded
	case strings.Contains(errStr, "timeout"):
		return ErrOpenAIUnavailable
	case strings.Contains(errStr, "connection"):
		return ErrOpenAIUnavailable
	case strings.Contains(errStr, "context_length_exceeded"):
		return ErrContextTooLong
	case strings.Contains(errStr, "invalid_request_error"):
		return ErrInvalidInput
	case strings.Contains(errStr, "service_unavailable"):
		return ErrOpenAIUnavailable
	case strings.Contains(errStr, "server_error"):
		return ErrOpenAIUnavailable
	default:
		return fmt.Errorf("OpenAI API error: %w", err)
	}
}

// getFallbackResponse returns a fallback response when AI is unavailable
func (s *aiService) getFallbackResponse(message string) string {
	fallbackResponses := []string{
		"I'm currently experiencing technical difficulties. Please try your question again in a moment.",
		"I'm having trouble connecting to my AI services right now. Please try again shortly.",
		"I'm temporarily unavailable due to technical issues. Please try your request again.",
		"I'm experiencing some technical problems at the moment. Please try again in a few minutes.",
	}

	if message != "" {
		return message
	}

	// Return a random fallback response (using time as seed for simplicity)
	index := int(time.Now().UnixNano()) % len(fallbackResponses)
	return fallbackResponses[index]
}

// handleStravaError converts Strava service errors to user-friendly messages
func (s *aiService) handleStravaError(err error, operation string) error {
	if err == nil {
		return nil
	}

	log.Printf("Strava API error during %s: %v", operation, err)

	switch {
	case errors.Is(err, ErrRateLimitExceeded):
		return fmt.Errorf("strava API rate limit exceeded. Please try again in a few minutes")
	case errors.Is(err, ErrTokenExpired):
		return fmt.Errorf("your Strava connection has expired. Please reconnect your Strava account")
	case errors.Is(err, ErrInvalidToken):
		return fmt.Errorf("invalid Strava credentials. Please reconnect your Strava account")
	case errors.Is(err, ErrActivityNotFound):
		return fmt.Errorf("the requested activity was not found or is not accessible")
	case errors.Is(err, ErrNetworkTimeout):
		return fmt.Errorf("connection to Strava timed out. Please try again")
	case errors.Is(err, ErrServiceUnavailable):
		return fmt.Errorf("strava service is temporarily unavailable. Please try again later")
	default:
		return fmt.Errorf("unable to retrieve %s from Strava: %v", operation, err)
	}
}
