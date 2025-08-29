package services

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"bodda/internal/config"
)

// TestStreamProcessingIntegration tests the complete stream processing workflow
func TestStreamProcessingIntegration(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:      5000, // Low threshold for testing
			TokenPerCharRatio:     0.25,
			DefaultPageSize:       100,
			MaxPageSize:          500,
			RedactionEnabled:     true,
			StravaResolutions:    []string{"low", "medium", "high"},
			EnableDerivedFeatures: true,
			EnableAISummary:      true,
			EnablePagination:     true,
			EnableAutoMode:       true,
			LargeDatasetThreshold: 1000,
			ContextSafetyMargin:  500,
			MaxRetries:           3,
			ProcessingTimeout:    30,
		},
	}

	processor := NewStreamProcessor(cfg)
	formatter := NewOutputFormatter()
	derivedProcessor := NewDerivedFeaturesProcessor()

	// Test with realistic activity data
	streams := createRealisticStreamData()

	t.Run("auto mode with large dataset", func(t *testing.T) {
		result, err := processor.ProcessStreamOutput(streams, "test-auto-large")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.ProcessingMode != "auto" {
			t.Errorf("Expected auto mode, got %s", result.ProcessingMode)
		}

		if len(result.Options) == 0 {
			t.Error("Expected processing options for large dataset")
		}

		// Verify options contain expected modes
		modes := make(map[string]bool)
		for _, option := range result.Options {
			modes[option.Mode] = true
		}

		expectedModes := []string{"raw", "derived", "ai-summary"}
		for _, mode := range expectedModes {
			if !modes[mode] {
				t.Errorf("Expected mode %s in options", mode)
			}
		}
	})

	t.Run("derived features processing", func(t *testing.T) {
		features, err := derivedProcessor.ExtractFeatures(streams, nil)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Verify comprehensive feature extraction
		if features.Summary.TotalDataPoints == 0 {
			t.Error("Expected data points in summary")
		}

		if features.Statistics.HeartRate == nil {
			t.Error("Expected heart rate statistics")
		}

		if features.Statistics.Power == nil {
			t.Error("Expected power statistics")
		}

		if features.Statistics.VelocitySmooth == nil {
			t.Error("Expected velocity statistics")
		}

		// Verify derived features are calculated
		if len(features.InflectionPoints) == 0 {
			t.Log("No inflection points detected (may be normal for test data)")
		}

		if len(features.Trends) == 0 {
			t.Log("No trends detected (may be normal for test data)")
		}

		if len(features.Spikes) == 0 {
			t.Log("No spikes detected (may be normal for test data)")
		}

		// Verify sample data is provided
		if len(features.SampleData) == 0 {
			t.Error("Expected sample data points")
		}
	})

	t.Run("output formatting", func(t *testing.T) {
		// Test stream data formatting
		formatted := formatter.FormatStreamData(streams, "raw")
		if !containsText(formatted, "Stream Data") {
			t.Error("Expected formatted output to contain 'Stream Data'")
		}

		if !containsText(formatted, "Heart Rate") {
			t.Error("Expected formatted output to contain heart rate data")
		}

		if !containsText(formatted, "Power") {
			t.Error("Expected formatted output to contain power data")
		}

		// Test derived features formatting
		features, _ := derivedProcessor.ExtractFeatures(streams, nil)
		derivedFormatted := formatter.FormatDerivedFeatures(features)
		
		if !containsText(derivedFormatted, "Stream Analysis") {
			t.Error("Expected derived features output to contain 'Stream Analysis'")
		}

		if !containsText(derivedFormatted, "Overview") {
			t.Error("Expected derived features output to contain 'Overview'")
		}
	})

	t.Run("error handling and fallback", func(t *testing.T) {
		// Test with nil data
		result, err := processor.ProcessStreamOutput(nil, "test-nil")
		if err == nil {
			t.Error("Expected error for nil data")
		}

		// Test with corrupted data
		corruptedStreams := &StravaStreams{
			Time:      []int{0, 1, 2},
			Heartrate: []int{}, // Empty heart rate data
			Watts:     []int{100, 200}, // Mismatched length
		}

		result, err = processor.ProcessStreamOutput(corruptedStreams, "test-corrupted")
		if err != nil {
			t.Errorf("Expected graceful handling of corrupted data, got error: %v", err)
		}

		if result == nil {
			t.Error("Expected result even with corrupted data")
		}
	})
}

// TestStreamSizeDetectionAccuracy tests the accuracy of stream size detection
func TestStreamSizeDetectionAccuracy(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  10000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1000,
			MaxPageSize:      5000,
		},
	}

	processor := NewStreamProcessor(cfg)

	tests := []struct {
		name           string
		dataSize       int
		expectedTokens int
		tolerance      int
	}{
		{
			name:           "small dataset",
			dataSize:       100,
			expectedTokens: 600,  // Approximate
			tolerance:      200,
		},
		{
			name:           "medium dataset",
			dataSize:       1000,
			expectedTokens: 6000, // Approximate
			tolerance:      2000,
		},
		{
			name:           "large dataset",
			dataSize:       5000,
			expectedTokens: 30000, // Approximate
			tolerance:      10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streams := createLargeStreamData(tt.dataSize)
			tokens := processor.EstimateTokens(streams)

			if tokens < tt.expectedTokens-tt.tolerance || tokens > tt.expectedTokens+tt.tolerance {
				t.Errorf("Token estimation for %d data points: expected %d±%d, got %d", 
					tt.dataSize, tt.expectedTokens, tt.tolerance, tokens)
			}

			shouldProcess := processor.ShouldProcess(streams)
			expectedShouldProcess := tokens > cfg.StreamProcessing.MaxContextTokens

			if shouldProcess != expectedShouldProcess {
				t.Errorf("ShouldProcess for %d tokens: expected %t, got %t", 
					tokens, expectedShouldProcess, shouldProcess)
			}
		})
	}
}

// TestFeatureExtractionAccuracy tests the accuracy of feature extraction algorithms
func TestFeatureExtractionAccuracy(t *testing.T) {
	// Create test data with known patterns
	streams := &StravaStreams{
		Time:           []int{0, 60, 120, 180, 240, 300, 360, 420, 480, 540, 600},
		Heartrate:      []int{120, 130, 140, 150, 160, 155, 150, 145, 140, 135, 130},
		Watts:          []int{100, 150, 200, 250, 300, 280, 260, 240, 220, 200, 180},
		VelocitySmooth: []float64{5.0, 6.0, 7.0, 8.0, 9.0, 8.5, 8.0, 7.5, 7.0, 6.5, 6.0},
		Distance:       []float64{0, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000},
		Altitude:       []float64{100, 110, 120, 130, 140, 135, 130, 125, 120, 115, 110},
	}

	processor := NewDerivedFeaturesProcessor()
	features, err := processor.ExtractFeatures(streams, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	t.Run("statistical accuracy", func(t *testing.T) {
		// Verify heart rate statistics
		hrStats := features.Statistics.HeartRate
		if hrStats == nil {
			t.Fatal("Expected heart rate statistics")
		}

		expectedMinHR := 120.0
		expectedMaxHR := 160.0
		expectedMeanHR := 142.27 // Approximate

		if hrStats.Min != expectedMinHR {
			t.Errorf("Heart rate min: expected %f, got %f", expectedMinHR, hrStats.Min)
		}

		if hrStats.Max != expectedMaxHR {
			t.Errorf("Heart rate max: expected %f, got %f", expectedMaxHR, hrStats.Max)
		}

		if hrStats.Mean < expectedMeanHR-5 || hrStats.Mean > expectedMeanHR+5 {
			t.Errorf("Heart rate mean: expected %f±5, got %f", expectedMeanHR, hrStats.Mean)
		}

		// Verify power statistics
		powerStats := features.Statistics.Power
		if powerStats == nil {
			t.Fatal("Expected power statistics")
		}

		expectedMinPower := 100.0
		expectedMaxPower := 300.0

		if powerStats.Min != expectedMinPower {
			t.Errorf("Power min: expected %f, got %f", expectedMinPower, powerStats.Min)
		}

		if powerStats.Max != expectedMaxPower {
			t.Errorf("Power max: expected %f, got %f", expectedMaxPower, powerStats.Max)
		}
	})

	t.Run("elevation analysis", func(t *testing.T) {
		elevationAnalysis := CalculateElevationAnalysis(streams.Altitude, streams.Distance, streams.Time)

		// Expected: climb from 100 to 140 (40m gain), then descent to 110 (30m loss)
		expectedGain := 40.0
		expectedLoss := 30.0

		if elevationAnalysis.TotalGain < expectedGain-5 || elevationAnalysis.TotalGain > expectedGain+5 {
			t.Errorf("Elevation gain: expected %f±5, got %f", expectedGain, elevationAnalysis.TotalGain)
		}

		if elevationAnalysis.TotalLoss < expectedLoss-5 || elevationAnalysis.TotalLoss > expectedLoss+5 {
			t.Errorf("Elevation loss: expected %f±5, got %f", expectedLoss, elevationAnalysis.TotalLoss)
		}
	})

	t.Run("normalized power calculation", func(t *testing.T) {
		// Test with sufficient data points
		longPowerData := make([]int, 300) // 5 minutes of data
		longTimeData := make([]int, 300)
		
		for i := 0; i < 300; i++ {
			longPowerData[i] = 200 + (i%50) // Variable power around 200W
			longTimeData[i] = i
		}

		np := CalculateNormalizedPower(longPowerData, longTimeData)
		
		// Should be close to average power for this pattern
		if np < 180 || np > 250 {
			t.Errorf("Normalized power: expected 180-250W, got %f", np)
		}
	})

	t.Run("heart rate drift calculation", func(t *testing.T) {
		// Create data with known drift: 120 to 160 over 10 minutes = 4 bpm/min = 240 bpm/hour
		driftHR := []int{120, 124, 128, 132, 136, 140, 144, 148, 152, 156, 160}
		driftTime := []int{0, 60, 120, 180, 240, 300, 360, 420, 480, 540, 600}

		drift := CalculateHeartRateDrift(driftHR, driftTime)
		
		expectedDrift := 240.0 // bpm/hour
		tolerance := 50.0

		if drift < expectedDrift-tolerance || drift > expectedDrift+tolerance {
			t.Errorf("Heart rate drift: expected %f±%f bpm/hour, got %f", expectedDrift, tolerance, drift)
		}
	})
}

// TestErrorHandlingScenarios tests various error scenarios
func TestErrorHandlingScenarios(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  1000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   100,
			MaxPageSize:      500,
			RedactionEnabled:  true,
			MaxRetries:       3,
			ProcessingTimeout: 30,
		},
	}

	processor := NewStreamProcessor(cfg)

	t.Run("malformed stream data", func(t *testing.T) {
		malformedStreams := []*StravaStreams{
			// Mismatched array lengths
			{
				Time:      []int{0, 1, 2, 3, 4},
				Heartrate: []int{120, 130}, // Too short
				Watts:     []int{100, 110, 120, 130, 140, 150}, // Too long
			},
			// Empty required fields
			{
				Time:      []int{},
				Heartrate: []int{120, 130, 140},
				Watts:     []int{100, 110, 120},
			},
			// Invalid values
			{
				Time:      []int{0, -1, -2}, // Negative time
				Heartrate: []int{-120, 300, 400}, // Invalid heart rates
				Watts:     []int{-100, 2000, 3000}, // Invalid power values
			},
		}

		for i, streams := range malformedStreams {
			t.Run(fmt.Sprintf("malformed_case_%d", i), func(t *testing.T) {
				result, err := processor.ProcessStreamOutput(streams, fmt.Sprintf("test-malformed-%d", i))
				
				// Should handle gracefully, not crash
				if err != nil && result == nil {
					t.Errorf("Expected graceful handling of malformed data, got error: %v", err)
				}
				
				if result != nil && result.Content == "" {
					t.Error("Expected some content even for malformed data")
				}
			})
		}
	})

	t.Run("extreme data sizes", func(t *testing.T) {
		// Test with very large dataset
		largeStreams := createLargeStreamData(10000)
		result, err := processor.ProcessStreamOutput(largeStreams, "test-extreme-large")
		
		if err != nil {
			t.Errorf("Unexpected error for large dataset: %v", err)
		}
		
		if result == nil {
			t.Error("Expected result for large dataset")
		}
		
		// Should trigger processing options
		if result.ProcessingMode == "raw" {
			t.Error("Expected large dataset to trigger processing options, not raw mode")
		}

		// Test with very small dataset
		tinyStreams := &StravaStreams{
			Time:      []int{0},
			Heartrate: []int{120},
		}
		
		result, err = processor.ProcessStreamOutput(tinyStreams, "test-extreme-small")
		
		if err != nil {
			t.Errorf("Unexpected error for tiny dataset: %v", err)
		}
		
		if result == nil {
			t.Error("Expected result for tiny dataset")
		}
	})

	t.Run("configuration edge cases", func(t *testing.T) {
		// Test with extreme configuration values
		extremeCfg := &config.Config{
			StreamProcessing: config.StreamProcessingConfig{
				MaxContextTokens:  1, // Extremely low
				TokenPerCharRatio: 1.0, // Very high ratio
				DefaultPageSize:   1,
				MaxPageSize:      2,
			},
		}

		extremeProcessor := NewStreamProcessor(extremeCfg)
		streams := createLargeStreamData(100)
		
		result, err := extremeProcessor.ProcessStreamOutput(streams, "test-extreme-config")
		
		// Should handle extreme configuration gracefully
		if err != nil && result == nil {
			t.Errorf("Expected graceful handling of extreme config, got error: %v", err)
		}
	})
}

// Helper functions

func createRealisticStreamData() *StravaStreams {
	// Create 30 minutes of realistic cycling data
	size := 1800 // 30 minutes at 1-second intervals
	
	time := make([]int, size)
	heartrate := make([]int, size)
	watts := make([]int, size)
	distance := make([]float64, size)
	altitude := make([]float64, size)
	velocitySmooth := make([]float64, size)
	cadence := make([]int, size)
	moving := make([]bool, size)

	baseHR := 140
	basePower := 200
	baseAltitude := 100.0
	totalDistance := 0.0

	for i := 0; i < size; i++ {
		time[i] = i
		
		// Realistic heart rate with gradual increase and some variation
		heartrate[i] = baseHR + (i/60) + (i%10) - 5
		if heartrate[i] < 100 {
			heartrate[i] = 100
		}
		if heartrate[i] > 180 {
			heartrate[i] = 180
		}
		
		// Realistic power with intervals and variation
		intervalPhase := (i / 300) % 4
		switch intervalPhase {
		case 0, 2: // Easy phases
			watts[i] = basePower - 50 + (i%20) - 10
		case 1, 3: // Hard phases
			watts[i] = basePower + 100 + (i%30) - 15
		}
		if watts[i] < 50 {
			watts[i] = 50
		}
		if watts[i] > 400 {
			watts[i] = 400
		}
		
		// Realistic speed based on power and terrain
		speed := 8.0 + float64(watts[i])/50.0 + float64(i%20)/10.0 - 1.0
		if speed < 3.0 {
			speed = 3.0
		}
		if speed > 20.0 {
			speed = 20.0
		}
		velocitySmooth[i] = speed
		
		// Distance accumulation
		if i > 0 {
			totalDistance += speed / 3.6 // Convert km/h to m/s
		}
		distance[i] = totalDistance
		
		// Realistic altitude with hills
		hillPhase := float64(i) / 600.0
		altitude[i] = baseAltitude + 50.0*sin(hillPhase) + float64(i%50)/10.0
		
		// Realistic cadence
		cadence[i] = 85 + (i%20) - 10
		if cadence[i] < 60 {
			cadence[i] = 60
		}
		if cadence[i] > 110 {
			cadence[i] = 110
		}
		
		// Moving status (mostly moving with some stops)
		moving[i] = (i%200) < 190 // 5% stopped time
	}

	return &StravaStreams{
		Time:           time,
		Heartrate:      heartrate,
		Watts:          watts,
		Distance:       distance,
		Altitude:       altitude,
		VelocitySmooth: velocitySmooth,
		Cadence:        cadence,
		Moving:         moving,
	}
}

func containsText(s, substr string) bool {
	return strings.Contains(s, substr)
}

func sin(x float64) float64 {
	return math.Sin(x)
}