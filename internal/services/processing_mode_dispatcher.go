package services

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// ProcessingModeDispatcher handles routing to different processing modes
type ProcessingModeDispatcher interface {
	Dispatch(mode string, data *StravaStreams, params ProcessingParams) (*ProcessedStreamResult, error)
	ValidateMode(mode string) error
	GetSupportedModes() []string
}

// ProcessingParams contains parameters for processing modes
type ProcessingParams struct {
	ToolCallID      string
	ActivityID      int64
	PageNumber      int
	PageSize        int
	SummaryPrompt   string
	Resolution      string
	StreamTypes     []string
	ProcessingMode  string
}

// processingModeDispatcher implements the ProcessingModeDispatcher interface
type processingModeDispatcher struct {
	streamProcessor  StreamProcessor
	summaryProcessor SummaryProcessor
}

// NewProcessingModeDispatcher creates a new processing mode dispatcher
func NewProcessingModeDispatcher(streamProcessor StreamProcessor, summaryProcessor SummaryProcessor) ProcessingModeDispatcher {
	return &processingModeDispatcher{
		streamProcessor:  streamProcessor,
		summaryProcessor: summaryProcessor,
	}
}

// Dispatch routes processing to the appropriate mode handler
func (pmd *processingModeDispatcher) Dispatch(mode string, data *StravaStreams, params ProcessingParams) (*ProcessedStreamResult, error) {
	if err := pmd.ValidateMode(mode); err != nil {
		streamErr := NewStreamProcessingError("invalid_request", err.Error(), params.ActivityID, mode).
			WithOriginalError(err).
			WithContext("requested_mode", mode).
			WithContext("supported_modes", pmd.GetSupportedModes())
		return nil, streamErr
	}

	log.Printf("Dispatching stream processing with mode: %s, activity: %d", mode, params.ActivityID)

	// Validate stream data before processing
	if data == nil {
		streamErr := NewStreamProcessingError("data_corrupted", "Stream data is nil", params.ActivityID, mode).
			WithContext("validation_stage", "pre_processing")
		return nil, streamErr
	}

	// Dispatch to appropriate handler with error recovery
	var result *ProcessedStreamResult
	var err error

	switch mode {
	case "auto":
		result, err = pmd.handleAutoModeWithFallback(data, params)
	case "raw":
		result, err = pmd.handleRawMode(data, params)
	case "derived":
		result, err = pmd.handleDerivedModeWithFallback(data, params)
	case "ai-summary":
		result, err = pmd.handleAISummaryMode(data, params)
	default:
		streamErr := NewStreamProcessingError("invalid_request", fmt.Sprintf("Unsupported processing mode: %s", mode), params.ActivityID, mode).
			WithContext("requested_mode", mode)
		return nil, streamErr
	}

	// If processing failed and no result was returned, create emergency fallback
	if err != nil && result == nil {
		log.Printf("All processing failed for activity %d mode %s, creating emergency fallback", params.ActivityID, mode)
		result = pmd.createEmergencyFallback(data, params, err)
	}

	return result, err
}

// handleAutoModeWithFallback handles auto mode with enhanced error recovery
func (pmd *processingModeDispatcher) handleAutoModeWithFallback(data *StravaStreams, params ProcessingParams) (*ProcessedStreamResult, error) {
	// Try the original auto mode logic
	result, err := pmd.handleAutoMode(data, params)
	if err == nil {
		return result, nil
	}

	log.Printf("Auto mode failed for activity %d: %v", params.ActivityID, err)

	// If auto mode fails, try raw mode as fallback
	log.Printf("Attempting raw mode fallback for activity %d", params.ActivityID)
	fallbackResult, fallbackErr := pmd.handleRawMode(data, params)
	if fallbackErr == nil {
		// Add fallback notice
		fallbackNotice := "🔄 **Fallback Mode:** raw (auto mode failed)\n\n"
		fallbackResult.Content = fallbackNotice + fallbackResult.Content
		fallbackResult.ProcessingMode = "raw-fallback"
		return fallbackResult, nil
	}

	// Return original error if fallback also fails
	return nil, err
}

// handleDerivedModeWithFallback handles derived mode with fallback to raw
func (pmd *processingModeDispatcher) handleDerivedModeWithFallback(data *StravaStreams, params ProcessingParams) (*ProcessedStreamResult, error) {
	// Try the original derived mode logic
	result, err := pmd.handleDerivedMode(data, params)
	if err == nil {
		return result, nil
	}

	log.Printf("Derived mode failed for activity %d: %v", params.ActivityID, err)

	// If derived mode fails, try raw mode as fallback
	log.Printf("Attempting raw mode fallback for activity %d", params.ActivityID)
	fallbackResult, fallbackErr := pmd.handleRawMode(data, params)
	if fallbackErr == nil {
		// Add fallback notice
		fallbackNotice := "🔄 **Fallback Mode:** raw (derived mode failed)\n\n"
		fallbackResult.Content = fallbackNotice + fallbackResult.Content
		fallbackResult.ProcessingMode = "raw-fallback"
		return fallbackResult, nil
	}

	// Return original error if fallback also fails
	return nil, err
}

// createEmergencyFallback creates a basic result when all processing fails
func (pmd *processingModeDispatcher) createEmergencyFallback(data *StravaStreams, params ProcessingParams, originalErr error) *ProcessedStreamResult {
	log.Printf("Creating emergency fallback for activity %d", params.ActivityID)

	var content strings.Builder
	content.WriteString("🚨 **Emergency Fallback**\n\n")
	content.WriteString(fmt.Sprintf("**Activity ID:** %d\n", params.ActivityID))
	content.WriteString(fmt.Sprintf("**Requested Mode:** %s\n", params.ProcessingMode))
	content.WriteString(fmt.Sprintf("**Error:** %v\n\n", originalErr))

	if data != nil {
		// Provide very basic information
		dataPoints := 0
		streamTypes := []string{}

		if len(data.Time) > 0 {
			dataPoints = len(data.Time)
			streamTypes = append(streamTypes, "time")
		}
		if len(data.Heartrate) > 0 {
			streamTypes = append(streamTypes, "heartrate")
		}
		if len(data.Watts) > 0 {
			streamTypes = append(streamTypes, "watts")
		}
		if len(data.Distance) > 0 {
			streamTypes = append(streamTypes, "distance")
		}

		content.WriteString("📊 **Basic Information:**\n")
		content.WriteString(fmt.Sprintf("- Data points: %d\n", dataPoints))
		content.WriteString(fmt.Sprintf("- Stream types: %s\n", strings.Join(streamTypes, ", ")))
	} else {
		content.WriteString("❌ **No stream data available**\n")
	}

	content.WriteString("\n💡 **Suggestions:**\n")
	content.WriteString("- Check your Strava connection\n")
	content.WriteString("- Verify the activity ID is correct\n")
	content.WriteString("- Try a different processing mode\n")
	content.WriteString("- Contact support if the issue persists\n")

	return &ProcessedStreamResult{
		ToolCallID:     params.ToolCallID,
		Content:        content.String(),
		ProcessingMode: "emergency-fallback",
		Data:           originalErr,
	}
}

// ValidateMode checks if the processing mode is supported
func (pmd *processingModeDispatcher) ValidateMode(mode string) error {
	supportedModes := pmd.GetSupportedModes()
	for _, supportedMode := range supportedModes {
		if mode == supportedMode {
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrInvalidProcessingMode, mode)
}

// GetSupportedModes returns list of supported processing modes
func (pmd *processingModeDispatcher) GetSupportedModes() []string {
	return []string{"auto", "raw", "derived", "ai-summary"}
}

// handleAutoMode automatically determines the best processing approach
func (pmd *processingModeDispatcher) handleAutoMode(data *StravaStreams, params ProcessingParams) (*ProcessedStreamResult, error) {
	// Use the stream processor to determine if processing is needed
	if !pmd.streamProcessor.ShouldProcess(data) {
		// Data is small enough, return raw formatted data
		return pmd.handleRawMode(data, params)
	}

	// Data is too large, return processing options
	return pmd.streamProcessor.ProcessStreamOutput(data, params.ToolCallID)
}

// handleRawMode returns raw stream data formatted as human-readable text
func (pmd *processingModeDispatcher) handleRawMode(data *StravaStreams, params ProcessingParams) (*ProcessedStreamResult, error) {
	// Validate stream data
	if data == nil {
		streamErr := NewStreamProcessingError("data_corrupted", "Stream data is nil or corrupted", params.ActivityID, "raw").
			WithContext("data_validation", "nil_data")
		return nil, streamErr
	}

	// Check if pagination is needed
	if pmd.streamProcessor.ShouldProcess(data) && params.PageSize > 0 {
		// Data is large and pagination is requested
		return pmd.handlePaginatedRawMode(data, params)
	}

	// Return formatted raw data
	result, err := pmd.streamProcessor.ProcessStreamOutput(data, params.ToolCallID)
	if err != nil {
		// Create detailed error for raw processing failure
		streamErr := NewStreamProcessingError("processing_failure", "Failed to process raw stream data", params.ActivityID, "raw").
			WithOriginalError(err).
			WithDataSize(pmd.streamProcessor.EstimateTokens(data))
		
		// Attempt to create fallback formatter
		streamProcessor := pmd.streamProcessor.(*streamProcessor)
		fallbackContent := streamProcessor.CreateFallbackFormatter(data, params.ActivityID, "raw")
		
		return &ProcessedStreamResult{
			ToolCallID:     params.ToolCallID,
			Content:        fallbackContent,
			ProcessingMode: "raw-fallback",
			Data:           streamErr,
		}, nil
	}

	// Override processing mode to raw
	result.ProcessingMode = "raw"
	result.Options = nil // Remove options for raw mode

	return result, nil
}

// handleDerivedMode processes stream data to extract derived features
func (pmd *processingModeDispatcher) handleDerivedMode(data *StravaStreams, params ProcessingParams) (*ProcessedStreamResult, error) {
	// For now, return a placeholder indicating derived features processing is not yet implemented
	// This will be implemented in task 3
	content := fmt.Sprintf("🔧 **Derived Features Processing**\n\n"+
		"Derived features processing for activity %d is not yet implemented.\n"+
		"This will extract statistical patterns, inflection points, trends, and insights from the stream data.\n\n"+
		"**Available data points:** %d\n"+
		"**Processing mode:** derived\n\n"+
		"Please use 'raw' mode for now to access the stream data directly.",
		params.ActivityID, pmd.streamProcessor.EstimateTokens(data)/4) // Rough estimate of data points

	return &ProcessedStreamResult{
		ToolCallID:     params.ToolCallID,
		Content:        content,
		ProcessingMode: "derived",
		Data:           data,
	}, nil
}

// handleAISummaryMode processes stream data using AI summarization
func (pmd *processingModeDispatcher) handleAISummaryMode(data *StravaStreams, params ProcessingParams) (*ProcessedStreamResult, error) {
	// Validate that summary prompt is provided
	if params.SummaryPrompt == "" {
		streamErr := NewStreamProcessingError("invalid_request", "summary_prompt is required for ai-summary mode", params.ActivityID, "ai-summary").
			WithContext("missing_parameter", "summary_prompt")
		return nil, streamErr
	}

	// Check if summary processor is available
	if pmd.summaryProcessor == nil {
		streamErr := NewStreamProcessingError("processor_unavailable", "AI summary processor not available", params.ActivityID, "ai-summary").
			WithContext("processor_type", "summary")
		return nil, streamErr
	}

	// Generate AI summary using the summary processor
	ctx := context.Background()
	summary, err := pmd.summaryProcessor.GenerateSummary(ctx, data, params.ActivityID, params.SummaryPrompt)
	if err != nil {
		log.Printf("AI summarization failed for activity %d: %v", params.ActivityID, err)
		
		// Create detailed error for AI failure
		streamErr := NewStreamProcessingError("ai_summary_failure", "AI summarization failed", params.ActivityID, "ai-summary").
			WithOriginalError(err).
			WithContext("summary_prompt", params.SummaryPrompt)
		
		// Attempt automatic fallback to derived features mode
		log.Printf("Attempting automatic fallback to derived features mode for activity %d", params.ActivityID)
		fallbackResult, fallbackErr := pmd.handleDerivedMode(data, params)
		if fallbackErr == nil {
			// Add fallback notice to the result
			fallbackNotice := "🔄 **Fallback Mode:** derived (AI summarization failed)\n\n"
			fallbackResult.Content = fallbackNotice + fallbackResult.Content
			fallbackResult.ProcessingMode = "derived-fallback"
			return fallbackResult, nil
		}
		
		// If fallback also fails, return the original AI error
		log.Printf("Fallback to derived features also failed for activity %d: %v", params.ActivityID, fallbackErr)
		return nil, streamErr
	}

	// Return the raw AI summary output without additional formatting
	// The AI provides its own structure and formatting
	return &ProcessedStreamResult{
		ToolCallID:     params.ToolCallID,
		Content:        summary.Summary,
		ProcessingMode: "ai-summary",
		Data:           summary,
	}, nil
}

// handlePaginatedRawMode handles raw mode with pagination
func (pmd *processingModeDispatcher) handlePaginatedRawMode(data *StravaStreams, params ProcessingParams) (*ProcessedStreamResult, error) {
	// For now, return a placeholder indicating pagination is not yet implemented
	// This will be implemented in task 5
	content := fmt.Sprintf("📄 **Paginated Raw Data Processing**\n\n"+
		"Paginated processing for activity %d is not yet implemented.\n"+
		"This will divide the stream data into manageable pages for incremental processing.\n\n"+
		"**Requested page:** %d\n"+
		"**Page size:** %d\n"+
		"**Available data points:** %d\n"+
		"**Processing mode:** raw (paginated)\n\n"+
		"Please use a smaller dataset or wait for pagination implementation.",
		params.ActivityID, params.PageNumber, params.PageSize, pmd.streamProcessor.EstimateTokens(data)/4)

	return &ProcessedStreamResult{
		ToolCallID:     params.ToolCallID,
		Content:        content,
		ProcessingMode: "raw",
		Data:           data,
	}, nil
}