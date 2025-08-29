package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"runtime"
	"sort"
	"sync"
)

// OptimizedStatisticsCalculator provides optimized statistical calculations for large datasets
type OptimizedStatisticsCalculator struct {
	maxWorkers int
	chunkSize  int
}

// NewOptimizedStatisticsCalculator creates a new optimized statistics calculator
func NewOptimizedStatisticsCalculator() *OptimizedStatisticsCalculator {
	return &OptimizedStatisticsCalculator{
		maxWorkers: runtime.NumCPU(),
		chunkSize:  1000, // Process data in chunks of 1000 points
	}
}

// CalculateMetricStatsOptimized calculates statistics for numeric data using parallel processing
func (osc *OptimizedStatisticsCalculator) CalculateMetricStatsOptimized(ctx context.Context, data []float64) *MetricStats {
	if len(data) == 0 {
		return &MetricStats{}
	}

	// For small datasets, use simple calculation
	if len(data) < osc.chunkSize {
		return osc.calculateMetricStatsSimple(data)
	}

	// Use parallel processing for large datasets
	return osc.calculateMetricStatsParallel(ctx, data)
}

// calculateMetricStatsSimple calculates statistics using simple sequential processing
func (osc *OptimizedStatisticsCalculator) calculateMetricStatsSimple(data []float64) *MetricStats {
	if len(data) == 0 {
		return &MetricStats{}
	}

	// Calculate min, max, and sum in single pass
	min, max, sum := data[0], data[0], 0.0
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}

	mean := sum / float64(len(data))

	// Calculate variance in second pass
	variance := 0.0
	for _, v := range data {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(data))
	stdDev := math.Sqrt(variance)

	// Calculate median and quartiles
	sortedData := make([]float64, len(data))
	copy(sortedData, data)
	sort.Float64s(sortedData)

	median := osc.calculatePercentile(sortedData, 0.5)
	q25 := osc.calculatePercentile(sortedData, 0.25)
	q75 := osc.calculatePercentile(sortedData, 0.75)

	variability := 0.0
	if mean != 0 {
		variability = stdDev / math.Abs(mean) // Coefficient of variation
	}

	return &MetricStats{
		Min:         min,
		Max:         max,
		Mean:        mean,
		Median:      median,
		StdDev:      stdDev,
		Variability: variability,
		Range:       max - min,
		Q25:         q25,
		Q75:         q75,
		Count:       len(data),
	}
}

// calculateMetricStatsParallel calculates statistics using parallel processing
func (osc *OptimizedStatisticsCalculator) calculateMetricStatsParallel(ctx context.Context, data []float64) *MetricStats {
	numWorkers := osc.maxWorkers
	if numWorkers > len(data)/osc.chunkSize {
		numWorkers = len(data)/osc.chunkSize + 1
	}

	// Channel for chunk results
	type chunkResult struct {
		min, max, sum float64
		count         int
		sumSquares    float64 // For variance calculation
	}

	resultChan := make(chan chunkResult, numWorkers)
	chunkSize := len(data) / numWorkers

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			start := workerID * chunkSize
			end := start + chunkSize
			if workerID == numWorkers-1 {
				end = len(data) // Last worker handles remaining data
			}

			if start >= len(data) {
				return
			}

			chunk := data[start:end]
			if len(chunk) == 0 {
				return
			}

			// Calculate chunk statistics
			min, max, sum := chunk[0], chunk[0], 0.0
			sumSquares := 0.0

			for _, v := range chunk {
				if v < min {
					min = v
				}
				if v > max {
					max = v
				}
				sum += v
				sumSquares += v * v
			}

			select {
			case resultChan <- chunkResult{
				min:        min,
				max:        max,
				sum:        sum,
				count:      len(chunk),
				sumSquares: sumSquares,
			}:
			case <-ctx.Done():
				return
			}
		}(i)
	}

	// Close channel when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Aggregate results
	var globalMin, globalMax, globalSum, globalSumSquares float64
	var totalCount int
	first := true

	for result := range resultChan {
		if first {
			globalMin = result.min
			globalMax = result.max
			first = false
		} else {
			if result.min < globalMin {
				globalMin = result.min
			}
			if result.max > globalMax {
				globalMax = result.max
			}
		}
		globalSum += result.sum
		globalSumSquares += result.sumSquares
		totalCount += result.count
	}

	if totalCount == 0 {
		return &MetricStats{}
	}

	// Calculate derived statistics
	mean := globalSum / float64(totalCount)
	variance := (globalSumSquares / float64(totalCount)) - (mean * mean)
	stdDev := math.Sqrt(math.Max(0, variance)) // Ensure non-negative

	// For median and quartiles, we need to sort (can't parallelize this efficiently)
	sortedData := make([]float64, len(data))
	copy(sortedData, data)
	sort.Float64s(sortedData)

	median := osc.calculatePercentile(sortedData, 0.5)
	q25 := osc.calculatePercentile(sortedData, 0.25)
	q75 := osc.calculatePercentile(sortedData, 0.75)

	variability := 0.0
	if mean != 0 {
		variability = stdDev / math.Abs(mean)
	}

	return &MetricStats{
		Min:         globalMin,
		Max:         globalMax,
		Mean:        mean,
		Median:      median,
		StdDev:      stdDev,
		Variability: variability,
		Range:       globalMax - globalMin,
		Q25:         q25,
		Q75:         q75,
		Count:       totalCount,
	}
}

// calculatePercentile calculates the specified percentile from sorted data
func (osc *OptimizedStatisticsCalculator) calculatePercentile(sortedData []float64, percentile float64) float64 {
	if len(sortedData) == 0 {
		return 0
	}

	if len(sortedData) == 1 {
		return sortedData[0]
	}

	index := percentile * float64(len(sortedData)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedData[lower]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sortedData[lower]*(1-weight) + sortedData[upper]*weight
}

// OptimizedInflectionPointDetector provides optimized inflection point detection
type OptimizedInflectionPointDetector struct {
	windowSize    int
	threshold     float64
	maxWorkers    int
	minDataPoints int
}

// NewOptimizedInflectionPointDetector creates a new optimized inflection point detector
func NewOptimizedInflectionPointDetector() *OptimizedInflectionPointDetector {
	return &OptimizedInflectionPointDetector{
		windowSize:    10,  // Look at 10 points around each candidate
		threshold:     0.1, // Minimum change threshold
		maxWorkers:    runtime.NumCPU(),
		minDataPoints: 50, // Minimum data points to use parallel processing
	}
}

// DetectInflectionPointsOptimized detects inflection points using optimized algorithms
func (oipd *OptimizedInflectionPointDetector) DetectInflectionPointsOptimized(ctx context.Context, data []float64, streamType string) []InflectionPoint {
	if len(data) < oipd.windowSize*2 {
		return []InflectionPoint{}
	}

	// For small datasets, use simple detection
	if len(data) < oipd.minDataPoints {
		return oipd.detectInflectionPointsSimple(data, streamType)
	}

	// Use parallel processing for large datasets
	return oipd.detectInflectionPointsParallel(ctx, data, streamType)
}

// detectInflectionPointsSimple detects inflection points using simple sequential processing
func (oipd *OptimizedInflectionPointDetector) detectInflectionPointsSimple(data []float64, streamType string) []InflectionPoint {
	var inflectionPoints []InflectionPoint

	for i := oipd.windowSize; i < len(data)-oipd.windowSize; i++ {
		// Calculate slopes before and after the point
		slopeBefore := oipd.calculateSlope(data, i-oipd.windowSize, i)
		slopeAfter := oipd.calculateSlope(data, i, i+oipd.windowSize)

		// Check for significant change in slope
		slopeChange := math.Abs(slopeAfter - slopeBefore)
		if slopeChange > oipd.threshold {
			inflectionType := "increase"
			if slopeAfter < slopeBefore {
				inflectionType = "decrease"
			}

			inflectionPoints = append(inflectionPoints, InflectionPoint{
				Index:     i,
				Value:     data[i],
				Direction: inflectionType,
				Magnitude: slopeChange,
				Metric:    streamType,
			})
		}
	}

	return inflectionPoints
}

// detectInflectionPointsParallel detects inflection points using parallel processing
func (oipd *OptimizedInflectionPointDetector) detectInflectionPointsParallel(ctx context.Context, data []float64, streamType string) []InflectionPoint {
	numWorkers := oipd.maxWorkers
	chunkSize := (len(data) - 2*oipd.windowSize) / numWorkers

	if chunkSize < oipd.windowSize {
		// Not enough data for parallel processing
		return oipd.detectInflectionPointsSimple(data, streamType)
	}

	resultChan := make(chan []InflectionPoint, numWorkers)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			start := oipd.windowSize + workerID*chunkSize
			end := start + chunkSize
			if workerID == numWorkers-1 {
				end = len(data) - oipd.windowSize // Last worker handles remaining data
			}

			var workerInflectionPoints []InflectionPoint

			for j := start; j < end; j++ {
				// Calculate slopes before and after the point
				slopeBefore := oipd.calculateSlope(data, j-oipd.windowSize, j)
				slopeAfter := oipd.calculateSlope(data, j, j+oipd.windowSize)

				// Check for significant change in slope
				slopeChange := math.Abs(slopeAfter - slopeBefore)
				if slopeChange > oipd.threshold {
					inflectionType := "increase"
					if slopeAfter < slopeBefore {
						inflectionType = "decrease"
					}

					workerInflectionPoints = append(workerInflectionPoints, InflectionPoint{
						Index:     j,
						Value:     data[j],
						Direction: inflectionType,
						Magnitude: slopeChange,
						Metric:    streamType,
					})
				}
			}

			select {
			case resultChan <- workerInflectionPoints:
			case <-ctx.Done():
				return
			}
		}(i)
	}

	// Close channel when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Aggregate results
	var allInflectionPoints []InflectionPoint
	for workerResults := range resultChan {
		allInflectionPoints = append(allInflectionPoints, workerResults...)
	}

	// Sort by index
	sort.Slice(allInflectionPoints, func(i, j int) bool {
		return allInflectionPoints[i].Index < allInflectionPoints[j].Index
	})

	return allInflectionPoints
}

// calculateSlope calculates the slope between two points in the data
func (oipd *OptimizedInflectionPointDetector) calculateSlope(data []float64, start, end int) float64 {
	if start >= end || end >= len(data) {
		return 0
	}

	return (data[end] - data[start]) / float64(end-start)
}

// MemoryOptimizedProcessor provides memory-efficient processing for large datasets
type MemoryOptimizedProcessor struct {
	maxMemoryMB int
	chunkSize   int
}

// NewMemoryOptimizedProcessor creates a new memory-optimized processor
func NewMemoryOptimizedProcessor(maxMemoryMB int) *MemoryOptimizedProcessor {
	return &MemoryOptimizedProcessor{
		maxMemoryMB: maxMemoryMB,
		chunkSize:   10000, // Process 10k points at a time by default
	}
}

// ProcessStreamDataInChunks processes stream data in memory-efficient chunks
func (mop *MemoryOptimizedProcessor) ProcessStreamDataInChunks(ctx context.Context, data *StravaStreams, processor func(chunk *StravaStreams) error) error {
	if data == nil {
		return fmt.Errorf("stream data is nil")
	}

	// Determine the size of the largest stream
	maxSize := 0
	if len(data.Time) > maxSize {
		maxSize = len(data.Time)
	}
	if len(data.Heartrate) > maxSize {
		maxSize = len(data.Heartrate)
	}
	if len(data.Watts) > maxSize {
		maxSize = len(data.Watts)
	}
	// ... check other streams

	if maxSize <= mop.chunkSize {
		// Data is small enough to process in one chunk
		return processor(data)
	}

	// Process data in chunks
	for start := 0; start < maxSize; start += mop.chunkSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		end := start + mop.chunkSize
		if end > maxSize {
			end = maxSize
		}

		// Create chunk
		chunk := &StravaStreams{}
		
		// Copy relevant portions of each stream
		if len(data.Time) > start {
			chunkEnd := end
			if chunkEnd > len(data.Time) {
				chunkEnd = len(data.Time)
			}
			chunk.Time = data.Time[start:chunkEnd]
		}
		
		if len(data.Heartrate) > start {
			chunkEnd := end
			if chunkEnd > len(data.Heartrate) {
				chunkEnd = len(data.Heartrate)
			}
			chunk.Heartrate = data.Heartrate[start:chunkEnd]
		}
		
		if len(data.Watts) > start {
			chunkEnd := end
			if chunkEnd > len(data.Watts) {
				chunkEnd = len(data.Watts)
			}
			chunk.Watts = data.Watts[start:chunkEnd]
		}

		// Process chunk
		if err := processor(chunk); err != nil {
			return fmt.Errorf("failed to process chunk %d-%d: %w", start, end, err)
		}

		// Force garbage collection after each chunk to manage memory
		runtime.GC()
	}

	return nil
}

// EstimateMemoryUsage estimates memory usage for stream data
func (mop *MemoryOptimizedProcessor) EstimateMemoryUsage(data *StravaStreams) int64 {
	if data == nil {
		return 0
	}

	var totalBytes int64

	// Estimate memory for each stream type
	totalBytes += int64(len(data.Time) * 4)           // int32 = 4 bytes
	totalBytes += int64(len(data.Distance) * 8)       // float64 = 8 bytes
	totalBytes += int64(len(data.Heartrate) * 4)      // int32 = 4 bytes
	totalBytes += int64(len(data.Watts) * 4)          // int32 = 4 bytes
	totalBytes += int64(len(data.Cadence) * 4)        // int32 = 4 bytes
	totalBytes += int64(len(data.Altitude) * 8)       // float64 = 8 bytes
	totalBytes += int64(len(data.VelocitySmooth) * 8) // float64 = 8 bytes
	totalBytes += int64(len(data.Temp) * 4)           // int32 = 4 bytes
	totalBytes += int64(len(data.GradeSmooth) * 8)    // float64 = 8 bytes
	totalBytes += int64(len(data.Moving) * 1)         // bool = 1 byte
	totalBytes += int64(len(data.Latlng) * 16)        // [2]float64 = 16 bytes

	// Add overhead for slice headers and struct
	totalBytes += 1024 // Rough estimate for overhead

	return totalBytes
}

// ShouldUseChunkedProcessing determines if chunked processing should be used
func (mop *MemoryOptimizedProcessor) ShouldUseChunkedProcessing(data *StravaStreams) bool {
	estimatedMemoryMB := mop.EstimateMemoryUsage(data) / 1024 / 1024
	return int(estimatedMemoryMB) > mop.maxMemoryMB
}

// LogMemoryUsage logs current memory usage statistics
func (mop *MemoryOptimizedProcessor) LogMemoryUsage(operation string) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	log.Printf("Memory usage for %s:", operation)
	log.Printf("  Allocated: %d KB", memStats.Alloc/1024)
	log.Printf("  Total Allocated: %d KB", memStats.TotalAlloc/1024)
	log.Printf("  System Memory: %d KB", memStats.Sys/1024)
	log.Printf("  GC Cycles: %d", memStats.NumGC)
}