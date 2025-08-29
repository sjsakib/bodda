package services

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"
)

// StreamPerformanceMetrics tracks performance metrics for stream processing operations
type StreamPerformanceMetrics struct {
	mu                    sync.RWMutex
	ProcessingTimes       map[string][]time.Duration // Processing times by operation type
	MemoryUsage          map[string][]int64          // Memory usage by operation type
	DataSizes            map[string][]int            // Data sizes processed by operation type
	ErrorCounts          map[string]int              // Error counts by error type
	TotalOperations      int64                       // Total number of operations
	TotalProcessingTime  time.Duration               // Total processing time across all operations
	PeakMemoryUsage      int64                       // Peak memory usage observed
	AverageDataSize      float64                     // Average data size processed
	SuccessRate          float64                     // Success rate percentage
	LastResetTime        time.Time                   // When metrics were last reset
}

// StreamPerformanceMonitor provides performance monitoring for stream processing
type StreamPerformanceMonitor struct {
	metrics *StreamPerformanceMetrics
	enabled bool
}

// NewStreamPerformanceMonitor creates a new performance monitor
func NewStreamPerformanceMonitor(enabled bool) *StreamPerformanceMonitor {
	return &StreamPerformanceMonitor{
		metrics: &StreamPerformanceMetrics{
			ProcessingTimes: make(map[string][]time.Duration),
			MemoryUsage:     make(map[string][]int64),
			DataSizes:       make(map[string][]int),
			ErrorCounts:     make(map[string]int),
			LastResetTime:   time.Now(),
		},
		enabled: enabled,
	}
}

// OperationTimer tracks timing and memory usage for a single operation
type OperationTimer struct {
	monitor       *StreamPerformanceMonitor
	operationType string
	startTime     time.Time
	startMemory   int64
	dataSize      int
	ctx           context.Context
}

// StartOperation begins tracking a stream processing operation
func (spm *StreamPerformanceMonitor) StartOperation(ctx context.Context, operationType string, dataSize int) *OperationTimer {
	if !spm.enabled {
		return &OperationTimer{monitor: spm, ctx: ctx}
	}

	var memStats runtime.MemStats
	runtime.GC() // Force garbage collection for accurate memory measurement
	runtime.ReadMemStats(&memStats)

	return &OperationTimer{
		monitor:       spm,
		operationType: operationType,
		startTime:     time.Now(),
		startMemory:   int64(memStats.Alloc),
		dataSize:      dataSize,
		ctx:           ctx,
	}
}

// EndOperation completes tracking and records metrics
func (ot *OperationTimer) EndOperation(err error) {
	if !ot.monitor.enabled || ot.operationType == "" {
		return
	}

	duration := time.Since(ot.startTime)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	currentMemory := int64(memStats.Alloc)
	memoryUsed := currentMemory - ot.startMemory

	ot.monitor.recordMetrics(ot.operationType, duration, memoryUsed, ot.dataSize, err)

	// Log performance information
	if err != nil {
		log.Printf("Stream operation %s failed after %v (data size: %d, memory: %d bytes): %v",
			ot.operationType, duration, ot.dataSize, memoryUsed, err)
	} else {
		log.Printf("Stream operation %s completed in %v (data size: %d, memory: %d bytes)",
			ot.operationType, duration, ot.dataSize, memoryUsed)
	}
}

// recordMetrics records performance metrics for an operation
func (spm *StreamPerformanceMonitor) recordMetrics(operationType string, duration time.Duration, memoryUsed int64, dataSize int, err error) {
	spm.metrics.mu.Lock()
	defer spm.metrics.mu.Unlock()

	// Record processing time
	spm.metrics.ProcessingTimes[operationType] = append(spm.metrics.ProcessingTimes[operationType], duration)

	// Record memory usage
	spm.metrics.MemoryUsage[operationType] = append(spm.metrics.MemoryUsage[operationType], memoryUsed)

	// Record data size
	spm.metrics.DataSizes[operationType] = append(spm.metrics.DataSizes[operationType], dataSize)

	// Update totals
	spm.metrics.TotalOperations++
	spm.metrics.TotalProcessingTime += duration

	// Update peak memory usage
	if memoryUsed > spm.metrics.PeakMemoryUsage {
		spm.metrics.PeakMemoryUsage = memoryUsed
	}

	// Record errors
	if err != nil {
		errorType := "unknown_error"
		if streamErr, ok := err.(*StreamProcessingError); ok {
			errorType = streamErr.Type
		}
		spm.metrics.ErrorCounts[errorType]++
	}

	// Update derived metrics
	spm.updateDerivedMetrics()
}

// updateDerivedMetrics calculates derived performance metrics
func (spm *StreamPerformanceMonitor) updateDerivedMetrics() {
	// Calculate average data size
	totalDataSize := 0
	totalOperations := 0
	for _, sizes := range spm.metrics.DataSizes {
		for _, size := range sizes {
			totalDataSize += size
			totalOperations++
		}
	}
	if totalOperations > 0 {
		spm.metrics.AverageDataSize = float64(totalDataSize) / float64(totalOperations)
	}

	// Calculate success rate
	totalErrors := 0
	for _, count := range spm.metrics.ErrorCounts {
		totalErrors += count
	}
	if spm.metrics.TotalOperations > 0 {
		successfulOperations := spm.metrics.TotalOperations - int64(totalErrors)
		spm.metrics.SuccessRate = float64(successfulOperations) / float64(spm.metrics.TotalOperations) * 100
	}
}

// GetMetrics returns a copy of current performance metrics
func (spm *StreamPerformanceMonitor) GetMetrics() StreamPerformanceMetrics {
	spm.metrics.mu.RLock()
	defer spm.metrics.mu.RUnlock()

	// Create a deep copy of metrics
	metricsCopy := StreamPerformanceMetrics{
		ProcessingTimes:     make(map[string][]time.Duration),
		MemoryUsage:         make(map[string][]int64),
		DataSizes:           make(map[string][]int),
		ErrorCounts:         make(map[string]int),
		TotalOperations:     spm.metrics.TotalOperations,
		TotalProcessingTime: spm.metrics.TotalProcessingTime,
		PeakMemoryUsage:     spm.metrics.PeakMemoryUsage,
		AverageDataSize:     spm.metrics.AverageDataSize,
		SuccessRate:         spm.metrics.SuccessRate,
		LastResetTime:       spm.metrics.LastResetTime,
	}

	// Copy maps
	for k, v := range spm.metrics.ProcessingTimes {
		metricsCopy.ProcessingTimes[k] = make([]time.Duration, len(v))
		copy(metricsCopy.ProcessingTimes[k], v)
	}
	for k, v := range spm.metrics.MemoryUsage {
		metricsCopy.MemoryUsage[k] = make([]int64, len(v))
		copy(metricsCopy.MemoryUsage[k], v)
	}
	for k, v := range spm.metrics.DataSizes {
		metricsCopy.DataSizes[k] = make([]int, len(v))
		copy(metricsCopy.DataSizes[k], v)
	}
	for k, v := range spm.metrics.ErrorCounts {
		metricsCopy.ErrorCounts[k] = v
	}

	return metricsCopy
}

// ResetMetrics clears all performance metrics
func (spm *StreamPerformanceMonitor) ResetMetrics() {
	spm.metrics.mu.Lock()
	defer spm.metrics.mu.Unlock()

	spm.metrics.ProcessingTimes = make(map[string][]time.Duration)
	spm.metrics.MemoryUsage = make(map[string][]int64)
	spm.metrics.DataSizes = make(map[string][]int)
	spm.metrics.ErrorCounts = make(map[string]int)
	spm.metrics.TotalOperations = 0
	spm.metrics.TotalProcessingTime = 0
	spm.metrics.PeakMemoryUsage = 0
	spm.metrics.AverageDataSize = 0
	spm.metrics.SuccessRate = 0
	spm.metrics.LastResetTime = time.Now()
}

// GetOperationStats returns statistics for a specific operation type
func (spm *StreamPerformanceMonitor) GetOperationStats(operationType string) OperationStats {
	spm.metrics.mu.RLock()
	defer spm.metrics.mu.RUnlock()

	times := spm.metrics.ProcessingTimes[operationType]
	memory := spm.metrics.MemoryUsage[operationType]
	sizes := spm.metrics.DataSizes[operationType]

	if len(times) == 0 {
		return OperationStats{OperationType: operationType}
	}

	stats := OperationStats{
		OperationType:    operationType,
		TotalOperations:  len(times),
		AverageTime:      calculateAverageDuration(times),
		MinTime:          findMinDuration(times),
		MaxTime:          findMaxDuration(times),
		AverageMemory:    calculateAverageInt64(memory),
		PeakMemory:       findMaxInt64(memory),
		AverageDataSize:  calculateAverageInt(sizes),
		MaxDataSize:      findMaxInt(sizes),
		ThroughputPerSec: calculateThroughput(sizes, times),
	}

	return stats
}

// OperationStats contains performance statistics for a specific operation type
type OperationStats struct {
	OperationType    string        `json:"operation_type"`
	TotalOperations  int           `json:"total_operations"`
	AverageTime      time.Duration `json:"average_time"`
	MinTime          time.Duration `json:"min_time"`
	MaxTime          time.Duration `json:"max_time"`
	AverageMemory    int64         `json:"average_memory"`
	PeakMemory       int64         `json:"peak_memory"`
	AverageDataSize  float64       `json:"average_data_size"`
	MaxDataSize      int           `json:"max_data_size"`
	ThroughputPerSec float64       `json:"throughput_per_sec"` // Data points processed per second
}

// LogPerformanceSummary logs a summary of performance metrics
func (spm *StreamPerformanceMonitor) LogPerformanceSummary() {
	if !spm.enabled {
		return
	}

	metrics := spm.GetMetrics()
	
	log.Printf("=== Stream Processing Performance Summary ===")
	log.Printf("Total Operations: %d", metrics.TotalOperations)
	log.Printf("Total Processing Time: %v", metrics.TotalProcessingTime)
	log.Printf("Success Rate: %.2f%%", metrics.SuccessRate)
	log.Printf("Peak Memory Usage: %d bytes (%.2f MB)", metrics.PeakMemoryUsage, float64(metrics.PeakMemoryUsage)/1024/1024)
	log.Printf("Average Data Size: %.0f data points", metrics.AverageDataSize)
	
	if metrics.TotalOperations > 0 {
		avgTimePerOp := metrics.TotalProcessingTime / time.Duration(metrics.TotalOperations)
		log.Printf("Average Time Per Operation: %v", avgTimePerOp)
	}

	// Log stats for each operation type
	for operationType := range metrics.ProcessingTimes {
		stats := spm.GetOperationStats(operationType)
		log.Printf("--- %s Stats ---", operationType)
		log.Printf("  Operations: %d", stats.TotalOperations)
		log.Printf("  Avg Time: %v (Min: %v, Max: %v)", stats.AverageTime, stats.MinTime, stats.MaxTime)
		log.Printf("  Avg Memory: %d bytes (Peak: %d bytes)", stats.AverageMemory, stats.PeakMemory)
		log.Printf("  Avg Data Size: %.0f points (Max: %d points)", stats.AverageDataSize, stats.MaxDataSize)
		log.Printf("  Throughput: %.2f points/sec", stats.ThroughputPerSec)
	}

	// Log error summary
	if len(metrics.ErrorCounts) > 0 {
		log.Printf("--- Error Summary ---")
		for errorType, count := range metrics.ErrorCounts {
			log.Printf("  %s: %d occurrences", errorType, count)
		}
	}

	log.Printf("=== End Performance Summary ===")
}

// Helper functions for statistical calculations

func calculateAverageDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}

func findMinDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	min := durations[0]
	for _, d := range durations[1:] {
		if d < min {
			min = d
		}
	}
	return min
}

func findMaxDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	max := durations[0]
	for _, d := range durations[1:] {
		if d > max {
			max = d
		}
	}
	return max
}

func calculateAverageInt64(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	
	var total int64
	for _, v := range values {
		total += v
	}
	return total / int64(len(values))
}

func findMaxInt64(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func calculateAverageInt(values []int) float64 {
	if len(values) == 0 {
		return 0
	}
	
	total := 0
	for _, v := range values {
		total += v
	}
	return float64(total) / float64(len(values))
}

func findMaxInt(values []int) int {
	if len(values) == 0 {
		return 0
	}
	
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func calculateThroughput(dataSizes []int, processingTimes []time.Duration) float64 {
	if len(dataSizes) == 0 || len(processingTimes) == 0 || len(dataSizes) != len(processingTimes) {
		return 0
	}
	
	totalDataPoints := 0
	var totalTime time.Duration
	
	for i := 0; i < len(dataSizes); i++ {
		totalDataPoints += dataSizes[i]
		totalTime += processingTimes[i]
	}
	
	if totalTime == 0 {
		return 0
	}
	
	return float64(totalDataPoints) / totalTime.Seconds()
}

// StreamProcessingOptimizer provides optimization recommendations based on performance metrics
type StreamProcessingOptimizer struct {
	monitor *StreamPerformanceMonitor
}

// NewStreamProcessingOptimizer creates a new optimizer
func NewStreamProcessingOptimizer(monitor *StreamPerformanceMonitor) *StreamProcessingOptimizer {
	return &StreamProcessingOptimizer{
		monitor: monitor,
	}
}

// OptimizationRecommendation represents a performance optimization suggestion
type OptimizationRecommendation struct {
	Type        string  `json:"type"`
	Priority    string  `json:"priority"` // "high", "medium", "low"
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Metric      float64 `json:"metric,omitempty"`
}

// GetOptimizationRecommendations analyzes performance metrics and provides optimization suggestions
func (spo *StreamProcessingOptimizer) GetOptimizationRecommendations() []OptimizationRecommendation {
	metrics := spo.monitor.GetMetrics()
	var recommendations []OptimizationRecommendation

	// Check success rate
	if metrics.SuccessRate < 95.0 && metrics.TotalOperations > 10 {
		recommendations = append(recommendations, OptimizationRecommendation{
			Type:        "error_rate",
			Priority:    "high",
			Description: "High error rate detected in stream processing operations",
			Impact:      "Implement better error handling and fallback mechanisms",
			Metric:      metrics.SuccessRate,
		})
	}

	// Check memory usage
	if metrics.PeakMemoryUsage > 100*1024*1024 { // 100MB
		recommendations = append(recommendations, OptimizationRecommendation{
			Type:        "memory_usage",
			Priority:    "medium",
			Description: "High peak memory usage detected",
			Impact:      "Consider implementing streaming processing or data chunking",
			Metric:      float64(metrics.PeakMemoryUsage),
		})
	}

	// Check processing time for specific operations
	for operationType, times := range metrics.ProcessingTimes {
		if len(times) > 5 {
			avgTime := calculateAverageDuration(times)
			maxTime := findMaxDuration(times)
			
			// Check for slow operations
			if avgTime > 5*time.Second {
				recommendations = append(recommendations, OptimizationRecommendation{
					Type:        "slow_operation",
					Priority:    "medium",
					Description: fmt.Sprintf("Operation '%s' has slow average processing time", operationType),
					Impact:      "Optimize algorithms or implement caching",
					Metric:      avgTime.Seconds(),
				})
			}
			
			// Check for inconsistent performance
			if maxTime > avgTime*3 {
				recommendations = append(recommendations, OptimizationRecommendation{
					Type:        "inconsistent_performance",
					Priority:    "low",
					Description: fmt.Sprintf("Operation '%s' has inconsistent performance", operationType),
					Impact:      "Investigate performance bottlenecks and optimize worst-case scenarios",
					Metric:      float64(maxTime) / float64(avgTime),
				})
			}
		}
	}

	// Check data size efficiency
	if metrics.AverageDataSize > 5000 {
		recommendations = append(recommendations, OptimizationRecommendation{
			Type:        "large_data_size",
			Priority:    "medium",
			Description: "Large average data sizes being processed",
			Impact:      "Consider implementing more aggressive pagination or data filtering",
			Metric:      metrics.AverageDataSize,
		})
	}

	return recommendations
}

// LogOptimizationRecommendations logs optimization recommendations
func (spo *StreamProcessingOptimizer) LogOptimizationRecommendations() {
	recommendations := spo.GetOptimizationRecommendations()
	
	if len(recommendations) == 0 {
		log.Printf("Stream processing performance is optimal - no recommendations")
		return
	}

	log.Printf("=== Stream Processing Optimization Recommendations ===")
	
	for _, rec := range recommendations {
		log.Printf("[%s] %s: %s", strings.ToUpper(rec.Priority), rec.Type, rec.Description)
		log.Printf("  Impact: %s", rec.Impact)
		if rec.Metric > 0 {
			log.Printf("  Metric: %.2f", rec.Metric)
		}
	}
	
	log.Printf("=== End Optimization Recommendations ===")
}