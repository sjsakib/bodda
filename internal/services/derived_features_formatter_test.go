package services

import (
	"strings"
	"testing"
)

func TestFormatDerivedFeatures(t *testing.T) {
	formatter := NewOutputFormatter()

	// Test with nil data
	result := formatter.FormatDerivedFeatures(nil)
	if !strings.Contains(result, "No derived features data available") {
		t.Error("Expected 'No derived features data available' message for nil input")
	}

	// Test with valid derived features data
	features := &DerivedFeatures{
		ActivityID: 12345,
		Summary: FeatureSummary{
			TotalDataPoints:   1000,
			Duration:          3600, // 1 hour
			TotalDistance:     10000, // 10km
			ElevationGain:     200,
			AvgSpeed:          2.78, // 10 km/h in m/s
			MaxSpeed:          5.56, // 20 km/h in m/s
			AvgHeartRate:      150,
			MaxHeartRate:      180,
			AvgPower:          200,
			MaxPower:          400,
			MovingTimePercent: 95.0,
			StreamTypes:       []string{"time", "heartrate", "power", "speed"},
		},
		Statistics: StreamStatistics{
			HeartRate: &MetricStats{
				Min:         120,
				Max:         180,
				Mean:        150,
				Median:      148,
				StdDev:      15,
				Variability: 0.1,
				Q25:         140,
				Q75:         160,
				Count:       1000,
			},
			Power: &MetricStats{
				Min:         50,
				Max:         400,
				Mean:        200,
				Median:      195,
				StdDev:      50,
				Variability: 0.25,
				Q25:         160,
				Q75:         240,
				Count:       1000,
			},
		},
		Trends: []Trend{
			{
				Metric:     "heartrate",
				Direction:  "increasing",
				StartTime:  0,
				EndTime:    1800,
				Magnitude:  10,
				Confidence: 0.85,
			},
		},
		InflectionPoints: []InflectionPoint{
			{
				Time:      1800,
				Value:     160,
				Metric:    "heartrate",
				Direction: "peak",
				Magnitude: 2.5,
			},
		},
		Spikes: []Spike{
			{
				Time:      2700,
				Value:     380,
				Metric:    "power",
				Magnitude: 3.6,
				Duration:  30,
			},
		},
		SampleData: []DataPoint{
			{
				TimeOffset: 900,
				Values: map[string]interface{}{
					"heartrate": 155,
					"power":     220,
					"speed":     3.2,
				},
			},
		},
	}

	result = formatter.FormatDerivedFeatures(features)

	// Test that the result contains expected sections
	expectedSections := []string{
		"📊 **Stream Analysis** (Activity ID: 12345)",
		"## 📈 **Overview**",
		"## 📊 **Statistical Analysis**",
		"### 💓 **Heart Rate Analysis**",
		"### ⚡ **Power Analysis**",
	}

	for _, section := range expectedSections {
		if !strings.Contains(result, section) {
			t.Errorf("Expected section '%s' not found in output", section)
		}
	}

	// Test specific content formatting
	if !strings.Contains(result, "Duration:** 01:00:00 (1000 data points)") {
		t.Error("Duration formatting incorrect")
	}
	if !strings.Contains(result, "Distance:** 10.00km") {
		t.Error("Distance formatting incorrect")
	}
	if !strings.Contains(result, "200m elevation gain") {
		t.Error("Elevation gain not formatted correctly")
	}
	if !strings.Contains(result, "Range:** 120.0 - 180.0 bpm") {
		t.Error("Heart rate range not formatted correctly")
	}
}

func TestFormatDerivedFeaturesWithLapAnalysis(t *testing.T) {
	formatter := NewOutputFormatter()

	features := &DerivedFeatures{
		ActivityID: 12345,
		Summary: FeatureSummary{
			TotalDataPoints: 500,
			Duration:        1800, // 30 minutes
			StreamTypes:     []string{"time", "heartrate"},
		},
		Statistics: StreamStatistics{},
		LapAnalysis: &LapAnalysis{
			TotalLaps:        2,
			SegmentationType: "laps",
			LapSummaries: []LapSummary{
				{
					LapNumber:    1,
					Duration:     900,
					Distance:     5000,
					AvgSpeed:     5.56,
					AvgHeartRate: 145,
					AvgPower:     180,
				},
				{
					LapNumber:    2,
					Duration:     900,
					Distance:     5000,
					AvgSpeed:     5.56,
					AvgHeartRate: 150,
					AvgPower:     190,
				},
			},
			LapComparisons: LapComparisons{
				FastestLap:       1,
				SlowestLap:       2,
				HighestPowerLap:  2,
				ConsistencyScore: 0.85,
				SpeedVariation:   0.05,
			},
		},
	}

	result := formatter.FormatDerivedFeatures(features)

	// Test that lap analysis section is included
	expectedLapSections := []string{
		"## 🏁 **Lap-by-Lap Analysis**",
		"**Segmentation:** laps (2 segments)",
		"### 📊 **Lap Performance Summary**",
		"### 🏆 **Lap Comparisons**",
		"**Lap 1:** 5.00km in 15:00",
		"**Lap 2:** 5.00km in 15:00",
		"**Fastest Lap:** Lap 1",
		"**Consistency Score:** 8.5/10",
	}

	for _, section := range expectedLapSections {
		if !strings.Contains(result, section) {
			t.Errorf("Expected lap section '%s' not found in output", section)
		}
	}
}

func TestFormatDerivedFeaturesEmptyData(t *testing.T) {
	formatter := NewOutputFormatter()

	// Test with empty derived features
	features := &DerivedFeatures{
		ActivityID: 12345,
		Summary: FeatureSummary{
			TotalDataPoints: 0,
			Duration:        0,
			StreamTypes:     []string{},
		},
		Statistics: StreamStatistics{},
		Trends:     []Trend{},
		Spikes:     []Spike{},
		SampleData: []DataPoint{},
	}

	result := formatter.FormatDerivedFeatures(features)

	// Should still contain basic structure
	if !strings.Contains(result, "📊 **Stream Analysis** (Activity ID: 12345)") {
		t.Error("Should contain activity ID header")
	}
	if !strings.Contains(result, "## 📈 **Overview**") {
		t.Error("Should contain overview section")
	}
	if !strings.Contains(result, "Duration:** 00:00 (0 data points)") {
		t.Error("Should format zero duration correctly")
	}
}