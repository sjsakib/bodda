package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStravaService_GetActivityDetail_Enhanced(t *testing.T) {
	// Mock enhanced activity detail with all new fields
	mockActivity := StravaActivityDetail{
		StravaActivity: StravaActivity{
			ID:                 123456,
			Name:               "Enhanced Morning Run",
			Distance:           10000.0,
			MovingTime:         3600,
			ElapsedTime:        3720,
			Type:               "Run",
			SportType:          "Run",
			AverageHeartrate:   150.0,
			MaxHeartrate:       180.0,
			AveragePower:       250.0,
			MaxPower:           400.0,
			StartDateLocal:     "2024-01-15T08:00:00Z",
			TotalElevationGain: 200.0,
			AverageSpeed:       2.78,
			MaxSpeed:           4.5,
			AverageTemp:        15.0,
			KudosCount:         25,
			CommentCount:       5,
			PRCount:            2,
			Kilojoules:         1200.0,
			DeviceWatts:        true,
		},
		ResourceState:                  3,
		Description:                    "Great morning run with comprehensive data",
		Calories:                       500.0,
		StartLatlng:                    []float64{37.7749, -122.4194},
		EndLatlng:                      []float64{37.7849, -122.4094},
		LocationCity:                   "San Francisco",
		LocationState:                  "CA",
		LocationCountry:                "United States",
		AchievementCount:               3,
		TotalPhotoCount:                2,
		HasKudoed:                      true,
		AverageCadence:                 180.0,
		WeightedAverageWatts:           275.0,
		SufferScore:                    85.0,
		PerceivedExertion:              7,
		PreferPerceivedExertion:        false,
		HeartrateOptOut:                false,
		DisplayHideHeartrateOption:     false,
		HideFromHome:                   false,
		UploadID:                       987654321,
		UploadIDStr:                    "987654321",
		ExternalID:                     "garmin_123456789",
		FromAcceptedTag:                false,
		DeviceName:                     "Garmin Forerunner 945",
		AvailableZones:                 []string{"heartrate", "power"},
		Athlete: StravaAthleteRef{
			ID:            789,
			ResourceState: 2,
		},
		SplitsStandard: []StravaSplitStandard{
			{
				Distance:                  1000.0,
				ElapsedTime:               360,
				MovingTime:                350,
				Split:                     1,
				AverageSpeed:              2.86,
				AverageGradeAdjustedSpeed: 2.90,
				AverageHeartrate:          145.0,
				AveragePower:              240.0,
				AverageCadence:            175.0,
				ElevationDifference:       10.0,
				PaceZone:                  2,
			},
			{
				Distance:                  1000.0,
				ElapsedTime:               355,
				MovingTime:                345,
				Split:                     2,
				AverageSpeed:              2.90,
				AverageGradeAdjustedSpeed: 2.95,
				AverageHeartrate:          150.0,
				AveragePower:              250.0,
				AverageCadence:            180.0,
				ElevationDifference:       15.0,
				PaceZone:                  3,
			},
		},
		BestEfforts: []StravaBestEffort{
			{
				ID:               1001,
				ResourceState:    3,
				Name:             "1 mile",
				Activity: StravaActivityRef{
					ID:            123456,
					ResourceState: 2,
				},
				Athlete: StravaAthleteRef{
					ID:            789,
					ResourceState: 2,
				},
				ElapsedTime:      420,
				MovingTime:       415,
				StartDate:        "2024-01-15T08:05:00Z",
				StartDateLocal:   "2024-01-15T08:05:00Z",
				Distance:         1609.34,
				StartIndex:       100,
				EndIndex:         520,
				AverageHeartrate: 165.0,
				MaxHeartrate:     175.0,
				AveragePower:     280.0,
				MaxPower:         350.0,
				AverageCadence:   185.0,
				PRRank:           3,
				Achievements: []StravaAchievement{
					{
						TypeID: 1,
						Type:   "pr",
						Rank:   3,
					},
				},
			},
			{
				ID:               1002,
				ResourceState:    3,
				Name:             "5k",
				Activity: StravaActivityRef{
					ID:            123456,
					ResourceState: 2,
				},
				Athlete: StravaAthleteRef{
					ID:            789,
					ResourceState: 2,
				},
				ElapsedTime:      1200,
				MovingTime:       1180,
				StartDate:        "2024-01-15T08:00:00Z",
				StartDateLocal:   "2024-01-15T08:00:00Z",
				Distance:         5000.0,
				StartIndex:       0,
				EndIndex:         1200,
				AverageHeartrate: 155.0,
				MaxHeartrate:     170.0,
				AveragePower:     260.0,
				MaxPower:         320.0,
				AverageCadence:   178.0,
				PRRank:           1,
				Achievements: []StravaAchievement{
					{
						TypeID: 2,
						Type:   "pr",
						Rank:   1,
					},
				},
			},
		},
		SimilarActivities: StravaSimilarActivities{
			EffortCount:         15,
			AverageSpeed:        2.75,
			MinAverageSpeed:     2.50,
			MidAverageSpeed:     2.75,
			MaxAverageSpeed:     3.00,
			PRRank:              5,
			FrequencyMilestone:  "This is your 10th run this month!",
			ResourceState:       2,
			TrendStats: StravaTrendStats{
				Speeds:                   []float64{2.60, 2.65, 2.70, 2.75, 2.78},
				CurrentActivityIndex:     4,
				MinSpeed:                 2.60,
				MidSpeed:                 2.70,
				MaxSpeed:                 2.78,
				Direction:                1,
			},
		},
		Laps: []StravaLap{
			{
				ID:                 2001,
				ResourceState:      3,
				Name:               "Lap 1",
				Activity: StravaActivityRef{
					ID:            123456,
					ResourceState: 2,
				},
				Athlete: StravaAthleteRef{
					ID:            789,
					ResourceState: 2,
				},
				ElapsedTime:        1800,
				MovingTime:         1750,
				StartDate:          "2024-01-15T08:00:00Z",
				StartDateLocal:     "2024-01-15T08:00:00Z",
				Distance:           5000.0,
				StartIndex:         0,
				EndIndex:           1800,
				TotalElevationGain: 100.0,
				AverageSpeed:       2.86,
				MaxSpeed:           4.2,
				AverageHeartrate:   148.0,
				MaxHeartrate:       165.0,
				AveragePower:       245.0,
				MaxPower:           320.0,
				AverageCadence:     178.0,
				DeviceWatts:        true,
				AverageWatts:       245.0,
				LapIndex:           1,
				Split:              1,
				PaceZone:           2,
			},
			{
				ID:                 2002,
				ResourceState:      3,
				Name:               "Lap 2",
				Activity: StravaActivityRef{
					ID:            123456,
					ResourceState: 2,
				},
				Athlete: StravaAthleteRef{
					ID:            789,
					ResourceState: 2,
				},
				ElapsedTime:        1920,
				MovingTime:         1850,
				StartDate:          "2024-01-15T08:30:00Z",
				StartDateLocal:     "2024-01-15T08:30:00Z",
				Distance:           5000.0,
				StartIndex:         1800,
				EndIndex:           3720,
				TotalElevationGain: 100.0,
				AverageSpeed:       2.70,
				MaxSpeed:           3.8,
				AverageHeartrate:   152.0,
				MaxHeartrate:       175.0,
				AveragePower:       255.0,
				MaxPower:           380.0,
				AverageCadence:     182.0,
				DeviceWatts:        true,
				AverageWatts:       255.0,
				LapIndex:           2,
				Split:              2,
				PaceZone:           3,
			},
		},
		Gear: StravaGear{
			ID:          "g123",
			Name:        "Nike Pegasus 40",
			BrandName:   "Nike",
			ModelName:   "Air Zoom Pegasus 40",
			Distance:    500000.0,
			Description: "Primary running shoes",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/activities/123456", r.URL.Path)
		assert.Equal(t, "Bearer test_token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockActivity)
	}))
	defer server.Close()

	cfg := &config.Config{}
	mockUserRepo := &MockStravaUserRepository{}
	service := NewTestStravaService(cfg, server.URL, mockUserRepo)

	testUser := &models.User{
		ID:           "test-user-id",
		AccessToken:  "test_token",
		RefreshToken: "test_refresh_token",
		TokenExpiry:  time.Now().Add(time.Hour),
	}

	activity, err := service.GetActivityDetail(testUser, 123456)

	require.NoError(t, err)
	
	// Test basic fields
	assert.Equal(t, mockActivity.ID, activity.ID)
	assert.Equal(t, mockActivity.Name, activity.Name)
	assert.Equal(t, mockActivity.Description, activity.Description)
	
	// Test enhanced fields
	assert.Equal(t, 3, activity.ResourceState)
	assert.Equal(t, []float64{37.7749, -122.4194}, activity.StartLatlng)
	assert.Equal(t, []float64{37.7849, -122.4094}, activity.EndLatlng)
	assert.Equal(t, "San Francisco", activity.LocationCity)
	assert.Equal(t, "CA", activity.LocationState)
	assert.Equal(t, "United States", activity.LocationCountry)
	assert.Equal(t, 3, activity.AchievementCount)
	assert.Equal(t, 2, activity.TotalPhotoCount)
	assert.True(t, activity.HasKudoed)
	assert.Equal(t, 180.0, activity.AverageCadence)
	assert.Equal(t, 275.0, activity.WeightedAverageWatts)
	assert.Equal(t, 85.0, activity.SufferScore)
	assert.Equal(t, 7, activity.PerceivedExertion)
	assert.False(t, activity.PreferPerceivedExertion)
	assert.False(t, activity.HeartrateOptOut)
	assert.False(t, activity.DisplayHideHeartrateOption)
	assert.False(t, activity.HideFromHome)
	assert.Equal(t, int64(987654321), activity.UploadID)
	assert.Equal(t, "987654321", activity.UploadIDStr)
	assert.Equal(t, "garmin_123456789", activity.ExternalID)
	assert.False(t, activity.FromAcceptedTag)
	assert.Equal(t, "Garmin Forerunner 945", activity.DeviceName)
	assert.Equal(t, []string{"heartrate", "power"}, activity.AvailableZones)
	
	// Test athlete reference
	assert.Equal(t, int64(789), activity.Athlete.ID)
	assert.Equal(t, 2, activity.Athlete.ResourceState)
	
	// Test standard splits
	require.Len(t, activity.SplitsStandard, 2)
	assert.Equal(t, 1, activity.SplitsStandard[0].Split)
	assert.Equal(t, 145.0, activity.SplitsStandard[0].AverageHeartrate)
	assert.Equal(t, 240.0, activity.SplitsStandard[0].AveragePower)
	assert.Equal(t, 175.0, activity.SplitsStandard[0].AverageCadence)
	assert.Equal(t, 10.0, activity.SplitsStandard[0].ElevationDifference)
	assert.Equal(t, 2.90, activity.SplitsStandard[0].AverageGradeAdjustedSpeed)
	assert.Equal(t, 2, activity.SplitsStandard[0].PaceZone)
	
	// Test best efforts
	require.Len(t, activity.BestEfforts, 2)
	assert.Equal(t, int64(1001), activity.BestEfforts[0].ID)
	assert.Equal(t, 3, activity.BestEfforts[0].ResourceState)
	assert.Equal(t, "1 mile", activity.BestEfforts[0].Name)
	assert.Equal(t, int64(123456), activity.BestEfforts[0].Activity.ID)
	assert.Equal(t, 2, activity.BestEfforts[0].Activity.ResourceState)
	assert.Equal(t, int64(789), activity.BestEfforts[0].Athlete.ID)
	assert.Equal(t, 2, activity.BestEfforts[0].Athlete.ResourceState)
	assert.Equal(t, 420, activity.BestEfforts[0].ElapsedTime)
	assert.Equal(t, "2024-01-15T08:05:00Z", activity.BestEfforts[0].StartDate)
	assert.Equal(t, "2024-01-15T08:05:00Z", activity.BestEfforts[0].StartDateLocal)
	assert.Equal(t, 1609.34, activity.BestEfforts[0].Distance)
	assert.Equal(t, 165.0, activity.BestEfforts[0].AverageHeartrate)
	assert.Equal(t, 280.0, activity.BestEfforts[0].AveragePower)
	assert.Equal(t, 185.0, activity.BestEfforts[0].AverageCadence)
	assert.Equal(t, 3, activity.BestEfforts[0].PRRank)
	require.Len(t, activity.BestEfforts[0].Achievements, 1)
	assert.Equal(t, "pr", activity.BestEfforts[0].Achievements[0].Type)
	
	// Test laps
	require.Len(t, activity.Laps, 2)
	assert.Equal(t, int64(2001), activity.Laps[0].ID)
	assert.Equal(t, 3, activity.Laps[0].ResourceState)
	assert.Equal(t, "Lap 1", activity.Laps[0].Name)
	assert.Equal(t, int64(123456), activity.Laps[0].Activity.ID)
	assert.Equal(t, int64(789), activity.Laps[0].Athlete.ID)
	assert.Equal(t, 1800, activity.Laps[0].ElapsedTime)
	assert.Equal(t, "2024-01-15T08:00:00Z", activity.Laps[0].StartDate)
	assert.Equal(t, "2024-01-15T08:00:00Z", activity.Laps[0].StartDateLocal)
	assert.Equal(t, 5000.0, activity.Laps[0].Distance)
	assert.Equal(t, 100.0, activity.Laps[0].TotalElevationGain)
	assert.Equal(t, 148.0, activity.Laps[0].AverageHeartrate)
	assert.Equal(t, 245.0, activity.Laps[0].AveragePower)
	assert.Equal(t, 178.0, activity.Laps[0].AverageCadence)
	assert.True(t, activity.Laps[0].DeviceWatts)
	assert.Equal(t, 245.0, activity.Laps[0].AverageWatts)
	assert.Equal(t, 1, activity.Laps[0].LapIndex)
	assert.Equal(t, 2, activity.Laps[0].PaceZone)
	
	// Test similar activities
	assert.Equal(t, 15, activity.SimilarActivities.EffortCount)
	assert.Equal(t, 2.75, activity.SimilarActivities.AverageSpeed)
	assert.Equal(t, 5, activity.SimilarActivities.PRRank)
	assert.Equal(t, "This is your 10th run this month!", activity.SimilarActivities.FrequencyMilestone)
	assert.Equal(t, 1, activity.SimilarActivities.TrendStats.Direction)
	assert.Equal(t, 4, activity.SimilarActivities.TrendStats.CurrentActivityIndex)
}

func TestStravaService_GetActivityZones(t *testing.T) {
	// Mock zones response
	mockZones := []StravaZoneDistribution{
		{
			Type:          "heartrate",
			ResourceState: 3,
			SensorBased:   true,
			CustomZones:   false,
			Zones: []StravaZoneData{
				{Min: 0, Max: 142, Time: 300},
				{Min: 142, Max: 155, Time: 900},
				{Min: 155, Max: 168, Time: 1800},
				{Min: 168, Max: 181, Time: 600},
				{Min: 181, Max: 220, Time: 0},
			},
		},
		{
			Type:          "power",
			ResourceState: 3,
			SensorBased:   true,
			CustomZones:   false,
			Zones: []StravaZoneData{
				{Min: 0, Max: 200, Time: 600},
				{Min: 200, Max: 250, Time: 1200},
				{Min: 250, Max: 300, Time: 1500},
				{Min: 300, Max: 350, Time: 300},
				{Min: 350, Max: 500, Time: 0},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/activities/123456/zones", r.URL.Path)
		assert.Equal(t, "Bearer test_token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockZones)
	}))
	defer server.Close()

	cfg := &config.Config{}
	mockUserRepo := &MockStravaUserRepository{}
	service := NewTestStravaService(cfg, server.URL, mockUserRepo)

	testUser := &models.User{
		ID:           "test-user-id",
		AccessToken:  "test_token",
		RefreshToken: "test_refresh_token",
		TokenExpiry:  time.Now().Add(time.Hour),
	}

	zones, err := service.GetActivityZones(testUser, 123456)

	require.NoError(t, err)
	require.NotNil(t, zones)
	
	// Test heart rate zones
	require.NotNil(t, zones.HeartRate)
	assert.Equal(t, "heartrate", zones.HeartRate.Type)
	assert.True(t, zones.HeartRate.SensorBased)
	assert.False(t, zones.HeartRate.CustomZones)
	require.Len(t, zones.HeartRate.Zones, 5)
	
	// Test specific heart rate zone
	assert.Equal(t, 142, zones.HeartRate.Zones[1].Min)
	assert.Equal(t, 155, zones.HeartRate.Zones[1].Max)
	assert.Equal(t, 900, zones.HeartRate.Zones[1].Time)
	
	// Test power zones
	require.NotNil(t, zones.Power)
	assert.Equal(t, "power", zones.Power.Type)
	assert.True(t, zones.Power.SensorBased)
	require.Len(t, zones.Power.Zones, 5)
	
	// Test specific power zone
	assert.Equal(t, 250, zones.Power.Zones[2].Min)
	assert.Equal(t, 300, zones.Power.Zones[2].Max)
	assert.Equal(t, 1500, zones.Power.Zones[2].Time)
	
	// Pace zones should be nil in this test
	assert.Nil(t, zones.Pace)
}

func TestStravaService_GetActivityZones_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Record Not Found", "errors": [{"resource": "Activity", "field": "zones", "code": "not found"}]}`))
	}))
	defer server.Close()

	cfg := &config.Config{}
	mockUserRepo := &MockStravaUserRepository{}
	service := NewTestStravaService(cfg, server.URL, mockUserRepo)

	testUser := &models.User{
		ID:           "test-user-id",
		AccessToken:  "test_token",
		RefreshToken: "test_refresh_token",
		TokenExpiry:  time.Now().Add(time.Hour),
	}

	zones, err := service.GetActivityZones(testUser, 123456)

	assert.Error(t, err)
	assert.Nil(t, zones)
	assert.Contains(t, err.Error(), "failed to get activity zones")
}

func TestOutputFormatter_FormatActivityDetails_Enhanced(t *testing.T) {
	formatter := NewOutputFormatter()

	t.Run("enhanced activity details with all fields", func(t *testing.T) {
		details := &StravaActivityDetail{
			StravaActivity: StravaActivity{
				ID:                 123456789,
				Name:               "Enhanced Test Run",
				Distance:           5000.0,
				MovingTime:         1800,
				Type:               "Run",
				AverageHeartrate:   150.0,
				MaxHeartrate:       180.0,
				AveragePower:       250.0,
				MaxPower:           400.0,
				StartDateLocal:     "2024-01-15T08:00:00Z",
				KudosCount:         15,
				CommentCount:       3,
				PRCount:            2,
			},
			ResourceState:        3,
			Description:          "Enhanced test run with comprehensive data",
			StartLatlng:          []float64{37.7749, -122.4194},
			EndLatlng:            []float64{37.7849, -122.4094},
			LocationCity:         "San Francisco",
			LocationState:        "CA",
			LocationCountry:      "United States",
			AchievementCount:     5,
			AverageCadence:       180.0,
			WeightedAverageWatts: 275.0,
			SufferScore:          85.0,
			PerceivedExertion:    7,
			AvailableZones:       []string{"heartrate", "power"},
			BestEfforts: []StravaBestEffort{
				{
					Name:             "1 mile",
					ElapsedTime:      420,
					AverageHeartrate: 165.0,
					AveragePower:     280.0,
					PRRank:           3,
				},
			},
			SplitsStandard: []StravaSplitStandard{
				{
					Split:               1,
					ElapsedTime:         360,
					AverageSpeed:        2.86,
					AverageHeartrate:    145.0,
					ElevationDifference: 10.0,
				},
			},
			SimilarActivities: StravaSimilarActivities{
				EffortCount:        12,
				PRRank:             4,
				AverageSpeed:       2.75,
				FrequencyMilestone: "This is your 8th run this month!",
			},
		}

		result := formatter.FormatActivityDetails(details)

		// Test basic formatting
		assert.Contains(t, result, "🏃 **Enhanced Test Run** (ID: 123456789)")
		assert.Contains(t, result, "Type: Run")
		assert.Contains(t, result, "Date: 1/15/2024, 8:00:00 AM")
		
		// Test enhanced location data
		assert.Contains(t, result, "Location: San Francisco, CA, United States")
		assert.Contains(t, result, "Start Coordinates: 37.774900, -122.419400")
		assert.Contains(t, result, "End Coordinates: 37.784900, -122.409400")
		
		// Test enhanced power metrics
		assert.Contains(t, result, "Avg Power: 250.0W")
		assert.Contains(t, result, "Max Power: 400W")
		assert.Contains(t, result, "Weighted Avg: 275.0W")
		
		// Test cadence
		assert.Contains(t, result, "Average Cadence: 180.0 rpm")
		
		// Test Strava-specific metrics
		assert.Contains(t, result, "Suffer Score: 85")
		assert.Contains(t, result, "Perceived Exertion: 7/10")
		
		// Test enhanced achievements
		assert.Contains(t, result, "Achievements: 5")
		
		// Test best efforts
		assert.Contains(t, result, "🏆 **Best Efforts:**")
		assert.Contains(t, result, "**1 mile:** 07:00 (PR #3)")
		assert.Contains(t, result, "HR: 165 bpm")
		assert.Contains(t, result, "Power: 280W")
		
		// Test standard splits
		assert.Contains(t, result, "📏 **Standard Splits:**")
		assert.Contains(t, result, "**Split 1:** 06:00")
		assert.Contains(t, result, "HR: 145 bpm")
		assert.Contains(t, result, "Elev: +10m")
		
		// Test performance comparison
		assert.Contains(t, result, "📊 **Performance Comparison:**")
		assert.Contains(t, result, "Similar Activities: 12 efforts")
		assert.Contains(t, result, "Performance Rank: #4")
		assert.Contains(t, result, "This is your 8th run this month!")
		
		// Test available zones
		assert.Contains(t, result, "📈 **Available Training Zones:** heartrate, power")
		
		// Test description
		assert.Contains(t, result, "📝 **Description:**")
		assert.Contains(t, result, "Enhanced test run with comprehensive data")
	})
}

func TestOutputFormatter_FormatActivityZones(t *testing.T) {
	formatter := NewOutputFormatter()

	t.Run("complete zones data", func(t *testing.T) {
		zones := &StravaActivityZones{
			HeartRate: &StravaZoneDistribution{
				Type:        "heartrate",
				SensorBased: true,
				Zones: []StravaZoneData{
					{Min: 0, Max: 142, Time: 300},
					{Min: 142, Max: 155, Time: 900},
					{Min: 155, Max: 168, Time: 1800},
					{Min: 168, Max: 181, Time: 600},
					{Min: 181, Max: 220, Time: 0},
				},
			},
			Power: &StravaZoneDistribution{
				Type:        "power",
				SensorBased: true,
				Zones: []StravaZoneData{
					{Min: 0, Max: 200, Time: 600},
					{Min: 200, Max: 250, Time: 1200},
					{Min: 250, Max: 300, Time: 1500},
					{Min: 300, Max: 350, Time: 300},
					{Min: 350, Max: 500, Time: 0},
				},
			},
		}

		result := formatter.FormatActivityZones(zones)

		// Test header
		assert.Contains(t, result, "📈 **Training Zone Analysis**")
		
		// Test heart rate zones
		assert.Contains(t, result, "💓 **Heart Rate Zones:**")
		assert.Contains(t, result, "**Zone 1** (0-142 bpm): 05:00 (8.3%)")
		assert.Contains(t, result, "**Zone 2** (142-155 bpm): 15:00 (25.0%)")
		assert.Contains(t, result, "**Zone 3** (155-168 bpm): 30:00 (50.0%)")
		assert.Contains(t, result, "**Zone 4** (168-181 bpm): 10:00 (16.7%)")
		assert.Contains(t, result, "**Zone 5** (181-220 bpm): 00:00 (0.0%)")
		assert.Contains(t, result, "Data Source: Heart Rate Sensor")
		
		// Test power zones
		assert.Contains(t, result, "⚡ **Power Zones:**")
		assert.Contains(t, result, "**Zone 1** (0-200 watts): 10:00 (16.7%)")
		assert.Contains(t, result, "**Zone 2** (200-250 watts): 20:00 (33.3%)")
		assert.Contains(t, result, "**Zone 3** (250-300 watts): 25:00 (41.7%)")
		assert.Contains(t, result, "**Zone 4** (300-350 watts): 05:00 (8.3%)")
		assert.Contains(t, result, "**Zone 5** (350-500 watts): 00:00 (0.0%)")
		assert.Contains(t, result, "Data Source: Power Meter")
	})

	t.Run("no zones data", func(t *testing.T) {
		result := formatter.FormatActivityZones(nil)
		assert.Contains(t, result, "❌ **No training zone data available**")
	})

	t.Run("empty zones", func(t *testing.T) {
		zones := &StravaActivityZones{}
		result := formatter.FormatActivityZones(zones)
		assert.Contains(t, result, "❌ No zone data available for this activity")
	})
}