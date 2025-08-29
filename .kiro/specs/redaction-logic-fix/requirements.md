# Requirements Document

## Introduction

This feature addresses an issue in the current redaction logic where tool call results are being redacted unnecessarily. The system should only redact tool call results when there are subsequent non-tool call messages in the conversation. If a tool call result is followed only by other tool calls or is the final message, it should not be redacted.

## Requirements

### Requirement 1

**User Story:** As a user interacting with the AI system, I want to see complete tool call results when they are not followed by non-tool call messages, so that I can understand the full outcome without unnecessary redaction.

#### Acceptance Criteria

1. WHEN a tool call result is followed by non-tool call messages THEN the system SHALL redact the tool call result
2. WHEN a tool call result is followed only by other tool calls THEN the system SHALL NOT redact the tool call result
3. WHEN a tool call result is the final message in the conversation THEN the system SHALL NOT redact the tool call result
4. WHEN the system processes a single tool call with no subsequent messages THEN the system SHALL NOT redact the response content

### Requirement 2

**User Story:** As a developer maintaining the system, I want clear logic for determining when redaction should be applied based on subsequent message types, so that the behavior is predictable and maintainable.

#### Acceptance Criteria

1. WHEN evaluating redaction rules THEN the system SHALL examine all messages that follow the current tool call result
2. WHEN analyzing subsequent messages THEN the system SHALL distinguish between tool call messages and non-tool call messages
3. WHEN processing streaming responses THEN the system SHALL defer redaction decisions until the message sequence is complete or can be determined
4. WHEN logging redaction decisions THEN the system SHALL record whether redaction was applied based on subsequent non-tool call messages

### Requirement 3

**User Story:** As a user of the AI system, I want consistent behavior where tool call results are only hidden when followed by explanatory text, so that I can see technical details when they're the final output.

#### Acceptance Criteria

1. WHEN a conversation flow ends with a tool call result THEN the system SHALL ensure the result is fully visible
2. WHEN multiple tool calls are chained together with no intervening non-tool call messages THEN the system SHALL show all results without redaction
3. WHEN error conditions occur in tool call sequences THEN the system SHALL not redact error messages unless followed by non-tool call explanatory messages
4. WHEN the system processes concurrent tool calls THEN the system SHALL apply redaction rules individually based on what follows each tool call result

### Requirement 4

**User Story:** As a system administrator, I want to enable or disable redaction functionality using an environment variable, so that I can easily control whether tool call results are ever redacted.

#### Acceptance Criteria

1. WHEN the redaction environment variable is set to false/disabled THEN the system SHALL never redact tool call results regardless of subsequent messages
2. WHEN the redaction environment variable is set to true/enabled THEN the system SHALL apply the sequence-based redaction logic
3. WHEN the redaction environment variable is not set THEN the system SHALL default to enabled redaction behavior
4. WHEN the system starts up THEN the system SHALL read the redaction environment variable and apply the setting for all conversations