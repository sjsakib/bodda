package services

import (
	"strings"
	"testing"
)

func TestZoneFormattingWithPlusSign(t *testing.T) {
	formatter := NewOutputFormatter()

	// Test data with zones that have -1 as max value (no upper limit)
	profile := &StravaAthleteWithZones{
		StravaAthlete: &StravaAthlete{
			ID:        12345,
			Firstname: "Test",
			Lastname:  "User",
		},
		Zones: &StravaAthleteZones{
			HeartRate: &StravaZoneSet{
				Zones: []StravaZone{
					{Min: 100, Max: 120},  // Zone 1: normal range
					{Min: 120, Max: 140},  // Zone 2: normal range
					{Min: 140, Max: 160},  // Zone 3: normal range
					{Min: 160, Max: 180},  // Zone 4: normal range
					{Min: 180, Max: -1},   // Zone 5: no upper limit (should show as 180+)
				},
			},
			Power: &StravaZoneSet{
				Zones: []StravaZone{
					{Min: 100, Max: 150},  // Zone 1: normal range
					{Min: 150, Max: 200},  // Zone 2: normal range
					{Min: 200, Max: 250},  // Zone 3: normal range
					{Min: 250, Max: 300},  // Zone 4: normal range
					{Min: 300, Max: -1},   // Zone 5: no upper limit (should show as 300+)
				},
			},
			Pace: &StravaZoneSet{
				Zones: []StravaZone{
					{Min: 240, Max: 300},  // Zone 1: normal range (4:00-5:00 per km)
					{Min: 300, Max: 360},  // Zone 2: normal range (5:00-6:00 per km)
					{Min: 360, Max: 420},  // Zone 3: normal range (6:00-7:00 per km)
					{Min: 420, Max: 480},  // Zone 4: normal range (7:00-8:00 per km)
					{Min: 480, Max: -1},   // Zone 5: no upper limit (should show as 8:00+ per km)
				},
			},
		},
	}

	result := formatter.FormatAthleteProfile(profile)

	// Test heart rate zones formatting
	if !strings.Contains(result, "Zone 5: 180+ bpm") {
		t.Errorf("Expected heart rate zone 5 to show '180+ bpm', but got: %s", result)
	}
	if strings.Contains(result, "180 - -1") {
		t.Errorf("Heart rate zone should not contain '180 - -1', but got: %s", result)
	}

	// Test power zones formatting
	if !strings.Contains(result, "Zone 5: 300+ watts") {
		t.Errorf("Expected power zone 5 to show '300+ watts', but got: %s", result)
	}
	if strings.Contains(result, "300 - -1") {
		t.Errorf("Power zone should not contain '300 - -1', but got: %s", result)
	}

	// Test pace zones formatting
	if !strings.Contains(result, "Zone 5: 8:00+ per km") {
		t.Errorf("Expected pace zone 5 to show '8:00+ per km', but got: %s", result)
	}
	if strings.Contains(result, "8:00 - -1") {
		t.Errorf("Pace zone should not contain '8:00 - -1', but got: %s", result)
	}

	// Verify normal zones still work correctly
	if !strings.Contains(result, "Zone 1: 100-120 bpm") {
		t.Errorf("Expected normal heart rate zone formatting, but got: %s", result)
	}
	if !strings.Contains(result, "Zone 1: 100-150 watts") {
		t.Errorf("Expected normal power zone formatting, but got: %s", result)
	}
}

func TestActivityZoneFormattingWithPlusSign(t *testing.T) {
	formatter := NewOutputFormatter()

	// Test activity zones with -1 max values
	zones := &StravaActivityZones{
		HeartRate: &StravaZoneDistribution{
			Zones: []StravaZoneData{
				{Min: 100, Max: 120, Time: 300},  // Zone 1: 5 minutes
				{Min: 120, Max: 140, Time: 600},  // Zone 2: 10 minutes
				{Min: 140, Max: 160, Time: 900},  // Zone 3: 15 minutes
				{Min: 160, Max: 180, Time: 600},  // Zone 4: 10 minutes
				{Min: 180, Max: -1, Time: 300},   // Zone 5: 5 minutes, no upper limit
			},
			SensorBased: true,
		},
		Power: &StravaZoneDistribution{
			Zones: []StravaZoneData{
				{Min: 100, Max: 150, Time: 400},  // Zone 1
				{Min: 150, Max: 200, Time: 500},  // Zone 2
				{Min: 200, Max: 250, Time: 600},  // Zone 3
				{Min: 250, Max: 300, Time: 700},  // Zone 4
				{Min: 300, Max: -1, Time: 500},   // Zone 5: no upper limit
			},
			SensorBased: true,
		},
	}

	result := formatter.FormatActivityZones(zones)

	// Test heart rate zones formatting in activity zones
	if !strings.Contains(result, "Zone 5** (180+ bpm)") {
		t.Errorf("Expected activity heart rate zone 5 to show '(180+ bpm)', but got: %s", result)
	}
	if strings.Contains(result, "180--1") || strings.Contains(result, "180 - -1") {
		t.Errorf("Activity heart rate zone should not contain '180--1' or '180 - -1', but got: %s", result)
	}

	// Test power zones formatting in activity zones
	if !strings.Contains(result, "Zone 5** (300+ watts)") {
		t.Errorf("Expected activity power zone 5 to show '(300+ watts)', but got: %s", result)
	}
	if strings.Contains(result, "300--1") || strings.Contains(result, "300 - -1") {
		t.Errorf("Activity power zone should not contain '300--1' or '300 - -1', but got: %s", result)
	}

	// Verify normal zones still work correctly
	if !strings.Contains(result, "Zone 1** (100-120 bpm)") {
		t.Errorf("Expected normal activity heart rate zone formatting, but got: %s", result)
	}
	if !strings.Contains(result, "Zone 1** (100-150 watts)") {
		t.Errorf("Expected normal activity power zone formatting, but got: %s", result)
	}
}

func TestZoneFormattingSkipsInvalidZones(t *testing.T) {
	formatter := NewOutputFormatter()

	// Test data with invalid zones (Min = -1)
	profile := &StravaAthleteWithZones{
		StravaAthlete: &StravaAthlete{
			ID:        12345,
			Firstname: "Test",
			Lastname:  "User",
		},
		Zones: &StravaAthleteZones{
			HeartRate: &StravaZoneSet{
				Zones: []StravaZone{
					{Min: 100, Max: 120},  // Zone 1: valid
					{Min: -1, Max: -1},    // Invalid zone (should be skipped)
					{Min: 140, Max: 160},  // Zone 2: valid (should be numbered as zone 2)
					{Min: 160, Max: -1},   // Zone 3: valid with no upper limit
				},
			},
		},
	}

	result := formatter.FormatAthleteProfile(profile)

	// Should not contain any reference to -1 values in zone formatting
	if strings.Contains(result, " -1") || strings.Contains(result, "--1") {
		t.Errorf("Result should not contain any -1 values in zone formatting, but got: %s", result)
	}

	// Should contain valid zones with correct numbering
	if !strings.Contains(result, "Zone 1: 100-120 bpm") {
		t.Errorf("Expected zone 1 to be formatted correctly, but got: %s", result)
	}
	if !strings.Contains(result, "Zone 2: 140-160 bpm") {
		t.Errorf("Expected zone 2 to be formatted correctly (skipping invalid zone), but got: %s", result)
	}
	if !strings.Contains(result, "Zone 3: 160+ bpm") {
		t.Errorf("Expected zone 3 to show '160+ bpm', but got: %s", result)
	}
}