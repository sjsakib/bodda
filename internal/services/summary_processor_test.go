package services

import (
	"context"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestNewSummaryProcessor(t *testing.T) {
	client := openai.NewClient("test-api-key")
	processor := NewSummaryProcessor(client)

	if processor == nil {
		t.Fatal("Expected summary processor to be created, got nil")
	}

	// Test that it implements the interface
	var _ SummaryProcessor = processor
}

func TestSummaryProcessor_PrepareStreamDataForSummarization(t *testing.T) {
	client := openai.NewClient("test-api-key")
	processor := NewSummaryProcessor(client)

	tests := []struct {
		name        string
		data        *StravaStreams
		expectError bool
	}{
		{
			name:        "nil data",
			data:        nil,
			expectError: true,
		},
		{
			name: "empty data",
			data: &StravaStreams{},
			expectError: false,
		},
		{
			name: "basic stream data",
			data: &StravaStreams{
				Time:      []int{0, 1, 2, 3, 4},
				Heartrate: []int{120, 125, 130, 135, 140},
				Watts:     []int{100, 110, 120, 130, 140},
			},
			expectError: false,
		},
		{
			name: "comprehensive stream data",
			data: &StravaStreams{
				Time:           []int{0, 1, 2, 3, 4},
				Distance:       []float64{0, 10, 20, 30, 40},
				Heartrate:      []int{120, 125, 130, 135, 140},
				Watts:          []int{100, 110, 120, 130, 140},
				Cadence:        []int{80, 85, 90, 95, 100},
				Altitude:       []float64{100, 105, 110, 115, 120},
				VelocitySmooth: []float64{5.0, 5.5, 6.0, 6.5, 7.0},
				Temp:           []int{20, 21, 22, 23, 24},
				GradeSmooth:    []float64{0.01, 0.02, 0.03, 0.04, 0.05},
				Moving:         []bool{true, true, true, true, true},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.PrepareStreamDataForSummarization(tt.data)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == "" {
				t.Error("Expected non-empty result")
			}

			// Check that result contains expected sections
			if tt.data != nil && len(tt.data.Time) > 0 {
				if !contains(result, "STREAM DATA SUMMARY") {
					t.Error("Expected result to contain 'STREAM DATA SUMMARY'")
				}
				if !contains(result, "SAMPLE DATA POINTS") {
					t.Error("Expected result to contain 'SAMPLE DATA POINTS'")
				}
			}
		})
	}
}

func TestSummaryProcessor_GenerateSummary_ValidationErrors(t *testing.T) {
	client := openai.NewClient("test-api-key")
	processor := NewSummaryProcessor(client)
	ctx := context.Background()

	tests := []struct {
		name        string
		data        *StravaStreams
		activityID  int64
		prompt      string
		expectError bool
	}{
		{
			name:        "nil data",
			data:        nil,
			activityID:  123,
			prompt:      "test prompt",
			expectError: true,
		},
		{
			name: "empty prompt",
			data: &StravaStreams{
				Time:      []int{0, 1, 2},
				Heartrate: []int{120, 125, 130},
			},
			activityID:  123,
			prompt:      "",
			expectError: true,
		},
		{
			name: "valid inputs but no API call",
			data: &StravaStreams{
				Time:      []int{0, 1, 2},
				Heartrate: []int{120, 125, 130},
			},
			activityID:  123,
			prompt:      "test prompt",
			expectError: true, // Will fail due to invalid API key, but that's expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processor.GenerateSummary(ctx, tt.data, tt.activityID, tt.prompt)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}