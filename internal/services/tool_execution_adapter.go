package services

import (
	"context"
)

// toolExecutionAdapter adapts the existing AI service to provide tool execution capabilities
type toolExecutionAdapter struct {
	aiService AIService
}

// NewToolExecutionAdapter creates a new tool execution adapter
func NewToolExecutionAdapter(aiService AIService) ToolExecutionService {
	return &toolExecutionAdapter{
		aiService: aiService,
	}
}

// ExecuteGetAthleteProfile executes the get-athlete-profile tool
func (tea *toolExecutionAdapter) ExecuteGetAthleteProfile(ctx context.Context, msgCtx *MessageContext) (string, error) {
	return tea.aiService.ExecuteGetAthleteProfile(ctx, msgCtx)
}

// ExecuteGetRecentActivities executes the get-recent-activities tool
func (tea *toolExecutionAdapter) ExecuteGetRecentActivities(ctx context.Context, msgCtx *MessageContext, perPage int) (string, error) {
	return tea.aiService.ExecuteGetRecentActivities(ctx, msgCtx, perPage)
}

// ExecuteGetActivityDetails executes the get-activity-details tool
func (tea *toolExecutionAdapter) ExecuteGetActivityDetails(ctx context.Context, msgCtx *MessageContext, activityID int64) (string, error) {
	return tea.aiService.ExecuteGetActivityDetails(ctx, msgCtx, activityID)
}

// ExecuteGetActivityStreams executes the get-activity-streams tool
func (tea *toolExecutionAdapter) ExecuteGetActivityStreams(ctx context.Context, msgCtx *MessageContext, activityID int64, streamTypes []string, resolution string, processingMode string, pageNumber int, pageSize int, summaryPrompt string) (string, error) {
	return tea.aiService.ExecuteGetActivityStreams(ctx, msgCtx, activityID, streamTypes, resolution, processingMode, pageNumber, pageSize, summaryPrompt)
}

// ExecuteUpdateAthleteLogbook executes the update-athlete-logbook tool
func (tea *toolExecutionAdapter) ExecuteUpdateAthleteLogbook(ctx context.Context, msgCtx *MessageContext, content string) (string, error) {
	return tea.aiService.ExecuteUpdateAthleteLogbook(ctx, msgCtx, content)
}