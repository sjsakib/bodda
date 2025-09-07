package services

import (
	"context"
	"errors"
	"testing"

	"bodda/internal/config"
)

// Mock failing summary processor for testing fallback logic
type mockFailingSummaryProcessor struct {
	shouldFail bool
}

func (m *mockFailingSummaryProcessor) GenerateSummary(ctx context.Context, data *StravaStreams, activityID int64, prompt string) (*StreamSummary, error) {
	if m.shouldFail {
		return nil, errors.New("mock AI summary failure")
	}
	
	return &StreamSummary{
		ActivityID: activityID,
		Summary:    "Mock AI summary",
	}, nil
}

func (m *mockFailingSummaryProcessor) GenerateSummaryWithResponsesAPI(ctx context.Context, data *StravaStreams, activityID int64, prompt string) (*StreamSummary, error) {
	if m.shouldFail {
		return nil, errors.New("mock AI summary failure with Responses API")
	}
	return &StreamSummary{
		ActivityID: activityID, SummaryPrompt: prompt, Summary: "Mock summary with Responses API", Model: "mock-model",
	}, nil
}

func (m *mockFailingSummaryProcessor) ProcessStreamDataWithResponsesAPI(ctx context.Context, data *StravaStreams, activityID int64, prompt string) (*StreamSummary, error) {
	if m.shouldFail {
		return nil, errors.New("mock stream processing failure with Responses API")
	}
	return &StreamSummary{
		ActivityID: activityID, SummaryPrompt: prompt, Summary: "Mock processed data with Responses API", Model: "mock-model",
	}, nil
}

func (m *mockFailingSummaryProcessor) PrepareStreamDataForSummarization(data *StravaStreams) (string, error) {
	return "prepared data", nil
}

func TestProcessingModeDispatcherFallbackLogic(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  15000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1000,
			MaxPageSize:       5000,
			RedactionEnabled:  true,
		},
	}

	streamProcessor := NewStreamProcessor(cfg)
	mockSummary := &mockFailingSummaryProcessor{shouldFail: true}
	
	dispatcher := NewProcessingModeDispatcher(streamProcessor, mockSummary)

	data := &StravaStreams{
		Time:      []int{0, 60, 120, 180, 240},
		Heartrate: []int{120, 125, 130, 135, 140},
	}

	params := ProcessingParams{
		ToolCallID:      "test_call",
		ActivityID:      12345,
		ProcessingMode:  "ai-summary",
		SummaryPrompt:   "Test prompt",
	}

	// Test AI summary failure with automatic fallback
	result, err := dispatcher.Dispatch("ai-summary", data, params)

	// Should succeed with fallback
	if err != nil {
		t.Errorf("Expected success with fallback, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Should indicate fallback was used
	if result.ProcessingMode != "derived-fallback" {
		t.Errorf("Expected processing mode 'derived-fallback', got '%s'", result.ProcessingMode)
	}

	// Should contain fallback notice
	if !containsString(result.Content, "Fallback Mode") {
		t.Error("Expected result to contain fallback notice")
	}
}

func TestProcessingModeDispatcherEmergencyFallback(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  15000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1000,
			MaxPageSize:       5000,
			RedactionEnabled:  true,
		},
	}

	streamProcessor := NewStreamProcessor(cfg)
	mockSummary := &mockFailingSummaryProcessor{shouldFail: true}
	
	dispatcher := NewProcessingModeDispatcher(streamProcessor, mockSummary)

	// Test with nil data to trigger emergency fallback
	params := ProcessingParams{
		ToolCallID:      "test_call",
		ActivityID:      12345,
		ProcessingMode:  "raw",
	}

	_, err := dispatcher.Dispatch("raw", nil, params)

	// Should return error for nil data
	if err == nil {
		t.Error("Expected error for nil data")
	}

	// Should be a StreamProcessingError
	var streamErr *StreamProcessingError
	if !errors.As(err, &streamErr) {
		t.Error("Expected StreamProcessingError")
	} else {
		if streamErr.Type != "data_corrupted" {
			t.Errorf("Expected error type 'data_corrupted', got '%s'", streamErr.Type)
		}
	}
}

func TestFallbackModeSelection(t *testing.T) {
	config := &StreamConfig{
		MaxContextTokens:  15000,
		TokenPerCharRatio: 0.25,
		DefaultPageSize:   1000,
		MaxPageSize:       5000,
		RedactionEnabled:  true,
	}

	// Create a mock unified stream processor to test fallback mode selection
	processor := &UnifiedStreamProcessor{config: config}

	tests := []struct {
		originalMode    string
		expectedFallbacks []string
	}{
		{"ai-summary", []string{"derived", "raw"}},
		{"derived", []string{"raw"}},
		{"raw", []string{"derived"}},
		{"unknown", []string{"raw", "derived"}},
	}

	for _, test := range tests {
		fallbacks := processor.getFallbackModes(test.originalMode)
		
		if len(fallbacks) != len(test.expectedFallbacks) {
			t.Errorf("Expected %d fallback modes for %s, got %d", 
				len(test.expectedFallbacks), test.originalMode, len(fallbacks))
			continue
		}

		for i, expected := range test.expectedFallbacks {
			if fallbacks[i] != expected {
				t.Errorf("Expected fallback mode %s at position %d for %s, got %s", 
					expected, i, test.originalMode, fallbacks[i])
			}
		}
	}
}