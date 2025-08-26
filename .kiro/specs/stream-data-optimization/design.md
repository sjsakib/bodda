# Stream Data Optimization Design

## Overview

This design implements a stream data optimization system that handles large tool outputs exceeding LLM context window limits. The system provides three processing modes: derived feature extraction, AI-powered summarization, and paginated reading. It also includes context optimization through redaction of previous stream tool outputs.

The solution integrates into the existing AI service architecture, specifically targeting the `get-activity-streams` tool which can return large time-series datasets from Strava activities.

## Architecture

### Core Components

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   AI Service    │───▶│ Stream Processor │───▶│ Processing Mode │
│                 │    │                  │    │   Dispatcher    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │ Context Manager │    │ Mode Processors │
                       │   (Redaction)   │    │ • Derived       │
                       └─────────────────┘    │ • Summary       │
                                              │ • Paginated     │
                                              └─────────────────┘
```

### Integration Points

- **AI Service**: Modified `executeGetActivityStreams` method to check output size
- **Tool Definition**: Enhanced `get-activity-streams` tool with new processing parameters
- **Tool Result Processing**: Enhanced to handle oversized outputs and processing modes
- **Output Formatting**: Added human-readable formatting for all Strava tool outputs
- **Context Management**: Added redaction logic for previous stream outputs
- **Response Generation**: Modified to present processing options to LLM

## Components and Interfaces

### 1. Stream Processor

```go
type StreamProcessor interface {
    ProcessStreamOutput(data *StravaStreams, toolCallID string) (*ProcessedStreamResult, error)
    ShouldProcess(data *StravaStreams) bool
    GetProcessingOptions() []ProcessingOption
}

type ProcessedStreamResult struct {
    ToolCallID      string                 `json:"tool_call_id"`
    Content         string                 `json:"content"` // Human-readable formatted text
    ProcessingMode  string                 `json:"processing_mode,omitempty"`
    Options         []ProcessingOption     `json:"options,omitempty"`
    Data            interface{}            `json:"data,omitempty"` // Raw data for internal use
}

type ProcessingOption struct {
    Mode        string `json:"mode"`
    Description string `json:"description"`
    Command     string `json:"command"`
}
```

### 2. Output Formatter

```go
type OutputFormatter interface {
    FormatAthleteProfile(profile *StravaAthlete) string
    FormatActivities(activities []*StravaActivity) string
    FormatActivityDetails(details *StravaActivityDetail) string
    FormatStreamData(streams *StravaStreams, mode string) string
    FormatDerivedFeatures(features *DerivedFeatures) string
    FormatStreamSummary(summary *StreamSummary) string
    FormatStreamPage(page *StreamPage) string
}
```

### 2. Processing Mode Interfaces

```go
type DerivedFeaturesProcessor interface {
    ExtractFeatures(data *StravaStreams) (*DerivedFeatures, error)
}

type SummaryProcessor interface {
    GenerateSummary(data *StravaStreams, prompt string) (*StreamSummary, error)
}

type PaginatedProcessor interface {
    GetStreamPage(req *PaginatedStreamRequest) (*StreamPage, error)
    EstimatePageCount(activityID int64, streamTypes []string) (int, error)
}
```

### 3. Context Manager

```go
type ContextManager interface {
    RedactPreviousStreamOutputs(messages []openai.ChatCompletionMessage) []openai.ChatCompletionMessage
    ShouldRedact(toolCallName string) bool
}
```

## Data Models

### Stream Processing Models

```go
type DerivedFeatures struct {
    ActivityID      int64                  `json:"activity_id"`
    Summary         FeatureSummary         `json:"summary"`
    InflectionPoints []InflectionPoint     `json:"inflection_points"`
    Statistics      StreamStatistics       `json:"statistics"`
    Trends          []Trend               `json:"trends"`
    Spikes          []Spike               `json:"spikes"`
    SampleData      []DataPoint           `json:"sample_data"`
}

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

type MetricStats struct {
    Min         float64 `json:"min"`
    Max         float64 `json:"max"`
    Mean        float64 `json:"mean"`
    Median      float64 `json:"median"`
    StdDev      float64 `json:"std_dev"`
    Variability float64 `json:"variability"`
    Range       float64 `json:"range"`
    Q25         float64 `json:"q25"`
    Q75         float64 `json:"q75"`
    Count       int     `json:"count"`
}

type BooleanStats struct {
    TrueCount    int     `json:"true_count"`
    FalseCount   int     `json:"false_count"`
    TruePercent  float64 `json:"true_percent"`
    FalsePercent float64 `json:"false_percent"`
    TotalCount   int     `json:"total_count"`
}

type LocationStats struct {
    StartLat    float64 `json:"start_lat"`
    StartLng    float64 `json:"start_lng"`
    EndLat      float64 `json:"end_lat"`
    EndLng      float64 `json:"end_lng"`
    BoundingBox BoundingBox `json:"bounding_box"`
    TotalPoints int     `json:"total_points"`
}

type BoundingBox struct {
    NorthLat float64 `json:"north_lat"`
    SouthLat float64 `json:"south_lat"`
    EastLng  float64 `json:"east_lng"`
    WestLng  float64 `json:"west_lng"`
}
```

### Pagination Models

```go
type PaginatedStreamRequest struct {
    ActivityID    int64    `json:"activity_id"`
    StreamTypes   []string `json:"stream_types"`
    Resolution    string   `json:"resolution"`
    ProcessingMode string  `json:"processing_mode"` // "raw", "derived", "ai-summary"
    PageNumber    int      `json:"page_number"`
    PageSize      int      `json:"page_size"`
}

type StreamPage struct {
    ActivityID     int64                  `json:"activity_id"`
    PageNumber     int                    `json:"page_number"`
    TotalPages     int                    `json:"total_pages"`
    ProcessingMode string                 `json:"processing_mode"`
    Data           interface{}            `json:"data"` // Raw streams, derived features, or summary
    TimeRange      TimeRange              `json:"time_range"`
    Instructions   string                 `json:"instructions"`
    HasNextPage    bool                   `json:"has_next_page"`
}

type TimeRange struct {
    StartTime int `json:"start_time"`
    EndTime   int `json:"end_time"`
}
```

## Error Handling

### Error Types

```go
var (
    ErrStreamTooLarge     = errors.New("stream data exceeds context window limits")
    ErrProcessingFailed   = errors.New("stream processing failed")
    ErrInvalidProcessingMode = errors.New("invalid processing mode specified")
    ErrPageNotFound       = errors.New("requested page not found")
    ErrStreamExpired      = errors.New("paginated stream has expired")
)
```

### Error Handling Strategy

1. **Graceful Degradation**: If processing fails, provide basic size information
2. **Fallback Options**: Always present alternative processing methods
3. **Clear Messaging**: Inform LLM of specific failure reasons
4. **No Automatic Retries**: Let LLM decide on alternative approaches

## Testing Strategy

### Unit Tests

1. **Stream Size Detection**: Test threshold calculations
2. **Feature Extraction**: Validate derived features accuracy
3. **Pagination Logic**: Test page boundary calculations
4. **Context Redaction**: Verify proper message filtering
5. **Error Scenarios**: Test all failure modes

### Integration Tests

1. **End-to-End Processing**: Test complete workflow from stream tool to processed output
2. **AI Service Integration**: Verify proper integration with existing tool execution
3. **Context Window Limits**: Test with various data sizes
4. **Multi-Round Processing**: Test redaction across multiple tool calls

### Performance Tests

1. **Large Dataset Processing**: Test with maximum expected stream sizes
2. **Feature Extraction Speed**: Benchmark derived feature calculations
3. **Memory Usage**: Monitor memory consumption during processing
4. **Concurrent Processing**: Test multiple simultaneous stream processing

## Enhanced Tool Definition

### Updated get-activity-streams Tool Parameters

```json
{
  "name": "get-activity-streams",
  "description": "Get time-series data streams from a Strava activity with processing options for large datasets",
  "parameters": {
    "type": "object",
    "properties": {
      "activity_id": {
        "type": "integer",
        "description": "The Strava activity ID"
      },
      "stream_types": {
        "type": "array",
        "description": "Types of streams to retrieve",
        "items": {
          "type": "string",
          "enum": ["time", "distance", "latlng", "altitude", "velocity_smooth", "heartrate", "cadence", "watts", "temp", "moving", "grade_smooth"]
        },
        "default": ["time", "distance", "heartrate", "watts"]
      },
      "resolution": {
        "type": "string",
        "description": "Resolution of the data",
        "enum": ["low", "medium", "high"],
        "default": "medium"
      },
      "processing_mode": {
        "type": "string",
        "description": "How to process large datasets that exceed context limits",
        "enum": ["auto", "raw", "derived", "ai-summary"],
        "default": "auto"
      },
      "page_number": {
        "type": "integer",
        "description": "Page number for paginated processing (1-based)",
        "minimum": 1,
        "default": 1
      },
      "page_size": {
        "type": "integer",
        "description": "Number of data points per page. Use negative value to request full dataset (subject to context limits)",
        "minimum": -1,
        "maximum": 5000,
        "default": 1000
      },
      "summary_prompt": {
        "type": "string",
        "description": "Custom prompt for AI summarization mode (required when processing_mode is 'ai-summary')"
      }
    },
    "required": ["activity_id"],
    "if": {
      "properties": {
        "processing_mode": {"const": "ai-summary"}
      }
    },
    "then": {
      "required": ["activity_id", "summary_prompt"]
    }
  }
}
```

### Processing Mode Behaviors

All stream requests use pagination by default. Processing modes determine how data is presented:

- **auto**: Automatically detect if data is too large and present processing options
- **raw**: Return raw stream data with minimal processing (paginated if necessary)
- **derived**: Extract key features, statistics, and patterns from the stream data (paginated or full)
- **ai-summary**: Use AI to summarize the stream data with optional custom prompt (paginated or full)

**Page Size Handling:**
- Positive page_size: Return data in pages of specified size
- Negative page_size: Attempt to return full dataset, applying processing mode if too large for context

### Mode Descriptions for LLM

When presenting processing options, include these descriptions:

```
📊 **Processing Mode Options:**

🔍 **raw** - Get the actual stream data points (time, heart rate, power, etc.)
   Best for: Detailed analysis, specific time intervals, technical examination
   
📈 **derived** - Get calculated features, statistics, and insights from the data
   Best for: Performance analysis, training insights, pattern identification
   
🤖 **ai-summary** - Get an AI-generated summary focusing on key findings (requires summary_prompt)
   Best for: Quick overview, coaching insights, narrative understanding
   
⚡ **auto** - Let the system choose the best approach based on data size
   Best for: When unsure which mode to use

📏 **Token Usage Estimates:**
- Page size 500: ~1,200 tokens per page
- Page size 1000: ~2,400 tokens per page  
- Page size 2000: ~4,800 tokens per page
- Full dataset (-1): ~15,000 tokens (may require processing)

💡 **Current context usage: 3,200 tokens remaining**
```

## Strava API Integration

### Leveraging Native Pagination

The Strava Streams API supports several parameters that enable efficient pagination:

- **Resolution**: `low`, `medium`, `high` - controls data density
- **Series Type**: Specific stream types to reduce payload size
- **Time-based Filtering**: Can be implemented using resolution and post-processing

### Pagination Implementation Strategy

1. **Initial Size Check**: Request with `low` resolution to estimate full data size
2. **Dynamic Resolution**: Choose appropriate resolution based on context limits
3. **Chunked Requests**: For large datasets, make multiple API calls with different time ranges
4. **Cross-Mode Processing**: Apply derived features, summary, or raw processing to each chunk

## Implementation Details

### Configuration

```go
type StreamConfig struct {
    MaxContextTokens    int     `json:"max_context_tokens"`
    TokenPerCharRatio   float64 `json:"token_per_char_ratio"`
    DefaultPageSize     int     `json:"default_page_size"`
    MaxPageSize         int     `json:"max_page_size"`
    RedactionEnabled    bool    `json:"redaction_enabled"`
    StravaResolutions   []string `json:"strava_resolutions"` // ["low", "medium", "high"]
}
```

### Size Calculation

- Estimate tokens using character count with configurable ratio
- Account for JSON serialization overhead
- Include safety margin for context preservation

### Feature Extraction Algorithms

1. **Inflection Points**: Detect significant changes in gradient for all numeric streams
2. **Spikes**: Identify values exceeding statistical thresholds across all metrics
3. **Trends**: Calculate moving averages and trend directions for time-series data
4. **Variability**: Compute coefficient of variation and standard deviation for all numeric streams
5. **Elevation Analysis**: Calculate gains, losses, and grade changes from altitude data
6. **Speed Analysis**: Analyze velocity patterns, acceleration, and deceleration phases
7. **Geographic Analysis**: Process GPS coordinates for route characteristics and bounding boxes
8. **Moving Time Analysis**: Analyze moving vs stopped time patterns
9. **Temperature Patterns**: Identify temperature trends and variations during activity
10. **Multi-Metric Correlations**: Analyze relationships between different stream types (e.g., power vs heart rate)

### Pagination Strategy

- **Leverage Strava API Pagination**: Use Strava's native pagination parameters for stream data
- **API-Level Chunking**: Request stream data in chunks using Strava's resolution and time-based parameters
- **Cross-Mode Compatibility**: Pagination works with all processing modes (derived features, summary, raw data)
- **Stateless Approach**: Each page request goes directly to Strava API rather than storing data locally

### Context Redaction

- Simple string replacement of previous stream tool outputs
- Preserve tool call structure while removing content
- Maintain conversation flow and context

## Security Considerations

1. **Data Sanitization**: Ensure processed outputs don't contain sensitive information
2. **Temporary Storage**: Secure handling of paginated stream data
3. **Memory Management**: Prevent memory leaks from large datasets
4. **Access Control**: Verify user permissions for stream data access

## Output Formatting Strategy

### Human-Readable Format Examples

#### Activity List Format
```
🏃 **Morning Ride** (ID: 15347546137) — 29.17km on 8/5/2025
🚴 **Evening Run** (ID: 15347546138) — 10.5km on 8/4/2025
🏊 **Pool Swim** (ID: 15347546139) — 2.0km on 8/3/2025
```

#### Activity Details Format
```
🏃 **Morning Ride** (ID: 15347546137)
- Type: Ride (Cycling)
- Date: 8/5/2025, 12:21:19 PM
- Moving Time: 01:27:10, Elapsed Time: 01:35:15
- Distance: 29.17 km
- Elevation Gain: 104 m
- Average Speed: 20.1 km/h, Max Speed: 39.3 km/h
- Avg Cadence: 83.8 rpm
- Avg Power: 98.8W, Max Power: 245W
- Avg Heart Rate: 144.6 bpm, Max Heart Rate: 171 bpm
- Calories: 649
- Gear: Java Siluro 6 Top
```

#### Stream Data Derived Features Format
```
📊 **Stream Analysis for Morning Ride** (ID: 15347546137)

**Overview:**
- Duration: 01:27:10 (5,230 data points)
- Distance: 29.17 km with 104m elevation gain

**Heart Rate Analysis:**
- Average: 144.6 bpm (Range: 98-171 bpm)
- Time in Zone 2: 45.2%, Zone 3: 32.1%, Zone 4: 18.7%
- Heart rate variability: CV: 12.3%
- Heart rate drift: +3.2 bpm/hour

**Power Analysis:**
- Average: 98.8W (Range: 0-245W)
- Normalized Power: 112W, Intensity Factor: 0.68
- Training Stress Score: 45.2
- Power spikes detected: 3 intervals >200W

**Speed & Cadence:**
- Average speed: 20.1 km/h (Max: 39.3 km/h)
- Average cadence: 83.8 rpm (Range: 45-105 rpm)
- Moving time: 92.1% of total time

**Statistical Summary:**
- Total data points: 5,230
- Data quality: Complete (no gaps detected)
- Stream types: time, distance, heartrate, watts, cadence
```

### Formatting Principles

1. **Markdown Structure**: Use proper markdown headers (##, ###), bullet points, and formatting
2. **Emoji Usage**: Appropriate activity type emojis (🏃 🚴 🏊 📊 ⚡ 💓)
3. **Clear Hierarchy**: Use markdown headers, bullet points, and spacing for readability
4. **Contextual Units**: Always include appropriate units (km, bpm, W, etc.)
5. **Meaningful Groupings**: Group related metrics together logically with markdown sections
6. **Factual Presentation**: Present data and statistics without interpretive insights (except in ai-summary mode)
7. **Flexible Detail Levels**: Allow LLM to request more or less detail as needed
8. **Consistent Formatting**: Use consistent markdown patterns across all tool outputs

## Performance Considerations

1. **Lazy Processing**: Only process when size limits are exceeded
2. **Efficient Algorithms**: Use optimized statistical calculations
3. **Memory Streaming**: Process large datasets without loading entirely into memory
4. **Caching**: Cache derived features for repeated access
5. **Cleanup**: Automatic cleanup of expired paginated streams
6. **Format Caching**: Cache formatted outputs to avoid re-formatting identical data