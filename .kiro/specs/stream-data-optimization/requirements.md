# Requirements Document

## Introduction

This feature addresses the challenge of handling large stream tool outputs that exceed the LLM's context window limits. When stream tools generate outputs too large to fit in context, the system will provide three distinct processing options: derived feature extraction, AI-powered summarization, and paginated reading. This ensures the LLM can always work with stream data regardless of size constraints.

## Requirements

### Requirement 1

**User Story:** As an LLM processing stream tool outputs, I want the system to automatically detect when output exceeds my context window, so that I can receive processed data instead of truncated or failed responses.

#### Acceptance Criteria

1. WHEN a stream tool output exceeds the context window limit THEN the system SHALL detect this condition and prevent direct output
2. WHEN the output size limit is exceeded THEN the system SHALL return a notification message like "Output too long, use summary or different mode"
3. WHEN presenting the notification THEN the system SHALL clearly list the three available processing options: derived features, AI summarization, and paginated reading
4. IF the system detects oversized output THEN it SHALL replace the original output with the notification message and processing options

### Requirement 2

**User Story:** As an LLM analyzing stream data, I want to receive derived features from large datasets (inflection points, min, max, variability, spikes, trends), so that I can understand key patterns without processing the full dataset.

#### Acceptance Criteria

1. WHEN the user selects derived features option THEN the system SHALL analyze the stream data for statistical patterns
2. WHEN processing for derived features THEN the system SHALL extract inflection points, minimum values, maximum values, variability metrics, spike detection, and trend analysis, per lap/kilometer summary etc whichever is appropriate for the specific stream
3. WHEN derived features are calculated THEN the system SHALL provide a compressed sample of representative data points
4. WHEN presenting derived features THEN the system SHALL include confidence levels and statistical significance where applicable

### Requirement 2A

**User Story:** As an LLM analyzing activity data, I want to receive lap-by-lap analysis of stream data using the activity's lap information, so that I can understand performance patterns and variations across different segments of the activity.

#### Acceptance Criteria

1. WHEN processing stream data with derived features THEN the system SHALL retrieve lap information from the activity detail API if available
2. WHEN lap data is available THEN the system SHALL segment stream data according to lap boundaries using start_index and end_index
3. WHEN presenting lap-by-lap analysis THEN the system SHALL provide statistical summaries for each lap including pace, heart rate, power, elevation, and other available metrics
4. WHEN lap boundaries are used THEN the system SHALL calculate lap-specific derived features (min, max, average, trends) for each stream type
5. WHEN no lap data is available THEN the system SHALL fallback to distance-based segmentation (per kilometer or mile)
6. WHEN presenting lap analysis THEN the system SHALL include lap comparison metrics showing relative performance across laps

### Requirement 3

**User Story:** As an LLM working with complex data streams, I want the system to summarize large datasets using a smaller AI model, so that I can quickly understand the key insights without processing the full stream.

#### Acceptance Criteria

1. WHEN the LLM selects AI summarization option THEN the system SHALL use a smaller, faster model for initial processing
2. WHEN generating summaries THEN the main LLM SHALL provide the summarization prompt to ensure consistency with the current context
3. WHEN preparing data for AI summarization THEN the system SHALL include the complete stream data, not just statistical summaries, to enable detailed analysis
4. WHEN AI summarization is complete THEN the system SHALL present key insights, patterns, and actionable information
5. IF summarization fails THEN the system SHALL fallback to alternative processing methods

### Requirement 4

**User Story:** As an LLM analyzing large datasets, I want to read stream data in manageable pages, so that I can process information incrementally without overwhelming my context window.

#### Acceptance Criteria

1. WHEN the LLM selects paginated reading THEN the system SHALL divide the stream into logical page boundaries
2. WHEN presenting paginated data THEN the system SHALL provide controls for the LLM to request specific pages
3. WHEN on any page THEN the LLM SHALL be able to see the current position and total page count
4. WHEN requesting pages THEN the system SHALL maintain stream context and allow jumping to specific sections
5. WHEN presenting each page THEN the system SHALL instruct the LLM to make conclusions about the current page before requesting the next page
6. WHEN the LLM moves to a new page THEN the system SHALL redact previous page content as per redaction requirements

### Requirement 5

**User Story:** As an LLM working with stream data, I want the system to have appropriate thresholds configured for stream processing, so that I receive optimized processing based on my context window limits and capabilities.

#### Acceptance Criteria

1. WHEN the system is configured THEN it SHALL have appropriate context window size limits set for optimal LLM processing
2. WHEN processing stream data THEN the system SHALL use validated threshold values that ensure reliable operation
3. IF threshold values are misconfigured THEN the system SHALL handle this gracefully and use safe defaults
4. WHEN the LLM requests stream processing THEN the system SHALL apply current threshold settings without delay

### Requirement 6

**User Story:** As an LLM using any processing option, I want clear error notifications when methods fail, so that I can decide which alternative approach to use.

#### Acceptance Criteria

1. WHEN any processing method fails THEN the system SHALL notify the LLM of the failure with specific error details
2. WHEN a processing method fails THEN the system SHALL provide information about available alternative methods
3. WHEN notifying of failure THEN the system SHALL include relevant context about the data size and processing constraints
4. IF no processing is possible THEN the system SHALL provide the LLM with clear information about why processing failed and suggest alternative approaches

### Requirement 7

**User Story:** As an LLM working with stream data, I want previous stream tool outputs to be redacted from context when processing new streams, so that token usage is optimized and context window limits are respected.

#### Acceptance Criteria

1. WHEN a new stream tool is called THEN the system SHALL redact outputs from previous stream tool calls only
2. WHEN redacting previous stream outputs THEN the system SHALL use simple redaction without complex analysis
3. WHEN stream tool redaction occurs THEN the system SHALL leave other tool outputs unchanged in context
4. IF multiple previous stream outputs exist THEN the system SHALL redact all of them using straightforward replacement

### Requirement 8

**User Story:** As an LLM using any Strava tool, I want all tool outputs to be formatted in human-readable text instead of raw JSON, so that I can easily understand and work with the data without parsing complex structures.

#### Acceptance Criteria

1. WHEN any Strava tool returns data THEN the system SHALL format the output as human-readable text with appropriate emojis and formatting
2. WHEN formatting tool outputs THEN the system SHALL include all relevant data fields available from the Strava API
3. WHEN presenting activity data THEN the system SHALL use clear labels, proper units, and logical grouping of information
4. WHEN stream data is processed THEN the system SHALL present derived features and statistics in readable format with context and explanations
