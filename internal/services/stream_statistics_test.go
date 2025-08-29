package services

import (
	"math"
	"testing"
)

func TestCalculateIntStats(t *testing.T) {
	tests := []struct {
		name     string
		data     []int
		expected *MetricStats
	}{
		{
			name: "empty slice",
			data: []int{},
			expected: &MetricStats{Count: 0},
		},
		{
			name: "single value",
			data: []int{100},
			expected: &MetricStats{
				Min:         100,
				Max:         100,
				Mean:        100,
				Median:      100,
				StdDev:      0,
				Variability: 0,
				Range:       0,
				Q25:         100,
				Q75:         100,
				Count:       1,
			},
		},
		{
			name: "heart rate data",
			data: []int{120, 130, 140, 150, 160},
			expected: &MetricStats{
				Min:         120,
				Max:         160,
				Mean:        140,
				Median:      140,
				Range:       40,
				Q25:         130,
				Q75:         150,
				Count:       5,
			},
		},
		{
			name: "power data with zeros",
			data: []int{0, 100, 200, 0, 300, 250},
			expected: &MetricStats{
				Min:    100,
				Max:    300,
				Mean:   212.5,
				Median: 225,
				Range:  200,
				Q25:    175,
				Q75:    275,
				Count:  4, // zeros filtered out
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateIntStats(tt.data)
			
			if result.Count != tt.expected.Count {
				t.Errorf("Count = %d, want %d", result.Count, tt.expected.Count)
			}
			
			if tt.expected.Count == 0 {
				return // Skip other checks for empty data
			}
			
			if math.Abs(result.Min-tt.expected.Min) > 0.001 {
				t.Errorf("Min = %f, want %f", result.Min, tt.expected.Min)
			}
			if math.Abs(result.Max-tt.expected.Max) > 0.001 {
				t.Errorf("Max = %f, want %f", result.Max, tt.expected.Max)
			}
			if math.Abs(result.Mean-tt.expected.Mean) > 0.001 {
				t.Errorf("Mean = %f, want %f", result.Mean, tt.expected.Mean)
			}
			if math.Abs(result.Median-tt.expected.Median) > 0.001 {
				t.Errorf("Median = %f, want %f", result.Median, tt.expected.Median)
			}
		})
	}
}

func TestCalculateFloatStats(t *testing.T) {
	tests := []struct {
		name     string
		data     []float64
		expected *MetricStats
	}{
		{
			name: "empty slice",
			data: []float64{},
			expected: &MetricStats{Count: 0},
		},
		{
			name: "speed data",
			data: []float64{10.5, 12.3, 15.7, 18.2, 20.1},
			expected: &MetricStats{
				Min:    10.5,
				Max:    20.1,
				Mean:   15.36,
				Median: 15.7,
				Range:  9.6,
				Count:  5,
			},
		},
		{
			name: "altitude data with variation",
			data: []float64{100.0, 105.5, 98.2, 110.3, 95.7, 108.1},
			expected: &MetricStats{
				Min:    95.7,
				Max:    110.3,
				Range:  14.6,
				Count:  6,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateFloatStats(tt.data)
			
			if result.Count != tt.expected.Count {
				t.Errorf("Count = %d, want %d", result.Count, tt.expected.Count)
			}
			
			if tt.expected.Count == 0 {
				return
			}
			
			if math.Abs(result.Min-tt.expected.Min) > 0.001 {
				t.Errorf("Min = %f, want %f", result.Min, tt.expected.Min)
			}
			if math.Abs(result.Max-tt.expected.Max) > 0.001 {
				t.Errorf("Max = %f, want %f", result.Max, tt.expected.Max)
			}
			if math.Abs(result.Range-tt.expected.Range) > 0.001 {
				t.Errorf("Range = %f, want %f", result.Range, tt.expected.Range)
			}
		})
	}
}

func TestCalculateBooleanStats(t *testing.T) {
	tests := []struct {
		name     string
		data     []bool
		expected *BooleanStats
	}{
		{
			name: "empty slice",
			data: []bool{},
			expected: &BooleanStats{TotalCount: 0},
		},
		{
			name: "all true",
			data: []bool{true, true, true, true},
			expected: &BooleanStats{
				TrueCount:    4,
				FalseCount:   0,
				TotalCount:   4,
				TruePercent:  100.0,
				FalsePercent: 0.0,
			},
		},
		{
			name: "mixed moving data",
			data: []bool{true, true, false, true, false, false, true},
			expected: &BooleanStats{
				TrueCount:    4,
				FalseCount:   3,
				TotalCount:   7,
				TruePercent:  57.14285714285714,
				FalsePercent: 42.857142857142854,
			},
		},
		{
			name: "all false",
			data: []bool{false, false, false},
			expected: &BooleanStats{
				TrueCount:    0,
				FalseCount:   3,
				TotalCount:   3,
				TruePercent:  0.0,
				FalsePercent: 100.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateBooleanStats(tt.data)
			
			if result.TrueCount != tt.expected.TrueCount {
				t.Errorf("TrueCount = %d, want %d", result.TrueCount, tt.expected.TrueCount)
			}
			if result.FalseCount != tt.expected.FalseCount {
				t.Errorf("FalseCount = %d, want %d", result.FalseCount, tt.expected.FalseCount)
			}
			if result.TotalCount != tt.expected.TotalCount {
				t.Errorf("TotalCount = %d, want %d", result.TotalCount, tt.expected.TotalCount)
			}
			if math.Abs(result.TruePercent-tt.expected.TruePercent) > 0.001 {
				t.Errorf("TruePercent = %f, want %f", result.TruePercent, tt.expected.TruePercent)
			}
			if math.Abs(result.FalsePercent-tt.expected.FalsePercent) > 0.001 {
				t.Errorf("FalsePercent = %f, want %f", result.FalsePercent, tt.expected.FalsePercent)
			}
		})
	}
}

func TestCalculateLocationStats(t *testing.T) {
	tests := []struct {
		name     string
		data     [][]float64
		expected *LocationStats
	}{
		{
			name: "empty slice",
			data: [][]float64{},
			expected: &LocationStats{TotalPoints: 0},
		},
		{
			name: "single coordinate",
			data: [][]float64{{40.7128, -74.0060}}, // NYC
			expected: &LocationStats{
				StartLat:    40.7128,
				StartLng:    -74.0060,
				EndLat:      40.7128,
				EndLng:      -74.0060,
				TotalPoints: 1,
				BoundingBox: BoundingBox{
					NorthLat: 40.7128,
					SouthLat: 40.7128,
					EastLng:  -74.0060,
					WestLng:  -74.0060,
				},
			},
		},
		{
			name: "route coordinates",
			data: [][]float64{
				{40.7128, -74.0060}, // NYC
				{40.7589, -73.9851}, // Times Square
				{40.7831, -73.9712}, // Central Park
				{40.7505, -73.9934}, // Empire State
			},
			expected: &LocationStats{
				StartLat:    40.7128,
				StartLng:    -74.0060,
				EndLat:      40.7505,
				EndLng:      -73.9934,
				TotalPoints: 4,
				BoundingBox: BoundingBox{
					NorthLat: 40.7831,
					SouthLat: 40.7128,
					EastLng:  -73.9712,
					WestLng:  -74.0060,
				},
			},
		},
		{
			name: "coordinates with zeros filtered",
			data: [][]float64{
				{0, 0},               // Should be filtered
				{40.7128, -74.0060},  // NYC
				{40.7589, -73.9851},  // Times Square
				{0, 0},               // Should be filtered
			},
			expected: &LocationStats{
				StartLat:    40.7128,
				StartLng:    -74.0060,
				EndLat:      40.7589,
				EndLng:      -73.9851,
				TotalPoints: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateLocationStats(tt.data)
			
			if result.TotalPoints != tt.expected.TotalPoints {
				t.Errorf("TotalPoints = %d, want %d", result.TotalPoints, tt.expected.TotalPoints)
			}
			
			if tt.expected.TotalPoints == 0 {
				return
			}
			
			if math.Abs(result.StartLat-tt.expected.StartLat) > 0.0001 {
				t.Errorf("StartLat = %f, want %f", result.StartLat, tt.expected.StartLat)
			}
			if math.Abs(result.StartLng-tt.expected.StartLng) > 0.0001 {
				t.Errorf("StartLng = %f, want %f", result.StartLng, tt.expected.StartLng)
			}
			if math.Abs(result.EndLat-tt.expected.EndLat) > 0.0001 {
				t.Errorf("EndLat = %f, want %f", result.EndLat, tt.expected.EndLat)
			}
			if math.Abs(result.EndLng-tt.expected.EndLng) > 0.0001 {
				t.Errorf("EndLng = %f, want %f", result.EndLng, tt.expected.EndLng)
			}
		})
	}
}

func TestCalculatePercentile(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	
	tests := []struct {
		percentile float64
		expected   float64
	}{
		{0.0, 1.0},
		{0.25, 3.25},
		{0.5, 5.5},
		{0.75, 7.75},
		{1.0, 10.0},
	}

	for _, tt := range tests {
		result := calculatePercentile(data, tt.percentile)
		if math.Abs(result-tt.expected) > 0.001 {
			t.Errorf("calculatePercentile(%f) = %f, want %f", tt.percentile, result, tt.expected)
		}
	}
}

func TestCalculateQuartiles(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	
	q1, q2, q3 := CalculateQuartiles(data)
	
	expectedQ1 := 3.25
	expectedQ2 := 5.5
	expectedQ3 := 7.75
	
	if math.Abs(q1-expectedQ1) > 0.001 {
		t.Errorf("Q1 = %f, want %f", q1, expectedQ1)
	}
	if math.Abs(q2-expectedQ2) > 0.001 {
		t.Errorf("Q2 = %f, want %f", q2, expectedQ2)
	}
	if math.Abs(q3-expectedQ3) > 0.001 {
		t.Errorf("Q3 = %f, want %f", q3, expectedQ3)
	}
}

func TestCalculateVariabilityMetrics(t *testing.T) {
	data := []float64{10, 12, 14, 16, 18, 20}
	
	cv, iqr, mad := CalculateVariabilityMetrics(data)
	
	// Expected values calculated manually
	expectedIQR := 5.0 // Q3(17.5) - Q1(12.5) = 5.0
	
	if math.Abs(iqr-expectedIQR) > 0.1 {
		t.Errorf("IQR = %f, want %f", iqr, expectedIQR)
	}
	
	// CV should be positive for non-zero data
	if cv <= 0 {
		t.Errorf("CV = %f, should be positive", cv)
	}
	
	// MAD should be positive for varied data
	if mad <= 0 {
		t.Errorf("MAD = %f, should be positive", mad)
	}
}

func TestStatisticsWithRealWorldData(t *testing.T) {
	// Simulate realistic heart rate data during a workout
	heartRateData := []int{
		65, 68, 72, 78, 85, 92, 98, 105, 112, 118, 125, 132, 138, 145, 152,
		158, 165, 162, 159, 156, 153, 150, 147, 144, 141, 138, 135, 132, 129,
		126, 123, 120, 117, 114, 111, 108, 105, 102, 99, 96, 93, 90, 87, 84,
	}
	
	stats := CalculateIntStats(heartRateData)
	
	// Verify reasonable values for heart rate data
	if stats.Min < 60 || stats.Min > 70 {
		t.Errorf("Min heart rate %f seems unrealistic", stats.Min)
	}
	if stats.Max < 160 || stats.Max > 170 {
		t.Errorf("Max heart rate %f seems unrealistic", stats.Max)
	}
	if stats.Mean < 110 || stats.Mean > 130 {
		t.Errorf("Mean heart rate %f seems unrealistic", stats.Mean)
	}
	if stats.Count != len(heartRateData) {
		t.Errorf("Count = %d, want %d", stats.Count, len(heartRateData))
	}
}