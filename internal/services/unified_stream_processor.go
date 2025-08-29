package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"bodda/internal/models"
)

// UnifiedStreamProcessor handles all stream requests through pagination interface
type UnifiedStreamProcessor struct {
	config               *StreamConfig
	paginationCalculator *PaginationCalculator
	derivedProcessor     DerivedFeaturesProcessor
	summaryProcessor     SummaryProcessor
	outputFormatter      OutputFormatter
	stravaService        StravaService
	performanceMonitor   *StreamPerformanceMonitor
}

// NewUnifiedStreamProcessor creates a new unified stream processor
func NewUnifiedStreamProcessor(
	config *StreamConfig,
	stravaService StravaService,
	derivedProcessor DerivedFeaturesProcessor,
	summaryProcessor SummaryProcessor,
	outputFormatter OutputFormatter,
) *UnifiedStreamProcessor {
	paginationCalculator := NewPaginationCalculator(config, stravaService)
	performanceMonitor := NewStreamPerformanceMonitor(true) // Enable performance monitoring
	
	return &UnifiedStreamProcessor{
		config:               config,
		paginationCalculator: paginationCalculator,
		derivedProcessor:     derivedProcessor,
		summaryProcessor:     summaryProcessor,
		outputFormatter:      outputFormatter,
		stravaService:        stravaService,
		performanceMonitor:   performanceMonitor,
	}
}

// ProcessPaginatedStreamRequest processes a paginated stream request with the specified mode
func (usp *UnifiedStreamProcessor) ProcessPaginatedStreamRequest(user *models.User, req *PaginatedStreamRequest, currentContextTokens int) (*StreamPage, error) {
	// Start performance monitoring
	timer := usp.performanceMonitor.StartOperation(context.Background(), "paginated_stream_request", req.PageSize)
	defer func() {
		timer.EndOperation(nil) // Will be updated with actual error if one occurs
	}()

	// Validate request
	if err := usp.validateRequest(req); err != nil {
		// Create detailed error for invalid request
		streamErr := NewStreamProcessingError("invalid_request", err.Error(), req.ActivityID, req.ProcessingMode).
			WithOriginalError(err).
			WithContext("page_number", req.PageNumber).
			WithContext("page_size", req.PageSize).
			WithContext("stream_types", req.StreamTypes)
		
		return nil, streamErr
	}
	
	// Handle negative page size (full dataset request)
	if req.PageSize < 0 {
		return usp.processFullDatasetRequest(user, req, currentContextTokens)
	}
	
	// Calculate optimal page size if not specified or if current page size is too large
	if req.PageSize == 0 || usp.wouldExceedContext(req.PageSize, len(req.StreamTypes), currentContextTokens) {
		req.PageSize = usp.paginationCalculator.CalculateOptimalPageSize(currentContextTokens)
		log.Printf("Adjusted page size to %d based on available context", req.PageSize)
	}
	
	// Estimate total pages
	totalPages, err := usp.paginationCalculator.EstimateTotalPages(user, req.ActivityID, req.StreamTypes, req.Resolution, req.PageSize)
	if err != nil {
		// Create detailed error for pagination failure
		streamErr := NewStreamProcessingError("pagination_failure", "Failed to estimate total pages", req.ActivityID, req.ProcessingMode).
			WithOriginalError(err).
			WithContext("resolution", req.Resolution).
			WithContext("stream_types", req.StreamTypes)
		
		return nil, streamErr
	}
	
	// Validate page number
	if req.PageNumber < 1 || req.PageNumber > totalPages {
		streamErr := NewStreamProcessingError("invalid_request", 
			fmt.Sprintf("Invalid page number %d (total pages: %d)", req.PageNumber, totalPages), 
			req.ActivityID, req.ProcessingMode).
			WithContext("requested_page", req.PageNumber).
			WithContext("total_pages", totalPages).
			WithContext("page_size", req.PageSize)
		
		return nil, streamErr
	}
	
	// Request the specific data chunk
	streamData, err := usp.paginationCalculator.RequestSpecificDataChunk(user, req)
	if err != nil {
		// Create detailed error for Strava API failure
		streamErr := NewStreamProcessingError("strava_api_failure", "Failed to retrieve stream data from Strava API", req.ActivityID, req.ProcessingMode).
			WithOriginalError(err).
			WithContext("page_number", req.PageNumber).
			WithContext("page_size", req.PageSize).
			WithContext("resolution", req.Resolution)
		
		return nil, streamErr
	}
	
	// Apply processing mode to the paginated data
	processedData, err := usp.applyProcessingModeWithFallback(user, req, streamData)
	if err != nil {
		// Update timer with error before creating fallback
		timer.EndOperation(err)
		
		// If processing fails, create fallback result
		log.Printf("Processing mode %s failed for activity %d, creating fallback", req.ProcessingMode, req.ActivityID)
		
		fallbackData := usp.createFallbackStreamPage(req, streamData, err)
		return fallbackData, nil
	}
	
	// Calculate time range for this page
	timeRange := usp.calculateTimeRange(streamData)
	
	// Create the stream page result
	streamPage := &StreamPage{
		ActivityID:      req.ActivityID,
		PageNumber:      req.PageNumber,
		TotalPages:      totalPages,
		ProcessingMode:  req.ProcessingMode,
		Data:            processedData,
		TimeRange:       timeRange,
		Instructions:    usp.generatePageInstructions(req, totalPages),
		HasNextPage:     req.PageNumber < totalPages,
		EstimatedTokens: usp.paginationCalculator.EstimatePageTokens(req.PageSize, len(req.StreamTypes)),
	}
	
	return streamPage, nil
}

// processFullDatasetRequest handles requests for the full dataset (negative page size)
func (usp *UnifiedStreamProcessor) processFullDatasetRequest(user *models.User, req *PaginatedStreamRequest, currentContextTokens int) (*StreamPage, error) {
	log.Printf("Processing full dataset request for activity %d", req.ActivityID)
	
	// Get the full stream data
	streamData, err := usp.stravaService.GetActivityStreams(user, req.ActivityID, req.StreamTypes, req.Resolution)
	if err != nil {
		// Create detailed error for Strava API failure
		streamErr := NewStreamProcessingError("strava_api_failure", "Failed to retrieve full stream data from Strava API", req.ActivityID, req.ProcessingMode).
			WithOriginalError(err).
			WithContext("resolution", req.Resolution).
			WithContext("stream_types", req.StreamTypes)
		
		return nil, streamErr
	}
	
	// Check if the full dataset would exceed context limits
	estimatedTokens := usp.estimateStreamTokens(streamData)
	availableTokens := usp.config.MaxContextTokens - currentContextTokens
	
	if estimatedTokens > availableTokens {
		// Full dataset is too large, must use processing
		if req.ProcessingMode == "raw" {
			streamErr := NewStreamProcessingError("context_exceeded", 
				fmt.Sprintf("Full dataset too large for raw mode (%d tokens estimated, %d available)", estimatedTokens, availableTokens), 
				req.ActivityID, req.ProcessingMode).
				WithDataSize(estimatedTokens).
				WithAvailableTokens(availableTokens).
				WithContext("suggested_modes", []string{"derived", "ai-summary"})
			
			return nil, streamErr
		}
	}
	
	// Apply processing mode to the full dataset
	processedData, err := usp.applyProcessingModeWithFallback(user, req, streamData)
	if err != nil {
		// If processing fails, create fallback result
		log.Printf("Processing mode %s failed for full dataset activity %d, creating fallback", req.ProcessingMode, req.ActivityID)
		
		fallbackData := usp.createFallbackStreamPage(req, streamData, err)
		return fallbackData, nil
	}
	
	// Calculate time range for the full dataset
	timeRange := usp.calculateTimeRange(streamData)
	
	// Create the stream page result for full dataset
	streamPage := &StreamPage{
		ActivityID:      req.ActivityID,
		PageNumber:      1,
		TotalPages:      1,
		ProcessingMode:  req.ProcessingMode,
		Data:            processedData,
		TimeRange:       timeRange,
		Instructions:    "Full dataset processed. No pagination needed.",
		HasNextPage:     false,
		EstimatedTokens: estimatedTokens,
	}
	
	return streamPage, nil
}

// applyProcessingMode applies the specified processing mode to the stream data
func (usp *UnifiedStreamProcessor) applyProcessingMode(user *models.User, req *PaginatedStreamRequest, streamData *StravaStreams) (interface{}, error) {
	switch req.ProcessingMode {
	case "raw":
		// Return formatted raw stream data
		return usp.outputFormatter.FormatStreamData(streamData, "raw"), nil
		
	case "derived":
		// Get lap data if available for enhanced analysis
		var laps []StravaLap
		if activityDetail, err := usp.stravaService.GetActivityDetail(user, req.ActivityID); err == nil {
			laps = activityDetail.Laps
		}
		
		// Extract derived features
		features, err := usp.derivedProcessor.ExtractFeatures(streamData, laps)
		if err != nil {
			return nil, fmt.Errorf("failed to extract derived features: %w", err)
		}
		
		// Format the derived features
		return usp.outputFormatter.FormatDerivedFeatures(features), nil
		
	case "ai-summary":
		// Validate that summary prompt is provided
		if req.SummaryPrompt == "" {
			return nil, fmt.Errorf("summary_prompt is required for ai-summary mode")
		}
		
		// Generate AI summary using context and proper interface
		ctx := context.Background()
		summary, err := usp.summaryProcessor.GenerateSummary(ctx, streamData, req.ActivityID, req.SummaryPrompt)
		if err != nil {
			return nil, fmt.Errorf("failed to generate AI summary: %w", err)
		}
		
		// Format the summary
		return usp.outputFormatter.FormatStreamSummary(summary), nil
		
	default:
		return nil, fmt.Errorf("unsupported processing mode: %s", req.ProcessingMode)
	}
}

// validateRequest validates the paginated stream request
func (usp *UnifiedStreamProcessor) validateRequest(req *PaginatedStreamRequest) error {
	if req.ActivityID <= 0 {
		return fmt.Errorf("activity_id must be positive")
	}
	
	if len(req.StreamTypes) == 0 {
		return fmt.Errorf("stream_types cannot be empty")
	}
	
	if req.PageNumber < 1 && req.PageSize >= 0 {
		return fmt.Errorf("page_number must be >= 1 for paginated requests")
	}
	
	if req.PageSize == 0 {
		req.PageSize = usp.config.DefaultPageSize
	}
	
	if req.PageSize > usp.config.MaxPageSize {
		return fmt.Errorf("page_size cannot exceed %d", usp.config.MaxPageSize)
	}
	
	validModes := map[string]bool{
		"raw":        true,
		"derived":    true,
		"ai-summary": true,
	}
	
	if !validModes[req.ProcessingMode] {
		return fmt.Errorf("invalid processing_mode: %s", req.ProcessingMode)
	}
	
	if req.ProcessingMode == "ai-summary" && req.SummaryPrompt == "" {
		return fmt.Errorf("summary_prompt is required for ai-summary mode")
	}
	
	return nil
}

// wouldExceedContext checks if the given page size would exceed context limits
func (usp *UnifiedStreamProcessor) wouldExceedContext(pageSize int, streamTypeCount int, currentContextTokens int) bool {
	estimatedTokens := usp.paginationCalculator.EstimatePageTokens(pageSize, streamTypeCount)
	availableTokens := usp.config.MaxContextTokens - currentContextTokens
	
	// Add 20% buffer for safety
	return estimatedTokens > int(float64(availableTokens)*0.8)
}

// estimateStreamTokens estimates tokens for stream data
func (usp *UnifiedStreamProcessor) estimateStreamTokens(data *StravaStreams) int {
	if data == nil {
		return 0
	}
	
	dataPoints := usp.paginationCalculator.countDataPoints(data)
	streamTypeCount := usp.countStreamTypes(data)
	
	return usp.paginationCalculator.EstimatePageTokens(dataPoints, streamTypeCount)
}

// countStreamTypes counts the number of available stream types
func (usp *UnifiedStreamProcessor) countStreamTypes(data *StravaStreams) int {
	count := 0
	
	if len(data.Time) > 0 {
		count++
	}
	if len(data.Distance) > 0 {
		count++
	}
	if len(data.Heartrate) > 0 {
		count++
	}
	if len(data.Watts) > 0 {
		count++
	}
	if len(data.Cadence) > 0 {
		count++
	}
	if len(data.Altitude) > 0 {
		count++
	}
	if len(data.VelocitySmooth) > 0 {
		count++
	}
	if len(data.Temp) > 0 {
		count++
	}
	if len(data.GradeSmooth) > 0 {
		count++
	}
	if len(data.Moving) > 0 {
		count++
	}
	if len(data.Latlng) > 0 {
		count++
	}
	
	return count
}

// calculateTimeRange calculates the time range covered by the stream data
func (usp *UnifiedStreamProcessor) calculateTimeRange(data *StravaStreams) TimeRange {
	if data == nil || len(data.Time) == 0 {
		return TimeRange{}
	}
	
	return TimeRange{
		StartTime: data.Time[0],
		EndTime:   data.Time[len(data.Time)-1],
	}
}

// generatePageInstructions generates navigation instructions for the LLM
func (usp *UnifiedStreamProcessor) generatePageInstructions(req *PaginatedStreamRequest, totalPages int) string {
	var builder strings.Builder
	
	builder.WriteString(fmt.Sprintf("📄 **Page %d of %d** for Activity %d\n\n", req.PageNumber, totalPages, req.ActivityID))
	
	if req.PageNumber < totalPages {
		builder.WriteString("**Navigation:**\n")
		builder.WriteString(fmt.Sprintf("- Next page: Call get-activity-streams with page_number=%d\n", req.PageNumber+1))
		
		if req.PageNumber > 1 {
			builder.WriteString(fmt.Sprintf("- Previous page: Call get-activity-streams with page_number=%d\n", req.PageNumber-1))
		}
		
		builder.WriteString(fmt.Sprintf("- Jump to specific page: Call get-activity-streams with page_number=X (1-%d)\n", totalPages))
		builder.WriteString("- Get full dataset: Call get-activity-streams with page_size=-1\n\n")
	}
	
	builder.WriteString("**Processing Mode:** ")
	switch req.ProcessingMode {
	case "raw":
		builder.WriteString("Raw stream data points")
	case "derived":
		builder.WriteString("Derived features and statistics")
	case "ai-summary":
		builder.WriteString("AI-generated summary")
	}
	builder.WriteString("\n\n")
	
	if req.PageNumber < totalPages {
		builder.WriteString("💡 **Tip:** Analyze this page's data before requesting the next page to maintain context efficiency.")
	} else {
		builder.WriteString("✅ **Complete:** This is the final page of data for this activity.")
	}
	
	return builder.String()
}

// FormatPaginatedResult formats a paginated result with context and navigation info
func (usp *UnifiedStreamProcessor) FormatPaginatedResult(page *StreamPage) string {
	var builder strings.Builder
	
	// Add page header with navigation info
	builder.WriteString(page.Instructions)
	builder.WriteString("\n")
	
	// Add time range information
	if page.TimeRange.StartTime > 0 && page.TimeRange.EndTime > 0 {
		duration := page.TimeRange.EndTime - page.TimeRange.StartTime
		builder.WriteString(fmt.Sprintf("**Time Range:** %d-%d seconds (Duration: %d seconds)\n\n", 
			page.TimeRange.StartTime, page.TimeRange.EndTime, duration))
	}
	
	// Add the processed data
	if dataStr, ok := page.Data.(string); ok {
		builder.WriteString(dataStr)
	} else {
		builder.WriteString("**Processed Data:**\n")
		builder.WriteString(fmt.Sprintf("%+v", page.Data))
	}
	
	// Add footer with token usage info
	builder.WriteString(fmt.Sprintf("\n\n📊 **Page Stats:** %d estimated tokens", page.EstimatedTokens))
	
	if page.HasNextPage {
		builder.WriteString(fmt.Sprintf(" | Next: page %d", page.PageNumber+1))
	}
	
	return builder.String()
}

// applyProcessingModeWithFallback applies processing mode with automatic fallback on failure
func (usp *UnifiedStreamProcessor) applyProcessingModeWithFallback(user *models.User, req *PaginatedStreamRequest, streamData *StravaStreams) (interface{}, error) {
	// Try the requested processing mode first
	result, err := usp.applyProcessingMode(user, req, streamData)
	if err == nil {
		return result, nil
	}
	
	log.Printf("Processing mode %s failed for activity %d: %v", req.ProcessingMode, req.ActivityID, err)
	
	// Attempt fallback processing modes
	fallbackModes := usp.getFallbackModes(req.ProcessingMode)
	
	for _, fallbackMode := range fallbackModes {
		log.Printf("Attempting fallback to %s mode for activity %d", fallbackMode, req.ActivityID)
		
		// Create fallback request
		fallbackReq := *req
		fallbackReq.ProcessingMode = fallbackMode
		
		// Clear summary prompt for non-AI modes
		if fallbackMode != "ai-summary" {
			fallbackReq.SummaryPrompt = ""
		}
		
		result, fallbackErr := usp.applyProcessingMode(user, &fallbackReq, streamData)
		if fallbackErr == nil {
			log.Printf("Successfully fell back to %s mode for activity %d", fallbackMode, req.ActivityID)
			
			// Add fallback notice to the result
			if resultStr, ok := result.(string); ok {
				fallbackNotice := fmt.Sprintf("🔄 **Fallback Mode:** %s (original mode %s failed)\n\n", fallbackMode, req.ProcessingMode)
				result = fallbackNotice + resultStr
			}
			
			return result, nil
		}
		
		log.Printf("Fallback to %s mode also failed for activity %d: %v", fallbackMode, req.ActivityID, fallbackErr)
	}
	
	// All processing modes failed, return the original error
	return nil, err
}

// getFallbackModes returns ordered list of fallback modes for a given processing mode
func (usp *UnifiedStreamProcessor) getFallbackModes(originalMode string) []string {
	switch originalMode {
	case "ai-summary":
		return []string{"derived", "raw"}
	case "derived":
		return []string{"raw"}
	case "raw":
		return []string{"derived"} // Try derived as fallback for raw if it fails
	default:
		return []string{"raw", "derived"}
	}
}

// createFallbackStreamPage creates a fallback stream page when all processing fails
func (usp *UnifiedStreamProcessor) createFallbackStreamPage(req *PaginatedStreamRequest, streamData *StravaStreams, originalErr error) *StreamPage {
	log.Printf("Creating fallback stream page for activity %d", req.ActivityID)
	
	// Create a basic stream processor for fallback formatting
	streamProcessor := &streamProcessor{config: usp.config}
	
	// Generate fallback content
	fallbackContent := streamProcessor.CreateFallbackFormatter(streamData, req.ActivityID, req.ProcessingMode)
	
	// Add error information
	var errorInfo strings.Builder
	errorInfo.WriteString("⚠️ **Processing Error**\n")
	errorInfo.WriteString(fmt.Sprintf("Original error: %v\n\n", originalErr))
	
	// Combine error info with fallback content
	finalContent := errorInfo.String() + fallbackContent
	
	// Calculate basic time range
	timeRange := usp.calculateTimeRange(streamData)
	
	// Create fallback stream page
	return &StreamPage{
		ActivityID:      req.ActivityID,
		PageNumber:      req.PageNumber,
		TotalPages:      1, // Fallback is always single page
		ProcessingMode:  "fallback",
		Data:            finalContent,
		TimeRange:       timeRange,
		Instructions:    "Fallback processing applied due to processing errors. Basic stream information provided.",
		HasNextPage:     false,
		EstimatedTokens: usp.estimateStreamTokens(streamData),
	}
}

// Enhanced error handling for processing modes
func (usp *UnifiedStreamProcessor) handleProcessingModeError(err error, req *PaginatedStreamRequest, context string) error {
	// Create detailed stream processing error
	var streamErr *StreamProcessingError
	
	// Check if it's already a stream processing error
	if errors.As(err, &streamErr) {
		// Add additional context
		streamErr.WithContext("processing_context", context)
		return streamErr
	}
	
	// Create new stream processing error
	errorType := "processing_failure"
	
	// Determine specific error type based on error message
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "strava") || strings.Contains(errMsg, "API"):
		errorType = "strava_api_failure"
	case strings.Contains(errMsg, "context") || strings.Contains(errMsg, "token"):
		errorType = "context_exceeded"
	case strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "required"):
		errorType = "invalid_request"
	case strings.Contains(errMsg, "corrupt") || strings.Contains(errMsg, "incomplete"):
		errorType = "data_corrupted"
	case strings.Contains(errMsg, "unavailable") || strings.Contains(errMsg, "not available"):
		errorType = "processor_unavailable"
	}
	
	streamErr = NewStreamProcessingError(errorType, err.Error(), req.ActivityID, req.ProcessingMode).
		WithOriginalError(err).
		WithContext("processing_context", context)
	
	return streamErr
}