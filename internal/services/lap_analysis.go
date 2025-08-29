package services

import (
	"fmt"
	"math"
)

// LapAnalysis represents comprehensive lap-by-lap analysis
type LapAnalysis struct {
	TotalLaps        int                    `json:"total_laps"`
	LapSummaries     []LapSummary          `json:"lap_summaries"`
	LapComparisons   LapComparisons        `json:"lap_comparisons"`
	SegmentationType string                `json:"segmentation_type"` // "laps", "distance", "time"
}

// LapSummary represents statistics for a single lap
type LapSummary struct {
	LapNumber       int                    `json:"lap_number"`
	LapName         string                 `json:"lap_name,omitempty"`
	StartTime       int                    `json:"start_time"`
	EndTime         int                    `json:"end_time"`
	Duration        int                    `json:"duration"`
	Distance        float64                `json:"distance"`
	ElevationGain   float64                `json:"elevation_gain,omitempty"`
	ElevationLoss   float64                `json:"elevation_loss,omitempty"`
	AvgSpeed        float64                `json:"avg_speed,omitempty"`
	MaxSpeed        float64                `json:"max_speed,omitempty"`
	AvgHeartRate    float64                `json:"avg_heart_rate,omitempty"`
	MaxHeartRate    int                    `json:"max_heart_rate,omitempty"`
	AvgPower        float64                `json:"avg_power,omitempty"`
	MaxPower        int                    `json:"max_power,omitempty"`
	AvgCadence      float64                `json:"avg_cadence,omitempty"`
	MaxCadence      int                    `json:"max_cadence,omitempty"`
	AvgTemperature  float64                `json:"avg_temperature,omitempty"`
	Statistics      LapStatistics          `json:"statistics"`
	Trends          []LapTrend            `json:"trends"`
	Spikes          []LapSpike            `json:"spikes"`
}

// LapComparisons represents comparisons across all laps
type LapComparisons struct {
	FastestLap      int                    `json:"fastest_lap"`
	SlowestLap      int                    `json:"slowest_lap"`
	HighestPowerLap int                    `json:"highest_power_lap,omitempty"`
	LowestPowerLap  int                    `json:"lowest_power_lap,omitempty"`
	HighestHRLap    int                    `json:"highest_hr_lap,omitempty"`
	LowestHRLap     int                    `json:"lowest_hr_lap,omitempty"`
	SpeedVariation  float64                `json:"speed_variation"`
	PowerVariation  float64                `json:"power_variation,omitempty"`
	HRVariation     float64                `json:"hr_variation,omitempty"`
	ConsistencyScore float64               `json:"consistency_score"`
}

// LapStatistics represents statistical analysis for each metric within a lap
type LapStatistics struct {
	HeartRate      *MetricStats `json:"heart_rate,omitempty"`
	Power          *MetricStats `json:"power,omitempty"`
	Speed          *MetricStats `json:"speed,omitempty"`
	Cadence        *MetricStats `json:"cadence,omitempty"`
	Elevation      *MetricStats `json:"elevation,omitempty"`
	Temperature    *MetricStats `json:"temperature,omitempty"`
}

// LapTrend represents trends within a single lap
type LapTrend struct {
	Metric      string  `json:"metric"`
	Direction   string  `json:"direction"` // "increasing", "decreasing", "stable"
	Magnitude   float64 `json:"magnitude"`
	Confidence  float64 `json:"confidence"`
}

// LapSpike represents spikes within a single lap
type LapSpike struct {
	Metric      string  `json:"metric"`
	TimeOffset  int     `json:"time_offset"`
	Value       float64 `json:"value"`
	Magnitude   float64 `json:"magnitude"`
	Duration    int     `json:"duration"`
}

// DistanceSegment represents a distance-based segment when lap data is unavailable
type DistanceSegment struct {
	SegmentNumber int     `json:"segment_number"`
	StartDistance float64 `json:"start_distance"`
	EndDistance   float64 `json:"end_distance"`
	StartIndex    int     `json:"start_index"`
	EndIndex      int     `json:"end_index"`
	Distance      float64 `json:"distance"` // Should be close to segmentSize
}

// AnalyzeLapByLap performs comprehensive lap-by-lap analysis using activity lap data
func AnalyzeLapByLap(streams *StravaStreams, laps []StravaLap) *LapAnalysis {
	if len(laps) == 0 {
		// Fallback to distance-based segmentation
		return AnalyzeDistanceSegments(streams, 1000) // 1km segments
	}

	analysis := &LapAnalysis{
		TotalLaps:        len(laps),
		SegmentationType: "laps",
		LapSummaries:     make([]LapSummary, len(laps)),
	}

	// Analyze each lap
	for i, lap := range laps {
		lapSummary := analyzeSingleLap(streams, lap, i+1)
		analysis.LapSummaries[i] = lapSummary
	}

	// Calculate lap comparisons
	analysis.LapComparisons = calculateLapComparisons(analysis.LapSummaries)

	return analysis
}

// AnalyzeDistanceSegments creates distance-based segments when lap data is unavailable
func AnalyzeDistanceSegments(streams *StravaStreams, segmentSize float64) *LapAnalysis {
	if len(streams.Distance) == 0 || len(streams.Time) == 0 {
		return &LapAnalysis{SegmentationType: "distance"}
	}

	segments := createDistanceSegments(streams.Distance, segmentSize)
	if len(segments) == 0 {
		return &LapAnalysis{SegmentationType: "distance"}
	}

	analysis := &LapAnalysis{
		TotalLaps:        len(segments),
		SegmentationType: "distance",
		LapSummaries:     make([]LapSummary, len(segments)),
	}

	// Convert distance segments to lap format for analysis
	for i, segment := range segments {
		// Create a synthetic lap from the distance segment
		syntheticLap := StravaLap{
			LapIndex:   i,
			StartIndex: segment.StartIndex,
			EndIndex:   segment.EndIndex,
			Distance:   segment.Distance,
		}

		// Calculate timing if time data is available
		if len(streams.Time) > segment.EndIndex {
			syntheticLap.ElapsedTime = streams.Time[segment.EndIndex] - streams.Time[segment.StartIndex]
			syntheticLap.MovingTime = syntheticLap.ElapsedTime // Simplified
		}

		lapSummary := analyzeSingleLap(streams, syntheticLap, i+1)
		lapSummary.LapName = fmt.Sprintf("Segment %d (%.1fkm)", i+1, segment.Distance/1000)
		analysis.LapSummaries[i] = lapSummary
	}

	// Calculate comparisons
	analysis.LapComparisons = calculateLapComparisons(analysis.LapSummaries)

	return analysis
}

// analyzeSingleLap performs detailed analysis of a single lap
func analyzeSingleLap(streams *StravaStreams, lap StravaLap, lapNumber int) LapSummary {
	summary := LapSummary{
		LapNumber: lapNumber,
		LapName:   lap.Name,
		Distance:  lap.Distance,
		Duration:  lap.ElapsedTime,
	}

	// Validate indices
	if lap.StartIndex < 0 || lap.EndIndex >= len(streams.Time) || lap.StartIndex >= lap.EndIndex {
		return summary
	}

	// Set timing information
	if len(streams.Time) > lap.EndIndex {
		summary.StartTime = streams.Time[lap.StartIndex]
		summary.EndTime = streams.Time[lap.EndIndex]
	}

	// Analyze each metric within the lap boundaries
	summary.Statistics = LapStatistics{}

	// Heart Rate Analysis
	if len(streams.Heartrate) > lap.EndIndex {
		hrData := streams.Heartrate[lap.StartIndex:lap.EndIndex+1]
		if len(hrData) > 0 {
			hrFloat := make([]float64, len(hrData))
			sum := 0
			max := 0
			for i, hr := range hrData {
				hrFloat[i] = float64(hr)
				sum += hr
				if hr > max {
					max = hr
				}
			}
			summary.AvgHeartRate = float64(sum) / float64(len(hrData))
			summary.MaxHeartRate = max
			summary.Statistics.HeartRate = CalculateFloatStats(hrFloat)
		}
	}

	// Power Analysis
	if len(streams.Watts) > lap.EndIndex {
		powerData := streams.Watts[lap.StartIndex:lap.EndIndex+1]
		if len(powerData) > 0 {
			powerFloat := make([]float64, len(powerData))
			sum := 0
			max := 0
			for i, power := range powerData {
				powerFloat[i] = float64(power)
				sum += power
				if power > max {
					max = power
				}
			}
			summary.AvgPower = float64(sum) / float64(len(powerData))
			summary.MaxPower = max
			summary.Statistics.Power = CalculateFloatStats(powerFloat)
		}
	}

	// Speed Analysis
	if len(streams.VelocitySmooth) > lap.EndIndex {
		speedData := streams.VelocitySmooth[lap.StartIndex:lap.EndIndex+1]
		if len(speedData) > 0 {
			sum := 0.0
			max := 0.0
			for _, speed := range speedData {
				sum += speed
				if speed > max {
					max = speed
				}
			}
			summary.AvgSpeed = sum / float64(len(speedData))
			summary.MaxSpeed = max
			summary.Statistics.Speed = CalculateFloatStats(speedData)
		}
	}

	// Cadence Analysis
	if len(streams.Cadence) > lap.EndIndex {
		cadenceData := streams.Cadence[lap.StartIndex:lap.EndIndex+1]
		if len(cadenceData) > 0 {
			cadenceFloat := make([]float64, len(cadenceData))
			sum := 0
			max := 0
			for i, cadence := range cadenceData {
				cadenceFloat[i] = float64(cadence)
				sum += cadence
				if cadence > max {
					max = cadence
				}
			}
			summary.AvgCadence = float64(sum) / float64(len(cadenceData))
			summary.MaxCadence = max
			summary.Statistics.Cadence = CalculateFloatStats(cadenceFloat)
		}
	}

	// Elevation Analysis
	if len(streams.Altitude) > lap.EndIndex {
		elevationData := streams.Altitude[lap.StartIndex:lap.EndIndex+1]
		if len(elevationData) > 0 {
			summary.Statistics.Elevation = CalculateFloatStats(elevationData)
			
			// Calculate elevation gain/loss for this lap
			gain := 0.0
			loss := 0.0
			for i := 1; i < len(elevationData); i++ {
				diff := elevationData[i] - elevationData[i-1]
				if diff > 0 {
					gain += diff
				} else {
					loss += math.Abs(diff)
				}
			}
			summary.ElevationGain = gain
			summary.ElevationLoss = loss
		}
	}

	// Temperature Analysis
	if len(streams.Temp) > lap.EndIndex {
		tempData := streams.Temp[lap.StartIndex:lap.EndIndex+1]
		if len(tempData) > 0 {
			tempFloat := make([]float64, len(tempData))
			sum := 0
			for i, temp := range tempData {
				tempFloat[i] = float64(temp)
				sum += temp
			}
			summary.AvgTemperature = float64(sum) / float64(len(tempData))
			summary.Statistics.Temperature = CalculateFloatStats(tempFloat)
		}
	}

	// Analyze trends within the lap
	summary.Trends = analyzeLapTrends(streams, lap)

	// Detect spikes within the lap
	summary.Spikes = detectLapSpikes(streams, lap)

	return summary
}

// calculateLapComparisons compares performance across all laps
func calculateLapComparisons(lapSummaries []LapSummary) LapComparisons {
	if len(lapSummaries) == 0 {
		return LapComparisons{}
	}

	comparisons := LapComparisons{}
	
	// Find fastest and slowest laps (by average speed)
	fastestSpeed := 0.0
	slowestSpeed := math.MaxFloat64
	
	var avgSpeeds []float64
	var avgPowers []float64
	var avgHeartRates []float64

	for i, lap := range lapSummaries {
		// Speed comparisons
		if lap.AvgSpeed > fastestSpeed {
			fastestSpeed = lap.AvgSpeed
			comparisons.FastestLap = i + 1
		}
		if lap.AvgSpeed < slowestSpeed && lap.AvgSpeed > 0 {
			slowestSpeed = lap.AvgSpeed
			comparisons.SlowestLap = i + 1
		}
		avgSpeeds = append(avgSpeeds, lap.AvgSpeed)

		// Power comparisons
		if lap.AvgPower > 0 {
			if comparisons.HighestPowerLap == 0 || lap.AvgPower > lapSummaries[comparisons.HighestPowerLap-1].AvgPower {
				comparisons.HighestPowerLap = i + 1
			}
			if comparisons.LowestPowerLap == 0 || lap.AvgPower < lapSummaries[comparisons.LowestPowerLap-1].AvgPower {
				comparisons.LowestPowerLap = i + 1
			}
			avgPowers = append(avgPowers, lap.AvgPower)
		}

		// Heart rate comparisons
		if lap.AvgHeartRate > 0 {
			if comparisons.HighestHRLap == 0 || lap.AvgHeartRate > lapSummaries[comparisons.HighestHRLap-1].AvgHeartRate {
				comparisons.HighestHRLap = i + 1
			}
			if comparisons.LowestHRLap == 0 || lap.AvgHeartRate < lapSummaries[comparisons.LowestHRLap-1].AvgHeartRate {
				comparisons.LowestHRLap = i + 1
			}
			avgHeartRates = append(avgHeartRates, lap.AvgHeartRate)
		}
	}

	// Calculate variation coefficients
	if len(avgSpeeds) > 0 {
		speedStats := CalculateFloatStats(avgSpeeds)
		comparisons.SpeedVariation = speedStats.Variability
	}

	if len(avgPowers) > 0 {
		powerStats := CalculateFloatStats(avgPowers)
		comparisons.PowerVariation = powerStats.Variability
	}

	if len(avgHeartRates) > 0 {
		hrStats := CalculateFloatStats(avgHeartRates)
		comparisons.HRVariation = hrStats.Variability
	}

	// Calculate overall consistency score (lower variation = higher consistency)
	totalVariation := comparisons.SpeedVariation
	if comparisons.PowerVariation > 0 {
		totalVariation += comparisons.PowerVariation
	}
	if comparisons.HRVariation > 0 {
		totalVariation += comparisons.HRVariation
	}

	// Convert to consistency score (0-1, where 1 is perfectly consistent)
	if totalVariation > 0 {
		comparisons.ConsistencyScore = math.Max(0, 1.0 - totalVariation)
	} else {
		comparisons.ConsistencyScore = 1.0
	}

	return comparisons
}

// createDistanceSegments divides the activity into equal distance segments
func createDistanceSegments(distanceData []float64, segmentSize float64) []DistanceSegment {
	if len(distanceData) == 0 {
		return []DistanceSegment{}
	}

	var segments []DistanceSegment
	totalDistance := distanceData[len(distanceData)-1]
	
	if totalDistance < segmentSize {
		// If total distance is less than segment size, create one segment
		return []DistanceSegment{{
			SegmentNumber: 1,
			StartDistance: 0,
			EndDistance:   totalDistance,
			StartIndex:    0,
			EndIndex:      len(distanceData) - 1,
			Distance:      totalDistance,
		}}
	}

	segmentNumber := 1
	currentDistance := 0.0

	for currentDistance < totalDistance {
		targetDistance := currentDistance + segmentSize
		if targetDistance > totalDistance {
			targetDistance = totalDistance
		}

		// Find indices for this segment
		startIndex := findDistanceIndex(distanceData, currentDistance)
		endIndex := findDistanceIndex(distanceData, targetDistance)

		if startIndex < endIndex {
			segments = append(segments, DistanceSegment{
				SegmentNumber: segmentNumber,
				StartDistance: currentDistance,
				EndDistance:   targetDistance,
				StartIndex:    startIndex,
				EndIndex:      endIndex,
				Distance:      targetDistance - currentDistance,
			})
		}

		currentDistance = targetDistance
		segmentNumber++
	}

	return segments
}

// findDistanceIndex finds the index in distance data closest to the target distance
func findDistanceIndex(distanceData []float64, targetDistance float64) int {
	if len(distanceData) == 0 {
		return 0
	}

	// Binary search for efficiency
	left, right := 0, len(distanceData)-1
	
	for left < right {
		mid := (left + right) / 2
		if distanceData[mid] < targetDistance {
			left = mid + 1
		} else {
			right = mid
		}
	}

	return left
}

// analyzeLapTrends identifies trends within a single lap
func analyzeLapTrends(streams *StravaStreams, lap StravaLap) []LapTrend {
	var trends []LapTrend

	// Analyze heart rate trend
	if len(streams.Heartrate) > lap.EndIndex {
		hrData := streams.Heartrate[lap.StartIndex:lap.EndIndex+1]
		if len(hrData) > 5 {
			hrFloat := make([]float64, len(hrData))
			timeData := make([]int, len(hrData))
			for i, hr := range hrData {
				hrFloat[i] = float64(hr)
				if len(streams.Time) > lap.StartIndex+i {
					timeData[i] = streams.Time[lap.StartIndex+i]
				}
			}
			
			slope := calculateSlope(hrFloat, timeData)
			direction := "stable"
			if slope > 0.01 {
				direction = "increasing"
			} else if slope < -0.01 {
				direction = "decreasing"
			}

			trends = append(trends, LapTrend{
				Metric:     "heart_rate",
				Direction:  direction,
				Magnitude:  math.Abs(slope),
				Confidence: calculateTrendConfidence(hrFloat),
			})
		}
	}

	// Analyze power trend
	if len(streams.Watts) > lap.EndIndex {
		powerData := streams.Watts[lap.StartIndex:lap.EndIndex+1]
		if len(powerData) > 5 {
			powerFloat := make([]float64, len(powerData))
			timeData := make([]int, len(powerData))
			for i, power := range powerData {
				powerFloat[i] = float64(power)
				if len(streams.Time) > lap.StartIndex+i {
					timeData[i] = streams.Time[lap.StartIndex+i]
				}
			}
			
			slope := calculateSlope(powerFloat, timeData)
			direction := "stable"
			if slope > 0.1 {
				direction = "increasing"
			} else if slope < -0.1 {
				direction = "decreasing"
			}

			trends = append(trends, LapTrend{
				Metric:     "power",
				Direction:  direction,
				Magnitude:  math.Abs(slope),
				Confidence: calculateTrendConfidence(powerFloat),
			})
		}
	}

	// Analyze speed trend
	if len(streams.VelocitySmooth) > lap.EndIndex {
		speedData := streams.VelocitySmooth[lap.StartIndex:lap.EndIndex+1]
		if len(speedData) > 5 {
			timeData := make([]int, len(speedData))
			for i := range speedData {
				if len(streams.Time) > lap.StartIndex+i {
					timeData[i] = streams.Time[lap.StartIndex+i]
				}
			}
			
			slope := calculateSlope(speedData, timeData)
			direction := "stable"
			if slope > 0.001 {
				direction = "increasing"
			} else if slope < -0.001 {
				direction = "decreasing"
			}

			trends = append(trends, LapTrend{
				Metric:     "speed",
				Direction:  direction,
				Magnitude:  math.Abs(slope),
				Confidence: calculateTrendConfidence(speedData),
			})
		}
	}

	return trends
}

// detectLapSpikes finds spikes within a single lap
func detectLapSpikes(streams *StravaStreams, lap StravaLap) []LapSpike {
	var spikes []LapSpike

	// Detect power spikes
	if len(streams.Watts) > lap.EndIndex {
		powerData := streams.Watts[lap.StartIndex:lap.EndIndex+1]
		if len(powerData) > 3 {
			powerFloat := make([]float64, len(powerData))
			for i, power := range powerData {
				powerFloat[i] = float64(power)
			}
			
			stats := CalculateFloatStats(powerFloat)
			threshold := 2.0 // 2 standard deviations
			
			for i, power := range powerFloat {
				if math.Abs(power-stats.Mean) > threshold*stats.StdDev {
					timeOffset := 0
					if len(streams.Time) > lap.StartIndex+i && len(streams.Time) > lap.StartIndex {
						timeOffset = streams.Time[lap.StartIndex+i] - streams.Time[lap.StartIndex]
					}
					
					spikes = append(spikes, LapSpike{
						Metric:     "power",
						TimeOffset: timeOffset,
						Value:      power,
						Magnitude:  math.Abs(power-stats.Mean) / stats.StdDev,
						Duration:   1, // Simplified duration
					})
				}
			}
		}
	}

	// Detect heart rate spikes
	if len(streams.Heartrate) > lap.EndIndex {
		hrData := streams.Heartrate[lap.StartIndex:lap.EndIndex+1]
		if len(hrData) > 3 {
			hrFloat := make([]float64, len(hrData))
			for i, hr := range hrData {
				hrFloat[i] = float64(hr)
			}
			
			stats := CalculateFloatStats(hrFloat)
			threshold := 2.0 // 2 standard deviations
			
			for i, hr := range hrFloat {
				if math.Abs(hr-stats.Mean) > threshold*stats.StdDev {
					timeOffset := 0
					if len(streams.Time) > lap.StartIndex+i && len(streams.Time) > lap.StartIndex {
						timeOffset = streams.Time[lap.StartIndex+i] - streams.Time[lap.StartIndex]
					}
					
					spikes = append(spikes, LapSpike{
						Metric:     "heart_rate",
						TimeOffset: timeOffset,
						Value:      hr,
						Magnitude:  math.Abs(hr-stats.Mean) / stats.StdDev,
						Duration:   1, // Simplified duration
					})
				}
			}
		}
	}

	return spikes
}