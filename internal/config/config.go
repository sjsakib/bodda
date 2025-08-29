package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	
	// Strava OAuth
	StravaClientID     string
	StravaClientSecret string
	StravaRedirectURL  string
	
	// OpenAI
	OpenAIAPIKey string
	
	// Frontend URL for CORS
	FrontendURL string
	
	// Development mode
	IsDevelopment bool
	
	// Stream Processing
	StreamProcessing StreamProcessingConfig
	
	// Tool Execution Monitoring
	ToolMonitoring ToolMonitoringConfig
	
	// Tool Execution Settings
	ToolExecution ToolExecutionConfig
}

// StreamProcessingConfig holds configuration for stream data processing
type StreamProcessingConfig struct {
	MaxContextTokens    int
	TokenPerCharRatio   float64
	DefaultPageSize     int
	MaxPageSize         int
	RedactionEnabled    bool
	StravaResolutions   []string
	
	// Feature flags for processing modes
	EnableDerivedFeatures bool
	EnableAISummary      bool
	EnablePagination     bool
	EnableAutoMode       bool
	
	// Processing thresholds
	LargeDatasetThreshold int
	ContextSafetyMargin   int
	MaxRetries           int
	ProcessingTimeout    int // seconds
}

// ToolMonitoringConfig holds configuration for tool execution monitoring
type ToolMonitoringConfig struct {
	Enabled                bool
	EnableParameterLogging bool
	
	// Performance thresholds
	MaxExecutionTimeMs    int64
	MaxConcurrentExecs    int64
	MaxQueueDepth         int64
	MaxErrorRatePercent   float64
	MaxTimeoutRatePercent float64
	AlertRetentionHours   int
	
	// Rate limiting
	RateLimitPerMinute    int
	MaxConcurrentPerUser  int64
	
	// Logging configuration
	EnableDetailedLogging bool
	LogLevel             string
}

// ToolExecutionConfig holds configuration for tool execution settings
type ToolExecutionConfig struct {
	// Timeout settings
	DefaultTimeout         int  // seconds
	MaxTimeout            int  // seconds
	
	// Rate limiting
	RateLimitPerMinute    int
	MaxConcurrentExecs    int
	
	// Caching
	EnableCaching         bool
	CacheTTL             int  // seconds
	
	// Logging
	EnableDetailedLogging bool
	
	// Performance thresholds
	PerformanceThresholds PerformanceThresholds
}

// PerformanceThresholds holds performance monitoring thresholds
type PerformanceThresholds struct {
	MaxExecutionTimeMs int     // milliseconds
	MaxCPUUsagePercent float64 // percentage
	MaxMemoryUsageMB   int64   // megabytes
	MaxQueueDepth      int     // number of queued requests
}

func Load() *Config {
	config := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/bodda?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		
		StravaClientID:     getEnv("STRAVA_CLIENT_ID", ""),
		StravaClientSecret: getEnv("STRAVA_CLIENT_SECRET", ""),
		StravaRedirectURL:  getEnv("STRAVA_REDIRECT_URL", "http://localhost:8080/auth/callback"),
		
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
		
		IsDevelopment: getEnvBool("DEVELOPMENT_MODE", true),
		
		StreamProcessing: StreamProcessingConfig{
			MaxContextTokens:      getEnvInt("STREAM_MAX_CONTEXT_TOKENS", 15000),
			TokenPerCharRatio:     getEnvFloat("STREAM_TOKEN_PER_CHAR_RATIO", 0.25),
			DefaultPageSize:       getEnvInt("STREAM_DEFAULT_PAGE_SIZE", 1000),
			MaxPageSize:           getEnvInt("STREAM_MAX_PAGE_SIZE", 5000),
			RedactionEnabled:      getEnvBool("STREAM_REDACTION_ENABLED", true),
			StravaResolutions:     getEnvStringSlice("STREAM_STRAVA_RESOLUTIONS", []string{"low", "medium", "high"}),
			
			// Feature flags
			EnableDerivedFeatures: getEnvBool("STREAM_ENABLE_DERIVED_FEATURES", true),
			EnableAISummary:      getEnvBool("STREAM_ENABLE_AI_SUMMARY", true),
			EnablePagination:     getEnvBool("STREAM_ENABLE_PAGINATION", true),
			EnableAutoMode:       getEnvBool("STREAM_ENABLE_AUTO_MODE", true),
			
			// Processing thresholds
			LargeDatasetThreshold: getEnvInt("STREAM_LARGE_DATASET_THRESHOLD", 10000),
			ContextSafetyMargin:   getEnvInt("STREAM_CONTEXT_SAFETY_MARGIN", 2000),
			MaxRetries:           getEnvInt("STREAM_MAX_RETRIES", 3),
			ProcessingTimeout:    getEnvInt("STREAM_PROCESSING_TIMEOUT", 30),
		},
		
		ToolMonitoring: ToolMonitoringConfig{
			Enabled:                getEnvBool("TOOL_MONITORING_ENABLED", true),
			EnableParameterLogging: getEnvBool("TOOL_ENABLE_PARAMETER_LOGGING", true),
			
			// Performance thresholds
			MaxExecutionTimeMs:    int64(getEnvInt("TOOL_MAX_EXECUTION_TIME_MS", 30000)),
			MaxConcurrentExecs:    int64(getEnvInt("TOOL_MAX_CONCURRENT_EXECUTIONS", 10)),
			MaxQueueDepth:         int64(getEnvInt("TOOL_MAX_QUEUE_DEPTH", 20)),
			MaxErrorRatePercent:   getEnvFloat("TOOL_MAX_ERROR_RATE_PERCENT", 10.0),
			MaxTimeoutRatePercent: getEnvFloat("TOOL_MAX_TIMEOUT_RATE_PERCENT", 5.0),
			AlertRetentionHours:   getEnvInt("TOOL_ALERT_RETENTION_HOURS", 24),
			
			// Rate limiting
			RateLimitPerMinute:   getEnvInt("TOOL_RATE_LIMIT_PER_MINUTE", 60),
			MaxConcurrentPerUser: int64(getEnvInt("TOOL_MAX_CONCURRENT_PER_USER", 3)),
			
			// Logging configuration
			EnableDetailedLogging: getEnvBool("TOOL_ENABLE_DETAILED_LOGGING", true),
			LogLevel:             getEnv("TOOL_LOG_LEVEL", "info"),
		},
		
		ToolExecution: ToolExecutionConfig{
			// Timeout settings
			DefaultTimeout:        getEnvInt("TOOL_EXECUTION_TIMEOUT", 30),
			MaxTimeout:           getEnvInt("TOOL_MAX_TIMEOUT", 300),
			
			// Rate limiting
			RateLimitPerMinute:   getEnvInt("TOOL_EXECUTION_RATE_LIMIT_PER_MINUTE", 60),
			MaxConcurrentExecs:   getEnvInt("TOOL_EXECUTION_MAX_CONCURRENT_EXECUTIONS", 5),
			
			// Caching
			EnableCaching:        getEnvBool("TOOL_EXECUTION_ENABLE_CACHING", false),
			CacheTTL:            getEnvInt("TOOL_EXECUTION_CACHE_TTL", 300),
			
			// Logging
			EnableDetailedLogging: getEnvBool("TOOL_EXECUTION_ENABLE_DETAILED_LOGGING", true),
			
			// Performance thresholds
			PerformanceThresholds: PerformanceThresholds{
				MaxExecutionTimeMs: getEnvInt("TOOL_EXECUTION_PERFORMANCE_ALERT_THRESHOLD_MS", 5000),
				MaxCPUUsagePercent: getEnvFloat("TOOL_EXECUTION_MAX_CPU_USAGE_PERCENT", 80.0),
				MaxMemoryUsageMB:   int64(getEnvInt("TOOL_EXECUTION_MAX_MEMORY_USAGE_MB", 512)),
				MaxQueueDepth:      getEnvInt("TOOL_EXECUTION_MAX_QUEUE_DEPTH", 10),
			},
		},
	}
	
	// Validate configuration
	config.validateStreamProcessingConfig()
	config.validateToolMonitoringConfig()
	config.validateToolExecutionConfig()
	
	return config
}

// validateStreamProcessingConfig ensures stream processing configuration is valid
func (c *Config) validateStreamProcessingConfig() {
	sp := &c.StreamProcessing
	
	// Ensure page sizes are reasonable
	if sp.DefaultPageSize <= 0 {
		sp.DefaultPageSize = 1000
	}
	if sp.MaxPageSize <= 0 || sp.MaxPageSize < sp.DefaultPageSize {
		sp.MaxPageSize = sp.DefaultPageSize * 5
	}
	
	// Ensure context limits are reasonable
	if sp.MaxContextTokens <= 0 {
		sp.MaxContextTokens = 15000
	}
	if sp.ContextSafetyMargin <= 0 {
		sp.ContextSafetyMargin = 2000
	}
	if sp.ContextSafetyMargin >= sp.MaxContextTokens {
		sp.ContextSafetyMargin = sp.MaxContextTokens / 10
	}
	
	// Ensure token ratio is reasonable
	if sp.TokenPerCharRatio <= 0 || sp.TokenPerCharRatio > 1 {
		sp.TokenPerCharRatio = 0.25
	}
	
	// Ensure thresholds are reasonable
	if sp.LargeDatasetThreshold <= 0 {
		sp.LargeDatasetThreshold = 10000
	}
	
	// Ensure retry and timeout values are reasonable
	if sp.MaxRetries < 0 {
		sp.MaxRetries = 3
	}
	if sp.ProcessingTimeout <= 0 {
		sp.ProcessingTimeout = 30
	}
	
	// Ensure we have valid Strava resolutions
	if len(sp.StravaResolutions) == 0 {
		sp.StravaResolutions = []string{"low", "medium", "high"}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		result := make([]string, 0)
		for _, item := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

// validateToolMonitoringConfig ensures tool monitoring configuration is valid
func (c *Config) validateToolMonitoringConfig() {
	tm := &c.ToolMonitoring
	
	// Ensure execution time threshold is reasonable
	if tm.MaxExecutionTimeMs <= 0 {
		tm.MaxExecutionTimeMs = 30000 // 30 seconds
	}
	
	// Ensure concurrent execution limits are reasonable
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
	
	// Ensure queue depth is reasonable
	if tm.MaxQueueDepth <= 0 {
		tm.MaxQueueDepth = 20
	}
	
	// Ensure error rate thresholds are reasonable
	if tm.MaxErrorRatePercent < 0 || tm.MaxErrorRatePercent > 100 {
		tm.MaxErrorRatePercent = 10.0
	}
	if tm.MaxTimeoutRatePercent < 0 || tm.MaxTimeoutRatePercent > 100 {
		tm.MaxTimeoutRatePercent = 5.0
	}
	
	// Ensure alert retention is reasonable
	if tm.AlertRetentionHours <= 0 {
		tm.AlertRetentionHours = 24
	}
	
	// Ensure rate limiting is reasonable
	if tm.RateLimitPerMinute <= 0 {
		tm.RateLimitPerMinute = 60
	}
	
	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[strings.ToLower(tm.LogLevel)] {
		tm.LogLevel = "info"
	}
}

// validateToolExecutionConfig ensures tool execution configuration is valid
func (c *Config) validateToolExecutionConfig() {
	te := &c.ToolExecution
	
	// Ensure timeout values are reasonable
	if te.DefaultTimeout <= 0 {
		te.DefaultTimeout = 30 // 30 seconds
	}
	if te.MaxTimeout <= 0 || te.MaxTimeout < te.DefaultTimeout {
		te.MaxTimeout = 300 // 5 minutes
	}
	
	// Ensure rate limiting is reasonable
	if te.RateLimitPerMinute <= 0 {
		te.RateLimitPerMinute = 60
	}
	if te.MaxConcurrentExecs <= 0 {
		te.MaxConcurrentExecs = 5
	}
	
	// Ensure cache TTL is reasonable
	if te.CacheTTL <= 0 {
		te.CacheTTL = 300 // 5 minutes
	}
	
	// Validate performance thresholds
	pt := &te.PerformanceThresholds
	
	if pt.MaxExecutionTimeMs <= 0 {
		pt.MaxExecutionTimeMs = 5000 // 5 seconds
	}
	
	if pt.MaxCPUUsagePercent <= 0 || pt.MaxCPUUsagePercent > 100 {
		pt.MaxCPUUsagePercent = 80.0
	}
	
	if pt.MaxMemoryUsageMB <= 0 {
		pt.MaxMemoryUsageMB = 512 // 512 MB
	}
	
	if pt.MaxQueueDepth <= 0 {
		pt.MaxQueueDepth = 10
	}
}

// IsToolExecutionEnabled returns true if tool execution endpoint should be available
func (c *Config) IsToolExecutionEnabled() bool {
	return c.IsDevelopment
}

// GetToolExecutionTimeout returns the appropriate timeout for tool execution
func (c *Config) GetToolExecutionTimeout(requestedTimeout int) int {
	if requestedTimeout <= 0 {
		return c.ToolExecution.DefaultTimeout
	}
	
	if requestedTimeout > c.ToolExecution.MaxTimeout {
		return c.ToolExecution.MaxTimeout
	}
	
	return requestedTimeout
}