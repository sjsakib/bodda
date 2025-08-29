package services

import (
	"testing"
)

func TestPaginationCalculator_CalculateOptimalPageSize(t *testing.T) {
	config := &StreamConfig{
		MaxContextTokens:  15000,
		TokenPerCharRatio: 0.25,
		DefaultPageSize:   1000,
		MaxPageSize:       5000,
	}

	pc := NewPaginationCalculator(config, nil)

	tests := []struct {
		name                string
		currentContextTokens int
		expectedMinSize     int
		expectedMaxSize     int
	}{
		{
			name:                "High available context",
			currentContextTokens: 5000,
			expectedMinSize:     1000,
			expectedMaxSize:     5000,
		},
		{
			name:                "Low available context",
			currentContextTokens: 14000,
			expectedMinSize:     100,
			expectedMaxSize:     1000,
		},
		{
			name:                "Very low available context",
			currentContextTokens: 14900,
			expectedMinSize:     100,
			expectedMaxSize:     200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageSize := pc.CalculateOptimalPageSize(tt.currentContextTokens)
			
			if pageSize < tt.expectedMinSize {
				t.Errorf("CalculateOptimalPageSize() = %d, expected >= %d", pageSize, tt.expectedMinSize)
			}
			
			if pageSize > tt.expectedMaxSize {
				t.Errorf("CalculateOptimalPageSize() = %d, expected <= %d", pageSize, tt.expectedMaxSize)
			}
		})
	}
}

func TestPaginationCalculator_SliceStreams(t *testing.T) {
	config := &StreamConfig{
		MaxContextTokens:  15000,
		TokenPerCharRatio: 0.25,
		DefaultPageSize:   1000,
		MaxPageSize:       5000,
	}

	pc := NewPaginationCalculator(config, nil)

	// Create test stream data
	streams := &StravaStreams{
		Time:      []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		Heartrate: []int{120, 125, 130, 135, 140, 145, 150, 155, 160, 165},
		Watts:     []int{100, 110, 120, 130, 140, 150, 160, 170, 180, 190},
		Distance:  []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90},
	}

	tests := []struct {
		name       string
		startIndex int
		endIndex   int
		expectedLen int
	}{
		{
			name:       "First half",
			startIndex: 0,
			endIndex:   5,
			expectedLen: 5,
		},
		{
			name:       "Second half",
			startIndex: 5,
			endIndex:   10,
			expectedLen: 5,
		},
		{
			name:       "Middle section",
			startIndex: 2,
			endIndex:   7,
			expectedLen: 5,
		},
		{
			name:       "Single element",
			startIndex: 3,
			endIndex:   4,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sliced := pc.sliceStreams(streams, tt.startIndex, tt.endIndex)
			
			if sliced == nil {
				t.Fatal("sliceStreams() returned nil")
			}
			
			if len(sliced.Time) != tt.expectedLen {
				t.Errorf("sliceStreams() Time length = %d, expected %d", len(sliced.Time), tt.expectedLen)
			}
			
			if len(sliced.Heartrate) != tt.expectedLen {
				t.Errorf("sliceStreams() Heartrate length = %d, expected %d", len(sliced.Heartrate), tt.expectedLen)
			}
			
			if len(sliced.Watts) != tt.expectedLen {
				t.Errorf("sliceStreams() Watts length = %d, expected %d", len(sliced.Watts), tt.expectedLen)
			}
			
			if len(sliced.Distance) != tt.expectedLen {
				t.Errorf("sliceStreams() Distance length = %d, expected %d", len(sliced.Distance), tt.expectedLen)
			}
			
			// Verify data integrity
			if len(sliced.Time) > 0 {
				expectedStartTime := streams.Time[tt.startIndex]
				if sliced.Time[0] != expectedStartTime {
					t.Errorf("sliceStreams() first Time = %d, expected %d", sliced.Time[0], expectedStartTime)
				}
			}
		})
	}
}

func TestPaginationCalculator_EstimatePageTokens(t *testing.T) {
	config := &StreamConfig{
		MaxContextTokens:  15000,
		TokenPerCharRatio: 0.25,
		DefaultPageSize:   1000,
		MaxPageSize:       5000,
	}

	pc := NewPaginationCalculator(config, nil)

	tests := []struct {
		name            string
		pageSize        int
		streamTypeCount int
		expectedMin     int
		expectedMax     int
	}{
		{
			name:            "Small page, few streams",
			pageSize:        100,
			streamTypeCount: 3,
			expectedMin:     200,
			expectedMax:     500,
		},
		{
			name:            "Large page, many streams",
			pageSize:        1000,
			streamTypeCount: 8,
			expectedMin:     5000,
			expectedMax:     15000,
		},
		{
			name:            "Medium page, medium streams",
			pageSize:        500,
			streamTypeCount: 5,
			expectedMin:     1000,
			expectedMax:     5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := pc.EstimatePageTokens(tt.pageSize, tt.streamTypeCount)
			
			if tokens < tt.expectedMin {
				t.Errorf("EstimatePageTokens() = %d, expected >= %d", tokens, tt.expectedMin)
			}
			
			if tokens > tt.expectedMax {
				t.Errorf("EstimatePageTokens() = %d, expected <= %d", tokens, tt.expectedMax)
			}
		})
	}
}

func TestPaginationCalculator_CountDataPoints(t *testing.T) {
	config := &StreamConfig{
		MaxContextTokens:  15000,
		TokenPerCharRatio: 0.25,
		DefaultPageSize:   1000,
		MaxPageSize:       5000,
	}

	pc := NewPaginationCalculator(config, nil)

	tests := []struct {
		name     string
		streams  *StravaStreams
		expected int
	}{
		{
			name:     "Nil streams",
			streams:  nil,
			expected: 0,
		},
		{
			name:     "Empty streams",
			streams:  &StravaStreams{},
			expected: 0,
		},
		{
			name: "Single stream type",
			streams: &StravaStreams{
				Time: []int{0, 1, 2, 3, 4},
			},
			expected: 5,
		},
		{
			name: "Multiple stream types, same length",
			streams: &StravaStreams{
				Time:      []int{0, 1, 2, 3, 4},
				Heartrate: []int{120, 125, 130, 135, 140},
				Watts:     []int{100, 110, 120, 130, 140},
			},
			expected: 5,
		},
		{
			name: "Multiple stream types, different lengths",
			streams: &StravaStreams{
				Time:      []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, // 10 points
				Heartrate: []int{120, 125, 130, 135, 140},        // 5 points
				Watts:     []int{100, 110, 120},                  // 3 points
			},
			expected: 10, // Should return the maximum
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := pc.countDataPoints(tt.streams)
			
			if count != tt.expected {
				t.Errorf("countDataPoints() = %d, expected %d", count, tt.expected)
			}
		})
	}
}

func TestPaginatedStreamRequest_Validation(t *testing.T) {
	// This test would be part of the unified stream processor tests
	// but we can test the basic structure here
	
	req := &PaginatedStreamRequest{
		ActivityID:     12345,
		StreamTypes:    []string{"time", "heartrate", "watts"},
		Resolution:     "medium",
		ProcessingMode: "raw",
		PageNumber:     1,
		PageSize:       1000,
	}

	// Basic validation checks
	if req.ActivityID <= 0 {
		t.Error("ActivityID should be positive")
	}
	
	if len(req.StreamTypes) == 0 {
		t.Error("StreamTypes should not be empty")
	}
	
	if req.PageNumber < 1 {
		t.Error("PageNumber should be >= 1")
	}
	
	validModes := map[string]bool{
		"raw":        true,
		"derived":    true,
		"ai-summary": true,
	}
	
	if !validModes[req.ProcessingMode] {
		t.Errorf("ProcessingMode %s is not valid", req.ProcessingMode)
	}
}

func TestStreamPage_Structure(t *testing.T) {
	// Test the StreamPage structure
	page := &StreamPage{
		ActivityID:      12345,
		PageNumber:      1,
		TotalPages:      5,
		ProcessingMode:  "raw",
		Data:            "test data",
		TimeRange:       TimeRange{StartTime: 0, EndTime: 100},
		Instructions:    "test instructions",
		HasNextPage:     true,
		EstimatedTokens: 1500,
	}

	if page.ActivityID != 12345 {
		t.Error("ActivityID not set correctly")
	}
	
	if page.PageNumber != 1 {
		t.Error("PageNumber not set correctly")
	}
	
	if page.TotalPages != 5 {
		t.Error("TotalPages not set correctly")
	}
	
	if !page.HasNextPage {
		t.Error("HasNextPage should be true")
	}
	
	if page.TimeRange.StartTime != 0 || page.TimeRange.EndTime != 100 {
		t.Error("TimeRange not set correctly")
	}
}