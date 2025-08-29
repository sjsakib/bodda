package monitoring

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
	level slog.Level
}

// NewLogger creates a new structured logger
func NewLogger(level LogLevel, isDevelopment bool) *Logger {
	var slogLevel slog.Level
	switch strings.ToLower(string(level)) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: slogLevel,
		AddSource: isDevelopment,
	}

	if isDevelopment {
		// Pretty text output for development
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		// JSON output for production
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	return &Logger{
		Logger: logger,
		level:  slogLevel,
	}
}

// WithContext adds context information to the logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract request ID or other context values if available
	if requestID := ctx.Value("request_id"); requestID != nil {
		return &Logger{
			Logger: l.Logger.With("request_id", requestID),
			level:  l.level,
		}
	}
	return l
}

// WithFields adds structured fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		Logger: l.Logger.With(args...),
		level:  l.level,
	}
}

// HTTP request logging middleware for Gin
func (l *Logger) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		
		// Get status code and size
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		// Build full path
		if raw != "" {
			path = path + "?" + raw
		}

		// Determine log level based on status code
		logLevel := slog.LevelInfo
		if statusCode >= 400 && statusCode < 500 {
			logLevel = slog.LevelWarn
		} else if statusCode >= 500 {
			logLevel = slog.LevelError
		}

		// Log the request
		l.Logger.Log(context.Background(), logLevel, "HTTP Request",
			"method", c.Request.Method,
			"path", path,
			"status", statusCode,
			"latency", latency.String(),
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"body_size", bodySize,
		)
	}
}

// Error logging helper
func (l *Logger) LogError(ctx context.Context, err error, message string, fields ...interface{}) {
	args := []interface{}{"error", err.Error()}
	args = append(args, fields...)
	l.WithContext(ctx).Error(message, args...)
}

// API operation logging helper
func (l *Logger) LogAPIOperation(ctx context.Context, operation string, duration time.Duration, success bool, fields ...interface{}) {
	args := []interface{}{
		"operation", operation,
		"duration", duration.String(),
		"success", success,
	}
	args = append(args, fields...)
	
	if success {
		l.WithContext(ctx).Info("API Operation", args...)
	} else {
		l.WithContext(ctx).Warn("API Operation Failed", args...)
	}
}

// Database operation logging helper
func (l *Logger) LogDBOperation(ctx context.Context, query string, duration time.Duration, err error) {
	fields := []interface{}{
		"query_type", extractQueryType(query),
		"duration", duration.String(),
	}
	
	if err != nil {
		fields = append(fields, "error", err.Error())
		l.WithContext(ctx).Error("Database Operation Failed", fields...)
	} else {
		l.WithContext(ctx).Debug("Database Operation", fields...)
	}
}

// extractQueryType extracts the type of SQL query (SELECT, INSERT, etc.)
func extractQueryType(query string) string {
	query = strings.TrimSpace(strings.ToUpper(query))
	parts := strings.Fields(query)
	if len(parts) > 0 {
		return parts[0]
	}
	return "UNKNOWN"
}

// Global logger instance
var globalLogger *Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(level LogLevel, isDevelopment bool) {
	globalLogger = NewLogger(level, isDevelopment)
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if globalLogger == nil {
		// Fallback to a default logger
		globalLogger = NewLogger(LevelInfo, false)
	}
	return globalLogger
}

// Convenience functions for global logger
func Debug(msg string, args ...interface{}) {
	GetLogger().Debug(msg, args...)
}

func Info(msg string, args ...interface{}) {
	GetLogger().Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	GetLogger().Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	GetLogger().Error(msg, args...)
}

func LogError(ctx context.Context, err error, message string, fields ...interface{}) {
	GetLogger().LogError(ctx, err, message, fields...)
}