package monitoring

import (
	"context"
	"sync"
	"time"
)

// ToolPerformanceMetrics tracks performance metrics for tool executions
type ToolPerformanceMetrics struct {
	mu                    sync.RWMutex
	ExecutionCount        map[string]int64          `json:"execution_count"`
	SuccessCount          map[string]int64          `json:"success_count"`
	ErrorCount            map[string]int64          `json:"error_count"`
	TimeoutCount          map[string]int64          `json:"timeout_count"`
	TotalExecutionTime    map[string]int64          `json:"total_execution_time_ms"`
	AverageExecutionTime  map[string]float64        `json:"average_execution_time_ms"`
	MaxExecutionTime      map[string]int64          `json:"max_execution_time_ms"`
	MinExecutionTime      map[string]int64          `json:"min_execution_time_ms"`
	ConcurrentExecutions  int64                     `json:"concurrent_executions"`
	QueueDepth            int64                     `json:"queue_depth"`
	ActiveExecutions      map[string]int64          `json:"active_executions"`
	UserExecutionCount    map[string]int64          `json:"user_execution_count"`
	LastExecutionTime     map[string]time.Time      `json:"last_execution_time"`
	PerformanceAlerts     []PerformanceAlert        `json:"performance_alerts"`
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	Timestamp   time.Time `json:"timestamp"`
	AlertType   string    `json:"alert_type"`
	ToolName    string    `json:"tool_name,omitempty"`
	UserID      string    `json:"user_id,omitempty"`
	Message     string    `json:"message"`
	Severity    string    `json:"severity"`
	Value       float64   `json:"value,omitempty"`
	Threshold   float64   `json:"threshold,omitempty"`
}

// PerformanceThresholds defines thresholds for performance alerts
type PerformanceThresholds struct {
	MaxExecutionTimeMs     int64   `json:"max_execution_time_ms"`
	MaxConcurrentExecs     int64   `json:"max_concurrent_executions"`
	MaxQueueDepth          int64   `json:"max_queue_depth"`
	MaxErrorRatePercent    float64 `json:"max_error_rate_percent"`
	MaxTimeoutRatePercent  float64 `json:"max_timeout_rate_percent"`
	AlertRetentionHours    int     `json:"alert_retention_hours"`
}

// ToolPerformanceTracker manages performance metrics and alerting for tool executions
type ToolPerformanceTracker struct {
	metrics    *ToolPerformanceMetrics
	thresholds PerformanceThresholds
	logger     *Logger
	alertChan  chan PerformanceAlert
}

// NewToolPerformanceTracker creates a new tool performance tracker
func NewToolPerformanceTracker(logger *Logger, thresholds PerformanceThresholds) *ToolPerformanceTracker {
	tracker := &ToolPerformanceTracker{
		metrics: &ToolPerformanceMetrics{
			ExecutionCount:       make(map[string]int64),
			SuccessCount:         make(map[string]int64),
			ErrorCount:           make(map[string]int64),
			TimeoutCount:         make(map[string]int64),
			TotalExecutionTime:   make(map[string]int64),
			AverageExecutionTime: make(map[string]float64),
			MaxExecutionTime:     make(map[string]int64),
			MinExecutionTime:     make(map[string]int64),
			ActiveExecutions:     make(map[string]int64),
			UserExecutionCount:   make(map[string]int64),
			LastExecutionTime:    make(map[string]time.Time),
			PerformanceAlerts:    make([]PerformanceAlert, 0),
		},
		thresholds: thresholds,
		logger:     logger,
		alertChan:  make(chan PerformanceAlert, 100),
	}

	// Start alert processing goroutine
	go tracker.processAlerts()
	
	// Start cleanup goroutine for old alerts
	go tracker.cleanupOldAlerts()

	return tracker
}

// RecordToolExecutionStart records the start of a tool execution
func (tpt *ToolPerformanceTracker) RecordToolExecutionStart(ctx context.Context, toolName string, userID string) {
	tpt.metrics.mu.Lock()
	defer tpt.metrics.mu.Unlock()

	tpt.metrics.ConcurrentExecutions++
	tpt.metrics.ActiveExecutions[toolName]++
	
	// Check concurrent execution threshold
	if tpt.metrics.ConcurrentExecutions > tpt.thresholds.MaxConcurrentExecs {
		alert := PerformanceAlert{
			Timestamp: time.Now(),
			AlertType: "high_concurrent_executions",
			Message:   "High number of concurrent tool executions",
			Severity:  "warning",
			Value:     float64(tpt.metrics.ConcurrentExecutions),
			Threshold: float64(tpt.thresholds.MaxConcurrentExecs),
		}
		tpt.sendAlert(alert)
	}
}

// RecordToolExecutionEnd records the completion of a tool execution
func (tpt *ToolPerformanceTracker) RecordToolExecutionEnd(ctx context.Context, toolName string, userID string, duration time.Duration, success bool, isTimeout bool) {
	tpt.metrics.mu.Lock()
	defer tpt.metrics.mu.Unlock()

	durationMs := duration.Milliseconds()
	
	// Update basic counters
	tpt.metrics.ExecutionCount[toolName]++
	tpt.metrics.ConcurrentExecutions--
	tpt.metrics.ActiveExecutions[toolName]--
	tpt.metrics.UserExecutionCount[userID]++
	tpt.metrics.LastExecutionTime[toolName] = time.Now()

	// Update success/error counters
	if success {
		tpt.metrics.SuccessCount[toolName]++
	} else {
		tpt.metrics.ErrorCount[toolName]++
		if isTimeout {
			tpt.metrics.TimeoutCount[toolName]++
		}
	}

	// Update timing metrics
	tpt.metrics.TotalExecutionTime[toolName] += durationMs
	
	// Update min/max execution times
	if tpt.metrics.MaxExecutionTime[toolName] < durationMs {
		tpt.metrics.MaxExecutionTime[toolName] = durationMs
	}
	if tpt.metrics.MinExecutionTime[toolName] == 0 || tpt.metrics.MinExecutionTime[toolName] > durationMs {
		tpt.metrics.MinExecutionTime[toolName] = durationMs
	}

	// Calculate average execution time
	if tpt.metrics.ExecutionCount[toolName] > 0 {
		tpt.metrics.AverageExecutionTime[toolName] = float64(tpt.metrics.TotalExecutionTime[toolName]) / float64(tpt.metrics.ExecutionCount[toolName])
	}

	// Check performance thresholds
	tpt.checkPerformanceThresholds(toolName, userID, durationMs, success, isTimeout)
}

// RecordQueueDepth records the current queue depth
func (tpt *ToolPerformanceTracker) RecordQueueDepth(depth int64) {
	tpt.metrics.mu.Lock()
	defer tpt.metrics.mu.Unlock()

	tpt.metrics.QueueDepth = depth
	
	// Check queue depth threshold
	if depth > tpt.thresholds.MaxQueueDepth {
		alert := PerformanceAlert{
			Timestamp: time.Now(),
			AlertType: "high_queue_depth",
			Message:   "Tool execution queue depth is high",
			Severity:  "warning",
			Value:     float64(depth),
			Threshold: float64(tpt.thresholds.MaxQueueDepth),
		}
		tpt.sendAlert(alert)
	}
}

// checkPerformanceThresholds checks various performance thresholds and generates alerts
func (tpt *ToolPerformanceTracker) checkPerformanceThresholds(toolName string, userID string, durationMs int64, success bool, isTimeout bool) {
	// Check execution time threshold
	if durationMs > tpt.thresholds.MaxExecutionTimeMs {
		alert := PerformanceAlert{
			Timestamp: time.Now(),
			AlertType: "slow_execution",
			ToolName:  toolName,
			UserID:    userID,
			Message:   "Tool execution exceeded time threshold",
			Severity:  "warning",
			Value:     float64(durationMs),
			Threshold: float64(tpt.thresholds.MaxExecutionTimeMs),
		}
		tpt.sendAlert(alert)
	}

	// Check error rate threshold (only if we have enough executions)
	if tpt.metrics.ExecutionCount[toolName] >= 10 {
		errorRate := float64(tpt.metrics.ErrorCount[toolName]) / float64(tpt.metrics.ExecutionCount[toolName]) * 100
		if errorRate > tpt.thresholds.MaxErrorRatePercent {
			alert := PerformanceAlert{
				Timestamp: time.Now(),
				AlertType: "high_error_rate",
				ToolName:  toolName,
				Message:   "Tool has high error rate",
				Severity:  "error",
				Value:     errorRate,
				Threshold: tpt.thresholds.MaxErrorRatePercent,
			}
			tpt.sendAlert(alert)
		}
	}

	// Check timeout rate threshold (only if we have enough executions)
	if tpt.metrics.ExecutionCount[toolName] >= 10 {
		timeoutRate := float64(tpt.metrics.TimeoutCount[toolName]) / float64(tpt.metrics.ExecutionCount[toolName]) * 100
		if timeoutRate > tpt.thresholds.MaxTimeoutRatePercent {
			alert := PerformanceAlert{
				Timestamp: time.Now(),
				AlertType: "high_timeout_rate",
				ToolName:  toolName,
				Message:   "Tool has high timeout rate",
				Severity:  "error",
				Value:     timeoutRate,
				Threshold: tpt.thresholds.MaxTimeoutRatePercent,
			}
			tpt.sendAlert(alert)
		}
	}
}

// sendAlert sends an alert through the alert channel
func (tpt *ToolPerformanceTracker) sendAlert(alert PerformanceAlert) {
	select {
	case tpt.alertChan <- alert:
		// Alert sent successfully
	default:
		// Alert channel is full, log the issue
		tpt.logger.Warn("Alert channel is full, dropping alert",
			"alert_type", alert.AlertType,
			"tool_name", alert.ToolName,
			"severity", alert.Severity,
		)
	}
}

// processAlerts processes alerts from the alert channel
func (tpt *ToolPerformanceTracker) processAlerts() {
	for alert := range tpt.alertChan {
		// Add alert to metrics
		tpt.metrics.mu.Lock()
		tpt.metrics.PerformanceAlerts = append(tpt.metrics.PerformanceAlerts, alert)
		tpt.metrics.mu.Unlock()

		// Log the alert
		logLevel := "info"
		switch alert.Severity {
		case "warning":
			logLevel = "warn"
		case "error":
			logLevel = "error"
		}

		fields := []interface{}{
			"alert_type", alert.AlertType,
			"tool_name", alert.ToolName,
			"user_id", alert.UserID,
			"severity", alert.Severity,
			"message", alert.Message,
		}

		if alert.Value != 0 {
			fields = append(fields, "value", alert.Value)
		}
		if alert.Threshold != 0 {
			fields = append(fields, "threshold", alert.Threshold)
		}

		switch logLevel {
		case "warn":
			tpt.logger.Warn("Performance alert", fields...)
		case "error":
			tpt.logger.Error("Performance alert", fields...)
		default:
			tpt.logger.Info("Performance alert", fields...)
		}
	}
}

// cleanupOldAlerts removes old alerts based on retention policy
func (tpt *ToolPerformanceTracker) cleanupOldAlerts() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		tpt.metrics.mu.Lock()
		cutoff := time.Now().Add(-time.Duration(tpt.thresholds.AlertRetentionHours) * time.Hour)
		
		// Filter out old alerts
		var filteredAlerts []PerformanceAlert
		for _, alert := range tpt.metrics.PerformanceAlerts {
			if alert.Timestamp.After(cutoff) {
				filteredAlerts = append(filteredAlerts, alert)
			}
		}
		
		tpt.metrics.PerformanceAlerts = filteredAlerts
		tpt.metrics.mu.Unlock()
	}
}

// GetMetrics returns a copy of current performance metrics
func (tpt *ToolPerformanceTracker) GetMetrics() *ToolPerformanceMetrics {
	tpt.metrics.mu.RLock()
	defer tpt.metrics.mu.RUnlock()

	// Create a deep copy
	copy := &ToolPerformanceMetrics{
		ExecutionCount:       make(map[string]int64),
		SuccessCount:         make(map[string]int64),
		ErrorCount:           make(map[string]int64),
		TimeoutCount:         make(map[string]int64),
		TotalExecutionTime:   make(map[string]int64),
		AverageExecutionTime: make(map[string]float64),
		MaxExecutionTime:     make(map[string]int64),
		MinExecutionTime:     make(map[string]int64),
		ActiveExecutions:     make(map[string]int64),
		UserExecutionCount:   make(map[string]int64),
		LastExecutionTime:    make(map[string]time.Time),
		ConcurrentExecutions: tpt.metrics.ConcurrentExecutions,
		QueueDepth:           tpt.metrics.QueueDepth,
		PerformanceAlerts:    make([]PerformanceAlert, len(tpt.metrics.PerformanceAlerts)),
	}

	// Copy maps
	for k, v := range tpt.metrics.ExecutionCount {
		copy.ExecutionCount[k] = v
	}
	for k, v := range tpt.metrics.SuccessCount {
		copy.SuccessCount[k] = v
	}
	for k, v := range tpt.metrics.ErrorCount {
		copy.ErrorCount[k] = v
	}
	for k, v := range tpt.metrics.TimeoutCount {
		copy.TimeoutCount[k] = v
	}
	for k, v := range tpt.metrics.TotalExecutionTime {
		copy.TotalExecutionTime[k] = v
	}
	for k, v := range tpt.metrics.AverageExecutionTime {
		copy.AverageExecutionTime[k] = v
	}
	for k, v := range tpt.metrics.MaxExecutionTime {
		copy.MaxExecutionTime[k] = v
	}
	for k, v := range tpt.metrics.MinExecutionTime {
		copy.MinExecutionTime[k] = v
	}
	for k, v := range tpt.metrics.ActiveExecutions {
		copy.ActiveExecutions[k] = v
	}
	for k, v := range tpt.metrics.UserExecutionCount {
		copy.UserExecutionCount[k] = v
	}
	for k, v := range tpt.metrics.LastExecutionTime {
		copy.LastExecutionTime[k] = v
	}

	// Copy alerts
	for i, alert := range tpt.metrics.PerformanceAlerts {
		copy.PerformanceAlerts[i] = alert
	}

	return copy
}

// GetToolMetrics returns metrics for a specific tool
func (tpt *ToolPerformanceTracker) GetToolMetrics(toolName string) map[string]interface{} {
	tpt.metrics.mu.RLock()
	defer tpt.metrics.mu.RUnlock()

	metrics := map[string]interface{}{
		"tool_name":              toolName,
		"execution_count":        tpt.metrics.ExecutionCount[toolName],
		"success_count":          tpt.metrics.SuccessCount[toolName],
		"error_count":            tpt.metrics.ErrorCount[toolName],
		"timeout_count":          tpt.metrics.TimeoutCount[toolName],
		"average_execution_time": tpt.metrics.AverageExecutionTime[toolName],
		"max_execution_time":     tpt.metrics.MaxExecutionTime[toolName],
		"min_execution_time":     tpt.metrics.MinExecutionTime[toolName],
		"active_executions":      tpt.metrics.ActiveExecutions[toolName],
		"last_execution_time":    tpt.metrics.LastExecutionTime[toolName],
	}

	// Calculate success rate
	if tpt.metrics.ExecutionCount[toolName] > 0 {
		successRate := float64(tpt.metrics.SuccessCount[toolName]) / float64(tpt.metrics.ExecutionCount[toolName]) * 100
		metrics["success_rate_percent"] = successRate
	}

	// Calculate error rate
	if tpt.metrics.ExecutionCount[toolName] > 0 {
		errorRate := float64(tpt.metrics.ErrorCount[toolName]) / float64(tpt.metrics.ExecutionCount[toolName]) * 100
		metrics["error_rate_percent"] = errorRate
	}

	// Calculate timeout rate
	if tpt.metrics.ExecutionCount[toolName] > 0 {
		timeoutRate := float64(tpt.metrics.TimeoutCount[toolName]) / float64(tpt.metrics.ExecutionCount[toolName]) * 100
		metrics["timeout_rate_percent"] = timeoutRate
	}

	return metrics
}

// GetRecentAlerts returns recent performance alerts
func (tpt *ToolPerformanceTracker) GetRecentAlerts(hours int) []PerformanceAlert {
	tpt.metrics.mu.RLock()
	defer tpt.metrics.mu.RUnlock()

	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	var recentAlerts []PerformanceAlert

	for _, alert := range tpt.metrics.PerformanceAlerts {
		if alert.Timestamp.After(cutoff) {
			recentAlerts = append(recentAlerts, alert)
		}
	}

	return recentAlerts
}

// Close closes the performance tracker and stops background goroutines
func (tpt *ToolPerformanceTracker) Close() {
	close(tpt.alertChan)
}