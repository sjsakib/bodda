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

func TestStravaService_GetAthleteProfile_ZoneIntegration(t *testing.T) {
	t.Run("profile with heart rate zones", func(t *testing.T) {
		mockAthlete := StravaAthlete{
			ID:        12345,
			Username:  "testuser",
			Firstname: "Test",
			Lastname:  "User",
			Premium:   true,
		}

		mockZones := []StravaZoneSet{
			{
				CustomZones: false,
				Zones: []StravaZone{
					{Min: 60, Max: 120},
					{Min: 120, Max: 140},
					{Min: 140, Max: 160},
					{Min: 160, Max: 180},
					{Min: 180, Max: 200},
				},
				ResourceState: 3,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch r.URL.Path {
			case "/athlete":
				json.NewEncoder(w).Encode(mockAthlete)
			case "/athlete/zones":
				json.NewEncoder(w).Encode(mockZones)
			default:
				t.Errorf("Unexpected path: %s", r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
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

		result, err := service.GetAthleteProfile(testUser)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.StravaAthlete)
		require.NotNil(t, result.Zones)
		require.NotNil(t, result.Zones.HeartRate)

		// Verify athlete data
		assert.Equal(t, mockAthlete.ID, result.StravaAthlete.ID)
		assert.Equal(t, mockAthlete.Username, result.StravaAthlete.Username)

		// Verify heart rate zones
		assert.Equal(t, 5, len(result.Zones.HeartRate.Zones))
		assert.Equal(t, 60, result.Zones.HeartRate.Zones[0].Min)
		assert.Equal(t, 120, result.Zones.HeartRate.Zones[0].Max)
		assert.Equal(t, 180, result.Zones.HeartRate.Zones[4].Min)
		assert.Equal(t, 200, result.Zones.HeartRate.Zones[4].Max)
	})

	t.Run("profile with power zones", func(t *testing.T) {
		mockAthlete := StravaAthlete{
			ID:        12345,
			Username:  "testuser",
			Firstname: "Test",
			Lastname:  "User",
			Premium:   true,
		}

		mockZones := []StravaZoneSet{
			{
				CustomZones: true,
				Zones: []StravaZone{
					{Min: 0, Max: 150},
					{Min: 150, Max: 200},
					{Min: 200, Max: 250},
					{Min: 250, Max: 300},
					{Min: 300, Max: 350},
					{Min: 350, Max: 400},
					{Min: 400, Max: 500},
				},
				ResourceState: 3,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch r.URL.Path {
			case "/athlete":
				json.NewEncoder(w).Encode(mockAthlete)
			case "/athlete/zones":
				json.NewEncoder(w).Encode(mockZones)
			default:
				t.Errorf("Unexpected path: %s", r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
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

		result, err := service.GetAthleteProfile(testUser)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.StravaAthlete)
		require.NotNil(t, result.Zones)
		require.NotNil(t, result.Zones.Power)

		// Verify power zones
		assert.Equal(t, 7, len(result.Zones.Power.Zones))
		assert.Equal(t, 0, result.Zones.Power.Zones[0].Min)
		assert.Equal(t, 150, result.Zones.Power.Zones[0].Max)
		assert.Equal(t, 400, result.Zones.Power.Zones[6].Min)
		assert.Equal(t, 500, result.Zones.Power.Zones[6].Max)
		assert.True(t, result.Zones.Power.CustomZones)
	})

	t.Run("profile with zones API failure", func(t *testing.T) {
		mockAthlete := StravaAthlete{
			ID:        12345,
			Username:  "testuser",
			Firstname: "Test",
			Lastname:  "User",
			Premium:   true,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/athlete":
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(mockAthlete)
			case "/athlete/zones":
				// Simulate zones API failure
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"message": "Access denied"}`))
			default:
				t.Errorf("Unexpected path: %s", r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
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

		result, err := service.GetAthleteProfile(testUser)

		// Should still succeed even if zones fail
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.StravaAthlete)

		// Verify athlete data is still present
		assert.Equal(t, mockAthlete.ID, result.StravaAthlete.ID)
		assert.Equal(t, mockAthlete.Username, result.StravaAthlete.Username)

		// Zones should be nil due to API failure
		assert.Nil(t, result.Zones)
	})

	t.Run("profile with no zones configured", func(t *testing.T) {
		mockAthlete := StravaAthlete{
			ID:        12345,
			Username:  "testuser",
			Firstname: "Test",
			Lastname:  "User",
			Premium:   false,
		}

		// Empty zones array
		mockZones := []StravaZoneSet{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch r.URL.Path {
			case "/athlete":
				json.NewEncoder(w).Encode(mockAthlete)
			case "/athlete/zones":
				json.NewEncoder(w).Encode(mockZones)
			default:
				t.Errorf("Unexpected path: %s", r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
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

		result, err := service.GetAthleteProfile(testUser)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.StravaAthlete)

		// Verify athlete data
		assert.Equal(t, mockAthlete.ID, result.StravaAthlete.ID)
		assert.Equal(t, mockAthlete.Username, result.StravaAthlete.Username)

		// Zones should be present but empty
		require.NotNil(t, result.Zones)
		assert.Nil(t, result.Zones.HeartRate)
		assert.Nil(t, result.Zones.Power)
		assert.Nil(t, result.Zones.Pace)
	})
}

func TestOutputFormatter_FormatAthleteProfile_ZoneIntegration(t *testing.T) {
	formatter := NewOutputFormatter()

	t.Run("profile with multiple zone types", func(t *testing.T) {
		profile := &StravaAthleteWithZones{
			StravaAthlete: &StravaAthlete{
				ID:        12345,
				Username:  "testuser",
				Firstname: "John",
				Lastname:  "Doe",
				Premium:   true,
			},
			Zones: &StravaAthleteZones{
				HeartRate: &StravaZoneSet{
					CustomZones: false,
					Zones: []StravaZone{
						{Min: 60, Max: 120},
						{Min: 120, Max: 140},
						{Min: 140, Max: 160},
						{Min: 160, Max: 180},
						{Min: 180, Max: 200},
					},
					ResourceState: 3,
				},
				Power: &StravaZoneSet{
					CustomZones: true,
					Zones: []StravaZone{
						{Min: 0, Max: 150},
						{Min: 150, Max: 200},
						{Min: 200, Max: 250},
						{Min: 250, Max: 300},
						{Min: 300, Max: 350},
						{Min: 350, Max: 400},
						{Min: 400, Max: 500},
					},
					ResourceState: 3,
				},
				Pace: &StravaZoneSet{
					CustomZones: false,
					Zones: []StravaZone{
						{Min: 240, Max: 300}, // 4:00-5:00 per km
						{Min: 210, Max: 240}, // 3:30-4:00 per km
						{Min: 180, Max: 210}, // 3:00-3:30 per km
					},
					ResourceState: 3,
				},
			},
		}

		result := formatter.FormatAthleteProfile(profile)

		// Check for all zone types
		assert.Contains(t, result, "🎯 **Training Zones:**")
		assert.Contains(t, result, "**Heart Rate Zones:**")
		assert.Contains(t, result, "Zone 1: 60-120 bpm")
		assert.Contains(t, result, "Zone 5: 180-200 bpm")
		assert.Contains(t, result, "**Power Zones:**")
		assert.Contains(t, result, "Zone 1: 0-150 watts")
		assert.Contains(t, result, "Zone 7: 400-500 watts")
		assert.Contains(t, result, "**Pace Zones:**")
		assert.Contains(t, result, "Zone 1: 5:00-4:00 per km")
		assert.Contains(t, result, "Zone 3: 3:30-3:00 per km")
	})

	t.Run("profile with empty zones", func(t *testing.T) {
		profile := &StravaAthleteWithZones{
			StravaAthlete: &StravaAthlete{
				ID:        12345,
				Username:  "testuser",
				Firstname: "John",
				Lastname:  "Doe",
				Premium:   false,
			},
			Zones: &StravaAthleteZones{
				// All zones are nil
			},
		}

		result := formatter.FormatAthleteProfile(profile)

		assert.Contains(t, result, "🎯 **Training Zones:**")
		assert.Contains(t, result, "No training zones configured")
	})

	t.Run("profile with zones containing -1 values", func(t *testing.T) {
		profile := &StravaAthleteWithZones{
			StravaAthlete: &StravaAthlete{
				ID:        12345,
				Username:  "testuser",
				Firstname: "John",
				Lastname:  "Doe",
				Premium:   true,
			},
			Zones: &StravaAthleteZones{
				HeartRate: &StravaZoneSet{
					CustomZones: false,
					Zones: []StravaZone{
						{Min: 60, Max: 120},   // Valid zone
						{Min: 120, Max: 140},  // Valid zone
						{Min: -1, Max: -1},    // Invalid zone - should be omitted
						{Min: 160, Max: 180},  // Valid zone
						{Min: 180, Max: -1},   // Invalid zone - should be omitted
					},
					ResourceState: 3,
				},
				Power: &StravaZoneSet{
					CustomZones: true,
					Zones: []StravaZone{
						{Min: 0, Max: 150},    // Valid zone
						{Min: -1, Max: 200},   // Invalid zone - should be omitted
						{Min: 200, Max: 250},  // Valid zone
					},
					ResourceState: 3,
				},
			},
		}

		result := formatter.FormatAthleteProfile(profile)

		// Should contain valid zones with sequential numbering
		assert.Contains(t, result, "🎯 **Training Zones:**")
		assert.Contains(t, result, "**Heart Rate Zones:**")
		assert.Contains(t, result, "Zone 1: 60-120 bpm")
		assert.Contains(t, result, "Zone 2: 120-140 bpm")
		assert.Contains(t, result, "Zone 3: 160-180 bpm")
		assert.Contains(t, result, "**Power Zones:**")
		assert.Contains(t, result, "Zone 1: 0-150 watts")
		assert.Contains(t, result, "Zone 2: 200-250 watts")

		// Should NOT contain zones with -1 values
		assert.NotContains(t, result, "-1 bpm")
		assert.NotContains(t, result, "-1 watts")
		assert.NotContains(t, result, "Zone 4: 180--1 bpm")
		assert.NotContains(t, result, "Zone 5: -1-")
	})
}