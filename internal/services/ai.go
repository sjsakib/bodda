package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"strings"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"

	"github.com/sashabaranov/go-openai"
)

const MODEL = openai.GPT5

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

// IterativeProcessor manages multiple rounds of data analysis and tool execution
type IterativeProcessor struct {
	MaxRounds        int                            // Maximum tool call rounds (default: 5)
	CurrentRound     int                            // Current analysis round
	ProgressCallback func(string)                   // Stream progress updates
	ToolResults      [][]ToolResult                 // Results from each round
	Context          *MessageContext                // Persistent context
	Messages         []openai.ChatCompletionMessage // Accumulated conversation context
}

// NewIterativeProcessor creates a new iterative processor with default settings
func NewIterativeProcessor(msgCtx *MessageContext, progressCallback func(string)) *IterativeProcessor {
	return &IterativeProcessor{
		MaxRounds:        10,
		CurrentRound:     0,
		ProgressCallback: progressCallback,
		ToolResults:      make([][]ToolResult, 0),
		Context:          msgCtx,
		Messages:         make([]openai.ChatCompletionMessage, 0),
	}
}

// AddToolResults adds tool results for the current round and increments the round counter
func (ip *IterativeProcessor) AddToolResults(results []ToolResult) {
	ip.ToolResults = append(ip.ToolResults, results)
	ip.CurrentRound++
}

// GetTotalToolCalls returns the total number of tool calls across all rounds
func (ip *IterativeProcessor) GetTotalToolCalls() int {
	total := 0
	for _, roundResults := range ip.ToolResults {
		total += len(roundResults)
	}
	return total
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

	// Tool execution methods for the tool execution endpoint
	ExecuteGetAthleteProfile(ctx context.Context, msgCtx *MessageContext) (string, error)
	ExecuteGetRecentActivities(ctx context.Context, msgCtx *MessageContext, perPage int) (string, error)
	ExecuteGetActivityDetails(ctx context.Context, msgCtx *MessageContext, activityID int64) (string, error)
	ExecuteGetActivityStreams(ctx context.Context, msgCtx *MessageContext, activityID int64, streamTypes []string, resolution string, processingMode string, pageNumber int, pageSize int, summaryPrompt string) (string, error)
	ExecuteUpdateAthleteLogbook(ctx context.Context, msgCtx *MessageContext, content string) (string, error)
}

type aiService struct {
	client               *openai.Client
	stravaService        StravaService
	logbookService       LogbookService
	formatter            OutputFormatter
	config               *config.Config
	streamProcessor      StreamProcessor
	summaryProcessor     SummaryProcessor
	processingDispatcher ProcessingModeDispatcher
	unifiedProcessor     *UnifiedStreamProcessor
	contextManager       ContextManager
}

// NewAIService creates a new AI service instance
func NewAIService(cfg *config.Config, stravaService StravaService, logbookService LogbookService) AIService {
	client := openai.NewClient(cfg.OpenAIAPIKey)

	// Create stream processing components
	streamProcessor := NewStreamProcessor(cfg)
	summaryProcessor := NewSummaryProcessor(client)
	processingDispatcher := NewProcessingModeDispatcher(streamProcessor, summaryProcessor)

	// Create derived features processor and unified stream processor
	derivedProcessor := NewDerivedFeaturesProcessor()
	outputFormatter := NewOutputFormatter()
	unifiedProcessor := NewUnifiedStreamProcessor(
		&StreamConfig{
			MaxContextTokens:  cfg.StreamProcessing.MaxContextTokens,
			TokenPerCharRatio: cfg.StreamProcessing.TokenPerCharRatio,
			DefaultPageSize:   cfg.StreamProcessing.DefaultPageSize,
			MaxPageSize:       cfg.StreamProcessing.MaxPageSize,
			RedactionEnabled:  cfg.StreamProcessing.RedactionEnabled,
			StravaResolutions: []string{"low", "medium", "high"},
		},
		stravaService,
		derivedProcessor,
		summaryProcessor,
		outputFormatter,
	)

	// Create context manager for stream output redaction
	contextManager := NewContextManager(cfg.StreamProcessing.RedactionEnabled)

	return &aiService{
		client:               client,
		stravaService:        stravaService,
		logbookService:       logbookService,
		formatter:            outputFormatter,
		config:               cfg,
		streamProcessor:      streamProcessor,
		summaryProcessor:     summaryProcessor,
		processingDispatcher: processingDispatcher,
		unifiedProcessor:     unifiedProcessor,
		contextManager:       contextManager,
	}
}

// ProcessMessage processes a user message and returns a streaming response channel
func (s *aiService) ProcessMessage(ctx context.Context, msgCtx *MessageContext) (<-chan string, error) {
	// Validate input
	if err := s.validateMessageContext(msgCtx); err != nil {
		return nil, err
	}

	responseChan := make(chan string, 100)

	// Create iterative processor with progress callback
	processor := NewIterativeProcessor(msgCtx, func(message string) {
		responseChan <- message
	})

	go func() {
		defer close(responseChan)

		// Process message with iterative tool calling
		err := s.processIterativeToolCalls(ctx, processor, responseChan)
		if err != nil {
			aiErr := s.handleOpenAIError(err)
			if errors.Is(aiErr, ErrOpenAIUnavailable) {
				message := s.getRandomMessage([]string{
					"I'm having some difficulties right now. Please give me a moment and try your question again.",
					"I'm experiencing some connectivity issues at the moment. Please try reaching out again in a few minutes.",
					"There's a hiccup on my end right now. Please try your question again shortly.",
				})
				responseChan <- message
			} else {
				message := s.getRandomMessage([]string{
					"I ran into an issue while analyzing your training data. Please try your question again.",
					"Something went wrong while I was processing your request. Please give it another try.",
					"I encountered a problem while working on your question. Please try asking again.",
				})
				responseChan <- message
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

	// Use the same iterative processor logic as streaming, but collect all output
	var responseBuilder strings.Builder
	responseChan := make(chan string, 100)

	// Create iterative processor with response collection
	processor := NewIterativeProcessor(msgCtx, func(message string) {
		responseChan <- message
	})

	// Process in a goroutine and collect all responses
	go func() {
		defer close(responseChan)
		err := s.processIterativeToolCalls(ctx, processor, responseChan)
		if err != nil {
			aiErr := s.handleOpenAIError(err)
			if errors.Is(aiErr, ErrOpenAIUnavailable) {
				message := s.getRandomMessage([]string{
					"I'm experiencing technical difficulties right now. Please try again in a moment.",
					"I'm having some connectivity issues at the moment. Please try reaching out again in a few minutes.",
					"There's a hiccup on my end right now. Please try your question again shortly.",
				})
				responseChan <- message
			} else {
				message := s.getRandomMessage([]string{
					"I ran into an issue while analyzing your training data. Please try your question again.",
					"Something went wrong while I was processing your request. Please give it another try.",
					"I encountered a problem while working on your question. Please try asking again.",
				})
				responseChan <- message
			}
		}
	}()

	// Collect all responses
	for chunk := range responseChan {
		responseBuilder.WriteString(chunk)
	}

	return responseBuilder.String(), nil
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
				Description: "Get time-series data streams from a Strava activity with processing options for large datasets",
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
							"description": "Resolution of the data",
							"enum":        []string{"low", "medium", "high"},
							"default":     "medium",
						},
						"processing_mode": map[string]interface{}{
							"type":        "string",
							"description": "How to process large datasets that exceed context limits.\n- ai-summary: summarizes the stream with a small LLM model, recommended mode",
							"enum":        []string{"ai-summary"},
							"default":     "ai-summary",
						},
						"page_number": map[string]interface{}{
							"type":        "integer",
							"description": "Page number for paginated processing (1-based)",
							"minimum":     1,
							"default":     1,
						},
						"page_size": map[string]interface{}{
							"type":        "integer",
							"description": "Number of data points per page. Use negative value to request full dataset (subject to context limits)",
							"minimum":     -1,
							"maximum":     5000,
							"default":     1000,
						},
						"summary_prompt": map[string]interface{}{
							"type":        "string",
							"description": "Custom prompt for AI summarization mode (required when processing_mode is 'ai-summary')",
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

// processIterativeToolCalls handles multi-turn tool calling with progress streaming
func (s *aiService) processIterativeToolCalls(ctx context.Context, processor *IterativeProcessor, responseChan chan<- string) error {
	// Prepare initial messages with accumulated context
	processor.Messages = s.buildConversationContext(processor.Context)
	tools := s.getAvailableTools()

	for {
		// Apply context redaction before creating the request
		redactedMessages := s.contextManager.RedactPreviousStreamOutputs(processor.Messages)

		// Create chat completion request with enhanced context
		req := openai.ChatCompletionRequest{
			Model:    MODEL,
			Messages: redactedMessages,
			Tools:    tools,
			Stream:   true,
			// Temperature: 0.7,
		}

		messagesTotalLen := 0
		for _, message := range redactedMessages {
			messagesTotalLen += len(message.Content)
		}

		slog.InfoContext(ctx, "Calling LLM", "message_count", len(redactedMessages), "message_total_len", messagesTotalLen)

		stream, err := s.client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			return s.handleStreamingError(err, processor, responseChan)
		}

		var toolCalls []openai.ToolCall
		var currentToolCall *openai.ToolCall
		var responseContent strings.Builder
		var hasContent bool

		// Process streaming response with enhanced error handling
		for {
			response, err := stream.Recv()
			if err != nil {
				stream.Close()
				if err == io.EOF {
					break
				}
				return s.handleStreamingError(err, processor, responseChan)
			}

			if len(response.Choices) == 0 {
				continue
			}

			delta := response.Choices[0].Delta

			// Handle content streaming
			if delta.Content != "" {
				responseContent.WriteString(delta.Content)
				responseChan <- delta.Content
				hasContent = true
			}

			// Handle tool calls with improved parsing
			if len(delta.ToolCalls) > 0 {
				toolCalls = s.parseToolCallsFromDelta(delta.ToolCalls, toolCalls, &currentToolCall)
			}
		}

		stream.Close()

		// Determine next action based on tool calls and analysis depth
		if len(toolCalls) > 0 {
			// Check if we should continue with another round of analysis
			shouldContinue, reason := s.shouldContinueAnalysis(processor, toolCalls, hasContent)
			if !shouldContinue {
				// Provide final response based on accumulated insights
				finalResponse := s.generateFinalResponse(processor, reason, hasContent)
				if finalResponse != "" {
					responseChan <- finalResponse
				}
				break
			}

			// Stream natural progress update
			progressMsg := s.getCoachingProgressMessage(processor, toolCalls)
			responseChan <- fmt.Sprintf("\n\n*%s*\n\n", progressMsg)

			// Execute tools with enhanced error handling
			toolResults, err := s.executeToolsWithRecovery(ctx, processor.Context, toolCalls)
			if err != nil {
				return s.handleToolExecutionError(err, processor, responseChan)
			}

			// Build enhanced conversation context with accumulated insights
			processor = s.accumulateAnalysisContext(processor, toolCalls, toolResults, responseContent.String())

			// Continue to next iteration with enhanced context
			continue
		}

		// No tool calls, analysis complete
		break
	}

	return nil
}

// buildConversationContext creates enhanced conversation context with accumulated insights
func (s *aiService) buildConversationContext(msgCtx *MessageContext) []openai.ChatCompletionMessage {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: s.buildEnhancedSystemPrompt(msgCtx),
		},
	}

	// Add conversation history with context preservation
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

	// Add current message with analysis context
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: msgCtx.Message,
	})

	return messages
}

var systemPrompt = `You are Bodda, an elite running and/or cycling coach mentoring an athlete with access to their Strava profile and all of their activities. Your responses should look and feel like it is coming from an elite professional coach.

When asked about any particular workout, provide a thorough, data-driven assessment, combining both quantitative insights and textual interpretation. Begin your report with a written summary that highlights key findings and context. Add clear coaching feedback and personalized training recommendations. These should be practical, actionable, and grounded solely in the data provided—no assumptions or fabrications. Do not hide or sugarcoat weakness.

LOGBOOK MANAGEMENT:
- The logbook has NO predefined schema - you have complete freedom to structure it based on coaching best practices
- If no logbook exists, use appropriate tools to get athlete's profile and recent activities to create one and then save it with the provided tool.
- You should get last 30 activities for the logbook in addition to the athlete profile.
- You can update the logbook profile section when you determine it needs fresh data from Strava
- Whenever you think the logbook needs update, you should do it with the provided tool. It could be after analyzing an activity, providing suggestion, plan, athlete sharing their constraint, preference etc. All significant or useful info about the athlete should be in the logbook.
- The logbook is stored using the athlete's Strava ID, ensuring their data persists across login sessions
- Include the athlete's Strava ID and current timestamp in the logbook

COACHING APPROACH:
- Use the logbook context to provide personalized coaching based on the athlete's complete history
- Structure the logbook content however you think will be most effective for coaching

RESPONSE FORMAT:
- Your response will be rendered as markdown, so feel free to format your response using markdown when appropriate.

Available tools:
- get-athlete-profile: Get complete Strava athlete profile
- get-recent-activities: Get recent activities (configurable count)
- get-activity-details: Get detailed information about a specific activity
- get-activity-streams: Get time-series data from an activity (heart rate, power, etc.)
- update-athlete-logbook: Update the athlete's logbook with new information

**Your Final Goal**
Provide professional grade coaching to your athlete to help them improve their performance, achieve their goals. Make them feel good and inspire them to continue when they actually are making progress.`

// buildEnhancedSystemPrompt creates system prompt with iterative analysis guidance
func (s *aiService) buildEnhancedSystemPrompt(msgCtx *MessageContext) string {
	basePrompt := systemPrompt

	// Add athlete logbook context if available
	if msgCtx.AthleteLogbook != nil && msgCtx.AthleteLogbook.Content != "" {
		basePrompt += fmt.Sprintf("\n\nCurrent Athlete Logbook:\n%s", msgCtx.AthleteLogbook.Content)
	} else {
		basePrompt += "\n\nNo athlete logbook exists yet. You should create one."
	}

	return basePrompt
}

// shouldContinueAnalysis determines if another round of analysis should be performed
func (s *aiService) shouldContinueAnalysis(processor *IterativeProcessor, toolCalls []openai.ToolCall, hasContent bool) (bool, string) {
	// Don't continue if no tool calls
	if len(toolCalls) == 0 {
		return false, "no_tools"
	}

	// Don't continue if max rounds reached
	if processor.CurrentRound >= processor.MaxRounds {
		return false, "max_rounds"
	}

	// for now we will continue as long as there is tool calls and max round not reached
	return true, "continue_analysis"
}

// getCoachingProgressMessage returns natural coaching-focused progress messages
func (s *aiService) getCoachingProgressMessage(processor *IterativeProcessor, toolCalls []openai.ToolCall) string {
	// Determine message based on tool calls and current context
	if len(toolCalls) > 0 {
		return s.getContextualProgressMessage(processor, toolCalls)
	}

	// Fallback to round-based messages with coaching tone
	return s.getRoundBasedProgressMessage(processor)
}

// getContextualProgressMessage returns progress messages based on specific tool calls
func (s *aiService) getContextualProgressMessage(processor *IterativeProcessor, toolCalls []openai.ToolCall) string {
	// Analyze the combination of tool calls to provide contextual messages
	hasProfile := false
	hasActivities := false
	hasDetails := false
	hasStreams := false
	hasLogbookUpdate := false

	for _, toolCall := range toolCalls {
		switch toolCall.Function.Name {
		case "get-athlete-profile":
			hasProfile = true
		case "get-recent-activities":
			hasActivities = true
		case "get-activity-details":
			hasDetails = true
		case "get-activity-streams":
			hasStreams = true
		case "update-athlete-logbook":
			hasLogbookUpdate = true
		}
	}

	// Provide contextual messages based on the combination of tools being used
	if hasLogbookUpdate {
		return s.getRandomMessage([]string{
			"Updating your training insights with what I've learned...",
			"Recording important observations about your training...",
			"Adding new insights to your coaching profile...",
		})
	}

	if hasStreams {
		return s.getRandomMessage([]string{
			"Diving deep into your workout data to understand your performance patterns...",
			"Examining your heart rate and power data for detailed insights...",
			"Looking at the fine details of your training metrics...",
			"Analyzing your performance data to spot trends and opportunities...",
		})
	}

	if hasDetails {
		return s.getRandomMessage([]string{
			"Taking a closer look at your specific workouts...",
			"Examining the details of your recent training sessions...",
			"Getting a better understanding of your workout structure...",
			"Reviewing the specifics of your training efforts...",
			"Analyzing your training zones...",
			"Reviewing zone distribution...",
		})
	}

	if hasActivities {
		if processor.CurrentRound == 0 {
			return s.getRandomMessage([]string{
				"Reviewing your recent training activities...",
				"Looking at your recent workouts to understand your current training...",
				"Checking out what you've been up to in your training lately...",
				"Getting familiar with your recent training sessions...",
			})
		} else {
			return s.getRandomMessage([]string{
				"Gathering more details about your training history...",
				"Looking at additional activities to get the full picture...",
				"Expanding my view of your training patterns...",
			})
		}
	}

	if hasProfile {
		return s.getRandomMessage([]string{
			"Getting to know you better as an athlete...",
			"Learning about your training background and preferences...",
			"Understanding your athletic profile and goals...",
			"Familiarizing myself with your training setup...",
		})
	}

	// Default contextual message
	return s.getRoundBasedProgressMessage(processor)
}

// getRoundBasedProgressMessage provides round-based progress messages with coaching tone
func (s *aiService) getRoundBasedProgressMessage(processor *IterativeProcessor) string {
	roundMessages := [][]string{
		// Round 0 - Initial analysis
		{
			"Let me take a look at your training data...",
			"Starting to analyze your training information...",
			"Beginning my review of your athletic data...",
		},
		// Round 1 - Deeper dive
		{
			"Digging deeper into your training patterns...",
			"Looking for insights in your workout data...",
			"Examining your training trends more closely...",
		},
		// Round 2 - Detailed analysis
		{
			"Analyzing the details to understand your performance...",
			"Connecting the dots in your training data...",
			"Piecing together your training story...",
		},
		// Round 3 - Advanced insights
		{
			"Uncovering deeper insights about your training...",
			"Looking at the bigger picture of your athletic development...",
			"Finding patterns that will help optimize your training...",
		},
		// Round 4+ - Final analysis
		{
			"Putting together my final analysis...",
			"Synthesizing everything I've learned about your training...",
			"Preparing comprehensive recommendations based on your data...",
		},
	}

	roundIndex := processor.CurrentRound
	if roundIndex >= len(roundMessages) {
		roundIndex = len(roundMessages) - 1
	}

	return s.getRandomMessage(roundMessages[roundIndex])
}

// getRandomMessage returns a random message from the provided slice
func (s *aiService) getRandomMessage(messages []string) string {
	if len(messages) == 0 {
		return "Continuing my analysis..."
	}

	// Use current time as a simple randomization method
	index := int(time.Now().UnixNano()) % len(messages)
	return messages[index]
}

// generateFinalResponse creates appropriate final response based on analysis state
func (s *aiService) generateFinalResponse(processor *IterativeProcessor, reason string, hasContent bool) string {
	if hasContent {
		// AI already provided some content, no need for additional response
		return ""
	}

	switch reason {
	case "max_rounds":
		return s.getRandomMessage([]string{
			"I've thoroughly reviewed your training data. Here's what I found and what I recommend:",
			"After analyzing your training comprehensively, here are my insights and suggestions:",
			"Based on my complete review of your training, here's my coaching advice:",
		})
	case "sufficient_data":
		return s.getRandomMessage([]string{
			"Perfect! I have everything I need. Here's my analysis and recommendations:",
			"Great! Based on what I've learned about your training, here's what I suggest:",
			"Excellent! I've got a clear picture now. Here are my coaching insights:",
		})
	case "no_tools":
		return s.getRandomMessage([]string{
			"Let me share my thoughts based on our conversation:",
			"Here's my coaching perspective on what you've shared:",
			"Based on what you've told me, here's what I think:",
		})
	default:
		return s.getRandomMessage([]string{
			"Here are my insights based on your training data:",
			"Let me share what I've learned about your training:",
			"Based on my analysis, here's my coaching advice:",
		})
	}
}

// executeToolsWithRecovery executes tools with enhanced error recovery
func (s *aiService) executeToolsWithRecovery(ctx context.Context, msgCtx *MessageContext, toolCalls []openai.ToolCall) ([]ToolResult, error) {
	results, err := s.executeTools(ctx, msgCtx, toolCalls)
	if err != nil {
		log.Printf("Tool execution error: %v", err)
	}

	// Filter and recover from partial failures
	var successfulResults []ToolResult
	var failedCount int

	for _, result := range results {
		if result.Error == "" {
			successfulResults = append(successfulResults, result)
		} else {
			failedCount++
			log.Printf("Tool call failed: %s - %s", result.ToolCallID, result.Error)
		}
	}

	// If we have some successful results, continue with those
	if len(successfulResults) > 0 {
		if failedCount > 0 {
			log.Printf("Continuing with %d successful results out of %d total", len(successfulResults), len(results))
		}
		return successfulResults, nil
	}

	// All tools failed
	if err != nil {
		return nil, fmt.Errorf("all tool calls failed: %w", err)
	}

	// If we reach here, all tools failed but executeTools didn't return an error
	// This means all results have Error fields set
	return nil, fmt.Errorf("all tool calls failed")
}

// accumulateAnalysisContext builds enhanced context with accumulated insights
func (s *aiService) accumulateAnalysisContext(processor *IterativeProcessor, toolCalls []openai.ToolCall, toolResults []ToolResult, responseContent string) *IterativeProcessor {
	// Add tool results to processor
	processor.AddToolResults(toolResults)

	// Add assistant message with tool calls to conversation
	processor.Messages = append(processor.Messages, openai.ChatCompletionMessage{
		Role:      openai.ChatMessageRoleAssistant,
		Content:   responseContent,
		ToolCalls: toolCalls,
	})

	// Add tool results as messages with enhanced context
	for _, result := range toolResults {
		processor.Messages = append(processor.Messages, openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    result.Content,
			ToolCallID: result.ToolCallID,
		})
	}

	return processor
}

// parseToolCallsFromDelta parses tool calls from streaming delta with improved handling
func (s *aiService) parseToolCallsFromDelta(deltaToolCalls []openai.ToolCall, existingToolCalls []openai.ToolCall, currentToolCall **openai.ToolCall) []openai.ToolCall {
	for _, toolCall := range deltaToolCalls {
		if toolCall.Index != nil {
			// New tool call or existing one
			for len(existingToolCalls) <= *toolCall.Index {
				existingToolCalls = append(existingToolCalls, openai.ToolCall{})
			}
			*currentToolCall = &existingToolCalls[*toolCall.Index]

			if toolCall.ID != "" {
				(*currentToolCall).ID = toolCall.ID
			}
			if toolCall.Type != "" {
				(*currentToolCall).Type = toolCall.Type
			}
			if toolCall.Function.Name != "" {
				(*currentToolCall).Function.Name = toolCall.Function.Name
			}
		}

		if *currentToolCall != nil && toolCall.Function.Arguments != "" {
			(*currentToolCall).Function.Arguments += toolCall.Function.Arguments
		}
	}
	return existingToolCalls
}

// handleStreamingError handles streaming errors with appropriate recovery
func (s *aiService) handleStreamingError(err error, processor *IterativeProcessor, responseChan chan<- string) error {
	log.Printf("Streaming error in round %d: %v", processor.CurrentRound, err)

	aiErr := s.handleOpenAIError(err)
	if errors.Is(aiErr, ErrOpenAIUnavailable) {
		message := s.getRandomMessage([]string{
			"\n\nI'm having some difficulties right now. Let me work with what I've already learned about your training to give you some insights.",
			"\n\nI'm experiencing some connectivity issues, but I can still help you based on the training data I've already reviewed.",
			"\n\nThere's a hiccup on my end, but let me share what I've discovered about your training so far.",
		})
		responseChan <- message
		return nil // Continue with partial analysis
	}

	return aiErr
}

// handleToolExecutionError handles tool execution errors during iterative analysis
func (s *aiService) handleToolExecutionError(err error, processor *IterativeProcessor, responseChan chan<- string) error {
	log.Printf("Tool execution error in round %d: %v", processor.CurrentRound, err)

	// If we have some previous results, try to provide coaching based on that
	if processor.GetTotalToolCalls() > 0 {
		message := s.getRandomMessage([]string{
			"\n\nI'm having trouble accessing some of your training data right now, but let me work with what I've already gathered to give you some helpful insights.",
			"\n\nThere's an issue connecting to your training data at the moment, but I can still provide coaching advice based on what I've already reviewed.",
			"\n\nI'm experiencing some difficulties accessing additional training information, but let me share recommendations based on the data I've already analyzed.",
		})
		responseChan <- message
		return nil
	}

	// No previous data, return error
	return fmt.Errorf("unable to gather training data: %w", err)
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
			content, err := s.executeGetAthleteProfile(ctx, msgCtx)
			if err != nil {
				result.Error = err.Error()
				result.Content = fmt.Sprintf("Error getting athlete profile: %v", err)
			} else {
				result.Content = content
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
				content, err := s.executeGetRecentActivities(ctx, msgCtx, args.PerPage)
				if err != nil {
					result.Error = err.Error()
					result.Content = fmt.Sprintf("Error getting recent activities: %v", err)
				} else {
					result.Content = content
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
				content, err := s.executeGetActivityDetails(ctx, msgCtx, args.ActivityID)
				if err != nil {
					result.Error = err.Error()
					result.Content = fmt.Sprintf("Error getting activity details: %v", err)
				} else {
					result.Content = content
				}
			}

		case "get-activity-streams":
			var args struct {
				ActivityID     int64    `json:"activity_id"`
				StreamTypes    []string `json:"stream_types"`
				Resolution     string   `json:"resolution"`
				ProcessingMode string   `json:"processing_mode"`
				PageNumber     int      `json:"page_number"`
				PageSize       int      `json:"page_size"`
				SummaryPrompt  string   `json:"summary_prompt"`
			}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				result.Error = err.Error()
				result.Content = fmt.Sprintf("Error parsing arguments: %v", err)
			} else {
				// Set defaults
				if len(args.StreamTypes) == 0 {
					args.StreamTypes = []string{"time", "distance", "heartrate", "watts"}
				}
				if args.Resolution == "" {
					args.Resolution = "medium"
				}
				if args.ProcessingMode == "" {
					args.ProcessingMode = "ai-summary"
				}
				if args.PageNumber == 0 {
					args.PageNumber = 1
				}
				if args.PageSize == 0 {
					args.PageSize = 1000
				}

				// Validate ai-summary mode requires summary_prompt
				if args.ProcessingMode == "ai-summary" && args.SummaryPrompt == "" {
					result.Error = "summary_prompt is required when processing_mode is 'ai-summary'"
					result.Content = "Error: summary_prompt parameter is required when using ai-summary processing mode"
				} else {
					processedResult, err := s.executeGetActivityStreamsWithProcessing(ctx, msgCtx, args.ActivityID, args.StreamTypes, args.Resolution, args.ProcessingMode, args.PageNumber, args.PageSize, args.SummaryPrompt)
					if err != nil {
						result.Error = err.Error()
						result.Content = fmt.Sprintf("Error getting activity streams: %v", err)
					} else {
						result.Data = processedResult.Data
						result.Content = processedResult.Content
					}
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

func (s *aiService) executeGetAthleteProfile(ctx context.Context, msgCtx *MessageContext) (string, error) {
	if msgCtx.User == nil {
		return "", fmt.Errorf("user context is required")
	}

	profile, err := s.stravaService.GetAthleteProfile(msgCtx.User)
	if err != nil {
		return "", s.handleStravaError(err, "athlete profile")
	}

	return s.formatter.FormatAthleteProfile(profile), nil
}

func (s *aiService) executeGetRecentActivities(ctx context.Context, msgCtx *MessageContext, perPage int) (string, error) {
	if msgCtx.User == nil {
		return "", fmt.Errorf("user context is required")
	}

	params := ActivityParams{
		PerPage: perPage,
	}

	activities, err := s.stravaService.GetActivities(msgCtx.User, params)
	if err != nil {
		return "", s.handleStravaError(err, "recent activities")
	}

	return s.formatter.FormatActivities(activities), nil
}

func (s *aiService) executeGetActivityDetails(ctx context.Context, msgCtx *MessageContext, activityID int64) (string, error) {
	if msgCtx.User == nil {
		return "", fmt.Errorf("user context is required")
	}

	// Use the integrated method to get activity details with zones
	detailsWithZones, err := s.stravaService.GetActivityDetailWithZones(msgCtx.User, activityID)
	if err != nil {
		return "", s.handleStravaError(err, "activity details")
	}

	// Format the integrated activity details and zones data
	return s.formatter.FormatActivityDetailsWithZones(detailsWithZones), nil
}

func (s *aiService) executeGetActivityStreams(ctx context.Context, msgCtx *MessageContext, activityID int64, streamTypes []string, resolution string) (*StravaStreams, error) {
	if msgCtx.User == nil {
		return nil, fmt.Errorf("user context is required")
	}

	streams, err := s.stravaService.GetActivityStreams(msgCtx.User, activityID, streamTypes, resolution)
	if err != nil {
		return nil, s.handleStravaError(err, "activity streams")
	}

	return streams, nil
}

func (s *aiService) executeGetActivityStreamsWithProcessing(ctx context.Context, msgCtx *MessageContext, activityID int64, streamTypes []string, resolution string, processingMode string, pageNumber int, pageSize int, summaryPrompt string) (*ProcessedStreamResult, error) {
	if msgCtx.User == nil {
		return nil, fmt.Errorf("user context is required")
	}

	// Handle auto mode - check if data needs processing first
	if processingMode == "auto" {
		// Get a sample of the data to determine if processing is needed
		streams, err := s.stravaService.GetActivityStreams(msgCtx.User, activityID, streamTypes, "low")
		if err != nil {
			return nil, s.handleStravaError(err, "activity streams")
		}

		// Use stream processor to determine if processing is needed
		if !s.streamProcessor.ShouldProcess(streams) {
			// Data is small enough, get full resolution and return raw formatted data
			fullStreams, err := s.stravaService.GetActivityStreams(msgCtx.User, activityID, streamTypes, resolution)
			if err != nil {
				return nil, s.handleStravaError(err, "activity streams")
			}

			content := s.formatter.FormatStreamData(fullStreams, "raw")
			return &ProcessedStreamResult{
				ToolCallID:     fmt.Sprintf("streams_%d", activityID),
				Content:        content,
				ProcessingMode: "raw",
			}, nil
		}

		// Data is too large, return processing options with current context information
		currentContextTokens := s.estimateCurrentContextTokens(msgCtx)
		result, err := s.streamProcessor.ProcessStreamOutputWithContext(streams, fmt.Sprintf("streams_%d", activityID), currentContextTokens)
		if err != nil {
			return nil, fmt.Errorf("failed to process stream output: %w", err)
		}
		return result, nil
	}

	// For specific processing modes, use the unified stream processor
	req := &PaginatedStreamRequest{
		ActivityID:     activityID,
		StreamTypes:    streamTypes,
		Resolution:     resolution,
		ProcessingMode: processingMode,
		PageNumber:     pageNumber,
		PageSize:       pageSize,
		SummaryPrompt:  summaryPrompt,
	}

	// Estimate current context tokens (simplified estimation)
	currentContextTokens := s.estimateCurrentContextTokens(msgCtx)

	// Process the paginated stream request
	streamPage, err := s.unifiedProcessor.ProcessPaginatedStreamRequest(msgCtx.User, req, currentContextTokens)
	if err != nil {
		log.Printf("Unified stream processing failed for activity %d with mode %s: %v", activityID, processingMode, err)
		return nil, fmt.Errorf("failed to process stream data: %w", err)
	}

	// Format the result using the unified processor's formatter
	content := s.unifiedProcessor.FormatPaginatedResult(streamPage)

	log.Printf("Successfully processed stream data for activity %d with mode %s (page %d/%d)",
		activityID, processingMode, streamPage.PageNumber, streamPage.TotalPages)

	return &ProcessedStreamResult{
		ToolCallID:     fmt.Sprintf("streams_%d", activityID),
		Content:        content,
		ProcessingMode: processingMode,
		Data:           streamPage,
	}, nil
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

// estimateCurrentContextTokens estimates the current context usage for pagination decisions
func (s *aiService) estimateCurrentContextTokens(msgCtx *MessageContext) int {
	// Simple estimation based on conversation history
	totalChars := 0

	// Count characters in conversation history
	for _, msg := range msgCtx.ConversationHistory {
		totalChars += len(msg.Content)
	}

	// Add current message
	totalChars += len(msgCtx.Message)

	// Add system prompt (rough estimate)
	totalChars += 2000

	// Convert to tokens using the configured ratio
	estimatedTokens := int(float64(totalChars) * s.config.StreamProcessing.TokenPerCharRatio)

	log.Printf("Estimated current context tokens: %d (based on %d characters)", estimatedTokens, totalChars)
	return estimatedTokens
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

// Exported tool execution methods for the tool execution endpoint

// ExecuteGetAthleteProfile executes the get-athlete-profile tool
func (s *aiService) ExecuteGetAthleteProfile(ctx context.Context, msgCtx *MessageContext) (string, error) {
	return s.executeGetAthleteProfile(ctx, msgCtx)
}

// ExecuteGetRecentActivities executes the get-recent-activities tool
func (s *aiService) ExecuteGetRecentActivities(ctx context.Context, msgCtx *MessageContext, perPage int) (string, error) {
	return s.executeGetRecentActivities(ctx, msgCtx, perPage)
}

// ExecuteGetActivityDetails executes the get-activity-details tool
func (s *aiService) ExecuteGetActivityDetails(ctx context.Context, msgCtx *MessageContext, activityID int64) (string, error) {
	return s.executeGetActivityDetails(ctx, msgCtx, activityID)
}

// ExecuteGetActivityStreams executes the get-activity-streams tool
func (s *aiService) ExecuteGetActivityStreams(ctx context.Context, msgCtx *MessageContext, activityID int64, streamTypes []string, resolution string, processingMode string, pageNumber int, pageSize int, summaryPrompt string) (string, error) {
	// Handle different processing modes
	if processingMode == "auto" || processingMode == "summary" || processingMode == "detailed" || processingMode == "paginated" || processingMode == "ai-summary" || processingMode == "raw" || processingMode == "derived" {
		// Use the processing version
		result, err := s.executeGetActivityStreamsWithProcessing(ctx, msgCtx, activityID, streamTypes, resolution, processingMode, pageNumber, pageSize, summaryPrompt)
		if err != nil {
			return "", err
		}

		// Return the content from the processed result
		return result.Content, nil
	} else {
		// Use the basic version
		streams, err := s.executeGetActivityStreams(ctx, msgCtx, activityID, streamTypes, resolution)
		if err != nil {
			return "", err
		}

		// Convert streams to string format using proper formatter
		return s.formatter.FormatStreamData(streams, "raw"), nil
	}
}

// ExecuteUpdateAthleteLogbook executes the update-athlete-logbook tool
func (s *aiService) ExecuteUpdateAthleteLogbook(ctx context.Context, msgCtx *MessageContext, content string) (string, error) {
	logbook, err := s.executeUpdateAthleteLogbook(ctx, msgCtx, content)
	if err != nil {
		return "", err
	}

	// Return a success message with the updated content
	return fmt.Sprintf("Logbook updated successfully. Content: %s", logbook.Content), nil
}
