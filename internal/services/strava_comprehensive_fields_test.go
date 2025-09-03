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

func TestStravaService_GetActivityDetail_ComprehensiveFields(t *testing.T) {
	// Test comprehensive activity detail with all possible fields from Strava API
	mockActivity := StravaActivityDetail{
		StravaActivity: StravaActivity{
			ID:                 987654321,
			Name:               "Comprehensive Test Activity",
			Distance:           21097.5, // Half marathon
			MovingTime:         7200,    // 2 hours
			ElapsedTime:        7500,    // 2 hours 5 minutes
			Type:               "Run",
			SportType:          "Run",
			AverageHeartrate:   155.0,
			MaxHeartrate:       185.0,
			AveragePower:       280.0,
			MaxPower:           450.0,
			StartDateLocal:     "2024-01-20T07:00:00Z",
			TotalElevationGain: 350.0,
			AverageSpeed:       2.93, // ~6:50/mile pace
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
			Flagged:            false,
		},
		ResourceState:                  3,
		Description:                    "Perfect half marathon with comprehensive tracking data",
		Calories:                       1250.0,
		StartLatlng:                    []float64{40.7589, -73.9851}, // NYC Central Park
		EndLatlng:                      []float64{40.7614, -73.9776},
		LocationCity:                   "New York",
		LocationState:                  "NY",
		LocationCountry:                "United States",
		AchievementCount:               5,
		TotalPhotoCount:                3,
		HasKudoed:                      false,
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
				Distance:                  1609.34, // 1 mile
				ElapsedTime:               410,     // 6:50
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
				Distance:                  1609.34, // 1 mile
				ElapsedTime:               415,     // 6:55
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
				ID:            2001,
				ResourceState: 3,
				Name:          "1 mile",
				Activity: StravaActivityRef{
					ID:            987654321,
					ResourceState: 2,
				},
				Athlete: StravaAthleteRef{
					ID:            12345,
					ResourceState: 2,
				},
				ElapsedTime:      395,
				MovingTime:       390,
				StartDate:        "2024-01-20T07:05:00Z",
				StartDateLocal:   "2024-01-20T07:05:00Z",
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
			{
				ID:            2002,
				ResourceState: 3,
				Name:          "5k",
				Activity: StravaActivityRef{
					ID:            987654321,
					ResourceState: 2,
				},
				Athlete: StravaAthleteRef{
					ID:            12345,
					ResourceState: 2,
				},
				ElapsedTime:      1245,
				MovingTime:       1230,
				StartDate:        "2024-01-20T07:00:00Z",
				StartDateLocal:   "2024-01-20T07:00:00Z",
				Distance:         5000.0,
				StartIndex:       0,
				EndIndex:         1245,
				AverageHeartrate: 158.0,
				MaxHeartrate:     172.0,
				AveragePower:     285.0,
				MaxPower:         350.0,
				AverageCadence:   183.0,
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
		Laps: []StravaLap{
			{
				ID:            3001,
				ResourceState: 3,
				Name:          "Lap 1",
				Activity: StravaActivityRef{
					ID:            987654321,
					ResourceState: 2,
				},
				Athlete: StravaAthleteRef{
					ID:            12345,
					ResourceState: 2,
				},
				ElapsedTime:        3600,
				MovingTime:         3550,
				StartDate:          "2024-01-20T07:00:00Z",
				StartDateLocal:     "2024-01-20T07:00:00Z",
				Distance:           10548.75, // Half of half marathon
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
				ID:            3002,
				ResourceState: 3,
				Name:          "Lap 2",
				Activity: StravaActivityRef{
					ID:            987654321,
					ResourceState: 2,
				},
				Athlete: StravaAthleteRef{
					ID:            12345,
					ResourceState: 2,
				},
				ElapsedTime:        3900,
				MovingTime:         3850,
				StartDate:          "2024-01-20T08:00:00Z",
				StartDateLocal:     "2024-01-20T08:00:00Z",
				Distance:           10548.75, // Second half
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
			TrendStats: StravaTrendStats{
				Speeds:               []float64{2.70, 2.75, 2.80, 2.85, 2.93},
				CurrentActivityIndex: 4,
				MinSpeed:             2.70,
				MidSpeed:             2.80,
				MaxSpeed:             2.93,
				Direction:            1, // Improving
			},
		},
		Gear: StravaGear{
			ID:          "g456",
			Name:        "Nike Vaporfly Next% 2",
			BrandName:   "Nike",
			ModelName:   "Air Zoom Alphafly NEXT%",
			Distance:    750000.0, // 750km
			Description: "Race day shoes",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/activities/987654321", r.URL.Path)
		assert.Equal(t, "Bearer comprehensive_test_token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockActivity)
	}))
	defer server.Close()

	cfg := &config.Config{}
	mockUserRepo := &MockStravaUserRepository{}
	service := NewTestStravaService(cfg, server.URL, mockUserRepo)

	testUser := &models.User{
		ID:           "comprehensive-test-user",
		AccessToken:  "comprehensive_test_token",
		RefreshToken: "comprehensive_refresh_token",
		TokenExpiry:  time.Now().Add(time.Hour),
	}

	activity, err := service.GetActivityDetail(testUser, 987654321)

	require.NoError(t, err)
	require.NotNil(t, activity)

	// Test all comprehensive fields
	t.Run("basic activity fields", func(t *testing.T) {
		assert.Equal(t, int64(987654321), activity.ID)
		assert.Equal(t, "Comprehensive Test Activity", activity.Name)
		assert.Equal(t, 21097.5, activity.Distance)
		assert.Equal(t, 7200, activity.MovingTime)
		assert.Equal(t, 7500, activity.ElapsedTime)
		assert.Equal(t, "Run", activity.Type)
		assert.Equal(t, 2016.0, activity.Kilojoules)
		assert.True(t, activity.DeviceWatts)
	})

	t.Run("enhanced metadata fields", func(t *testing.T) {
		assert.Equal(t, 3, activity.ResourceState)
		assert.Equal(t, "Perfect half marathon with comprehensive tracking data", activity.Description)
		assert.Equal(t, 1250.0, activity.Calories)
		assert.Equal(t, int64(123456789012), activity.UploadID)
		assert.Equal(t, "123456789012", activity.UploadIDStr)
		assert.Equal(t, "garmin_activity_987654321", activity.ExternalID)
		assert.Equal(t, "Garmin Forerunner 965", activity.DeviceName)
		assert.False(t, activity.FromAcceptedTag)
	})

	t.Run("location and coordinates", func(t *testing.T) {
		assert.Equal(t, []float64{40.7589, -73.9851}, activity.StartLatlng)
		assert.Equal(t, []float64{40.7614, -73.9776}, activity.EndLatlng)
		assert.Equal(t, "New York", activity.LocationCity)
		assert.Equal(t, "NY", activity.LocationState)
		assert.Equal(t, "United States", activity.LocationCountry)
	})

	t.Run("social and achievement metrics", func(t *testing.T) {
		assert.Equal(t, 5, activity.AchievementCount)
		assert.Equal(t, 3, activity.TotalPhotoCount)
		assert.False(t, activity.HasKudoed)
		assert.Equal(t, 42, activity.KudosCount)
		assert.Equal(t, 8, activity.CommentCount)
		assert.Equal(t, 3, activity.PRCount)
	})

	t.Run("performance metrics", func(t *testing.T) {
		assert.Equal(t, 182.0, activity.AverageCadence)
		assert.Equal(t, 295.0, activity.WeightedAverageWatts)
		assert.Equal(t, 142.0, activity.SufferScore)
		assert.Equal(t, 8, activity.PerceivedExertion)
		assert.True(t, activity.PreferPerceivedExertion)
	})

	t.Run("privacy and display options", func(t *testing.T) {
		assert.False(t, activity.HeartrateOptOut)
		assert.False(t, activity.DisplayHideHeartrateOption)
		assert.False(t, activity.HideFromHome)
	})

	t.Run("athlete reference", func(t *testing.T) {
		assert.Equal(t, int64(12345), activity.Athlete.ID)
		assert.Equal(t, 2, activity.Athlete.ResourceState)
	})

	t.Run("enhanced standard splits", func(t *testing.T) {
		require.Len(t, activity.SplitsStandard, 2)
		
		// First split
		split1 := activity.SplitsStandard[0]
		assert.Equal(t, 1609.34, split1.Distance)
		assert.Equal(t, 410, split1.ElapsedTime)
		assert.Equal(t, 405, split1.MovingTime)
		assert.Equal(t, 1, split1.Split)
		assert.Equal(t, 3.97, split1.AverageSpeed)
		assert.Equal(t, 4.02, split1.AverageGradeAdjustedSpeed)
		assert.Equal(t, 148.0, split1.AverageHeartrate)
		assert.Equal(t, 275.0, split1.AveragePower)
		assert.Equal(t, 180.0, split1.AverageCadence)
		assert.Equal(t, 12.0, split1.ElevationDifference)
		assert.Equal(t, 2, split1.PaceZone)
	})

	t.Run("enhanced best efforts", func(t *testing.T) {
		require.Len(t, activity.BestEfforts, 2)
		
		// First best effort (1 mile)
		effort1 := activity.BestEfforts[0]
		assert.Equal(t, int64(2001), effort1.ID)
		assert.Equal(t, 3, effort1.ResourceState)
		assert.Equal(t, "1 mile", effort1.Name)
		assert.Equal(t, int64(987654321), effort1.Activity.ID)
		assert.Equal(t, 2, effort1.Activity.ResourceState)
		assert.Equal(t, int64(12345), effort1.Athlete.ID)
		assert.Equal(t, 2, effort1.Athlete.ResourceState)
		assert.Equal(t, 395, effort1.ElapsedTime)
		assert.Equal(t, 390, effort1.MovingTime)
		assert.Equal(t, "2024-01-20T07:05:00Z", effort1.StartDate)
		assert.Equal(t, "2024-01-20T07:05:00Z", effort1.StartDateLocal)
		assert.Equal(t, 1609.34, effort1.Distance)
		assert.Equal(t, 300, effort1.StartIndex)
		assert.Equal(t, 695, effort1.EndIndex)
		assert.Equal(t, 162.0, effort1.AverageHeartrate)
		assert.Equal(t, 175.0, effort1.MaxHeartrate)
		assert.Equal(t, 295.0, effort1.AveragePower)
		assert.Equal(t, 380.0, effort1.MaxPower)
		assert.Equal(t, 185.0, effort1.AverageCadence)
		assert.Equal(t, 2, effort1.PRRank)
		require.Len(t, effort1.Achievements, 2)
		assert.Equal(t, "pr", effort1.Achievements[0].Type)
		assert.Equal(t, 2, effort1.Achievements[0].Rank)
	})

	t.Run("enhanced laps", func(t *testing.T) {
		require.Len(t, activity.Laps, 2)
		
		// First lap
		lap1 := activity.Laps[0]
		assert.Equal(t, int64(3001), lap1.ID)
		assert.Equal(t, 3, lap1.ResourceState)
		assert.Equal(t, "Lap 1", lap1.Name)
		assert.Equal(t, int64(987654321), lap1.Activity.ID)
		assert.Equal(t, 2, lap1.Activity.ResourceState)
		assert.Equal(t, int64(12345), lap1.Athlete.ID)
		assert.Equal(t, 2, lap1.Athlete.ResourceState)
		assert.Equal(t, 3600, lap1.ElapsedTime)
		assert.Equal(t, 3550, lap1.MovingTime)
		assert.Equal(t, "2024-01-20T07:00:00Z", lap1.StartDate)
		assert.Equal(t, "2024-01-20T07:00:00Z", lap1.StartDateLocal)
		assert.Equal(t, 10548.75, lap1.Distance)
		assert.Equal(t, 0, lap1.StartIndex)
		assert.Equal(t, 3600, lap1.EndIndex)
		assert.Equal(t, 175.0, lap1.TotalElevationGain)
		assert.Equal(t, 2.97, lap1.AverageSpeed)
		assert.Equal(t, 4.8, lap1.MaxSpeed)
		assert.Equal(t, 152.0, lap1.AverageHeartrate)
		assert.Equal(t, 168.0, lap1.MaxHeartrate)
		assert.Equal(t, 275.0, lap1.AveragePower)
		assert.Equal(t, 420.0, lap1.MaxPower)
		assert.Equal(t, 181.0, lap1.AverageCadence)
		assert.True(t, lap1.DeviceWatts)
		assert.Equal(t, 275.0, lap1.AverageWatts)
		assert.Equal(t, 1, lap1.LapIndex)
		assert.Equal(t, 1, lap1.Split)
		assert.Equal(t, 2, lap1.PaceZone)
	})

	t.Run("similar activities with trend data", func(t *testing.T) {
		similar := activity.SimilarActivities
		assert.Equal(t, 25, similar.EffortCount)
		assert.Equal(t, 2.85, similar.AverageSpeed)
		assert.Equal(t, 2.65, similar.MinAverageSpeed)
		assert.Equal(t, 2.85, similar.MidAverageSpeed)
		assert.Equal(t, 3.05, similar.MaxAverageSpeed)
		assert.Equal(t, 3, similar.PRRank)
		assert.Equal(t, "This is your 5th half marathon this year!", similar.FrequencyMilestone)
		assert.Equal(t, 2, similar.ResourceState)
		
		// Test trend stats
		trend := similar.TrendStats
		assert.Equal(t, []float64{2.70, 2.75, 2.80, 2.85, 2.93}, trend.Speeds)
		assert.Equal(t, 4, trend.CurrentActivityIndex)
		assert.Equal(t, 2.70, trend.MinSpeed)
		assert.Equal(t, 2.80, trend.MidSpeed)
		assert.Equal(t, 2.93, trend.MaxSpeed)
		assert.Equal(t, 1, trend.Direction) // Improving
	})

	t.Run("available zones", func(t *testing.T) {
		assert.Equal(t, []string{"heartrate", "power", "pace"}, activity.AvailableZones)
	})
}