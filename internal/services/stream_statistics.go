package services

import (
	"math"
	"sort"
)

// MetricStats represents statistical analysis for numeric streams
type MetricStats struct {
	Min         float64 `json:"min"`
	Max         float64 `json:"max"`
	Mean        float64 `json:"mean"`
	Median      float64 `json:"median"`
	StdDev      float64 `json:"std_dev"`
	Variability float64 `json:"variability"`
	Range       float64 `json:"range"`
	Q25         float64 `json:"q25"`
	Q75         float64 `json:"q75"`
	Count       int     `json:"count"`
}

// BooleanStats represents statistical analysis for boolean streams (moving time data)
type BooleanStats struct {
	TrueCount    int     `json:"true_count"`
	FalseCount   int     `json:"false_count"`
	TruePercent  float64 `json:"true_percent"`
	FalsePercent float64 `json:"false_percent"`
	TotalCount   int     `json:"total_count"`
}

// LocationStats represents statistical analysis for GPS coordinates
type LocationStats struct {
	StartLat    float64     `json:"start_lat"`
	StartLng    float64     `json:"start_lng"`
	EndLat      float64     `json:"end_lat"`
	EndLng      float64     `json:"end_lng"`
	BoundingBox BoundingBox `json:"bounding_box"`
	TotalPoints int         `json:"total_points"`
}

// BoundingBox represents the geographic bounds of GPS coordinates
type BoundingBox struct {
	NorthLat float64 `json:"north_lat"`
	SouthLat float64 `json:"south_lat"`
	EastLng  float64 `json:"east_lng"`
	WestLng  float64 `json:"west_lng"`
}

// CalculateIntStats calculates statistics for integer slices
func CalculateIntStats(data []int) *MetricStats {
	if len(data) == 0 {
		return &MetricStats{Count: 0}
	}

	// Convert to float64 for calculations
	floatData := make([]float64, len(data))
	for i, v := range data {
		floatData[i] = float64(v)
	}

	return CalculateFloatStats(floatData)
}

// CalculateFloatStats calculates comprehensive statistics for float64 slices
func CalculateFloatStats(data []float64) *MetricStats {
	if len(data) == 0 {
		return &MetricStats{Count: 0}
	}

	// Filter out zero values for more meaningful statistics
	nonZeroData := make([]float64, 0, len(data))
	for _, v := range data {
		if v != 0 {
			nonZeroData = append(nonZeroData, v)
		}
	}

	// If all values are zero, use original data
	if len(nonZeroData) == 0 {
		nonZeroData = data
	}

	// Sort data for percentile calculations
	sortedData := make([]float64, len(nonZeroData))
	copy(sortedData, nonZeroData)
	sort.Float64s(sortedData)

	stats := &MetricStats{
		Count: len(nonZeroData),
		Min:   sortedData[0],
		Max:   sortedData[len(sortedData)-1],
	}

	stats.Range = stats.Max - stats.Min

	// Calculate mean
	sum := 0.0
	for _, v := range nonZeroData {
		sum += v
	}
	stats.Mean = sum / float64(len(nonZeroData))

	// Calculate median
	n := len(sortedData)
	if n%2 == 0 {
		stats.Median = (sortedData[n/2-1] + sortedData[n/2]) / 2
	} else {
		stats.Median = sortedData[n/2]
	}

	// Calculate quartiles
	stats.Q25 = calculatePercentile(sortedData, 0.25)
	stats.Q75 = calculatePercentile(sortedData, 0.75)

	// Calculate standard deviation
	sumSquaredDiff := 0.0
	for _, v := range nonZeroData {
		diff := v - stats.Mean
		sumSquaredDiff += diff * diff
	}
	stats.StdDev = math.Sqrt(sumSquaredDiff / float64(len(nonZeroData)))

	// Calculate coefficient of variation (variability)
	if stats.Mean != 0 {
		stats.Variability = stats.StdDev / stats.Mean
	}

	return stats
}

// CalculateBooleanStats calculates statistics for boolean slices (moving time data)
func CalculateBooleanStats(data []bool) *BooleanStats {
	if len(data) == 0 {
		return &BooleanStats{TotalCount: 0}
	}

	trueCount := 0
	for _, v := range data {
		if v {
			trueCount++
		}
	}

	falseCount := len(data) - trueCount
	totalCount := len(data)

	stats := &BooleanStats{
		TrueCount:    trueCount,
		FalseCount:   falseCount,
		TotalCount:   totalCount,
		TruePercent:  float64(trueCount) / float64(totalCount) * 100,
		FalsePercent: float64(falseCount) / float64(totalCount) * 100,
	}

	return stats
}

// CalculateLocationStats calculates statistics for GPS coordinate data
func CalculateLocationStats(data [][]float64) *LocationStats {
	if len(data) == 0 {
		return &LocationStats{TotalPoints: 0}
	}

	// Filter out invalid coordinates (0,0)
	validCoords := make([][]float64, 0, len(data))
	for _, coord := range data {
		if len(coord) == 2 && (coord[0] != 0 || coord[1] != 0) {
			validCoords = append(validCoords, coord)
		}
	}

	if len(validCoords) == 0 {
		return &LocationStats{TotalPoints: 0}
	}

	stats := &LocationStats{
		TotalPoints: len(validCoords),
		StartLat:    validCoords[0][0],
		StartLng:    validCoords[0][1],
		EndLat:      validCoords[len(validCoords)-1][0],
		EndLng:      validCoords[len(validCoords)-1][1],
	}

	// Calculate bounding box
	minLat, maxLat := validCoords[0][0], validCoords[0][0]
	minLng, maxLng := validCoords[0][1], validCoords[0][1]

	for _, coord := range validCoords {
		lat, lng := coord[0], coord[1]
		if lat < minLat {
			minLat = lat
		}
		if lat > maxLat {
			maxLat = lat
		}
		if lng < minLng {
			minLng = lng
		}
		if lng > maxLng {
			maxLng = lng
		}
	}

	stats.BoundingBox = BoundingBox{
		NorthLat: maxLat,
		SouthLat: minLat,
		EastLng:  maxLng,
		WestLng:  minLng,
	}

	return stats
}

// calculatePercentile calculates the value at a given percentile in sorted data
func calculatePercentile(sortedData []float64, percentile float64) float64 {
	if len(sortedData) == 0 {
		return 0
	}

	index := percentile * float64(len(sortedData)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedData[lower]
	}

	weight := index - float64(lower)
	return sortedData[lower]*(1-weight) + sortedData[upper]*weight
}

// CalculateQuartiles calculates Q1, Q2 (median), and Q3 for a dataset
func CalculateQuartiles(data []float64) (q1, q2, q3 float64) {
	if len(data) == 0 {
		return 0, 0, 0
	}

	sortedData := make([]float64, len(data))
	copy(sortedData, data)
	sort.Float64s(sortedData)

	q1 = calculatePercentile(sortedData, 0.25)
	q2 = calculatePercentile(sortedData, 0.50)
	q3 = calculatePercentile(sortedData, 0.75)

	return q1, q2, q3
}

// CalculateVariabilityMetrics calculates various variability measures
func CalculateVariabilityMetrics(data []float64) (cv, iqr, mad float64) {
	if len(data) == 0 {
		return 0, 0, 0
	}

	stats := CalculateFloatStats(data)
	
	// Coefficient of Variation
	cv = stats.Variability

	// Interquartile Range
	iqr = stats.Q75 - stats.Q25

	// Median Absolute Deviation
	sortedData := make([]float64, len(data))
	copy(sortedData, data)
	sort.Float64s(sortedData)
	
	median := stats.Median
	deviations := make([]float64, len(data))
	for i, v := range data {
		deviations[i] = math.Abs(v - median)
	}
	sort.Float64s(deviations)
	
	mad = calculatePercentile(deviations, 0.50)

	return cv, iqr, mad
}