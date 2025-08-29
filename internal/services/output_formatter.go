package services

import (
	"fmt"
	"strings"
	"time"
)

// DerivedFeatures represents comprehensive stream analysis data
type DerivedFeatures struct {
	ActivityID      int64                  `json:"activity_id"`
	Summary         FeatureSummary         `json:"summary"`
	InflectionPoints []InflectionPoint     `json:"inflection_points"`
	Statistics      StreamStatistics       `json:"statistics"`
	Trends          []Trend               `json:"trends"`
	Spikes          []Spike               `json:"spikes"`
	SampleData      []DataPoint           `json:"sample_data"`
	LapAnalysis     *LapAnalysis          `json:"lap_analysis,omitempty"`
}

// FeatureSummary contains high-level activity metrics
type FeatureSummary struct {
	TotalDataPoints   int     `json:"total_data_points"`
	Duration          int     `json:"duration_seconds"`
	TotalDistance     float64 `json:"total_distance,omitempty"`
	ElevationGain     float64 `json:"elevation_gain,omitempty"`
	ElevationLoss     float64 `json:"elevation_loss,omitempty"`
	AvgSpeed          float64 `json:"avg_speed,omitempty"`
	MaxSpeed          float64 `json:"max_speed,omitempty"`
	AvgHeartRate      float64 `json:"avg_heart_rate,omitempty"`
	MaxHeartRate      int     `json:"max_heart_rate,omitempty"`
	AvgPower          float64 `json:"avg_power,omitempty"`
	MaxPower          int     `json:"max_power,omitempty"`
	NormalizedPower   float64 `json:"normalized_power,omitempty"`
	IntensityFactor   float64 `json:"intensity_factor,omitempty"`
	TrainingStressScore float64 `json:"training_stress_score,omitempty"`
	HeartRateDrift    float64 `json:"heart_rate_drift,omitempty"` // bpm per hour
	AvgCadence        float64 `json:"avg_cadence,omitempty"`
	MaxCadence        int     `json:"max_cadence,omitempty"`
	AvgTemperature    float64 `json:"avg_temperature,omitempty"`
	MovingTimePercent float64 `json:"moving_time_percent,omitempty"`
	StreamTypes       []string `json:"stream_types"`
}

// StreamStatistics contains statistical analysis for all stream types
type StreamStatistics struct {
	Time           *MetricStats `json:"time,omitempty"`
	Distance       *MetricStats `json:"distance,omitempty"`
	Altitude       *MetricStats `json:"altitude,omitempty"`
	VelocitySmooth *MetricStats `json:"velocity_smooth,omitempty"`
	HeartRate      *MetricStats `json:"heart_rate,omitempty"`
	Cadence        *MetricStats `json:"cadence,omitempty"`
	Power          *MetricStats `json:"power,omitempty"`
	Temperature    *MetricStats `json:"temperature,omitempty"`
	Grade          *MetricStats `json:"grade_smooth,omitempty"`
	Moving         *BooleanStats `json:"moving,omitempty"`
	LatLng         *LocationStats `json:"latlng,omitempty"`
}

// DataPoint represents a sample data point from the stream
type DataPoint struct {
	TimeOffset int                    `json:"time_offset"`
	Values     map[string]interface{} `json:"values"`
}

// OutputFormatter interface defines methods for formatting Strava tool outputs
type OutputFormatter interface {
	FormatAthleteProfile(profile *StravaAthlete) string
	FormatActivities(activities []*StravaActivity) string
	FormatActivityDetails(details *StravaActivityDetail) string
	FormatStreamData(streams *StravaStreams, mode string) string
	FormatDerivedFeatures(features interface{}) string
	FormatStreamSummary(summary interface{}) string
	FormatStreamPage(page interface{}) string
}

// outputFormatter implements the OutputFormatter interface
type outputFormatter struct{}

// NewOutputFormatter creates a new output formatter instance
func NewOutputFormatter() OutputFormatter {
	return &outputFormatter{}
}

// FormatAthleteProfile formats athlete profile data with emojis and readable markdown structure
func (f *outputFormatter) FormatAthleteProfile(profile *StravaAthlete) string {
	if profile == nil {
		return "❌ **No athlete profile data available**"
	}

	var builder strings.Builder
	
	// Header with athlete name and basic info
	builder.WriteString(fmt.Sprintf("👤 **%s %s**", profile.Firstname, profile.Lastname))
	if profile.Username != "" {
		builder.WriteString(fmt.Sprintf(" (@%s)", profile.Username))
	}
	builder.WriteString("\n\n")

	// Location information
	if profile.City != "" || profile.State != "" || profile.Country != "" {
		builder.WriteString("📍 **Location:** ")
		locationParts := []string{}
		if profile.City != "" {
			locationParts = append(locationParts, profile.City)
		}
		if profile.State != "" {
			locationParts = append(locationParts, profile.State)
		}
		if profile.Country != "" {
			locationParts = append(locationParts, profile.Country)
		}
		builder.WriteString(strings.Join(locationParts, ", "))
		builder.WriteString("\n")
	}

	// Account details
	builder.WriteString("⚙️ **Account Details:**\n")
	if profile.Sex != "" {
		builder.WriteString(fmt.Sprintf("- Gender: %s\n", profile.Sex))
	}
	if profile.Premium {
		builder.WriteString("- Premium: ✅ Yes\n")
	} else {
		builder.WriteString("- Premium: ❌ No\n")
	}
	if profile.Summit {
		builder.WriteString("- Summit: ✅ Yes\n")
	} else {
		builder.WriteString("- Summit: ❌ No\n")
	}

	// Performance metrics
	if profile.Weight > 0 || profile.FTP > 0 {
		builder.WriteString("\n📊 **Performance Metrics:**\n")
		if profile.Weight > 0 {
			builder.WriteString(fmt.Sprintf("- Weight: %.1f kg\n", profile.Weight))
		}
		if profile.FTP > 0 {
			builder.WriteString(fmt.Sprintf("- FTP: %d watts\n", profile.FTP))
		}
	}

	// Account dates
	if profile.CreatedAt != "" {
		builder.WriteString("\n📅 **Account Information:**\n")
		if createdTime, err := time.Parse(time.RFC3339, profile.CreatedAt); err == nil {
			builder.WriteString(fmt.Sprintf("- Member since: %s\n", createdTime.Format("January 2, 2006")))
		}
		if profile.UpdatedAt != "" {
			if updatedTime, err := time.Parse(time.RFC3339, profile.UpdatedAt); err == nil {
				builder.WriteString(fmt.Sprintf("- Last updated: %s\n", updatedTime.Format("January 2, 2006")))
			}
		}
	}

	return builder.String()
}

// FormatActivities formats activity lists with concise summary format in markdown
func (f *outputFormatter) FormatActivities(activities []*StravaActivity) string {
	if len(activities) == 0 {
		return "📭 **No recent activities found**"
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("🏃 **Recent Activities** (%d activities)\n\n", len(activities)))

	for _, activity := range activities {
		// Activity type emoji
		emoji := f.getActivityEmoji(activity.Type, activity.SportType)
		
		// Parse date
		var dateStr string
		if startTime, err := time.Parse(time.RFC3339, activity.StartDateLocal); err == nil {
			dateStr = startTime.Format("1/2/2006")
		} else {
			dateStr = "Unknown date"
		}

		// Format distance
		distanceStr := f.formatDistance(activity.Distance)
		
		// Build activity line
		builder.WriteString(fmt.Sprintf("%s **%s** (ID: %d) — %s on %s\n", 
			emoji, activity.Name, activity.ID, distanceStr, dateStr))
	}

	return builder.String()
}

// FormatActivityDetails formats detailed activity information with comprehensive metrics in markdown
func (f *outputFormatter) FormatActivityDetails(details *StravaActivityDetail) string {
	if details == nil {
		return "❌ **No activity details available**"
	}

	var builder strings.Builder
	
	// Header with activity name and type
	emoji := f.getActivityEmoji(details.Type, details.SportType)
	builder.WriteString(fmt.Sprintf("%s **%s** (ID: %d)\n", emoji, details.Name, details.ID))

	// Activity type and date
	builder.WriteString(fmt.Sprintf("- Type: %s", details.Type))
	if details.SportType != "" && details.SportType != details.Type {
		builder.WriteString(fmt.Sprintf(" (%s)", details.SportType))
	}
	builder.WriteString("\n")

	// Date and time
	if startTime, err := time.Parse(time.RFC3339, details.StartDateLocal); err == nil {
		builder.WriteString(fmt.Sprintf("- Date: %s\n", startTime.Format("1/2/2006, 3:04:05 PM")))
	}

	// Duration information
	if details.MovingTime > 0 || details.ElapsedTime > 0 {
		builder.WriteString("- ")
		if details.MovingTime > 0 {
			builder.WriteString(fmt.Sprintf("Moving Time: %s", f.formatDuration(details.MovingTime)))
		}
		if details.ElapsedTime > 0 && details.ElapsedTime != details.MovingTime {
			if details.MovingTime > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprintf("Elapsed Time: %s", f.formatDuration(details.ElapsedTime)))
		}
		builder.WriteString("\n")
	}

	// Distance and elevation
	if details.Distance > 0 {
		builder.WriteString(fmt.Sprintf("- Distance: %s\n", f.formatDistance(details.Distance)))
	}
	if details.TotalElevationGain > 0 {
		builder.WriteString(fmt.Sprintf("- Elevation Gain: %.0f m\n", details.TotalElevationGain))
	}

	// Speed metrics
	if details.AverageSpeed > 0 || details.MaxSpeed > 0 {
		builder.WriteString("- ")
		if details.AverageSpeed > 0 {
			builder.WriteString(fmt.Sprintf("Average Speed: %s", f.formatSpeed(details.AverageSpeed)))
		}
		if details.MaxSpeed > 0 {
			if details.AverageSpeed > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprintf("Max Speed: %s", f.formatSpeed(details.MaxSpeed)))
		}
		builder.WriteString("\n")
	}

	// Cadence metrics
	if details.AveragePower > 0 || details.MaxPower > 0 {
		builder.WriteString("- ")
		if details.AveragePower > 0 {
			builder.WriteString(fmt.Sprintf("Avg Power: %.1fW", details.AveragePower))
		}
		if details.MaxPower > 0 {
			if details.AveragePower > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprintf("Max Power: %.0fW", details.MaxPower))
		}
		builder.WriteString("\n")
	}

	// Heart rate metrics
	if details.AverageHeartrate > 0 || details.MaxHeartrate > 0 {
		builder.WriteString("- ")
		if details.AverageHeartrate > 0 {
			builder.WriteString(fmt.Sprintf("Avg Heart Rate: %.1f bpm", details.AverageHeartrate))
		}
		if details.MaxHeartrate > 0 {
			if details.AverageHeartrate > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprintf("Max Heart Rate: %.0f bpm", details.MaxHeartrate))
		}
		builder.WriteString("\n")
	}

	// Energy and temperature
	if details.Calories > 0 {
		builder.WriteString(fmt.Sprintf("- Calories: %.0f\n", details.Calories))
	}
	if details.AverageTemp > 0 {
		builder.WriteString(fmt.Sprintf("- Average Temperature: %.1f°C\n", details.AverageTemp))
	}

	// Gear information
	if details.Gear.Name != "" {
		builder.WriteString(fmt.Sprintf("- Gear: %s", details.Gear.Name))
		if details.Gear.BrandName != "" || details.Gear.ModelName != "" {
			builder.WriteString(" (")
			if details.Gear.BrandName != "" {
				builder.WriteString(details.Gear.BrandName)
				if details.Gear.ModelName != "" {
					builder.WriteString(" ")
				}
			}
			if details.Gear.ModelName != "" {
				builder.WriteString(details.Gear.ModelName)
			}
			builder.WriteString(")")
		}
		builder.WriteString("\n")
	}

	// Social metrics
	if details.KudosCount > 0 || details.CommentCount > 0 || details.PRCount > 0 {
		builder.WriteString("\n🎯 **Social & Achievements:**\n")
		if details.KudosCount > 0 {
			builder.WriteString(fmt.Sprintf("- Kudos: %d\n", details.KudosCount))
		}
		if details.CommentCount > 0 {
			builder.WriteString(fmt.Sprintf("- Comments: %d\n", details.CommentCount))
		}
		if details.PRCount > 0 {
			builder.WriteString(fmt.Sprintf("- Personal Records: %d\n", details.PRCount))
		}
	}

	// Activity flags
	flags := []string{}
	if details.Trainer {
		flags = append(flags, "Indoor/Trainer")
	}
	if details.Commute {
		flags = append(flags, "Commute")
	}
	if details.Manual {
		flags = append(flags, "Manual Entry")
	}
	if details.Private {
		flags = append(flags, "Private")
	}
	if len(flags) > 0 {
		builder.WriteString(fmt.Sprintf("\n🏷️ **Activity Flags:** %s\n", strings.Join(flags, ", ")))
	}

	// Description
	if details.Description != "" {
		builder.WriteString(fmt.Sprintf("\n📝 **Description:**\n%s\n", details.Description))
	}

	return builder.String()
}

// FormatStreamData formats stream data based on the specified mode
func (f *outputFormatter) FormatStreamData(streams *StravaStreams, mode string) string {
	if streams == nil {
		return "❌ **No stream data available**"
	}

	var builder strings.Builder
	
	// Header with mode information
	builder.WriteString(fmt.Sprintf("📊 **Stream Data** (%s mode)\n\n", mode))
	
	// Count total data points
	totalPoints := f.countStreamDataPoints(streams)
	builder.WriteString(fmt.Sprintf("**Total Data Points:** %d\n\n", totalPoints))
	
	// List available stream types
	streamTypes := f.getStreamTypes(streams)
	builder.WriteString("**Available Streams:**\n")
	for _, streamType := range streamTypes {
		builder.WriteString(fmt.Sprintf("- %s\n", streamType))
	}
	
	builder.WriteString("\n**Stream Data Summary:**\n")
	
	// Format each stream type with basic statistics
	if len(streams.Time) > 0 {
		builder.WriteString(fmt.Sprintf("- **Time:** %d data points (0-%d seconds)\n", len(streams.Time), streams.Time[len(streams.Time)-1]))
	}
	if len(streams.Distance) > 0 {
		builder.WriteString(fmt.Sprintf("- **Distance:** %d data points (%.2f-%.2f meters)\n", len(streams.Distance), streams.Distance[0], streams.Distance[len(streams.Distance)-1]))
	}
	if len(streams.Heartrate) > 0 {
		min, max := f.findMinMaxInt(streams.Heartrate)
		avg := f.calculateAvgInt(streams.Heartrate)
		builder.WriteString(fmt.Sprintf("- **Heart Rate:** %d data points (%d-%d bpm, avg: %.1f bpm)\n", len(streams.Heartrate), min, max, avg))
	}
	if len(streams.Watts) > 0 {
		min, max := f.findMinMaxInt(streams.Watts)
		avg := f.calculateAvgInt(streams.Watts)
		builder.WriteString(fmt.Sprintf("- **Power:** %d data points (%d-%d watts, avg: %.1f watts)\n", len(streams.Watts), min, max, avg))
	}
	if len(streams.Cadence) > 0 {
		min, max := f.findMinMaxInt(streams.Cadence)
		avg := f.calculateAvgInt(streams.Cadence)
		builder.WriteString(fmt.Sprintf("- **Cadence:** %d data points (%d-%d rpm, avg: %.1f rpm)\n", len(streams.Cadence), min, max, avg))
	}
	if len(streams.Altitude) > 0 {
		min, max := f.findMinMaxFloat(streams.Altitude)
		avg := f.calculateAvgFloat(streams.Altitude)
		builder.WriteString(fmt.Sprintf("- **Altitude:** %d data points (%.1f-%.1f meters, avg: %.1f meters)\n", len(streams.Altitude), min, max, avg))
	}
	if len(streams.VelocitySmooth) > 0 {
		min, max := f.findMinMaxFloat(streams.VelocitySmooth)
		avg := f.calculateAvgFloat(streams.VelocitySmooth)
		builder.WriteString(fmt.Sprintf("- **Velocity:** %d data points (%.2f-%.2f m/s, avg: %.2f m/s)\n", len(streams.VelocitySmooth), min, max, avg))
	}
	if len(streams.Temp) > 0 {
		min, max := f.findMinMaxInt(streams.Temp)
		avg := f.calculateAvgInt(streams.Temp)
		builder.WriteString(fmt.Sprintf("- **Temperature:** %d data points (%d-%d°C, avg: %.1f°C)\n", len(streams.Temp), min, max, avg))
	}
	if len(streams.GradeSmooth) > 0 {
		min, max := f.findMinMaxFloat(streams.GradeSmooth)
		avg := f.calculateAvgFloat(streams.GradeSmooth)
		builder.WriteString(fmt.Sprintf("- **Grade:** %d data points (%.1f%%-%.1f%%, avg: %.1f%%)\n", len(streams.GradeSmooth), min*100, max*100, avg*100))
	}
	if len(streams.Moving) > 0 {
		trueCount := 0
		for _, moving := range streams.Moving {
			if moving {
				trueCount++
			}
		}
		movingPercent := float64(trueCount) / float64(len(streams.Moving)) * 100
		builder.WriteString(fmt.Sprintf("- **Moving:** %d data points (%.1f%% moving time)\n", len(streams.Moving), movingPercent))
	}
	if len(streams.Latlng) > 0 {
		builder.WriteString(fmt.Sprintf("- **GPS Coordinates:** %d data points\n", len(streams.Latlng)))
	}
	
	return builder.String()
}

// FormatDerivedFeatures formats derived features with comprehensive stream analysis
func (f *outputFormatter) FormatDerivedFeatures(features interface{}) string {
	derivedFeatures, ok := features.(*DerivedFeatures)
	if !ok || derivedFeatures == nil {
		return "❌ **No derived features data available**"
	}

	var builder strings.Builder
	
	// Header with activity ID
	builder.WriteString(fmt.Sprintf("📊 **Stream Analysis** (Activity ID: %d)\n\n", derivedFeatures.ActivityID))

	// Overview section
	f.formatOverviewSection(&builder, &derivedFeatures.Summary)

	// Statistical analysis section
	f.formatStatisticsSection(&builder, &derivedFeatures.Statistics)

	// Trends and patterns section
	f.formatTrendsSection(&builder, derivedFeatures.Trends)

	// Spikes and anomalies section
	f.formatSpikesSection(&builder, derivedFeatures.Spikes)

	// Inflection points section
	f.formatInflectionPointsSection(&builder, derivedFeatures.InflectionPoints)

	// Lap-by-lap analysis section (if available)
	if derivedFeatures.LapAnalysis != nil {
		f.formatLapAnalysisSection(&builder, derivedFeatures.LapAnalysis)
	}

	// Sample data section
	f.formatSampleDataSection(&builder, derivedFeatures.SampleData)

	return builder.String()
}

// FormatStreamSummary formats AI-generated stream summary
func (f *outputFormatter) FormatStreamSummary(summary interface{}) string {
	streamSummary, ok := summary.(*StreamSummary)
	if !ok || streamSummary == nil {
		return "❌ **No stream summary data available**"
	}

	var builder strings.Builder
	
	// Header with activity ID and model info
	builder.WriteString(fmt.Sprintf("🤖 **AI Stream Summary** (Activity ID: %d)\n\n", streamSummary.ActivityID))
	
	if streamSummary.Model != "" {
		builder.WriteString(fmt.Sprintf("**Model:** %s", streamSummary.Model))
		if streamSummary.TokensUsed > 0 {
			builder.WriteString(fmt.Sprintf(" | **Tokens Used:** %d", streamSummary.TokensUsed))
		}
		builder.WriteString("\n\n")
	}
	
	// Show the custom prompt used
	if streamSummary.SummaryPrompt != "" {
		builder.WriteString("**Analysis Request:**\n")
		builder.WriteString(fmt.Sprintf("> %s\n\n", streamSummary.SummaryPrompt))
	}
	
	// Add the AI-generated summary
	builder.WriteString("**AI Analysis:**\n\n")
	builder.WriteString(streamSummary.Summary)
	
	return builder.String()
}

// FormatStreamPage formats paginated stream page with navigation info
func (f *outputFormatter) FormatStreamPage(page interface{}) string {
	streamPage, ok := page.(*StreamPage)
	if !ok || streamPage == nil {
		return "❌ **No stream page data available**"
	}

	var builder strings.Builder
	
	// Page header with navigation info
	builder.WriteString(fmt.Sprintf("📄 **Page %d of %d** for Activity %d\n\n", 
		streamPage.PageNumber, streamPage.TotalPages, streamPage.ActivityID))
	
	// Add time range information
	if streamPage.TimeRange.StartTime > 0 && streamPage.TimeRange.EndTime > 0 {
		duration := streamPage.TimeRange.EndTime - streamPage.TimeRange.StartTime
		builder.WriteString(fmt.Sprintf("**Time Range:** %d-%d seconds (Duration: %d seconds)\n\n", 
			streamPage.TimeRange.StartTime, streamPage.TimeRange.EndTime, duration))
	}
	
	// Processing mode information
	builder.WriteString("**Processing Mode:** ")
	switch streamPage.ProcessingMode {
	case "raw":
		builder.WriteString("Raw stream data points")
	case "derived":
		builder.WriteString("Derived features and statistics")
	case "ai-summary":
		builder.WriteString("AI-generated summary")
	default:
		builder.WriteString(streamPage.ProcessingMode)
	}
	builder.WriteString("\n\n")
	
	// Add the processed data
	if dataStr, ok := streamPage.Data.(string); ok {
		builder.WriteString(dataStr)
	} else {
		builder.WriteString("**Processed Data:**\n")
		builder.WriteString(fmt.Sprintf("%+v", streamPage.Data))
	}
	
	// Navigation instructions
	if streamPage.HasNextPage {
		builder.WriteString("\n\n**Navigation:**\n")
		builder.WriteString(fmt.Sprintf("- Next page: Call get-activity-streams with page_number=%d\n", streamPage.PageNumber+1))
		
		if streamPage.PageNumber > 1 {
			builder.WriteString(fmt.Sprintf("- Previous page: Call get-activity-streams with page_number=%d\n", streamPage.PageNumber-1))
		}
		
		builder.WriteString(fmt.Sprintf("- Jump to specific page: Call get-activity-streams with page_number=X (1-%d)\n", streamPage.TotalPages))
		builder.WriteString("- Get full dataset: Call get-activity-streams with page_size=-1\n")
		
		builder.WriteString("\n💡 **Tip:** Analyze this page's data before requesting the next page to maintain context efficiency.")
	} else {
		builder.WriteString("\n\n✅ **Complete:** This is the final page of data for this activity.")
	}
	
	// Add footer with token usage info
	builder.WriteString(fmt.Sprintf("\n\n📊 **Page Stats:** %d estimated tokens", streamPage.EstimatedTokens))
	
	if streamPage.HasNextPage {
		builder.WriteString(fmt.Sprintf(" | Next: page %d", streamPage.PageNumber+1))
	}
	
	return builder.String()
}

// Helper methods

// getActivityEmoji returns appropriate emoji for activity type
func (f *outputFormatter) getActivityEmoji(activityType, sportType string) string {
	// Use sport type first if available, then fall back to activity type
	typeToCheck := sportType
	if typeToCheck == "" {
		typeToCheck = activityType
	}

	switch strings.ToLower(typeToCheck) {
	case "run", "running":
		return "🏃"
	case "ride", "cycling", "bike":
		return "🚴"
	case "swim", "swimming":
		return "🏊"
	case "walk", "walking":
		return "🚶"
	case "hike", "hiking":
		return "🥾"
	case "workout", "crosstraining":
		return "💪"
	case "yoga":
		return "🧘"
	case "ski", "skiing":
		return "⛷️"
	case "snowboard", "snowboarding":
		return "🏂"
	case "kayak", "kayaking", "canoe", "canoeing":
		return "🛶"
	case "rowing":
		return "🚣"
	case "golf":
		return "⛳"
	case "tennis":
		return "🎾"
	case "soccer", "football":
		return "⚽"
	case "basketball":
		return "🏀"
	case "climbing", "rockclimbing":
		return "🧗"
	default:
		return "🏃" // Default to running emoji
	}
}

// formatDistance formats distance in appropriate units
func (f *outputFormatter) formatDistance(distanceMeters float64) string {
	if distanceMeters < 1000 {
		return fmt.Sprintf("%.0fm", distanceMeters)
	}
	return fmt.Sprintf("%.2fkm", distanceMeters/1000)
}

// formatSpeed formats speed in km/h
func (f *outputFormatter) formatSpeed(speedMPS float64) string {
	speedKMH := speedMPS * 3.6
	return fmt.Sprintf("%.1f km/h", speedKMH)
}

// formatDuration formats duration in human-readable format
func (f *outputFormatter) formatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}

// formatOverviewSection formats the overview with key metrics
func (f *outputFormatter) formatOverviewSection(builder *strings.Builder, summary *FeatureSummary) {
	builder.WriteString("## 📈 **Overview**\n\n")
	
	// Duration and data points
	builder.WriteString(fmt.Sprintf("- **Duration:** %s (%d data points)\n", 
		f.formatDuration(summary.Duration), summary.TotalDataPoints))
	
	// Distance and elevation
	if summary.TotalDistance > 0 {
		builder.WriteString(fmt.Sprintf("- **Distance:** %s", f.formatDistance(summary.TotalDistance)))
		if summary.ElevationGain > 0 {
			builder.WriteString(fmt.Sprintf(" with %.0fm elevation gain", summary.ElevationGain))
		}
		if summary.ElevationLoss > 0 {
			builder.WriteString(fmt.Sprintf(" and %.0fm elevation loss", summary.ElevationLoss))
		}
		builder.WriteString("\n")
	}
	
	// Stream types available
	if len(summary.StreamTypes) > 0 {
		builder.WriteString(fmt.Sprintf("- **Available Data:** %s\n", strings.Join(summary.StreamTypes, ", ")))
	}
	
	// Moving time percentage
	if summary.MovingTimePercent > 0 {
		builder.WriteString(fmt.Sprintf("- **Moving Time:** %.1f%% of total time\n", summary.MovingTimePercent))
	}
	
	builder.WriteString("\n")
}

// formatStatisticsSection formats statistical analysis for all metrics
func (f *outputFormatter) formatStatisticsSection(builder *strings.Builder, stats *StreamStatistics) {
	builder.WriteString("## 📊 **Statistical Analysis**\n\n")
	
	// Heart Rate Analysis
	if stats.HeartRate != nil {
		builder.WriteString("### 💓 **Heart Rate Analysis**\n")
		f.formatMetricStats(builder, "Heart Rate", stats.HeartRate, "bpm")
		builder.WriteString("\n")
	}
	
	// Power Analysis
	if stats.Power != nil {
		builder.WriteString("### ⚡ **Power Analysis**\n")
		f.formatMetricStats(builder, "Power", stats.Power, "W")
		builder.WriteString("\n")
	}
	
	// Speed Analysis
	if stats.VelocitySmooth != nil {
		builder.WriteString("### 🏃 **Speed Analysis**\n")
		f.formatSpeedStats(builder, stats.VelocitySmooth)
		builder.WriteString("\n")
	}
	
	// Elevation Analysis
	if stats.Altitude != nil {
		builder.WriteString("### ⛰️ **Elevation Analysis**\n")
		f.formatMetricStats(builder, "Altitude", stats.Altitude, "m")
		builder.WriteString("\n")
	}
	
	// Cadence Analysis
	if stats.Cadence != nil {
		builder.WriteString("### 🔄 **Cadence Analysis**\n")
		f.formatMetricStats(builder, "Cadence", stats.Cadence, "rpm")
		builder.WriteString("\n")
	}
	
	// Temperature Analysis
	if stats.Temperature != nil {
		builder.WriteString("### 🌡️ **Temperature Analysis**\n")
		f.formatMetricStats(builder, "Temperature", stats.Temperature, "°C")
		builder.WriteString("\n")
	}
	
	// Grade Analysis
	if stats.Grade != nil {
		builder.WriteString("### 📐 **Grade Analysis**\n")
		f.formatMetricStats(builder, "Grade", stats.Grade, "%")
		builder.WriteString("\n")
	}
	
	// Moving Time Analysis
	if stats.Moving != nil {
		builder.WriteString("### 🚶 **Moving Time Analysis**\n")
		f.formatBooleanStats(builder, stats.Moving)
		builder.WriteString("\n")
	}
	
	// Location Analysis
	if stats.LatLng != nil {
		builder.WriteString("### 🗺️ **Location Analysis**\n")
		f.formatLocationStats(builder, stats.LatLng)
		builder.WriteString("\n")
	}
}

// formatMetricStats formats statistics for a numeric metric
func (f *outputFormatter) formatMetricStats(builder *strings.Builder, metricName string, stats *MetricStats, unit string) {
	builder.WriteString(fmt.Sprintf("- **Range:** %.1f - %.1f %s (Mean: %.1f %s)\n", 
		stats.Min, stats.Max, unit, stats.Mean, unit))
	builder.WriteString(fmt.Sprintf("- **Median:** %.1f %s (Q25: %.1f, Q75: %.1f)\n", 
		stats.Median, unit, stats.Q25, stats.Q75))
	builder.WriteString(fmt.Sprintf("- **Variability:** %.1f%% (StdDev: %.1f %s)\n", 
		stats.Variability*100, stats.StdDev, unit))
	builder.WriteString(fmt.Sprintf("- **Data Points:** %d\n", stats.Count))
}

// formatSpeedStats formats speed statistics with km/h conversion
func (f *outputFormatter) formatSpeedStats(builder *strings.Builder, stats *MetricStats) {
	// Convert m/s to km/h for display
	minKmh := stats.Min * 3.6
	maxKmh := stats.Max * 3.6
	meanKmh := stats.Mean * 3.6
	medianKmh := stats.Median * 3.6
	
	builder.WriteString(fmt.Sprintf("- **Range:** %.1f - %.1f km/h (Mean: %.1f km/h)\n", 
		minKmh, maxKmh, meanKmh))
	builder.WriteString(fmt.Sprintf("- **Median:** %.1f km/h\n", medianKmh))
	builder.WriteString(fmt.Sprintf("- **Variability:** %.1f%% (StdDev: %.1f km/h)\n", 
		stats.Variability*100, stats.StdDev*3.6))
	builder.WriteString(fmt.Sprintf("- **Data Points:** %d\n", stats.Count))
}

// formatBooleanStats formats statistics for boolean metrics
func (f *outputFormatter) formatBooleanStats(builder *strings.Builder, stats *BooleanStats) {
	builder.WriteString(fmt.Sprintf("- **Moving:** %d data points (%.1f%%)\n", 
		stats.TrueCount, stats.TruePercent))
	builder.WriteString(fmt.Sprintf("- **Stopped:** %d data points (%.1f%%)\n", 
		stats.FalseCount, stats.FalsePercent))
	builder.WriteString(fmt.Sprintf("- **Total Data Points:** %d\n", stats.TotalCount))
}

// formatLocationStats formats GPS location statistics
func (f *outputFormatter) formatLocationStats(builder *strings.Builder, stats *LocationStats) {
	builder.WriteString(fmt.Sprintf("- **Start:** %.6f, %.6f\n", stats.StartLat, stats.StartLng))
	builder.WriteString(fmt.Sprintf("- **End:** %.6f, %.6f\n", stats.EndLat, stats.EndLng))
	builder.WriteString("- **Bounding Box:**\n")
	builder.WriteString(fmt.Sprintf("  - North: %.6f, South: %.6f\n", 
		stats.BoundingBox.NorthLat, stats.BoundingBox.SouthLat))
	builder.WriteString(fmt.Sprintf("  - East: %.6f, West: %.6f\n", 
		stats.BoundingBox.EastLng, stats.BoundingBox.WestLng))
	builder.WriteString(fmt.Sprintf("- **GPS Points:** %d\n", stats.TotalPoints))
}

// formatTrendsSection formats trend analysis
func (f *outputFormatter) formatTrendsSection(builder *strings.Builder, trends []Trend) {
	if len(trends) == 0 {
		return
	}
	
	builder.WriteString("## 📈 **Trend Analysis**\n\n")
	
	for _, trend := range trends {
		directionEmoji := f.getTrendEmoji(trend.Direction)
		builder.WriteString(fmt.Sprintf("- **%s %s:** %s trend from %s to %s (Magnitude: %.2f, Confidence: %.1f%%)\n",
			directionEmoji, trend.Metric, trend.Direction,
			f.formatDuration(trend.StartTime), f.formatDuration(trend.EndTime),
			trend.Magnitude, trend.Confidence*100))
	}
	
	builder.WriteString("\n")
}

// formatSpikesSection formats spike analysis
func (f *outputFormatter) formatSpikesSection(builder *strings.Builder, spikes []Spike) {
	if len(spikes) == 0 {
		return
	}
	
	builder.WriteString("## 🔥 **Spikes and Anomalies**\n\n")
	
	for _, spike := range spikes {
		builder.WriteString(fmt.Sprintf("- **%s Spike:** %.1f at %s (Magnitude: %.2fx, Duration: %ds)\n",
			spike.Metric, spike.Value, f.formatDuration(spike.Time),
			spike.Magnitude, spike.Duration))
	}
	
	builder.WriteString("\n")
}

// formatInflectionPointsSection formats inflection point analysis
func (f *outputFormatter) formatInflectionPointsSection(builder *strings.Builder, points []InflectionPoint) {
	if len(points) == 0 {
		return
	}
	
	builder.WriteString("## 🔄 **Inflection Points**\n\n")
	
	for _, point := range points {
		builder.WriteString(fmt.Sprintf("- **%s:** %.1f at %s (Direction: %s, Magnitude: %.2f)\n",
			point.Metric, point.Value, f.formatDuration(point.Time),
			point.Direction, point.Magnitude))
	}
	
	builder.WriteString("\n")
}

// formatLapAnalysisSection formats lap-by-lap analysis
func (f *outputFormatter) formatLapAnalysisSection(builder *strings.Builder, lapAnalysis *LapAnalysis) {
	builder.WriteString("## 🏁 **Lap-by-Lap Analysis**\n\n")
	
	// Overview
	builder.WriteString(fmt.Sprintf("**Segmentation:** %s (%d segments)\n\n", 
		lapAnalysis.SegmentationType, lapAnalysis.TotalLaps))
	
	// Lap summaries
	builder.WriteString("### 📊 **Lap Performance Summary**\n\n")
	for _, lap := range lapAnalysis.LapSummaries {
		f.formatLapSummary(builder, &lap)
	}
	
	// Lap comparisons
	builder.WriteString("### 🏆 **Lap Comparisons**\n\n")
	f.formatLapComparisons(builder, &lapAnalysis.LapComparisons)
}

// formatLapSummary formats a single lap summary
func (f *outputFormatter) formatLapSummary(builder *strings.Builder, lap *LapSummary) {
	lapName := fmt.Sprintf("Lap %d", lap.LapNumber)
	if lap.LapName != "" {
		lapName = lap.LapName
	}
	
	builder.WriteString(fmt.Sprintf("**%s:** ", lapName))
	
	// Distance and duration
	if lap.Distance > 0 {
		builder.WriteString(fmt.Sprintf("%.2fkm in %s", lap.Distance/1000, f.formatDuration(lap.Duration)))
	} else {
		builder.WriteString(fmt.Sprintf("%s", f.formatDuration(lap.Duration)))
	}
	
	// Key metrics
	metrics := []string{}
	if lap.AvgSpeed > 0 {
		metrics = append(metrics, fmt.Sprintf("Avg Speed: %.1f km/h", lap.AvgSpeed*3.6))
	}
	if lap.AvgHeartRate > 0 {
		metrics = append(metrics, fmt.Sprintf("Avg HR: %.0f bpm", lap.AvgHeartRate))
	}
	if lap.AvgPower > 0 {
		metrics = append(metrics, fmt.Sprintf("Avg Power: %.0fW", lap.AvgPower))
	}
	
	if len(metrics) > 0 {
		builder.WriteString(fmt.Sprintf(" - %s", strings.Join(metrics, ", ")))
	}
	
	builder.WriteString("\n")
}

// formatLapComparisons formats lap comparison metrics
func (f *outputFormatter) formatLapComparisons(builder *strings.Builder, comparisons *LapComparisons) {
	builder.WriteString(fmt.Sprintf("- **Fastest Lap:** Lap %d\n", comparisons.FastestLap))
	builder.WriteString(fmt.Sprintf("- **Slowest Lap:** Lap %d\n", comparisons.SlowestLap))
	
	if comparisons.HighestPowerLap > 0 {
		builder.WriteString(fmt.Sprintf("- **Highest Power:** Lap %d\n", comparisons.HighestPowerLap))
	}
	if comparisons.HighestHRLap > 0 {
		builder.WriteString(fmt.Sprintf("- **Highest HR:** Lap %d\n", comparisons.HighestHRLap))
	}
	
	builder.WriteString(fmt.Sprintf("- **Speed Variation:** %.1f%% across laps\n", comparisons.SpeedVariation*100))
	if comparisons.PowerVariation > 0 {
		builder.WriteString(fmt.Sprintf("- **Power Variation:** %.1f%% across laps\n", comparisons.PowerVariation*100))
	}
	if comparisons.HRVariation > 0 {
		builder.WriteString(fmt.Sprintf("- **HR Variation:** %.1f%% across laps\n", comparisons.HRVariation*100))
	}
	
	builder.WriteString(fmt.Sprintf("- **Consistency Score:** %.1f/10\n", comparisons.ConsistencyScore*10))
	builder.WriteString("\n")
}

// formatSampleDataSection formats representative sample data points
func (f *outputFormatter) formatSampleDataSection(builder *strings.Builder, sampleData []DataPoint) {
	if len(sampleData) == 0 {
		return
	}
	
	builder.WriteString("## 📋 **Sample Data Points**\n\n")
	builder.WriteString("Representative data points from the activity:\n\n")
	
	// Show up to 5 sample points
	maxSamples := len(sampleData)
	if maxSamples > 5 {
		maxSamples = 5
	}
	
	for i := 0; i < maxSamples; i++ {
		point := sampleData[i]
		builder.WriteString(fmt.Sprintf("**Time %s:**", f.formatDuration(point.TimeOffset)))
		
		values := []string{}
		for metric, value := range point.Values {
			switch v := value.(type) {
			case float64:
				values = append(values, fmt.Sprintf("%s: %.1f", metric, v))
			case int:
				values = append(values, fmt.Sprintf("%s: %d", metric, v))
			case bool:
				values = append(values, fmt.Sprintf("%s: %t", metric, v))
			default:
				values = append(values, fmt.Sprintf("%s: %v", metric, v))
			}
		}
		
		if len(values) > 0 {
			builder.WriteString(fmt.Sprintf(" %s", strings.Join(values, ", ")))
		}
		builder.WriteString("\n")
	}
	
	if len(sampleData) > 5 {
		builder.WriteString(fmt.Sprintf("\n*... and %d more data points*\n", len(sampleData)-5))
	}
	
	builder.WriteString("\n")
}

// getTrendEmoji returns appropriate emoji for trend direction
func (f *outputFormatter) getTrendEmoji(direction string) string {
	switch direction {
	case "increasing":
		return "📈"
	case "decreasing":
		return "📉"
	case "stable":
		return "➡️"
	default:
		return "📊"
	}
}

// Helper methods for stream data formatting

func (f *outputFormatter) countStreamDataPoints(data *StravaStreams) int {
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

func (f *outputFormatter) getStreamTypes(data *StravaStreams) []string {
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

func (f *outputFormatter) findMinMaxInt(data []int) (int, int) {
	if len(data) == 0 {
		return 0, 0
	}
	
	min, max := data[0], data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

func (f *outputFormatter) findMinMaxFloat(data []float64) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}
	
	min, max := data[0], data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

func (f *outputFormatter) calculateAvgInt(data []int) float64 {
	if len(data) == 0 {
		return 0
	}
	
	sum := 0
	for _, v := range data {
		sum += v
	}
	return float64(sum) / float64(len(data))
}

func (f *outputFormatter) calculateAvgFloat(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}