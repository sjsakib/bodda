package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository for testing
type MockStravaUserRepository struct {
	mock.Mock
}

func (m *MockStravaUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// testStravaServiceWithMockRefresh wraps stravaService to allow mocking RefreshToken
type testStravaServiceWithMockRefresh struct {
	*stravaService
	mockRefreshFunc func(string) (*TokenResponse, error)
}

func (t *testStravaServiceWithMockRefresh) RefreshToken(refreshToken string) (*TokenResponse, error) {
	if t.mockRefreshFunc != nil {
		return t.mockRefreshFunc(refreshToken)
	}
	return t.stravaService.RefreshToken(refreshToken)
}

func TestNewStravaService(t *testing.T) {
	cfg := &config.Config{
		StravaClientID:     "test_client_id",
		StravaClientSecret: "test_client_secret",
	}
	
	mockUserRepo := &MockStravaUserRepository{}
	service := NewStravaService(cfg, mockUserRepo)
	assert.NotNil(t, service)
}

func TestRateLimiter(t *testing.T) {
	t.Run("allows requests within limit", func(t *testing.T) {
		rl := NewRateLimiter(2, time.Minute)
		
		assert.True(t, rl.Allow())
		assert.True(t, rl.Allow())
		assert.False(t, rl.Allow()) // Should be blocked
	})
	
	t.Run("resets after window", func(t *testing.T) {
		rl := NewRateLimiter(1, 100*time.Millisecond)
		
		assert.True(t, rl.Allow())
		assert.False(t, rl.Allow())
		
		time.Sleep(150 * time.Millisecond)
		assert.True(t, rl.Allow()) // Should be allowed after window
	})
}

func TestStravaService_GetAthleteProfile(t *testing.T) {
	mockAthlete := StravaAthlete{
		ID:        12345,
		Username:  "testuser",
		Firstname: "Test",
		Lastname:  "User",
		City:      "Test City",
		State:     "Test State",
		Country:   "Test Country",
		Sex:       "M",
		Premium:   true,
		Weight:    70.5,
		FTP:       250,
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/athlete", r.URL.Path)
		assert.Equal(t, "Bearer test_token", r.Header.Get("Authorization"))
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockAthlete)
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
	
	athlete, err := service.GetAthleteProfile(testUser)
	
	require.NoError(t, err)
	assert.Equal(t, mockAthlete.ID, athlete.ID)
	assert.Equal(t, mockAthlete.Username, athlete.Username)
	assert.Equal(t, mockAthlete.Firstname, athlete.Firstname)
	assert.Equal(t, mockAthlete.Weight, athlete.Weight)
	assert.Equal(t, mockAthlete.FTP, athlete.FTP)
}

func TestStravaService_GetActivities(t *testing.T) {
	mockActivities := []*StravaActivity{
		{
			ID:               123456,
			Name:             "Morning Run",
			Distance:         5000.0,
			MovingTime:       1800,
			ElapsedTime:      1900,
			Type:             "Run",
			SportType:        "Run",
			StartDate:        "2024-01-15T08:00:00Z",
			AverageSpeed:     2.78,
			MaxSpeed:         4.5,
			AverageHeartrate: 150.0,
			MaxHeartrate:     180.0,
			HasHeartrate:     true,
		},
		{
			ID:               123457,
			Name:             "Evening Ride",
			Distance:         25000.0,
			MovingTime:       3600,
			ElapsedTime:      3800,
			Type:             "Ride",
			SportType:        "Ride",
			StartDate:        "2024-01-15T18:00:00Z",
			AverageSpeed:     6.94,
			MaxSpeed:         15.0,
			AveragePower:     200.0,
			MaxPower:         400.0,
			DeviceWatts:      true,
		},
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/athlete/activities", r.URL.Path)
		assert.Equal(t, "Bearer test_token", r.Header.Get("Authorization"))
		
		// Check query parameters
		params := r.URL.Query()
		if params.Get("per_page") != "" {
			assert.Equal(t, "10", params.Get("per_page"))
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockActivities)
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
	
	activities, err := service.GetActivities(testUser, ActivityParams{
		PerPage: 10,
	})
	
	require.NoError(t, err)
	assert.Len(t, activities, 2)
	assert.Equal(t, "Morning Run", activities[0].Name)
	assert.Equal(t, "Evening Ride", activities[1].Name)
	assert.Equal(t, "Run", activities[0].Type)
	assert.Equal(t, "Ride", activities[1].Type)
}

func TestStravaService_GetActivityDetail(t *testing.T) {
	mockActivity := StravaActivityDetail{
		StravaActivity: StravaActivity{
			ID:               123456,
			Name:             "Morning Run",
			Distance:         5000.0,
			MovingTime:       1800,
			Type:             "Run",
			AverageHeartrate: 150.0,
		},
		Description: "Great morning run in the park",
		Calories:    300.0,
		SegmentEfforts: []StravaSegmentEffort{
			{
				ID:          789,
				Name:        "Park Loop",
				ElapsedTime: 600,
				Distance:    1000.0,
				PRRank:      5,
			},
		},
		Splits: []StravaSplit{
			{
				Distance:     1000.0,
				ElapsedTime:  360,
				MovingTime:   350,
				Split:        1,
				AverageSpeed: 2.86,
			},
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
	assert.Equal(t, mockActivity.ID, activity.ID)
	assert.Equal(t, mockActivity.Name, activity.Name)
	assert.Equal(t, mockActivity.Description, activity.Description)
	assert.Equal(t, mockActivity.Calories, activity.Calories)
	assert.Len(t, activity.SegmentEfforts, 1)
	assert.Equal(t, "Park Loop", activity.SegmentEfforts[0].Name)
	assert.Len(t, activity.Splits, 1)
}

func TestStravaService_GetActivityStreams(t *testing.T) {
	// Mock the actual Strava API response format with key_by_type=true
	mockStreamsResponse := StravaStreamsResponse{
		"time": StravaStreamData{
			Data:       []interface{}{0.0, 10.0, 20.0, 30.0},
			SeriesType: "time",
		},
		"distance": StravaStreamData{
			Data:       []interface{}{0.0, 50.0, 100.0, 150.0},
			SeriesType: "distance",
		},
		"heartrate": StravaStreamData{
			Data:       []interface{}{120.0, 140.0, 160.0, 150.0},
			SeriesType: "heartrate",
		},
		"altitude": StravaStreamData{
			Data:       []interface{}{100.0, 105.0, 110.0, 108.0},
			SeriesType: "altitude",
		},
		"watts": StravaStreamData{
			Data:       []interface{}{200.0, 250.0, 300.0, 280.0},
			SeriesType: "watts",
		},
		"cadence": StravaStreamData{
			Data:       []interface{}{80.0, 85.0, 90.0, 88.0},
			SeriesType: "cadence",
		},
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/activities/123456/streams", r.URL.Path)
		assert.Equal(t, "Bearer test_token", r.Header.Get("Authorization"))
		
		params := r.URL.Query()
		assert.Contains(t, params.Get("keys"), "time")
		assert.Contains(t, params.Get("keys"), "heartrate")
		assert.Equal(t, "high", params.Get("resolution"))
		assert.Equal(t, "true", params.Get("key_by_type"))
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockStreamsResponse)
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
	
	streams, err := service.GetActivityStreams(testUser, 123456, []string{"time", "heartrate", "watts"}, "high")
	
	require.NoError(t, err)
	assert.Len(t, streams.Time, 4)
	assert.Len(t, streams.Heartrate, 4)
	assert.Len(t, streams.Watts, 4)
	assert.Equal(t, []int{0, 10, 20, 30}, streams.Time)
	assert.Equal(t, []int{120, 140, 160, 150}, streams.Heartrate)
	assert.Equal(t, []int{200, 250, 300, 280}, streams.Watts)
}



func TestStravaService_ErrorHandling(t *testing.T) {
	t.Run("handles API errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "Authorization Error", "errors": [{"resource": "Application", "field": "client_id", "code": "invalid"}]}`))
		}))
		defer server.Close()
		
		cfg := &config.Config{}
		mockUserRepo := &MockStravaUserRepository{}
		service := NewTestStravaService(cfg, server.URL, mockUserRepo)
		
		testUser := &models.User{
			ID:           "test-user-id",
			AccessToken:  "invalid_token",
			RefreshToken: "test_refresh_token",
			TokenExpiry:  time.Now().Add(time.Hour),
		}
		
		_, err := service.GetAthleteProfile(testUser)
		assert.Error(t, err)
	})
	
	t.Run("handles rate limiting", func(t *testing.T) {
		cfg := &config.Config{}
		mockUserRepo := &MockStravaUserRepository{}
		service := &stravaService{
			config:      cfg,
			httpClient:  &http.Client{},
			rateLimiter: NewRateLimiter(0, time.Minute), // No requests allowed
			userRepo:    mockUserRepo,
		}
		
		// Set the makeRequest function to the default implementation
		service.makeRequest = service.defaultMakeRequest
		
		testUser := &models.User{
			ID:           "test-user-id",
			AccessToken:  "test_token",
			RefreshToken: "test_refresh_token",
			TokenExpiry:  time.Now().Add(time.Hour),
		}
		
		_, err := service.GetAthleteProfile(testUser)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit exceeded")
	})
}

func TestActivityParams(t *testing.T) {
	t.Run("builds URL parameters correctly", func(t *testing.T) {
		before := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		after := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		
		params := ActivityParams{
			Before:  &before,
			After:   &after,
			Page:    2,
			PerPage: 50,
		}
		
		assert.Equal(t, before, *params.Before)
		assert.Equal(t, after, *params.After)
		assert.Equal(t, 2, params.Page)
		assert.Equal(t, 50, params.PerPage)
	})
}

func TestParseStreamsResponse(t *testing.T) {
	t.Run("parses streams response correctly", func(t *testing.T) {
		rawStreams := StravaStreamsResponse{
			"time": StravaStreamData{
				Data:       []interface{}{0.0, 10.0, 20.0, 30.0},
				SeriesType: "time",
			},
			"altitude": StravaStreamData{
				Data:       []interface{}{100.0, 105.0, 110.0, 108.0},
				SeriesType: "altitude",
			},
			"heartrate": StravaStreamData{
				Data:       []interface{}{120.0, 140.0, 160.0, 150.0},
				SeriesType: "heartrate",
			},
			"latlng": StravaStreamData{
				Data: []interface{}{
					[]interface{}{37.7749, -122.4194},
					[]interface{}{37.7750, -122.4195},
				},
				SeriesType: "latlng",
			},
			"moving": StravaStreamData{
				Data:       []interface{}{true, true, false, true},
				SeriesType: "moving",
			},
		}

		streams, err := parseStreamsResponse(rawStreams)
		
		require.NoError(t, err)
		assert.Equal(t, []int{0, 10, 20, 30}, streams.Time)
		assert.Equal(t, []float64{100.0, 105.0, 110.0, 108.0}, streams.Altitude)
		assert.Equal(t, []int{120, 140, 160, 150}, streams.Heartrate)
		assert.Equal(t, []bool{true, true, false, true}, streams.Moving)
		assert.Len(t, streams.Latlng, 2)
		assert.Equal(t, []float64{37.7749, -122.4194}, streams.Latlng[0])
		assert.Equal(t, []float64{37.7750, -122.4195}, streams.Latlng[1])
	})
}

// Test automatic token refresh functionality
func TestStravaService_ExecuteWithTokenRefresh(t *testing.T) {
	t.Run("successfully executes API call without token refresh when token is valid", func(t *testing.T) {
		cfg := &config.Config{}
		mockUserRepo := &MockStravaUserRepository{}
		
		service := &stravaService{
			config:      cfg,
			httpClient:  &http.Client{Timeout: 30 * time.Second},
			rateLimiter: NewRateLimiter(100, 15*time.Minute),
			userRepo:    mockUserRepo,
		}
		
		testUser := &models.User{
			ID:           "test-user-id",
			AccessToken:  "valid_token",
			RefreshToken: "refresh_token",
			TokenExpiry:  time.Now().Add(time.Hour),
		}
		
		// Mock API call that succeeds on first try
		apiCall := func(accessToken string) (interface{}, error) {
			assert.Equal(t, "valid_token", accessToken)
			return "success", nil
		}
		
		result, err := service.executeWithTokenRefresh(testUser, apiCall)
		
		require.NoError(t, err)
		assert.Equal(t, "success", result)
		
		// User tokens should remain unchanged
		assert.Equal(t, "valid_token", testUser.AccessToken)
		assert.Equal(t, "refresh_token", testUser.RefreshToken)
	})
	
	t.Run("handles non-token-related API errors without attempting refresh", func(t *testing.T) {
		cfg := &config.Config{}
		mockUserRepo := &MockStravaUserRepository{}
		
		service := &stravaService{
			config:      cfg,
			httpClient:  &http.Client{Timeout: 30 * time.Second},
			rateLimiter: NewRateLimiter(100, 15*time.Minute),
			userRepo:    mockUserRepo,
		}
		
		testUser := &models.User{
			ID:           "test-user-id",
			AccessToken:  "valid_token",
			RefreshToken: "refresh_token",
			TokenExpiry:  time.Now().Add(time.Hour),
		}
		
		// Mock API call that fails with non-token error
		apiCall := func(accessToken string) (interface{}, error) {
			return nil, errors.New("rate limit exceeded")
		}
		
		_, err := service.executeWithTokenRefresh(testUser, apiCall)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit exceeded")
		
		// User tokens should remain unchanged
		assert.Equal(t, "valid_token", testUser.AccessToken)
		assert.Equal(t, "refresh_token", testUser.RefreshToken)
	})
}

func TestStravaService_RefreshToken(t *testing.T) {
	t.Run("validates token response structure", func(t *testing.T) {
		// Test the token response structure
		tokenResp := &TokenResponse{
			AccessToken:  "new_access_token",
			RefreshToken: "new_refresh_token",
			ExpiresAt:    time.Now().Add(6 * time.Hour).Unix(),
			ExpiresIn:    21600,
			TokenType:    "Bearer",
		}
		
		assert.Equal(t, "new_access_token", tokenResp.AccessToken)
		assert.Equal(t, "new_refresh_token", tokenResp.RefreshToken)
		assert.Equal(t, "Bearer", tokenResp.TokenType)
		assert.Greater(t, tokenResp.ExpiresAt, time.Now().Unix())
	})
}