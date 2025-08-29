package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"bodda/internal/models"

	"github.com/sashabaranov/go-openai"
)

// ToolExecutionService defines the interface for executing individual tools
type ToolExecutionService interface {
	ExecuteGetAthleteProfile(ctx context.Context, msgCtx *MessageContext) (string, error)
	ExecuteGetRecentActivities(ctx context.Context, msgCtx *MessageContext, perPage int) (string, error)
	ExecuteGetActivityDetails(ctx context.Context, msgCtx *MessageContext, activityID int64) (string, error)
	ExecuteGetActivityStreams(ctx context.Context, msgCtx *MessageContext, activityID int64, streamTypes []string, resolution string, processingMode string, pageNumber int, pageSize int, summaryPrompt string) (string, error)
	ExecuteUpdateAthleteLogbook(ctx context.Context, msgCtx *MessageContext, content string) (string, error)
}

// toolExecutor implements the ToolExecutor interface with enhanced timeout and streaming support
type toolExecutor struct {
	toolService    ToolExecutionService
	registry       ToolRegistry
	defaultTimeout time.Duration
	maxTimeout     time.Duration
	mu             sync.RWMutex
	activeJobs     map[string]context.CancelFunc
}

// NewToolExecutor creates a new tool executor with enhanced timeout and streaming support
func NewToolExecutor(toolService ToolExecutionService, registry ToolRegistry) ToolExecutor {
	return &toolExecutor{
		toolService:    toolService,
		registry:       registry,
		defaultTimeout: 30 * time.Second,
		maxTimeout:     300 * time.Second,
		activeJobs:     make(map[string]context.CancelFunc),
	}
}

// NewToolExecutorWithConfig creates a new tool executor with custom timeout configuration
func NewToolExecutorWithConfig(toolService ToolExecutionService, registry ToolRegistry, defaultTimeout, maxTimeout time.Duration) ToolExecutor {
	return &toolExecutor{
		toolService:    toolService,
		registry:       registry,
		defaultTimeout: defaultTimeout,
		maxTimeout:     maxTimeout,
		activeJobs:     make(map[string]context.CancelFunc),
	}
}

// ExecuteTool executes a tool with the given parameters and context
func (te *toolExecutor) ExecuteTool(ctx context.Context, toolName string, parameters map[string]interface{}, msgCtx *MessageContext) (*models.ToolExecutionResult, error) {
	return te.ExecuteToolWithOptions(ctx, toolName, parameters, msgCtx, nil)
}

// ExecuteToolWithOptions executes a tool with additional execution options including timeout and streaming support
func (te *toolExecutor) ExecuteToolWithOptions(ctx context.Context, toolName string, parameters map[string]interface{}, msgCtx *MessageContext, options *models.ExecutionOptions) (*models.ToolExecutionResult, error) {
	startTime := time.Now()
	jobID := fmt.Sprintf("job_%d", time.Now().UnixNano())
	
	result := &models.ToolExecutionResult{
		ToolName:  toolName,
		Timestamp: startTime,
	}

	// Validate tool exists
	if !te.registry.IsToolAvailable(toolName) {
		result.Success = false
		result.Error = fmt.Sprintf("tool '%s' not found", toolName)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, fmt.Errorf("tool '%s' not found", toolName)
	}

	// Validate parameters
	if err := te.registry.ValidateToolCall(toolName, parameters); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("parameter validation failed: %v", err)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}

	// Set up timeout context with enhanced timeout handling
	execCtx, cancel := te.setupTimeoutContext(ctx, options, jobID)
	defer func() {
		cancel()
		te.cleanupJob(jobID)
	}()

	// Log execution start
	log.Printf("Starting tool execution: tool=%s, job_id=%s, timeout=%v", toolName, jobID, te.getTimeoutDuration(options))

	// Handle streaming vs buffered execution
	if options != nil && options.Streaming {
		return te.executeToolStreaming(execCtx, toolName, parameters, msgCtx, result, jobID)
	} else {
		return te.executeToolBuffered(execCtx, toolName, parameters, msgCtx, result, jobID)
	}
}

// setupTimeoutContext creates a timeout context with job tracking for graceful cleanup
func (te *toolExecutor) setupTimeoutContext(ctx context.Context, options *models.ExecutionOptions, jobID string) (context.Context, context.CancelFunc) {
	timeout := te.getTimeoutDuration(options)
	
	// Enforce maximum timeout limit
	if timeout > te.maxTimeout {
		timeout = te.maxTimeout
		log.Printf("Timeout capped at maximum: %v for job %s", te.maxTimeout, jobID)
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	
	// Track active job for cleanup
	te.mu.Lock()
	te.activeJobs[jobID] = cancel
	te.mu.Unlock()

	return execCtx, cancel
}

// getTimeoutDuration determines the timeout duration from options or defaults
func (te *toolExecutor) getTimeoutDuration(options *models.ExecutionOptions) time.Duration {
	if options != nil && options.Timeout > 0 {
		return time.Duration(options.Timeout) * time.Second
	}
	return te.defaultTimeout
}

// cleanupJob removes a job from active tracking
func (te *toolExecutor) cleanupJob(jobID string) {
	te.mu.Lock()
	delete(te.activeJobs, jobID)
	te.mu.Unlock()
}

// executeToolStreaming executes a tool with streaming output support
func (te *toolExecutor) executeToolStreaming(ctx context.Context, toolName string, parameters map[string]interface{}, msgCtx *MessageContext, result *models.ToolExecutionResult, jobID string) (*models.ToolExecutionResult, error) {
	log.Printf("Executing tool in streaming mode: tool=%s, job_id=%s", toolName, jobID)
	
	// For streaming mode, we'll execute the tool and collect output progressively
	// This is a simplified implementation - in a real streaming scenario, 
	// you might want to return a channel or use a callback mechanism
	
	// Create a done channel to signal completion
	done := make(chan struct{})
	var toolResult *ToolResult
	var execErr error
	
	// Execute tool in a goroutine to support streaming
	go func() {
		defer close(done)
		toolResult, execErr = te.executeToolInternal(ctx, toolName, parameters, msgCtx, jobID)
	}()
	
	// Wait for completion with timeout handling
	select {
	case <-ctx.Done():
		result.Success = false
		result.Error = "execution timed out"
		result.Duration = time.Since(result.Timestamp).Milliseconds()
		log.Printf("Tool execution timed out: tool=%s, job_id=%s, duration=%dms", toolName, jobID, result.Duration)
		return result, ctx.Err()
		
	case <-done:
		if execErr != nil {
			result.Success = false
			result.Error = execErr.Error()
			result.Duration = time.Since(result.Timestamp).Milliseconds()
			log.Printf("Tool execution failed (streaming): tool=%s, job_id=%s, duration=%dms, error=%v", 
				toolName, jobID, result.Duration, execErr)
			return result, execErr
		}
		
		if toolResult == nil {
			result.Success = false
			result.Error = "no result received from tool execution"
			result.Duration = time.Since(result.Timestamp).Milliseconds()
			log.Printf("Tool execution failed (streaming): tool=%s, job_id=%s, duration=%dms, error=nil result", 
				toolName, jobID, result.Duration)
			return result, fmt.Errorf("no result received from tool execution")
		}
		
		// Process successful result
		result.Success = toolResult.Error == ""
		result.Data = toolResult.Content
		if toolResult.Error != "" {
			result.Error = toolResult.Error
		}
		result.Duration = time.Since(result.Timestamp).Milliseconds()
		
		log.Printf("Tool execution completed (streaming): tool=%s, job_id=%s, duration=%dms, success=%v", 
			toolName, jobID, result.Duration, result.Success)
		return result, nil
	}
}

// executeToolBuffered executes a tool with buffered output (traditional mode)
func (te *toolExecutor) executeToolBuffered(ctx context.Context, toolName string, parameters map[string]interface{}, msgCtx *MessageContext, result *models.ToolExecutionResult, jobID string) (*models.ToolExecutionResult, error) {
	log.Printf("Executing tool in buffered mode: tool=%s, job_id=%s", toolName, jobID)
	
	// Execute with timeout monitoring
	done := make(chan struct{})
	var toolResult *ToolResult
	var execErr error
	
	go func() {
		defer close(done)
		toolResult, execErr = te.executeToolInternal(ctx, toolName, parameters, msgCtx, jobID)
	}()
	
	select {
	case <-ctx.Done():
		result.Success = false
		result.Error = "execution timed out"
		result.Duration = time.Since(result.Timestamp).Milliseconds()
		log.Printf("Tool execution timed out: tool=%s, job_id=%s, duration=%dms", toolName, jobID, result.Duration)
		return result, ctx.Err()
		
	case <-done:
		if execErr != nil {
			result.Success = false
			result.Error = execErr.Error()
			result.Duration = time.Since(result.Timestamp).Milliseconds()
			log.Printf("Tool execution failed (buffered): tool=%s, job_id=%s, duration=%dms, error=%v", 
				toolName, jobID, result.Duration, execErr)
			return result, execErr
		}
		
		// Process successful result
		result.Success = toolResult.Error == ""
		result.Data = toolResult.Content
		if toolResult.Error != "" {
			result.Error = toolResult.Error
		}
		result.Duration = time.Since(result.Timestamp).Milliseconds()
		
		log.Printf("Tool execution completed (buffered): tool=%s, job_id=%s, duration=%dms, success=%v", 
			toolName, jobID, result.Duration, result.Success)
		return result, nil
	}
}

// executeToolInternal performs the actual tool execution
func (te *toolExecutor) executeToolInternal(ctx context.Context, toolName string, parameters map[string]interface{}, msgCtx *MessageContext, jobID string) (*ToolResult, error) {
	// Create a mock tool call to reuse existing execution logic
	parametersJSON, err := json.Marshal(parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	toolCall := openai.ToolCall{
		ID:   fmt.Sprintf("call_%s", jobID),
		Type: openai.ToolTypeFunction,
		Function: openai.FunctionCall{
			Name:      toolName,
			Arguments: string(parametersJSON),
		},
	}

	// Execute the tool using existing AI service logic
	toolResults, err := te.executeToolCall(ctx, msgCtx, toolCall)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	// Process the result
	if len(toolResults) > 0 {
		return &toolResults[0], nil
	}
	
	return nil, fmt.Errorf("no result returned from tool execution")
}

// CancelJob cancels an active tool execution job
func (te *toolExecutor) CancelJob(jobID string) bool {
	te.mu.RLock()
	cancel, exists := te.activeJobs[jobID]
	te.mu.RUnlock()
	
	if exists {
		cancel()
		log.Printf("Cancelled tool execution job: %s", jobID)
		return true
	}
	
	return false
}

// GetActiveJobCount returns the number of currently active jobs
func (te *toolExecutor) GetActiveJobCount() int {
	te.mu.RLock()
	defer te.mu.RUnlock()
	return len(te.activeJobs)
}

// executeToolCall executes a single tool call using the existing AI service logic
func (te *toolExecutor) executeToolCall(ctx context.Context, msgCtx *MessageContext, toolCall openai.ToolCall) ([]ToolResult, error) {
	// This method reuses the existing tool execution logic from the AI service
	// We need to access the private executeTools method, so we'll implement the logic here
	
	var results []ToolResult
	result := ToolResult{
		ToolCallID: toolCall.ID,
	}

	switch toolCall.Function.Name {
	case "get-athlete-profile":
		content, err := te.toolService.ExecuteGetAthleteProfile(ctx, msgCtx)
		if err != nil {
			result.Error = err.Error()
			result.Content = fmt.Sprintf("Error getting athlete profile: %v", err)
		} else {
			result.Content = content
		}

	case "get-recent-activities":
		var args struct {
			PerPage int `json:"per_page"`
		}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			result.Error = err.Error()
			result.Content = fmt.Sprintf("Error parsing arguments: %v", err)
		} else {
			if args.PerPage == 0 {
				args.PerPage = 30
			}
			content, err := te.toolService.ExecuteGetRecentActivities(ctx, msgCtx, args.PerPage)
			if err != nil {
				result.Error = err.Error()
				result.Content = fmt.Sprintf("Error getting recent activities: %v", err)
			} else {
				result.Content = content
			}
		}

	case "get-activity-details":
		var args struct {
			ActivityID int64 `json:"activity_id"`
		}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			result.Error = err.Error()
			result.Content = fmt.Sprintf("Error parsing arguments: %v", err)
		} else {
			content, err := te.toolService.ExecuteGetActivityDetails(ctx, msgCtx, args.ActivityID)
			if err != nil {
				result.Error = err.Error()
				result.Content = fmt.Sprintf("Error getting activity details: %v", err)
			} else {
				result.Content = content
			}
		}

	case "get-activity-streams":
		var args struct {
			ActivityID     int64    `json:"activity_id"`
			StreamTypes    []string `json:"stream_types"`
			Resolution     string   `json:"resolution"`
			ProcessingMode string   `json:"processing_mode"`
			PageNumber     int      `json:"page_number"`
			PageSize       int      `json:"page_size"`
			SummaryPrompt  string   `json:"summary_prompt"`
		}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			result.Error = err.Error()
			result.Content = fmt.Sprintf("Error parsing arguments: %v", err)
		} else {
			// Set defaults
			if len(args.StreamTypes) == 0 {
				args.StreamTypes = []string{"time", "distance", "heartrate", "watts"}
			}
			if args.Resolution == "" {
				args.Resolution = "medium"
			}
			if args.ProcessingMode == "" {
				args.ProcessingMode = "auto"
			}
			if args.PageNumber == 0 {
				args.PageNumber = 1
			}
			if args.PageSize == 0 {
				args.PageSize = 1000
			}

			content, err := te.toolService.ExecuteGetActivityStreams(ctx, msgCtx, args.ActivityID, args.StreamTypes, args.Resolution, args.ProcessingMode, args.PageNumber, args.PageSize, args.SummaryPrompt)
			if err != nil {
				result.Error = err.Error()
				result.Content = fmt.Sprintf("Error getting activity streams: %v", err)
			} else {
				result.Content = content
			}
		}

	case "update-athlete-logbook":
		var args struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			result.Error = err.Error()
			result.Content = fmt.Sprintf("Error parsing arguments: %v", err)
		} else {
			content, err := te.toolService.ExecuteUpdateAthleteLogbook(ctx, msgCtx, args.Content)
			if err != nil {
				result.Error = err.Error()
				result.Content = fmt.Sprintf("Error updating athlete logbook: %v", err)
			} else {
				result.Content = content
			}
		}

	default:
		result.Error = fmt.Sprintf("unknown tool: %s", toolCall.Function.Name)
		result.Content = fmt.Sprintf("Tool '%s' is not supported", toolCall.Function.Name)
	}

	results = append(results, result)
	return results, nil
}