package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"bodda/internal/models"
)

// mockToolExecutionService implements ToolExecutionService for testing
type mockToolExecutionService struct {
	delay       time.Duration
	shouldError bool
	response    string
}

func (m *mockToolExecutionService) ExecuteGetAthleteProfile(ctx context.Context, msgCtx *MessageContext) (string, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	
	if m.shouldError {
		return "", errors.New("mock error")
	}
	
	return m.response, nil
}

func (m *mockToolExecutionService) ExecuteGetRecentActivities(ctx context.Context, msgCtx *MessageContext, perPage int) (string, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	
	if m.shouldError {
		return "", errors.New("mock error")
	}
	
	return m.response, nil
}

func (m *mockToolExecutionService) ExecuteGetActivityDetails(ctx context.Context, msgCtx *MessageContext, activityID int64) (string, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	
	if m.shouldError {
		return "", errors.New("mock error")
	}
	
	return m.response, nil
}

func (m *mockToolExecutionService) ExecuteGetActivityStreams(ctx context.Context, msgCtx *MessageContext, activityID int64, streamTypes []string, resolution string, processingMode string, pageNumber int, pageSize int, summaryPrompt string) (string, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	
	if m.shouldError {
		return "", errors.New("mock error")
	}
	
	return m.response, nil
}

func (m *mockToolExecutionService) ExecuteUpdateAthleteLogbook(ctx context.Context, msgCtx *MessageContext, content string) (string, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	
	if m.shouldError {
		return "", errors.New("mock error")
	}
	
	return m.response, nil
}

func TestToolExecutorWithTimeout(t *testing.T) {
	// Create mock services
	mockService := &mockToolExecutionService{
		delay:    2 * time.Second, // Longer than timeout
		response: "test response",
	}
	registry := NewToolRegistry()
	executor := NewToolExecutorWithConfig(mockService, registry, 1*time.Second, 5*time.Second)

	// Create test context
	ctx := context.Background()
	msgCtx := &MessageContext{
		UserID:    "test-user",
		SessionID: "test-session",
		Message:   "test message",
	}

	// Test timeout behavior
	options := &models.ExecutionOptions{
		Timeout: 1, // 1 second timeout
	}

	result, err := executor.ExecuteToolWithOptions(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx, options)

	// Should timeout
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if result.Success {
		t.Error("Expected failed result due to timeout")
	}

	if result.Error != "execution timed out" {
		t.Errorf("Expected timeout error message, got: %s", result.Error)
	}

	if result.Duration < 1000 { // Should be at least 1 second
		t.Errorf("Expected duration >= 1000ms, got: %dms", result.Duration)
	}
}

func TestToolExecutorSuccessfulExecution(t *testing.T) {
	// Create mock services
	mockService := &mockToolExecutionService{
		delay:    100 * time.Millisecond, // Short delay
		response: "successful response",
	}
	registry := NewToolRegistry()
	executor := NewToolExecutor(mockService, registry)

	// Create test context
	ctx := context.Background()
	msgCtx := &MessageContext{
		UserID:    "test-user",
		SessionID: "test-session",
		Message:   "test message",
	}

	// Test successful execution
	result, err := executor.ExecuteTool(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful result")
	}

	if result.Data != "successful response" {
		t.Errorf("Expected 'successful response', got: %v", result.Data)
	}

	if result.Duration < 100 { // Should be at least 100ms
		t.Errorf("Expected duration >= 100ms, got: %dms", result.Duration)
	}
}

func TestToolExecutorStreamingMode(t *testing.T) {
	// Create mock services
	mockService := &mockToolExecutionService{
		delay:    50 * time.Millisecond,
		response: "streaming response",
	}
	registry := NewToolRegistry()
	executor := NewToolExecutor(mockService, registry)

	// Create test context
	ctx := context.Background()
	msgCtx := &MessageContext{
		UserID:    "test-user",
		SessionID: "test-session",
		Message:   "test message",
	}

	// Test streaming execution
	options := &models.ExecutionOptions{
		Streaming: true,
		Timeout:   5, // 5 second timeout
	}

	result, err := executor.ExecuteToolWithOptions(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx, options)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful result")
	}

	if result.Data != "streaming response" {
		t.Errorf("Expected 'streaming response', got: %v", result.Data)
	}
}

func TestToolExecutorBufferedMode(t *testing.T) {
	// Create mock services
	mockService := &mockToolExecutionService{
		delay:    50 * time.Millisecond,
		response: "buffered response",
	}
	registry := NewToolRegistry()
	executor := NewToolExecutor(mockService, registry)

	// Create test context
	ctx := context.Background()
	msgCtx := &MessageContext{
		UserID:    "test-user",
		SessionID: "test-session",
		Message:   "test message",
	}

	// Test buffered execution (default mode)
	options := &models.ExecutionOptions{
		BufferedOutput: true,
		Timeout:        5, // 5 second timeout
	}

	result, err := executor.ExecuteToolWithOptions(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx, options)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful result")
	}

	if result.Data != "buffered response" {
		t.Errorf("Expected 'buffered response', got: %v", result.Data)
	}
}

func TestToolExecutorInvalidTool(t *testing.T) {
	// Create mock services
	mockService := &mockToolExecutionService{
		response: "test response",
	}
	registry := NewToolRegistry()
	executor := NewToolExecutor(mockService, registry)

	// Create test context
	ctx := context.Background()
	msgCtx := &MessageContext{
		UserID:    "test-user",
		SessionID: "test-session",
		Message:   "test message",
	}

	// Test invalid tool
	result, err := executor.ExecuteTool(ctx, "invalid-tool", map[string]interface{}{}, msgCtx)

	if err == nil {
		t.Error("Expected error for invalid tool")
	}

	if result.Success {
		t.Error("Expected failed result for invalid tool")
	}

	if result.Error != "tool 'invalid-tool' not found" {
		t.Errorf("Expected tool not found error, got: %s", result.Error)
	}
}

func TestToolExecutorParameterValidation(t *testing.T) {
	// Create mock services
	mockService := &mockToolExecutionService{
		response: "test response",
	}
	registry := NewToolRegistry()
	executor := NewToolExecutor(mockService, registry)

	// Create test context
	ctx := context.Background()
	msgCtx := &MessageContext{
		UserID:    "test-user",
		SessionID: "test-session",
		Message:   "test message",
	}

	// Test missing required parameter for get-activity-details
	result, err := executor.ExecuteTool(ctx, "get-activity-details", map[string]interface{}{}, msgCtx)

	if err == nil {
		t.Error("Expected error for missing required parameter")
	}

	if result.Success {
		t.Error("Expected failed result for missing required parameter")
	}

	if !strings.Contains(result.Error, "parameter validation failed") {
		t.Errorf("Expected parameter validation error, got: %s", result.Error)
	}
}

func TestToolExecutorJobTracking(t *testing.T) {
	// Create mock services
	mockService := &mockToolExecutionService{
		delay:    1 * time.Second, // Long enough to track
		response: "test response",
	}
	registry := NewToolRegistry()
	executor := NewToolExecutor(mockService, registry)

	// Cast to concrete type to access job tracking methods
	concreteExecutor := executor.(*toolExecutor)

	// Check initial job count
	if concreteExecutor.GetActiveJobCount() != 0 {
		t.Errorf("Expected 0 active jobs initially, got: %d", concreteExecutor.GetActiveJobCount())
	}

	// Create test context
	ctx := context.Background()
	msgCtx := &MessageContext{
		UserID:    "test-user",
		SessionID: "test-session",
		Message:   "test message",
	}

	// Start execution in background
	done := make(chan struct{})
	go func() {
		defer close(done)
		executor.ExecuteTool(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx)
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Check that job is tracked
	if concreteExecutor.GetActiveJobCount() != 1 {
		t.Errorf("Expected 1 active job during execution, got: %d", concreteExecutor.GetActiveJobCount())
	}

	// Wait for completion
	<-done

	// Give it a moment to clean up
	time.Sleep(100 * time.Millisecond)

	// Check that job is cleaned up
	if concreteExecutor.GetActiveJobCount() != 0 {
		t.Errorf("Expected 0 active jobs after completion, got: %d", concreteExecutor.GetActiveJobCount())
	}
}

func TestToolExecutorMaxTimeout(t *testing.T) {
	// Create mock services
	mockService := &mockToolExecutionService{
		delay:    100 * time.Millisecond,
		response: "test response",
	}
	registry := NewToolRegistry()
	
	// Create executor with very short max timeout
	executor := NewToolExecutorWithConfig(mockService, registry, 1*time.Second, 2*time.Second)

	// Create test context
	ctx := context.Background()
	msgCtx := &MessageContext{
		UserID:    "test-user",
		SessionID: "test-session",
		Message:   "test message",
	}

	// Test that timeout is capped at max timeout
	options := &models.ExecutionOptions{
		Timeout: 10, // Request 10 seconds, but max is 2
	}

	start := time.Now()
	result, err := executor.ExecuteToolWithOptions(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx, options)
	duration := time.Since(start)

	// Should succeed because mock delay (100ms) is less than max timeout (2s)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful result")
	}

	// Duration should be much less than requested 10 seconds
	if duration > 3*time.Second {
		t.Errorf("Expected duration < 3s (due to max timeout cap), got: %v", duration)
	}
}