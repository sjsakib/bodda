package monitoring

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestToolExecutionLogger(t *testing.T) {
	logger := NewLogger(LevelInfo, true)
	execLogger := NewToolExecutionLogger(logger, true)

	ctx := context.Background()
	toolName := "test-tool"
	parameters := map[string]interface{}{
		"param1": "value1",
		"password": "secret123", // Should be redacted
		"api_key": "key123",     // Should be redacted
	}
	userID := "user123"

	t.Run("LogToolExecution", func(t *testing.T) {
		duration := 100 * time.Millisecond
		responseSize := int64(1024)

		// This should not panic and should log successfully
		execLogger.LogToolExecution(ctx, toolName, parameters, duration, responseSize, userID)
	})

	t.Run("LogToolExecutionError", func(t *testing.T) {
		duration := 200 * time.Millisecond
		err := errors.New("test error")

		// This should not panic and should log the error
		execLogger.LogToolExecutionError(ctx, toolName, parameters, duration, err, userID)
	})

	t.Run("LogToolExecutionTimeout", func(t *testing.T) {
		timeout := 30 * time.Second

		// This should not panic and should log the timeout
		execLogger.LogToolExecutionTimeout(ctx, toolName, parameters, timeout, userID)
	})

	t.Run("LogToolValidationError", func(t *testing.T) {
		validationErr := errors.New("validation failed")

		// This should not panic and should log the validation error
		execLogger.LogToolValidationError(ctx, toolName, parameters, validationErr, userID)
	})

	t.Run("ParameterSanitization", func(t *testing.T) {
		sanitized := execLogger.sanitizeParameters(parameters)
		
		if sanitized["param1"] != "value1" {
			t.Errorf("Expected param1 to be 'value1', got %v", sanitized["param1"])
		}
		
		if sanitized["password"] != "[REDACTED]" {
			t.Errorf("Expected password to be redacted, got %v", sanitized["password"])
		}
		
		if sanitized["api_key"] != "[REDACTED]" {
			t.Errorf("Expected api_key to be redacted, got %v", sanitized["api_key"])
		}
	})
}

func TestToolPerformanceTracker(t *testing.T) {
	logger := NewLogger(LevelInfo, true)
	thresholds := PerformanceThresholds{
		MaxExecutionTimeMs:    1000, // 1 second
		MaxConcurrentExecs:    5,
		MaxQueueDepth:         10,
		MaxErrorRatePercent:   20.0,
		MaxTimeoutRatePercent: 10.0,
		AlertRetentionHours:   1,
	}
	
	tracker := NewToolPerformanceTracker(logger, thresholds)
	defer tracker.Close()

	ctx := context.Background()
	toolName := "test-tool"
	userID := "user123"

	t.Run("RecordSuccessfulExecution", func(t *testing.T) {
		tracker.RecordToolExecutionStart(ctx, toolName, userID)
		duration := 500 * time.Millisecond
		tracker.RecordToolExecutionEnd(ctx, toolName, userID, duration, true, false)

		metrics := tracker.GetMetrics()
		if metrics.ExecutionCount[toolName] != 1 {
			t.Errorf("Expected execution count to be 1, got %d", metrics.ExecutionCount[toolName])
		}
		if metrics.SuccessCount[toolName] != 1 {
			t.Errorf("Expected success count to be 1, got %d", metrics.SuccessCount[toolName])
		}
		if metrics.ErrorCount[toolName] != 0 {
			t.Errorf("Expected error count to be 0, got %d", metrics.ErrorCount[toolName])
		}
	})

	t.Run("RecordFailedExecution", func(t *testing.T) {
		tracker.RecordToolExecutionStart(ctx, toolName, userID)
		duration := 300 * time.Millisecond
		tracker.RecordToolExecutionEnd(ctx, toolName, userID, duration, false, false)

		metrics := tracker.GetMetrics()
		if metrics.ExecutionCount[toolName] != 2 {
			t.Errorf("Expected execution count to be 2, got %d", metrics.ExecutionCount[toolName])
		}
		if metrics.ErrorCount[toolName] != 1 {
			t.Errorf("Expected error count to be 1, got %d", metrics.ErrorCount[toolName])
		}
	})

	t.Run("RecordTimeoutExecution", func(t *testing.T) {
		tracker.RecordToolExecutionStart(ctx, toolName, userID)
		duration := 2 * time.Second // Exceeds threshold
		tracker.RecordToolExecutionEnd(ctx, toolName, userID, duration, false, true)

		metrics := tracker.GetMetrics()
		if metrics.TimeoutCount[toolName] != 1 {
			t.Errorf("Expected timeout count to be 1, got %d", metrics.TimeoutCount[toolName])
		}
	})

	t.Run("QueueDepthTracking", func(t *testing.T) {
		tracker.RecordQueueDepth(5)
		metrics := tracker.GetMetrics()
		if metrics.QueueDepth != 5 {
			t.Errorf("Expected queue depth to be 5, got %d", metrics.QueueDepth)
		}
	})

	t.Run("GetToolMetrics", func(t *testing.T) {
		toolMetrics := tracker.GetToolMetrics(toolName)
		
		if toolMetrics["tool_name"] != toolName {
			t.Errorf("Expected tool name to be %s, got %v", toolName, toolMetrics["tool_name"])
		}
		
		if toolMetrics["execution_count"].(int64) != 3 {
			t.Errorf("Expected execution count to be 3, got %v", toolMetrics["execution_count"])
		}
	})

	t.Run("AlertGeneration", func(t *testing.T) {
		// Wait a bit for alerts to be processed
		time.Sleep(100 * time.Millisecond)
		
		alerts := tracker.GetRecentAlerts(1)
		if len(alerts) == 0 {
			t.Log("No alerts generated (this might be expected depending on thresholds)")
		} else {
			t.Logf("Generated %d alerts", len(alerts))
		}
	})
}

func TestToolMonitoringSystem(t *testing.T) {
	logger := NewLogger(LevelInfo, true)
	config := DefaultToolMonitoringConfig()
	
	monitoringSystem := NewToolMonitoringSystem(logger, config)
	defer monitoringSystem.Close()

	ctx := context.Background()
	toolName := "test-tool"
	parameters := map[string]interface{}{
		"param1": "value1",
	}
	userID := "user123"

	t.Run("MonitorSuccessfulExecution", func(t *testing.T) {
		execution := func() (interface{}, int64, error) {
			time.Sleep(50 * time.Millisecond) // Simulate work
			return "success result", 100, nil
		}

		result, err := monitoringSystem.MonitorToolExecution(ctx, toolName, parameters, userID, execution)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if result != "success result" {
			t.Errorf("Expected result to be 'success result', got %v", result)
		}

		// Check metrics
		metrics := monitoringSystem.GetPerformanceMetrics()
		if metrics.ExecutionCount[toolName] != 1 {
			t.Errorf("Expected execution count to be 1, got %d", metrics.ExecutionCount[toolName])
		}
	})

	t.Run("MonitorFailedExecution", func(t *testing.T) {
		execution := func() (interface{}, int64, error) {
			time.Sleep(30 * time.Millisecond)
			return nil, 0, errors.New("execution failed")
		}

		result, err := monitoringSystem.MonitorToolExecution(ctx, toolName, parameters, userID, execution)
		
		if err == nil {
			t.Error("Expected error, got nil")
		}
		
		if result != nil {
			t.Errorf("Expected result to be nil, got %v", result)
		}

		// Check metrics
		metrics := monitoringSystem.GetPerformanceMetrics()
		if metrics.ErrorCount[toolName] != 1 {
			t.Errorf("Expected error count to be 1, got %d", metrics.ErrorCount[toolName])
		}
	})

	t.Run("MonitorToolValidation", func(t *testing.T) {
		validationError := errors.New("invalid parameters")
		
		// This should not panic
		monitoringSystem.MonitorToolValidation(ctx, toolName, parameters, userID, validationError)
	})

	t.Run("MonitorToolDiscovery", func(t *testing.T) {
		execution := func() error {
			time.Sleep(10 * time.Millisecond)
			return nil
		}

		err := monitoringSystem.MonitorToolDiscovery(ctx, "list_tools", "", userID, execution)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("GetToolMetrics", func(t *testing.T) {
		toolMetrics := monitoringSystem.GetToolMetrics(toolName)
		
		if len(toolMetrics) == 0 {
			t.Error("Expected tool metrics to be populated")
		}
	})

	t.Run("RecordQueueDepth", func(t *testing.T) {
		monitoringSystem.RecordQueueDepth(3)
		
		metrics := monitoringSystem.GetPerformanceMetrics()
		if metrics.QueueDepth != 3 {
			t.Errorf("Expected queue depth to be 3, got %d", metrics.QueueDepth)
		}
	})
}

func TestDisabledMonitoring(t *testing.T) {
	logger := NewLogger(LevelInfo, true)
	config := DefaultToolMonitoringConfig()
	config.Enabled = false
	
	monitoringSystem := NewToolMonitoringSystem(logger, config)
	defer monitoringSystem.Close()

	if monitoringSystem.IsEnabled() {
		t.Error("Expected monitoring to be disabled")
	}

	ctx := context.Background()
	toolName := "test-tool"
	parameters := map[string]interface{}{"param1": "value1"}
	userID := "user123"

	t.Run("DisabledMonitoringExecution", func(t *testing.T) {
		execution := func() (interface{}, int64, error) {
			return "result", 100, nil
		}

		result, err := monitoringSystem.MonitorToolExecution(ctx, toolName, parameters, userID, execution)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if result != "result" {
			t.Errorf("Expected result to be 'result', got %v", result)
		}

		// Metrics should be empty
		metrics := monitoringSystem.GetPerformanceMetrics()
		if len(metrics.ExecutionCount) != 0 {
			t.Error("Expected empty metrics when monitoring is disabled")
		}
	})
}

func TestMonitoringMiddleware(t *testing.T) {
	logger := NewLogger(LevelInfo, true)
	config := DefaultToolMonitoringConfig()
	
	monitoringSystem := NewToolMonitoringSystem(logger, config)
	defer monitoringSystem.Close()

	middleware := NewMonitoringMiddleware(monitoringSystem)

	ctx := context.Background()
	toolName := "test-tool"
	userID := "user123"
	parameters := map[string]interface{}{"param1": "value1"}

	t.Run("WrapToolExecution", func(t *testing.T) {
		originalExecution := func(ctx context.Context, parameters map[string]interface{}) (interface{}, int64, error) {
			time.Sleep(25 * time.Millisecond)
			return "wrapped result", 200, nil
		}

		wrappedExecution := middleware.WrapToolExecution(toolName, userID, originalExecution)
		
		result, err := wrappedExecution(ctx, parameters)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if result != "wrapped result" {
			t.Errorf("Expected result to be 'wrapped result', got %v", result)
		}

		// Check that metrics were recorded
		metrics := monitoringSystem.GetPerformanceMetrics()
		if metrics.ExecutionCount[toolName] != 1 {
			t.Errorf("Expected execution count to be 1, got %d", metrics.ExecutionCount[toolName])
		}
	})
}

func TestToolExecutionContext(t *testing.T) {
	toolName := "test-tool"
	parameters := map[string]interface{}{"param1": "value1"}
	userID := "user123"
	requestID := "req123"

	ctx := NewToolExecutionContext(toolName, parameters, userID, requestID)

	if ctx.ToolName != toolName {
		t.Errorf("Expected tool name to be %s, got %s", toolName, ctx.ToolName)
	}

	if ctx.UserID != userID {
		t.Errorf("Expected user ID to be %s, got %s", userID, ctx.UserID)
	}

	if ctx.RequestID != requestID {
		t.Errorf("Expected request ID to be %s, got %s", requestID, ctx.RequestID)
	}

	// Test duration
	time.Sleep(10 * time.Millisecond)
	duration := ctx.Duration()
	if duration < 10*time.Millisecond {
		t.Errorf("Expected duration to be at least 10ms, got %v", duration)
	}
}

func TestGlobalMonitoringSystem(t *testing.T) {
	logger := NewLogger(LevelInfo, true)
	config := DefaultToolMonitoringConfig()

	// Initialize global monitoring
	InitGlobalToolMonitoring(logger, config)

	// Test global functions
	ctx := context.Background()
	toolName := "global-test-tool"
	parameters := map[string]interface{}{"param1": "value1"}
	userID := "user123"

	t.Run("GlobalMonitorToolExecution", func(t *testing.T) {
		execution := func() (interface{}, int64, error) {
			return "global result", 150, nil
		}

		result, err := MonitorToolExecution(ctx, toolName, parameters, userID, execution)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if result != "global result" {
			t.Errorf("Expected result to be 'global result', got %v", result)
		}
	})

	t.Run("GlobalMonitorToolValidation", func(t *testing.T) {
		validationError := errors.New("global validation error")
		
		// This should not panic
		MonitorToolValidation(ctx, toolName, parameters, userID, validationError)
	})

	t.Run("GlobalMonitorToolDiscovery", func(t *testing.T) {
		execution := func() error {
			return nil
		}

		err := MonitorToolDiscovery(ctx, "global_list_tools", "", userID, execution)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("GlobalRecordQueueDepth", func(t *testing.T) {
		// This should not panic
		RecordQueueDepth(7)
	})
}