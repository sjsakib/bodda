package services

import (
	"math"
	"testing"
)

func TestDetectInflectionPoints(t *testing.T) {
	tests := []struct {
		name      string
		data      []float64
		timeData  []int
		metric    string
		threshold float64
		minPoints int
	}{
		{
			name:      "empty data",
			data:      []float64{},
			timeData:  []int{},
			metric:    "test",
			threshold: 1.0,
			minPoints: 0,
		},
		{
			name:      "heart rate with peak",
			data:      []float64{120, 125, 130, 140, 150, 145, 140, 135, 130, 125, 120},
			timeData:  []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
			metric:    "heartrate",
			threshold: 0.5,
			minPoints: 1,
		},
		{
			name:      "power with spikes",
			data:      []float64{100, 105, 110, 200, 250, 120, 115, 110, 105, 100},
			timeData:  []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90},
			metric:    "power",
			threshold: 1.0,
			minPoints: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points := DetectInflectionPoints(tt.data, tt.timeData, tt.metric, tt.threshold)
			
			if len(points) < tt.minPoints {
				t.Errorf("Expected at least %d inflection points, got %d", tt.minPoints, len(points))
			}

			// Verify all points have valid data
			for _, point := range points {
				if point.Index < 0 || point.Index >= len(tt.data) {
					t.Errorf("Invalid index %d for data length %d", point.Index, len(tt.data))
				}
				if point.Metric != tt.metric {
					t.Errorf("Expected metric %s, got %s", tt.metric, point.Metric)
				}
				if point.Direction == "" {
					t.Errorf("Direction should not be empty")
				}
			}
		})
	}
}

func TestDetectSpikes(t *testing.T) {
	tests := []struct {
		name         string
		data         []float64
		timeData     []int
		metric       string
		threshold    float64
		expectedMin  int
	}{
		{
			name:         "empty data",
			data:         []float64{},
			timeData:     []int{},
			metric:       "test",
			threshold:    2.0,
			expectedMin:  0,
		},
		{
			name:         "normal heart rate",
			data:         []float64{140, 142, 141, 143, 142, 141, 140, 139},
			timeData:     []int{0, 10, 20, 30, 40, 50, 60, 70},
			metric:       "heartrate",
			threshold:    2.0,
			expectedMin:  0,
		},
		{
			name:         "power with clear spike",
			data:         []float64{100, 105, 110, 500, 115, 110, 105, 100},
			timeData:     []int{0, 10, 20, 30, 40, 50, 60, 70},
			metric:       "power",
			threshold:    2.0,
			expectedMin:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spikes := DetectSpikes(tt.data, tt.timeData, tt.metric, tt.threshold)
			
			if len(spikes) < tt.expectedMin {
				t.Errorf("Expected at least %d spikes, got %d", tt.expectedMin, len(spikes))
			}

			// Verify spike data
			for _, spike := range spikes {
				if spike.Index < 0 || spike.Index >= len(tt.data) {
					t.Errorf("Invalid spike index %d", spike.Index)
				}
				if spike.Magnitude <= 0 {
					t.Errorf("Spike magnitude should be positive, got %f", spike.Magnitude)
				}
				if spike.Metric != tt.metric {
					t.Errorf("Expected metric %s, got %s", tt.metric, spike.Metric)
				}
			}
		})
	}
}

func TestAnalyzeTrends(t *testing.T) {
	tests := []struct {
		name         string
		data         []float64
		timeData     []int
		metric       string
		windowSize   int
		expectedMin  int
	}{
		{
			name:         "insufficient data",
			data:         []float64{1, 2, 3},
			timeData:     []int{0, 10, 20},
			metric:       "test",
			windowSize:   5,
			expectedMin:  0,
		},
		{
			name:         "increasing trend",
			data:         []float64{100, 105, 110, 115, 120, 125, 130, 135, 140, 145, 150},
			timeData:     []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
			metric:       "power",
			windowSize:   3,
			expectedMin:  1,
		},
		{
			name:         "mixed trends",
			data:         []float64{100, 110, 120, 130, 125, 120, 115, 110, 115, 120, 125},
			timeData:     []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
			metric:       "heartrate",
			windowSize:   3,
			expectedMin:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trends := AnalyzeTrends(tt.data, tt.timeData, tt.metric, tt.windowSize)
			
			if len(trends) < tt.expectedMin {
				t.Errorf("Expected at least %d trends, got %d", tt.expectedMin, len(trends))
			}

			// Verify trend data
			for _, trend := range trends {
				if trend.StartIndex < 0 || trend.EndIndex >= len(tt.data) {
					t.Errorf("Invalid trend indices: start=%d, end=%d", trend.StartIndex, trend.EndIndex)
				}
				if trend.StartIndex >= trend.EndIndex {
					t.Errorf("Start index should be less than end index")
				}
				if trend.Confidence < 0 || trend.Confidence > 1 {
					t.Errorf("Confidence should be between 0 and 1, got %f", trend.Confidence)
				}
				if trend.Direction == "" {
					t.Errorf("Direction should not be empty")
				}
			}
		})
	}
}

func TestCalculateElevationAnalysis(t *testing.T) {
	tests := []struct {
		name     string
		altitude []float64
		distance []float64
		timeData []int
		expected *ElevationAnalysis
	}{
		{
			name:     "empty data",
			altitude: []float64{},
			distance: []float64{},
			timeData: []int{},
			expected: &ElevationAnalysis{},
		},
		{
			name:     "simple climb",
			altitude: []float64{100, 105, 110, 115, 120, 115, 110, 105, 100},
			distance: []float64{0, 100, 200, 300, 400, 500, 600, 700, 800},
			timeData: []int{0, 10, 20, 30, 40, 50, 60, 70, 80},
			expected: &ElevationAnalysis{
				TotalGain: 20,
				TotalLoss: 20,
				NetElevation: 0,
			},
		},
		{
			name:     "net elevation gain",
			altitude: []float64{100, 110, 120, 130, 140},
			distance: []float64{0, 100, 200, 300, 400},
			timeData: []int{0, 10, 20, 30, 40},
			expected: &ElevationAnalysis{
				TotalGain: 40,
				TotalLoss: 0,
				NetElevation: 40,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateElevationAnalysis(tt.altitude, tt.distance, tt.timeData)
			
			if math.Abs(result.TotalGain-tt.expected.TotalGain) > 0.001 {
				t.Errorf("TotalGain = %f, want %f", result.TotalGain, tt.expected.TotalGain)
			}
			if math.Abs(result.TotalLoss-tt.expected.TotalLoss) > 0.001 {
				t.Errorf("TotalLoss = %f, want %f", result.TotalLoss, tt.expected.TotalLoss)
			}
			if math.Abs(result.NetElevation-tt.expected.NetElevation) > 0.001 {
				t.Errorf("NetElevation = %f, want %f", result.NetElevation, tt.expected.NetElevation)
			}
		})
	}
}

func TestCalculateNormalizedPower(t *testing.T) {
	tests := []struct {
		name     string
		power    []int
		timeData []int
		expected float64
		tolerance float64
	}{
		{
			name:     "insufficient data",
			power:    []int{100, 200},
			timeData: []int{0, 10},
			expected: 0,
			tolerance: 0,
		},
		{
			name:     "constant power",
			power:    make([]int, 60), // 60 seconds of data
			timeData: make([]int, 60),
			expected: 200,
			tolerance: 1,
		},
		{
			name:     "variable power",
			power:    []int{100, 150, 200, 250, 300, 250, 200, 150, 100, 150, 200, 250, 300, 250, 200, 150, 100, 150, 200, 250, 300, 250, 200, 150, 100, 150, 200, 250, 300, 250, 200, 150, 100, 150, 200, 250, 300, 250, 200, 150},
			timeData: make([]int, 40),
			expected: 200,
			tolerance: 50,
		},
	}

	// Initialize constant power test data
	for i := range tests[1].power {
		tests[1].power[i] = 200
		tests[1].timeData[i] = i
	}

	// Initialize variable power test time data
	for i := range tests[2].timeData {
		tests[2].timeData[i] = i
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateNormalizedPower(tt.power, tt.timeData)
			
			if tt.tolerance == 0 {
				if result != tt.expected {
					t.Errorf("NormalizedPower = %f, want %f", result, tt.expected)
				}
			} else {
				if math.Abs(result-tt.expected) > tt.tolerance {
					t.Errorf("NormalizedPower = %f, want %f ± %f", result, tt.expected, tt.tolerance)
				}
			}
		})
	}
}

func TestCalculateHeartRateDrift(t *testing.T) {
	tests := []struct {
		name      string
		heartRate []int
		timeData  []int
		expected  float64
		tolerance float64
	}{
		{
			name:      "insufficient data",
			heartRate: []int{140, 142},
			timeData:  []int{0, 10},
			expected:  0,
			tolerance: 0,
		},
		{
			name:      "no drift",
			heartRate: []int{140, 140, 140, 140, 140, 140, 140, 140, 140, 140},
			timeData:  []int{0, 60, 120, 180, 240, 300, 360, 420, 480, 540},
			expected:  0,
			tolerance: 1,
		},
		{
			name:      "positive drift",
			heartRate: []int{140, 142, 144, 146, 148, 150, 152, 154, 156, 158},
			timeData:  []int{0, 60, 120, 180, 240, 300, 360, 420, 480, 540},
			expected:  120, // 18 bpm over 9 minutes = 2 bpm/min = 120 bpm/hour
			tolerance: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateHeartRateDrift(tt.heartRate, tt.timeData)
			
			if tt.tolerance == 0 {
				if result != tt.expected {
					t.Errorf("HeartRateDrift = %f, want %f", result, tt.expected)
				}
			} else {
				if math.Abs(result-tt.expected) > tt.tolerance {
					t.Errorf("HeartRateDrift = %f, want %f ± %f", result, tt.expected, tt.tolerance)
				}
			}
		})
	}
}

func TestCalculateCorrelations(t *testing.T) {
	tests := []struct {
		name     string
		streams  *StravaStreams
		expected *CorrelationAnalysis
	}{
		{
			name:     "empty streams",
			streams:  &StravaStreams{},
			expected: &CorrelationAnalysis{},
		},
		{
			name: "perfect positive correlation",
			streams: &StravaStreams{
				Watts:     []int{100, 200, 300, 400, 500},
				Heartrate: []int{120, 140, 160, 180, 200},
			},
			expected: &CorrelationAnalysis{
				PowerHeartRate: 1.0,
			},
		},
		{
			name: "mixed correlations",
			streams: &StravaStreams{
				Watts:          []int{100, 150, 200, 250, 300},
				Heartrate:      []int{120, 130, 140, 150, 160},
				VelocitySmooth: []float64{10, 12, 14, 16, 18},
				Cadence:        []int{80, 85, 90, 95, 100},
			},
			expected: &CorrelationAnalysis{
				PowerHeartRate: 1.0,
				SpeedHeartRate: 1.0,
				CadencePower:   1.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCorrelations(tt.streams)
			
			tolerance := 0.1
			
			if math.Abs(result.PowerHeartRate-tt.expected.PowerHeartRate) > tolerance {
				t.Errorf("PowerHeartRate = %f, want %f", result.PowerHeartRate, tt.expected.PowerHeartRate)
			}
			if math.Abs(result.SpeedHeartRate-tt.expected.SpeedHeartRate) > tolerance {
				t.Errorf("SpeedHeartRate = %f, want %f", result.SpeedHeartRate, tt.expected.SpeedHeartRate)
			}
			if math.Abs(result.CadencePower-tt.expected.CadencePower) > tolerance {
				t.Errorf("CadencePower = %f, want %f", result.CadencePower, tt.expected.CadencePower)
			}
		})
	}
}

func TestCalculateSlope(t *testing.T) {
	tests := []struct {
		name     string
		yData    []float64
		xData    []int
		expected float64
		tolerance float64
	}{
		{
			name:     "empty data",
			yData:    []float64{},
			xData:    []int{},
			expected: 0,
			tolerance: 0,
		},
		{
			name:     "positive slope",
			yData:    []float64{1, 2, 3, 4, 5},
			xData:    []int{0, 1, 2, 3, 4},
			expected: 1.0,
			tolerance: 0.001,
		},
		{
			name:     "negative slope",
			yData:    []float64{5, 4, 3, 2, 1},
			xData:    []int{0, 1, 2, 3, 4},
			expected: -1.0,
			tolerance: 0.001,
		},
		{
			name:     "zero slope",
			yData:    []float64{3, 3, 3, 3, 3},
			xData:    []int{0, 1, 2, 3, 4},
			expected: 0,
			tolerance: 0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateSlope(tt.yData, tt.xData)
			
			if tt.tolerance == 0 {
				if result != tt.expected {
					t.Errorf("calculateSlope = %f, want %f", result, tt.expected)
				}
			} else {
				if math.Abs(result-tt.expected) > tt.tolerance {
					t.Errorf("calculateSlope = %f, want %f ± %f", result, tt.expected, tt.tolerance)
				}
			}
		})
	}
}

func TestCalculateMovingAverage(t *testing.T) {
	tests := []struct {
		name       string
		data       []float64
		windowSize int
		expected   []float64
	}{
		{
			name:       "insufficient data",
			data:       []float64{1, 2},
			windowSize: 3,
			expected:   []float64{},
		},
		{
			name:       "simple moving average",
			data:       []float64{1, 2, 3, 4, 5},
			windowSize: 3,
			expected:   []float64{2, 3, 4},
		},
		{
			name:       "window size 1",
			data:       []float64{1, 2, 3, 4, 5},
			windowSize: 1,
			expected:   []float64{1, 2, 3, 4, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMovingAverage(tt.data, tt.windowSize)
			
			if len(result) != len(tt.expected) {
				t.Errorf("Length = %d, want %d", len(result), len(tt.expected))
				return
			}

			for i, expected := range tt.expected {
				if math.Abs(result[i]-expected) > 0.001 {
					t.Errorf("MovingAverage[%d] = %f, want %f", i, result[i], expected)
				}
			}
		})
	}
}

func TestCalculateCorrelation(t *testing.T) {
	tests := []struct {
		name      string
		x         []float64
		y         []float64
		expected  float64
		tolerance float64
	}{
		{
			name:      "perfect positive correlation",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{2, 4, 6, 8, 10},
			expected:  1.0,
			tolerance: 0.001,
		},
		{
			name:      "perfect negative correlation",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{10, 8, 6, 4, 2},
			expected:  -1.0,
			tolerance: 0.001,
		},
		{
			name:      "no correlation",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{3, 3, 3, 3, 3},
			expected:  0.0,
			tolerance: 0.001,
		},
		{
			name:      "insufficient data",
			x:         []float64{1},
			y:         []float64{2},
			expected:  0.0,
			tolerance: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateCorrelation(tt.x, tt.y)
			
			if tt.tolerance == 0 {
				if result != tt.expected {
					t.Errorf("calculateCorrelation = %f, want %f", result, tt.expected)
				}
			} else {
				if math.Abs(result-tt.expected) > tt.tolerance {
					t.Errorf("calculateCorrelation = %f, want %f ± %f", result, tt.expected, tt.tolerance)
				}
			}
		})
	}
}