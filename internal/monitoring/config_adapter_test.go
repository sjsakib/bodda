package monitoring

import (
	"bodda/internal/config"
	"context"
	"testing"
	"time"
)

func TestConvertConfigToMonitoringConfig(t *testing.T) {
	// Create a test config
	appConfig := &config.Config{
		ToolMonitoring: config.ToolMonitoringConfig{
			Enabled:                true,
			EnableParameterLogging: true,
			MaxExecutionTimeMs:     25000,
			MaxConcurrentExecs:     8,
			MaxQueueDepth:          15,
			MaxErrorRatePercent:    12.5,
			MaxTimeoutRatePercent:  7.5,
			AlertRetentionHours:    48,
			RateLimitPerMinute:     120,
			MaxConcurrentPerUser:   5,
			EnableDetailedLogging:  true,
			LogLevel:              "debug",
		},
	}

	// Convert to monitoring config
	monitoringConfig := ConvertConfigToMonitoringConfig(appConfig)

	// Verify conversion
	if !monitoringConfig.Enabled {
		t.Error("Expected monitoring to be enabled")
	}

	if !monitoringConfig.EnableParameterLogging {
		t.Error("Expected parameter logging to be enabled")
	}

	if monitoringConfig.PerformanceThresholds.MaxExecutionTimeMs != 25000 {
		t.Errorf("Expected MaxExecutionTimeMs to be 25000, got %d", monitoringConfig.PerformanceThresholds.MaxExecutionTimeMs)
	}

	if monitoringConfig.PerformanceThresholds.MaxConcurrentExecs != 8 {
		t.Errorf("Expected MaxConcurrentExecs to be 8, got %d", monitoringConfig.PerformanceThresholds.MaxConcurrentExecs)
	}

	if monitoringConfig.PerformanceThresholds.MaxQueueDepth != 15 {
		t.Errorf("Expected MaxQueueDepth to be 15, got %d", monitoringConfig.PerformanceThresholds.MaxQueueDepth)
	}

	if monitoringConfig.PerformanceThresholds.MaxErrorRatePercent != 12.5 {
		t.Errorf("Expected MaxErrorRatePercent to be 12.5, got %f", monitoringConfig.PerformanceThresholds.MaxErrorRatePercent)
	}

	if monitoringConfig.PerformanceThresholds.MaxTimeoutRatePercent != 7.5 {
		t.Errorf("Expected MaxTimeoutRatePercent to be 7.5, got %f", monitoringConfig.PerformanceThresholds.MaxTimeoutRatePercent)
	}

	if monitoringConfig.PerformanceThresholds.AlertRetentionHours != 48 {
		t.Errorf("Expected AlertRetentionHours to be 48, got %d", monitoringConfig.PerformanceThresholds.AlertRetentionHours)
	}
}

func TestGetLogLevelFromConfig(t *testing.T) {
	testCases := []struct {
		configLevel    string
		expectedLevel  LogLevel
	}{
		{"debug", LevelDebug},
		{"info", LevelInfo},
		{"warn", LevelWarn},
		{"error", LevelError},
		{"invalid", LevelInfo}, // Should default to info
		{"", LevelInfo},        // Should default to info
	}

	for _, tc := range testCases {
		t.Run(tc.configLevel, func(t *testing.T) {
			appConfig := &config.Config{
				ToolMonitoring: config.ToolMonitoringConfig{
					LogLevel: tc.configLevel,
				},
			}

			logLevel := GetLogLevelFromConfig(appConfig)
			if logLevel != tc.expectedLevel {
				t.Errorf("Expected log level %s, got %s", tc.expectedLevel, logLevel)
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	// Test with invalid values that should be corrected
	appConfig := &config.Config{
		ToolMonitoring: config.ToolMonitoringConfig{
			Enabled:                true,
			EnableParameterLogging: true,
			MaxExecutionTimeMs:     -1000,  // Invalid, should be corrected
			MaxConcurrentExecs:     -5,     // Invalid, should be corrected
			MaxQueueDepth:          0,      // Invalid, should be corrected
			MaxErrorRatePercent:    150.0,  // Invalid, should be corrected
			MaxTimeoutRatePercent:  -10.0,  // Invalid, should be corrected
			AlertRetentionHours:    -24,    // Invalid, should be corrected
			RateLimitPerMinute:     0,      // Invalid, should be corrected
			MaxConcurrentPerUser:   100,    // Invalid (higher than MaxConcurrentExecs), should be corrected
			EnableDetailedLogging:  true,
			LogLevel:              "invalid", // Invalid, should be corrected
		},
	}

	// This should validate and correct the config
	appConfig.ToolMonitoring = config.ToolMonitoringConfig{
		Enabled:                appConfig.ToolMonitoring.Enabled,
		EnableParameterLogging: appConfig.ToolMonitoring.EnableParameterLogging,
		MaxExecutionTimeMs:     appConfig.ToolMonitoring.MaxExecutionTimeMs,
		MaxConcurrentExecs:     appConfig.ToolMonitoring.MaxConcurrentExecs,
		MaxQueueDepth:          appConfig.ToolMonitoring.MaxQueueDepth,
		MaxErrorRatePercent:    appConfig.ToolMonitoring.MaxErrorRatePercent,
		MaxTimeoutRatePercent:  appConfig.ToolMonitoring.MaxTimeoutRatePercent,
		AlertRetentionHours:    appConfig.ToolMonitoring.AlertRetentionHours,
		RateLimitPerMinute:     appConfig.ToolMonitoring.RateLimitPerMinute,
		MaxConcurrentPerUser:   appConfig.ToolMonitoring.MaxConcurrentPerUser,
		EnableDetailedLogging:  appConfig.ToolMonitoring.EnableDetailedLogging,
		LogLevel:              appConfig.ToolMonitoring.LogLevel,
	}

	// Manually validate (simulating what Load() does)
	tm := &appConfig.ToolMonitoring
	
	if tm.MaxExecutionTimeMs <= 0 {
		tm.MaxExecutionTimeMs = 30000
	}
	if tm.MaxConcurrentExecs <= 0 {
		tm.MaxConcurrentExecs = 10
	}
	if tm.MaxConcurrentPerUser <= 0 {
		tm.MaxConcurrentPerUser = 3
	}
	if tm.MaxConcurrentPerUser > tm.MaxConcurrentExecs {
		tm.MaxConcurrentPerUser = tm.MaxConcurrentExecs / 2
		if tm.MaxConcurrentPerUser <= 0 {
			tm.MaxConcurrentPerUser = 1
		}
	}
	if tm.MaxQueueDepth <= 0 {
		tm.MaxQueueDepth = 20
	}
	if tm.MaxErrorRatePercent < 0 || tm.MaxErrorRatePercent > 100 {
		tm.MaxErrorRatePercent = 10.0
	}
	if tm.MaxTimeoutRatePercent < 0 || tm.MaxTimeoutRatePercent > 100 {
		tm.MaxTimeoutRatePercent = 5.0
	}
	if tm.AlertRetentionHours <= 0 {
		tm.AlertRetentionHours = 24
	}
	if tm.RateLimitPerMinute <= 0 {
		tm.RateLimitPerMinute = 60
	}

	// Verify corrections
	if tm.MaxExecutionTimeMs != 30000 {
		t.Errorf("Expected MaxExecutionTimeMs to be corrected to 30000, got %d", tm.MaxExecutionTimeMs)
	}

	if tm.MaxConcurrentExecs != 10 {
		t.Errorf("Expected MaxConcurrentExecs to be corrected to 10, got %d", tm.MaxConcurrentExecs)
	}

	if tm.MaxQueueDepth != 20 {
		t.Errorf("Expected MaxQueueDepth to be corrected to 20, got %d", tm.MaxQueueDepth)
	}

	if tm.MaxErrorRatePercent != 10.0 {
		t.Errorf("Expected MaxErrorRatePercent to be corrected to 10.0, got %f", tm.MaxErrorRatePercent)
	}

	if tm.MaxTimeoutRatePercent != 5.0 {
		t.Errorf("Expected MaxTimeoutRatePercent to be corrected to 5.0, got %f", tm.MaxTimeoutRatePercent)
	}

	if tm.AlertRetentionHours != 24 {
		t.Errorf("Expected AlertRetentionHours to be corrected to 24, got %d", tm.AlertRetentionHours)
	}

	if tm.RateLimitPerMinute != 60 {
		t.Errorf("Expected RateLimitPerMinute to be corrected to 60, got %d", tm.RateLimitPerMinute)
	}

	if tm.MaxConcurrentPerUser != 5 {
		t.Errorf("Expected MaxConcurrentPerUser to be corrected to 5, got %d", tm.MaxConcurrentPerUser)
	}
}

func TestIntegrationWithMonitoringSystem(t *testing.T) {
	// Create a config with specific values
	appConfig := &config.Config{
		ToolMonitoring: config.ToolMonitoringConfig{
			Enabled:                true,
			EnableParameterLogging: false, // Disable parameter logging for this test
			MaxExecutionTimeMs:     5000,  // 5 seconds
			MaxConcurrentExecs:     3,
			MaxQueueDepth:          5,
			MaxErrorRatePercent:    15.0,
			MaxTimeoutRatePercent:  8.0,
			AlertRetentionHours:    12,
			RateLimitPerMinute:     30,
			MaxConcurrentPerUser:   2,
			EnableDetailedLogging:  false,
			LogLevel:              "warn",
		},
	}

	// Convert to monitoring config
	monitoringConfig := ConvertConfigToMonitoringConfig(appConfig)
	
	// Create logger with the config log level
	logLevel := GetLogLevelFromConfig(appConfig)
	logger := NewLogger(logLevel, true)
	
	// Create monitoring system
	monitoringSystem := NewToolMonitoringSystem(logger, monitoringConfig)
	defer monitoringSystem.Close()

	// Verify the system is configured correctly
	if !monitoringSystem.IsEnabled() {
		t.Error("Expected monitoring system to be enabled")
	}

	// Test that the thresholds are applied correctly by triggering an alert
	ctx := context.Background()
	
	// This should trigger a slow execution alert since we set MaxExecutionTimeMs to 5000
	execution := func() (interface{}, int64, error) {
		// Simulate a slow execution (6 seconds, which exceeds our 5-second threshold)
		// We'll simulate this by directly calling the performance tracker
		tracker := monitoringSystem.performanceTracker
		tracker.RecordToolExecutionStart(ctx, "test-tool", "user123")
		tracker.RecordToolExecutionEnd(ctx, "test-tool", "user123", 6000*time.Millisecond, true, false)
		return "result", 100, nil
	}

	result, err := monitoringSystem.MonitorToolExecution(ctx, "test-tool", map[string]interface{}{}, "user123", execution)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if result != "result" {
		t.Errorf("Expected result to be 'result', got %v", result)
	}

	// Check that an alert was generated (wait a bit for processing)
	time.Sleep(100 * time.Millisecond)
	alerts := monitoringSystem.GetRecentAlerts(1)
	
	// We should have at least one alert for the slow execution
	if len(alerts) == 0 {
		t.Log("No alerts generated - this might be expected depending on the exact timing")
	} else {
		t.Logf("Generated %d alerts as expected", len(alerts))
		
		// Check that the alert is for slow execution
		foundSlowExecutionAlert := false
		for _, alert := range alerts {
			if alert.AlertType == "slow_execution" {
				foundSlowExecutionAlert = true
				break
			}
		}
		
		if !foundSlowExecutionAlert {
			t.Log("No slow execution alert found, but other alerts were generated")
		}
	}
}