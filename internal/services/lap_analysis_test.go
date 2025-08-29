package services

import (
	"math"
	"testing"
)

func TestAnalyzeLapByLap(t *testing.T) {
	// Create test stream data
	streams := &StravaStreams{
		Time:           []int{0, 60, 120, 180, 240, 300, 360, 420, 480, 540, 600},
		Distance:       []float64{0, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000},
		Heartrate:      []int{120, 125, 130, 135, 140, 145, 150, 145, 140, 135, 130},
		Watts:          []int{100, 110, 120, 130, 140, 150, 160, 150, 140, 130, 120},
		VelocitySmooth: []float64{5.0, 5.2, 5.4, 5.6, 5.8, 6.0, 6.2, 6.0, 5.8, 5.6, 5.4},
		Altitude:       []float64{100, 102, 104, 106, 108, 110, 108, 106, 104, 102, 100},
	}

	// Create test lap data
	laps := []StravaLap{
		{
			LapIndex:   0,
			Name:       "Lap 1",
			StartIndex: 0,
			EndIndex:   4,
			Distance:   400,
			ElapsedTime: 240,
		},
		{
			LapIndex:   1,
			Name:       "Lap 2",
			StartIndex: 5,
			EndIndex:   9,
			Distance:   400,
			ElapsedTime: 240,
		},
	}

	tests := []struct {
		name     string
		streams  *StravaStreams
		laps     []StravaLap
		expected int // expected number of laps
	}{
		{
			name:     "normal lap analysis",
			streams:  streams,
			laps:     laps,
			expected: 2,
		},
		{
			name:     "empty laps fallback to distance",
			streams:  streams,
			laps:     []StravaLap{},
			expected: 1, // 1km total distance with 1km segments
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnalyzeLapByLap(tt.streams, tt.laps)
			
			if result.TotalLaps != tt.expected {
				t.Errorf("TotalLaps = %d, want %d", result.TotalLaps, tt.expected)
			}

			if len(result.LapSummaries) != tt.expected {
				t.Errorf("LapSummaries length = %d, want %d", len(result.LapSummaries), tt.expected)
			}

			// Verify lap summaries have reasonable data
			for i, lap := range result.LapSummaries {
				if lap.LapNumber != i+1 {
					t.Errorf("Lap %d has wrong lap number: %d", i, lap.LapNumber)
				}
				if lap.Duration <= 0 && len(tt.laps) > 0 {
					t.Errorf("Lap %d has invalid duration: %d", i, lap.Duration)
				}
				if lap.AvgHeartRate <= 0 {
					t.Errorf("Lap %d has invalid average heart rate: %f", i, lap.AvgHeartRate)
				}
			}

			// Verify segmentation type
			expectedType := "laps"
			if len(tt.laps) == 0 {
				expectedType = "distance"
			}
			if result.SegmentationType != expectedType {
				t.Errorf("SegmentationType = %s, want %s", result.SegmentationType, expectedType)
			}
		})
	}
}

func TestAnalyzeDistanceSegments(t *testing.T) {
	streams := &StravaStreams{
		Time:           []int{0, 60, 120, 180, 240, 300, 360, 420, 480, 540, 600},
		Distance:       []float64{0, 200, 400, 600, 800, 1000, 1200, 1400, 1600, 1800, 2000},
		Heartrate:      []int{120, 125, 130, 135, 140, 145, 150, 145, 140, 135, 130},
		VelocitySmooth: []float64{12.0, 12.2, 12.4, 12.6, 12.8, 13.0, 13.2, 13.0, 12.8, 12.6, 12.4},
	}

	tests := []struct {
		name        string
		streams     *StravaStreams
		segmentSize float64
		expected    int
	}{
		{
			name:        "1km segments",
			streams:     streams,
			segmentSize: 1000,
			expected:    2, // 2km total distance
		},
		{
			name:        "500m segments",
			streams:     streams,
			segmentSize: 500,
			expected:    4, // 2km total distance
		},
		{
			name:        "segment larger than total distance",
			streams:     streams,
			segmentSize: 5000,
			expected:    1, // One segment for entire activity
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnalyzeDistanceSegments(tt.streams, tt.segmentSize)
			
			if result.TotalLaps != tt.expected {
				t.Errorf("TotalLaps = %d, want %d", result.TotalLaps, tt.expected)
			}

			if result.SegmentationType != "distance" {
				t.Errorf("SegmentationType = %s, want distance", result.SegmentationType)
			}

			// Verify segments have reasonable data
			for i, segment := range result.LapSummaries {
				if segment.LapNumber != i+1 {
					t.Errorf("Segment %d has wrong number: %d", i, segment.LapNumber)
				}
				if segment.Distance <= 0 {
					t.Errorf("Segment %d has invalid distance: %f", i, segment.Distance)
				}
			}
		})
	}
}

func TestAnalyzeSingleLap(t *testing.T) {
	streams := &StravaStreams{
		Time:           []int{0, 60, 120, 180, 240, 300},
		Heartrate:      []int{120, 125, 130, 135, 140, 145},
		Watts:          []int{100, 110, 120, 130, 140, 150},
		VelocitySmooth: []float64{5.0, 5.2, 5.4, 5.6, 5.8, 6.0},
		Cadence:        []int{80, 82, 84, 86, 88, 90},
		Altitude:       []float64{100, 102, 104, 106, 108, 110},
		Temp:           []int{20, 21, 22, 23, 24, 25},
	}

	lap := StravaLap{
		LapIndex:    0,
		Name:        "Test Lap",
		StartIndex:  0,
		EndIndex:    5,
		Distance:    500,
		ElapsedTime: 300,
	}

	result := analyzeSingleLap(streams, lap, 1)

	// Test basic lap information
	if result.LapNumber != 1 {
		t.Errorf("LapNumber = %d, want 1", result.LapNumber)
	}
	if result.LapName != "Test Lap" {
		t.Errorf("LapName = %s, want Test Lap", result.LapName)
	}
	if result.Distance != 500 {
		t.Errorf("Distance = %f, want 500", result.Distance)
	}
	if result.Duration != 300 {
		t.Errorf("Duration = %d, want 300", result.Duration)
	}

	// Test heart rate calculations
	expectedAvgHR := 132.5 // (120+125+130+135+140+145)/6
	if math.Abs(result.AvgHeartRate-expectedAvgHR) > 0.1 {
		t.Errorf("AvgHeartRate = %f, want %f", result.AvgHeartRate, expectedAvgHR)
	}
	if result.MaxHeartRate != 145 {
		t.Errorf("MaxHeartRate = %d, want 145", result.MaxHeartRate)
	}

	// Test power calculations
	expectedAvgPower := 125.0 // (100+110+120+130+140+150)/6
	if math.Abs(result.AvgPower-expectedAvgPower) > 0.1 {
		t.Errorf("AvgPower = %f, want %f", result.AvgPower, expectedAvgPower)
	}
	if result.MaxPower != 150 {
		t.Errorf("MaxPower = %d, want 150", result.MaxPower)
	}

	// Test speed calculations
	expectedAvgSpeed := 5.5 // (5.0+5.2+5.4+5.6+5.8+6.0)/6
	if math.Abs(result.AvgSpeed-expectedAvgSpeed) > 0.1 {
		t.Errorf("AvgSpeed = %f, want %f", result.AvgSpeed, expectedAvgSpeed)
	}
	if result.MaxSpeed != 6.0 {
		t.Errorf("MaxSpeed = %f, want 6.0", result.MaxSpeed)
	}

	// Test elevation calculations
	expectedElevationGain := 10.0 // 110 - 100
	if math.Abs(result.ElevationGain-expectedElevationGain) > 0.1 {
		t.Errorf("ElevationGain = %f, want %f", result.ElevationGain, expectedElevationGain)
	}

	// Test statistics are populated
	if result.Statistics.HeartRate == nil {
		t.Error("HeartRate statistics should not be nil")
	}
	if result.Statistics.Power == nil {
		t.Error("Power statistics should not be nil")
	}
	if result.Statistics.Speed == nil {
		t.Error("Speed statistics should not be nil")
	}
}

func TestCalculateLapComparisons(t *testing.T) {
	lapSummaries := []LapSummary{
		{
			LapNumber:    1,
			AvgSpeed:     5.0,
			AvgPower:     100,
			AvgHeartRate: 140,
		},
		{
			LapNumber:    2,
			AvgSpeed:     6.0, // Fastest
			AvgPower:     120, // Highest
			AvgHeartRate: 150, // Highest
		},
		{
			LapNumber:    3,
			AvgSpeed:     4.0, // Slowest
			AvgPower:     80,  // Lowest
			AvgHeartRate: 130, // Lowest
		},
	}

	result := calculateLapComparisons(lapSummaries)

	if result.FastestLap != 2 {
		t.Errorf("FastestLap = %d, want 2", result.FastestLap)
	}
	if result.SlowestLap != 3 {
		t.Errorf("SlowestLap = %d, want 3", result.SlowestLap)
	}
	if result.HighestPowerLap != 2 {
		t.Errorf("HighestPowerLap = %d, want 2", result.HighestPowerLap)
	}
	if result.LowestPowerLap != 3 {
		t.Errorf("LowestPowerLap = %d, want 3", result.LowestPowerLap)
	}
	if result.HighestHRLap != 2 {
		t.Errorf("HighestHRLap = %d, want 2", result.HighestHRLap)
	}
	if result.LowestHRLap != 3 {
		t.Errorf("LowestHRLap = %d, want 3", result.LowestHRLap)
	}

	// Test that variations are calculated (should be > 0 for varied data)
	if result.SpeedVariation <= 0 {
		t.Errorf("SpeedVariation should be > 0, got %f", result.SpeedVariation)
	}
	if result.PowerVariation <= 0 {
		t.Errorf("PowerVariation should be > 0, got %f", result.PowerVariation)
	}
	if result.HRVariation <= 0 {
		t.Errorf("HRVariation should be > 0, got %f", result.HRVariation)
	}

	// Consistency score should be between 0 and 1
	if result.ConsistencyScore < 0 || result.ConsistencyScore > 1 {
		t.Errorf("ConsistencyScore should be between 0 and 1, got %f", result.ConsistencyScore)
	}
}

func TestCreateDistanceSegments(t *testing.T) {
	tests := []struct {
		name         string
		distanceData []float64
		segmentSize  float64
		expected     int
	}{
		{
			name:         "empty data",
			distanceData: []float64{},
			segmentSize:  1000,
			expected:     0,
		},
		{
			name:         "normal segmentation",
			distanceData: []float64{0, 500, 1000, 1500, 2000, 2500},
			segmentSize:  1000,
			expected:     3, // 0-1000, 1000-2000, 2000-2500
		},
		{
			name:         "exact fit",
			distanceData: []float64{0, 1000, 2000},
			segmentSize:  1000,
			expected:     2, // 0-1000, 1000-2000
		},
		{
			name:         "short activity",
			distanceData: []float64{0, 200, 400, 600},
			segmentSize:  1000,
			expected:     1, // One segment for entire activity
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createDistanceSegments(tt.distanceData, tt.segmentSize)
			
			if len(result) != tt.expected {
				t.Errorf("Number of segments = %d, want %d", len(result), tt.expected)
			}

			// Verify segment properties
			for i, segment := range result {
				if segment.SegmentNumber != i+1 {
					t.Errorf("Segment %d has wrong number: %d", i, segment.SegmentNumber)
				}
				if segment.StartIndex < 0 || segment.EndIndex >= len(tt.distanceData) {
					t.Errorf("Segment %d has invalid indices: start=%d, end=%d", i, segment.StartIndex, segment.EndIndex)
				}
				if segment.StartIndex >= segment.EndIndex {
					t.Errorf("Segment %d start index should be less than end index", i)
				}
				if segment.Distance <= 0 {
					t.Errorf("Segment %d has invalid distance: %f", i, segment.Distance)
				}
			}
		})
	}
}

func TestFindDistanceIndex(t *testing.T) {
	distanceData := []float64{0, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000}

	tests := []struct {
		name           string
		targetDistance float64
		expected       int
	}{
		{
			name:           "exact match",
			targetDistance: 500,
			expected:       5,
		},
		{
			name:           "between values",
			targetDistance: 550,
			expected:       6, // Should round up to next index
		},
		{
			name:           "start of range",
			targetDistance: 0,
			expected:       0,
		},
		{
			name:           "end of range",
			targetDistance: 1000,
			expected:       10,
		},
		{
			name:           "beyond range",
			targetDistance: 1500,
			expected:       10, // Should return last index
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findDistanceIndex(distanceData, tt.targetDistance)
			
			if result != tt.expected {
				t.Errorf("findDistanceIndex(%f) = %d, want %d", tt.targetDistance, result, tt.expected)
			}
		})
	}
}

func TestAnalyzeLapTrends(t *testing.T) {
	streams := &StravaStreams{
		Time:           []int{0, 60, 120, 180, 240, 300, 360, 420, 480, 540},
		Heartrate:      []int{120, 125, 130, 135, 140, 145, 150, 155, 160, 165}, // Increasing
		Watts:          []int{200, 190, 180, 170, 160, 150, 140, 130, 120, 110}, // Decreasing
		VelocitySmooth: []float64{5.0, 5.0, 5.0, 5.0, 5.0, 5.0, 5.0, 5.0, 5.0, 5.0}, // Stable
	}

	lap := StravaLap{
		StartIndex: 0,
		EndIndex:   9,
	}

	result := analyzeLapTrends(streams, lap)

	// Should detect trends for heart rate, power, and speed
	if len(result) < 3 {
		t.Errorf("Expected at least 3 trends, got %d", len(result))
	}

	// Find specific trends
	var hrTrend, powerTrend, speedTrend *LapTrend
	for i := range result {
		switch result[i].Metric {
		case "heart_rate":
			hrTrend = &result[i]
		case "power":
			powerTrend = &result[i]
		case "speed":
			speedTrend = &result[i]
		}
	}

	// Test heart rate trend (should be increasing)
	if hrTrend == nil {
		t.Error("Heart rate trend not found")
	} else if hrTrend.Direction != "increasing" {
		t.Errorf("Heart rate trend direction = %s, want increasing", hrTrend.Direction)
	}

	// Test power trend (should be decreasing)
	if powerTrend == nil {
		t.Error("Power trend not found")
	} else if powerTrend.Direction != "decreasing" {
		t.Errorf("Power trend direction = %s, want decreasing", powerTrend.Direction)
	}

	// Test speed trend (should be stable)
	if speedTrend == nil {
		t.Error("Speed trend not found")
	} else if speedTrend.Direction != "stable" {
		t.Errorf("Speed trend direction = %s, want stable", speedTrend.Direction)
	}
}

func TestDetectLapSpikes(t *testing.T) {
	streams := &StravaStreams{
		Time:      []int{0, 60, 120, 180, 240, 300, 360, 420, 480, 540},
		Heartrate: []int{140, 142, 141, 180, 143, 142, 141, 140, 139, 141}, // Spike at index 3
		Watts:     []int{150, 155, 160, 400, 165, 160, 155, 150, 145, 155}, // Spike at index 3
	}

	lap := StravaLap{
		StartIndex: 0,
		EndIndex:   9,
	}

	result := detectLapSpikes(streams, lap)

	// Should detect spikes in both power and heart rate
	if len(result) < 2 {
		t.Errorf("Expected at least 2 spikes, got %d", len(result))
	}

	// Verify spike properties
	for _, spike := range result {
		if spike.Magnitude <= 2.0 {
			t.Errorf("Spike magnitude should be > 2.0, got %f", spike.Magnitude)
		}
		if spike.TimeOffset < 0 {
			t.Errorf("Spike time offset should be >= 0, got %d", spike.TimeOffset)
		}
		if spike.Metric != "power" && spike.Metric != "heart_rate" {
			t.Errorf("Unexpected spike metric: %s", spike.Metric)
		}
	}
}