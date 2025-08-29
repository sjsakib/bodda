package services

import (
	"math"
)

// InflectionPoint represents a significant change in data trend
type InflectionPoint struct {
	Index     int     `json:"index"`
	Time      int     `json:"time"`
	Value     float64 `json:"value"`
	Metric    string  `json:"metric"`
	Direction string  `json:"direction"` // "increase", "decrease", "peak", "valley"
	Magnitude float64 `json:"magnitude"` // Change magnitude
}

// Spike represents a significant deviation from normal values
type Spike struct {
	Index     int     `json:"index"`
	Time      int     `json:"time"`
	Value     float64 `json:"value"`
	Metric    string  `json:"metric"`
	Magnitude float64 `json:"magnitude"` // How many standard deviations from mean
	Duration  int     `json:"duration"`  // Duration in seconds
}

// Trend represents a directional pattern in the data
type Trend struct {
	StartIndex int     `json:"start_index"`
	EndIndex   int     `json:"end_index"`
	StartTime  int     `json:"start_time"`
	EndTime    int     `json:"end_time"`
	Metric     string  `json:"metric"`
	Direction  string  `json:"direction"` // "increasing", "decreasing", "stable"
	Slope      float64 `json:"slope"`
	Magnitude  float64 `json:"magnitude"`
	Confidence float64 `json:"confidence"` // 0-1, how confident we are in this trend
}

// ElevationAnalysis represents elevation-specific calculations
type ElevationAnalysis struct {
	TotalGain     float64 `json:"total_gain"`
	TotalLoss     float64 `json:"total_loss"`
	NetElevation  float64 `json:"net_elevation"`
	MaxGrade      float64 `json:"max_grade"`
	MinGrade      float64 `json:"min_grade"`
	AvgGrade      float64 `json:"avg_grade"`
	ClimbSegments []ClimbSegment `json:"climb_segments"`
}

// ClimbSegment represents a continuous climbing section
type ClimbSegment struct {
	StartIndex    int     `json:"start_index"`
	EndIndex      int     `json:"end_index"`
	StartTime     int     `json:"start_time"`
	EndTime       int     `json:"end_time"`
	ElevationGain float64 `json:"elevation_gain"`
	Distance      float64 `json:"distance"`
	AvgGrade      float64 `json:"avg_grade"`
	MaxGrade      float64 `json:"max_grade"`
}

// PowerAnalysis represents cycling power-specific calculations
type PowerAnalysis struct {
	NormalizedPower     float64 `json:"normalized_power"`
	IntensityFactor     float64 `json:"intensity_factor"`
	TrainingStressScore float64 `json:"training_stress_score"`
	VariabilityIndex    float64 `json:"variability_index"`
	PowerZones          PowerZoneDistribution `json:"power_zones"`
}

// PowerZoneDistribution represents time spent in different power zones
type PowerZoneDistribution struct {
	Zone1Percent float64 `json:"zone1_percent"` // Active Recovery
	Zone2Percent float64 `json:"zone2_percent"` // Endurance
	Zone3Percent float64 `json:"zone3_percent"` // Tempo
	Zone4Percent float64 `json:"zone4_percent"` // Lactate Threshold
	Zone5Percent float64 `json:"zone5_percent"` // VO2 Max
	Zone6Percent float64 `json:"zone6_percent"` // Anaerobic Capacity
	Zone7Percent float64 `json:"zone7_percent"` // Neuromuscular Power
}

// HeartRateAnalysis represents heart rate-specific calculations
type HeartRateAnalysis struct {
	Drift            float64 `json:"drift"`             // bpm per hour
	TimeInZones      HeartRateZoneDistribution `json:"time_in_zones"`
	RecoverySegments []RecoverySegment `json:"recovery_segments"`
}

// HeartRateZoneDistribution represents time spent in different HR zones
type HeartRateZoneDistribution struct {
	Zone1Percent float64 `json:"zone1_percent"` // Recovery
	Zone2Percent float64 `json:"zone2_percent"` // Aerobic Base
	Zone3Percent float64 `json:"zone3_percent"` // Aerobic
	Zone4Percent float64 `json:"zone4_percent"` // Lactate Threshold
	Zone5Percent float64 `json:"zone5_percent"` // VO2 Max
}

// RecoverySegment represents periods of heart rate recovery
type RecoverySegment struct {
	StartIndex   int     `json:"start_index"`
	EndIndex     int     `json:"end_index"`
	StartTime    int     `json:"start_time"`
	EndTime      int     `json:"end_time"`
	StartHR      int     `json:"start_hr"`
	EndHR        int     `json:"end_hr"`
	RecoveryRate float64 `json:"recovery_rate"` // bpm per minute
}

// CorrelationAnalysis represents relationships between different metrics
type CorrelationAnalysis struct {
	PowerHeartRate    float64 `json:"power_heart_rate"`
	SpeedHeartRate    float64 `json:"speed_heart_rate"`
	CadencePower      float64 `json:"cadence_power"`
	ElevationSpeed    float64 `json:"elevation_speed"`
	TemperatureHR     float64 `json:"temperature_hr"`
}

// DetectInflectionPoints finds significant changes in data trends
func DetectInflectionPoints(data []float64, timeData []int, metric string, threshold float64) []InflectionPoint {
	if len(data) < 5 || len(timeData) != len(data) {
		return []InflectionPoint{}
	}

	var points []InflectionPoint
	windowSize := 2 // Smaller window for more sensitivity

	for i := windowSize; i < len(data)-windowSize; i++ {
		// Calculate slopes before and after current point
		beforeSlope := calculateSlope(data[i-windowSize:i+1], timeData[i-windowSize:i+1])
		afterSlope := calculateSlope(data[i:i+windowSize+1], timeData[i:i+windowSize+1])

		// Check for significant change in slope
		slopeChange := math.Abs(afterSlope - beforeSlope)
		if slopeChange > threshold {
			direction := "stable"
			if beforeSlope < 0 && afterSlope > 0 {
				direction = "valley"
			} else if beforeSlope > 0 && afterSlope < 0 {
				direction = "peak"
			} else if afterSlope > beforeSlope {
				direction = "increase"
			} else {
				direction = "decrease"
			}

			points = append(points, InflectionPoint{
				Index:     i,
				Time:      timeData[i],
				Value:     data[i],
				Metric:    metric,
				Direction: direction,
				Magnitude: slopeChange,
			})
		}
	}

	return points
}

// DetectSpikes finds values that deviate significantly from the mean
func DetectSpikes(data []float64, timeData []int, metric string, stdDevThreshold float64) []Spike {
	if len(data) < 3 || len(timeData) != len(data) {
		return []Spike{}
	}

	stats := CalculateFloatStats(data)
	if stats.Count == 0 {
		return []Spike{}
	}

	var spikes []Spike

	for i := 0; i < len(data); i++ {
		if math.Abs(data[i]-stats.Mean) > stdDevThreshold*stats.StdDev {
			// Calculate spike duration
			duration := 1
			if i > 0 && i < len(timeData)-1 {
				duration = timeData[i+1] - timeData[i-1]
			}

			spikes = append(spikes, Spike{
				Index:     i,
				Time:      timeData[i],
				Value:     data[i],
				Metric:    metric,
				Magnitude: math.Abs(data[i]-stats.Mean) / stats.StdDev,
				Duration:  duration,
			})
		}
	}

	return spikes
}

// AnalyzeTrends identifies directional patterns using moving averages
func AnalyzeTrends(data []float64, timeData []int, metric string, windowSize int) []Trend {
	if len(data) < windowSize*2 || len(timeData) != len(data) {
		return []Trend{}
	}

	var trends []Trend
	movingAvg := calculateMovingAverage(data, windowSize)

	// Find trend segments
	currentTrend := &Trend{
		StartIndex: windowSize,
		StartTime:  timeData[windowSize],
		Metric:     metric,
	}

	for i := windowSize + 1; i < len(movingAvg)-1; i++ {
		slope := movingAvg[i+1] - movingAvg[i-1]
		direction := "stable"
		
		if slope > 0.1 {
			direction = "increasing"
		} else if slope < -0.1 {
			direction = "decreasing"
		}

		// If direction changes or we reach the end, close current trend
		if direction != currentTrend.Direction || i == len(movingAvg)-2 {
			if currentTrend.Direction != "" {
				currentTrend.EndIndex = i
				currentTrend.EndTime = timeData[i]
				currentTrend.Slope = calculateSlope(
					data[currentTrend.StartIndex:currentTrend.EndIndex+1],
					timeData[currentTrend.StartIndex:currentTrend.EndIndex+1],
				)
				currentTrend.Magnitude = math.Abs(data[currentTrend.EndIndex] - data[currentTrend.StartIndex])
				currentTrend.Confidence = calculateTrendConfidence(
					data[currentTrend.StartIndex:currentTrend.EndIndex+1],
				)

				trends = append(trends, *currentTrend)
			}

			// Start new trend
			currentTrend = &Trend{
				StartIndex: i,
				StartTime:  timeData[i],
				Metric:     metric,
				Direction:  direction,
			}
		} else {
			currentTrend.Direction = direction
		}
	}

	return trends
}

// CalculateElevationAnalysis performs elevation-specific calculations
func CalculateElevationAnalysis(altitude []float64, distance []float64, timeData []int) *ElevationAnalysis {
	if len(altitude) < 2 {
		return &ElevationAnalysis{}
	}

	analysis := &ElevationAnalysis{}
	var grades []float64

	// Calculate elevation gain/loss and grades
	for i := 1; i < len(altitude); i++ {
		elevDiff := altitude[i] - altitude[i-1]
		
		if elevDiff > 0 {
			analysis.TotalGain += elevDiff
		} else {
			analysis.TotalLoss += math.Abs(elevDiff)
		}

		// Calculate grade if distance data is available
		if len(distance) == len(altitude) && distance[i] > distance[i-1] {
			distDiff := distance[i] - distance[i-1]
			if distDiff > 0 {
				grade := (elevDiff / distDiff) * 100
				grades = append(grades, grade)
			}
		}
	}

	analysis.NetElevation = altitude[len(altitude)-1] - altitude[0]

	// Calculate grade statistics
	if len(grades) > 0 {
		gradeStats := CalculateFloatStats(grades)
		analysis.MaxGrade = gradeStats.Max
		analysis.MinGrade = gradeStats.Min
		analysis.AvgGrade = gradeStats.Mean

		// Find climb segments (sustained positive grades)
		analysis.ClimbSegments = findClimbSegments(altitude, distance, timeData, grades)
	}

	return analysis
}

// CalculateNormalizedPower calculates normalized power for cycling activities
func CalculateNormalizedPower(power []int, timeData []int) float64 {
	if len(power) < 30 {
		return 0
	}

	// Convert to float64 and calculate 30-second rolling average
	powerFloat := make([]float64, len(power))
	for i, p := range power {
		powerFloat[i] = float64(p)
	}

	rollingAvg := calculateMovingAverage(powerFloat, 30)
	
	// Raise each value to the 4th power
	var sum float64
	for _, avg := range rollingAvg {
		sum += math.Pow(avg, 4)
	}

	// Take the 4th root of the average
	if len(rollingAvg) > 0 {
		return math.Pow(sum/float64(len(rollingAvg)), 0.25)
	}

	return 0
}

// CalculateHeartRateDrift calculates heart rate drift over time
func CalculateHeartRateDrift(heartRate []int, timeData []int) float64 {
	if len(heartRate) < 10 || len(timeData) != len(heartRate) {
		return 0
	}

	// Convert to float64 for calculations
	hrFloat := make([]float64, len(heartRate))
	timeFloat := make([]float64, len(timeData))
	
	for i := range heartRate {
		hrFloat[i] = float64(heartRate[i])
		timeFloat[i] = float64(timeData[i])
	}

	// Calculate slope (bpm per second)
	timeInt := make([]int, len(timeData))
	for i, t := range timeData {
		timeInt[i] = t
	}
	slope := calculateSlope(hrFloat, timeInt)
	
	// Convert to bpm per hour
	return slope * 3600
}

// CalculateCorrelations analyzes relationships between different metrics
func CalculateCorrelations(streams *StravaStreams) *CorrelationAnalysis {
	analysis := &CorrelationAnalysis{}

	// Power vs Heart Rate correlation
	if len(streams.Watts) > 0 && len(streams.Heartrate) > 0 {
		powerFloat := make([]float64, len(streams.Watts))
		hrFloat := make([]float64, len(streams.Heartrate))
		
		minLen := len(streams.Watts)
		if len(streams.Heartrate) < minLen {
			minLen = len(streams.Heartrate)
		}

		for i := 0; i < minLen; i++ {
			powerFloat[i] = float64(streams.Watts[i])
			hrFloat[i] = float64(streams.Heartrate[i])
		}

		analysis.PowerHeartRate = calculateCorrelation(powerFloat[:minLen], hrFloat[:minLen])
	}

	// Speed vs Heart Rate correlation
	if len(streams.VelocitySmooth) > 0 && len(streams.Heartrate) > 0 {
		hrFloat := make([]float64, len(streams.Heartrate))
		
		minLen := len(streams.VelocitySmooth)
		if len(streams.Heartrate) < minLen {
			minLen = len(streams.Heartrate)
		}

		for i := 0; i < minLen; i++ {
			hrFloat[i] = float64(streams.Heartrate[i])
		}

		analysis.SpeedHeartRate = calculateCorrelation(streams.VelocitySmooth[:minLen], hrFloat[:minLen])
	}

	// Cadence vs Power correlation
	if len(streams.Cadence) > 0 && len(streams.Watts) > 0 {
		cadenceFloat := make([]float64, len(streams.Cadence))
		powerFloat := make([]float64, len(streams.Watts))
		
		minLen := len(streams.Cadence)
		if len(streams.Watts) < minLen {
			minLen = len(streams.Watts)
		}

		for i := 0; i < minLen; i++ {
			cadenceFloat[i] = float64(streams.Cadence[i])
			powerFloat[i] = float64(streams.Watts[i])
		}

		analysis.CadencePower = calculateCorrelation(cadenceFloat[:minLen], powerFloat[:minLen])
	}

	return analysis
}

// Helper functions

func calculateSlope(yData []float64, xData []int) float64 {
	if len(yData) != len(xData) || len(yData) < 2 {
		return 0
	}

	n := float64(len(yData))
	var sumX, sumY, sumXY, sumXX float64

	for i := 0; i < len(yData); i++ {
		x := float64(xData[i])
		y := yData[i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	denominator := n*sumXX - sumX*sumX
	if denominator == 0 {
		return 0
	}

	return (n*sumXY - sumX*sumY) / denominator
}

func calculateMovingAverage(data []float64, windowSize int) []float64 {
	if len(data) < windowSize {
		return []float64{}
	}

	result := make([]float64, len(data)-windowSize+1)
	
	for i := 0; i <= len(data)-windowSize; i++ {
		sum := 0.0
		for j := i; j < i+windowSize; j++ {
			sum += data[j]
		}
		result[i] = sum / float64(windowSize)
	}

	return result
}

func calculateTrendConfidence(data []float64) float64 {
	if len(data) < 3 {
		return 0
	}

	// Calculate R-squared for linear trend
	stats := CalculateFloatStats(data)
	if stats.StdDev == 0 {
		return 1.0 // Perfect trend if no variation
	}

	// Simple confidence based on consistency of direction
	increases := 0
	decreases := 0
	
	for i := 1; i < len(data); i++ {
		if data[i] > data[i-1] {
			increases++
		} else if data[i] < data[i-1] {
			decreases++
		}
	}

	total := increases + decreases
	if total == 0 {
		return 1.0
	}

	maxDirection := increases
	if decreases > increases {
		maxDirection = decreases
	}

	return float64(maxDirection) / float64(total)
}

func findClimbSegments(altitude []float64, distance []float64, timeData []int, grades []float64) []ClimbSegment {
	var segments []ClimbSegment
	
	if len(grades) == 0 {
		return segments
	}

	var currentSegment *ClimbSegment
	climbThreshold := 3.0 // 3% grade threshold

	for i, grade := range grades {
		if grade > climbThreshold {
			if currentSegment == nil {
				// Start new climb segment
				currentSegment = &ClimbSegment{
					StartIndex: i,
					StartTime:  timeData[i],
					AvgGrade:   grade,
					MaxGrade:   grade,
				}
			} else {
				// Continue current segment
				currentSegment.MaxGrade = math.Max(currentSegment.MaxGrade, grade)
			}
		} else {
			if currentSegment != nil {
				// End current segment
				currentSegment.EndIndex = i
				currentSegment.EndTime = timeData[i]
				currentSegment.ElevationGain = altitude[i] - altitude[currentSegment.StartIndex]
				
				if len(distance) > i && len(distance) > currentSegment.StartIndex {
					currentSegment.Distance = distance[i] - distance[currentSegment.StartIndex]
				}

				// Calculate average grade for the segment
				segmentGrades := grades[currentSegment.StartIndex:i]
				if len(segmentGrades) > 0 {
					sum := 0.0
					for _, g := range segmentGrades {
						sum += g
					}
					currentSegment.AvgGrade = sum / float64(len(segmentGrades))
				}

				segments = append(segments, *currentSegment)
				currentSegment = nil
			}
		}
	}

	// Handle case where climb continues to the end
	if currentSegment != nil {
		i := len(grades) - 1
		currentSegment.EndIndex = i
		currentSegment.EndTime = timeData[i]
		currentSegment.ElevationGain = altitude[i] - altitude[currentSegment.StartIndex]
		
		if len(distance) > i && len(distance) > currentSegment.StartIndex {
			currentSegment.Distance = distance[i] - distance[currentSegment.StartIndex]
		}

		segments = append(segments, *currentSegment)
	}

	return segments
}

func calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	n := float64(len(x))
	var sumX, sumY, sumXY, sumXX, sumYY float64

	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumXX += x[i] * x[i]
		sumYY += y[i] * y[i]
	}

	numerator := n*sumXY - sumX*sumY
	denominator := math.Sqrt((n*sumXX - sumX*sumX) * (n*sumYY - sumY*sumY))

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}