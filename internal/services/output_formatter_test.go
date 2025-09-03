package services

import (
	"strings"
	"testing"
)

func TestNewOutputFormatter(t *testing.T) {
	formatter := NewOutputFormatter()
	if formatter == nil {
		t.Fatal("NewOutputFormatter() returned nil")
	}
}

func TestFormatAthleteProfile(t *testing.T) {
	formatter := NewOutputFormatter()

	t.Run("nil profile", func(t *testing.T) {
		result := formatter.FormatAthleteProfile(nil)
		expected := "❌ **No athlete profile data available**"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("complete profile", func(t *testing.T) {
		profile := &StravaAthleteWithZones{
			StravaAthlete: &StravaAthlete{
				ID:        12345,
				Username:  "testuser",
				Firstname: "John",
				Lastname:  "Doe",
				City:      "San Francisco",
				State:     "CA",
				Country:   "USA",
				Sex:       "M",
				Premium:   true,
				Summit:    false,
				CreatedAt: "2020-01-15T10:30:00Z",
				UpdatedAt: "2024-01-15T10:30:00Z",
				Weight:    75.5,
				FTP:       250,
			},
			Zones: &StravaAthleteZones{
				HeartRate: &StravaZoneSet{
					CustomZones: false,
					Zones: []StravaZone{
						{Min: 50, Max: 100},
						{Min: 100, Max: 130},
						{Min: 130, Max: 150},
						{Min: 150, Max: 170},
						{Min: 170, Max: 190},
					},
					ResourceState: 3,
				},
			},
		}

		result := formatter.FormatAthleteProfile(profile)

		// Check for key components
		if !strings.Contains(result, "👤 **John Doe** (@testuser)") {
			t.Error("Missing or incorrect header")
		}
		if !strings.Contains(result, "📍 **Location:** San Francisco, CA, USA") {
			t.Error("Missing or incorrect location")
		}
		if !strings.Contains(result, "Premium: ✅ Yes") {
			t.Error("Missing or incorrect premium status")
		}
		if !strings.Contains(result, "Summit: ❌ No") {
			t.Error("Missing or incorrect summit status")
		}
		if !strings.Contains(result, "Weight: 75.5 kg") {
			t.Error("Missing or incorrect weight")
		}
		if !strings.Contains(result, "FTP: 250 watts") {
			t.Error("Missing or incorrect FTP")
		}
		if !strings.Contains(result, "Member since: January 15, 2020") {
			t.Error("Missing or incorrect creation date")
		}
		if !strings.Contains(result, "🎯 **Training Zones:**") {
			t.Error("Missing training zones section")
		}
		if !strings.Contains(result, "Zone 1: 50-100 bpm") {
			t.Error("Missing or incorrect heart rate zone 1")
		}
		if !strings.Contains(result, "Zone 5: 170-190 bpm") {
			t.Error("Missing or incorrect heart rate zone 5")
		}
	})

	t.Run("profile without zones", func(t *testing.T) {
		profile := &StravaAthleteWithZones{
			StravaAthlete: &StravaAthlete{
				ID:        12345,
				Username:  "testuser",
				Firstname: "John",
				Lastname:  "Doe",
				Premium:   true,
			},
			Zones: nil,
		}

		result := formatter.FormatAthleteProfile(profile)

		if !strings.Contains(result, "👤 **John Doe** (@testuser)") {
			t.Error("Missing or incorrect header")
		}
		if !strings.Contains(result, "🎯 **Training Zones:** Not available or not configured") {
			t.Error("Missing or incorrect zones unavailable message")
		}
	})
}

func TestFormatActivities(t *testing.T) {
	formatter := NewOutputFormatter()

	t.Run("empty activities", func(t *testing.T) {
		result := formatter.FormatActivities([]*StravaActivity{})
		expected := "📭 **No recent activities found**"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("single activity", func(t *testing.T) {
		activities := []*StravaActivity{
			{
				ID:             123456789,
				Name:           "Morning Run",
				Type:           "Run",
				SportType:      "running",
				Distance:       5000.0, // 5km
				StartDateLocal: "2024-01-15T07:30:00Z",
			},
		}

		result := formatter.FormatActivities(activities)

		if !strings.Contains(result, "🏃 **Recent Activities** (1 activities)") {
			t.Error("Missing or incorrect header")
		}
		if !strings.Contains(result, "🏃 **Morning Run** (ID: 123456789) — 5.00km on 1/15/2024") {
			t.Error("Missing or incorrect activity line")
		}
	})
}

func TestFormatActivityDetails(t *testing.T) {
	formatter := NewOutputFormatter()

	t.Run("nil details", func(t *testing.T) {
		result := formatter.FormatActivityDetails(nil)
		expected := "❌ **No activity details available**"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("basic activity details", func(t *testing.T) {
		details := &StravaActivityDetail{
			StravaActivity: StravaActivity{
				ID:                 123456789,
				Name:               "Epic Mountain Ride",
				Type:               "Ride",
				SportType:          "cycling",
				Distance:           45000.0, // 45km
				MovingTime:         7200,    // 2 hours
				TotalElevationGain: 1200.0,
				StartDateLocal:     "2024-01-15T09:30:00Z",
				AverageSpeed:       6.25, // m/s = 22.5 km/h
				MaxSpeed:          15.0,  // m/s = 54 km/h
			},
		}

		result := formatter.FormatActivityDetails(details)

		// Check key components
		if !strings.Contains(result, "🚴 **Epic Mountain Ride** (ID: 123456789)") {
			t.Error("Missing or incorrect header")
		}
		if !strings.Contains(result, "Type: Ride (cycling)") {
			t.Error("Missing or incorrect type")
		}
		if !strings.Contains(result, "Distance: 45.00km") {
			t.Error("Missing or incorrect distance")
		}
		if !strings.Contains(result, "Elevation Gain: 1200 m") {
			t.Error("Missing or incorrect elevation")
		}
	})
}

func TestGetActivityEmoji(t *testing.T) {
	formatter := &outputFormatter{}

	tests := []struct {
		activityType string
		sportType    string
		expected     string
	}{
		{"Run", "running", "🏃"},
		{"Ride", "cycling", "🚴"},
		{"Swim", "swimming", "🏊"},
		{"Walk", "walking", "🚶"},
		{"Hike", "hiking", "🥾"},
		{"Unknown", "", "🏃"}, // Default case
	}

	for _, test := range tests {
		result := formatter.getActivityEmoji(test.activityType, test.sportType)
		if result != test.expected {
			t.Errorf("getActivityEmoji(%q, %q) = %q, expected %q", 
				test.activityType, test.sportType, result, test.expected)
		}
	}
}

func TestFormatDistance(t *testing.T) {
	formatter := &outputFormatter{}

	tests := []struct {
		meters   float64
		expected string
	}{
		{500.0, "500m"},
		{999.0, "999m"},
		{1000.0, "1.00km"},
		{5000.0, "5.00km"},
	}

	for _, test := range tests {
		result := formatter.formatDistance(test.meters)
		if result != test.expected {
			t.Errorf("formatDistance(%.1f) = %q, expected %q", 
				test.meters, result, test.expected)
		}
	}
}

func TestFormatSpeed(t *testing.T) {
	formatter := &outputFormatter{}

	tests := []struct {
		speedMPS float64
		expected string
	}{
		{5.0, "18.0 km/h"},   // 5 m/s = 18 km/h
		{10.0, "36.0 km/h"},  // 10 m/s = 36 km/h
	}

	for _, test := range tests {
		result := formatter.formatSpeed(test.speedMPS)
		if result != test.expected {
			t.Errorf("formatSpeed(%.2f) = %q, expected %q", 
				test.speedMPS, result, test.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	formatter := &outputFormatter{}

	tests := []struct {
		seconds  int
		expected string
	}{
		{30, "00:30"},
		{90, "01:30"},
		{3600, "01:00:00"},
		{7200, "02:00:00"},
	}

	for _, test := range tests {
		result := formatter.formatDuration(test.seconds)
		if result != test.expected {
			t.Errorf("formatDuration(%d) = %q, expected %q", 
				test.seconds, result, test.expected)
		}
	}
}

func TestImplementedMethods(t *testing.T) {
	formatter := NewOutputFormatter()

	// Test that methods handle nil input gracefully
	streamResult := formatter.FormatStreamData(nil, "raw")
	if !strings.Contains(streamResult, "No stream data available") {
		t.Error("FormatStreamData should return no data available message for nil input")
	}

	featuresResult := formatter.FormatDerivedFeatures(nil)
	if !strings.Contains(featuresResult, "No derived features data available") {
		t.Error("FormatDerivedFeatures should return no data available message")
	}

	summaryResult := formatter.FormatStreamSummary(nil)
	if !strings.Contains(summaryResult, "No stream summary data available") {
		t.Error("FormatStreamSummary should return no data available message for nil input")
	}

	pageResult := formatter.FormatStreamPage(nil)
	if !strings.Contains(pageResult, "No stream page data available") {
		t.Error("FormatStreamPage should return no data available message for nil input")
	}

	// Test with valid data
	streams := &StravaStreams{
		Time:      []int{0, 1, 2, 3, 4},
		Heartrate: []int{120, 125, 130, 135, 140},
		Watts:     []int{100, 110, 120, 130, 140},
	}

	streamResult = formatter.FormatStreamData(streams, "raw")
	if !strings.Contains(streamResult, "Stream Data") {
		t.Error("FormatStreamData should format valid stream data")
	}
	if !strings.Contains(streamResult, "Heart Rate") {
		t.Error("FormatStreamData should include heart rate data")
	}
}