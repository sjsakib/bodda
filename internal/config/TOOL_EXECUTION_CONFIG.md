# Tool Execution Configuration

This document describes the configuration options for the LLM Tool Execution Endpoint feature.

## Overview

The tool execution endpoint allows clients to execute any available tool programmatically. This endpoint is **development-only** for security reasons and is completely disabled in production environments.

## Environment Variables

### Development Mode

| Variable | Default | Description |
|----------|---------|-------------|
| `DEVELOPMENT_MODE` | `false` | **Required** - Must be `true` to enable the tool execution endpoint |

### Timeout Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `TOOL_EXECUTION_TIMEOUT` | `30` | Default timeout for tool execution (seconds) |
| `TOOL_MAX_TIMEOUT` | `300` | Maximum allowed timeout for tool execution (seconds) |

### Rate Limiting

| Variable | Default | Description |
|----------|---------|-------------|
| `TOOL_EXECUTION_RATE_LIMIT_PER_MINUTE` | `60` | Maximum tool executions per user per minute |
| `TOOL_EXECUTION_MAX_CONCURRENT_EXECUTIONS` | `5` | Maximum concurrent tool executions system-wide |

### Caching

| Variable | Default | Description |
|----------|---------|-------------|
| `TOOL_EXECUTION_ENABLE_CACHING` | `false` | Enable caching of tool execution results |
| `TOOL_EXECUTION_CACHE_TTL` | `300` | Cache time-to-live in seconds |

### Logging

| Variable | Default | Description |
|----------|---------|-------------|
| `TOOL_EXECUTION_ENABLE_DETAILED_LOGGING` | `true` | Enable detailed logging of tool executions |

### Performance Thresholds

| Variable | Default | Description |
|----------|---------|-------------|
| `TOOL_EXECUTION_PERFORMANCE_ALERT_THRESHOLD_MS` | `5000` | Alert threshold for execution time (milliseconds) |
| `TOOL_EXECUTION_MAX_CPU_USAGE_PERCENT` | `80.0` | Maximum CPU usage percentage before alerting |
| `TOOL_EXECUTION_MAX_MEMORY_USAGE_MB` | `512` | Maximum memory usage in MB before alerting |
| `TOOL_EXECUTION_MAX_QUEUE_DEPTH` | `10` | Maximum queue depth before alerting |

## Configuration Validation

The configuration system automatically validates and corrects invalid values:

- **Timeouts**: Negative or zero values are reset to defaults
- **Rate Limits**: Negative or zero values are reset to defaults
- **Performance Thresholds**: Invalid percentages (>100% or <0%) are reset to defaults
- **Memory Limits**: Negative values are reset to defaults

## Environment-Specific Configurations

### Development Environment

```bash
# Enable tool execution endpoint
DEVELOPMENT_MODE=true

# Relaxed settings for development
TOOL_EXECUTION_TIMEOUT=30
TOOL_MAX_TIMEOUT=300
TOOL_EXECUTION_RATE_LIMIT_PER_MINUTE=120
TOOL_EXECUTION_MAX_CONCURRENT_EXECUTIONS=8

# Verbose logging for debugging
TOOL_EXECUTION_ENABLE_DETAILED_LOGGING=true

# Relaxed performance thresholds
TOOL_EXECUTION_PERFORMANCE_ALERT_THRESHOLD_MS=10000
TOOL_EXECUTION_MAX_CPU_USAGE_PERCENT=90.0
TOOL_EXECUTION_MAX_MEMORY_USAGE_MB=1024
TOOL_EXECUTION_MAX_QUEUE_DEPTH=20
```

### Production Environment

```bash
# DISABLE tool execution endpoint (security)
DEVELOPMENT_MODE=false

# Strict settings for production
TOOL_EXECUTION_TIMEOUT=15
TOOL_MAX_TIMEOUT=60
TOOL_EXECUTION_RATE_LIMIT_PER_MINUTE=30
TOOL_EXECUTION_MAX_CONCURRENT_EXECUTIONS=3

# Enable caching for performance
TOOL_EXECUTION_ENABLE_CACHING=true
TOOL_EXECUTION_CACHE_TTL=600

# Minimal logging for production
TOOL_EXECUTION_ENABLE_DETAILED_LOGGING=false

# Strict performance thresholds
TOOL_EXECUTION_PERFORMANCE_ALERT_THRESHOLD_MS=3000
TOOL_EXECUTION_MAX_CPU_USAGE_PERCENT=70.0
TOOL_EXECUTION_MAX_MEMORY_USAGE_MB=256
TOOL_EXECUTION_MAX_QUEUE_DEPTH=5
```

## Usage in Code

### Checking if Tool Execution is Enabled

```go
config := config.Load()

if config.IsToolExecutionEnabled() {
    // Tool execution endpoint is available
    // Register routes and middleware
} else {
    // Tool execution is disabled (production mode)
    // Return 404 for any tool execution requests
}
```

### Getting Appropriate Timeout

```go
config := config.Load()

// Get timeout with validation
timeout := config.GetToolExecutionTimeout(requestedTimeout)

// This will:
// - Return default timeout if requestedTimeout <= 0
// - Return requestedTimeout if valid
// - Return max timeout if requestedTimeout > max
```

### Accessing Configuration Values

```go
config := config.Load()

// Access tool execution settings
toolConfig := config.ToolExecution

// Timeout settings
defaultTimeout := toolConfig.DefaultTimeout
maxTimeout := toolConfig.MaxTimeout

// Rate limiting
rateLimit := toolConfig.RateLimitPerMinute
maxConcurrent := toolConfig.MaxConcurrentExecs

// Performance thresholds
thresholds := toolConfig.PerformanceThresholds
maxExecTime := thresholds.MaxExecutionTimeMs
maxCPU := thresholds.MaxCPUUsagePercent
maxMemory := thresholds.MaxMemoryUsageMB
maxQueue := thresholds.MaxQueueDepth
```

## Security Considerations

1. **Development Only**: The tool execution endpoint is completely disabled in production
2. **Environment Variable Required**: `DEVELOPMENT_MODE=true` must be explicitly set
3. **Production Safety**: Setting `DEVELOPMENT_MODE=false` (or omitting it) disables the endpoint
4. **404 Response**: Production environments return 404 as if the endpoint doesn't exist

## Monitoring Integration

The configuration integrates with the existing monitoring system:

- Performance thresholds trigger alerts when exceeded
- Detailed logging can be enabled/disabled per environment
- Resource usage monitoring uses configured thresholds
- Queue depth monitoring prevents system overload

## Testing

The configuration includes comprehensive tests:

- Default value validation
- Environment variable parsing
- Configuration validation and correction
- Helper function behavior
- Development mode detection

Run tests with:
```bash
go test ./internal/config -v
```