# Design Document

## Overview

This design implements a system to externalize hardcoded system prompts into separate files that can be managed independently from the codebase. The solution will create a dedicated prompt management system that loads prompts from external files, supports environment-specific configurations, and ensures these files are excluded from git commits.

Based on the current codebase analysis, there are at least two main system prompts that need to be externalized:
1. The main AI coaching prompt in `internal/services/ai.go` (variable `systemPrompt`)
2. The summary processor prompt in `internal/services/summary_processor.go`

## Architecture

### Directory Structure
```
prompts/
├── examples/                    # Template files (committed to git)
│   ├── ai-coach.txt.example
│   ├── summary-processor.txt.example
│   └── README.md
├── ai-coach.txt                # Actual prompt files (ignored by git)
├── summary-processor.txt
├── development/                # Environment-specific overrides
│   ├── ai-coach.txt
│   └── summary-processor.txt
├── production/
│   ├── ai-coach.txt
│   └── summary-processor.txt
└── test/
    ├── ai-coach.txt
    └── summary-processor.txt
```

### Loading Priority
1. Environment-specific file (`prompts/{env}/{prompt-name}.txt`)
2. Default file (`prompts/{prompt-name}.txt`)
3. Fallback to hardcoded prompt (with warning)
4. Error if no prompt found and no fallback available

## Components and Interfaces

### PromptManager Interface
```go
type PromptManager interface {
    LoadPrompt(name string) (string, error)
    ValidatePrompts() error
    GetAvailablePrompts() []string
    ReloadPrompts() error
}
```

### PromptLoader Implementation
```go
type PromptLoader struct {
    baseDir     string
    environment string
    cache       map[string]string
    logger      *log.Logger
}
```

### Configuration Integration
The prompt system will integrate with the existing config system to:
- Set the base prompts directory (default: `./prompts`)
- Define the current environment
- Enable/disable prompt caching
- Configure fallback behavior

## Data Models

### Prompt Configuration
```go
type PromptConfig struct {
    BaseDir     string `yaml:"base_dir" env:"PROMPTS_BASE_DIR" default:"./prompts"`
    Environment string `yaml:"environment" env:"PROMPTS_ENV" default:"development"`
    CacheEnabled bool  `yaml:"cache_enabled" env:"PROMPTS_CACHE_ENABLED" default:"true"`
    FallbackEnabled bool `yaml:"fallback_enabled" env:"PROMPTS_FALLBACK_ENABLED" default:"true"`
}
```

### Prompt Metadata
```go
type PromptInfo struct {
    Name         string
    FilePath     string
    Environment  string
    LastModified time.Time
    Size         int64
    Checksum     string
}
```

## Error Handling

### Error Types
1. **PromptNotFoundError**: When a required prompt file doesn't exist
2. **PromptReadError**: When a prompt file exists but cannot be read
3. **PromptValidationError**: When a prompt file has invalid format/content
4. **PromptCacheError**: When caching operations fail

### Error Recovery
- Missing prompts: Log warning and use fallback if available
- Read errors: Retry once, then fail with detailed error
- Validation errors: Prevent startup and provide specific guidance
- Cache errors: Continue without cache, log warning

### Logging Strategy
- Info: Successful prompt loads, cache hits
- Warn: Fallback usage, cache misses, file not found with fallback
- Error: Read failures, validation failures, missing required prompts
- Debug: File paths, environment resolution, cache operations

## Testing Strategy

### Unit Tests
1. **PromptLoader Tests**
   - Test loading from different environments
   - Test fallback behavior
   - Test caching functionality
   - Test error conditions

2. **Configuration Tests**
   - Test environment variable parsing
   - Test default value handling
   - Test validation logic

3. **Integration Tests**
   - Test with actual file system
   - Test environment switching
   - Test concurrent access

### Test Data Structure
```
testdata/
├── prompts/
│   ├── valid-prompt.txt
│   ├── empty-prompt.txt
│   ├── development/
│   │   └── dev-prompt.txt
│   └── production/
│       └── prod-prompt.txt
└── invalid/
    └── unreadable-prompt.txt
```

### Validation Tests
- Empty prompt files
- Very large prompt files
- Files with special characters
- Files with different encodings
- Missing directory permissions
- Concurrent file access

## Implementation Plan Integration

### Service Modifications
1. **AI Service**: Replace hardcoded `systemPrompt` variable with prompt loader
2. **Summary Processor**: Replace hardcoded prompt with prompt loader
3. **Config Service**: Add prompt configuration support

### Startup Sequence
1. Load configuration (including prompt config)
2. Initialize prompt manager
3. Validate all required prompts exist
4. Cache prompts if enabled
5. Continue with normal service initialization

### Runtime Behavior
- Prompts loaded once at startup (unless reload triggered)
- Cache used for subsequent access
- Environment changes require service restart
- File changes can trigger reload via signal or API

## Security Considerations

### File System Security
- Validate all file paths to prevent directory traversal
- Ensure prompt directory has appropriate permissions
- Log all file access attempts for auditing

### Content Validation
- Validate prompt file size limits (prevent DoS)
- Check for malicious content patterns
- Ensure proper encoding (UTF-8)

### Environment Isolation
- Prevent cross-environment prompt access
- Validate environment names against whitelist
- Secure handling of sensitive prompts in production

## Performance Considerations

### Caching Strategy
- In-memory cache for loaded prompts
- Cache invalidation on file changes (if file watching enabled)
- Configurable cache TTL for development environments

### File I/O Optimization
- Batch load all prompts at startup
- Async reload capability for runtime updates
- Minimal file system calls during normal operation

### Memory Management
- Reasonable limits on prompt file sizes
- Efficient string handling for large prompts
- Garbage collection considerations for cached content