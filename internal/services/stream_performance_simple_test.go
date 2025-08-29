package services

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"bodda/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStreamPerformanceMonitoring tests the performance monitoring functionality
func TestStreamPerformanceMonitoring(t *testing.T) {
	monitor := NewStreamPerformanceMonitor(true)
	
	// Test basic operation tracking
	ctx := context.Background()
	
	// Simulate some operations
	for i := 0; i < 5; i++ {
		timer := monitor.StartOperation(ctx, "test_operation", 1000+i*100)
		
		// Simulate processing time
		time.Sleep(time.Duration(10+i*5) * time.Millisecond)
		
		var err error
		if i == 2 {
			err = fmt.Errorf("simulated error")
		}
		
		timer.EndOperation(err)
	}
	
	// Get metrics
	metrics := monitor.GetMetrics()
	
	assert.Equal(t, int64(5), metrics.TotalOperations)
	assert.Equal(t, 80.0, metrics.SuccessRate) // 4 success out of 5
	assert.Greater(t, metrics.TotalProcessingTime, time.Duration(0))
	assert.Greater(t, metrics.AverageDataSize, 0.0)
	
	// Test operation stats
	stats := monitor.GetOperationStats("test_operation")
	assert.Equal(t, "test_operation", stats.OperationType)
	assert.Equal(t, 5, stats.TotalOperations)
	assert.Greater(t, stats.AverageTime, time.Duration(0))
	
	// Test error tracking
	assert.Equal(t, 1, metrics.ErrorCounts["unknown_error"])
	
	monitor.LogPerformanceSummary()
}

// TestOptimizedStatisticsCalculator tests the optimized statistics functionality
func TestOptimizedStatisticsCalculator(t *testing.T) {
	calculator := NewOptimizedStatisticsCalculator()
	
	// Test with small dataset (should use simple calculation)
	smallData := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	stats := calculator.CalculateMetricStatsOptimized(context.Background(), smallData)
	
	assert.Equal(t, 1.0, stats.Min)
	assert.Equal(t, 10.0, stats.Max)
	assert.Equal(t, 5.5, stats.Mean)
	assert.Equal(t, 10, stats.Count)
	
	// Test with larger dataset (should use parallel calculation)
	largeData := make([]float64, 5000)
	for i := 0; i < len(largeData); i++ {
		largeData[i] = float64(i % 100)
	}
	
	startTime := time.Now()
	largeStats := calculator.CalculateMetricStatsOptimized(context.Background(), largeData)
	duration := time.Since(startTime)
	
	assert.Equal(t, 0.0, largeStats.Min)
	assert.Equal(t, 99.0, largeStats.Max)
	assert.Equal(t, 5000, largeStats.Count)
	assert.Less(t, duration, 5*time.Second) // Should complete quickly
	
	t.Logf("Large dataset statistics calculated in %v", duration)
}

// TestInflectionPointDetection tests the optimized inflection point detection
func TestInflectionPointDetection(t *testing.T) {
	detector := NewOptimizedInflectionPointDetector()
	
	// Create test data with known inflection points
	data := make([]float64, 1000)
	for i := 0; i < len(data); i++ {
		// Create a more pronounced pattern with clear inflection points
		if i < 200 {
			data[i] = float64(i) // Increasing
		} else if i < 400 {
			data[i] = 200 - float64(i-200) // Decreasing
		} else if i < 600 {
			data[i] = float64(i-400) // Increasing again
		} else {
			data[i] = 200 - float64(i-600) // Decreasing again
		}
	}
	
	startTime := time.Now()
	points := detector.DetectInflectionPointsOptimized(context.Background(), data, "test_metric")
	duration := time.Since(startTime)
	
	assert.Greater(t, len(points), 0, "Should detect some inflection points")
	assert.Less(t, duration, 2*time.Second, "Detection should complete quickly")
	
	// Verify inflection point structure
	for _, point := range points {
		assert.Greater(t, point.Index, 0)
		assert.Contains(t, []string{"increase", "decrease"}, point.Direction)
		assert.Equal(t, "test_metric", point.Metric)
		assert.Greater(t, point.Magnitude, 0.0)
	}
	
	t.Logf("Detected %d inflection points in %v", len(points), duration)
}

// TestMemoryOptimizedProcessing tests memory-optimized processing
func TestMemoryOptimizedProcessing(t *testing.T) {
	processor := NewMemoryOptimizedProcessor(50) // 50MB limit
	
	// Create test stream data
	streams := &StravaStreams{
		Time:      make([]int, 10000),
		Heartrate: make([]int, 10000),
		Watts:     make([]int, 10000),
	}
	
	for i := 0; i < 10000; i++ {
		streams.Time[i] = i
		streams.Heartrate[i] = 120 + i%50
		streams.Watts[i] = 150 + i%100
	}
	
	// Test memory estimation
	estimatedMemory := processor.EstimateMemoryUsage(streams)
	assert.Greater(t, estimatedMemory, int64(0))
	
	t.Logf("Estimated memory usage: %d bytes", estimatedMemory)
	
	// Test chunked processing
	chunkCount := 0
	totalProcessed := 0
	
	err := processor.ProcessStreamDataInChunks(context.Background(), streams, func(chunk *StravaStreams) error {
		chunkCount++
		totalProcessed += len(chunk.Time)
		
		// Verify chunk is not empty
		assert.Greater(t, len(chunk.Time), 0)
		
		return nil
	})
	
	require.NoError(t, err)
	assert.Equal(t, 10000, totalProcessed)
	assert.Greater(t, chunkCount, 0)
	
	t.Logf("Processed %d data points in %d chunks", totalProcessed, chunkCount)
}

// TestStreamProcessingOptimizer tests optimization recommendations
func TestStreamProcessingOptimizer(t *testing.T) {
	monitor := NewStreamPerformanceMonitor(true)
	optimizer := NewStreamProcessingOptimizer(monitor)
	
	ctx := context.Background()
	
	// Simulate operations with different performance characteristics
	
	// Good operations
	for i := 0; i < 10; i++ {
		timer := monitor.StartOperation(ctx, "good_operation", 1000)
		time.Sleep(50 * time.Millisecond)
		timer.EndOperation(nil)
	}
	
	// Failed operations to trigger error rate recommendation
	for i := 0; i < 8; i++ {
		timer := monitor.StartOperation(ctx, "failing_operation", 2000)
		time.Sleep(10 * time.Millisecond)
		timer.EndOperation(fmt.Errorf("test error"))
	}
	
	// Get recommendations
	recommendations := optimizer.GetOptimizationRecommendations()
	
	assert.NotEmpty(t, recommendations, "Should have recommendations")
	
	// Should have recommendations for error rate
	hasErrorRateRecommendation := false
	for _, rec := range recommendations {
		if rec.Type == "error_rate" {
			hasErrorRateRecommendation = true
		}
		t.Logf("Recommendation: %s (%s) - %s", rec.Type, rec.Priority, rec.Description)
	}
	
	assert.True(t, hasErrorRateRecommendation, "Should recommend addressing error rate")
	
	optimizer.LogOptimizationRecommendations()
}

// TestStreamProcessingLogger tests the logging functionality
func TestStreamProcessingLogger(t *testing.T) {
	logger, err := NewStreamProcessingLogger(true, "/tmp/test_logs", LogLevelInfo)
	require.NoError(t, err)
	defer logger.Close()
	
	ctx := context.Background()
	
	// Log some operations
	logger.LogOperation(ctx, "test_operation", LogLevelInfo, map[string]interface{}{
		"activity_id":      int64(12345),
		"user_id":         "test-user",
		"duration":        100 * time.Millisecond,
		"data_size":       1000,
		"processing_mode": "derived",
	})
	
	// Log an error
	logger.LogError(ctx, "error_operation", fmt.Errorf("test error"), map[string]interface{}{
		"activity_id": int64(67890),
		"user_id":    "test-user",
	})
	
	// Get recent operations
	recent := logger.GetRecentOperations(10)
	assert.Len(t, recent, 2)
	
	// Get error operations
	errors := logger.GetErrorOperations(10)
	assert.Len(t, errors, 1)
	assert.False(t, errors[0].Success)
	
	// Generate report
	report := logger.GenerateOperationReport(time.Now().Add(-1 * time.Hour))
	assert.Equal(t, 2, report.TotalOperations)
	assert.Equal(t, 1, report.SuccessfulOperations)
	assert.Equal(t, 1, report.FailedOperations)
	
	logger.LogOperationReport(time.Now().Add(-1 * time.Hour))
}

// TestStreamProcessorWithPerformanceMonitoring tests integration with existing stream processor
func TestStreamProcessorWithPerformanceMonitoring(t *testing.T) {
	cfg := &config.Config{
		StreamProcessing: config.StreamProcessingConfig{
			MaxContextTokens:  15000,
			TokenPerCharRatio: 0.25,
			DefaultPageSize:   1000,
			MaxPageSize:       5000,
			RedactionEnabled:  false,
		},
	}
	
	processor := NewStreamProcessor(cfg)
	
	// Test with different data sizes
	testSizes := []int{100, 1000, 5000}
	
	for _, size := range testSizes {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			streams := &StravaStreams{
				Time:      make([]int, size),
				Heartrate: make([]int, size),
				Watts:     make([]int, size),
			}
			
			for i := 0; i < size; i++ {
				streams.Time[i] = i
				streams.Heartrate[i] = 120 + i%50
				streams.Watts[i] = 150 + i%100
			}
			
			startTime := time.Now()
			result, err := processor.ProcessStreamOutput(streams, fmt.Sprintf("test_%d", size))
			duration := time.Since(startTime)
			
			require.NoError(t, err)
			assert.NotNil(t, result)
			
			t.Logf("Size %d processed in %v", size, duration)
			
			// Performance should be reasonable
			assert.Less(t, duration, 10*time.Second)
		})
	}
}

// BenchmarkStreamProcessingComponents benchmarks key components
func BenchmarkStreamProcessingComponents(b *testing.B) {
	b.Run("StatisticsCalculation", func(b *testing.B) {
		calculator := NewOptimizedStatisticsCalculator()
		data := make([]float64, 5000)
		for i := 0; i < len(data); i++ {
			data[i] = float64(i % 1000)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = calculator.CalculateMetricStatsOptimized(context.Background(), data)
		}
	})
	
	b.Run("InflectionPointDetection", func(b *testing.B) {
		detector := NewOptimizedInflectionPointDetector()
		data := make([]float64, 2000)
		for i := 0; i < len(data); i++ {
			data[i] = 50 + 20*math.Sin(float64(i)/100.0)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = detector.DetectInflectionPointsOptimized(context.Background(), data, "test")
		}
	})
	
	b.Run("PerformanceMonitoring", func(b *testing.B) {
		monitor := NewStreamPerformanceMonitor(true)
		ctx := context.Background()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			timer := monitor.StartOperation(ctx, "benchmark", 1000)
			time.Sleep(1 * time.Millisecond)
			timer.EndOperation(nil)
		}
	})
}