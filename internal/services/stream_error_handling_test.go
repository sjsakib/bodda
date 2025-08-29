package services

import (
	"errors"
	"testing"

	"bodda/internal/config"
)

func TestStreamProcessingError(t *testing.T) {
	// Test creating a new stream processing error
	err := NewStreamProcessingError("test_error", "Test error message", 12345, "raw")
	
	if err.Type != "test_error" {
		t.Errorf("Expected error type 'test_error', got '%s'", err.Type)
	}
	
	if err.Message != "Test error message" {
		t.Errorf("Expected message 'Test error message', got '%s'", err.Message)
	}
	
	if err.ActivityID != 12345 {
		t.Errorf("Expected activity ID 12345, got %d", err.ActivityID)
	}
	
	if err.ProcessingMode != "raw" {
		t.Errorf("Expected processing mode 'raw', got '%s'", err.ProcessingMode)
	}
}

func TestStreamProcessingErrorChaining(t *testing.T) {
	originalErr := errors.New("original error")
	
	err := NewStreamProcessingError("test_error", "Test error message", 12345, "raw").
		WithOriginalError(originalErr).
		WithDataSize(1000).
		WithAvailableTokens(500).
		WithContext("test_key", "test_value")
	
	if err.OriginalError != originalErr {
		t.Errorf("Expected original error to be preserved")
	}
	
	if err.DataSize != 1000 {
		t.Errorf("Expected data size 1000, got %d", err.DataSize)
	}
	
	if err.AvailableTokens != 500 {
		t.Errorf("Expected available tokens 500, got %d", err.AvailableTokens)
	}
	
	if err.Context["test_key"] != "test_value" {
		t.Errorf("Expected context value 'test_value', got '%v'", err.Context["test_key"])
	}
}

func TestStreamProcessorErrorHandling(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  15000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1000,
			MaxPageSize:       5000,
			RedactionEnabled:  true,
		},
	}

	processor := NewStreamProcessor(cfg)
	streamProcessor := processor.(*streamProcessor)

	// Test handling a generic error
	originalErr := errors.New("test processing error")
	data := &StravaStreams{
		Time:      []int{1, 2, 3, 4, 5},
		Heartrate: []int{120, 125, 130, 135, 140},
	}

	result := streamProcessor.HandleProcessingError(originalErr, 12345, "raw", data)

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.ProcessingMode != "error" {
		t.Errorf("Expected processing mode 'error', got '%s'", result.ProcessingMode)
	}

	if len(result.Options) == 0 {
		t.Error("Expected processing options to be provided")
	}

	// Verify the content contains error information
	if !containsString(result.Content, "Stream Processing Error") {
		t.Errorf("Expected error content to contain 'Stream Processing Error', got: %s", result.Content)
	}

	if !containsString(result.Content, "12345") {
		t.Errorf("Expected error content to contain activity ID 12345, got: %s", result.Content)
	}
}

func TestStreamProcessorFallbackFormatter(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  15000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1000,
			MaxPageSize:       5000,
			RedactionEnabled:  true,
		},
	}

	processor := NewStreamProcessor(cfg)
	streamProcessor := processor.(*streamProcessor)

	// Test with valid data
	data := &StravaStreams{
		Time:      []int{0, 60, 120, 180, 240},
		Heartrate: []int{120, 125, 130, 135, 140},
		Watts:     []int{200, 210, 220, 215, 205},
		Distance:  []float64{0, 100, 200, 300, 400},
		Altitude:  []float64{100, 105, 110, 108, 106},
	}

	result := streamProcessor.CreateFallbackFormatter(data, 12345, "derived")

	if !containsString(result, "Fallback Stream Information") {
		t.Errorf("Expected fallback content to contain 'Fallback Stream Information', got: %s", result)
	}

	if !containsString(result, "12345") {
		t.Errorf("Expected fallback content to contain activity ID 12345, got: %s", result)
	}

	if !containsString(result, "derived") {
		t.Errorf("Expected fallback content to contain failed processing mode 'derived', got: %s", result)
	}

	if !containsString(result, "Heart Rate:") {
		t.Error("Expected fallback content to contain heart rate statistics")
	}

	if !containsString(result, "Power:") {
		t.Error("Expected fallback content to contain power statistics")
	}

	// Test with nil data
	nilResult := streamProcessor.CreateFallbackFormatter(nil, 12345, "raw")

	if !containsString(nilResult, "No stream data available") {
		t.Error("Expected nil data fallback to contain 'No stream data available'")
	}
}

func TestErrorRecoverySuggestions(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  15000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1000,
			MaxPageSize:       5000,
			RedactionEnabled:  true,
		},
	}

	processor := NewStreamProcessor(cfg)
	streamProcessor := processor.(*streamProcessor)

	tests := []struct {
		errorType           string
		expectedSuggestion  string
	}{
		{"strava_api_failure", "Check your Strava API connection"},
		{"context_exceeded", "Use pagination with smaller page_size"},
		{"processing_failure", "Try a different processing mode"},
		{"invalid_request", "Verify all required parameters"},
		{"data_corrupted", "Try requesting different stream types"},
		{"unknown_error", "Try a different processing mode"},
	}

	for _, test := range tests {
		err := NewStreamProcessingError(test.errorType, "Test error", 12345, "raw")
		suggestions := streamProcessor.getRecoverySuggestions(err)

		if !containsString(suggestions, test.expectedSuggestion) {
			t.Errorf("Expected suggestions for %s to contain '%s', got: %s", 
				test.errorType, test.expectedSuggestion, suggestions)
		}
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}