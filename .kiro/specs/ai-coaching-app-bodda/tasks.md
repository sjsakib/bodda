# Implementation Plan

- [x] 1. Set up project structure and core configuration

  - Create Go backend directory structure with main.go, internal packages for services
  - Create React frontend with Vite, TypeScript, and Tailwind CSS setup
  - Set up PostgreSQL database with Docker compose for development
  - Create environment configuration files and basic project documentation
  - _Requirements: All requirements need foundational project structure_

- [x] 2. Implement database schema and models

  - Create PostgreSQL migration files for users, sessions, messages, and athlete_logbooks tables
  - Implement Go structs for User, Session, Message, and AthleteLogbook models
  - Create database connection utilities and migration runner
  - Write unit tests for database models and basic CRUD operations
  - _Requirements: 2.2, 6.1, 6.3, 7.1, 7.3_

- [x] 3. Build Strava OAuth authentication system

  - Implement Strava OAuth flow with redirect handling in Go backend
  - Create authentication middleware for protected routes
  - Build token storage and refresh mechanism for Strava API access
  - Create user registration and login endpoints with JWT session management
  - Write tests for authentication flow and token management
  - _Requirements: 2.1, 2.2, 2.3, 2.4_

- [x] 4. Create Strava API integration service

  - Implement StravaService with methods for athlete profile, activities, and activity details
  - Add activity streams fetching with configurable stream types and resolution
  - Implement rate limiting and error handling for Strava API calls
  - Create token refresh logic when Strava tokens expire
  - Write unit tests with mocked Strava API responses
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [x] 5. Implement athlete logbook management

  - Create LogbookService for CRUD operations on athlete logbooks
  - Implement initial logbook creation from Strava athlete profile data
  - Build logbook update functionality accepting free-form string content
  - Create database queries for efficient logbook retrieval and updates
  - Write tests for logbook operations and string content handling
  - _Requirements: 6.1, 6.2, 6.4, 6.6_

- [x] 6. Build chat session management system

  - Implement ChatService for creating and managing conversation sessions
  - Create session CRUD operations with user association
  - Build message persistence with role-based storage (user/assistant)
  - Implement session history retrieval with pagination support
  - Write tests for session and message management
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 7.1, 7.2, 7.3, 7.4_

- [x] 7. Create OpenAI integration and AI service

  - Implement AIService with OpenAI client for chat completions
  - Build function calling system for Strava tools (get-athlete-profile, get-recent-activities, etc.)
  - Implement logbook update tool accepting string content for AI to structure freely
  - Create message context preparation with conversation history and logbook
  - Add streaming response handling for real-time AI responses
  - Write tests with mocked OpenAI responses and tool executions
  - _Requirements: 3.2, 3.3, 5.1, 5.2, 6.2, 6.6_

- [x] 8. Build backend API endpoints and routing

  - Create REST API endpoints for authentication (/auth/strava, /auth/callback)
  - Implement session management endpoints (/api/sessions, /api/sessions/:id/messages)
  - Build chat messaging endpoint with AI integration (/api/sessions/:id/messages)
  - Create Server-Sent Events endpoint for streaming AI responses
  - Add middleware for authentication, CORS, and request logging
  - Write integration tests for all API endpoints
  - _Requirements: 2.1, 2.3, 3.1, 3.2, 4.1, 4.2_

- [x] 8.1 Implement authentication status endpoint

  - Create /api/auth/check endpoint to verify current authentication status
  - Return authenticated user information when valid token is present
  - Handle unauthenticated requests gracefully with appropriate status codes
  - Add endpoint to existing auth middleware for proper token validation
  - Write tests for authentication status checking functionality
  - _Requirements: 2.3, 2.4_

- [x] 9. Create React frontend landing page

  - Build responsive landing page component with Strava connect button
  - Implement disclaimer display about AI advice usage risks
  - Add routing setup with React Router for navigation
  - Create authentication state management and redirect logic
  - Style with Tailwind CSS for comfortable, readable design
  - Write component tests for landing page functionality
  - _Requirements: 1.1, 1.2, 1.3, 8.1, 8.2, 8.3_

- [x] 10. Implement chat interface frontend

  - Create chat interface component with message input and display
  - Implement real-time message streaming using Server-Sent Events
  - Build markdown rendering for AI responses using react-markdown
  - Add auto-scroll functionality for new messages
  - Create loading states and error handling for chat interactions
  - Write tests for chat component behavior and streaming
  - _Requirements: 3.1, 3.3, 3.4, 8.1, 8.4_

- [x] 11. Build session sidebar and navigation

  - Create session sidebar component displaying user's conversation history
  - Implement session selection and switching functionality
  - Build new session creation with automatic title generation
  - Add session metadata display (creation date, message count)
  - Create responsive design that works on different screen sizes
  - Write tests for session navigation and state management
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 8.4_

- [x] 12. Integrate frontend with backend APIs

  - Connect authentication flow from frontend to backend OAuth endpoints
  - Implement API client for session management and message sending
  - Add error handling and retry logic for network requests
  - Create loading states and user feedback for all async operations
  - Implement proper error display with user-friendly messages
  - Write end-to-end tests for complete user workflows
  - _Requirements: 2.3, 2.4, 3.2, 4.2, 7.2, 8.1_

- [x] 13. Add comprehensive error handling and edge cases

  - Implement graceful handling of Strava API failures and rate limits
  - Add fallback responses when OpenAI service is unavailable
  - Create proper error messages for authentication failures
  - Handle network disconnections and reconnection for streaming
  - Add input validation and sanitization for user messages
  - Write tests for error scenarios and recovery mechanisms
  - _Requirements: 2.4, 5.4, 3.5_

- [x] 14. Create development and deployment configuration

  - Set up Docker containers for backend, frontend, and database
  - Create development environment with hot reloading
  - Add environment-specific configuration management
  - Create database seeding scripts for development and testing
  - Set up basic logging and monitoring for the application
  - Write deployment documentation and setup instructions
  - _Requirements: Supporting infrastructure for all requirements_

- [x] 15. Update logbook implementation to use string-based content

  - Modify LogbookService to accept and store string content instead of structured data
  - Update AI service logbook tool to pass string content directly to LogbookService
  - Simplify database operations to handle string content storage and retrieval
  - Update existing tests to work with string-based logbook content
  - Ensure backward compatibility with existing logbook data if any exists
  - _Requirements: 6.2, 6.4, 6.6_

- [x] 16. Implement multi-turn iterative analysis infrastructure

  - Create IterativeProcessor struct to manage multiple rounds of data analysis
  - Add round tracking and maximum analysis limits (default 5 rounds)
  - Implement context accumulation across analysis rounds
  - Create progress streaming mechanism for coaching-focused status updates
  - Add safeguards to prevent infinite analysis loops
  - Write unit tests for iterative analysis processing logic
  - _Requirements: 9.1, 9.2, 9.3, 9.5_

- [x] 17. Enhance AI service to support iterative data analysis

  - Modify ProcessMessage method to handle multiple rounds of data gathering and analysis
  - Implement conversation context building with accumulated insights from each analysis round
  - Add logic to determine when to continue analysis vs provide final coaching response
  - Create natural progress update streaming during comprehensive analysis
  - Implement graceful error handling for failed data requests during analysis rounds
  - Update existing tests to cover multi-round analysis scenarios
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.6_

- [x] 18. Add user-friendly progress streaming for multi-turn analysis

  - Implement intermediate progress messages using coaching-focused language
  - Create natural status updates ("Reviewing your recent training...", "Analyzing your workout data...", "Looking at your performance trends...")
  - Add progress indicators for comprehensive analysis workflows
  - Ensure all progress messages sound like natural coaching communication
  - Avoid technical jargon like "executing tools" or "making API calls"
  - Write tests for user-friendly progress messaging
  - _Requirements: 9.4_

- [ ] 19. Create comprehensive tests for multi-turn analysis workflows

  - Write integration tests for complete multi-round analysis scenarios
  - Test maximum iteration limits and infinite loop prevention
  - Create test cases for partial data failures during analysis rounds
  - Test user-friendly progress streaming during comprehensive analysis
  - Add performance tests for iterative analysis overhead
  - Create end-to-end tests with realistic coaching analysis workflows
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6_

- [x] 20. Implement automatic Strava token refresh mechanism

  - Modify StravaService methods to accept User objects instead of raw tokens
  - Create executeWithTokenRefresh wrapper method for handling 401 errors automatically
  - Implement token refresh logic that updates database and retries original request
  - Add error handling for failed token refresh (invalid refresh token)
  - Update all Strava API calls to use the new token refresh wrapper
  - Write unit tests for token refresh scenarios and error handling
  - _Requirements: 10.1, 10.2, 10.3, 10.5, 10.6_

- [x] 21. Add logout functionality to chat interface

  - Add logout button to chat i nterface header with appropriate styling
  - Implement logout API endpoint that clears JWT tokens but preserves user data
  - Create frontend logout handler that clears browser tokens and redirects to landing page
  - Update authentication middleware to handle logout requests properly
  - Ensure logout button is positioned prominently but doesn't interfere with chat functionality
  - Verify that re-login via Strava OAuth reconnects to existing user account and data
  - Write tests for logout flow, token cleanup, and re-login data continuity
  - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6_

- [x] 22. Enhance activity details with comprehensive Strava data fields

  - Add all missing fields to StravaActivityDetail struct: resource_state, athlete (id, resource_state), location_city, location_state, location_country, achievement_count, start_latlng, end_latlng, average_cadence, average_temp, weighted_average_watts, device_watts, kilojoules, heartrate_opt_out, display_hide_heartrate_option, upload_id, upload_id_str, external_id, from_accepted_tag, total_photo_count, has_kudoed, suffer_score, calories, perceived_exertion, prefer_perceived_exertion, hide_from_home, device_name
  - Implement splits_standard array with distance, elapsed_time, elevation_difference, moving_time, split, average_speed, average_grade_adjusted_speed, average_heartrate, pace_zone fields
  - Add best_efforts array with id, resource_state, name, activity reference, athlete reference, elapsed_time, moving_time, start_date, start_date_local, distance, pr_rank, achievements, start_index, end_index fields
  - Create similar_activities struct with effort_count, average_speed, min_average_speed, mid_average_speed, max_average_speed, pr_rank, frequency_milestone, trend (speeds array, current_activity_index, min_speed, mid_speed, max_speed, direction), resource_state
  - Add enhanced laps array with all fields: id, resource_state, name, activity reference, athlete reference, elapsed_time, moving_time, start_date, start_date_local, distance, average_speed, max_speed, lap_index, split, start_index, end_index, total_elevation_gain, average_cadence, device_watts, average_watts, average_heartrate, max_heartrate, pace_zone
  - Include available_zones array (heartrate, pace, power) for training zone analysis capabilities
  - Add GetActivityZones method to StravaService to fetch detailed zone data from /activities/{id}/zones API
  - Create StravaActivityZones struct to handle heart rate, power, and pace zone distributions and time spent in each zone
  - Integrate zones data into activity detail responses for comprehensive training zone analysis
  - Update AI service tool responses to format all enhanced activity data in LLM-friendly format with clear labels, units, and structured presentation for optimal coaching analysis
  - Update JSON parsing to handle all new fields with proper type conversion and null handling for optional fields
  - Write comprehensive tests for enhanced activity detail parsing with real Strava API response samples including all new fields
  - _Requirements: 5.2, 5.3 - Enhanced activity data provides richer context for AI coaching analysis_

- [x] 23. Integrate athlete training zones into profile tool

  - Add internal getAthleteZones method to StravaService to fetch athlete's configured training zones from /athlete/zones API
  - Create StravaAthleteZones, StravaZoneSet, and StravaZone structs for zone boundary definitions (min/max values)
  - Modify GetAthleteProfile method to return StravaAthleteWithZones that includes zone data
  - Add zone parsing logic to handle different zone types (heart rate, power, pace) with proper type conversion
  - Update athlete profile tool response to include zone data when available for comprehensive coaching context
  - Handle cases where athletes haven't configured zones with appropriate messaging in profile response
  - Write unit tests for integrated zone data parsing and API integration with mock Strava responses
  - _Requirements: 5A.1, 5A.2, 5A.5 - Integrates athlete zone configuration into existing profile tool_

- [x] 24. Integrate activity zones into activity details tool

  - Add internal getActivityZones method to StravaService to fetch zone distribution data from /activities/{id}/zones API
  - Create StravaActivityZones and StravaZoneDistribution structs for zone time distribution data
  - Modify GetActivityDetail method to return StravaActivityDetailWithZones that includes zone distribution
  - Update activity details tool response to include zone distribution data showing time spent in each training zone
  - Add zone-specific progress messaging during analysis ("Analyzing your training zones...", "Reviewing zone distribution...")
  - Update tool response formatting to present integrated zone and activity data clearly for LLM analysis
  - Write tests for integrated zone tool execution and response formatting with realistic zone data
  - _Requirements: 5A.3, 5A.4 - Integrates zone distribution data into existing activity details tool_

- [ ] 25. Create comprehensive zone analysis workflows with integrated tools
  - Implement zone distribution analysis using integrated athlete profile and activity detail tools
  - Add zone-based training intensity assessment using time spent in each zone from activity details
  - Create zone trend analysis across multiple activities for training load evaluation using existing tools
  - Integrate zone analysis into logbook updates for persistent zone-based coaching insights
  - Add zone-specific coaching recommendations based on training distribution patterns from integrated data
  - Implement zone threshold analysis to identify when athletes are training outside their configured zones
  - Write integration tests for complete zone analysis workflows using enhanced profile and activity tools
  - _Requirements: 5A.3, 5A.4, 5A.6 - Provides comprehensive zone-based coaching analysis using integrated tools_
