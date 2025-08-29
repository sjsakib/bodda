package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/sashabaranov/go-openai"
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

// NewSummaryProcessor creates a new summary processor instance
func NewSummaryProcessor(client *openai.Client) SummaryProcessor {
	return &summaryProcessor{
		client: client,
		model:  openai.GPT4oMini, // Use smaller, faster model for summarization
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

	// Create chat completion request
	req := openai.ChatCompletionRequest{
		Model: sp.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Temperature: 0.3, // Lower temperature for more consistent, factual responses
	}

	// Call OpenAI API
	resp, err := sp.client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("OpenAI API error during stream summarization: %v", err)
		return nil, fmt.Errorf("failed to generate AI summary: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI summarization")
	}

	summary := &StreamSummary{
		ActivityID:    activityID,
		SummaryPrompt: prompt,
		Summary:       resp.Choices[0].Message.Content,
		TokensUsed:    resp.Usage.TotalTokens,
		Model:         sp.model,
	}

	log.Printf("Generated AI summary for activity %d using %d tokens", activityID, resp.Usage.TotalTokens)

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

	// Add statistical summaries for each stream type
	if len(data.Heartrate) > 0 {
		builder.WriteString("HEART RATE DATA:\n")
		sp.addStreamStatistics(&builder, "Heart Rate", data.Heartrate, "bpm")
		builder.WriteString("\n")
	}

	if len(data.Watts) > 0 {
		builder.WriteString("POWER DATA:\n")
		sp.addStreamStatistics(&builder, "Power", data.Watts, "watts")
		builder.WriteString("\n")
	}

	if len(data.VelocitySmooth) > 0 {
		builder.WriteString("SPEED DATA:\n")
		sp.addSpeedStatistics(&builder, data.VelocitySmooth)
		builder.WriteString("\n")
	}

	if len(data.Cadence) > 0 {
		builder.WriteString("CADENCE DATA:\n")
		sp.addStreamStatistics(&builder, "Cadence", data.Cadence, "rpm")
		builder.WriteString("\n")
	}

	if len(data.Altitude) > 0 {
		builder.WriteString("ELEVATION DATA:\n")
		sp.addElevationStatistics(&builder, data.Altitude)
		builder.WriteString("\n")
	}

	if len(data.Temp) > 0 {
		builder.WriteString("TEMPERATURE DATA:\n")
		sp.addStreamStatistics(&builder, "Temperature", data.Temp, "°C")
		builder.WriteString("\n")
	}

	if len(data.GradeSmooth) > 0 {
		builder.WriteString("GRADE DATA:\n")
		sp.addGradeStatistics(&builder, data.GradeSmooth)
		builder.WriteString("\n")
	}

	if len(data.Moving) > 0 {
		builder.WriteString("MOVING TIME DATA:\n")
		sp.addMovingTimeStatistics(&builder, data.Moving)
		builder.WriteString("\n")
	}

	if len(data.Distance) > 0 {
		builder.WriteString("DISTANCE DATA:\n")
		sp.addDistanceStatistics(&builder, data.Distance)
		builder.WriteString("\n")
	}

	// Add sample data points for context (every 10% of the activity)
	if len(data.Time) > 0 {
		builder.WriteString("SAMPLE DATA POINTS:\n")
		sp.addSampleDataPoints(&builder, data)
		builder.WriteString("\n")
	}

	return builder.String(), nil
}

// Helper methods for data preparation

func (sp *summaryProcessor) countDataPoints(data *StravaStreams) int {
	maxPoints := 0
	
	if len(data.Time) > maxPoints {
		maxPoints = len(data.Time)
	}
	if len(data.Distance) > maxPoints {
		maxPoints = len(data.Distance)
	}
	if len(data.Heartrate) > maxPoints {
		maxPoints = len(data.Heartrate)
	}
	if len(data.Watts) > maxPoints {
		maxPoints = len(data.Watts)
	}
	if len(data.Cadence) > maxPoints {
		maxPoints = len(data.Cadence)
	}
	if len(data.Altitude) > maxPoints {
		maxPoints = len(data.Altitude)
	}
	if len(data.VelocitySmooth) > maxPoints {
		maxPoints = len(data.VelocitySmooth)
	}
	if len(data.Temp) > maxPoints {
		maxPoints = len(data.Temp)
	}
	if len(data.GradeSmooth) > maxPoints {
		maxPoints = len(data.GradeSmooth)
	}
	if len(data.Moving) > maxPoints {
		maxPoints = len(data.Moving)
	}
	if len(data.Latlng) > maxPoints {
		maxPoints = len(data.Latlng)
	}
	
	return maxPoints
}

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
	if len(data.VelocitySmooth) > 0 {
		types = append(types, "velocity_smooth")
	}
	if len(data.Temp) > 0 {
		types = append(types, "temp")
	}
	if len(data.GradeSmooth) > 0 {
		types = append(types, "grade_smooth")
	}
	if len(data.Moving) > 0 {
		types = append(types, "moving")
	}
	if len(data.Latlng) > 0 {
		types = append(types, "latlng")
	}
	
	return types
}

func (sp *summaryProcessor) addStreamStatistics(builder *strings.Builder, name string, data []int, unit string) {
	if len(data) == 0 {
		return
	}

	min, max, avg := sp.calculateIntStats(data)
	builder.WriteString(fmt.Sprintf("- %s: %d-%d %s (avg: %.1f %s, %d points)\n", 
		name, min, max, unit, avg, unit, len(data)))
}

func (sp *summaryProcessor) addSpeedStatistics(builder *strings.Builder, data []float64) {
	if len(data) == 0 {
		return
	}

	min, max, avg := sp.calculateFloatStats(data)
	// Convert m/s to km/h for display
	minKmh := min * 3.6
	maxKmh := max * 3.6
	avgKmh := avg * 3.6
	
	builder.WriteString(fmt.Sprintf("- Speed: %.1f-%.1f km/h (avg: %.1f km/h, %d points)\n", 
		minKmh, maxKmh, avgKmh, len(data)))
}

func (sp *summaryProcessor) addElevationStatistics(builder *strings.Builder, data []float64) {
	if len(data) == 0 {
		return
	}

	min, max, avg := sp.calculateFloatStats(data)
	
	// Calculate elevation gain/loss
	elevationGain := 0.0
	elevationLoss := 0.0
	for i := 1; i < len(data); i++ {
		diff := data[i] - data[i-1]
		if diff > 0 {
			elevationGain += diff
		} else {
			elevationLoss += -diff
		}
	}
	
	builder.WriteString(fmt.Sprintf("- Elevation: %.1f-%.1fm (avg: %.1fm, gain: %.1fm, loss: %.1fm, %d points)\n", 
		min, max, avg, elevationGain, elevationLoss, len(data)))
}

func (sp *summaryProcessor) addGradeStatistics(builder *strings.Builder, data []float64) {
	if len(data) == 0 {
		return
	}

	min, max, avg := sp.calculateFloatStats(data)
	builder.WriteString(fmt.Sprintf("- Grade: %.1f%%-%.1f%% (avg: %.1f%%, %d points)\n", 
		min*100, max*100, avg*100, len(data)))
}

func (sp *summaryProcessor) addMovingTimeStatistics(builder *strings.Builder, data []bool) {
	if len(data) == 0 {
		return
	}

	movingCount := 0
	for _, moving := range data {
		if moving {
			movingCount++
		}
	}
	
	movingPercent := float64(movingCount) / float64(len(data)) * 100
	builder.WriteString(fmt.Sprintf("- Moving time: %d/%d points (%.1f%% moving)\n", 
		movingCount, len(data), movingPercent))
}

func (sp *summaryProcessor) addDistanceStatistics(builder *strings.Builder, data []float64) {
	if len(data) == 0 {
		return
	}

	totalDistance := data[len(data)-1] - data[0]
	builder.WriteString(fmt.Sprintf("- Distance: %.2fkm total (%d points)\n", 
		totalDistance/1000, len(data)))
}

func (sp *summaryProcessor) addSampleDataPoints(builder *strings.Builder, data *StravaStreams) {
	if len(data.Time) == 0 {
		return
	}

	// Sample at 0%, 25%, 50%, 75%, 100% of the activity
	sampleIndices := []int{
		0,
		len(data.Time) / 4,
		len(data.Time) / 2,
		len(data.Time) * 3 / 4,
		len(data.Time) - 1,
	}

	for _, idx := range sampleIndices {
		if idx >= len(data.Time) {
			continue
		}

		timeOffset := data.Time[idx]
		builder.WriteString(fmt.Sprintf("Time %ds: ", timeOffset))

		values := []string{}
		
		if idx < len(data.Heartrate) && data.Heartrate[idx] > 0 {
			values = append(values, fmt.Sprintf("HR %d", data.Heartrate[idx]))
		}
		if idx < len(data.Watts) && data.Watts[idx] > 0 {
			values = append(values, fmt.Sprintf("Power %dW", data.Watts[idx]))
		}
		if idx < len(data.VelocitySmooth) && data.VelocitySmooth[idx] > 0 {
			values = append(values, fmt.Sprintf("Speed %.1fkm/h", data.VelocitySmooth[idx]*3.6))
		}
		if idx < len(data.Cadence) && data.Cadence[idx] > 0 {
			values = append(values, fmt.Sprintf("Cadence %d", data.Cadence[idx]))
		}
		if idx < len(data.Altitude) {
			values = append(values, fmt.Sprintf("Alt %.1fm", data.Altitude[idx]))
		}

		if len(values) > 0 {
			builder.WriteString(strings.Join(values, ", "))
		}
		builder.WriteString("\n")
	}
}

func (sp *summaryProcessor) calculateIntStats(data []int) (int, int, float64) {
	if len(data) == 0 {
		return 0, 0, 0
	}

	min, max := data[0], data[0]
	sum := 0

	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}

	avg := float64(sum) / float64(len(data))
	return min, max, avg
}

func (sp *summaryProcessor) calculateFloatStats(data []float64) (float64, float64, float64) {
	if len(data) == 0 {
		return 0, 0, 0
	}

	min, max := data[0], data[0]
	sum := 0.0

	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}

	avg := sum / float64(len(data))
	return min, max, avg
}