# Requirements Document

## Introduction

The codebase currently has duplicate tool definitions in two separate files: `internal/services/ai.go` and `internal/services/tool_registry_impl.go`. This duplication creates maintenance overhead, potential inconsistencies, and violates the DRY (Don't Repeat Yourself) principle. We need to consolidate these definitions into a single source of truth.

## Requirements

### Requirement 1

**User Story:** As a developer, I want tool definitions to exist in only one place, so that I don't have to maintain duplicate code and risk inconsistencies.

#### Acceptance Criteria

1. WHEN I examine the codebase THEN there SHALL be only one source of truth for tool definitions
2. WHEN I need to modify a tool definition THEN I SHALL only need to update it in one location
3. WHEN the system initializes THEN both the AI service and tool registry SHALL use the same tool definitions

### Requirement 2

**User Story:** As a developer, I want the tool registry to be the authoritative source for tool definitions, so that all tool-related functionality uses consistent schemas and metadata.

#### Acceptance Criteria

1. WHEN the AI service needs tool definitions THEN it SHALL retrieve them from the tool registry
2. WHEN the tool registry initializes THEN it SHALL contain all available tool definitions with complete schemas
3. WHEN a tool is executed THEN both the AI service and tool executor SHALL use the same tool definition

### Requirement 3

**User Story:** As a developer, I want the refactoring to maintain backward compatibility, so that existing functionality continues to work without changes.

#### Acceptance Criteria

1. WHEN the refactoring is complete THEN all existing API endpoints SHALL continue to function identically
2. WHEN tools are executed THEN they SHALL produce the same results as before the refactoring
3. WHEN the system starts up THEN all tool-related functionality SHALL work without any configuration changes

### Requirement 4

**User Story:** As a developer, I want comprehensive tests to validate the consolidation, so that I can be confident the refactoring doesn't break existing functionality.

#### Acceptance Criteria

1. WHEN the refactoring is complete THEN all existing tests SHALL pass
2. WHEN I run integration tests THEN tool execution SHALL work correctly through both the AI service and direct tool execution endpoints
3. WHEN I validate tool schemas THEN they SHALL be consistent between the AI service and tool registry