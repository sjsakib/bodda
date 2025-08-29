package monitoring

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics holds application metrics
type Metrics struct {
	mu                sync.RWMutex
	RequestCount      map[string]int64  `json:"request_count"`
	RequestDuration   map[string]int64  `json:"request_duration_ms"`
	ErrorCount        map[string]int64  `json:"error_count"`
	ActiveConnections int64             `json:"active_connections"`
	StartTime         time.Time         `json:"start_time"`
	
	// Custom application metrics
	StravaAPICalls    int64 `json:"strava_api_calls"`
	OpenAICalls       int64 `json:"openai_calls"`
	DatabaseQueries   int64 `json:"database_queries"`
	ActiveSessions    int64 `json:"active_sessions"`
}

// MetricsCollector manages application metrics
type MetricsCollector struct {
	metrics *Metrics
	logger  *Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *Logger) *MetricsCollector {
	return &MetricsCollector{
		metrics: &Metrics{
			RequestCount:    make(map[string]int64),
			RequestDuration: make(map[string]int64),
			ErrorCount:      make(map[string]int64),
			StartTime:       time.Now(),
		},
		logger: logger,
	}
}

// RecordRequest records HTTP request metrics
func (mc *MetricsCollector) RecordRequest(method, path string, statusCode int, duration time.Duration) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()
	
	key := fmt.Sprintf("%s %s", method, path)
	mc.metrics.RequestCount[key]++
	mc.metrics.RequestDuration[key] += duration.Milliseconds()
	
	if statusCode >= 400 {
		errorKey := fmt.Sprintf("%s %d", key, statusCode)
		mc.metrics.ErrorCount[errorKey]++
	}
}

// RecordStravaAPICall records Strava API call
func (mc *MetricsCollector) RecordStravaAPICall() {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()
	mc.metrics.StravaAPICalls++
}

// RecordOpenAICall records OpenAI API call
func (mc *MetricsCollector) RecordOpenAICall() {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()
	mc.metrics.OpenAICalls++
}

// RecordDatabaseQuery records database query
func (mc *MetricsCollector) RecordDatabaseQuery() {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()
	mc.metrics.DatabaseQueries++
}

// IncrementActiveConnections increments active connection count
func (mc *MetricsCollector) IncrementActiveConnections() {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()
	mc.metrics.ActiveConnections++
}

// DecrementActiveConnections decrements active connection count
func (mc *MetricsCollector) DecrementActiveConnections() {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()
	mc.metrics.ActiveConnections--
}

// SetActiveSessions sets the number of active chat sessions
func (mc *MetricsCollector) SetActiveSessions(count int64) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()
	mc.metrics.ActiveSessions = count
}

// GetMetrics returns a copy of current metrics
func (mc *MetricsCollector) GetMetrics() *Metrics {
	mc.metrics.mu.RLock()
	defer mc.metrics.mu.RUnlock()
	
	// Create a deep copy
	copy := &Metrics{
		RequestCount:      make(map[string]int64),
		RequestDuration:   make(map[string]int64),
		ErrorCount:        make(map[string]int64),
		ActiveConnections: mc.metrics.ActiveConnections,
		StartTime:         mc.metrics.StartTime,
		StravaAPICalls:    mc.metrics.StravaAPICalls,
		OpenAICalls:       mc.metrics.OpenAICalls,
		DatabaseQueries:   mc.metrics.DatabaseQueries,
		ActiveSessions:    mc.metrics.ActiveSessions,
	}
	
	for k, v := range mc.metrics.RequestCount {
		copy.RequestCount[k] = v
	}
	for k, v := range mc.metrics.RequestDuration {
		copy.RequestDuration[k] = v
	}
	for k, v := range mc.metrics.ErrorCount {
		copy.ErrorCount[k] = v
	}
	
	return copy
}

// GetSystemMetrics returns system-level metrics
func (mc *MetricsCollector) GetSystemMetrics() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"memory": map[string]interface{}{
			"alloc_mb":      bToMb(m.Alloc),
			"total_alloc_mb": bToMb(m.TotalAlloc),
			"sys_mb":        bToMb(m.Sys),
			"num_gc":        m.NumGC,
		},
		"goroutines": runtime.NumGoroutine(),
		"uptime_seconds": time.Since(mc.metrics.StartTime).Seconds(),
	}
}

// bToMb converts bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// Gin middleware for metrics collection
func (mc *MetricsCollector) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Increment active connections
		mc.IncrementActiveConnections()
		defer mc.DecrementActiveConnections()
		
		// Process request
		c.Next()
		
		// Record metrics
		duration := time.Since(start)
		mc.RecordRequest(c.Request.Method, c.FullPath(), c.Writer.Status(), duration)
	}
}

// Health check endpoint data
type HealthStatus struct {
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Uptime      string                 `json:"uptime"`
	Version     string                 `json:"version,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	Checks      map[string]interface{} `json:"checks"`
}

// HealthChecker performs health checks
type HealthChecker struct {
	startTime time.Time
	version   string
	env       string
	logger    *Logger
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(version, env string, logger *Logger) *HealthChecker {
	return &HealthChecker{
		startTime: time.Now(),
		version:   version,
		env:       env,
		logger:    logger,
	}
}

// CheckHealth performs comprehensive health check
func (hc *HealthChecker) CheckHealth(ctx context.Context) *HealthStatus {
	checks := make(map[string]interface{})
	overallStatus := "healthy"
	
	// Add individual health checks here
	checks["database"] = "healthy" // This would be implemented with actual DB ping
	checks["memory"] = hc.checkMemoryUsage()
	
	// If any check fails, set overall status to unhealthy
	for _, check := range checks {
		if check == "unhealthy" {
			overallStatus = "unhealthy"
			break
		}
	}
	
	return &HealthStatus{
		Status:      overallStatus,
		Timestamp:   time.Now(),
		Uptime:      time.Since(hc.startTime).String(),
		Version:     hc.version,
		Environment: hc.env,
		Checks:      checks,
	}
}

// checkMemoryUsage checks if memory usage is within acceptable limits
func (hc *HealthChecker) checkMemoryUsage() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Consider unhealthy if using more than 1GB
	if m.Alloc > 1024*1024*1024 {
		return "unhealthy"
	}
	return "healthy"
}

// Global metrics collector
var globalMetrics *MetricsCollector

// InitGlobalMetrics initializes the global metrics collector
func InitGlobalMetrics(logger *Logger) {
	globalMetrics = NewMetricsCollector(logger)
}

// GetMetricsCollector returns the global metrics collector
func GetMetricsCollector() *MetricsCollector {
	return globalMetrics
}

// Convenience functions for recording metrics
func RecordStravaAPICall() {
	if globalMetrics != nil {
		globalMetrics.RecordStravaAPICall()
	}
}

func RecordOpenAICall() {
	if globalMetrics != nil {
		globalMetrics.RecordOpenAICall()
	}
}

func RecordDatabaseQuery() {
	if globalMetrics != nil {
		globalMetrics.RecordDatabaseQuery()
	}
}