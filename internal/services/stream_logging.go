package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// StreamProcessingLogger provides comprehensive logging for stream processing operations
type StreamProcessingLogger struct {
	mu           sync.RWMutex
	logFile      *os.File
	enabled      bool
	logLevel     LogLevel
	operations   []LogEntry
	maxEntries   int
	logDirectory string
}

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// LogEntry represents a single log entry for stream processing
type LogEntry struct {
	Timestamp     time.Time              `json:"timestamp"`
	Level         LogLevel               `json:"level"`
	Operation     string                 `json:"operation"`
	ActivityID    int64                  `json:"activity_id,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	Duration      time.Duration          `json:"duration,omitempty"`
	DataSize      int                    `json:"data_size,omitempty"`
	MemoryUsage   int64                  `json:"memory_usage,omitempty"`
	ProcessingMode string                `json:"processing_mode,omitempty"`
	Success       bool                   `json:"success"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	ErrorType     string                 `json:"error_type,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// NewStreamProcessingLogger creates a new stream processing logger
func NewStreamProcessingLogger(enabled bool, logDirectory string, logLevel LogLevel) (*StreamProcessingLogger, error) {
	logger := &StreamProcessingLogger{
		enabled:      enabled,
		logLevel:     logLevel,
		operations:   make([]LogEntry, 0),
		maxEntries:   10000, // Keep last 10k entries in memory
		logDirectory: logDirectory,
	}

	if enabled {
		if err := logger.initializeLogFile(); err != nil {
			return nil, fmt.Errorf("failed to initialize log file: %w", err)
		}
	}

	return logger, nil
}

// initializeLogFile creates and opens the log file
func (spl *StreamProcessingLogger) initializeLogFile() error {
	if spl.logDirectory == "" {
		spl.logDirectory = "logs"
	}

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(spl.logDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02")
	logFileName := fmt.Sprintf("stream_processing_%s.log", timestamp)
	logFilePath := filepath.Join(spl.logDirectory, logFileName)

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	spl.logFile = file
	return nil
}

// LogOperation logs a stream processing operation
func (spl *StreamProcessingLogger) LogOperation(ctx context.Context, operation string, level LogLevel, metadata map[string]interface{}) {
	if !spl.enabled || level < spl.logLevel {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Operation: operation,
		Success:   true,
		Metadata:  metadata,
	}

	// Extract common metadata
	if activityID, ok := metadata["activity_id"].(int64); ok {
		entry.ActivityID = activityID
	}
	if userID, ok := metadata["user_id"].(string); ok {
		entry.UserID = userID
	}
	if duration, ok := metadata["duration"].(time.Duration); ok {
		entry.Duration = duration
	}
	if dataSize, ok := metadata["data_size"].(int); ok {
		entry.DataSize = dataSize
	}
	if memoryUsage, ok := metadata["memory_usage"].(int64); ok {
		entry.MemoryUsage = memoryUsage
	}
	if processingMode, ok := metadata["processing_mode"].(string); ok {
		entry.ProcessingMode = processingMode
	}

	spl.writeLogEntry(entry)
}

// LogError logs an error during stream processing
func (spl *StreamProcessingLogger) LogError(ctx context.Context, operation string, err error, metadata map[string]interface{}) {
	if !spl.enabled {
		return
	}

	entry := LogEntry{
		Timestamp:    time.Now(),
		Level:        LogLevelError,
		Operation:    operation,
		Success:      false,
		ErrorMessage: err.Error(),
		Metadata:     metadata,
	}

	// Extract error type if it's a StreamProcessingError
	if streamErr, ok := err.(*StreamProcessingError); ok {
		entry.ErrorType = streamErr.Type
		entry.ActivityID = streamErr.ActivityID
		entry.ProcessingMode = streamErr.ProcessingMode
		entry.DataSize = streamErr.DataSize
	}

	// Extract common metadata
	if activityID, ok := metadata["activity_id"].(int64); ok {
		entry.ActivityID = activityID
	}
	if userID, ok := metadata["user_id"].(string); ok {
		entry.UserID = userID
	}
	if duration, ok := metadata["duration"].(time.Duration); ok {
		entry.Duration = duration
	}

	spl.writeLogEntry(entry)
}

// LogPerformanceMetrics logs performance metrics
func (spl *StreamProcessingLogger) LogPerformanceMetrics(metrics StreamPerformanceMetrics) {
	if !spl.enabled {
		return
	}

	metadata := map[string]interface{}{
		"total_operations":      metrics.TotalOperations,
		"total_processing_time": metrics.TotalProcessingTime,
		"success_rate":          metrics.SuccessRate,
		"peak_memory_usage":     metrics.PeakMemoryUsage,
		"average_data_size":     metrics.AverageDataSize,
		"error_counts":          metrics.ErrorCounts,
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LogLevelInfo,
		Operation: "performance_metrics",
		Success:   true,
		Metadata:  metadata,
	}

	spl.writeLogEntry(entry)
}

// writeLogEntry writes a log entry to file and memory
func (spl *StreamProcessingLogger) writeLogEntry(entry LogEntry) {
	spl.mu.Lock()
	defer spl.mu.Unlock()

	// Add to memory buffer
	spl.operations = append(spl.operations, entry)

	// Trim memory buffer if it exceeds max entries
	if len(spl.operations) > spl.maxEntries {
		spl.operations = spl.operations[len(spl.operations)-spl.maxEntries:]
	}

	// Write to file if available
	if spl.logFile != nil {
		jsonEntry, err := json.Marshal(entry)
		if err != nil {
			log.Printf("Failed to marshal log entry: %v", err)
			return
		}

		if _, err := spl.logFile.WriteString(string(jsonEntry) + "\n"); err != nil {
			log.Printf("Failed to write log entry: %v", err)
		}
	}

	// Also log to standard logger for important entries
	if entry.Level >= LogLevelWarn {
		levelStr := spl.levelToString(entry.Level)
		if entry.Success {
			log.Printf("[%s] %s: %s (Duration: %v, DataSize: %d)", 
				levelStr, entry.Operation, "Success", entry.Duration, entry.DataSize)
		} else {
			log.Printf("[%s] %s: %s (Error: %s)", 
				levelStr, entry.Operation, "Failed", entry.ErrorMessage)
		}
	}
}

// levelToString converts log level to string
func (spl *StreamProcessingLogger) levelToString(level LogLevel) string {
	switch level {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// GetRecentOperations returns recent operations from memory
func (spl *StreamProcessingLogger) GetRecentOperations(limit int) []LogEntry {
	spl.mu.RLock()
	defer spl.mu.RUnlock()

	if limit <= 0 || limit > len(spl.operations) {
		limit = len(spl.operations)
	}

	// Return last 'limit' operations
	start := len(spl.operations) - limit
	result := make([]LogEntry, limit)
	copy(result, spl.operations[start:])

	return result
}

// GetOperationsByType returns operations filtered by operation type
func (spl *StreamProcessingLogger) GetOperationsByType(operationType string, limit int) []LogEntry {
	spl.mu.RLock()
	defer spl.mu.RUnlock()

	var filtered []LogEntry
	for i := len(spl.operations) - 1; i >= 0 && len(filtered) < limit; i-- {
		if spl.operations[i].Operation == operationType {
			filtered = append(filtered, spl.operations[i])
		}
	}

	return filtered
}

// GetErrorOperations returns operations that resulted in errors
func (spl *StreamProcessingLogger) GetErrorOperations(limit int) []LogEntry {
	spl.mu.RLock()
	defer spl.mu.RUnlock()

	var errors []LogEntry
	for i := len(spl.operations) - 1; i >= 0 && len(errors) < limit; i-- {
		if !spl.operations[i].Success {
			errors = append(errors, spl.operations[i])
		}
	}

	return errors
}

// GenerateOperationReport generates a comprehensive operation report
func (spl *StreamProcessingLogger) GenerateOperationReport(since time.Time) OperationReport {
	spl.mu.RLock()
	defer spl.mu.RUnlock()

	report := OperationReport{
		GeneratedAt:   time.Now(),
		ReportPeriod:  time.Since(since),
		OperationStats: make(map[string]OperationTypeStats),
	}

	// Analyze operations since the specified time
	for _, entry := range spl.operations {
		if entry.Timestamp.Before(since) {
			continue
		}

		report.TotalOperations++
		if entry.Success {
			report.SuccessfulOperations++
		} else {
			report.FailedOperations++
		}

		report.TotalProcessingTime += entry.Duration
		if entry.DataSize > 0 {
			report.TotalDataProcessed += int64(entry.DataSize)
		}

		// Update operation type stats
		stats, exists := report.OperationStats[entry.Operation]
		if !exists {
			stats = OperationTypeStats{
				OperationType: entry.Operation,
			}
		}

		stats.Count++
		stats.TotalDuration += entry.Duration
		if entry.Success {
			stats.SuccessCount++
		} else {
			stats.ErrorCount++
		}

		if entry.DataSize > 0 {
			stats.TotalDataSize += int64(entry.DataSize)
			if entry.DataSize > stats.MaxDataSize {
				stats.MaxDataSize = entry.DataSize
			}
		}

		if entry.Duration > stats.MaxDuration {
			stats.MaxDuration = entry.Duration
		}

		report.OperationStats[entry.Operation] = stats
	}

	// Calculate derived metrics
	if report.TotalOperations > 0 {
		report.SuccessRate = float64(report.SuccessfulOperations) / float64(report.TotalOperations) * 100
		report.AverageProcessingTime = report.TotalProcessingTime / time.Duration(report.TotalOperations)
	}

	if report.TotalDataProcessed > 0 && report.TotalProcessingTime > 0 {
		report.ThroughputPerSecond = float64(report.TotalDataProcessed) / report.TotalProcessingTime.Seconds()
	}

	// Calculate per-operation averages
	for opType, stats := range report.OperationStats {
		if stats.Count > 0 {
			stats.AverageDuration = stats.TotalDuration / time.Duration(stats.Count)
			stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.Count) * 100
		}
		if stats.TotalDataSize > 0 && stats.Count > 0 {
			stats.AverageDataSize = float64(stats.TotalDataSize) / float64(stats.Count)
		}
		report.OperationStats[opType] = stats
	}

	return report
}

// OperationReport represents a comprehensive report of stream processing operations
type OperationReport struct {
	GeneratedAt            time.Time                      `json:"generated_at"`
	ReportPeriod           time.Duration                  `json:"report_period"`
	TotalOperations        int                            `json:"total_operations"`
	SuccessfulOperations   int                            `json:"successful_operations"`
	FailedOperations       int                            `json:"failed_operations"`
	SuccessRate            float64                        `json:"success_rate"`
	TotalProcessingTime    time.Duration                  `json:"total_processing_time"`
	AverageProcessingTime  time.Duration                  `json:"average_processing_time"`
	TotalDataProcessed     int64                          `json:"total_data_processed"`
	ThroughputPerSecond    float64                        `json:"throughput_per_second"`
	OperationStats         map[string]OperationTypeStats  `json:"operation_stats"`
}

// OperationTypeStats represents statistics for a specific operation type
type OperationTypeStats struct {
	OperationType   string        `json:"operation_type"`
	Count           int           `json:"count"`
	SuccessCount    int           `json:"success_count"`
	ErrorCount      int           `json:"error_count"`
	SuccessRate     float64       `json:"success_rate"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	MaxDuration     time.Duration `json:"max_duration"`
	TotalDataSize   int64         `json:"total_data_size"`
	AverageDataSize float64       `json:"average_data_size"`
	MaxDataSize     int           `json:"max_data_size"`
}

// LogOperationReport logs a comprehensive operation report
func (spl *StreamProcessingLogger) LogOperationReport(since time.Time) {
	if !spl.enabled {
		return
	}

	report := spl.GenerateOperationReport(since)

	log.Printf("=== Stream Processing Operation Report ===")
	log.Printf("Report Period: %v", report.ReportPeriod)
	log.Printf("Total Operations: %d (Success: %d, Failed: %d)", 
		report.TotalOperations, report.SuccessfulOperations, report.FailedOperations)
	log.Printf("Success Rate: %.2f%%", report.SuccessRate)
	log.Printf("Total Processing Time: %v (Average: %v)", 
		report.TotalProcessingTime, report.AverageProcessingTime)
	log.Printf("Total Data Processed: %d points", report.TotalDataProcessed)
	log.Printf("Throughput: %.2f points/second", report.ThroughputPerSecond)

	log.Printf("--- Operation Type Breakdown ---")
	for _, stats := range report.OperationStats {
		log.Printf("%s: %d operations (%.2f%% success, avg duration: %v, avg data size: %.0f)", 
			stats.OperationType, stats.Count, stats.SuccessRate, 
			stats.AverageDuration, stats.AverageDataSize)
	}

	log.Printf("=== End Operation Report ===")
}

// Close closes the logger and flushes any remaining data
func (spl *StreamProcessingLogger) Close() error {
	spl.mu.Lock()
	defer spl.mu.Unlock()

	if spl.logFile != nil {
		if err := spl.logFile.Sync(); err != nil {
			log.Printf("Failed to sync log file: %v", err)
		}
		if err := spl.logFile.Close(); err != nil {
			return fmt.Errorf("failed to close log file: %w", err)
		}
		spl.logFile = nil
	}

	return nil
}

// RotateLogFile rotates the current log file
func (spl *StreamProcessingLogger) RotateLogFile() error {
	spl.mu.Lock()
	defer spl.mu.Unlock()

	if spl.logFile != nil {
		if err := spl.logFile.Close(); err != nil {
			log.Printf("Failed to close current log file: %v", err)
		}
	}

	return spl.initializeLogFile()
}