# Requirements Document

## Introduction

This feature enhances the user experience by displaying session start times as human-readable session titles in the frontend. Instead of generic session identifiers, users will see meaningful timestamps that help them quickly identify and navigate between different chat sessions based on when they were created.

## Requirements

### Requirement 1

**User Story:** As a user, I want to see session start times as session titles, so that I can easily identify and distinguish between different chat sessions based on when they were created.

#### Acceptance Criteria

1. WHEN a session is displayed in the session sidebar THEN the frontend SHALL format and show the session start time as "DD MMM, HH:MM am/pm" (e.g., "3 Sep, 08:20 pm")
2. WHEN multiple sessions exist THEN the frontend SHALL display each session with its respective start time formatted as the display title
3. WHEN rendering session titles THEN the frontend SHALL use the existing session creation timestamp from the backend without storing additional title data
4. WHEN the session list is rendered THEN the frontend SHALL use the local timezone for displaying the timestamp

### Requirement 2

**User Story:** As a user, I want consistent timestamp formatting across all sessions, so that I can easily scan and compare session creation times.

#### Acceptance Criteria

1. WHEN displaying session timestamps THEN the frontend SHALL use a consistent date format across all sessions
2. WHEN the date is from the current year THEN the frontend SHALL omit the year from the display (e.g., "3 Sep, 08:20 pm")
3. WHEN the date is from a previous year THEN the frontend SHALL include the year in the format "DD MMM YYYY, HH:MM am/pm"
4. WHEN displaying time THEN the frontend SHALL use 12-hour format with am/pm indicators

### Requirement 3

**User Story:** As a user, I want the session titles to be readable and accessible, so that I can navigate sessions efficiently regardless of my device or accessibility needs.

#### Acceptance Criteria

1. WHEN session titles are displayed THEN the frontend SHALL ensure adequate contrast and readability
2. WHEN using screen readers THEN the frontend SHALL provide appropriate aria-labels that include the full timestamp information
3. WHEN session titles are too long for the available space THEN the frontend SHALL handle text overflow gracefully without breaking the layout
4. WHEN hovering over a session title THEN the frontend SHALL optionally show a tooltip with the full timestamp if truncated