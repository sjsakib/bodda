package services

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"

	openai "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/responses"
)

// StreamSummary represents the result of AI-powered stream summarization
type StreamSummary struct {
	ActivityID    int64  `json:"activity_id"`
	SummaryPrompt string `json:"summary_prompt"`
	Summary       string `json:"summary"`
	TokensUsed    int    `json:"tokens_used,omitempty"`
	Model         string `json:"model,omitempty"`
}

// SummaryProcessor interface defines methods for AI-powered stream summarization
type SummaryProcessor interface {
	GenerateSummary(ctx context.Context, data *StravaStreams, activityID int64, prompt string) (*StreamSummary, error)
	PrepareStreamDataForSummarization(data *StravaStreams) (string, error)
}

// summaryProcessor implements the SummaryProcessor interface
type summaryProcessor struct {
	client *openai.Client
	model  string
}

// NewSummaryProcessor creates a new summary processor instance using the official OpenAI SDK
func NewSummaryProcessor(client *openai.Client) SummaryProcessor {
	return &summaryProcessor{
		client: client,
		model:  "gpt-4o-mini", // Use smaller, faster model for summarization
	}
}

// GenerateSummary generates an AI-powered summary of stream data using custom prompt
func (sp *summaryProcessor) GenerateSummary(ctx context.Context, data *StravaStreams, activityID int64, prompt string) (*StreamSummary, error) {
	if data == nil {
		return nil, fmt.Errorf("stream data is nil")
	}

	if prompt == "" {
		return nil, fmt.Errorf("summary prompt is required")
	}

	// Prepare stream data for AI processing
	streamDataText, err := sp.PrepareStreamDataForSummarization(data)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare stream data: %w", err)
	}

	// Build the complete prompt for AI summarization
	systemPrompt := `You are an expert sports data analyst specializing in endurance training. You will receive time-series data from a training activity and a specific prompt about what to analyze.

Your task is to analyze the provided stream data and respond to the user's specific request. Focus on the data patterns, trends, and insights that directly address their question.

Provide your analysis in a clear, structured format using markdown. Be factual and data-driven in your response.`

	userPrompt := fmt.Sprintf(`Here is the stream data from activity %d:

%s

User's request: %s

Please analyze this data and provide insights based on the user's specific request.`, activityID, streamDataText, prompt)

	// Create responses API request
	params := responses.ResponseNewParams{
		Model: responses.ChatModelGPT5Nano,
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: []responses.ResponseInputItemUnionParam{
				responses.ResponseInputItemParamOfMessage(systemPrompt, responses.EasyInputMessageRoleSystem),
				responses.ResponseInputItemParamOfMessage(userPrompt, responses.EasyInputMessageRoleUser),
			},
		},
	}

	slog.InfoContext(ctx, "Invoking LLM for stream summary", "message_len", len(userPrompt))

	// Call OpenAI Responses API with streaming
	stream := sp.client.Responses.NewStreaming(ctx, params)
	defer stream.Close()

	var summaryContent strings.Builder

	// Process streaming response
	for stream.Next() {
		event := stream.Current()
		
		switch event.Type {
		case "response.output_text.delta":
			textEvent := event.AsResponseOutputTextDelta()
			if textEvent.Delta != "" {
				summaryContent.WriteString(textEvent.Delta)
			}
		case "response.completed":
			// Stream completed successfully
			break
		}
	}

	if err := stream.Err(); err != nil {
		log.Printf("OpenAI API streaming error during summarization: %v", err)
		return nil, fmt.Errorf("failed to generate AI summary: %w", err)
	}

	summary := &StreamSummary{
		ActivityID:    activityID,
		SummaryPrompt: prompt,
		Summary:       summaryContent.String(),
		Model:         sp.model,
	}

	log.Printf("Generated AI summary for activity %d", activityID)

	return summary, nil
}

// PrepareStreamDataForSummarization converts stream data to a text format suitable for AI processing
func (sp *summaryProcessor) PrepareStreamDataForSummarization(data *StravaStreams) (string, error) {
	if data == nil {
		return "", fmt.Errorf("stream data is nil")
	}

	var builder strings.Builder

	// Add metadata about the stream data
	builder.WriteString("STREAM DATA SUMMARY:\n\n")

	// Count total data points and duration
	totalPoints := sp.countDataPoints(data)
	builder.WriteString(fmt.Sprintf("Total data points: %d\n", totalPoints))

	if len(data.Time) > 0 {
		duration := data.Time[len(data.Time)-1] - data.Time[0]
		builder.WriteString(fmt.Sprintf("Duration: %d seconds\n", duration))
	}

	// List available stream types
	streamTypes := sp.getAvailableStreamTypes(data)
	builder.WriteString(fmt.Sprintf("Available streams: %s\n\n", strings.Join(streamTypes, ", ")))

	// Include sample data for analysis
	builder.WriteString("SAMPLE DATA (first 10 points):\n\n")
	
	maxLen := 10
	if len(data.Time) > 0 && len(data.Time) < maxLen {
		maxLen = len(data.Time)
	}

	for i := 0; i < maxLen; i++ {
		builder.WriteString(fmt.Sprintf("Point %d: ", i+1))
		
		if i < len(data.Time) {
			builder.WriteString(fmt.Sprintf("Time=%ds ", data.Time[i]))
		}
		if i < len(data.Distance) {
			builder.WriteString(fmt.Sprintf("Distance=%.1fm ", data.Distance[i]))
		}
		if i < len(data.Heartrate) {
			builder.WriteString(fmt.Sprintf("HR=%dbpm ", data.Heartrate[i]))
		}
		if i < len(data.Watts) {
			builder.WriteString(fmt.Sprintf("Power=%dw ", data.Watts[i]))
		}
		
		builder.WriteString("\n")
	}

	return builder.String(), nil
}

// countDataPoints counts the total number of data points across all streams
func (sp *summaryProcessor) countDataPoints(data *StravaStreams) int {
	total := 0
	total += len(data.Time)
	total += len(data.Distance)
	total += len(data.Heartrate)
	total += len(data.Watts)
	total += len(data.Cadence)
	total += len(data.Altitude)
	return total
}

// getAvailableStreamTypes returns a list of available stream types in the data
func (sp *summaryProcessor) getAvailableStreamTypes(data *StravaStreams) []string {
	var types []string
	
	if len(data.Time) > 0 {
		types = append(types, "time")
	}
	if len(data.Distance) > 0 {
		types = append(types, "distance")
	}
	if len(data.Heartrate) > 0 {
		types = append(types, "heartrate")
	}
	if len(data.Watts) > 0 {
		types = append(types, "watts")
	}
	if len(data.Cadence) > 0 {
		types = append(types, "cadence")
	}
	if len(data.Altitude) > 0 {
		types = append(types, "altitude")
	}
	
	return types
}