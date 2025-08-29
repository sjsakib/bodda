package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"PORT", "DATABASE_URL", "JWT_SECRET", "STRAVA_CLIENT_ID", "STRAVA_CLIENT_SECRET",
		"STRAVA_REDIRECT_URL", "OPENAI_API_KEY", "FRONTEND_URL",
		"STREAM_MAX_CONTEXT_TOKENS", "STREAM_TOKEN_PER_CHAR_RATIO", "STREAM_DEFAULT_PAGE_SIZE",
		"STREAM_MAX_PAGE_SIZE", "STREAM_REDACTION_ENABLED", "STREAM_STRAVA_RESOLUTIONS",
		"STREAM_ENABLE_DERIVED_FEATURES", "STREAM_ENABLE_AI_SUMMARY", "STREAM_ENABLE_PAGINATION",
		"STREAM_ENABLE_AUTO_MODE", "STREAM_LARGE_DATASET_THRESHOLD", "STREAM_CONTEXT_SAFETY_MARGIN",
		"STREAM_MAX_RETRIES", "STREAM_PROCESSING_TIMEOUT",
	}
	
	for _, env := range envVars {
		originalEnv[env] = os.Getenv(env)
		os.Unsetenv(env)
	}
	
	// Restore environment after test
	defer func() {
		for env, value := range originalEnv {
			if value != "" {
				os.Setenv(env, value)
			} else {
				os.Unsetenv(env)
			}
		}
	}()

	// Test with default values
	config := Load()
	
	if config.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", config.Port)
	}
	
	if config.StreamProcessing.MaxContextTokens != 15000 {
		t.Errorf("Expected default max context tokens 15000, got %d", config.StreamProcessing.MaxContextTokens)
	}
	
	if config.StreamProcessing.TokenPerCharRatio != 0.25 {
		t.Errorf("Expected default token per char ratio 0.25, got %f", config.StreamProcessing.TokenPerCharRatio)
	}
	
	if config.StreamProcessing.DefaultPageSize != 1000 {
		t.Errorf("Expected default page size 1000, got %d", config.StreamProcessing.DefaultPageSize)
	}
	
	if config.StreamProcessing.MaxPageSize != 5000 {
		t.Errorf("Expected default max page size 5000, got %d", config.StreamProcessing.MaxPageSize)
	}
	
	if !config.StreamProcessing.RedactionEnabled {
		t.Error("Expected redaction to be enabled by default")
	}
	
	if !config.StreamProcessing.EnableDerivedFeatures {
		t.Error("Expected derived features to be enabled by default")
	}
	
	if !config.StreamProcessing.EnableAISummary {
		t.Error("Expected AI summary to be enabled by default")
	}
	
	if !config.StreamProcessing.EnablePagination {
		t.Error("Expected pagination to be enabled by default")
	}
	
	if !config.StreamProcessing.EnableAutoMode {
		t.Error("Expected auto mode to be enabled by default")
	}
	
	expectedResolutions := []string{"low", "medium", "high"}
	if len(config.StreamProcessing.StravaResolutions) != len(expectedResolutions) {
		t.Errorf("Expected %d Strava resolutions, got %d", len(expectedResolutions), len(config.StreamProcessing.StravaResolutions))
	}
	
	for i, expected := range expectedResolutions {
		if config.StreamProcessing.StravaResolutions[i] != expected {
			t.Errorf("Expected resolution %s at index %d, got %s", expected, i, config.StreamProcessing.StravaResolutions[i])
		}
	}
}

func TestLoadWithCustomEnvironment(t *testing.T) {
	// Set custom environment variables
	os.Setenv("PORT", "9000")
	os.Setenv("STREAM_MAX_CONTEXT_TOKENS", "20000")
	os.Setenv("STREAM_TOKEN_PER_CHAR_RATIO", "0.3")
	os.Setenv("STREAM_DEFAULT_PAGE_SIZE", "2000")
	os.Setenv("STREAM_MAX_PAGE_SIZE", "8000")
	os.Setenv("STREAM_REDACTION_ENABLED", "false")
	os.Setenv("STREAM_STRAVA_RESOLUTIONS", "low,high")
	os.Setenv("STREAM_ENABLE_DERIVED_FEATURES", "false")
	os.Setenv("STREAM_ENABLE_AI_SUMMARY", "false")
	os.Setenv("STREAM_LARGE_DATASET_THRESHOLD", "5000")
	os.Setenv("STREAM_CONTEXT_SAFETY_MARGIN", "1000")
	os.Setenv("STREAM_MAX_RETRIES", "5")
	os.Setenv("STREAM_PROCESSING_TIMEOUT", "60")
	
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("STREAM_MAX_CONTEXT_TOKENS")
		os.Unsetenv("STREAM_TOKEN_PER_CHAR_RATIO")
		os.Unsetenv("STREAM_DEFAULT_PAGE_SIZE")
		os.Unsetenv("STREAM_MAX_PAGE_SIZE")
		os.Unsetenv("STREAM_REDACTION_ENABLED")
		os.Unsetenv("STREAM_STRAVA_RESOLUTIONS")
		os.Unsetenv("STREAM_ENABLE_DERIVED_FEATURES")
		os.Unsetenv("STREAM_ENABLE_AI_SUMMARY")
		os.Unsetenv("STREAM_LARGE_DATASET_THRESHOLD")
		os.Unsetenv("STREAM_CONTEXT_SAFETY_MARGIN")
		os.Unsetenv("STREAM_MAX_RETRIES")
		os.Unsetenv("STREAM_PROCESSING_TIMEOUT")
	}()
	
	config := Load()
	
	if config.Port != "9000" {
		t.Errorf("Expected custom port 9000, got %s", config.Port)
	}
	
	if config.StreamProcessing.MaxContextTokens != 20000 {
		t.Errorf("Expected custom max context tokens 20000, got %d", config.StreamProcessing.MaxContextTokens)
	}
	
	if config.StreamProcessing.TokenPerCharRatio != 0.3 {
		t.Errorf("Expected custom token per char ratio 0.3, got %f", config.StreamProcessing.TokenPerCharRatio)
	}
	
	if config.StreamProcessing.DefaultPageSize != 2000 {
		t.Errorf("Expected custom default page size 2000, got %d", config.StreamProcessing.DefaultPageSize)
	}
	
	if config.StreamProcessing.MaxPageSize != 8000 {
		t.Errorf("Expected custom max page size 8000, got %d", config.StreamProcessing.MaxPageSize)
	}
	
	if config.StreamProcessing.RedactionEnabled {
		t.Error("Expected redaction to be disabled")
	}
	
	if config.StreamProcessing.EnableDerivedFeatures {
		t.Error("Expected derived features to be disabled")
	}
	
	if config.StreamProcessing.EnableAISummary {
		t.Error("Expected AI summary to be disabled")
	}
	
	expectedResolutions := []string{"low", "high"}
	if len(config.StreamProcessing.StravaResolutions) != len(expectedResolutions) {
		t.Errorf("Expected %d Strava resolutions, got %d", len(expectedResolutions), len(config.StreamProcessing.StravaResolutions))
	}
	
	if config.StreamProcessing.LargeDatasetThreshold != 5000 {
		t.Errorf("Expected large dataset threshold 5000, got %d", config.StreamProcessing.LargeDatasetThreshold)
	}
	
	if config.StreamProcessing.ContextSafetyMargin != 1000 {
		t.Errorf("Expected context safety margin 1000, got %d", config.StreamProcessing.ContextSafetyMargin)
	}
	
	if config.StreamProcessing.MaxRetries != 5 {
		t.Errorf("Expected max retries 5, got %d", config.StreamProcessing.MaxRetries)
	}
	
	if config.StreamProcessing.ProcessingTimeout != 60 {
		t.Errorf("Expected processing timeout 60, got %d", config.StreamProcessing.ProcessingTimeout)
	}
}

func TestValidateStreamProcessingConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   StreamProcessingConfig
		expected StreamProcessingConfig
	}{
		{
			name: "invalid page sizes",
			config: StreamProcessingConfig{
				DefaultPageSize: -100,
				MaxPageSize:     -200,
			},
			expected: StreamProcessingConfig{
				DefaultPageSize: 1000,
				MaxPageSize:     5000,
			},
		},
		{
			name: "max page size smaller than default",
			config: StreamProcessingConfig{
				DefaultPageSize: 2000,
				MaxPageSize:     1000,
			},
			expected: StreamProcessingConfig{
				DefaultPageSize: 2000,
				MaxPageSize:     10000, // 5 * DefaultPageSize
			},
		},
		{
			name: "invalid context limits",
			config: StreamProcessingConfig{
				MaxContextTokens:    -1000,
				ContextSafetyMargin: -500,
			},
			expected: StreamProcessingConfig{
				MaxContextTokens:    15000,
				ContextSafetyMargin: 2000, // Default value
			},
		},
		{
			name: "safety margin too large",
			config: StreamProcessingConfig{
				MaxContextTokens:    1000,
				ContextSafetyMargin: 2000,
			},
			expected: StreamProcessingConfig{
				MaxContextTokens:    1000,
				ContextSafetyMargin: 100, // 1000 / 10
			},
		},
		{
			name: "invalid token ratio",
			config: StreamProcessingConfig{
				TokenPerCharRatio: -0.5,
			},
			expected: StreamProcessingConfig{
				TokenPerCharRatio: 0.25,
			},
		},
		{
			name: "token ratio too high",
			config: StreamProcessingConfig{
				TokenPerCharRatio: 1.5,
			},
			expected: StreamProcessingConfig{
				TokenPerCharRatio: 0.25,
			},
		},
		{
			name: "invalid thresholds",
			config: StreamProcessingConfig{
				LargeDatasetThreshold: -1000,
				MaxRetries:           -5,
				ProcessingTimeout:    -30,
			},
			expected: StreamProcessingConfig{
				LargeDatasetThreshold: 10000,
				MaxRetries:           3,
				ProcessingTimeout:    30,
			},
		},
		{
			name: "empty resolutions",
			config: StreamProcessingConfig{
				StravaResolutions: []string{},
			},
			expected: StreamProcessingConfig{
				StravaResolutions: []string{"low", "medium", "high"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{StreamProcessing: tt.config}
			config.validateStreamProcessingConfig()
			
			result := config.StreamProcessing
			
			if tt.expected.DefaultPageSize != 0 && result.DefaultPageSize != tt.expected.DefaultPageSize {
				t.Errorf("Expected DefaultPageSize %d, got %d", tt.expected.DefaultPageSize, result.DefaultPageSize)
			}
			
			if tt.expected.MaxPageSize != 0 && result.MaxPageSize != tt.expected.MaxPageSize {
				t.Errorf("Expected MaxPageSize %d, got %d", tt.expected.MaxPageSize, result.MaxPageSize)
			}
			
			if tt.expected.MaxContextTokens != 0 && result.MaxContextTokens != tt.expected.MaxContextTokens {
				t.Errorf("Expected MaxContextTokens %d, got %d", tt.expected.MaxContextTokens, result.MaxContextTokens)
			}
			
			if tt.expected.ContextSafetyMargin != 0 && result.ContextSafetyMargin != tt.expected.ContextSafetyMargin {
				t.Errorf("Expected ContextSafetyMargin %d, got %d", tt.expected.ContextSafetyMargin, result.ContextSafetyMargin)
			}
			
			if tt.expected.TokenPerCharRatio != 0 && result.TokenPerCharRatio != tt.expected.TokenPerCharRatio {
				t.Errorf("Expected TokenPerCharRatio %f, got %f", tt.expected.TokenPerCharRatio, result.TokenPerCharRatio)
			}
			
			if tt.expected.LargeDatasetThreshold != 0 && result.LargeDatasetThreshold != tt.expected.LargeDatasetThreshold {
				t.Errorf("Expected LargeDatasetThreshold %d, got %d", tt.expected.LargeDatasetThreshold, result.LargeDatasetThreshold)
			}
			
			if tt.expected.MaxRetries != 0 && result.MaxRetries != tt.expected.MaxRetries {
				t.Errorf("Expected MaxRetries %d, got %d", tt.expected.MaxRetries, result.MaxRetries)
			}
			
			if tt.expected.ProcessingTimeout != 0 && result.ProcessingTimeout != tt.expected.ProcessingTimeout {
				t.Errorf("Expected ProcessingTimeout %d, got %d", tt.expected.ProcessingTimeout, result.ProcessingTimeout)
			}
			
			if len(tt.expected.StravaResolutions) > 0 {
				if len(result.StravaResolutions) != len(tt.expected.StravaResolutions) {
					t.Errorf("Expected %d resolutions, got %d", len(tt.expected.StravaResolutions), len(result.StravaResolutions))
				}
				for i, expected := range tt.expected.StravaResolutions {
					if result.StravaResolutions[i] != expected {
						t.Errorf("Expected resolution %s at index %d, got %s", expected, i, result.StravaResolutions[i])
					}
				}
			}
		})
	}
}

func TestGetEnvStringSlice(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue []string
		expected     []string
	}{
		{
			name:         "empty environment variable",
			envValue:     "",
			defaultValue: []string{"default1", "default2"},
			expected:     []string{"default1", "default2"},
		},
		{
			name:         "single value",
			envValue:     "single",
			defaultValue: []string{"default1", "default2"},
			expected:     []string{"single"},
		},
		{
			name:         "multiple values",
			envValue:     "value1,value2,value3",
			defaultValue: []string{"default1"},
			expected:     []string{"value1", "value2", "value3"},
		},
		{
			name:         "values with spaces",
			envValue:     " value1 , value2 , value3 ",
			defaultValue: []string{"default1"},
			expected:     []string{"value1", "value2", "value3"},
		},
		{
			name:         "values with empty items",
			envValue:     "value1,,value3,",
			defaultValue: []string{"default1"},
			expected:     []string{"value1", "value3"},
		},
		{
			name:         "only commas and spaces",
			envValue:     " , , , ",
			defaultValue: []string{"default1", "default2"},
			expected:     []string{"default1", "default2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("TEST_STRING_SLICE", tt.envValue)
			} else {
				os.Unsetenv("TEST_STRING_SLICE")
			}
			
			defer os.Unsetenv("TEST_STRING_SLICE")
			
			result := getEnvStringSlice("TEST_STRING_SLICE", tt.defaultValue)
			
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d items, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected item %s at index %d, got %s", expected, i, result[i])
				}
			}
		})
	}
}

func TestGetEnvHelpers(t *testing.T) {
	// Test getEnvInt
	os.Setenv("TEST_INT", "12345")
	defer os.Unsetenv("TEST_INT")
	
	result := getEnvInt("TEST_INT", 999)
	if result != 12345 {
		t.Errorf("Expected getEnvInt to return 12345, got %d", result)
	}
	
	// Test with invalid int
	os.Setenv("TEST_INVALID_INT", "not_a_number")
	defer os.Unsetenv("TEST_INVALID_INT")
	
	result = getEnvInt("TEST_INVALID_INT", 999)
	if result != 999 {
		t.Errorf("Expected getEnvInt to return default 999 for invalid int, got %d", result)
	}
	
	// Test getEnvFloat
	os.Setenv("TEST_FLOAT", "3.14159")
	defer os.Unsetenv("TEST_FLOAT")
	
	floatResult := getEnvFloat("TEST_FLOAT", 2.71)
	if floatResult != 3.14159 {
		t.Errorf("Expected getEnvFloat to return 3.14159, got %f", floatResult)
	}
	
	// Test getEnvBool
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")
	
	boolResult := getEnvBool("TEST_BOOL", false)
	if !boolResult {
		t.Errorf("Expected getEnvBool to return true, got %t", boolResult)
	}
	
	os.Setenv("TEST_BOOL", "false")
	boolResult = getEnvBool("TEST_BOOL", true)
	if boolResult {
		t.Errorf("Expected getEnvBool to return false, got %t", boolResult)
	}
}