package config

import (
	"os"
	"testing"
)

func TestToolExecutionConfigDefaults(t *testing.T) {
	// Clear environment variables to test defaults
	clearToolExecutionEnvVars()
	
	config := Load()
	
	// Test default timeout values
	if config.ToolExecution.DefaultTimeout != 30 {
		t.Errorf("Expected DefaultTimeout to be 30, got %d", config.ToolExecution.DefaultTimeout)
	}
	
	if config.ToolExecution.MaxTimeout != 300 {
		t.Errorf("Expected MaxTimeout to be 300, got %d", config.ToolExecution.MaxTimeout)
	}
	
	// Test default rate limiting
	if config.ToolExecution.RateLimitPerMinute != 60 {
		t.Errorf("Expected RateLimitPerMinute to be 60, got %d", config.ToolExecution.RateLimitPerMinute)
	}
	
	if config.ToolExecution.MaxConcurrentExecs != 5 {
		t.Errorf("Expected MaxConcurrentExecs to be 5, got %d", config.ToolExecution.MaxConcurrentExecs)
	}
	
	// Test default caching
	if config.ToolExecution.EnableCaching != false {
		t.Errorf("Expected EnableCaching to be false, got %v", config.ToolExecution.EnableCaching)
	}
	
	if config.ToolExecution.CacheTTL != 300 {
		t.Errorf("Expected CacheTTL to be 300, got %d", config.ToolExecution.CacheTTL)
	}
	
	// Test default logging
	if config.ToolExecution.EnableDetailedLogging != true {
		t.Errorf("Expected EnableDetailedLogging to be true, got %v", config.ToolExecution.EnableDetailedLogging)
	}
	
	// Test default performance thresholds
	pt := config.ToolExecution.PerformanceThresholds
	
	if pt.MaxExecutionTimeMs != 5000 {
		t.Errorf("Expected MaxExecutionTimeMs to be 5000, got %d", pt.MaxExecutionTimeMs)
	}
	
	if pt.MaxCPUUsagePercent != 80.0 {
		t.Errorf("Expected MaxCPUUsagePercent to be 80.0, got %f", pt.MaxCPUUsagePercent)
	}
	
	if pt.MaxMemoryUsageMB != 512 {
		t.Errorf("Expected MaxMemoryUsageMB to be 512, got %d", pt.MaxMemoryUsageMB)
	}
	
	if pt.MaxQueueDepth != 10 {
		t.Errorf("Expected MaxQueueDepth to be 10, got %d", pt.MaxQueueDepth)
	}
}

func TestToolExecutionConfigFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("TOOL_EXECUTION_TIMEOUT", "45")
	os.Setenv("TOOL_MAX_TIMEOUT", "600")
	os.Setenv("TOOL_EXECUTION_RATE_LIMIT_PER_MINUTE", "120")
	os.Setenv("TOOL_EXECUTION_MAX_CONCURRENT_EXECUTIONS", "8")
	os.Setenv("TOOL_EXECUTION_ENABLE_CACHING", "true")
	os.Setenv("TOOL_EXECUTION_CACHE_TTL", "600")
	os.Setenv("TOOL_EXECUTION_ENABLE_DETAILED_LOGGING", "false")
	os.Setenv("TOOL_EXECUTION_PERFORMANCE_ALERT_THRESHOLD_MS", "3000")
	os.Setenv("TOOL_EXECUTION_MAX_CPU_USAGE_PERCENT", "90.0")
	os.Setenv("TOOL_EXECUTION_MAX_MEMORY_USAGE_MB", "1024")
	os.Setenv("TOOL_EXECUTION_MAX_QUEUE_DEPTH", "20")
	
	defer clearToolExecutionEnvVars()
	
	config := Load()
	
	// Test environment variable values
	if config.ToolExecution.DefaultTimeout != 45 {
		t.Errorf("Expected DefaultTimeout to be 45, got %d", config.ToolExecution.DefaultTimeout)
	}
	
	if config.ToolExecution.MaxTimeout != 600 {
		t.Errorf("Expected MaxTimeout to be 600, got %d", config.ToolExecution.MaxTimeout)
	}
	
	if config.ToolExecution.RateLimitPerMinute != 120 {
		t.Errorf("Expected RateLimitPerMinute to be 120, got %d", config.ToolExecution.RateLimitPerMinute)
	}
	
	if config.ToolExecution.MaxConcurrentExecs != 8 {
		t.Errorf("Expected MaxConcurrentExecs to be 8, got %d", config.ToolExecution.MaxConcurrentExecs)
	}
	
	if config.ToolExecution.EnableCaching != true {
		t.Errorf("Expected EnableCaching to be true, got %v", config.ToolExecution.EnableCaching)
	}
	
	if config.ToolExecution.CacheTTL != 600 {
		t.Errorf("Expected CacheTTL to be 600, got %d", config.ToolExecution.CacheTTL)
	}
	
	if config.ToolExecution.EnableDetailedLogging != false {
		t.Errorf("Expected EnableDetailedLogging to be false, got %v", config.ToolExecution.EnableDetailedLogging)
	}
	
	// Test performance thresholds from environment
	pt := config.ToolExecution.PerformanceThresholds
	
	if pt.MaxExecutionTimeMs != 3000 {
		t.Errorf("Expected MaxExecutionTimeMs to be 3000, got %d", pt.MaxExecutionTimeMs)
	}
	
	if pt.MaxCPUUsagePercent != 90.0 {
		t.Errorf("Expected MaxCPUUsagePercent to be 90.0, got %f", pt.MaxCPUUsagePercent)
	}
	
	if pt.MaxMemoryUsageMB != 1024 {
		t.Errorf("Expected MaxMemoryUsageMB to be 1024, got %d", pt.MaxMemoryUsageMB)
	}
	
	if pt.MaxQueueDepth != 20 {
		t.Errorf("Expected MaxQueueDepth to be 20, got %d", pt.MaxQueueDepth)
	}
}

func TestToolExecutionConfigValidation(t *testing.T) {
	// Test validation with invalid values
	os.Setenv("TOOL_EXECUTION_TIMEOUT", "-10")
	os.Setenv("TOOL_MAX_TIMEOUT", "5") // Less than default timeout
	os.Setenv("TOOL_EXECUTION_RATE_LIMIT_PER_MINUTE", "-5")
	os.Setenv("TOOL_EXECUTION_MAX_CONCURRENT_EXECUTIONS", "0")
	os.Setenv("TOOL_EXECUTION_CACHE_TTL", "-100")
	os.Setenv("TOOL_EXECUTION_PERFORMANCE_ALERT_THRESHOLD_MS", "-1000")
	os.Setenv("TOOL_EXECUTION_MAX_CPU_USAGE_PERCENT", "150.0")
	os.Setenv("TOOL_EXECUTION_MAX_MEMORY_USAGE_MB", "-512")
	os.Setenv("TOOL_EXECUTION_MAX_QUEUE_DEPTH", "-5")
	
	defer clearToolExecutionEnvVars()
	
	config := Load()
	
	// Test that validation corrected invalid values
	if config.ToolExecution.DefaultTimeout != 30 {
		t.Errorf("Expected DefaultTimeout to be corrected to 30, got %d", config.ToolExecution.DefaultTimeout)
	}
	
	if config.ToolExecution.MaxTimeout != 300 {
		t.Errorf("Expected MaxTimeout to be corrected to 300, got %d", config.ToolExecution.MaxTimeout)
	}
	
	if config.ToolExecution.RateLimitPerMinute != 60 {
		t.Errorf("Expected RateLimitPerMinute to be corrected to 60, got %d", config.ToolExecution.RateLimitPerMinute)
	}
	
	if config.ToolExecution.MaxConcurrentExecs != 5 {
		t.Errorf("Expected MaxConcurrentExecs to be corrected to 5, got %d", config.ToolExecution.MaxConcurrentExecs)
	}
	
	if config.ToolExecution.CacheTTL != 300 {
		t.Errorf("Expected CacheTTL to be corrected to 300, got %d", config.ToolExecution.CacheTTL)
	}
	
	// Test performance threshold validation
	pt := config.ToolExecution.PerformanceThresholds
	
	if pt.MaxExecutionTimeMs != 5000 {
		t.Errorf("Expected MaxExecutionTimeMs to be corrected to 5000, got %d", pt.MaxExecutionTimeMs)
	}
	
	if pt.MaxCPUUsagePercent != 80.0 {
		t.Errorf("Expected MaxCPUUsagePercent to be corrected to 80.0, got %f", pt.MaxCPUUsagePercent)
	}
	
	if pt.MaxMemoryUsageMB != 512 {
		t.Errorf("Expected MaxMemoryUsageMB to be corrected to 512, got %d", pt.MaxMemoryUsageMB)
	}
	
	if pt.MaxQueueDepth != 10 {
		t.Errorf("Expected MaxQueueDepth to be corrected to 10, got %d", pt.MaxQueueDepth)
	}
}

func TestDevelopmentModeConfiguration(t *testing.T) {
	// Test development mode enabled
	os.Setenv("DEVELOPMENT_MODE", "true")
	defer os.Unsetenv("DEVELOPMENT_MODE")
	
	config := Load()
	
	if !config.IsDevelopment {
		t.Errorf("Expected IsDevelopment to be true when DEVELOPMENT_MODE=true")
	}
	
	// Test development mode disabled (default)
	os.Setenv("DEVELOPMENT_MODE", "false")
	config = Load()
	
	if config.IsDevelopment {
		t.Errorf("Expected IsDevelopment to be false when DEVELOPMENT_MODE=false")
	}
}

func TestIsToolExecutionEnabled(t *testing.T) {
	// Test with development mode enabled
	os.Setenv("DEVELOPMENT_MODE", "true")
	defer os.Unsetenv("DEVELOPMENT_MODE")
	
	config := Load()
	
	if !config.IsToolExecutionEnabled() {
		t.Errorf("Expected IsToolExecutionEnabled to be true in development mode")
	}
	
	// Test with development mode disabled
	os.Setenv("DEVELOPMENT_MODE", "false")
	config = Load()
	
	if config.IsToolExecutionEnabled() {
		t.Errorf("Expected IsToolExecutionEnabled to be false in production mode")
	}
}

func TestGetToolExecutionTimeout(t *testing.T) {
	os.Setenv("TOOL_EXECUTION_TIMEOUT", "30")
	os.Setenv("TOOL_MAX_TIMEOUT", "300")
	defer clearToolExecutionEnvVars()
	
	config := Load()
	
	// Test with no requested timeout (should return default)
	timeout := config.GetToolExecutionTimeout(0)
	if timeout != 30 {
		t.Errorf("Expected timeout to be 30 (default), got %d", timeout)
	}
	
	// Test with negative requested timeout (should return default)
	timeout = config.GetToolExecutionTimeout(-10)
	if timeout != 30 {
		t.Errorf("Expected timeout to be 30 (default), got %d", timeout)
	}
	
	// Test with valid requested timeout
	timeout = config.GetToolExecutionTimeout(60)
	if timeout != 60 {
		t.Errorf("Expected timeout to be 60 (requested), got %d", timeout)
	}
	
	// Test with requested timeout exceeding maximum (should return max)
	timeout = config.GetToolExecutionTimeout(500)
	if timeout != 300 {
		t.Errorf("Expected timeout to be 300 (max), got %d", timeout)
	}
}

func clearToolExecutionEnvVars() {
	envVars := []string{
		"TOOL_EXECUTION_TIMEOUT",
		"TOOL_MAX_TIMEOUT",
		"TOOL_EXECUTION_RATE_LIMIT_PER_MINUTE",
		"TOOL_EXECUTION_MAX_CONCURRENT_EXECUTIONS",
		"TOOL_EXECUTION_ENABLE_CACHING",
		"TOOL_EXECUTION_CACHE_TTL",
		"TOOL_EXECUTION_ENABLE_DETAILED_LOGGING",
		"TOOL_EXECUTION_PERFORMANCE_ALERT_THRESHOLD_MS",
		"TOOL_EXECUTION_MAX_CPU_USAGE_PERCENT",
		"TOOL_EXECUTION_MAX_MEMORY_USAGE_MB",
		"TOOL_EXECUTION_MAX_QUEUE_DEPTH",
		"DEVELOPMENT_MODE",
	}
	
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}