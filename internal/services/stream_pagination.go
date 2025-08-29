package services

import (
	"fmt"
	"log"
	"math"

	"bodda/internal/models"
)

// PaginatedStreamRequest represents a request for paginated stream data
type PaginatedStreamRequest struct {
	ActivityID     int64    `json:"activity_id"`
	StreamTypes    []string `json:"stream_types"`
	Resolution     string   `json:"resolution"`
	ProcessingMode string   `json:"processing_mode"` // "raw", "derived", "ai-summary"
	PageNumber     int      `json:"page_number"`
	PageSize       int      `json:"page_size"`
	SummaryPrompt  string   `json:"summary_prompt,omitempty"`
}

// StreamPage represents a page of stream data with navigation info
type StreamPage struct {
	ActivityID     int64       `json:"activity_id"`
	PageNumber     int         `json:"page_number"`
	TotalPages     int         `json:"total_pages"`
	ProcessingMode string      `json:"processing_mode"`
	Data           interface{} `json:"data"` // Raw streams, derived features, or summary
	TimeRange      TimeRange   `json:"time_range"`
	Instructions   string      `json:"instructions"`
	HasNextPage    bool        `json:"has_next_page"`
	EstimatedTokens int        `json:"estimated_tokens"`
}

// TimeRange represents the time range covered by a page
type TimeRange struct {
	StartTime int `json:"start_time"`
	EndTime   int `json:"end_time"`
}

// PaginationCalculator handles pagination calculations and estimates
type PaginationCalculator struct {
	config         *StreamConfig
	stravaService  StravaService
}

// NewPaginationCalculator creates a new pagination calculator
func NewPaginationCalculator(config *StreamConfig, stravaService StravaService) *PaginationCalculator {
	return &PaginationCalculator{
		config:        config,
		stravaService: stravaService,
	}
}

// CalculateOptimalPageSize calculates the optimal page size based on available context
func (pc *PaginationCalculator) CalculateOptimalPageSize(currentContextTokens int) int {
	// Calculate available tokens for stream data
	availableTokens := pc.config.MaxContextTokens - currentContextTokens
	
	// Reserve some tokens for formatting and metadata (20% buffer)
	usableTokens := int(float64(availableTokens) * 0.8)
	
	// Estimate data points that fit in available tokens
	// Rough estimate: each data point uses about 4 characters when serialized
	charsPerDataPoint := 4.0
	tokensPerDataPoint := charsPerDataPoint * pc.config.TokenPerCharRatio
	
	optimalPageSize := int(float64(usableTokens) / tokensPerDataPoint)
	
	// Apply bounds
	if optimalPageSize < 100 {
		optimalPageSize = 100 // Minimum useful page size
	}
	if optimalPageSize > pc.config.MaxPageSize {
		optimalPageSize = pc.config.MaxPageSize
	}
	
	log.Printf("Calculated optimal page size: %d (available tokens: %d, usable: %d)", 
		optimalPageSize, availableTokens, usableTokens)
	
	return optimalPageSize
}

// EstimateTotalPages estimates the total number of pages for an activity
func (pc *PaginationCalculator) EstimateTotalPages(user *models.User, activityID int64, streamTypes []string, resolution string, pageSize int) (int, error) {
	// Handle negative page size (full dataset request)
	if pageSize < 0 {
		return 1, nil
	}
	
	// Get a sample of the stream data to estimate total size
	// Use low resolution for estimation to minimize API calls
	estimationResolution := "low"
	if resolution == "low" {
		estimationResolution = resolution
	}
	
	sampleStreams, err := pc.stravaService.GetActivityStreams(user, activityID, streamTypes, estimationResolution)
	if err != nil {
		return 0, fmt.Errorf("failed to get sample streams for estimation: %w", err)
	}
	
	sampleDataPoints := pc.countDataPoints(sampleStreams)
	if sampleDataPoints == 0 {
		return 1, nil
	}
	
	// Estimate full dataset size based on resolution multipliers
	var fullDataPoints int
	switch resolution {
	case "low":
		fullDataPoints = sampleDataPoints
	case "medium":
		// Medium resolution typically has 2-3x more data points than low
		fullDataPoints = sampleDataPoints * 3
	case "high":
		// High resolution typically has 5-10x more data points than low
		fullDataPoints = sampleDataPoints * 8
	default:
		// Default to medium resolution estimate
		fullDataPoints = sampleDataPoints * 3
	}
	
	// Calculate total pages
	totalPages := int(math.Ceil(float64(fullDataPoints) / float64(pageSize)))
	
	log.Printf("Estimated total pages for activity %d: %d (sample points: %d, estimated full: %d, page size: %d)", 
		activityID, totalPages, sampleDataPoints, fullDataPoints, pageSize)
	
	return totalPages, nil
}

// RequestSpecificDataChunk requests a specific chunk of stream data from Strava API
func (pc *PaginationCalculator) RequestSpecificDataChunk(user *models.User, req *PaginatedStreamRequest) (*StravaStreams, error) {
	// Handle negative page size (full dataset request)
	if req.PageSize < 0 {
		log.Printf("Requesting full dataset for activity %d", req.ActivityID)
		return pc.stravaService.GetActivityStreams(user, req.ActivityID, req.StreamTypes, req.Resolution)
	}
	
	// For positive page sizes, we need to implement chunking
	// Since Strava API doesn't support direct pagination, we'll get the full dataset
	// and then slice it to the requested page
	fullStreams, err := pc.stravaService.GetActivityStreams(user, req.ActivityID, req.StreamTypes, req.Resolution)
	if err != nil {
		return nil, fmt.Errorf("failed to get full stream data: %w", err)
	}
	
	// Calculate the slice boundaries for the requested page
	startIndex := (req.PageNumber - 1) * req.PageSize
	endIndex := startIndex + req.PageSize
	
	// Slice the streams to the requested page
	pagedStreams := pc.sliceStreams(fullStreams, startIndex, endIndex)
	
	log.Printf("Requested data chunk for activity %d, page %d (indices %d-%d)", 
		req.ActivityID, req.PageNumber, startIndex, endIndex)
	
	return pagedStreams, nil
}

// sliceStreams slices stream data to a specific range
func (pc *PaginationCalculator) sliceStreams(streams *StravaStreams, startIndex, endIndex int) *StravaStreams {
	if streams == nil {
		return nil
	}
	
	result := &StravaStreams{}
	
	// Helper function to safely slice integer arrays
	sliceIntArray := func(data []int, start, end int) []int {
		if len(data) == 0 {
			return nil
		}
		if start >= len(data) {
			return nil
		}
		if end > len(data) {
			end = len(data)
		}
		return data[start:end]
	}
	
	// Helper function to safely slice float arrays
	sliceFloatArray := func(data []float64, start, end int) []float64 {
		if len(data) == 0 {
			return nil
		}
		if start >= len(data) {
			return nil
		}
		if end > len(data) {
			end = len(data)
		}
		return data[start:end]
	}
	
	// Helper function to safely slice boolean arrays
	sliceBoolArray := func(data []bool, start, end int) []bool {
		if len(data) == 0 {
			return nil
		}
		if start >= len(data) {
			return nil
		}
		if end > len(data) {
			end = len(data)
		}
		return data[start:end]
	}
	
	// Helper function to safely slice 2D float arrays (for latlng)
	sliceLatLngArray := func(data [][]float64, start, end int) [][]float64 {
		if len(data) == 0 {
			return nil
		}
		if start >= len(data) {
			return nil
		}
		if end > len(data) {
			end = len(data)
		}
		return data[start:end]
	}
	
	// Slice each stream type
	result.Time = sliceIntArray(streams.Time, startIndex, endIndex)
	result.Distance = sliceFloatArray(streams.Distance, startIndex, endIndex)
	result.Latlng = sliceLatLngArray(streams.Latlng, startIndex, endIndex)
	result.Altitude = sliceFloatArray(streams.Altitude, startIndex, endIndex)
	result.VelocitySmooth = sliceFloatArray(streams.VelocitySmooth, startIndex, endIndex)
	result.Heartrate = sliceIntArray(streams.Heartrate, startIndex, endIndex)
	result.Cadence = sliceIntArray(streams.Cadence, startIndex, endIndex)
	result.Watts = sliceIntArray(streams.Watts, startIndex, endIndex)
	result.Temp = sliceIntArray(streams.Temp, startIndex, endIndex)
	result.Moving = sliceBoolArray(streams.Moving, startIndex, endIndex)
	result.GradeSmooth = sliceFloatArray(streams.GradeSmooth, startIndex, endIndex)
	
	return result
}

// countDataPoints counts the maximum number of data points across all streams
func (pc *PaginationCalculator) countDataPoints(data *StravaStreams) int {
	if data == nil {
		return 0
	}
	
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

// EstimatePageTokens estimates the token count for a specific page
func (pc *PaginationCalculator) EstimatePageTokens(pageSize int, streamTypeCount int) int {
	// Rough estimate: each data point uses about 4 characters per stream type when serialized
	charsPerDataPoint := 4.0 * float64(streamTypeCount)
	totalChars := charsPerDataPoint * float64(pageSize)
	
	// Add overhead for JSON structure and formatting
	totalChars *= 1.2
	
	estimatedTokens := int(totalChars * pc.config.TokenPerCharRatio)
	
	return estimatedTokens
}

// GetTokenUsageEstimates returns token usage estimates for different page sizes
func (pc *PaginationCalculator) GetTokenUsageEstimates(streamTypeCount int) map[int]int {
	pageSizes := []int{500, 1000, 2000, 5000}
	estimates := make(map[int]int)
	
	for _, pageSize := range pageSizes {
		estimates[pageSize] = pc.EstimatePageTokens(pageSize, streamTypeCount)
	}
	
	return estimates
}