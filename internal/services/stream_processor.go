package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"bodda/internal/config"
)

// Custom error types for stream processing
var (
	ErrStreamTooLarge         = errors.New("stream data exceeds context window limits")
	ErrProcessingFailed       = errors.New("stream processing failed")
	ErrInvalidProcessingMode  = errors.New("invalid processing mode specified")
	ErrPageNotFound          = errors.New("requested page not found")
	ErrStreamExpired         = errors.New("paginated stream has expired")
	ErrStravaAPIFailure      = errors.New("strava API request failed")
	ErrDerivedFeaturesFailure = errors.New("derived features extraction failed")
	ErrAISummaryFailure      = errors.New("AI summarization failed")
	ErrPaginationFailure     = errors.New("pagination processing failed")
	ErrContextExceeded       = errors.New("context window exceeded")
	ErrInvalidRequest        = errors.New("invalid request parameters")
	ErrDataCorrupted         = errors.New("stream data is corrupted or incomplete")
	ErrProcessorUnavailable  = errors.New("required processor is unavailable")
)

// StreamProcessingError provides detailed error information for stream processing failures
type StreamProcessingError struct {
	Type            string                 `json:"type"`
	Message         string                 `json:"message"`
	ActivityID      int64                  `json:"activity_id,omitempty"`
	ProcessingMode  string                 `json:"processing_mode,omitempty"`
	DataSize        int                    `json:"data_size,omitempty"`
	AvailableTokens int                    `json:"available_tokens,omitempty"`
	Alternatives    []ProcessingOption     `json:"alternatives,omitempty"`
	Context         map[string]interface{} `json:"context,omitempty"`
	OriginalError   error                  `json:"-"`
}

// Error implements the error interface
func (e *StreamProcessingError) Error() string {
	return e.Message
}

// Unwrap returns the original error for error unwrapping
func (e *StreamProcessingError) Unwrap() error {
	return e.OriginalError
}

// NewStreamProcessingError creates a new stream processing error with context
func NewStreamProcessingError(errorType string, message string, activityID int64, processingMode string) *StreamProcessingError {
	return &StreamProcessingError{
		Type:           errorType,
		Message:        message,
		ActivityID:     activityID,
		ProcessingMode: processingMode,
		Context:        make(map[string]interface{}),
	}
}

// WithDataSize adds data size information to the error
func (e *StreamProcessingError) WithDataSize(dataSize int) *StreamProcessingError {
	e.DataSize = dataSize
	return e
}

// WithAvailableTokens adds available tokens information to the error
func (e *StreamProcessingError) WithAvailableTokens(tokens int) *StreamProcessingError {
	e.AvailableTokens = tokens
	return e
}

// WithAlternatives adds alternative processing methods to the error
func (e *StreamProcessingError) WithAlternatives(alternatives []ProcessingOption) *StreamProcessingError {
	e.Alternatives = alternatives
	return e
}

// WithContext adds additional context information to the error
func (e *StreamProcessingError) WithContext(key string, value interface{}) *StreamProcessingError {
	e.Context[key] = value
	return e
}

// WithOriginalError adds the original error for wrapping
func (e *StreamProcessingError) WithOriginalError(err error) *StreamProcessingError {
	e.OriginalError = err
	return e
}

// ProcessingOption represents a processing mode option for the LLM
type ProcessingOption struct {
	Mode        string `json:"mode"`
	Description string `json:"description"`
	Command     string `json:"command"`
}

// ProcessedStreamResult represents the result of stream processing
type ProcessedStreamResult struct {
	ToolCallID      string                 `json:"tool_call_id"`
	Content         string                 `json:"content"` // Human-readable formatted text
	ProcessingMode  string                 `json:"processing_mode,omitempty"`
	Options         []ProcessingOption     `json:"options,omitempty"`
	Data            interface{}            `json:"data,omitempty"` // Raw data for internal use
}

// StreamConfig holds configuration for stream processing
type StreamConfig struct {
	MaxContextTokens    int     `json:"max_context_tokens"`
	TokenPerCharRatio   float64 `json:"token_per_char_ratio"`
	DefaultPageSize     int     `json:"default_page_size"`
	MaxPageSize         int     `json:"max_page_size"`
	RedactionEnabled    bool    `json:"redaction_enabled"`
	StravaResolutions   []string `json:"strava_resolutions"`
}

// StreamProcessor interface defines the core stream processing functionality
type StreamProcessor interface {
	ProcessStreamOutput(data *StravaStreams, toolCallID string) (*ProcessedStreamResult, error)
	ProcessStreamOutputWithContext(data *StravaStreams, toolCallID string, currentContextTokens int) (*ProcessedStreamResult, error)
	ShouldProcess(data *StravaStreams) bool
	GetProcessingOptions() []ProcessingOption
	EstimateTokens(data *StravaStreams) int
}

// streamProcessor implements the StreamProcessor interface
type streamProcessor struct {
	config *StreamConfig
}

// NewStreamProcessor creates a new stream processor with configuration
func NewStreamProcessor(cfg *config.Config) StreamProcessor {
	// Use configuration from config package
	streamConfig := &StreamConfig{
		MaxContextTokens:  cfg.StreamProcessing.MaxContextTokens,
		TokenPerCharRatio: cfg.StreamProcessing.TokenPerCharRatio,
		DefaultPageSize:   cfg.StreamProcessing.DefaultPageSize,
		MaxPageSize:       cfg.StreamProcessing.MaxPageSize,
		RedactionEnabled:  cfg.StreamProcessing.RedactionEnabled,
		StravaResolutions: []string{"low", "medium", "high"},
	}

	return &streamProcessor{
		config: streamConfig,
	}
}

// ProcessStreamOutput processes stream data and returns appropriate result
func (sp *streamProcessor) ProcessStreamOutput(data *StravaStreams, toolCallID string) (*ProcessedStreamResult, error) {
	return sp.ProcessStreamOutputWithContext(data, toolCallID, 0)
}

// ProcessStreamOutputWithContext processes stream data with current context information
func (sp *streamProcessor) ProcessStreamOutputWithContext(data *StravaStreams, toolCallID string, currentContextTokens int) (*ProcessedStreamResult, error) {
	if data == nil {
		return nil, fmt.Errorf("stream data is nil")
	}

	// Check if processing is needed
	if !sp.ShouldProcess(data) {
		// Return raw data formatted as human-readable text
		content := sp.formatRawStreamData(data)
		return &ProcessedStreamResult{
			ToolCallID:     toolCallID,
			Content:        content,
			ProcessingMode: "raw",
		}, nil
	}

	// Stream is too large, return processing options
	options := sp.GetProcessingOptions()
	content := sp.buildProcessingOptionsMessageWithContext(data, options, currentContextTokens)

	return &ProcessedStreamResult{
		ToolCallID:     toolCallID,
		Content:        content,
		ProcessingMode: "auto",
		Options:        options,
		Data:           data,
	}, nil
}

// ShouldProcess determines if stream data needs processing based on size
func (sp *streamProcessor) ShouldProcess(data *StravaStreams) bool {
	if data == nil {
		return false
	}

	estimatedTokens := sp.EstimateTokens(data)
	return estimatedTokens > sp.config.MaxContextTokens
}

// EstimateTokens estimates the token count for stream data
func (sp *streamProcessor) EstimateTokens(data *StravaStreams) int {
	if data == nil {
		return 0
	}

	// Serialize to JSON to get approximate size
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling stream data for token estimation: %v", err)
		return 0
	}

	// Estimate tokens based on character count
	charCount := len(jsonData)
	estimatedTokens := int(float64(charCount) * sp.config.TokenPerCharRatio)

	log.Printf("Stream data size estimation: %d characters, ~%d tokens", charCount, estimatedTokens)
	return estimatedTokens
}

// GetProcessingOptions returns available processing options
func (sp *streamProcessor) GetProcessingOptions() []ProcessingOption {
	return []ProcessingOption{
		{
			Mode:        "raw",
			Description: "Get the actual stream data points (time, heart rate, power, etc.)",
			Command:     "Best for: Detailed analysis, specific time intervals, technical examination",
		},
		{
			Mode:        "derived",
			Description: "Get calculated features, statistics, and insights from the data",
			Command:     "Best for: Performance analysis, training insights, pattern identification",
		},
		{
			Mode:        "ai-summary",
			Description: "Get an AI-generated summary focusing on key findings (requires summary_prompt)",
			Command:     "Best for: Quick overview, coaching insights, narrative understanding",
		},
		{
			Mode:        "auto",
			Description: "Let the system choose the best approach based on data size",
			Command:     "Best for: When unsure which mode to use",
		},
	}
}

// formatRawStreamData formats stream data as human-readable text
func (sp *streamProcessor) formatRawStreamData(data *StravaStreams) string {
	var builder strings.Builder
	
	builder.WriteString("📊 **Stream Data**\n\n")
	
	// Count total data points
	totalPoints := sp.countDataPoints(data)
	builder.WriteString(fmt.Sprintf("**Total Data Points:** %d\n\n", totalPoints))
	
	// List available stream types
	streamTypes := sp.getAvailableStreamTypes(data)
	builder.WriteString("**Available Streams:**\n")
	for _, streamType := range streamTypes {
		builder.WriteString(fmt.Sprintf("- %s\n", streamType))
	}
	
	builder.WriteString("\n**Stream Data:**\n")
	
	// Format each stream type
	if len(data.Time) > 0 {
		builder.WriteString(fmt.Sprintf("- **Time:** %d data points (0-%d seconds)\n", len(data.Time), data.Time[len(data.Time)-1]))
	}
	if len(data.Distance) > 0 {
		builder.WriteString(fmt.Sprintf("- **Distance:** %d data points (%.2f-%.2f meters)\n", len(data.Distance), data.Distance[0], data.Distance[len(data.Distance)-1]))
	}
	if len(data.Heartrate) > 0 {
		min, max := sp.findMinMaxInt(data.Heartrate)
		builder.WriteString(fmt.Sprintf("- **Heart Rate:** %d data points (%d-%d bpm)\n", len(data.Heartrate), min, max))
	}
	if len(data.Watts) > 0 {
		min, max := sp.findMinMaxInt(data.Watts)
		builder.WriteString(fmt.Sprintf("- **Power:** %d data points (%d-%d watts)\n", len(data.Watts), min, max))
	}
	if len(data.Cadence) > 0 {
		min, max := sp.findMinMaxInt(data.Cadence)
		builder.WriteString(fmt.Sprintf("- **Cadence:** %d data points (%d-%d rpm)\n", len(data.Cadence), min, max))
	}
	if len(data.Altitude) > 0 {
		min, max := sp.findMinMaxFloat(data.Altitude)
		builder.WriteString(fmt.Sprintf("- **Altitude:** %d data points (%.1f-%.1f meters)\n", len(data.Altitude), min, max))
	}
	if len(data.VelocitySmooth) > 0 {
		min, max := sp.findMinMaxFloat(data.VelocitySmooth)
		builder.WriteString(fmt.Sprintf("- **Velocity:** %d data points (%.2f-%.2f m/s)\n", len(data.VelocitySmooth), min, max))
	}
	if len(data.Temp) > 0 {
		min, max := sp.findMinMaxInt(data.Temp)
		builder.WriteString(fmt.Sprintf("- **Temperature:** %d data points (%d-%d°C)\n", len(data.Temp), min, max))
	}
	if len(data.GradeSmooth) > 0 {
		min, max := sp.findMinMaxFloat(data.GradeSmooth)
		builder.WriteString(fmt.Sprintf("- **Grade:** %d data points (%.1f%%-%.1f%%)\n", len(data.GradeSmooth), min*100, max*100))
	}
	if len(data.Moving) > 0 {
		trueCount := 0
		for _, moving := range data.Moving {
			if moving {
				trueCount++
			}
		}
		movingPercent := float64(trueCount) / float64(len(data.Moving)) * 100
		builder.WriteString(fmt.Sprintf("- **Moving:** %d data points (%.1f%% moving time)\n", len(data.Moving), movingPercent))
	}
	if len(data.Latlng) > 0 {
		builder.WriteString(fmt.Sprintf("- **GPS Coordinates:** %d data points\n", len(data.Latlng)))
	}
	
	return builder.String()
}

// buildProcessingOptionsMessage creates the message shown when data is too large
func (sp *streamProcessor) buildProcessingOptionsMessage(data *StravaStreams, options []ProcessingOption) string {
	return sp.buildProcessingOptionsMessageWithContext(data, options, 0)
}

// buildProcessingOptionsMessageWithContext creates the message with current context information
func (sp *streamProcessor) buildProcessingOptionsMessageWithContext(data *StravaStreams, options []ProcessingOption, currentContextTokens int) string {
	var builder strings.Builder
	
	estimatedTokens := sp.EstimateTokens(data)
	totalPoints := sp.countDataPoints(data)
	availableTokens := sp.config.MaxContextTokens - currentContextTokens
	
	builder.WriteString("⚠️ **Output too large for context window**\n\n")
	builder.WriteString(fmt.Sprintf("The stream data contains %d data points (~%d tokens) which exceeds the context window limit of %d tokens.\n\n", 
		totalPoints, estimatedTokens, sp.config.MaxContextTokens))
	
	builder.WriteString("📊 **Processing Mode Options:**\n\n")
	
	for _, option := range options {
		var emoji string
		switch option.Mode {
		case "raw":
			emoji = "🔍"
		case "derived":
			emoji = "📈"
		case "ai-summary":
			emoji = "🤖"
		case "auto":
			emoji = "⚡"
		}
		
		builder.WriteString(fmt.Sprintf("%s **%s** - %s\n", emoji, option.Mode, option.Description))
		builder.WriteString(fmt.Sprintf("   %s\n\n", option.Command))
	}
	
	builder.WriteString("📏 **Token Usage Estimates:**\n")
	builder.WriteString(fmt.Sprintf("- Page size 500: ~%d tokens per page\n", int(500*sp.config.TokenPerCharRatio*4))) // Rough estimate
	builder.WriteString(fmt.Sprintf("- Page size 1000: ~%d tokens per page\n", int(1000*sp.config.TokenPerCharRatio*4)))
	builder.WriteString(fmt.Sprintf("- Page size 2000: ~%d tokens per page\n", int(2000*sp.config.TokenPerCharRatio*4)))
	builder.WriteString(fmt.Sprintf("- Full dataset (-1): ~%d tokens (requires processing)\n", estimatedTokens))
	
	// Add current context information if available
	if currentContextTokens > 0 {
		builder.WriteString(fmt.Sprintf("\n💡 **Current context usage:** %d tokens (%d remaining)\n", 
			currentContextTokens, availableTokens))
		
		// Calculate optimal page size
		optimalPageSize := sp.calculateOptimalPageSize(availableTokens)
		builder.WriteString(fmt.Sprintf("**Recommended page size:** %d (fits comfortably in remaining context)\n", optimalPageSize))
	}
	
	builder.WriteString("\n💡 To proceed, call the get-activity-streams tool again with your preferred processing_mode parameter.")
	
	return builder.String()
}

// calculateOptimalPageSize calculates optimal page size based on available tokens
func (sp *streamProcessor) calculateOptimalPageSize(availableTokens int) int {
	// Reserve 20% buffer for safety
	usableTokens := int(float64(availableTokens) * 0.8)
	
	// Estimate data points that fit in available tokens
	// Rough estimate: each data point uses about 4 characters when serialized
	charsPerDataPoint := 4.0
	tokensPerDataPoint := charsPerDataPoint * sp.config.TokenPerCharRatio
	
	optimalPageSize := int(float64(usableTokens) / tokensPerDataPoint)
	
	// Apply bounds
	if optimalPageSize < 100 {
		optimalPageSize = 100 // Minimum useful page size
	}
	if optimalPageSize > sp.config.MaxPageSize {
		optimalPageSize = sp.config.MaxPageSize
	}
	
	return optimalPageSize
}

// Helper functions for data analysis
func (sp *streamProcessor) countDataPoints(data *StravaStreams) int {
	maxPoints := 0
	
	if len(data.Time) > maxPoints {
		maxPoints = len(data.Time)
	}
	if len(data.Distance) > maxPoints {
		maxPoints = len(data.Distance)
	}
	if len(data.Heartrate) > maxPoints {
		maxPoints = len(data.Heartrate)
	}
	if len(data.Watts) > maxPoints {
		maxPoints = len(data.Watts)
	}
	if len(data.Cadence) > maxPoints {
		maxPoints = len(data.Cadence)
	}
	if len(data.Altitude) > maxPoints {
		maxPoints = len(data.Altitude)
	}
	if len(data.VelocitySmooth) > maxPoints {
		maxPoints = len(data.VelocitySmooth)
	}
	if len(data.Temp) > maxPoints {
		maxPoints = len(data.Temp)
	}
	if len(data.GradeSmooth) > maxPoints {
		maxPoints = len(data.GradeSmooth)
	}
	if len(data.Moving) > maxPoints {
		maxPoints = len(data.Moving)
	}
	if len(data.Latlng) > maxPoints {
		maxPoints = len(data.Latlng)
	}
	
	return maxPoints
}

func (sp *streamProcessor) getAvailableStreamTypes(data *StravaStreams) []string {
	var types []string
	
	if len(data.Time) > 0 {
		types = append(types, "time")
	}
	if len(data.Distance) > 0 {
		types = append(types, "distance")
	}
	if len(data.Heartrate) > 0 {
		types = append(types, "heartrate")
	}
	if len(data.Watts) > 0 {
		types = append(types, "watts")
	}
	if len(data.Cadence) > 0 {
		types = append(types, "cadence")
	}
	if len(data.Altitude) > 0 {
		types = append(types, "altitude")
	}
	if len(data.VelocitySmooth) > 0 {
		types = append(types, "velocity_smooth")
	}
	if len(data.Temp) > 0 {
		types = append(types, "temp")
	}
	if len(data.GradeSmooth) > 0 {
		types = append(types, "grade_smooth")
	}
	if len(data.Moving) > 0 {
		types = append(types, "moving")
	}
	if len(data.Latlng) > 0 {
		types = append(types, "latlng")
	}
	
	return types
}

func (sp *streamProcessor) findMinMaxInt(data []int) (int, int) {
	if len(data) == 0 {
		return 0, 0
	}
	
	min, max := data[0], data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

func (sp *streamProcessor) findMinMaxFloat(data []float64) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}
	
	min, max := data[0], data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

// Error handling and fallback functions

// HandleProcessingError creates a user-friendly error message with alternatives
func (sp *streamProcessor) HandleProcessingError(err error, activityID int64, processingMode string, data *StravaStreams) *ProcessedStreamResult {
	log.Printf("Stream processing error for activity %d (mode: %s): %v", activityID, processingMode, err)
	
	// Extract or create stream processing error
	var streamErr *StreamProcessingError
	if errors.As(err, &streamErr) {
		// Use existing stream processing error
	} else {
		// Create new stream processing error from generic error
		streamErr = NewStreamProcessingError("processing_failure", err.Error(), activityID, processingMode).
			WithOriginalError(err)
	}
	
	// Add data size information if available
	if data != nil {
		dataSize := sp.EstimateTokens(data)
		streamErr = streamErr.WithDataSize(dataSize)
	}
	
	// Add alternative processing methods
	alternatives := sp.GetProcessingOptions()
	streamErr = streamErr.WithAlternatives(alternatives)
	
	// Format error message for LLM
	content := sp.formatErrorMessage(streamErr)
	
	return &ProcessedStreamResult{
		ToolCallID:     fmt.Sprintf("error_%d", activityID),
		Content:        content,
		ProcessingMode: "error",
		Options:        alternatives,
		Data:           streamErr,
	}
}

// formatErrorMessage creates a human-readable error message with alternatives
func (sp *streamProcessor) formatErrorMessage(err *StreamProcessingError) string {
	var builder strings.Builder
	
	// Error header with appropriate emoji
	var emoji string
	switch err.Type {
	case "strava_api_failure":
		emoji = "🔌"
	case "context_exceeded":
		emoji = "📏"
	case "processing_failure":
		emoji = "⚠️"
	case "invalid_request":
		emoji = "❌"
	case "data_corrupted":
		emoji = "🔧"
	default:
		emoji = "⚠️"
	}
	
	builder.WriteString(fmt.Sprintf("%s **Stream Processing Error**\n\n", emoji))
	builder.WriteString(fmt.Sprintf("**Error:** %s\n", err.Message))
	
	if err.ActivityID > 0 {
		builder.WriteString(fmt.Sprintf("**Activity ID:** %d\n", err.ActivityID))
	}
	
	if err.ProcessingMode != "" {
		builder.WriteString(fmt.Sprintf("**Processing Mode:** %s\n", err.ProcessingMode))
	}
	
	// Add data size context if available
	if err.DataSize > 0 {
		builder.WriteString(fmt.Sprintf("**Data Size:** ~%d tokens\n", err.DataSize))
	}
	
	if err.AvailableTokens > 0 {
		builder.WriteString(fmt.Sprintf("**Available Context:** %d tokens\n", err.AvailableTokens))
	}
	
	// Add specific error context
	if len(err.Context) > 0 {
		builder.WriteString("\n**Additional Context:**\n")
		for key, value := range err.Context {
			builder.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
		}
	}
	
	// Add alternative processing methods
	if len(err.Alternatives) > 0 {
		builder.WriteString("\n🔄 **Alternative Processing Methods:**\n\n")
		
		for _, alt := range err.Alternatives {
			// Skip the failed mode
			if alt.Mode == err.ProcessingMode {
				continue
			}
			
			var altEmoji string
			switch alt.Mode {
			case "raw":
				altEmoji = "🔍"
			case "derived":
				altEmoji = "📈"
			case "ai-summary":
				altEmoji = "🤖"
			case "auto":
				altEmoji = "⚡"
			}
			
			builder.WriteString(fmt.Sprintf("%s **%s** - %s\n", altEmoji, alt.Mode, alt.Description))
			builder.WriteString(fmt.Sprintf("   %s\n\n", alt.Command))
		}
	}
	
	// Add recovery suggestions
	builder.WriteString("💡 **Recovery Suggestions:**\n")
	builder.WriteString(sp.getRecoverySuggestions(err))
	
	return builder.String()
}

// getRecoverySuggestions provides specific recovery suggestions based on error type
func (sp *streamProcessor) getRecoverySuggestions(err *StreamProcessingError) string {
	var builder strings.Builder
	
	switch err.Type {
	case "strava_api_failure":
		builder.WriteString("- Check your Strava API connection and authentication\n")
		builder.WriteString("- Verify the activity ID exists and is accessible\n")
		builder.WriteString("- Try again in a few moments if this is a temporary API issue\n")
		
	case "context_exceeded":
		builder.WriteString("- Use pagination with smaller page_size (e.g., 500-1000 data points)\n")
		builder.WriteString("- Try 'derived' mode for statistical analysis instead of raw data\n")
		builder.WriteString("- Use 'ai-summary' mode for a condensed overview\n")
		
	case "processing_failure":
		builder.WriteString("- Try a different processing mode (raw, derived, or ai-summary)\n")
		builder.WriteString("- Use pagination to process data in smaller chunks\n")
		builder.WriteString("- Check if the activity has complete stream data\n")
		
	case "invalid_request":
		builder.WriteString("- Verify all required parameters are provided\n")
		builder.WriteString("- Check that activity_id is a valid positive integer\n")
		builder.WriteString("- Ensure processing_mode is one of: raw, derived, ai-summary, auto\n")
		
	case "data_corrupted":
		builder.WriteString("- Try requesting different stream types\n")
		builder.WriteString("- Use a different resolution (low, medium, high)\n")
		builder.WriteString("- Check if the activity was recorded properly in Strava\n")
		
	default:
		builder.WriteString("- Try a different processing mode\n")
		builder.WriteString("- Use pagination to reduce data size\n")
		builder.WriteString("- Contact support if the issue persists\n")
	}
	
	return builder.String()
}

// CreateFallbackFormatter creates a basic formatter for when primary processing fails
func (sp *streamProcessor) CreateFallbackFormatter(data *StravaStreams, activityID int64, failedMode string) string {
	var builder strings.Builder
	
	builder.WriteString("🔄 **Fallback Stream Information**\n\n")
	builder.WriteString(fmt.Sprintf("**Activity ID:** %d\n", activityID))
	builder.WriteString(fmt.Sprintf("**Failed Processing Mode:** %s\n", failedMode))
	builder.WriteString("**Status:** Providing basic stream information as fallback\n\n")
	
	if data == nil {
		builder.WriteString("❌ **No stream data available**\n")
		builder.WriteString("The activity may not have recorded stream data or there was an error retrieving it.\n\n")
		builder.WriteString("**Suggestions:**\n")
		builder.WriteString("- Verify the activity ID is correct\n")
		builder.WriteString("- Check if the activity has GPS/sensor data recorded\n")
		builder.WriteString("- Try requesting different stream types\n")
		return builder.String()
	}
	
	// Basic stream information
	totalPoints := sp.countDataPoints(data)
	streamTypes := sp.getAvailableStreamTypes(data)
	estimatedTokens := sp.EstimateTokens(data)
	
	builder.WriteString("📊 **Basic Stream Information:**\n")
	builder.WriteString(fmt.Sprintf("- **Total Data Points:** %d\n", totalPoints))
	builder.WriteString(fmt.Sprintf("- **Estimated Size:** ~%d tokens\n", estimatedTokens))
	builder.WriteString(fmt.Sprintf("- **Available Stream Types:** %s\n", strings.Join(streamTypes, ", ")))
	
	// Time range if available
	if len(data.Time) > 0 {
		duration := data.Time[len(data.Time)-1] - data.Time[0]
		builder.WriteString(fmt.Sprintf("- **Duration:** %d seconds (%.1f minutes)\n", duration, float64(duration)/60))
	}
	
	// Basic statistics for key metrics
	builder.WriteString("\n📈 **Basic Statistics:**\n")
	
	if len(data.Heartrate) > 0 {
		min, max := sp.findMinMaxInt(data.Heartrate)
		avg := sp.calculateAverageInt(data.Heartrate)
		builder.WriteString(fmt.Sprintf("- **Heart Rate:** %d-%d bpm (avg: %.1f bpm)\n", min, max, avg))
	}
	
	if len(data.Watts) > 0 {
		min, max := sp.findMinMaxInt(data.Watts)
		avg := sp.calculateAverageInt(data.Watts)
		builder.WriteString(fmt.Sprintf("- **Power:** %d-%d watts (avg: %.1f watts)\n", min, max, avg))
	}
	
	if len(data.VelocitySmooth) > 0 {
		min, max := sp.findMinMaxFloat(data.VelocitySmooth)
		avg := sp.calculateAverageFloat(data.VelocitySmooth)
		builder.WriteString(fmt.Sprintf("- **Speed:** %.2f-%.2f m/s (avg: %.2f m/s)\n", min, max, avg))
	}
	
	if len(data.Altitude) > 0 {
		min, max := sp.findMinMaxFloat(data.Altitude)
		elevationGain := sp.calculateElevationGain(data.Altitude)
		builder.WriteString(fmt.Sprintf("- **Elevation:** %.1f-%.1f m (gain: %.1f m)\n", min, max, elevationGain))
	}
	
	// Suggest next steps
	builder.WriteString("\n💡 **Next Steps:**\n")
	builder.WriteString("- Try 'derived' mode for detailed statistical analysis\n")
	builder.WriteString("- Use 'ai-summary' mode with a custom prompt for insights\n")
	builder.WriteString("- Use pagination (page_size parameter) for large datasets\n")
	
	return builder.String()
}

// Helper functions for fallback calculations
func (sp *streamProcessor) calculateAverageInt(data []int) float64 {
	if len(data) == 0 {
		return 0
	}
	
	sum := 0
	for _, v := range data {
		sum += v
	}
	return float64(sum) / float64(len(data))
}

func (sp *streamProcessor) calculateAverageFloat(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func (sp *streamProcessor) calculateElevationGain(altitude []float64) float64 {
	if len(altitude) < 2 {
		return 0
	}
	
	gain := 0.0
	for i := 1; i < len(altitude); i++ {
		diff := altitude[i] - altitude[i-1]
		if diff > 0 {
			gain += diff
		}
	}
	return gain
}