package services

import (
	"testing"
)

func TestDerivedFeaturesProcessor_ExtractFeatures(t *testing.T) {
	processor := NewDerivedFeaturesProcessor()

	// Test with nil data
	features, err := processor.ExtractFeatures(nil, nil)
	if err == nil {
		t.Error("Expected error for nil stream data")
	}
	if features != nil {
		t.Error("Expected nil features for nil stream data")
	}

	// Test with valid stream data
	streams := &StravaStreams{
		Time:      []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		Heartrate: []int{120, 125, 130, 135, 140, 145, 150, 155, 160, 165},
		Watts:     []int{100, 110, 120, 130, 140, 150, 160, 170, 180, 190},
		Distance:  []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90},
		Altitude:  []float64{100, 105, 110, 115, 120, 125, 130, 135, 140, 145},
	}

	features, err = processor.ExtractFeatures(streams, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if features == nil {
		t.Fatal("Expected features but got nil")
	}

	// Verify summary
	if features.Summary.TotalDataPoints != 10 {
		t.Errorf("Expected 10 data points, got %d", features.Summary.TotalDataPoints)
	}

	if features.Summary.Duration != 9 {
		t.Errorf("Expected duration 9, got %d", features.Summary.Duration)
	}

	if len(features.Summary.StreamTypes) == 0 {
		t.Error("Expected stream types but got empty slice")
	}

	// Verify statistics
	if features.Statistics.HeartRate == nil {
		t.Error("Expected heart rate statistics")
	}

	if features.Statistics.Power == nil {
		t.Error("Expected power statistics")
	}

	if features.Statistics.Altitude == nil {
		t.Error("Expected altitude statistics")
	}

	// Verify sample data
	if len(features.SampleData) == 0 {
		t.Error("Expected sample data but got empty slice")
	}

	// Verify that inflection points, trends, and spikes are initialized (even if empty)
	if features.InflectionPoints == nil {
		t.Error("Expected inflection points slice (even if empty)")
	} else if len(features.InflectionPoints) < 0 {
		t.Error("Inflection points slice should be valid")
	}

	if features.Trends == nil {
		t.Error("Expected trends slice (even if empty)")
	} else if len(features.Trends) < 0 {
		t.Error("Trends slice should be valid")
	}

	if features.Spikes == nil {
		t.Error("Expected spikes slice (even if empty)")
	} else if len(features.Spikes) < 0 {
		t.Error("Spikes slice should be valid")
	}
}

func TestDerivedFeaturesProcessor_ExtractLapFeatures(t *testing.T) {
	processor := NewDerivedFeaturesProcessor()

	// Test with nil data
	lapAnalysis, err := processor.ExtractLapFeatures(nil, nil)
	if err == nil {
		t.Error("Expected error for nil stream data")
	}
	if lapAnalysis != nil {
		t.Error("Expected nil lap analysis for nil stream data")
	}

	// Test with empty laps
	streams := &StravaStreams{
		Time:      []int{0, 1, 2, 3, 4},
		Heartrate: []int{120, 125, 130, 135, 140},
	}

	lapAnalysis, err = processor.ExtractLapFeatures(streams, []StravaLap{})
	if err == nil {
		t.Error("Expected error for empty laps")
	}

	// Test with valid lap data
	laps := []StravaLap{
		{
			LapIndex:         0,
			Name:             "Lap 1",
			StartIndex:       0,
			EndIndex:         2,
			Distance:         1000,
			ElapsedTime:      120,
			AverageSpeed:     8.33,
			MaxSpeed:         10.0,
			AverageHeartrate: 125,
			MaxHeartrate:     130,
			AveragePower:     150,
			MaxPower:         180,
		},
		{
			LapIndex:         1,
			Name:             "Lap 2",
			StartIndex:       2,
			EndIndex:         4,
			Distance:         1000,
			ElapsedTime:      130,
			AverageSpeed:     7.69,
			MaxSpeed:         9.0,
			AverageHeartrate: 135,
			MaxHeartrate:     140,
			AveragePower:     160,
			MaxPower:         190,
		},
	}

	lapAnalysis, err = processor.ExtractLapFeatures(streams, laps)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if lapAnalysis == nil {
		t.Fatal("Expected lap analysis but got nil")
	}

	if lapAnalysis.TotalLaps != 2 {
		t.Errorf("Expected 2 laps, got %d", lapAnalysis.TotalLaps)
	}

	if lapAnalysis.SegmentationType != "laps" {
		t.Errorf("Expected segmentation type 'laps', got %s", lapAnalysis.SegmentationType)
	}

	if len(lapAnalysis.LapSummaries) != 2 {
		t.Errorf("Expected 2 lap summaries, got %d", len(lapAnalysis.LapSummaries))
	}

	// Verify first lap summary
	lap1 := lapAnalysis.LapSummaries[0]
	if lap1.LapNumber != 1 {
		t.Errorf("Expected lap number 1, got %d", lap1.LapNumber)
	}

	if lap1.LapName != "Lap 1" {
		t.Errorf("Expected lap name 'Lap 1', got %s", lap1.LapName)
	}

	if lap1.Distance != 1000 {
		t.Errorf("Expected distance 1000, got %f", lap1.Distance)
	}

	// Verify lap comparisons
	if lapAnalysis.LapComparisons.FastestLap == 0 {
		t.Error("Expected fastest lap to be set")
	}

	if lapAnalysis.LapComparisons.SlowestLap == 0 {
		t.Error("Expected slowest lap to be set")
	}
}

func TestDerivedFeaturesProcessor_CalculateFeatureSummary(t *testing.T) {
	processor := &derivedFeaturesProcessor{}

	streams := &StravaStreams{
		Time:      []int{0, 10, 20, 30, 40},
		Distance:  []float64{0, 100, 200, 300, 400},
		Heartrate: []int{120, 130, 140, 150, 160},
		Watts:     []int{100, 150, 200, 250, 300},
		Altitude:  []float64{100, 110, 120, 130, 140},
		Moving:    []bool{true, true, false, true, true},
	}

	summary := processor.calculateFeatureSummary(streams)

	if summary.TotalDataPoints != 5 {
		t.Errorf("Expected 5 data points, got %d", summary.TotalDataPoints)
	}

	if summary.Duration != 40 {
		t.Errorf("Expected duration 40, got %d", summary.Duration)
	}

	if summary.TotalDistance != 400 {
		t.Errorf("Expected total distance 400, got %f", summary.TotalDistance)
	}

	if summary.AvgHeartRate != 140 {
		t.Errorf("Expected avg heart rate 140, got %f", summary.AvgHeartRate)
	}

	if summary.MaxHeartRate != 160 {
		t.Errorf("Expected max heart rate 160, got %d", summary.MaxHeartRate)
	}

	if summary.AvgPower != 200 {
		t.Errorf("Expected avg power 200, got %f", summary.AvgPower)
	}

	if summary.MaxPower != 300 {
		t.Errorf("Expected max power 300, got %d", summary.MaxPower)
	}

	if summary.MovingTimePercent != 80 {
		t.Errorf("Expected moving time percent 80, got %f", summary.MovingTimePercent)
	}

	if len(summary.StreamTypes) == 0 {
		t.Error("Expected stream types but got empty slice")
	}
}

func TestDerivedFeaturesProcessor_CountDataPoints(t *testing.T) {
	processor := &derivedFeaturesProcessor{}

	// Test with nil data
	count := processor.countDataPoints(nil)
	if count != 0 {
		t.Errorf("Expected 0 for nil data, got %d", count)
	}

	// Test with empty streams
	streams := &StravaStreams{}
	count = processor.countDataPoints(streams)
	if count != 0 {
		t.Errorf("Expected 0 for empty streams, got %d", count)
	}

	// Test with different length streams
	streams = &StravaStreams{
		Time:      []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, // 10 points
		Heartrate: []int{120, 125, 130, 135, 140},        // 5 points
		Watts:     []int{100, 110, 120},                  // 3 points
	}

	count = processor.countDataPoints(streams)
	if count != 10 {
		t.Errorf("Expected 10 (max length), got %d", count)
	}
}

func TestDerivedFeaturesProcessor_GetAvailableStreamTypes(t *testing.T) {
	processor := &derivedFeaturesProcessor{}

	streams := &StravaStreams{
		Time:      []int{0, 1, 2},
		Heartrate: []int{120, 125, 130},
		Watts:     []int{100, 110, 120},
		Distance:  []float64{0, 10, 20},
	}

	types := processor.getAvailableStreamTypes(streams)

	expectedTypes := []string{"time", "distance", "heartrate", "watts"}
	if len(types) != len(expectedTypes) {
		t.Errorf("Expected %d stream types, got %d", len(expectedTypes), len(types))
	}

	// Check that all expected types are present
	typeMap := make(map[string]bool)
	for _, t := range types {
		typeMap[t] = true
	}

	for _, expected := range expectedTypes {
		if !typeMap[expected] {
			t.Errorf("Expected stream type %s not found", expected)
		}
	}
}