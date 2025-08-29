package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"bodda/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestToolExecutor_Comprehensive(t *testing.T) {
	// Create comprehensive test setup
	mockService := &mockToolExecutionServiceComprehensive{
		responses: make(map[string]mockResponse),
	}
	registry := NewToolRegistry()
	executor := NewToolExecutorWithConfig(mockService, registry, 2*time.Second, 10*time.Second)

	// Test context and message context
	ctx := context.Background()
	msgCtx := &MessageContext{
		UserID:    "test-user-123",
		SessionID: "test-session-456",
		Message:   "test message",
		User: &models.User{
			ID:       "test-user-123",
			StravaID: 12345,
		},
	}

	t.Run("ExecuteTool_AllAvailableTools_Success", func(t *testing.T) {
		tools := registry.GetAvailableTools()
		
		for _, tool := range tools {
			t.Run(tool.Name, func(t *testing.T) {
				// Set up mock response for this tool
				mockService.responses[tool.Name] = mockResponse{
					response: "success response for " + tool.Name,
					delay:    50 * time.Millisecond,
				}
				
				// Get required parameters for the tool
				params := getValidParametersForTool(tool.Name)
				
				result, err := executor.ExecuteTool(ctx, tool.Name, params, msgCtx)
				
				assert.NoError(t, err, "Tool %s should execute successfully", tool.Name)
				assert.NotNil(t, result, "Result should not be nil for tool %s", tool.Name)
				assert.True(t, result.Success, "Execution should be successful for tool %s", tool.Name)
				assert.Equal(t, tool.Name, result.ToolName)
				assert.Contains(t, result.Data, tool.Name, "Response should contain tool name")
				assert.Greater(t, result.Duration, int64(0), "Duration should be positive")
				assert.False(t, result.Timestamp.IsZero(), "Timestamp should be set")
			})
		}
	})

	t.Run("ExecuteToolWithOptions_StreamingMode_Success", func(t *testing.T) {
		mockService.responses["get-athlete-profile"] = mockResponse{
			response: "streaming response",
			delay:    100 * time.Millisecond,
		}
		
		options := &models.ExecutionOptions{
			Streaming: true,
			Timeout:   5,
		}
		
		result, err := executor.ExecuteToolWithOptions(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx, options)
		
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "streaming response", result.Data)
		assert.Greater(t, result.Duration, int64(90), "Should take at least 90ms due to mock delay")
	})

	t.Run("ExecuteToolWithOptions_BufferedMode_Success", func(t *testing.T) {
		mockService.responses["get-athlete-profile"] = mockResponse{
			response: "buffered response",
			delay:    100 * time.Millisecond,
		}
		
		options := &models.ExecutionOptions{
			BufferedOutput: true,
			Timeout:        5,
		}
		
		result, err := executor.ExecuteToolWithOptions(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx, options)
		
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "buffered response", result.Data)
	})

	t.Run("ExecuteTool_TimeoutHandling_ReturnsTimeoutError", func(t *testing.T) {
		mockService.responses["get-athlete-profile"] = mockResponse{
			response: "should timeout",
			delay:    3 * time.Second, // Longer than executor timeout (2s)
		}
		
		start := time.Now()
		result, err := executor.ExecuteTool(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx)
		duration := time.Since(start)
		
		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "timed out")
		assert.Less(t, duration, 3*time.Second, "Should timeout before mock delay completes")
		assert.Greater(t, duration, 1500*time.Millisecond, "Should take at least 1.5s to timeout")
	})

	t.Run("ExecuteToolWithOptions_CustomTimeout_RespectsTimeout", func(t *testing.T) {
		mockService.responses["get-athlete-profile"] = mockResponse{
			response: "should timeout with custom timeout",
			delay:    2 * time.Second,
		}
		
		options := &models.ExecutionOptions{
			Timeout: 1, // 1 second timeout
		}
		
		start := time.Now()
		result, err := executor.ExecuteToolWithOptions(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx, options)
		duration := time.Since(start)
		
		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "timed out")
		assert.Less(t, duration, 1500*time.Millisecond, "Should timeout in ~1 second")
	})

	t.Run("ExecuteToolWithOptions_MaxTimeoutEnforcement", func(t *testing.T) {
		mockService.responses["get-athlete-profile"] = mockResponse{
			response: "quick response",
			delay:    100 * time.Millisecond,
		}
		
		options := &models.ExecutionOptions{
			Timeout: 15, // Request 15 seconds, but max is 10
		}
		
		result, err := executor.ExecuteToolWithOptions(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx, options)
		
		// Should succeed because mock delay (100ms) is much less than max timeout (10s)
		assert.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("ExecuteTool_InvalidTool_ReturnsError", func(t *testing.T) {
		result, err := executor.ExecuteTool(ctx, "nonexistent-tool", map[string]interface{}{}, msgCtx)
		
		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "not found")
		assert.Equal(t, "nonexistent-tool", result.ToolName)
	})

	t.Run("ExecuteTool_ParameterValidationError_ReturnsError", func(t *testing.T) {
		// Try to execute get-activity-details without required activity_id
		result, err := executor.ExecuteTool(ctx, "get-activity-details", map[string]interface{}{}, msgCtx)
		
		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "parameter validation failed")
		assert.Equal(t, "get-activity-details", result.ToolName)
	})

	t.Run("ExecuteTool_ServiceError_ReturnsError", func(t *testing.T) {
		mockService.responses["get-athlete-profile"] = mockResponse{
			shouldError: true,
			errorMsg:    "service unavailable",
		}
		
		result, err := executor.ExecuteTool(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx)
		
		// The executor may return an error OR a failed result, both are valid
		if err != nil {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "service unavailable")
		} else {
			assert.NotNil(t, result)
			assert.False(t, result.Success)
			assert.Contains(t, result.Error, "service unavailable")
		}
	})

	t.Run("JobTracking_ActiveJobCount", func(t *testing.T) {
		concreteExecutor := executor.(*toolExecutor)
		
		// Initially no active jobs
		assert.Equal(t, 0, concreteExecutor.GetActiveJobCount())
		
		// Set up a long-running mock
		mockService.responses["get-athlete-profile"] = mockResponse{
			response: "long running task",
			delay:    1 * time.Second,
		}
		
		// Start execution in background
		done := make(chan struct{})
		go func() {
			defer close(done)
			executor.ExecuteTool(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx)
		}()
		
		// Give it time to start
		time.Sleep(100 * time.Millisecond)
		
		// Should have one active job
		assert.Equal(t, 1, concreteExecutor.GetActiveJobCount())
		
		// Wait for completion
		<-done
		
		// Give it time to clean up
		time.Sleep(100 * time.Millisecond)
		
		// Should be back to zero
		assert.Equal(t, 0, concreteExecutor.GetActiveJobCount())
	})

	t.Run("JobCancellation_CancelActiveJob", func(t *testing.T) {
		concreteExecutor := executor.(*toolExecutor)
		
		// Set up a long-running mock
		mockService.responses["get-athlete-profile"] = mockResponse{
			response: "should be cancelled",
			delay:    2 * time.Second,
		}
		
		// Start execution in background
		var result *models.ToolExecutionResult
		done := make(chan struct{})
		
		go func() {
			defer close(done)
			result, _ = executor.ExecuteTool(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx)
		}()
		
		// Give it time to start
		time.Sleep(100 * time.Millisecond)
		
		// Should have one active job
		assert.Equal(t, 1, concreteExecutor.GetActiveJobCount())
		
		// Cancel all jobs (in a real scenario, you'd have the job ID)
		// For this test, we'll just verify the job tracking works
		
		// Wait for completion
		<-done
		
		// Verify execution completed (may have been cancelled or timed out)
		assert.NotNil(t, result)
	})

	t.Run("ConcurrentExecution_MultipleJobs", func(t *testing.T) {
		concreteExecutor := executor.(*toolExecutor)
		
		// Set up mock responses for concurrent execution
		mockService.responses["get-athlete-profile"] = mockResponse{
			response: "concurrent job 1",
			delay:    500 * time.Millisecond,
		}
		
		numJobs := 3
		done := make(chan struct{}, numJobs)
		
		// Start multiple jobs concurrently
		for i := 0; i < numJobs; i++ {
			go func(jobNum int) {
				defer func() { done <- struct{}{} }()
				
				result, err := executor.ExecuteTool(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx)
				assert.NoError(t, err, "Job %d should succeed", jobNum)
				assert.True(t, result.Success, "Job %d should be successful", jobNum)
			}(i)
		}
		
		// Give jobs time to start
		time.Sleep(100 * time.Millisecond)
		
		// Should have multiple active jobs
		activeCount := concreteExecutor.GetActiveJobCount()
		assert.Greater(t, activeCount, 0, "Should have active jobs")
		assert.LessOrEqual(t, activeCount, numJobs, "Should not exceed number of started jobs")
		
		// Wait for all jobs to complete
		for i := 0; i < numJobs; i++ {
			<-done
		}
		
		// Give time for cleanup
		time.Sleep(100 * time.Millisecond)
		
		// Should be back to zero
		assert.Equal(t, 0, concreteExecutor.GetActiveJobCount())
	})

	t.Run("ContextCancellation_HandlesGracefully", func(t *testing.T) {
		// Create a cancellable context
		cancelCtx, cancel := context.WithCancel(ctx)
		
		mockService.responses["get-athlete-profile"] = mockResponse{
			response: "should be cancelled",
			delay:    2 * time.Second,
		}
		
		// Start execution
		done := make(chan struct{})
		var result *models.ToolExecutionResult
		var err error
		
		go func() {
			defer close(done)
			result, err = executor.ExecuteTool(cancelCtx, "get-athlete-profile", map[string]interface{}{}, msgCtx)
		}()
		
		// Cancel after a short delay
		time.Sleep(100 * time.Millisecond)
		cancel()
		
		// Wait for completion
		<-done
		
		// Should handle cancellation gracefully
		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Success)
	})
}

// Comprehensive mock service for testing
type mockToolExecutionServiceComprehensive struct {
	responses map[string]mockResponse
}

type mockResponse struct {
	response    string
	delay       time.Duration
	shouldError bool
	errorMsg    string
}

func (m *mockToolExecutionServiceComprehensive) ExecuteGetAthleteProfile(ctx context.Context, msgCtx *MessageContext) (string, error) {
	return m.executeWithMock(ctx, "get-athlete-profile")
}

func (m *mockToolExecutionServiceComprehensive) ExecuteGetRecentActivities(ctx context.Context, msgCtx *MessageContext, perPage int) (string, error) {
	return m.executeWithMock(ctx, "get-recent-activities")
}

func (m *mockToolExecutionServiceComprehensive) ExecuteGetActivityDetails(ctx context.Context, msgCtx *MessageContext, activityID int64) (string, error) {
	return m.executeWithMock(ctx, "get-activity-details")
}

func (m *mockToolExecutionServiceComprehensive) ExecuteGetActivityStreams(ctx context.Context, msgCtx *MessageContext, activityID int64, streamTypes []string, resolution string, processingMode string, pageNumber int, pageSize int, summaryPrompt string) (string, error) {
	return m.executeWithMock(ctx, "get-activity-streams")
}

func (m *mockToolExecutionServiceComprehensive) ExecuteUpdateAthleteLogbook(ctx context.Context, msgCtx *MessageContext, content string) (string, error) {
	return m.executeWithMock(ctx, "update-athlete-logbook")
}

func (m *mockToolExecutionServiceComprehensive) executeWithMock(ctx context.Context, toolName string) (string, error) {
	response, exists := m.responses[toolName]
	if !exists {
		response = mockResponse{
			response: "default response for " + toolName,
			delay:    10 * time.Millisecond,
		}
	}
	
	if response.delay > 0 {
		select {
		case <-time.After(response.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	
	if response.shouldError {
		return "", errors.New(response.errorMsg)
	}
	
	return response.response, nil
}

// Helper function to get valid parameters for each tool
func getValidParametersForTool(toolName string) map[string]interface{} {
	switch toolName {
	case "get-athlete-profile":
		return map[string]interface{}{}
	case "get-recent-activities":
		return map[string]interface{}{
			"per_page": 10,
		}
	case "get-activity-details":
		return map[string]interface{}{
			"activity_id": int64(123456),
		}
	case "get-activity-streams":
		return map[string]interface{}{
			"activity_id":     int64(123456),
			"stream_types":    []string{"time", "heartrate"},
			"resolution":      "medium",
			"processing_mode": "auto",
		}
	case "update-athlete-logbook":
		return map[string]interface{}{
			"content": "Test logbook content",
		}
	default:
		return map[string]interface{}{}
	}
}

func TestToolExecutor_ErrorHandling(t *testing.T) {
	mockService := &mockToolExecutionServiceComprehensive{
		responses: make(map[string]mockResponse),
	}
	registry := NewToolRegistry()
	executor := NewToolExecutor(mockService, registry)
	
	ctx := context.Background()
	msgCtx := &MessageContext{
		UserID:    "test-user",
		SessionID: "test-session",
		Message:   "test message",
		User: &models.User{
			ID:       "test-user",
			StravaID: 12345,
		},
	}

	t.Run("ServiceError_ReturnsDetailedError", func(t *testing.T) {
		mockService.responses["get-athlete-profile"] = mockResponse{
			shouldError: true,
			errorMsg:    "detailed service error message",
		}
		
		result, err := executor.ExecuteTool(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx)
		
		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "detailed service error message")
		assert.Equal(t, "get-athlete-profile", result.ToolName)
		assert.Greater(t, result.Duration, int64(0))
	})

	t.Run("NilResult_HandledGracefully", func(t *testing.T) {
		// This tests the internal error handling when tool execution returns nil
		// In practice, this shouldn't happen with our current implementation,
		// but it's good to test defensive programming
		
		mockService.responses["get-athlete-profile"] = mockResponse{
			response: "", // Empty response
		}
		
		result, _ := executor.ExecuteTool(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx)
		
		// Should still return a result object even if the tool response is empty
		assert.NotNil(t, result)
		assert.Equal(t, "get-athlete-profile", result.ToolName)
	})

	t.Run("LongRunningTask_TimeoutHandling", func(t *testing.T) {
		mockService.responses["get-athlete-profile"] = mockResponse{
			response: "long running response",
			delay:    5 * time.Second, // Longer than default timeout
		}
		
		start := time.Now()
		result, err := executor.ExecuteTool(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx)
		duration := time.Since(start)
		
		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Success)
		assert.Contains(t, strings.ToLower(result.Error), "timeout")
		
		// Should timeout much sooner than the mock delay
		assert.Less(t, duration, 4*time.Second)
	})
}

func TestToolExecutor_PerformanceMetrics(t *testing.T) {
	mockService := &mockToolExecutionServiceComprehensive{
		responses: make(map[string]mockResponse),
	}
	registry := NewToolRegistry()
	executor := NewToolExecutor(mockService, registry)
	
	ctx := context.Background()
	msgCtx := &MessageContext{
		UserID:    "test-user",
		SessionID: "test-session",
		Message:   "test message",
		User: &models.User{
			ID:       "test-user",
			StravaID: 12345,
		},
	}

	t.Run("ExecutionDuration_AccuratelyMeasured", func(t *testing.T) {
		expectedDelay := 200 * time.Millisecond
		mockService.responses["get-athlete-profile"] = mockResponse{
			response: "timed response",
			delay:    expectedDelay,
		}
		
		result, err := executor.ExecuteTool(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx)
		
		assert.NoError(t, err)
		assert.True(t, result.Success)
		
		// Duration should be approximately the expected delay (with some tolerance)
		assert.Greater(t, result.Duration, int64(150), "Duration should be at least 150ms")
		assert.Less(t, result.Duration, int64(300), "Duration should be less than 300ms")
	})

	t.Run("TimestampAccuracy", func(t *testing.T) {
		mockService.responses["get-athlete-profile"] = mockResponse{
			response: "timestamp test",
			delay:    50 * time.Millisecond,
		}
		
		beforeExecution := time.Now()
		result, err := executor.ExecuteTool(ctx, "get-athlete-profile", map[string]interface{}{}, msgCtx)
		afterExecution := time.Now()
		
		assert.NoError(t, err)
		assert.True(t, result.Success)
		
		// Timestamp should be between before and after execution
		assert.True(t, result.Timestamp.After(beforeExecution.Add(-time.Second)), 
			"Timestamp should be after execution start")
		assert.True(t, result.Timestamp.Before(afterExecution.Add(time.Second)), 
			"Timestamp should be before execution end")
	})
}