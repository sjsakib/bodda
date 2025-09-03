# Implementation Plan

- [x] 1. Create core stream processing infrastructure

  - Create `StreamProcessor` interface and basic implementation
  - Add stream size detection logic with configurable thresholds
  - Implement processing mode dispatcher
  - _Requirements: 1.1, 1.2, 1.3_

- [x] 2. Implement output formatting system

  - [x] 2.1 Create `OutputFormatter` interface and implementation

    - Write formatter for athlete profile data with emojis and readable markdown structure
    - Write formatter for activity lists with concise summary format in markdown
    - Write formatter for detailed activity information with comprehensive metrics in markdown
    - Use consistent markdown formatting across all tool outputs
    - _Requirements: 8.1, 8.2, 8.3_

  - [x] 2.2 Update existing tool execution methods to use formatters
    - Modify `executeGetAthleteProfile` to return formatted text instead of JSON
    - Modify `executeGetRecentActivities` to return formatted activity list
    - Modify `executeGetActivityDetails` to return formatted activity details
    - _Requirements: 8.1, 8.2, 8.3_

- [-] 3. Implement derived features processing mode

  - [x] 3.1 Create statistical analysis functions

    - Write functions to calculate min, max, mean, median, std dev for all numeric streams
    - Write functions to calculate quartiles and variability metrics
    - Write boolean statistics calculator for moving time data
    - Write location statistics calculator for GPS coordinates
    - _Requirements: 2.1, 2.2, 2.3_

  - [x] 3.2 Create feature extraction algorithms

    - Implement inflection point detection for all numeric streams
    - Implement spike detection using statistical thresholds
    - Implement trend analysis with moving averages
    - Implement elevation gain/loss calculation from altitude data
    - Implement normalized power calculation for cycling activities
    - Implement heart rate drift calculation for endurance analysis
    - Implement multi-metric correlation analysis
    - _Requirements: 2.1, 2.2, 2.3_

  - [x] 3.3 Create lap-by-lap analysis functionality

    - Write function to retrieve activity detail with lap information from Strava API
    - Implement stream data segmentation using lap start_index and end_index boundaries
    - Create lap-specific statistical calculations for all stream types
    - Implement lap comparison metrics (fastest, slowest, most consistent)
    - Create fallback distance-based segmentation when lap data unavailable
    - Write lap progression analysis to detect pacing and fatigue patterns
    - _Requirements: 2A.1, 2A.2, 2A.3, 2A.4, 2A.5, 2A.6_

  - [x] 3.4 Create derived features formatter
    - Write formatter for comprehensive stream analysis with factual data presentation
    - Include clear metric labels and statistical values without interpretive insights
    - Format statistics in readable markdown sections with appropriate emojis
    - Use markdown headers, bullet points, and formatting for clear structure
    - Focus on data presentation rather than coaching interpretations
    - Add lap-by-lap analysis formatting with lap summaries and comparisons
    - _Requirements: 2.3, 2A.6, 8.4_

- [x] 4. Implement AI-powered summarization mode

  - [x] 4.1 Create summary processor interface and implementation

    - Write function to prepare stream data for AI summarization
    - Implement custom prompt handling for targeted summaries
    - Return raw AI summary output without additional formatting
    - **Fix data preparation to include full stream data instead of just statistics**
    - _Requirements: 3.1, 3.2, 3.3_

  - [x] 4.2 Integrate with existing AI service
    - Add summarization capability to AI service using existing OpenAI client
    - Handle summarization errors and fallback to derived features
    - Return AI-generated summaries without additional formatting (AI provides its own structure)
    - _Requirements: 3.1, 3.2, 3.3_

- [x] 5. Implement unified paginated stream processing

  - [x] 5.1 Create core pagination logic for all stream requests

    - Implement dynamic page size calculation based on available context window size
    - Create functions to estimate optimal page size considering current conversation context
    - Create functions to estimate total pages using Strava resolution parameters
    - Write logic to handle negative page size as indicator for full dataset request
    - Write logic to request specific data chunks from Strava API
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [x] 5.2 Create unified stream processor for all modes
    - Write function to handle all stream requests through pagination interface
    - Implement processing mode application to paginated data (raw, derived, ai-summary)
    - Implement page navigation with clear instructions for LLM
    - Create formatter for paginated results with context and navigation info in markdown
    - Handle full dataset requests (negative page size) with appropriate processing
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6_

- [x] 6. Implement context optimization and redaction

  - [x] 6.1 Create context manager for stream output redaction

    - Write function to identify previous stream tool outputs in conversation history
    - Implement simple redaction logic that replaces content while preserving structure
    - Ensure redaction only affects stream tool outputs, not other tools
    - _Requirements: 7.1, 7.2, 7.3, 7.4_

  - [x] 6.2 Integrate redaction into AI service processing
    - Modify iterative processor to apply redaction before new stream tool calls
    - Add redaction to both streaming and synchronous processing paths
    - Test redaction with multiple consecutive stream tool calls
    - _Requirements: 7.1, 7.2, 7.3, 7.4_

- [x] 7. Update get-activity-streams tool definition and execution

  - [x] 7.1 Update tool definition with new parameters

    - Add processing_mode parameter with auto, raw, derived, ai-summary options
    - Add page_number and page_size parameters for pagination
    - Add summary_prompt parameter for custom AI summarization
    - Update tool description to explain processing options
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

  - [x] 7.2 Enhance executeGetActivityStreams method

    - Implement unified pagination approach for all stream requests
    - Add logic to handle negative page size as full dataset indicator
    - Implement processing mode routing to appropriate processors
    - Handle all new parameters and validation (including required summary_prompt for ai-summary mode)
    - Integrate lap data retrieval from activity detail API when processing derived features
    - Return formatted markdown output instead of raw JSON
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 2A.1, 5.1, 5.2_

  - [x] 7.3 Create processing mode guidance for LLM

    - Add clear descriptions of each processing mode in tool responses
    - Provide usage examples and recommendations for when to use each mode
    - Include mode descriptions in auto mode when presenting options
    - Format mode descriptions in LLM-friendly text with practical guidance
    - _Requirements: 1.2, 1.3, 1.4_

  - [x] 7.4 Provide token estimation hints for pagination decisions
    - Calculate and display estimated token usage for different page sizes
    - Show relationship between page size and context consumption
    - Provide recommendations for optimal page sizes based on current context usage
    - Include token estimates in processing mode options presentation
    - _Requirements: 4.1, 4.2, 4.5_

- [x] 8. Implement error handling and fallback mechanisms

  - [x] 8.1 Create comprehensive error handling

    - Define specific error types for stream processing failures
    - Implement error notification with clear failure details
    - Provide information about available alternative processing methods
    - Include context about data size and processing constraints in error messages
    - _Requirements: 6.1, 6.2, 6.3, 6.4_

  - [x] 8.2 Create fallback processing logic
    - Implement automatic fallback from failed processing modes
    - Ensure LLM always receives useful information even when processing fails
    - Create fallback formatters for basic stream information
    - _Requirements: 6.1, 6.2, 6.3, 6.4_

- [x] 9. Add configuration and testing infrastructure

  - [x] 9.1 Create configuration system

    - Add stream processing configuration to application config
    - Include configurable thresholds for context limits and page sizes
    - Add feature flags for different processing modes
    - _Requirements: 1.1, 1.2, 1.3_

  - [x] 9.2 Write comprehensive unit tests
    - Test stream size detection with various data sizes
    - Test all statistical calculation functions for accuracy
    - Test feature extraction algorithms with sample stream data
    - Test output formatting for all tool types
    - Test error handling and fallback scenarios
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 6.1, 8.1_

- [x] 10. Integration testing and optimization

  - [x] 10.1 Create integration tests

    - Test end-to-end stream processing workflow
    - Test context redaction across multiple tool calls
    - Test pagination with real Strava API responses
    - Test AI summarization integration
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 7.1_

  - [x] 10.2 Performance optimization and monitoring
    - Add performance metrics for processing time and memory usage
    - Optimize statistical calculations for large datasets
    - Add logging for stream processing operations
    - Test with maximum expected stream sizes
    - _Requirements: 1.1, 2.1, 3.1, 4.1_
