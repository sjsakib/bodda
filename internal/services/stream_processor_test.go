package services

import (
	"testing"

	"bodda/internal/config"
)

func TestNewStreamProcessor(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  10000,
			TokenPerCharRatio: 0.3,
			DefaultPageSize:   500,
			MaxPageSize:       2000,
			RedactionEnabled:  false,
		},
	}

	processor := NewStreamProcessor(cfg)
	if processor == nil {
		t.Fatal("Expected stream processor to be created, got nil")
	}

	// Test that it implements the interface
	var _ StreamProcessor = processor
}

func TestStreamProcessor_ShouldProcess(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  1000, // Low threshold for testing
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   500,
			MaxPageSize:       2000,
			RedactionEnabled:  true,
		},
	}

	processor := NewStreamProcessor(cfg)

	tests := []struct {
		name     string
		data     *StravaStreams
		expected bool
	}{
		{
			name:     "nil data",
			data:     nil,
			expected: false,
		},
		{
			name:     "empty data",
			data:     &StravaStreams{},
			expected: false,
		},
		{
			name: "small data set",
			data: &StravaStreams{
				Time:      []int{0, 1, 2, 3, 4},
				Heartrate: []int{120, 125, 130, 135, 140},
			},
			expected: false,
		},
		{
			name: "large data set",
			data: createLargeStreamData(5000), // Should exceed token limit
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.ShouldProcess(tt.data)
			if result != tt.expected {
				t.Errorf("ShouldProcess() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestStreamProcessor_EstimateTokens(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  10000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1000,
			MaxPageSize:       5000,
			RedactionEnabled:  true,
		},
	}

	processor := NewStreamProcessor(cfg)

	tests := []struct {
		name     string
		data     *StravaStreams
		expected int
	}{
		{
			name:     "nil data",
			data:     nil,
			expected: 0,
		},
		{
			name:     "empty data",
			data:     &StravaStreams{},
			expected: 0, // Should be very small
		},
		{
			name: "small data set",
			data: &StravaStreams{
				Time:      []int{0, 1, 2},
				Heartrate: []int{120, 125, 130},
			},
			expected: 0, // Should be small but > 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.EstimateTokens(tt.data)
			if tt.name == "nil data" && result != tt.expected {
				t.Errorf("EstimateTokens() = %v, expected %v", result, tt.expected)
			} else if tt.name != "nil data" && result < 0 {
				t.Errorf("EstimateTokens() = %v, expected non-negative value", result)
			}
		})
	}
}

func TestStreamProcessor_GetProcessingOptions(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  10000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1000,
			MaxPageSize:       5000,
			RedactionEnabled:  true,
		},
	}

	processor := NewStreamProcessor(cfg)
	options := processor.GetProcessingOptions()

	expectedModes := []string{"raw", "derived", "ai-summary", "auto"}
	if len(options) != len(expectedModes) {
		t.Errorf("Expected %d options, got %d", len(expectedModes), len(options))
	}

	for i, option := range options {
		if option.Mode != expectedModes[i] {
			t.Errorf("Expected mode %s at index %d, got %s", expectedModes[i], i, option.Mode)
		}
		if option.Description == "" {
			t.Errorf("Expected non-empty description for mode %s", option.Mode)
		}
		if option.Command == "" {
			t.Errorf("Expected non-empty command for mode %s", option.Mode)
		}
	}
}

func TestStreamProcessor_ProcessStreamOutput(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  1000, // Low threshold for testing
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   500,
			MaxPageSize:       2000,
			RedactionEnabled:  true,
		},
	}

	processor := NewStreamProcessor(cfg)

	tests := []struct {
		name           string
		data           *StravaStreams
		toolCallID     string
		expectError    bool
		expectedMode   string
		expectOptions  bool
	}{
		{
			name:        "nil data",
			data:        nil,
			toolCallID:  "test-1",
			expectError: true,
		},
		{
			name: "small data - raw mode",
			data: &StravaStreams{
				Time:      []int{0, 1, 2},
				Heartrate: []int{120, 125, 130},
			},
			toolCallID:    "test-2",
			expectError:   false,
			expectedMode:  "raw",
			expectOptions: false,
		},
		{
			name:          "large data - auto mode with options",
			data:          createLargeStreamData(2000),
			toolCallID:    "test-3",
			expectError:   false,
			expectedMode:  "auto",
			expectOptions: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.ProcessStreamOutput(tt.data, tt.toolCallID)

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

			if result.ToolCallID != tt.toolCallID {
				t.Errorf("Expected ToolCallID %s, got %s", tt.toolCallID, result.ToolCallID)
			}

			if result.ProcessingMode != tt.expectedMode {
				t.Errorf("Expected ProcessingMode %s, got %s", tt.expectedMode, result.ProcessingMode)
			}

			if tt.expectOptions && len(result.Options) == 0 {
				t.Error("Expected processing options, got none")
			}

			if !tt.expectOptions && len(result.Options) > 0 {
				t.Error("Expected no processing options, got some")
			}

			if result.Content == "" {
				t.Error("Expected non-empty content")
			}
		})
	}
}

// Helper function to create large stream data for testing
func createLargeStreamData(size int) *StravaStreams {
	time := make([]int, size)
	heartrate := make([]int, size)
	watts := make([]int, size)
	distance := make([]float64, size)
	altitude := make([]float64, size)

	for i := 0; i < size; i++ {
		time[i] = i
		heartrate[i] = 120 + (i % 60)
		watts[i] = 100 + (i % 200)
		distance[i] = float64(i) * 10.5
		altitude[i] = 100.0 + float64(i%100)
	}

	return &StravaStreams{
		Time:      time,
		Heartrate: heartrate,
		Watts:     watts,
		Distance:  distance,
		Altitude:  altitude,
	}
}