# Tool Execution Monitoring System

This package provides comprehensive monitoring and logging capabilities for tool executions in the LLM Tool Execution Endpoint feature.

## Features

### 1. Execution Logging
- **Comprehensive Logging**: Logs all tool executions with timestamp, user, tool name, and duration
- **Error Logging**: Detailed error logging with stack traces and parameter values
- **Parameter Sanitization**: Automatically redacts sensitive parameters (passwords, API keys, tokens)
- **Structured Logging**: JSON-formatted logs for easy parsing and analysis

### 2. Performance Metrics
- **Execution Metrics**: Tracks execution count, success/error rates, timing statistics
- **Concurrent Execution Tracking**: Monitors active executions and queue depth
- **User-based Metrics**: Tracks per-user execution statistics
- **Real-time Alerting**: Configurable alerts for performance thresholds

### 3. Alerting System
- **Performance Alerts**: Alerts for slow executions, high error rates, high timeout rates
- **Resource Alerts**: Alerts for high concurrent executions and queue depth
- **Configurable Thresholds**: All alert thresholds are configurable via environment variables
- **Alert Retention**: Configurable retention period for alerts

## Configuration

The monitoring system is configured through environment variables:

```bash
# Enable/disable monitoring
TOOL_MONITORING_ENABLED=true

# Logging configuration
TOOL_ENABLE_PARAMETER_LOGGING=true
TOOL_ENABLE_DETAILED_LOGGING=true
TOOL_LOG_LEVEL=info

# Performance thresholds
TOOL_MAX_EXECUTION_TIME_MS=30000
TOOL_MAX_CONCURRENT_EXECUTIONS=10
TOOL_MAX_QUEUE_DEPTH=20
TOOL_MAX_ERROR_RATE_PERCENT=10.0
TOOL_MAX_TIMEOUT_RATE_PERCENT=5.0

# Alert configuration
TOOL_ALERT_RETENTION_HOURS=24

# Rate limiting
TOOL_RATE_LIMIT_PER_MINUTE=60
TOOL_MAX_CONCURRENT_PER_USER=3
```

## Usage

### Basic Setup

```go
import (
    "bodda/internal/config"
    "bodda/internal/monitoring"
)

// Load configuration
appConfig := config.Load()

// Create logger
logger := monitoring.NewLogger(
    monitoring.GetLogLevelFromConfig(appConfig), 
    appConfig.IsDevelopment,
)

// Initialize global monitoring
monitoringConfig := monitoring.ConvertConfigToMonitoringConfig(appConfig)
monitoring.InitGlobalToolMonitoring(logger, monitoringConfig)
```

### Monitoring Tool Execution

```go
// Using the global monitoring system
result, err := monitoring.MonitorToolExecution(
    ctx, 
    "tool-name", 
    parameters, 
    userID, 
    func() (interface{}, int64, error) {
        // Your tool execution logic here
        result := executeMyTool(parameters)
        responseSize := int64(len(result))
        return result, responseSize, nil
    },
)
```

### Using Monitoring Middleware

```go
// Create monitoring middleware
monitoringSystem := monitoring.GetGlobalToolMonitoring()
middleware := monitoring.NewMonitoringMiddleware(monitoringSystem)

// Wrap tool execution function
wrappedExecution := middleware.WrapToolExecution(
    "tool-name", 
    userID, 
    func(ctx context.Context, parameters map[string]interface{}) (interface{}, int64, error) {
        // Your tool execution logic
        return executeMyTool(ctx, parameters)
    },
)

// Execute with monitoring
result, err := wrappedExecution(ctx, parameters)
```

### Monitoring Tool Validation

```go
// Monitor parameter validation errors
if validationErr := validateParameters(parameters); validationErr != nil {
    monitoring.MonitorToolValidation(ctx, toolName, parameters, userID, validationErr)
    return nil, validationErr
}
```

### Monitoring Tool Discovery

```go
// Monitor tool discovery operations
err := monitoring.MonitorToolDiscovery(ctx, "list_tools", "", userID, func() error {
    // Your tool discovery logic
    return listAvailableTools()
})
```

## HTTP Endpoints

The monitoring system provides several HTTP endpoints for accessing metrics and alerts:

### Performance Metrics
```
GET /monitoring/tools/performance
```
Returns comprehensive performance metrics for all tools.

### Tool-Specific Metrics
```
GET /monitoring/tools/metrics/{toolName}
```
Returns detailed metrics for a specific tool.

### Recent Alerts
```
GET /monitoring/tools/alerts?hours=24
```
Returns recent performance alerts (default: last 24 hours).

## Example Response Formats

### Performance Metrics Response
```json
{
  "tool_performance_metrics": {
    "execution_count": {
      "get-athlete-profile": 150,
      "get-recent-activities": 89
    },
    "success_count": {
      "get-athlete-profile": 145,
      "get-recent-activities": 87
    },
    "error_count": {
      "get-athlete-profile": 5,
      "get-recent-activities": 2
    },
    "average_execution_time_ms": {
      "get-athlete-profile": 1250.5,
      "get-recent-activities": 890.2
    },
    "concurrent_executions": 3,
    "queue_depth": 0
  },
  "timestamp": "2025-08-28T17:46:18Z"
}
```

### Tool-Specific Metrics Response
```json
{
  "tool_metrics": {
    "tool_name": "get-athlete-profile",
    "execution_count": 150,
    "success_count": 145,
    "error_count": 5,
    "timeout_count": 1,
    "success_rate_percent": 96.67,
    "error_rate_percent": 3.33,
    "timeout_rate_percent": 0.67,
    "average_execution_time": 1250.5,
    "max_execution_time": 5000,
    "min_execution_time": 200,
    "active_executions": 1,
    "last_execution_time": "2025-08-28T17:45:30Z"
  },
  "timestamp": "2025-08-28T17:46:18Z"
}
```

### Alerts Response
```json
{
  "alerts": [
    {
      "timestamp": "2025-08-28T17:40:15Z",
      "alert_type": "slow_execution",
      "tool_name": "get-activity-streams",
      "user_id": "user123",
      "message": "Tool execution exceeded time threshold",
      "severity": "warning",
      "value": 35000,
      "threshold": 30000
    },
    {
      "timestamp": "2025-08-28T17:35:22Z",
      "alert_type": "high_error_rate",
      "tool_name": "get-athlete-profile",
      "message": "Tool has high error rate",
      "severity": "error",
      "value": 12.5,
      "threshold": 10.0
    }
  ],
  "hours": 24,
  "count": 2,
  "timestamp": "2025-08-28T17:46:18Z"
}
```

## Log Format Examples

### Successful Execution Log
```
time=2025-08-28T17:46:18.309+06:00 level=INFO msg="Tool execution completed" 
request_id=req_1756403178309399000 tool_name=get-athlete-profile user_id=user123 
duration_ms=1250 response_size_bytes=2048 success=true
```

### Error Execution Log
```
time=2025-08-28T17:46:18.309+06:00 level=ERROR msg="Tool execution failed" 
request_id=req_1756403178309779000 tool_name=get-recent-activities user_id=user123 
duration_ms=5000 error="API rate limit exceeded" success=false
```

### Performance Alert Log
```
time=2025-08-28T17:46:18.310+06:00 level=WARN msg="Performance alert" 
alert_type=slow_execution tool_name=get-activity-streams user_id=user123 
severity=warning message="Tool execution exceeded time threshold" value=35000 threshold=30000
```

## Integration with Existing Systems

The monitoring system is designed to integrate seamlessly with the existing tool execution infrastructure:

1. **Configuration Integration**: Uses the existing config system with validation
2. **Logging Integration**: Extends the existing structured logging system
3. **Metrics Integration**: Integrates with the existing metrics collection system
4. **HTTP Integration**: Adds endpoints to the existing monitoring routes

## Testing

The monitoring system includes comprehensive tests:

```bash
# Run all monitoring tests
go test ./internal/monitoring/... -v

# Run specific test suites
go test ./internal/monitoring/ -run TestToolExecutionLogger -v
go test ./internal/monitoring/ -run TestToolPerformanceTracker -v
go test ./internal/monitoring/ -run TestToolMonitoringSystem -v
```

## Security Considerations

- **Parameter Sanitization**: Sensitive parameters are automatically redacted in logs
- **Development-Only**: The monitoring endpoints respect the development-only restriction
- **Rate Limiting**: Built-in rate limiting prevents abuse
- **User Context**: All operations are tracked with user context for audit trails

## Performance Impact

The monitoring system is designed to have minimal performance impact:

- **Asynchronous Processing**: Alerts and detailed logging are processed asynchronously
- **Efficient Data Structures**: Uses efficient concurrent data structures for metrics
- **Configurable Detail Level**: Detailed logging can be disabled in production
- **Memory Management**: Automatic cleanup of old alerts and metrics