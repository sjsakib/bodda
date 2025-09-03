package services

import (
	"context"
	"testing"

	"bodda/internal/models"
)

// Mock implementations for testing

type mockUnifiedStravaService struct{}

func (m *mockUnifiedStravaService) GetAthleteProfile(user *models.User) (*StravaAthleteWithZones, error) {
	return nil, nil
}

func (m *mockUnifiedStravaService) GetAthleteZones(user *models.User) (*StravaAthleteZones, error) {
	return nil, nil
}

func (m *mockUnifiedStravaService) GetActivities(user *models.User, params ActivityParams) ([]*StravaActivity, error) {
	return nil, nil
}

func (m *mockUnifiedStravaService) GetActivityDetail(user *models.User, activityID int64) (*StravaActivityDetail, error) {
	return &StravaActivityDetail{
		Laps: []StravaLap{
			{
				LapIndex:   0,
				StartIndex: 0,
				EndIndex:   100,
				Distance:   1000,
				ElapsedTime: 300,
			},
		},
	}, nil
}

func (m *mockUnifiedStravaService) GetActivityDetailWithZones(user *models.User, activityID int64) (*StravaActivityDetailWithZones, error) {
	return &StravaActivityDetailWithZones{
		StravaActivityDetail: &StravaActivityDetail{
			Laps: []StravaLap{
				{
					LapIndex:   0,
					StartIndex: 0,
					EndIndex:   100,
					Distance:   1000,
					ElapsedTime: 300,
				},
			},
		},
		Zones: nil, // No zones for this mock
	}, nil
}

func (m *mockUnifiedStravaService) GetActivityStreams(user *models.User, activityID int64, streamTypes []string, resolution string) (*StravaStreams, error) {
	// Return mock stream data
	return &StravaStreams{
		Time:      []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		Heartrate: []int{120, 125, 130, 135, 140, 145, 150, 155, 160, 165},
		Watts:     []int{100, 110, 120, 130, 140, 150, 160, 170, 180, 190},
		Distance:  []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90},
	}, nil
}

func (m *mockUnifiedStravaService) GetActivityZones(user *models.User, activityID int64) (*StravaActivityZones, error) {
	return nil, nil
}

func (m *mockUnifiedStravaService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	return nil, nil
}

type mockUnifiedSummaryProcessor struct{}

func (m *mockUnifiedSummaryProcessor) GenerateSummary(ctx context.Context, data *StravaStreams, activityID int64, prompt string) (*StreamSummary, error) {
	return &StreamSummary{
		ActivityID:    activityID,
		SummaryPrompt: prompt,
		Summary:       "Mock AI summary of the stream data",
		TokensUsed:    100,
		Model:         "test-model",
	}, nil
}

func (m *mockUnifiedSummaryProcessor) PrepareStreamDataForSummarization(data *StravaStreams) (string, error) {
	return "prepared data", nil
}

func TestUnifiedStreamProcessor_ProcessPaginatedStreamRequest(t *testing.T) {
	// Setup
	config := &StreamConfig{
		MaxContextTokens:  15000,
		TokenPerCharRatio: 0.25,
		DefaultPageSize:   1000,
		MaxPageSize:       5000,
	}

	mockStrava := &mockUnifiedStravaService{}
	mockSummary := &mockUnifiedSummaryProcessor{}
	derivedProcessor := NewDerivedFeaturesProcessor()
	outputFormatter := NewOutputFormatter()

	processor := NewUnifiedStreamProcessor(
		config,
		mockStrava,
		derivedProcessor,
		mockSummary,
		outputFormatter,
	)

	user := &models.User{ID: "test-user"}

	tests := []struct {
		name           string
		request        *PaginatedStreamRequest
		expectError    bool
		expectedMode   string
	}{
		{
			name: "Raw mode processing",
			request: &PaginatedStreamRequest{
				ActivityID:     12345,
				StreamTypes:    []string{"time", "heartrate", "watts"},
				Resolution:     "medium",
				ProcessingMode: "raw",
				PageNumber:     1,
				PageSize:       5,
			},
			expectError:  false,
			expectedMode: "raw",
		},
		{
			name: "Derived mode processing",
			request: &PaginatedStreamRequest{
				ActivityID:     12345,
				StreamTypes:    []string{"time", "heartrate", "watts"},
				Resolution:     "medium",
				ProcessingMode: "derived",
				PageNumber:     1,
				PageSize:       5,
			},
			expectError:  false,
			expectedMode: "derived",
		},
		{
			name: "AI summary mode processing",
			request: &PaginatedStreamRequest{
				ActivityID:     12345,
				StreamTypes:    []string{"time", "heartrate", "watts"},
				Resolution:     "medium",
				ProcessingMode: "ai-summary",
				PageNumber:     1,
				PageSize:       5,
				SummaryPrompt:  "Analyze the heart rate trends",
			},
			expectError:  false,
			expectedMode: "ai-summary",
		},
		{
			name: "Full dataset request",
			request: &PaginatedStreamRequest{
				ActivityID:     12345,
				StreamTypes:    []string{"time", "heartrate", "watts"},
				Resolution:     "medium",
				ProcessingMode: "raw",
				PageNumber:     1,
				PageSize:       -1, // Full dataset
			},
			expectError:  false,
			expectedMode: "raw",
		},
		{
			name: "Invalid activity ID",
			request: &PaginatedStreamRequest{
				ActivityID:     0,
				StreamTypes:    []string{"time", "heartrate"},
				ProcessingMode: "raw",
				PageNumber:     1,
				PageSize:       100,
			},
			expectError: true,
		},
		{
			name: "Missing summary prompt for AI mode",
			request: &PaginatedStreamRequest{
				ActivityID:     12345,
				StreamTypes:    []string{"time", "heartrate"},
				ProcessingMode: "ai-summary",
				PageNumber:     1,
				PageSize:       100,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.ProcessPaginatedStreamRequest(user, tt.request, 5000)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected result but got nil")
				return
			}

			if result.ProcessingMode != tt.expectedMode {
				t.Errorf("Expected processing mode %s, got %s", tt.expectedMode, result.ProcessingMode)
			}

			if result.ActivityID != tt.request.ActivityID {
				t.Errorf("Expected activity ID %d, got %d", tt.request.ActivityID, result.ActivityID)
			}

			if tt.request.PageSize > 0 {
				if result.PageNumber != tt.request.PageNumber {
					t.Errorf("Expected page number %d, got %d", tt.request.PageNumber, result.PageNumber)
				}
			}

			// Verify data is present
			if result.Data == nil {
				t.Error("Expected data but got nil")
			}

			// Verify instructions are present
			if result.Instructions == "" {
				t.Error("Expected instructions but got empty string")
			}
		})
	}
}

func TestUnifiedStreamProcessor_ValidateRequest(t *testing.T) {
	config := &StreamConfig{
		MaxContextTokens:  15000,
		TokenPerCharRatio: 0.25,
		DefaultPageSize:   1000,
		MaxPageSize:       5000,
	}

	processor := &UnifiedStreamProcessor{config: config}

	tests := []struct {
		name        string
		request     *PaginatedStreamRequest
		expectError bool
	}{
		{
			name: "Valid request",
			request: &PaginatedStreamRequest{
				ActivityID:     12345,
				StreamTypes:    []string{"time", "heartrate"},
				ProcessingMode: "raw",
				PageNumber:     1,
				PageSize:       100,
			},
			expectError: false,
		},
		{
			name: "Invalid activity ID",
			request: &PaginatedStreamRequest{
				ActivityID:     0,
				StreamTypes:    []string{"time", "heartrate"},
				ProcessingMode: "raw",
				PageNumber:     1,
				PageSize:       100,
			},
			expectError: true,
		},
		{
			name: "Empty stream types",
			request: &PaginatedStreamRequest{
				ActivityID:     12345,
				StreamTypes:    []string{},
				ProcessingMode: "raw",
				PageNumber:     1,
				PageSize:       100,
			},
			expectError: true,
		},
		{
			name: "Invalid page number",
			request: &PaginatedStreamRequest{
				ActivityID:     12345,
				StreamTypes:    []string{"time", "heartrate"},
				ProcessingMode: "raw",
				PageNumber:     0,
				PageSize:       100,
			},
			expectError: true,
		},
		{
			name: "Page size too large",
			request: &PaginatedStreamRequest{
				ActivityID:     12345,
				StreamTypes:    []string{"time", "heartrate"},
				ProcessingMode: "raw",
				PageNumber:     1,
				PageSize:       10000,
			},
			expectError: true,
		},
		{
			name: "Invalid processing mode",
			request: &PaginatedStreamRequest{
				ActivityID:     12345,
				StreamTypes:    []string{"time", "heartrate"},
				ProcessingMode: "invalid",
				PageNumber:     1,
				PageSize:       100,
			},
			expectError: true,
		},
		{
			name: "AI summary without prompt",
			request: &PaginatedStreamRequest{
				ActivityID:     12345,
				StreamTypes:    []string{"time", "heartrate"},
				ProcessingMode: "ai-summary",
				PageNumber:     1,
				PageSize:       100,
			},
			expectError: true,
		},
		{
			name: "AI summary with prompt",
			request: &PaginatedStreamRequest{
				ActivityID:     12345,
				StreamTypes:    []string{"time", "heartrate"},
				ProcessingMode: "ai-summary",
				PageNumber:     1,
				PageSize:       100,
				SummaryPrompt:  "Test prompt",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.validateRequest(tt.request)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestUnifiedStreamProcessor_FormatPaginatedResult(t *testing.T) {
	config := &StreamConfig{
		MaxContextTokens:  15000,
		TokenPerCharRatio: 0.25,
		DefaultPageSize:   1000,
		MaxPageSize:       5000,
	}

	processor := &UnifiedStreamProcessor{config: config}

	page := &StreamPage{
		ActivityID:      12345,
		PageNumber:      1,
		TotalPages:      3,
		ProcessingMode:  "raw",
		Data:            "Test stream data content",
		TimeRange:       TimeRange{StartTime: 0, EndTime: 100},
		Instructions:    "Test instructions",
		HasNextPage:     true,
		EstimatedTokens: 1500,
	}

	result := processor.FormatPaginatedResult(page)

	if result == "" {
		t.Error("Expected formatted result but got empty string")
	}

	// Check that key information is included
	if !unifiedContains(result, "Test instructions") {
		t.Error("Expected instructions to be included")
	}

	if !unifiedContains(result, "Test stream data content") {
		t.Error("Expected data content to be included")
	}

	if !unifiedContains(result, "1500 estimated tokens") {
		t.Error("Expected token estimate to be included")
	}

	if !unifiedContains(result, "page 2") {
		t.Error("Expected next page information")
	}
}

// Helper function for string contains check
func unifiedContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		unifiedFindSubstring(s, substr))))
}

func unifiedFindSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}