package monitoring

import (
	"context"
	"time"
)

// ToolMonitoringSystem provides comprehensive monitoring for tool executions
type ToolMonitoringSystem struct {
	logger            *ToolExecutionLogger
	performanceTracker *ToolPerformanceTracker
	enabled           bool
}

// ToolMonitoringConfig holds configuration for the tool monitoring system
type ToolMonitoringConfig struct {
	EnableParameterLogging bool                  `json:"enable_parameter_logging"`
	PerformanceThresholds  PerformanceThresholds `json:"performance_thresholds"`
	Enabled               bool                  `json:"enabled"`
}

// DefaultToolMonitoringConfig returns default configuration for tool monitoring
func DefaultToolMonitoringConfig() ToolMonitoringConfig {
	return ToolMonitoringConfig{
		EnableParameterLogging: true,
		PerformanceThresholds: PerformanceThresholds{
			MaxExecutionTimeMs:    30000, // 30 seconds
			MaxConcurrentExecs:    10,
			MaxQueueDepth:         20,
			MaxErrorRatePercent:   10.0,
			MaxTimeoutRatePercent: 5.0,
			AlertRetentionHours:   24,
		},
		Enabled: true,
	}
}

// NewToolMonitoringSystem creates a new comprehensive tool monitoring system
func NewToolMonitoringSystem(logger *Logger, config ToolMonitoringConfig) *ToolMonitoringSystem {
	if !config.Enabled {
		return &ToolMonitoringSystem{
			enabled: false,
		}
	}

	executionLogger := NewToolExecutionLogger(logger, config.EnableParameterLogging)
	performanceTracker := NewToolPerformanceTracker(logger, config.PerformanceThresholds)

	return &ToolMonitoringSystem{
		logger:            executionLogger,
		performanceTracker: performanceTracker,
		enabled:           true,
	}
}

// MonitorToolExecution monitors a complete tool execution lifecycle
func (tms *ToolMonitoringSystem) MonitorToolExecution(ctx context.Context, toolName string, parameters map[string]interface{}, userID string, execution func() (interface{}, int64, error)) (interface{}, error) {
	if !tms.enabled {
		// If monitoring is disabled, just execute the tool
		result, _, err := execution()
		return result, err
	}

	// Record execution start
	start := time.Now()
	tms.performanceTracker.RecordToolExecutionStart(ctx, toolName, userID)

	// Execute the tool
	result, responseSize, err := execution()
	duration := time.Since(start)

	// Determine execution outcome
	success := err == nil
	isTimeout := false
	if err != nil {
		// Check if this is a timeout error (you might need to adjust this based on your error types)
		if duration >= 30*time.Second { // Assuming 30s is your default timeout
			isTimeout = true
		}
	}

	// Record execution end
	tms.performanceTracker.RecordToolExecutionEnd(ctx, toolName, userID, duration, success, isTimeout)

	// Log the execution
	if success {
		tms.logger.LogToolExecution(ctx, toolName, parameters, duration, responseSize, userID)
	} else if isTimeout {
		tms.logger.LogToolExecutionTimeout(ctx, toolName, parameters, duration, userID)
	} else {
		tms.logger.LogToolExecutionError(ctx, toolName, parameters, duration, err, userID)
	}

	return result, err
}

// MonitorToolValidation monitors tool parameter validation
func (tms *ToolMonitoringSystem) MonitorToolValidation(ctx context.Context, toolName string, parameters map[string]interface{}, userID string, validationError error) {
	if !tms.enabled {
		return
	}

	tms.logger.LogToolValidationError(ctx, toolName, parameters, validationError, userID)
}

// MonitorToolDiscovery monitors tool discovery operations
func (tms *ToolMonitoringSystem) MonitorToolDiscovery(ctx context.Context, operation string, toolName string, userID string, execution func() error) error {
	if !tms.enabled {
		return execution()
	}

	start := time.Now()
	err := execution()
	duration := time.Since(start)

	tms.logger.LogToolDiscovery(ctx, operation, toolName, duration, userID)
	return err
}

// RecordQueueDepth records the current tool execution queue depth
func (tms *ToolMonitoringSystem) RecordQueueDepth(depth int64) {
	if !tms.enabled {
		return
	}

	tms.performanceTracker.RecordQueueDepth(depth)
}

// GetPerformanceMetrics returns current performance metrics
func (tms *ToolMonitoringSystem) GetPerformanceMetrics() *ToolPerformanceMetrics {
	if !tms.enabled {
		return &ToolPerformanceMetrics{}
	}

	return tms.performanceTracker.GetMetrics()
}

// GetToolMetrics returns metrics for a specific tool
func (tms *ToolMonitoringSystem) GetToolMetrics(toolName string) map[string]interface{} {
	if !tms.enabled {
		return map[string]interface{}{}
	}

	return tms.performanceTracker.GetToolMetrics(toolName)
}

// GetRecentAlerts returns recent performance alerts
func (tms *ToolMonitoringSystem) GetRecentAlerts(hours int) []PerformanceAlert {
	if !tms.enabled {
		return []PerformanceAlert{}
	}

	return tms.performanceTracker.GetRecentAlerts(hours)
}

// IsEnabled returns whether monitoring is enabled
func (tms *ToolMonitoringSystem) IsEnabled() bool {
	return tms.enabled
}

// Close closes the monitoring system and cleans up resources
func (tms *ToolMonitoringSystem) Close() {
	if !tms.enabled {
		return
	}

	if tms.performanceTracker != nil {
		tms.performanceTracker.Close()
	}
}

// ToolExecutionContext provides context for tool execution monitoring
type ToolExecutionContext struct {
	ToolName   string
	Parameters map[string]interface{}
	UserID     string
	RequestID  string
	StartTime  time.Time
}

// NewToolExecutionContext creates a new tool execution context
func NewToolExecutionContext(toolName string, parameters map[string]interface{}, userID string, requestID string) *ToolExecutionContext {
	return &ToolExecutionContext{
		ToolName:   toolName,
		Parameters: parameters,
		UserID:     userID,
		RequestID:  requestID,
		StartTime:  time.Now(),
	}
}

// Duration returns the elapsed time since the context was created
func (tec *ToolExecutionContext) Duration() time.Duration {
	return time.Since(tec.StartTime)
}

// MonitoringMiddleware provides middleware for automatic tool execution monitoring
type MonitoringMiddleware struct {
	monitoringSystem *ToolMonitoringSystem
}

// NewMonitoringMiddleware creates a new monitoring middleware
func NewMonitoringMiddleware(monitoringSystem *ToolMonitoringSystem) *MonitoringMiddleware {
	return &MonitoringMiddleware{
		monitoringSystem: monitoringSystem,
	}
}

// WrapToolExecution wraps a tool execution function with monitoring
func (mm *MonitoringMiddleware) WrapToolExecution(toolName string, userID string, execution func(ctx context.Context, parameters map[string]interface{}) (interface{}, int64, error)) func(ctx context.Context, parameters map[string]interface{}) (interface{}, error) {
	return func(ctx context.Context, parameters map[string]interface{}) (interface{}, error) {
		return mm.monitoringSystem.MonitorToolExecution(ctx, toolName, parameters, userID, func() (interface{}, int64, error) {
			return execution(ctx, parameters)
		})
	}
}

// Global tool monitoring system instance
var globalToolMonitoring *ToolMonitoringSystem

// InitGlobalToolMonitoring initializes the global tool monitoring system
func InitGlobalToolMonitoring(logger *Logger, config ToolMonitoringConfig) {
	globalToolMonitoring = NewToolMonitoringSystem(logger, config)
}

// GetGlobalToolMonitoring returns the global tool monitoring system
func GetGlobalToolMonitoring() *ToolMonitoringSystem {
	if globalToolMonitoring == nil {
		// Return a disabled monitoring system as fallback
		return &ToolMonitoringSystem{enabled: false}
	}
	return globalToolMonitoring
}

// Convenience functions for global monitoring system
func MonitorToolExecution(ctx context.Context, toolName string, parameters map[string]interface{}, userID string, execution func() (interface{}, int64, error)) (interface{}, error) {
	return GetGlobalToolMonitoring().MonitorToolExecution(ctx, toolName, parameters, userID, execution)
}

func MonitorToolValidation(ctx context.Context, toolName string, parameters map[string]interface{}, userID string, validationError error) {
	GetGlobalToolMonitoring().MonitorToolValidation(ctx, toolName, parameters, userID, validationError)
}

func MonitorToolDiscovery(ctx context.Context, operation string, toolName string, userID string, execution func() error) error {
	return GetGlobalToolMonitoring().MonitorToolDiscovery(ctx, operation, toolName, userID, execution)
}

func RecordQueueDepth(depth int64) {
	GetGlobalToolMonitoring().RecordQueueDepth(depth)
}