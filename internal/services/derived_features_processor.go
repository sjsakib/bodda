package services

import (
	"context"
	"fmt"
	"log"
)

// DerivedFeaturesProcessor interface defines methods for extracting derived features from stream data
type DerivedFeaturesProcessor interface {
	ExtractFeatures(data *StravaStreams, laps []StravaLap) (*DerivedFeatures, error)
	ExtractLapFeatures(data *StravaStreams, laps []StravaLap) (*LapAnalysis, error)
}

// derivedFeaturesProcessor implements the DerivedFeaturesProcessor interface
type derivedFeaturesProcessor struct{
	performanceMonitor *StreamPerformanceMonitor
}

// NewDerivedFeaturesProcessor creates a new derived features processor
func NewDerivedFeaturesProcessor() DerivedFeaturesProcessor {
	return &derivedFeaturesProcessor{
		performanceMonitor: NewStreamPerformanceMonitor(true),
	}
}

// ExtractFeatures extracts comprehensive derived features from stream data
func (dfp *derivedFeaturesProcessor) ExtractFeatures(data *StravaStreams, laps []StravaLap) (*DerivedFeatures, error) {
	if data == nil {
		return nil, fmt.Errorf("stream data is nil")
	}

	// Start performance monitoring
	dataSize := len(data.Time)
	timer := dfp.performanceMonitor.StartOperation(context.Background(), "extract_features", dataSize)
	defer func() {
		timer.EndOperation(nil)
	}()

	log.Printf("Extracting derived features from stream data with %d time points", dataSize)

	// Create the derived features structure
	features := &DerivedFeatures{
		Summary:          dfp.calculateFeatureSummary(data),
		Statistics:       dfp.calculateStreamStatistics(data),
		InflectionPoints: []InflectionPoint{},
		Trends:           []Trend{},
		Spikes:           []Spike{},
		SampleData:       []DataPoint{},
	}

	// Extract derived features
	features.InflectionPoints = dfp.extractInflectionPoints(data)
	features.Trends = dfp.extractTrends(data)
	features.Spikes = dfp.extractSpikes(data)
	features.SampleData = dfp.extractSampleData(data)

	// Add lap analysis if lap data is available
	if len(laps) > 0 {
		lapAnalysis, err := dfp.ExtractLapFeatures(data, laps)
		if err != nil {
			log.Printf("Warning: Failed to extract lap features: %v", err)
		} else {
			features.LapAnalysis = lapAnalysis
		}
	}

	log.Printf("Successfully extracted derived features with %d inflection points, %d trends, %d spikes", 
		len(features.InflectionPoints), len(features.Trends), len(features.Spikes))

	return features, nil
}

// ExtractLapFeatures extracts lap-specific analysis from stream data
func (dfp *derivedFeaturesProcessor) ExtractLapFeatures(data *StravaStreams, laps []StravaLap) (*LapAnalysis, error) {
	if data == nil || len(laps) == 0 {
		return nil, fmt.Errorf("insufficient data for lap analysis")
	}

	log.Printf("Extracting lap features for %d laps", len(laps))

	// This is a placeholder implementation - the full lap analysis would be implemented in task 3.3
	lapAnalysis := &LapAnalysis{
		TotalLaps:        len(laps),
		SegmentationType: "laps",
		LapSummaries:     []LapSummary{},
		LapComparisons:   LapComparisons{},
	}

	// Basic lap summaries (simplified for now)
	for i, lap := range laps {
		summary := LapSummary{
			LapNumber:   i + 1,
			LapName:     lap.Name,
			StartTime:   lap.StartIndex,
			EndTime:     lap.EndIndex,
			Duration:    lap.ElapsedTime,
			Distance:    lap.Distance,
			AvgSpeed:    lap.AverageSpeed,
			MaxSpeed:    lap.MaxSpeed,
			AvgHeartRate: lap.AverageHeartrate,
			MaxHeartRate: int(lap.MaxHeartrate),
			AvgPower:    lap.AveragePower,
			MaxPower:    int(lap.MaxPower),
		}
		lapAnalysis.LapSummaries = append(lapAnalysis.LapSummaries, summary)
	}

	// Basic lap comparisons (simplified for now)
	if len(laps) > 0 {
		fastestLap := 1
		slowestLap := 1
		fastestSpeed := laps[0].AverageSpeed
		slowestSpeed := laps[0].AverageSpeed

		for i, lap := range laps {
			if lap.AverageSpeed > fastestSpeed {
				fastestSpeed = lap.AverageSpeed
				fastestLap = i + 1
			}
			if lap.AverageSpeed < slowestSpeed {
				slowestSpeed = lap.AverageSpeed
				slowestLap = i + 1
			}
		}

		lapAnalysis.LapComparisons = LapComparisons{
			FastestLap:       fastestLap,
			SlowestLap:       slowestLap,
			SpeedVariation:   (fastestSpeed - slowestSpeed) / fastestSpeed,
			ConsistencyScore: 0.8, // Placeholder value
		}
	}

	return lapAnalysis, nil
}

// calculateFeatureSummary calculates high-level summary metrics
func (dfp *derivedFeaturesProcessor) calculateFeatureSummary(data *StravaStreams) FeatureSummary {
	summary := FeatureSummary{
		StreamTypes: dfp.getAvailableStreamTypes(data),
	}

	// Count total data points
	summary.TotalDataPoints = dfp.countDataPoints(data)

	// Calculate duration
	if len(data.Time) > 0 {
		summary.Duration = data.Time[len(data.Time)-1] - data.Time[0]
	}

	// Calculate distance
	if len(data.Distance) > 0 {
		summary.TotalDistance = data.Distance[len(data.Distance)-1] - data.Distance[0]
	}

	// Calculate elevation metrics
	if len(data.Altitude) > 0 {
		elevAnalysis := CalculateElevationAnalysis(data.Altitude, data.Distance, data.Time)
		summary.ElevationGain = elevAnalysis.TotalGain
		summary.ElevationLoss = elevAnalysis.TotalLoss
	}

	// Calculate speed metrics
	if len(data.VelocitySmooth) > 0 {
		speedStats := CalculateFloatStats(data.VelocitySmooth)
		summary.AvgSpeed = speedStats.Mean
		summary.MaxSpeed = speedStats.Max
	}

	// Calculate heart rate metrics
	if len(data.Heartrate) > 0 {
		hrStats := CalculateIntStats(data.Heartrate)
		summary.AvgHeartRate = hrStats.Mean
		summary.MaxHeartRate = int(hrStats.Max)

		// Calculate heart rate drift
		summary.HeartRateDrift = CalculateHeartRateDrift(data.Heartrate, data.Time)
	}

	// Calculate power metrics
	if len(data.Watts) > 0 {
		powerStats := CalculateIntStats(data.Watts)
		summary.AvgPower = powerStats.Mean
		summary.MaxPower = int(powerStats.Max)

		// Calculate normalized power
		summary.NormalizedPower = CalculateNormalizedPower(data.Watts, data.Time)
	}

	// Calculate cadence metrics
	if len(data.Cadence) > 0 {
		cadenceStats := CalculateIntStats(data.Cadence)
		summary.AvgCadence = cadenceStats.Mean
		summary.MaxCadence = int(cadenceStats.Max)
	}

	// Calculate temperature metrics
	if len(data.Temp) > 0 {
		tempStats := CalculateIntStats(data.Temp)
		summary.AvgTemperature = tempStats.Mean
	}

	// Calculate moving time percentage
	if len(data.Moving) > 0 {
		movingStats := CalculateBooleanStats(data.Moving)
		summary.MovingTimePercent = movingStats.TruePercent
	}

	return summary
}

// calculateStreamStatistics calculates comprehensive statistics for all stream types
func (dfp *derivedFeaturesProcessor) calculateStreamStatistics(data *StravaStreams) StreamStatistics {
	stats := StreamStatistics{}

	if len(data.Time) > 0 {
		timeFloat := make([]float64, len(data.Time))
		for i, t := range data.Time {
			timeFloat[i] = float64(t)
		}
		stats.Time = CalculateFloatStats(timeFloat)
	}

	if len(data.Distance) > 0 {
		stats.Distance = CalculateFloatStats(data.Distance)
	}

	if len(data.Altitude) > 0 {
		stats.Altitude = CalculateFloatStats(data.Altitude)
	}

	if len(data.VelocitySmooth) > 0 {
		stats.VelocitySmooth = CalculateFloatStats(data.VelocitySmooth)
	}

	if len(data.Heartrate) > 0 {
		stats.HeartRate = CalculateIntStats(data.Heartrate)
	}

	if len(data.Cadence) > 0 {
		stats.Cadence = CalculateIntStats(data.Cadence)
	}

	if len(data.Watts) > 0 {
		stats.Power = CalculateIntStats(data.Watts)
	}

	if len(data.Temp) > 0 {
		stats.Temperature = CalculateIntStats(data.Temp)
	}

	if len(data.GradeSmooth) > 0 {
		stats.Grade = CalculateFloatStats(data.GradeSmooth)
	}

	if len(data.Moving) > 0 {
		stats.Moving = CalculateBooleanStats(data.Moving)
	}

	if len(data.Latlng) > 0 {
		stats.LatLng = CalculateLocationStats(data.Latlng)
	}

	return stats
}

// extractInflectionPoints finds significant changes in data trends
func (dfp *derivedFeaturesProcessor) extractInflectionPoints(data *StravaStreams) []InflectionPoint {
	points := make([]InflectionPoint, 0)

	// Extract inflection points for heart rate
	if len(data.Heartrate) > 0 && len(data.Time) > 0 {
		hrFloat := make([]float64, len(data.Heartrate))
		for i, hr := range data.Heartrate {
			hrFloat[i] = float64(hr)
		}
		hrPoints := DetectInflectionPoints(hrFloat, data.Time, "heart_rate", 5.0)
		points = append(points, hrPoints...)
	}

	// Extract inflection points for power
	if len(data.Watts) > 0 && len(data.Time) > 0 {
		powerFloat := make([]float64, len(data.Watts))
		for i, w := range data.Watts {
			powerFloat[i] = float64(w)
		}
		powerPoints := DetectInflectionPoints(powerFloat, data.Time, "power", 10.0)
		points = append(points, powerPoints...)
	}

	// Extract inflection points for speed
	if len(data.VelocitySmooth) > 0 && len(data.Time) > 0 {
		speedPoints := DetectInflectionPoints(data.VelocitySmooth, data.Time, "speed", 1.0)
		points = append(points, speedPoints...)
	}

	// Extract inflection points for altitude
	if len(data.Altitude) > 0 && len(data.Time) > 0 {
		altPoints := DetectInflectionPoints(data.Altitude, data.Time, "altitude", 5.0)
		points = append(points, altPoints...)
	}

	return points
}

// extractTrends identifies directional patterns in the data
func (dfp *derivedFeaturesProcessor) extractTrends(data *StravaStreams) []Trend {
	trends := make([]Trend, 0)
	windowSize := 30 // 30-second window for trend analysis

	// Analyze heart rate trends
	if len(data.Heartrate) > 0 && len(data.Time) > 0 {
		hrFloat := make([]float64, len(data.Heartrate))
		for i, hr := range data.Heartrate {
			hrFloat[i] = float64(hr)
		}
		hrTrends := AnalyzeTrends(hrFloat, data.Time, "heart_rate", windowSize)
		trends = append(trends, hrTrends...)
	}

	// Analyze power trends
	if len(data.Watts) > 0 && len(data.Time) > 0 {
		powerFloat := make([]float64, len(data.Watts))
		for i, w := range data.Watts {
			powerFloat[i] = float64(w)
		}
		powerTrends := AnalyzeTrends(powerFloat, data.Time, "power", windowSize)
		trends = append(trends, powerTrends...)
	}

	// Analyze speed trends
	if len(data.VelocitySmooth) > 0 && len(data.Time) > 0 {
		speedTrends := AnalyzeTrends(data.VelocitySmooth, data.Time, "speed", windowSize)
		trends = append(trends, speedTrends...)
	}

	return trends
}

// extractSpikes finds significant deviations from normal values
func (dfp *derivedFeaturesProcessor) extractSpikes(data *StravaStreams) []Spike {
	spikes := make([]Spike, 0)
	threshold := 2.5 // 2.5 standard deviations

	// Detect heart rate spikes
	if len(data.Heartrate) > 0 && len(data.Time) > 0 {
		hrFloat := make([]float64, len(data.Heartrate))
		for i, hr := range data.Heartrate {
			hrFloat[i] = float64(hr)
		}
		hrSpikes := DetectSpikes(hrFloat, data.Time, "heart_rate", threshold)
		spikes = append(spikes, hrSpikes...)
	}

	// Detect power spikes
	if len(data.Watts) > 0 && len(data.Time) > 0 {
		powerFloat := make([]float64, len(data.Watts))
		for i, w := range data.Watts {
			powerFloat[i] = float64(w)
		}
		powerSpikes := DetectSpikes(powerFloat, data.Time, "power", threshold)
		spikes = append(spikes, powerSpikes...)
	}

	// Detect speed spikes
	if len(data.VelocitySmooth) > 0 && len(data.Time) > 0 {
		speedSpikes := DetectSpikes(data.VelocitySmooth, data.Time, "speed", threshold)
		spikes = append(spikes, speedSpikes...)
	}

	return spikes
}

// extractSampleData creates representative sample data points
func (dfp *derivedFeaturesProcessor) extractSampleData(data *StravaStreams) []DataPoint {
	if len(data.Time) == 0 {
		return []DataPoint{}
	}

	sampleData := make([]DataPoint, 0)
	
	// Sample at 0%, 25%, 50%, 75%, 100% of the activity
	sampleIndices := []int{
		0,
		len(data.Time) / 4,
		len(data.Time) / 2,
		len(data.Time) * 3 / 4,
		len(data.Time) - 1,
	}

	for _, idx := range sampleIndices {
		if idx >= len(data.Time) {
			continue
		}

		values := make(map[string]interface{})
		
		if idx < len(data.Heartrate) && data.Heartrate[idx] > 0 {
			values["heart_rate"] = data.Heartrate[idx]
		}
		if idx < len(data.Watts) && data.Watts[idx] > 0 {
			values["power"] = data.Watts[idx]
		}
		if idx < len(data.VelocitySmooth) && data.VelocitySmooth[idx] > 0 {
			values["speed"] = data.VelocitySmooth[idx]
		}
		if idx < len(data.Cadence) && data.Cadence[idx] > 0 {
			values["cadence"] = data.Cadence[idx]
		}
		if idx < len(data.Altitude) {
			values["altitude"] = data.Altitude[idx]
		}
		if idx < len(data.Distance) {
			values["distance"] = data.Distance[idx]
		}

		sampleData = append(sampleData, DataPoint{
			TimeOffset: data.Time[idx],
			Values:     values,
		})
	}

	return sampleData
}

// Helper methods

func (dfp *derivedFeaturesProcessor) countDataPoints(data *StravaStreams) int {
	if data == nil {
		return 0
	}
	
	maxPoints := 0
	
	if len(data.Time) > maxPoints {
		maxPoints = len(data.Time)
	}
	if len(data.Distance) > maxPoints {
		maxPoints = len(data.Distance)
	}
	if len(data.Heartrate) > maxPoints {
		maxPoints = len(data.Heartrate)
	}
	if len(data.Watts) > maxPoints {
		maxPoints = len(data.Watts)
	}
	if len(data.Cadence) > maxPoints {
		maxPoints = len(data.Cadence)
	}
	if len(data.Altitude) > maxPoints {
		maxPoints = len(data.Altitude)
	}
	if len(data.VelocitySmooth) > maxPoints {
		maxPoints = len(data.VelocitySmooth)
	}
	if len(data.Temp) > maxPoints {
		maxPoints = len(data.Temp)
	}
	if len(data.GradeSmooth) > maxPoints {
		maxPoints = len(data.GradeSmooth)
	}
	if len(data.Moving) > maxPoints {
		maxPoints = len(data.Moving)
	}
	if len(data.Latlng) > maxPoints {
		maxPoints = len(data.Latlng)
	}
	
	return maxPoints
}

func (dfp *derivedFeaturesProcessor) getAvailableStreamTypes(data *StravaStreams) []string {
	if data == nil {
		return []string{}
	}
	
	var types []string
	
	if len(data.Time) > 0 {
		types = append(types, "time")
	}
	if len(data.Distance) > 0 {
		types = append(types, "distance")
	}
	if len(data.Heartrate) > 0 {
		types = append(types, "heartrate")
	}
	if len(data.Watts) > 0 {
		types = append(types, "watts")
	}
	if len(data.Cadence) > 0 {
		types = append(types, "cadence")
	}
	if len(data.Altitude) > 0 {
		types = append(types, "altitude")
	}
	if len(data.VelocitySmooth) > 0 {
		types = append(types, "velocity_smooth")
	}
	if len(data.Temp) > 0 {
		types = append(types, "temp")
	}
	if len(data.GradeSmooth) > 0 {
		types = append(types, "grade_smooth")
	}
	if len(data.Moving) > 0 {
		types = append(types, "moving")
	}
	if len(data.Latlng) > 0 {
		types = append(types, "latlng")
	}
	
	return types
}