package services

// This file contains integration examples showing how the stream processor
// would be integrated into the AI service for handling large stream outputs.
// This is for documentation purposes and will be used in later tasks.

import (
	"bodda/internal/config"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

// Example of how the stream processor would be integrated into the AI service
func ExampleStreamProcessorIntegration() {
	// This is how the stream processor would be used in the AI service
	// when executing the get-activity-streams tool

	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  15000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1000,
			MaxPageSize:       5000,
			RedactionEnabled:  true,
		},
	}

	// Create stream processor
	streamProcessor := NewStreamProcessor(cfg)
	
	// Create summary processor (would use real OpenAI client in production)
	client := openai.NewClient(option.WithAPIKey("mock-api-key"))
	summaryProcessor := NewSummaryProcessor(&client)
	
	// Create processing mode dispatcher
	dispatcher := NewProcessingModeDispatcher(streamProcessor, summaryProcessor)

	// Example usage in executeGetActivityStreams:
	// 1. Get stream data from Strava API
	streamData := &StravaStreams{
		Time:      []int{0, 1, 2, 3, 4, 5},
		Heartrate: []int{120, 125, 130, 135, 140, 145},
		Watts:     []int{100, 110, 120, 130, 140, 150},
	}

	// 2. Process based on mode
	params := ProcessingParams{
		ToolCallID:    "tool-call-123",
		ActivityID:    456789,
		PageNumber:    1,
		PageSize:      1000,
		SummaryPrompt: "Analyze this workout data",
		Resolution:    "medium",
		StreamTypes:   []string{"time", "heartrate", "watts"},
	}

	// 3. Dispatch to appropriate processing mode
	result, err := dispatcher.Dispatch("auto", streamData, params)
	if err != nil {
		// Handle error
		return
	}

	// 4. Return the processed result
	// result.Content contains human-readable formatted text
	// result.ProcessingMode indicates how the data was processed
	// result.Options contains available processing options (if applicable)
	_ = result
}

// Example of configuration in different environments
func ExampleStreamProcessorConfiguration() {
	// Development configuration - more permissive
	devConfig := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  20000, // Higher limit for development
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1500,
			MaxPageSize:       10000,
			RedactionEnabled:  false, // Disabled for debugging
		},
	}

	// Production configuration - conservative
	prodConfig := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  12000, // Conservative limit
			TokenPerCharRatio: 0.3,   // More conservative estimation
			DefaultPageSize:   800,
			MaxPageSize:       3000,
			RedactionEnabled:  true, // Enabled for optimization
		},
	}

	_ = devConfig
	_ = prodConfig
}

// Example of how different processing modes would be used
func ExampleProcessingModes() {
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
	client := openai.NewClient(option.WithAPIKey("mock-api-key"))
	summaryProcessor := NewSummaryProcessor(&client)
	dispatcher := NewProcessingModeDispatcher(streamProcessor, summaryProcessor)

	// largeStreamData := createLargeStreamData(5000) // From test helper - function not implemented
	largeStreamData := &StravaStreams{} // Placeholder for example

	// Auto mode - system decides
	autoParams := ProcessingParams{
		ToolCallID: "auto-example",
		ActivityID: 123,
	}
	autoResult, _ := dispatcher.Dispatch("auto", largeStreamData, autoParams)
	_ = autoResult // Would return processing options for large data

	// Raw mode - paginated access
	rawParams := ProcessingParams{
		ToolCallID: "raw-example",
		ActivityID: 123,
		PageNumber: 1,
		PageSize:   1000,
	}
	rawResult, _ := dispatcher.Dispatch("raw", largeStreamData, rawParams)
	_ = rawResult // Would return paginated raw data (placeholder for now)

	// Derived mode - statistical analysis
	derivedParams := ProcessingParams{
		ToolCallID: "derived-example",
		ActivityID: 123,
	}
	derivedResult, _ := dispatcher.Dispatch("derived", largeStreamData, derivedParams)
	_ = derivedResult // Would return derived features (placeholder for now)

	// AI Summary mode - AI-powered insights
	summaryParams := ProcessingParams{
		ToolCallID:    "summary-example",
		ActivityID:    123,
		SummaryPrompt: "Provide coaching insights from this workout data",
	}
	summaryResult, _ := dispatcher.Dispatch("ai-summary", largeStreamData, summaryParams)
	_ = summaryResult // Would return AI-generated summary (placeholder for now)
}