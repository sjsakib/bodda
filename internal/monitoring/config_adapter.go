package monitoring

import (
	"bodda/internal/config"
)

// ConvertConfigToMonitoringConfig converts the application config to monitoring system config
func ConvertConfigToMonitoringConfig(appConfig *config.Config) ToolMonitoringConfig {
	return ToolMonitoringConfig{
		EnableParameterLogging: appConfig.ToolMonitoring.EnableParameterLogging,
		PerformanceThresholds: PerformanceThresholds{
			MaxExecutionTimeMs:    appConfig.ToolMonitoring.MaxExecutionTimeMs,
			MaxConcurrentExecs:    appConfig.ToolMonitoring.MaxConcurrentExecs,
			MaxQueueDepth:         appConfig.ToolMonitoring.MaxQueueDepth,
			MaxErrorRatePercent:   appConfig.ToolMonitoring.MaxErrorRatePercent,
			MaxTimeoutRatePercent: appConfig.ToolMonitoring.MaxTimeoutRatePercent,
			AlertRetentionHours:   appConfig.ToolMonitoring.AlertRetentionHours,
		},
		Enabled: appConfig.ToolMonitoring.Enabled,
	}
}

// GetLogLevelFromConfig converts config log level to monitoring log level
func GetLogLevelFromConfig(appConfig *config.Config) LogLevel {
	switch appConfig.ToolMonitoring.LogLevel {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}