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
