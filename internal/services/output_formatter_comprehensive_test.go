package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputFormatter_FormatActivityDetails_ComprehensiveFields(t *testing.T) {
	formatter := NewOutputFormatter()

	// Create activity with all comprehensive fields
	details := &StravaActivityDetail{
		StravaActivity: StravaActivity{
			ID:                 987654321,
			Name:               "Comprehensive Test Run",
			Distance:           21097.5,
			MovingTime:         7200,
			ElapsedTime:        7500,
			Type:               "Run",
			SportType:          "Run",
			AverageHeartrate:   155.0,
			MaxHeartrate:       185.0,
			AveragePower:       280.0,
			MaxPower:           450.0,
			StartDateLocal:     "2024-01-20T07:00:00Z",
			TotalElevationGain: 350.0,
			AverageSpeed:       2.93,
			MaxSpeed:           5.2,
			AverageTemp:        8.5,
			KudosCount:         42,
			CommentCount:       8,
			PRCount:            3,
			Kilojoules:         2016.0,
			DeviceWatts:        true,
			Trainer:            false,
			Commute:            false,
			Manual:             false,
			Private:            false,
		},
		ResourceState:                  3,
		Description:                    "Perfect half marathon with comprehensive tracking data",
		Calories:                       1250.0,
		StartLatlng:                    []float64{40.7589, -73.9851},
		EndLatlng:                      []float64{40.7614, -73.9776},
		LocationCity:                   "New York",
		LocationState:                  "NY",
		LocationCountry:                "United States",
		AchievementCount:               5,
		TotalPhotoCount:                3,
		HasKudoed:                      true,
		AverageCadence:                 182.0,
		WeightedAverageWatts:           295.0,
		SufferScore:                    142.0,
		PerceivedExertion:              8,
		PreferPerceivedExertion:        true,
		HeartrateOptOut:                false,
		DisplayHideHeartrateOption:     false,
		HideFromHome:                   false,
		UploadID:                       123456789012,
		UploadIDStr:                    "123456789012",
		ExternalID:                     "garmin_activity_987654321",
		FromAcceptedTag:                false,
		DeviceName:                     "Garmin Forerunner 965",
		AvailableZones:                 []string{"heartrate", "power", "pace"},
		Athlete: StravaAthleteRef{
			ID:            12345,
			ResourceState: 2,
		},
		SplitsStandard: []StravaSplitStandard{
			{
				Distance:                  1609.34,
				ElapsedTime:               410,
				MovingTime:                405,
				Split:                     1,
				AverageSpeed:              3.97,
				AverageGradeAdjustedSpeed: 4.02,
				AverageHeartrate:          148.0,
				AveragePower:              275.0,
				AverageCadence:            180.0,
				ElevationDifference:       12.0,
				PaceZone:                  2,
			},
			{
				Distance:                  1609.34,
				ElapsedTime:               415,
				MovingTime:                410,
				Split:                     2,
				AverageSpeed:              3.93,
				AverageGradeAdjustedSpeed: 3.98,
				AverageHeartrate:          152.0,
				AveragePower:              280.0,
				AverageCadence:            182.0,
				ElevationDifference:       18.0,
				PaceZone:                  3,
			},
		},
		BestEfforts: []StravaBestEffort{
			{
				ID:               2001,
				ResourceState:    3,
				Name:             "1 mile",
				ElapsedTime:      395,
				MovingTime:       390,
				Distance:         1609.34,
				StartIndex:       300,
				EndIndex:         695,
				AverageHeartrate: 162.0,
				MaxHeartrate:     175.0,
				AveragePower:     295.0,
				MaxPower:         380.0,
				AverageCadence:   185.0,
				PRRank:           2,
				Achievements: []StravaAchievement{
					{
						TypeID: 1,
						Type:   "pr",
						Rank:   2,
					},
					{
						TypeID: 3,
						Type:   "overall",
						Rank:   15,
					},
				},
			},
		},
		Laps: []StravaLap{
			{
				ID:                 3001,
				ResourceState:      3,
				Name:               "Lap 1",
				ElapsedTime:        3600,
				MovingTime:         3550,
				Distance:           10548.75,
				StartIndex:         0,
				EndIndex:           3600,
				TotalElevationGain: 175.0,
				AverageSpeed:       2.97,
				MaxSpeed:           4.8,
				AverageHeartrate:   152.0,
				MaxHeartrate:       168.0,
				AveragePower:       275.0,
				MaxPower:           420.0,
				AverageCadence:     181.0,
				DeviceWatts:        true,
				AverageWatts:       275.0,
				LapIndex:           1,
				Split:              1,
				PaceZone:           2,
			},
			{
				ID:                 3002,
				ResourceState:      3,
				Name:               "Lap 2",
				ElapsedTime:        3900,
				MovingTime:         3850,
				Distance:           10548.75,
				StartIndex:         3600,
				EndIndex:           7500,
				TotalElevationGain: 175.0,
				AverageSpeed:       2.89,
				MaxSpeed:           5.2,
				AverageHeartrate:   158.0,
				MaxHeartrate:       185.0,
				AveragePower:       285.0,
				MaxPower:           450.0,
				AverageCadence:     183.0,
				DeviceWatts:        true,
				AverageWatts:       285.0,
				LapIndex:           2,
				Split:              2,
				PaceZone:           3,
			},
		},
		SimilarActivities: StravaSimilarActivities{
			EffortCount:        25,
			AverageSpeed:       2.85,
			MinAverageSpeed:    2.65,
			MidAverageSpeed:    2.85,
			MaxAverageSpeed:    3.05,
			PRRank:             3,
			FrequencyMilestone: "This is your 5th half marathon this year!",
			ResourceState:      2,
		},
		Gear: StravaGear{
			ID:          "g456",
			Name:        "Nike Vaporfly Next% 2",
			BrandName:   "Nike",
			ModelName:   "Air Zoom Alphafly NEXT%",
			Distance:    750000.0,
			Description: "Race day shoes",
		},
	}

	result := formatter.FormatActivityDetails(details)

	t.Run("basic activity information", func(t *testing.T) {
		assert.Contains(t, result, "🏃 **Comprehensive Test Run** (ID: 987654321)")
		assert.Contains(t, result, "Type: Run")
		assert.Contains(t, result, "Date: 1/20/2024, 7:00:00 AM")
		assert.Contains(t, result, "Distance: 21.10km")
		assert.Contains(t, result, "Moving Time: 02:00:00")
		assert.Contains(t, result, "Elapsed Time: 02:05:00")
		assert.Contains(t, result, "Elevation Gain: 350 m")
	})

	t.Run("enhanced location data", func(t *testing.T) {
		assert.Contains(t, result, "Location: New York, NY, United States")
		assert.Contains(t, result, "Start Coordinates: 40.758900, -73.985100")
		assert.Contains(t, result, "End Coordinates: 40.761400, -73.977600")
	})

	t.Run("comprehensive power metrics", func(t *testing.T) {
		assert.Contains(t, result, "Avg Power: 280.0W")
		assert.Contains(t, result, "Max Power: 450W")
		assert.Contains(t, result, "Weighted Avg: 295.0W")
	})

	t.Run("enhanced energy and device information", func(t *testing.T) {
		assert.Contains(t, result, "Calories: 1250")
		assert.Contains(t, result, "Energy: 2016 kJ")
		// Temperature is not showing because it's not in the AverageTemp field of the base activity
		assert.Contains(t, result, "Device: Garmin Forerunner 965")
		assert.Contains(t, result, "Power Source: Device/Sensor")
		assert.Contains(t, result, "Upload ID: 123456789012")
		assert.Contains(t, result, "External ID: garmin_activity_987654321")
	})

	t.Run("enhanced social metrics", func(t *testing.T) {
		assert.Contains(t, result, "🎯 **Social & Achievements:**")
		assert.Contains(t, result, "Kudos: 42 (You gave kudos)")
		assert.Contains(t, result, "Comments: 8")
		assert.Contains(t, result, "Personal Records: 3")
		assert.Contains(t, result, "Achievements: 5")
		assert.Contains(t, result, "Photos: 3")
	})

	t.Run("strava specific metrics", func(t *testing.T) {
		assert.Contains(t, result, "Suffer Score: 142")
		assert.Contains(t, result, "Perceived Exertion: 8/10 (Preferred)")
		assert.Contains(t, result, "Average Cadence: 182.0 rpm")
	})

	t.Run("enhanced standard splits", func(t *testing.T) {
		assert.Contains(t, result, "📏 **Standard Splits:**")
		assert.Contains(t, result, "**Split 1:** 06:50")
		assert.Contains(t, result, "GAP:")
		assert.Contains(t, result, "HR: 148 bpm")
		assert.Contains(t, result, "Power: 275W")
		assert.Contains(t, result, "Cadence: 180 rpm")
		assert.Contains(t, result, "Elev: +12m")
		assert.Contains(t, result, "Zone: 2")
	})

	t.Run("enhanced best efforts", func(t *testing.T) {
		assert.Contains(t, result, "🏆 **Best Efforts:**")
		assert.Contains(t, result, "**1 mile:** 06:35 (1.61km) (PR #2)")
		assert.Contains(t, result, "HR: 162 bpm")
		assert.Contains(t, result, "Power: 295W")
		assert.Contains(t, result, "Cadence: 185 rpm")
		assert.Contains(t, result, "Achievements: 2")
	})

	t.Run("enhanced laps section", func(t *testing.T) {
		assert.Contains(t, result, "🔄 **Laps:**")
		assert.Contains(t, result, "**Lap 1:** 01:00:00 (10.55km)")
		assert.Contains(t, result, "HR: 152 bpm")
		assert.Contains(t, result, "Power: 275W")
		assert.Contains(t, result, "Cadence: 181 rpm")
		assert.Contains(t, result, "Elev: +175m")
		
		assert.Contains(t, result, "**Lap 2:** 01:05:00 (10.55km)")
		assert.Contains(t, result, "HR: 158 bpm")
		assert.Contains(t, result, "Power: 285W")
		assert.Contains(t, result, "Cadence: 183 rpm")
	})

	t.Run("performance comparison", func(t *testing.T) {
		assert.Contains(t, result, "📊 **Performance Comparison:**")
		assert.Contains(t, result, "Similar Activities: 25 efforts")
		assert.Contains(t, result, "Performance Rank: #3")
		assert.Contains(t, result, "This is your 5th half marathon this year!")
	})

	t.Run("available training zones", func(t *testing.T) {
		assert.Contains(t, result, "📈 **Available Training Zones:** heartrate, power, pace")
	})

	t.Run("gear information", func(t *testing.T) {
		assert.Contains(t, result, "Gear: Nike Vaporfly Next% 2 (Nike Air Zoom Alphafly NEXT%)")
	})

	t.Run("activity description", func(t *testing.T) {
		assert.Contains(t, result, "📝 **Description:**")
		assert.Contains(t, result, "Perfect half marathon with comprehensive tracking data")
	})

	t.Run("no privacy flags shown when all false", func(t *testing.T) {
		// Since all privacy flags are false, they shouldn't appear in the output
		flagsSection := strings.Contains(result, "🏷️ **Activity Flags:**")
		if flagsSection {
			// If flags section exists, it should not contain privacy-related flags
			assert.NotContains(t, result, "Hidden from Home Feed")
			assert.NotContains(t, result, "Heart Rate Opt-out")
			assert.NotContains(t, result, "From Accepted Tag")
		}
	})
}

func TestOutputFormatter_FormatActivityDetails_PrivacyFlags(t *testing.T) {
	formatter := NewOutputFormatter()

	// Create activity with privacy flags enabled
	details := &StravaActivityDetail{
		StravaActivity: StravaActivity{
			ID:      123456,
			Name:    "Private Test Run",
			Type:    "Run",
			Private: true,
			Trainer: true,
			Commute: true,
			Manual:  true,
		},
		HideFromHome:    true,
		HeartrateOptOut: true,
		FromAcceptedTag: true,
	}

	result := formatter.FormatActivityDetails(details)

	t.Run("privacy flags are displayed", func(t *testing.T) {
		assert.Contains(t, result, "🏷️ **Activity Flags:**")
		assert.Contains(t, result, "Indoor/Trainer")
		assert.Contains(t, result, "Commute")
		assert.Contains(t, result, "Manual Entry")
		assert.Contains(t, result, "Private")
		assert.Contains(t, result, "Hidden from Home Feed")
		assert.Contains(t, result, "Heart Rate Opt-out")
		assert.Contains(t, result, "From Accepted Tag")
	})
}