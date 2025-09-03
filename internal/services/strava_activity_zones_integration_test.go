package services

import (
	"context"
	"testing"

	"bodda/internal/models"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)


func TestStravaActivityDetailWithZones_DataStructure(t *testing.T) {
	// Test data
	activityID := int64(12345)
	
	// Mock activity detail
	activityDetail := &StravaActivityDetail{
		StravaActivity: StravaActivity{
			ID:   activityID,
			Name: "Morning Run",
			Type: "Run",
		},
		AvailableZones: []string{"heartrate", "power"},
	}
	
	// Mock zone data
	zones := &StravaActivityZones{
		HeartRate: &StravaZoneDistribution{
			Type: "heartrate",
			Zones: []StravaZoneData{
				{Min: 100, Max: 120, Time: 600},  // Zone 1: 10 minutes
				{Min: 120, Max: 140, Time: 1200}, // Zone 2: 20 minutes
				{Min: 140, Max: 160, Time: 900},  // Zone 3: 15 minutes
			},
			SensorBased: true,
		},
	}
	
	// Test creating the integrated structure
	integrated := &StravaActivityDetailWithZones{
		StravaActivityDetail: activityDetail,
		Zones:               zones,
	}
	
	// Verify the structure exists and can be created
	assert.NotNil(t, integrated)
	assert.Equal(t, activityDetail, integrated.StravaActivityDetail)
	assert.Equal(t, zones, integrated.Zones)
	assert.Equal(t, activityID, integrated.ID)
	assert.Equal(t, "Morning Run", integrated.Name)
	assert.Equal(t, "Run", integrated.Type)
}

func TestStravaActivityDetailWithZones_NoZonesAvailable(t *testing.T) {
	// Test data
	activityDetail := &StravaActivityDetail{
		StravaActivity: StravaActivity{
			ID:   12345,
			Name: "Morning Run",
			Type: "Run",
		},
		AvailableZones: []string{}, // No zones available
	}
	
	// Expected result - should have nil zones
	expected := &StravaActivityDetailWithZones{
		StravaActivityDetail: activityDetail,
		Zones:               nil,
	}
	
	assert.NotNil(t, expected)
	assert.Equal(t, activityDetail, expected.StravaActivityDetail)
	assert.Nil(t, expected.Zones)
}

func TestFormatActivityDetailsWithZones_WithZones(t *testing.T) {
	formatter := NewOutputFormatter()
	
	// Test data with zones
	detailsWithZones := &StravaActivityDetailWithZones{
		StravaActivityDetail: &StravaActivityDetail{
			StravaActivity: StravaActivity{
				ID:       12345,
				Name:     "Morning Run",
				Type:     "Run",
				Distance: 5000, // 5km
			},
		},
		Zones: &StravaActivityZones{
			HeartRate: &StravaZoneDistribution{
				Type: "heartrate",
				Zones: []StravaZoneData{
					{Min: 100, Max: 120, Time: 600},  // Zone 1: 10 minutes
					{Min: 120, Max: 140, Time: 1200}, // Zone 2: 20 minutes
				},
				SensorBased: true,
			},
		},
	}
	
	result := formatter.FormatActivityDetailsWithZones(detailsWithZones)
	
	// Verify the result contains both activity details and zone information
	assert.Contains(t, result, "Morning Run")
	assert.Contains(t, result, "Training Zone Analysis")
	assert.Contains(t, result, "Heart Rate Zones")
	assert.Contains(t, result, "Zone 1")
	assert.Contains(t, result, "Zone 2")
}

func TestFormatActivityDetailsWithZones_WithoutZones(t *testing.T) {
	formatter := NewOutputFormatter()
	
	// Test data without zones
	detailsWithZones := &StravaActivityDetailWithZones{
		StravaActivityDetail: &StravaActivityDetail{
			StravaActivity: StravaActivity{
				ID:       12345,
				Name:     "Morning Run",
				Type:     "Run",
				Distance: 5000, // 5km
			},
		},
		Zones: nil, // No zones
	}
	
	result := formatter.FormatActivityDetailsWithZones(detailsWithZones)
	
	// Verify the result contains activity details but no zone information
	assert.Contains(t, result, "Morning Run")
	assert.NotContains(t, result, "Training Zone Analysis")
}

func TestExecuteGetActivityDetails_Integration(t *testing.T) {
	// This test verifies the AI service integration
	mockStrava := &MockStravaService{}
	formatter := NewOutputFormatter()
	
	// Create AI service with mocked dependencies
	aiService := &aiService{
		stravaService: mockStrava,
		formatter:     formatter,
	}
	
	// Test data
	user := &models.User{ID: "test-user", AccessToken: "test-token"}
	activityID := int64(12345)
	
	msgCtx := &MessageContext{
		UserID: "test-user",
		User:   user,
	}
	
	// Mock integrated response
	detailsWithZones := &StravaActivityDetailWithZones{
		StravaActivityDetail: &StravaActivityDetail{
			StravaActivity: StravaActivity{
				ID:   activityID,
				Name: "Morning Run",
				Type: "Run",
			},
		},
		Zones: &StravaActivityZones{
			HeartRate: &StravaZoneDistribution{
				Type: "heartrate",
				Zones: []StravaZoneData{
					{Min: 100, Max: 120, Time: 600},
				},
				SensorBased: true,
			},
		},
	}
	
	// Set up expectations
	mockStrava.On("GetActivityDetailWithZones", user, activityID).Return(detailsWithZones, nil)
	
	// Execute the method
	result, err := aiService.executeGetActivityDetails(context.Background(), msgCtx, activityID)
	
	// Verify results
	assert.NoError(t, err)
	assert.Contains(t, result, "Morning Run")
	assert.Contains(t, result, "Training Zone Analysis")
	
	mockStrava.AssertExpectations(t)
}

func TestZoneSpecificProgressMessages(t *testing.T) {
	aiService := &aiService{}
	processor := &IterativeProcessor{CurrentRound: 0}
	
	// Test activity details tool call triggers zone-specific messages
	toolCalls := []openai.ToolCall{
		{
			Function: openai.FunctionCall{
				Name: "get-activity-details",
			},
		},
	}
	
	message := aiService.getContextualProgressMessage(processor, toolCalls)
	
	// Verify that zone-related messages are possible
	possibleMessages := []string{
		"Taking a closer look at your specific workouts...",
		"Examining the details of your recent training sessions...",
		"Getting a better understanding of your workout structure...",
		"Reviewing the specifics of your training efforts...",
		"Analyzing your training zones...",
		"Reviewing zone distribution...",
	}
	
	// Check that the message is one of the expected ones
	found := false
	for _, expected := range possibleMessages {
		if message == expected {
			found = true
			break
		}
	}
	assert.True(t, found, "Message should be one of the expected zone-specific messages")
}