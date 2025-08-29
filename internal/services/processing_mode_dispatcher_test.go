package services

import (
	"context"
	"testing"

	"bodda/internal/config"
)

// mockSummaryProcessor is a simple mock for testing
type mockSummaryProcessor struct{}

func (m *mockSummaryProcessor) GenerateSummary(ctx context.Context, data *StravaStreams, activityID int64, prompt string) (*StreamSummary, error) {
	return &StreamSummary{
		ActivityID:    activityID,
		SummaryPrompt: prompt,
		Summary:       "Mock AI summary of the stream data",
		Model:         "mock-model",
	}, nil
}

func (m *mockSummaryProcessor) PrepareStreamDataForSummarization(data *StravaStreams) (string, error) {
	return "Mock stream data preparation", nil
}

func TestNewProcessingModeDispatcher(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  10000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1000,
			MaxPageSize:       5000,
			RedactionEnabled:  true,
		},
	}

	streamProcessor := NewStreamProcessor(cfg)
	summaryProcessor := &mockSummaryProcessor{}
	dispatcher := NewProcessingModeDispatcher(streamProcessor, summaryProcessor)

	if dispatcher == nil {
		t.Fatal("Expected dispatcher to be created, got nil")
	}

	// Test that it implements the interface
	var _ ProcessingModeDispatcher = dispatcher
}

func TestProcessingModeDispatcher_ValidateMode(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  10000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1000,
			MaxPageSize:       5000,
			RedactionEnabled:  true,
		},
	}

	streamProcessor := NewStreamProcessor(cfg)
	summaryProcessor := &mockSummaryProcessor{}
	dispatcher := NewProcessingModeDispatcher(streamProcessor, summaryProcessor)

	tests := []struct {
		name        string
		mode        string
		expectError bool
	}{
		{
			name:        "valid mode - auto",
			mode:        "auto",
			expectError: false,
		},
		{
			name:        "valid mode - raw",
			mode:        "raw",
			expectError: false,
		},
		{
			name:        "valid mode - derived",
			mode:        "derived",
			expectError: false,
		},
		{
			name:        "valid mode - ai-summary",
			mode:        "ai-summary",
			expectError: false,
		},
		{
			name:        "invalid mode",
			mode:        "invalid",
			expectError: true,
		},
		{
			name:        "empty mode",
			mode:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dispatcher.ValidateMode(tt.mode)
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestProcessingModeDispatcher_GetSupportedModes(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  10000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1000,
			MaxPageSize:       5000,
			RedactionEnabled:  true,
		},
	}

	streamProcessor := NewStreamProcessor(cfg)
	summaryProcessor := &mockSummaryProcessor{}
	dispatcher := NewProcessingModeDispatcher(streamProcessor, summaryProcessor)

	modes := dispatcher.GetSupportedModes()
	expectedModes := []string{"auto", "raw", "derived", "ai-summary"}

	if len(modes) != len(expectedModes) {
		t.Errorf("Expected %d modes, got %d", len(expectedModes), len(modes))
	}

	for i, mode := range modes {
		if mode != expectedModes[i] {
			t.Errorf("Expected mode %s at index %d, got %s", expectedModes[i], i, mode)
		}
	}
}

func TestProcessingModeDispatcher_Dispatch(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  1000, // Low threshold for testing
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   500,
			MaxPageSize:       2000,
			RedactionEnabled:  true,
		},
	}

	streamProcessor := NewStreamProcessor(cfg)
	summaryProcessor := &mockSummaryProcessor{}
	dispatcher := NewProcessingModeDispatcher(streamProcessor, summaryProcessor)

	smallData := &StravaStreams{
		Time:      []int{0, 1, 2},
		Heartrate: []int{120, 125, 130},
	}

	largeData := createLargeStreamData(2000)

	tests := []struct {
		name           string
		mode           string
		data           *StravaStreams
		params         ProcessingParams
		expectError    bool
		expectedMode   string
	}{
		{
			name: "invalid mode",
			mode: "invalid",
			data: smallData,
			params: ProcessingParams{
				ToolCallID: "test-1",
				ActivityID: 123,
			},
			expectError: true,
		},
		{
			name: "auto mode - small data",
			mode: "auto",
			data: smallData,
			params: ProcessingParams{
				ToolCallID: "test-2",
				ActivityID: 123,
			},
			expectError:  false,
			expectedMode: "raw",
		},
		{
			name: "auto mode - large data",
			mode: "auto",
			data: largeData,
			params: ProcessingParams{
				ToolCallID: "test-3",
				ActivityID: 123,
			},
			expectError:  false,
			expectedMode: "auto",
		},
		{
			name: "raw mode",
			mode: "raw",
			data: smallData,
			params: ProcessingParams{
				ToolCallID: "test-4",
				ActivityID: 123,
			},
			expectError:  false,
			expectedMode: "raw",
		},
		{
			name: "derived mode",
			mode: "derived",
			data: smallData,
			params: ProcessingParams{
				ToolCallID: "test-5",
				ActivityID: 123,
			},
			expectError:  false,
			expectedMode: "derived",
		},
		{
			name: "ai-summary mode with prompt",
			mode: "ai-summary",
			data: smallData,
			params: ProcessingParams{
				ToolCallID:    "test-6",
				ActivityID:    123,
				SummaryPrompt: "Analyze this workout",
			},
			expectError:  false,
			expectedMode: "ai-summary",
		},
		{
			name: "ai-summary mode without prompt",
			mode: "ai-summary",
			data: smallData,
			params: ProcessingParams{
				ToolCallID: "test-7",
				ActivityID: 123,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dispatcher.Dispatch(tt.mode, tt.data, tt.params)

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

			if result.ToolCallID != tt.params.ToolCallID {
				t.Errorf("Expected ToolCallID %s, got %s", tt.params.ToolCallID, result.ToolCallID)
			}

			if result.ProcessingMode != tt.expectedMode {
				t.Errorf("Expected ProcessingMode %s, got %s", tt.expectedMode, result.ProcessingMode)
			}

			if result.Content == "" {
				t.Error("Expected non-empty content")
			}
		})
	}
}