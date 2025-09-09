package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/packages/param"
	"github.com/openai/openai-go/v2/packages/ssestream"
	"github.com/openai/openai-go/v2/responses"
)

// SessionRepository interface for session operations needed by AI service
type SessionRepository interface {
	UpdateLastResponseID(ctx context.Context, sessionID string, responseID string) error
}

// MessageContext contains all the context needed for AI processing
type MessageContext struct {
	UserID              string
	SessionID           string
	Message             string
	ConversationHistory []*models.Message
	AthleteLogbook      *models.AthleteLogbook
	User                *models.User
	LastResponseID      string // OpenAI Response ID for multi-turn conversations
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
	MaxRounds        int                                     // Maximum tool call rounds (default: 5)
	CurrentRound     int                                     // Current analysis round
	ProgressCallback func(string)                            // Stream progress updates
	ToolResults      [][]ToolResult                          // Results from each round
	Context          *MessageContext                         // Persistent context
	Messages         []responses.ResponseInputItemUnionParam // Accumulated conversation context
}

// NewIterativeProcessor creates a new iterative processor with default settings
func NewIterativeProcessor(msgCtx *MessageContext, progressCallback func(string)) *IterativeProcessor {
	return &IterativeProcessor{
		MaxRounds:        10,
		CurrentRound:     0,
		ProgressCallback: progressCallback,
		ToolResults:      make([][]ToolResult, 0),
		Context:          msgCtx,
		Messages:         make([]responses.ResponseInputItemUnionParam, 0),
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
	client               openai.Client // Official OpenAI SDK client
	stravaService        StravaService
	logbookService       LogbookService
	sessionRepository    SessionRepository
	formatter            OutputFormatter
	config               *config.Config
	streamProcessor      StreamProcessor
	summaryProcessor     SummaryProcessor
	processingDispatcher ProcessingModeDispatcher
	unifiedProcessor     *UnifiedStreamProcessor
	contextManager       ContextManager
	toolRegistry         ToolRegistry
}

// NewAIService creates a new AI service instance
func NewAIService(cfg *config.Config, stravaService StravaService, logbookService LogbookService, sessionRepository SessionRepository, toolRegistry ToolRegistry) AIService {
	// Initialize OpenAI client
	client := openai.NewClient(option.WithAPIKey(cfg.OpenAIAPIKey))

	// Create stream processing components
	streamProcessor := NewStreamProcessor(cfg)
	summaryProcessor := NewSummaryProcessor(&client)
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

	slog.Info("AI Service initialized with Responses API", "implementation", "responses_api")

	return &aiService{
		client:               client,
		stravaService:        stravaService,
		logbookService:       logbookService,
		sessionRepository:    sessionRepository,
		formatter:            outputFormatter,
		config:               cfg,
		streamProcessor:      streamProcessor,
		summaryProcessor:     summaryProcessor,
		processingDispatcher: processingDispatcher,
		unifiedProcessor:     unifiedProcessor,
		contextManager:       contextManager,
		toolRegistry:         toolRegistry,
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

		slog.InfoContext(ctx, "Using Responses API implementation for message processing",
			"user_id", msgCtx.UserID,
			"session_id", msgCtx.SessionID,
			"implementation", "responses_api")

		err := s.processMessageWithResponsesAPI(ctx, processor, responseChan)

		if err != nil {
			aiErr := s.handleResponsesAPIError(err)
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

		slog.InfoContext(ctx, "Using Responses API implementation for sync message processing",
			"user_id", msgCtx.UserID,
			"session_id", msgCtx.SessionID,
			"implementation", "responses_api",
			"processing_mode", "sync")

		err := s.processMessageWithResponsesAPI(ctx, processor, responseChan)

		if err != nil {
			aiErr := s.handleResponsesAPIError(err)
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
func (s *aiService) getAvailableTools() []responses.ToolUnionParam {
	// Get tools from registry with error handling and fallback behavior
	if s.toolRegistry == nil {
		slog.Warn("Tool registry is not available, returning empty tool list")
		return []responses.ToolUnionParam{}
	}

	registryTools := s.toolRegistry.GetAvailableTools()
	if len(registryTools) == 0 {
		slog.Warn("No tools available from registry, returning empty tool list")
		return []responses.ToolUnionParam{}
	}

	// Convert registry tools to OpenAI format
	openAITools := make([]responses.ToolUnionParam, 0, len(registryTools))
	for _, tool := range registryTools {
		convertedTool := s.convertToolDefinitionToOpenAI(tool)
		openAITools = append(openAITools, convertedTool)

		slog.Debug("Converted tool from registry to OpenAI format",
			"tool_name", tool.Name,
			"description", tool.Description)
	}

	slog.Info("Successfully loaded tools from registry",
		"tool_count", len(openAITools),
		"implementation", "registry_based")

	return openAITools
}

// convertToolDefinitionToOpenAI converts a models.ToolDefinition to OpenAI's responses.ToolUnionParam format
func (s *aiService) convertToolDefinitionToOpenAI(tool models.ToolDefinition) responses.ToolUnionParam {
	// Create a copy of the parameters to modify for OpenAI compatibility
	params := make(map[string]interface{})
	for k, v := range tool.Parameters {
		params[k] = v
	}

	// OpenAI requires ALL properties to be listed in the required array
	// even if they're optional parameters with defaults
	if properties, ok := params["properties"].(map[string]interface{}); ok && len(properties) > 0 {
		allPropertyNames := make([]string, 0, len(properties))
		for propName := range properties {
			allPropertyNames = append(allPropertyNames, propName)
		}
		params["required"] = allPropertyNames
	}

	return responses.ToolParamOfFunction(
		tool.Name,
		params,
		true, // Enable strict mode for parameter validation
	)
}

// processMessageWithResponsesAPI handles multi-turn tool calling using the Responses API
func (s *aiService) processMessageWithResponsesAPI(ctx context.Context, processor *IterativeProcessor, responseChan chan<- string) error {
	// Prepare initial messages with accumulated context
	if len(processor.Messages) == 0 {
		processor.Messages = s.buildConversationContextForResponsesAPI(processor.Context)
	}
	tools := s.getAvailableTools()

	for {
		// Build input items for this iteration
		var inputItems []responses.ResponseInputItemUnionParam
		
		// For first iteration or when no response ID is available, include conversation context
		if processor.Context.LastResponseID == "" {
			inputItems = processor.Messages
		} else {
			// For subsequent iterations with response ID, only include new user message and tool results
			inputItems = append(inputItems, responses.ResponseInputItemParamOfMessage(
				processor.Context.Message,
				responses.EasyInputMessageRoleUser,
			))
			
			// Add tool results from previous iteration if any
			if processor.CurrentRound > 0 && len(processor.ToolResults) > 0 {
				lastRoundResults := processor.ToolResults[len(processor.ToolResults)-1]
				for _, result := range lastRoundResults {
					if result.ToolCallID != "" {
						inputItems = append(inputItems, responses.ResponseInputItemParamOfFunctionCallOutput(
							result.ToolCallID,
							result.Content,
						))
					}
				}
			}
		}

		// Create responses API request with system prompt as Instructions
		systemPrompt := s.buildEnhancedSystemPrompt(processor.Context)
		params := responses.ResponseNewParams{
			Model: responses.ChatModelGPT5,
			Input: responses.ResponseNewParamsInputUnion{
				OfInputItemList: inputItems,
			},
			Tools:        tools,
			Instructions: param.NewOpt(systemPrompt),
		}

		// Check if we have a previous response ID to reference
		if processor.Context.LastResponseID != "" {
			slog.Info("Including previous response ID in request context",
				"previous_response_id", processor.Context.LastResponseID)
			params.PreviousResponseID = param.NewOpt(processor.Context.LastResponseID)
		}

		slog.InfoContext(ctx, "Calling LLM with Responses API", 
			"message_count", len(inputItems),
			"has_response_id", processor.Context.LastResponseID != "",
			"iteration", processor.CurrentRound)

		stream := s.client.Responses.NewStreaming(ctx, params)

		// Clear processor messages after each call since Responses API uses LastResponseID for context
		processor.Messages = []responses.ResponseInputItemUnionParam{}

		var responseContent strings.Builder
		var hasContent bool
		var toolCalls []responses.ResponseFunctionToolCall

		// Process streaming response with event-based processing
		var responseID string
		err := s.processResponsesAPIStreamWithID(stream, responseChan, &responseContent, &hasContent, &toolCalls, &responseID)
		if err != nil {
			return s.handleResponsesAPIError(err)
		}

		// Store response ID for multi-turn conversations
		if responseID != "" {
			slog.Info("Captured response ID for multi-turn conversation", "response_id", responseID)
			// Store this in the processor context for later use when saving the assistant message
			processor.Context.LastResponseID = responseID

			// Update session's last_response_id for future multi-turn context
			if processor.Context.SessionID != "" {
				err := s.sessionRepository.UpdateLastResponseID(ctx, processor.Context.SessionID, responseID)
				if err != nil {
					slog.Error("Failed to update session last_response_id",
						"session_id", processor.Context.SessionID,
						"response_id", responseID,
						"error", err)
					// Don't fail the request if session update fails, just log the error
				} else {
					slog.Info("Successfully updated session last_response_id",
						"session_id", processor.Context.SessionID,
						"response_id", responseID)
				}
			}
		}

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
			// Use the proper function call output format with validated tool call IDs
			processor = s.accumulateAnalysisContext(processor, toolCalls, toolResults, responseContent.String())

			// Continue to next iteration with enhanced context
			continue
		}

		// No tool calls, analysis complete - log comprehensive summary
		slog.Info("Message processing completed successfully with comprehensive call_id summary",
			"implementation", "responses_api",
			"feature_flag", "enabled",
			"total_rounds", processor.CurrentRound,
			"total_tool_calls", processor.GetTotalToolCalls(),
			"final_message_count", len(processor.Messages),
			"processing_mode", "complete")
		break
	}

	return nil
}

// buildConversationContextForResponsesAPI creates conversation context directly in Responses API format
// Following OpenAI Responses API multi-turn pattern: only include new user message when previous response ID is available
func (s *aiService) buildConversationContextForResponsesAPI(msgCtx *MessageContext) []responses.ResponseInputItemUnionParam {
	var inputItems []responses.ResponseInputItemUnionParam

	// With Responses API, we only need to include the current user message
	// Previous conversation context is handled via LastResponseID parameter
	// System prompt is passed separately as Instructions parameter
	
	// Only include conversation history if we don't have a previous response ID
	if msgCtx.LastResponseID == "" && len(msgCtx.ConversationHistory) > 0 {
		slog.Info("No previous response ID available, including recent conversation history for context",
			"conversation_length", len(msgCtx.ConversationHistory))

		// Include only the last few messages to maintain context for first interaction
		recentMessages := msgCtx.ConversationHistory
		if len(recentMessages) > 4 { // Limit to last 4 messages for efficiency
			recentMessages = recentMessages[len(recentMessages)-4:]
		}

		for _, msg := range recentMessages {
			role := responses.EasyInputMessageRoleUser
			if msg.Role == "assistant" {
				role = responses.EasyInputMessageRoleAssistant
			}

			inputItems = append(inputItems, responses.ResponseInputItemParamOfMessage(
				msg.Content,
				role,
			))
		}
	} else if msgCtx.LastResponseID != "" {
		slog.Info("Using previous response ID for conversation context, skipping message history",
			"previous_response_id", msgCtx.LastResponseID,
			"conversation_length", len(msgCtx.ConversationHistory))
	}

	// Add current message
	inputItems = append(inputItems, responses.ResponseInputItemParamOfMessage(
		msgCtx.Message,
		responses.EasyInputMessageRoleUser,
	))

	return inputItems
}

// processResponsesAPIStreamWithID processes the streaming response from Responses API using event-based processing and captures response ID
func (s *aiService) processResponsesAPIStreamWithID(stream *ssestream.Stream[responses.ResponseStreamEventUnion], responseChan chan<- string, responseContent *strings.Builder, hasContent *bool, toolCalls *[]responses.ResponseFunctionToolCall, responseID *string) error {
	defer stream.Close()

	// Initialize tool call state manager for proper accumulation across multiple events
	toolCallState := NewToolCallState()

	// Process streaming events using stream.Next() and stream.Current() pattern
	for stream.Next() {
		event := stream.Current()

		// Log event type for debugging with enhanced context
		slog.Debug("Processing Responses API event",
			"event_type", event.Type,
			"stream_position", "active",
			"tool_calls_tracked", len(toolCallState.toolCalls))

		switch event.Type {
		case "response.output_text.delta":
			// Handle text content deltas
			textEvent := event.AsResponseOutputTextDelta()
			if textEvent.Delta != "" {
				responseContent.WriteString(textEvent.Delta)
				responseChan <- textEvent.Delta
				*hasContent = true
			}

		case "response.output_item.added":
			// Handle output item added events - check for function calls
			if err := s.handleOutputItemAdded(event, toolCallState); err != nil {
				slog.Error("Error processing output item added event", "event_type", event.Type, "error", err)
				// Continue processing other events even if one output item event fails
			}

		case "response.function_call_arguments.delta", "response.function_call.completed", "response.function_call.started":
			// Handle tool call related events using the new parseToolCallsFromEvents method
			if err := s.parseToolCallsFromEvents(event, toolCallState); err != nil {
				slog.Error("Error processing tool call event with call_id context",
					"event_type", event.Type,
					"error", err,
					"active_call_ids", s.getActiveCallIDs(toolCallState),
					"total_tool_calls", len(toolCallState.toolCalls))
				// Continue processing other events even if one tool call event fails
			}

		case "response.completed":
			// Handle completion event - finalize all accumulated tool calls and capture response ID
			completedEvent := event.AsResponseCompleted()
			completedToolCalls := s.GetCompletedToolCalls(toolCallState)

			// Enhanced completion summary logging with all processed call_id values
			allCallIDs := s.getAllCallIDs(toolCallState)
			slog.Info("Response completed, finalizing tool calls with comprehensive summary",
				"tool_call_count", len(completedToolCalls),
				"response_id", completedEvent.Response.ID,
				"all_call_ids", allCallIDs,
				"completed_call_ids", s.getCompletedCallIDs(toolCallState),
				"pending_call_ids", s.getPendingCallIDs(toolCallState))

			// Capture the response ID for multi-turn conversations
			if completedEvent.Response.ID != "" && responseID != nil {
				*responseID = completedEvent.Response.ID
				slog.Info("Captured response ID for multi-turn conversation with call_id context",
					"response_id", completedEvent.Response.ID,
					"associated_call_ids", allCallIDs)
			}

			// Add all results
			*toolCalls = append(*toolCalls, completedToolCalls...)

			// Enhanced individual tool call logging with call_id traceability
			for _, toolCall := range completedToolCalls {
				slog.Info("Finalized tool call with complete traceability",
					"call_id", toolCall.CallID,
					"tool_call_id", toolCall.CallID, // Include both for consistency
					"function_name", toolCall.Name,
					"args_length", len(toolCall.Arguments),
					"arguments", toolCall.Arguments,
					"item_id", toolCall.ID)
			}

			// Log completion summary with all call_id values processed
			slog.Info("Tool call processing pipeline completed successfully",
				"total_processed_call_ids", len(allCallIDs),
				"successful_completions", len(completedToolCalls),
				"call_id_summary", allCallIDs)

			return nil

		case "error":
			// Handle generic error events
			errorEvent := event.AsError()
			return fmt.Errorf("responses API error: %s", errorEvent.Message)

		default:
			// Handle other event types that might be available
			// For now, we'll log them for debugging and continue processing
			slog.Debug("Unhandled event type in Responses API stream", "event_type", event.Type)

			// Try to handle as text delta if it has similar structure
			if strings.Contains(event.Type, "delta") && strings.Contains(event.Type, "text") {
				// Attempt to extract text content from unknown text delta events
				// This is a fallback for potential text events we haven't explicitly handled
				slog.Debug("Attempting to handle unknown text delta event", "event_type", event.Type)
			}
		}
	}

	// Check for stream errors after processing all events
	if err := stream.Err(); err != nil {
		return fmt.Errorf("responses API stream error: %w", err)
	}

	// If we reach here without a completion event, finalize any accumulated tool calls
	completedToolCalls := s.GetCompletedToolCalls(toolCallState)
	slog.Debug("Stream ended without completion event, finalizing tool calls", "tool_call_count", len(completedToolCalls))

	// Add all completed tool calls to the results
	*toolCalls = append(*toolCalls, completedToolCalls...)

	return nil
}

// processResponsesAPIStream processes the streaming response from Responses API using event-based processing (backward compatibility)
func (s *aiService) processResponsesAPIStream(stream *ssestream.Stream[responses.ResponseStreamEventUnion], responseChan chan<- string, responseContent *strings.Builder, hasContent *bool, toolCalls *[]responses.ResponseFunctionToolCall) error {
	var responseID string
	return s.processResponsesAPIStreamWithID(stream, responseChan, responseContent, hasContent, toolCalls, &responseID)
}

// extractFunctionNameFromArguments attempts to extract function name from tool call arguments
// This is a fallback method when function name is not provided in separate events
func (s *aiService) extractFunctionNameFromArguments(arguments string) string {
	// Handle empty or malformed arguments
	if arguments == "" {
		return "get-athlete-profile" // Default to profile for empty args
	}

	// Try to parse the arguments as JSON to extract function name if it's embedded
	var argsMap map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &argsMap); err != nil {
		// If JSON parsing fails, try string-based heuristics
		slog.Debug("Failed to parse tool call arguments as JSON, using heuristics", "arguments", arguments, "error", err)
		return s.inferFunctionNameFromString(arguments)
	}

	// Check if there's a function name field in the arguments
	if funcName, ok := argsMap["function"].(string); ok {
		return funcName
	}

	// Infer function name based on argument structure and field presence
	return s.inferFunctionNameFromFields(argsMap)
}

// inferFunctionNameFromFields infers function name based on the presence of specific fields in arguments
func (s *aiService) inferFunctionNameFromFields(argsMap map[string]interface{}) string {
	// Check for activity_id field (used by get-activity-details and get-activity-streams)
	if _, hasActivityID := argsMap["activity_id"]; hasActivityID {
		// Check for stream-specific fields
		if _, hasStreamTypes := argsMap["stream_types"]; hasStreamTypes {
			return "get-activity-streams"
		}
		if _, hasResolution := argsMap["resolution"]; hasResolution {
			return "get-activity-streams"
		}
		if _, hasProcessingMode := argsMap["processing_mode"]; hasProcessingMode {
			return "get-activity-streams"
		}
		// Default to activity details if only activity_id is present
		return "get-activity-details"
	}

	// Check for per_page field (used by get-recent-activities)
	if _, hasPerPage := argsMap["per_page"]; hasPerPage {
		return "get-recent-activities"
	}

	// Check for content field (used by update-athlete-logbook)
	if _, hasContent := argsMap["content"]; hasContent {
		return "update-athlete-logbook"
	}

	// If no specific fields found, check if arguments are empty (get-athlete-profile)
	if len(argsMap) == 0 {
		return "get-athlete-profile"
	}

	// Default fallback based on most common usage patterns
	slog.Warn("Could not infer function name from arguments, using default", "args_fields", getMapKeys(argsMap))
	return "get-athlete-profile"
}

// inferFunctionNameFromString infers function name from string-based heuristics when JSON parsing fails
func (s *aiService) inferFunctionNameFromString(arguments string) string {
	// Simple heuristic: check if arguments contain fields specific to certain tools
	if strings.Contains(arguments, "activity_id") {
		if strings.Contains(arguments, "stream_types") || strings.Contains(arguments, "resolution") || strings.Contains(arguments, "processing_mode") {
			return "get-activity-streams"
		}
		return "get-activity-details"
	}
	if strings.Contains(arguments, "per_page") {
		return "get-recent-activities"
	}
	if strings.Contains(arguments, "content") {
		return "update-athlete-logbook"
	}
	if arguments == "{}" || strings.TrimSpace(arguments) == "" {
		return "get-athlete-profile"
	}

	// Default fallback
	return "get-athlete-profile"
}

// getMapKeys returns the keys of a map for logging purposes
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ToolCallState manages the state of tool calls during event-based processing
type ToolCallState struct {
	toolCalls    map[string]*responses.ResponseFunctionToolCall // keyed by call_id
	completed    map[string]bool                                // keyed by call_id
	itemToCallID map[string]string                              // maps item_id to call_id for delta event processing
}

// NewToolCallState creates a new tool call state manager
func NewToolCallState() *ToolCallState {
	return &ToolCallState{
		toolCalls:    make(map[string]*responses.ResponseFunctionToolCall),
		completed:    make(map[string]bool),
		itemToCallID: make(map[string]string),
	}
}

// parseToolCallsFromEvents processes Responses API events and manages tool call accumulation
// This method handles event-based tool call processing with proper state management
func (s *aiService) parseToolCallsFromEvents(event responses.ResponseStreamEventUnion, state *ToolCallState) error {
	switch event.Type {
	case "response.function_call_arguments.delta":
		// Handle tool call arguments deltas - accumulate arguments
		funcEvent := event.AsResponseFunctionCallArgumentsDelta()
		return s.handleFunctionCallArgumentsDelta(funcEvent, state)

	case "response.function_call.completed":
		// Handle tool call completion events
		return s.handleFunctionCallCompleted(event, state)

	case "response.function_call.started":
		// Handle tool call start events (if available)
		return s.handleFunctionCallStarted(event, state)

	default:
		// Not a tool call related event, ignore
		fmt.Println("unknown event found", event.Type, event.JSON)
		return nil
	}
}

// handleFunctionCallArgumentsDelta processes function call arguments delta events
func (s *aiService) handleFunctionCallArgumentsDelta(event responses.ResponseFunctionCallArgumentsDeltaEvent, state *ToolCallState) error {
	itemID := event.ItemID

	// Enhanced validation for missing or empty item ID values
	if itemID == "" {
		// Log detailed error information including event structure when extraction fails
		slog.Error("Function call arguments delta extraction failed: empty item ID",
			"event_type", "response.function_call_arguments.delta",
			"delta_content", event.Delta,
			"event_structure", fmt.Sprintf("%+v", event),
			"error", "empty_item_id")
		return fmt.Errorf("function call arguments delta extraction failed: empty item ID")
	}

	// Additional validation for whitespace-only item IDs
	if strings.TrimSpace(itemID) == "" {
		slog.Error("Function call arguments delta extraction failed: item ID contains only whitespace",
			"item_id", fmt.Sprintf("'%s'", itemID),
			"event_type", "response.function_call_arguments.delta",
			"delta_content", event.Delta,
			"error", "whitespace_only_item_id")
		return fmt.Errorf("function call arguments delta extraction failed: item ID contains only whitespace")
	}

	// Look up the correct call_id using the item_id from the delta event
	callID, exists := state.itemToCallID[itemID]
	if !exists {
		// Implement fallback to use item_id as call_id when mapping is not found
		// This can happen if delta events arrive before output_item.added events
		callID = itemID

		// Store the fallback mapping for consistency
		state.itemToCallID[itemID] = callID

		slog.Warn("Function call arguments delta processing using fallback: no call_id mapping found, using item_id as call_id",
			"item_id", itemID,
			"fallback_call_id", callID,
			"event_type", "response.function_call_arguments.delta",
			"delta_content", event.Delta,
			"available_mappings", len(state.itemToCallID),
			"fallback_reason", "missing_call_id_mapping")
	}

	slog.Debug("Processing function call arguments delta with call_id traceability",
		"call_id", callID,
		"item_id", itemID,
		"delta", event.Delta,
		"event_type", "response.function_call_arguments.delta",
		"total_active_calls", len(state.toolCalls))

	// Check if this tool call already exists and update it using the correct call_id
	if toolCall, exists := state.toolCalls[callID]; exists {
		toolCall.Arguments += event.Delta
		slog.Debug("Accumulated function call arguments delta with call_id traceability",
			"call_id", callID,
			"item_id", itemID,
			"delta_length", len(event.Delta),
			"total_args_length", len(toolCall.Arguments),
			"function_name", toolCall.Name)
	} else {
		// This should be rare since output_item.added should come first, but handle gracefully
		slog.Warn("Creating new tool call from arguments delta - output_item.added event may have been missed",
			"call_id", callID,
			"item_id", itemID,
			"initial_delta", event.Delta)

		toolCall := &responses.ResponseFunctionToolCall{
			ID:        itemID,
			CallID:    callID,
			Name:      "", // Name will be set when available or inferred later
			Arguments: event.Delta,
		}
		state.toolCalls[callID] = toolCall

		slog.Info("Created new function call from arguments delta using correct call_id",
			"call_id", callID,
			"item_id", itemID,
			"initial_delta", event.Delta)
	}

	return nil
}

// handleFunctionCallCompleted processes function call completion events
func (s *aiService) handleFunctionCallCompleted(event responses.ResponseStreamEventUnion, state *ToolCallState) error {
	// Try to extract tool call ID from completion event
	// Note: The exact structure may vary based on the actual API response format
	slog.Debug("Function call completion event received", "event_type", event.Type)

	// Mark all current tool calls as completed if we can't identify specific ones
	for toolCallID := range state.toolCalls {
		if !state.completed[toolCallID] {
			state.completed[toolCallID] = true
			slog.Debug("Marked tool call as completed", "tool_call_id", toolCallID)
		}
	}

	return nil
}

// handleFunctionCallStarted processes function call start events
func (s *aiService) handleFunctionCallStarted(event responses.ResponseStreamEventUnion, state *ToolCallState) error {
	// Try to extract function name and ID from start event
	// Note: The exact structure may vary based on the actual API response format
	slog.Debug("Function call start event received", "event_type", event.Type)

	fmt.Println("function call started")
	fmt.Println(event.JSON)

	// This would need to be implemented based on the actual event structure
	// For now, we'll rely on the arguments delta events to create tool calls

	return nil
}

// handleOutputItemAdded processes response.output_item.added events to extract call_id from function call items
func (s *aiService) handleOutputItemAdded(event responses.ResponseStreamEventUnion, state *ToolCallState) error {
	// Use event.AsResponseOutputItemAdded() to get the typed event structure
	outputItemEvent := event.AsResponseOutputItemAdded()

	slog.Info("Processing output item added event",
		"event_type", event.Type,
		"item_id", outputItemEvent.Item.ID,
		"item_type", outputItemEvent.Item.Type)

	// Check if the item type is function_call before processing
	if outputItemEvent.Item.Type != "function_call" {
		slog.Debug("Output item is not a function call, skipping",
			"item_type", outputItemEvent.Item.Type,
			"item_id", outputItemEvent.Item.ID)
		return nil
	}

	// Enhanced validation and error handling for call_id extraction
	callID := outputItemEvent.Item.CallID
	var fallbackUsed bool
	var fallbackReason string

	// Validate that call_id is not missing or empty
	if callID == "" {
		slog.Warn("Function call item missing call_id, detailed event structure",
			"item_id", outputItemEvent.Item.ID,
			"item_type", outputItemEvent.Item.Type,
			"function_name", outputItemEvent.Item.Name,
			"event_structure", fmt.Sprintf("%+v", outputItemEvent.Item))

		// Implement fallback to event.ItemID when call_id is unavailable
		fallbackID := outputItemEvent.Item.ID
		if fallbackID == "" {
			// Log detailed error information including event structure when extraction fails
			slog.Error("Call_id extraction failed: both call_id and item_id are empty",
				"event_type", event.Type,
				"item_structure", fmt.Sprintf("%+v", outputItemEvent.Item),
				"full_event", fmt.Sprintf("%+v", outputItemEvent),
				"error", "missing_identifiers")
			return fmt.Errorf("call_id extraction failed: both call_id and item_id are empty in function call item")
		}

		callID = fallbackID
		fallbackUsed = true
		fallbackReason = "missing_call_id"

		// Log fallback strategy usage with reasons
		slog.Info("Using fallback identification strategy",
			"fallback_call_id", callID,
			"original_item_id", outputItemEvent.Item.ID,
			"fallback_reason", fallbackReason,
			"strategy", "item_id_fallback")
	}

	// Additional validation for empty call_id after potential fallback
	if strings.TrimSpace(callID) == "" {
		slog.Error("Call_id extraction failed: call_id is empty after validation",
			"original_call_id", outputItemEvent.Item.CallID,
			"fallback_id", outputItemEvent.Item.ID,
			"event_structure", fmt.Sprintf("%+v", outputItemEvent.Item),
			"error", "empty_call_id")
		return fmt.Errorf("call_id extraction failed: call_id is empty after validation")
	}

	// Log successful extraction with fallback information
	if fallbackUsed {
		slog.Warn("Call_id extracted using fallback strategy",
			"extracted_call_id", callID,
			"fallback_reason", fallbackReason,
			"item_id", outputItemEvent.Item.ID,
			"function_name", outputItemEvent.Item.Name)
	} else {
		slog.Info("Successfully extracted call_id from function call item",
			"call_id", callID,
			"item_id", outputItemEvent.Item.ID,
			"function_name", outputItemEvent.Item.Name)
	}

	// Check if this tool call already exists in state and update it
	if existingToolCall, exists := state.toolCalls[callID]; exists {
		// Update existing tool call with additional information
		if existingToolCall.Name == "" && outputItemEvent.Item.Name != "" {
			existingToolCall.Name = outputItemEvent.Item.Name
			slog.Info("Updated existing tool call with function name",
				"call_id", callID,
				"function_name", outputItemEvent.Item.Name,
				"fallback_used", fallbackUsed)
		}
	} else {
		// Create new tool call entry with the extracted call_id
		toolCall := &responses.ResponseFunctionToolCall{
			ID:        outputItemEvent.Item.ID,   // Keep item ID for compatibility
			CallID:    callID,                    // Use extracted call_id as primary key
			Name:      outputItemEvent.Item.Name, // Function name from the event
			Arguments: "",                        // Arguments will be accumulated from delta events
		}

		// Add to tool call state using call_id as the primary key
		state.toolCalls[callID] = toolCall

		// Store mapping from item_id to call_id for delta event processing
		state.itemToCallID[outputItemEvent.Item.ID] = callID

		slog.Info("Created new tool call from output item",
			"call_id", callID,
			"item_id", outputItemEvent.Item.ID,
			"function_name", outputItemEvent.Item.Name,
			"fallback_used", fallbackUsed,
			"fallback_reason", fallbackReason)
	}

	return nil
}

// GetCompletedToolCalls returns all completed tool calls with proper function names
func (s *aiService) GetCompletedToolCalls(state *ToolCallState) []responses.ResponseFunctionToolCall {
	var completedCalls []responses.ResponseFunctionToolCall

	slog.Info("Processing completed tool calls", "total_tool_calls", len(state.toolCalls))

	for toolCallID, toolCall := range state.toolCalls {
		fmt.Println("---------------------------")
		fmt.Println(toolCall)
		slog.Debug("Processing tool call",
			"tool_call_id", toolCallID,
			"current_name", toolCall.Name,
			"args_length", len(toolCall.Arguments))

		// Ensure function name is set
		if toolCall.Name == "" {
			inferredName := s.extractFunctionNameFromArguments(toolCall.Arguments)
			toolCall.Name = inferredName
			slog.Info("Inferred function name for tool call",
				"tool_call_id", toolCallID,
				"inferred_name", inferredName)
		}

		// Validate and sanitize the tool call
		if s.validateToolCall(toolCall) {
			// Create a clean copy of the tool call
			cleanedToolCall := *toolCall
			cleanedToolCall.CallID = toolCallID
			completedCalls = append(completedCalls, cleanedToolCall)
			slog.Info("Added completed tool call",
				"tool_call_id", toolCallID,
				"function_name", cleanedToolCall.Name,
				"args_length", len(cleanedToolCall.Arguments))
		} else {
			slog.Warn("Skipping invalid tool call",
				"tool_call_id", toolCallID,
				"function_name", toolCall.Name,
				"args_length", len(toolCall.Arguments))
		}
	}

	slog.Info("Completed tool call processing", "valid_tool_calls", len(completedCalls))
	return completedCalls
}

// validateToolCall validates that a tool call has all required fields and is properly formed
func (s *aiService) validateToolCall(toolCall *responses.ResponseFunctionToolCall) bool {
	// Check required fields - validate CallID as primary key
	if toolCall.CallID == "" {
		slog.Warn("Tool call missing call_id", "tool_call_id", toolCall.CallID)
		return false
	}

	if toolCall.Name == "" {
		slog.Warn("Tool call missing function name", "tool_call_id", toolCall.CallID)
		return false
	}

	// Validate function name is one of our known tools
	knownTools := map[string]bool{
		"get-athlete-profile":    true,
		"get-recent-activities":  true,
		"get-activity-details":   true,
		"get-activity-streams":   true,
		"update-athlete-logbook": true,
	}

	if !knownTools[toolCall.Name] {
		slog.Warn("Unknown function name in tool call", "tool_call_id", toolCall.CallID, "function_name", toolCall.Name)
		return false
	}

	// Validate arguments are valid JSON
	if toolCall.Arguments != "" {
		var argsMap map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.Arguments), &argsMap); err != nil {
			slog.Warn("Tool call has invalid JSON arguments", "tool_call_id", toolCall.CallID, "function_name", toolCall.Name, "error", err)
			return false
		}
	}

	return true
}

// sanitizeToolCall cleans and normalizes a tool call
func (s *aiService) sanitizeToolCall(toolCall responses.ResponseFunctionToolCall) responses.ResponseFunctionToolCall {
	// Create a clean copy
	cleaned := responses.ResponseFunctionToolCall{
		ID:        strings.TrimSpace(toolCall.ID),
		CallID:    strings.TrimSpace(toolCall.CallID),
		Name:      strings.TrimSpace(toolCall.Name),
		Arguments: strings.TrimSpace(toolCall.Arguments),
	}

	// Ensure arguments is valid JSON, default to empty object if not
	if cleaned.Arguments == "" {
		cleaned.Arguments = "{}"
	} else {
		// Validate and potentially fix JSON
		var argsMap map[string]interface{}
		if err := json.Unmarshal([]byte(cleaned.Arguments), &argsMap); err != nil {
			slog.Warn("Sanitizing invalid JSON arguments", "tool_call_id", cleaned.CallID, "original_args", cleaned.Arguments)
			cleaned.Arguments = "{}"
		}
	}

	return cleaned
}

// IsToolCallComplete checks if a specific tool call is marked as completed
func (s *aiService) IsToolCallComplete(state *ToolCallState, toolCallID string) bool {
	return state.completed[toolCallID]
}

// GetToolCallCount returns the total number of tool calls being tracked
func (s *aiService) GetToolCallCount(state *ToolCallState) int {
	return len(state.toolCalls)
}

// GetCompletedToolCallCount returns the number of completed tool calls
func (s *aiService) GetCompletedToolCallCount(state *ToolCallState) int {
	count := 0
	for _, completed := range state.completed {
		if completed {
			count++
		}
	}
	return count
}

// MarkToolCallCompleted explicitly marks a tool call as completed
func (s *aiService) MarkToolCallCompleted(state *ToolCallState, toolCallID string) {
	if _, exists := state.toolCalls[toolCallID]; exists {
		state.completed[toolCallID] = true
		slog.Debug("Marked tool call as completed", "tool_call_id", toolCallID)
	} else {
		slog.Warn("Attempted to mark non-existent tool call as completed", "tool_call_id", toolCallID)
	}
}

// GetToolCallByID retrieves a specific tool call by ID
func (s *aiService) GetToolCallByID(state *ToolCallState, toolCallID string) (*responses.ResponseFunctionToolCall, bool) {
	toolCall, exists := state.toolCalls[toolCallID]
	return toolCall, exists
}

// HasPendingToolCalls checks if there are any tool calls that haven't been completed
func (s *aiService) HasPendingToolCalls(state *ToolCallState) bool {
	for toolCallID := range state.toolCalls {
		if !state.completed[toolCallID] {
			return true
		}
	}
	return false
}

// getActiveCallIDs returns a list of all active call_ids for logging purposes
func (s *aiService) getActiveCallIDs(state *ToolCallState) []string {
	var callIDs []string
	for callID := range state.toolCalls {
		callIDs = append(callIDs, callID)
	}
	return callIDs
}

// getAllCallIDs returns all call_ids that have been processed
func (s *aiService) getAllCallIDs(state *ToolCallState) []string {
	var callIDs []string
	for callID := range state.toolCalls {
		callIDs = append(callIDs, callID)
	}
	return callIDs
}

// getCompletedCallIDs returns only the completed call_ids
func (s *aiService) getCompletedCallIDs(state *ToolCallState) []string {
	var callIDs []string
	for callID, completed := range state.completed {
		if completed {
			callIDs = append(callIDs, callID)
		}
	}
	return callIDs
}

// getPendingCallIDs returns only the pending call_ids
func (s *aiService) getPendingCallIDs(state *ToolCallState) []string {
	var callIDs []string
	for callID := range state.toolCalls {
		if !state.completed[callID] {
			callIDs = append(callIDs, callID)
		}
	}
	return callIDs
}

// getContentPreview returns a truncated preview of content for logging purposes
func (s *aiService) getContentPreview(content string) string {
	const maxPreviewLength = 100
	if len(content) <= maxPreviewLength {
		return content
	}
	return content[:maxPreviewLength] + "..."
}

// parseToolCallsFromResponsesAPIEvents parses tool calls from Responses API events (legacy method for compatibility)
// Note: This method is now primarily used as a fallback. The main tool call processing
// is handled by parseToolCallsFromEvents for better event-based processing.
func (s *aiService) parseToolCallsFromResponsesAPIEvents(event responses.ResponseFunctionCallArgumentsDeltaEvent, existingToolCalls []responses.ResponseFunctionToolCall) []responses.ResponseFunctionToolCall {
	// Check if this tool call already exists and update it
	for i, existing := range existingToolCalls {
		if existing.ID == event.ItemID {
			existingToolCalls[i].Arguments += event.Delta
			return existingToolCalls
		}
	}

	// New tool call - create with accumulated delta
	toolCall := responses.ResponseFunctionToolCall{
		ID:        event.ItemID,
		Name:      "", // Will need to be set from function call start event
		Arguments: event.Delta,
	}

	return append(existingToolCalls, toolCall)
}

// getContextualProgressMessageForResponsesAPI returns progress messages based on specific tool calls for Responses API
func (s *aiService) getContextualProgressMessageForResponsesAPI(processor *IterativeProcessor, toolCalls []responses.ResponseFunctionToolCall) string {
	// Analyze the combination of tool calls to provide contextual messages
	hasProfile := false
	hasActivities := false
	hasDetails := false
	hasStreams := false
	hasLogbookUpdate := false

	for _, toolCall := range toolCalls {
		switch toolCall.Name {
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

// executeToolsFromResponsesAPI executes the tool calls from Responses API and returns the results
func (s *aiService) executeToolsFromResponsesAPI(ctx context.Context, msgCtx *MessageContext, toolCalls []responses.ResponseFunctionToolCall) ([]ToolResult, error) {
	var results []ToolResult

	// Enhanced logging with all call_id values for traceability
	allCallIDs := make([]string, len(toolCalls))
	for i, toolCall := range toolCalls {
		allCallIDs[i] = toolCall.CallID
	}

	slog.Info("Executing tool calls from Responses API with call_id traceability",
		"tool_call_count", len(toolCalls),
		"all_call_ids", allCallIDs,
		"implementation", "responses_api")

	for i, toolCall := range toolCalls {
		slog.Info("Executing individual tool call with call_id context",
			"index", i,
			"call_id", toolCall.CallID,
			"tool_call_id", toolCall.CallID, // Include both for consistency
			"function_name", toolCall.Name,
			"arguments", toolCall.Arguments,
			"item_id", toolCall.ID)

		result := ToolResult{
			ToolCallID: toolCall.CallID,
		}

		slog.Info("Creating ToolResult with extracted call_id",
			"tool_call_id", toolCall.CallID,
			"function_name", toolCall.Name,
			"item_id", toolCall.ID)

		switch toolCall.Name {
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
			if err := json.Unmarshal([]byte(toolCall.Arguments), &args); err != nil {
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
			if err := json.Unmarshal([]byte(toolCall.Arguments), &args); err != nil {
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
			if err := json.Unmarshal([]byte(toolCall.Arguments), &args); err != nil {
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
			if err := json.Unmarshal([]byte(toolCall.Arguments), &args); err != nil {
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
			result.Content = fmt.Sprintf("Unknown tool: %s", toolCall.Name)
		}

		slog.Info("Completed tool execution with call_id traceability",
			"call_id", result.ToolCallID,
			"tool_call_id", result.ToolCallID, // Include both for consistency
			"function_name", toolCall.Name,
			"has_error", result.Error != "",
			"content_length", len(result.Content),
			"execution_index", i)

		results = append(results, result)
	}

	// Enhanced completion summary with all call_id values processed
	completedCallIDs := make([]string, len(results))
	successfulExecutions := 0
	for i, result := range results {
		completedCallIDs[i] = result.ToolCallID
		if result.Error == "" {
			successfulExecutions++
		}
	}

	slog.Info("All tool executions completed with comprehensive call_id summary",
		"total_results", len(results),
		"successful_executions", successfulExecutions,
		"failed_executions", len(results)-successfulExecutions,
		"completed_call_ids", completedCallIDs,
		"implementation", "responses_api")

	return results, nil
}

// accumulateAnalysisContext builds enhanced context with accumulated insights
func (s *aiService) accumulateAnalysisContext(processor *IterativeProcessor, toolCalls []responses.ResponseFunctionToolCall, toolResults []ToolResult, responseContent string) *IterativeProcessor {
	// Add tool results to processor for tracking
	processor.AddToolResults(toolResults)

	// With Responses API and LastResponseID, we don't need to accumulate messages
	// The conversation context is maintained via the response ID
	// Only add tool results for the next iteration
	
	// Fix tool result IDs to match the current tool calls
	fixedToolResults := toolResults

	slog.Info("Processing tool call results for next iteration",
		"expected_tool_calls", len(toolCalls),
		"original_tool_results", len(toolResults),
		"fixed_tool_results", len(fixedToolResults),
		"using_response_id", processor.Context.LastResponseID != "")

	// Prepare tool results for the next API call
	// These will be included in the next request as tool call outputs
	successfullyProcessed := 0
	for _, result := range fixedToolResults {
		// Validate tool call ID
		if result.ToolCallID == "" {
			slog.Warn("Skipping tool result with empty call_id",
				"content_length", len(result.Content),
				"has_error", result.Error != "",
				"result_content_preview", s.getContentPreview(result.Content))
			continue
		}

		slog.Info("Processed tool call result for next iteration",
			"call_id", result.ToolCallID,
			"content_length", len(result.Content),
			"has_error", result.Error != "",
			"result_index", successfullyProcessed)

		// With Responses API, tool results will be included in the next request
		// rather than accumulated in processor.Messages
		successfullyProcessed++
	}

	// Enhanced completion summary
	processedCallIDs := make([]string, 0, successfullyProcessed)
	for _, result := range fixedToolResults {
		if result.ToolCallID != "" {
			processedCallIDs = append(processedCallIDs, result.ToolCallID)
		}
	}

	slog.Info("Successfully processed tool call results for next iteration",
		"processed_count", successfullyProcessed,
		"total_results", len(fixedToolResults),
		"processed_call_ids", processedCallIDs,
		"skipped_results", len(fixedToolResults)-successfullyProcessed)

	return processor
}

// fixToolResultIDs ensures tool results have the correct tool call IDs that match the current conversation
func (s *aiService) fixToolResultIDs(toolCalls []responses.ResponseFunctionToolCall, toolResults []ToolResult) []ToolResult {
	// Create a mapping from function name + arguments to tool call ID
	// This allows us to match tool results to the correct tool call IDs
	toolCallMap := make(map[string]string) // key: function_name, value: tool_call_id

	for _, toolCall := range toolCalls {
		if toolCall.Name != "" && toolCall.CallID != "" {
			// Use function name as the key for matching
			// In most cases, there's only one call per function per round
			toolCallMap[toolCall.Name] = toolCall.CallID

			slog.Debug("Mapped tool call",
				"function_name", toolCall.Name,
				"tool_call_id", toolCall.CallID)
		}
	}

	// Fix the tool result IDs
	var fixedResults []ToolResult
	for i, result := range toolResults {
		// Try to determine which tool call this result corresponds to
		// We need to infer the function name from the tool execution
		functionName := s.inferFunctionNameFromToolResult(result, i, toolCalls)

		if correctID, exists := toolCallMap[functionName]; exists {
			// Create a new result with the correct tool call ID
			fixedResult := ToolResult{
				ToolCallID: correctID,
				Content:    result.Content,
				Error:      result.Error,
				Data:       result.Data,
			}
			fixedResults = append(fixedResults, fixedResult)

			slog.Info("Fixed tool result call_id for proper correlation",
				"original_call_id", result.ToolCallID,
				"fixed_call_id", correctID,
				"function_name", functionName,
				"correlation_method", "function_name_mapping")
		} else {
			slog.Warn("Could not find matching tool call for result - call_id correlation failed",
				"original_call_id", result.ToolCallID,
				"inferred_function", functionName,
				"available_functions", getMapKeysString(toolCallMap),
				"correlation_issue", "no_matching_call_id")

			// Keep the original result if we can't fix it
			fixedResults = append(fixedResults, result)
		}
	}

	return fixedResults
}

// inferFunctionNameFromToolResult tries to determine which function a tool result corresponds to
func (s *aiService) inferFunctionNameFromToolResult(result ToolResult, index int, toolCalls []responses.ResponseFunctionToolCall) string {
	// If we have the same number of results as tool calls, match by index
	if len(toolCalls) > index {
		return toolCalls[index].Name
	}

	// Try to infer from the content or error message
	content := strings.ToLower(result.Content)
	if strings.Contains(content, "athlete profile") || strings.Contains(content, "profile") {
		return "get-athlete-profile"
	}
	if strings.Contains(content, "recent activities") || strings.Contains(content, "activities") {
		return "get-recent-activities"
	}
	if strings.Contains(content, "activity details") || strings.Contains(content, "activity") {
		return "get-activity-details"
	}
	if strings.Contains(content, "activity streams") || strings.Contains(content, "streams") {
		return "get-activity-streams"
	}
	if strings.Contains(content, "logbook updated") || strings.Contains(content, "logbook") {
		return "update-athlete-logbook"
	}

	// Default fallback - use the first tool call if available
	if len(toolCalls) > 0 {
		return toolCalls[0].Name
	}

	return "unknown"
}

// getMapKeysString returns the keys of a string map as a slice for logging
func getMapKeysString(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// accumulateAnalysisContextSafely builds enhanced context with accumulated insights while avoiding tool call ID mismatches
func (s *aiService) accumulateAnalysisContextSafely(processor *IterativeProcessor, toolCalls []responses.ResponseFunctionToolCall, toolResults []ToolResult, responseContent string) *IterativeProcessor {
	// Add tool results to processor
	processor.AddToolResults(toolResults)

	// Add assistant message with response content to conversation
	if responseContent != "" {
		processor.Messages = append(processor.Messages, responses.ResponseInputItemParamOfMessage(
			responseContent,
			responses.EasyInputMessageRoleAssistant,
		))
	}

	// Instead of adding function call outputs (which can cause ID mismatches),
	// add the tool results as regular user messages with clear context
	for _, result := range toolResults {
		if result.ToolCallID == "" {
			slog.Warn("Skipping tool result with empty call_id - safe context traceability issue",
				"content_length", len(result.Content),
				"has_error", result.Error != "",
				"result_content_preview", s.getContentPreview(result.Content),
				"error_type", "missing_call_id_safe_context")
			continue
		}

		slog.Info("Adding tool result as user message with call_id traceability",
			"call_id", result.ToolCallID,
			"tool_call_id", result.ToolCallID, // Include both for consistency
			"content_length", len(result.Content),
			"has_error", result.Error != "")

		// Format the tool result as a clear user message
		var toolResultMessage string
		if result.Error != "" {
			toolResultMessage = fmt.Sprintf("Tool execution error: %s", result.Error)
		} else {
			toolResultMessage = fmt.Sprintf("Tool execution result: %s", result.Content)
		}

		processor.Messages = append(processor.Messages, responses.ResponseInputItemParamOfMessage(
			toolResultMessage,
			responses.EasyInputMessageRoleUser,
		))
	}

	// Enhanced completion summary with call_id traceability
	processedCallIDs := make([]string, 0, len(toolResults))
	for _, result := range toolResults {
		if result.ToolCallID != "" {
			processedCallIDs = append(processedCallIDs, result.ToolCallID)
		}
	}

	slog.Info("Successfully accumulated analysis context safely with call_id summary",
		"total_messages", len(processor.Messages),
		"tool_results", len(toolResults),
		"processed_call_ids", processedCallIDs,
		"context_method", "safe_user_messages")

	return processor
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
- Keep the logbook as small as possible, only key insights and any current training plan should be part of it. Detailed analysis of any individual workout should not be part of it.

COACHING APPROACH:
- Use the logbook context to provide personalized coaching based on the athlete's complete history
- Structure the logbook content however you think will be most effective for coaching
- When athlete asks for an analysis for any of their workout ask them what they want next from you. Give them a workout or training plan only if they ask for it.
- Try not to simply repeat the data you get from the tools, rather try to present insights.

RESPONSE FORMAT:
- Your response will be rendered as markdown, so use headings, bold, italics, tables etc when appropriate.
- Mermaid and vega lite is also supported for graphics rendering. Provide simple graphics or diagrams when appropriate.

CRITICAL INSTRUCTION
- Whenever you provide analysis or information of a specific activity from strava, include a link back to the original activity in strava in markdown format.

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
func (s *aiService) shouldContinueAnalysis(processor *IterativeProcessor, toolCalls []responses.ResponseFunctionToolCall, hasContent bool) (bool, string) {
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
func (s *aiService) getCoachingProgressMessage(processor *IterativeProcessor, toolCalls []responses.ResponseFunctionToolCall) string {
	// Determine message based on tool calls and current context
	if len(toolCalls) > 0 {
		return s.getContextualProgressMessageForResponsesAPI(processor, toolCalls)
	}

	// Fallback to round-based messages with coaching tone
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
func (s *aiService) executeToolsWithRecovery(ctx context.Context, msgCtx *MessageContext, toolCalls []responses.ResponseFunctionToolCall) ([]ToolResult, error) {
	results, err := s.executeToolsFromResponsesAPI(ctx, msgCtx, toolCalls)
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
	case strings.Contains(errStr, "network is unreachable"):
		return ErrOpenAIUnavailable
	case strings.Contains(errStr, "no such host"):
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

// handleResponsesAPIError converts official OpenAI SDK errors to our custom error types
func (s *aiService) handleResponsesAPIError(err error) error {
	if err == nil {
		return nil
	}

	log.Printf("Official OpenAI SDK error: %v", err)

	// Check for official SDK error type using type assertion
	var apiErr *openai.Error
	if errors.As(err, &apiErr) && apiErr != nil {
		// Handle official SDK Error based on status code and error details
		switch apiErr.StatusCode {
		case 400:
			// Bad Request errors
			switch {
			case strings.Contains(apiErr.Message, "context_length_exceeded"):
				return ErrContextTooLong
			case strings.Contains(apiErr.Message, "invalid_request"):
				return ErrInvalidInput
			case strings.Contains(apiErr.Message, "model_not_found"):
				return ErrInvalidInput
			case strings.Contains(apiErr.Type, "invalid_request_error"):
				return ErrInvalidInput
			default:
				return fmt.Errorf("invalid request: %s", apiErr.Message)
			}
		case 401:
			// Unauthorized errors
			return fmt.Errorf("authentication failed: %s", apiErr.Message)
		case 403:
			// Forbidden errors
			return fmt.Errorf("permission denied: %s", apiErr.Message)
		case 404:
			// Not Found errors
			return fmt.Errorf("resource not found: %s", apiErr.Message)
		case 429:
			// Rate Limit errors
			return ErrOpenAIRateLimit
		case 500, 502, 503, 504:
			// Server errors
			return ErrOpenAIUnavailable
		default:
			// Other HTTP status codes
			return fmt.Errorf("OpenAI API error (status %d): %s", apiErr.StatusCode, apiErr.Message)
		}
	}

	// Check for string-based error patterns as fallback for non-API errors
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "rate limit"):
		return ErrOpenAIRateLimit
	case strings.Contains(errStr, "quota"):
		return ErrOpenAIQuotaExceeded
	case strings.Contains(errStr, "timeout"):
		return ErrOpenAIUnavailable
	case strings.Contains(errStr, "connection"):
		return ErrOpenAIUnavailable
	case strings.Contains(errStr, "network is unreachable"):
		return ErrOpenAIUnavailable
	case strings.Contains(errStr, "no such host"):
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
		return fmt.Errorf("OpenAI Responses API error: %w", err)
	}
}

// categorizeToolExecutionError provides enhanced error categorization for tool execution failures
func (s *aiService) categorizeToolExecutionError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	log.Printf("Categorizing tool execution error: %v", err)

	// Check for specific tool execution error patterns
	switch {
	case strings.Contains(errStr, "strava"):
		// Strava-related errors
		if strings.Contains(errStr, "rate limit") {
			return fmt.Errorf("Strava API rate limit exceeded. Please try again in a few minutes")
		}
		if strings.Contains(errStr, "token") || strings.Contains(errStr, "auth") {
			return fmt.Errorf("Strava authentication issue. Please reconnect your Strava account")
		}
		if strings.Contains(errStr, "not found") {
			return fmt.Errorf("Requested Strava data not found or not accessible")
		}
		return fmt.Errorf("Strava service error: %w", err)
	case strings.Contains(errStr, "timeout"):
		return fmt.Errorf("Request timed out while accessing training data")
	case strings.Contains(errStr, "network") || strings.Contains(errStr, "connection"):
		return fmt.Errorf("Network connectivity issue while accessing training data")
	case strings.Contains(errStr, "json") || strings.Contains(errStr, "parse"):
		return fmt.Errorf("Data format error while processing training information")
	case strings.Contains(errStr, "context") || strings.Contains(errStr, "length"):
		return fmt.Errorf("Training data too large to process in current context")
	default:
		return fmt.Errorf("Tool execution error: %w", err)
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
