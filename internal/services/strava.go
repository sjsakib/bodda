package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"
)

// Strava API data models
type StravaAthlete struct {
	ID        int64   `json:"id"`
	Username  string  `json:"username"`
	Firstname string  `json:"firstname"`
	Lastname  string  `json:"lastname"`
	City      string  `json:"city"`
	State     string  `json:"state"`
	Country   string  `json:"country"`
	Sex       string  `json:"sex"`
	Premium   bool    `json:"premium"`
	Summit    bool    `json:"summit"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	Profile   string  `json:"profile"`
	Weight    float64 `json:"weight"`
	FTP       int     `json:"ftp"`
}

type StravaActivity struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	Distance           float64 `json:"distance"`
	MovingTime         int     `json:"moving_time"`
	ElapsedTime        int     `json:"elapsed_time"`
	TotalElevationGain float64 `json:"total_elevation_gain"`
	Type               string  `json:"type"`
	SportType          string  `json:"sport_type"`
	StartDate          string  `json:"start_date"`
	StartDateLocal     string  `json:"start_date_local"`
	Timezone           string  `json:"timezone"`
	AverageSpeed       float64 `json:"average_speed"`
	MaxSpeed           float64 `json:"max_speed"`
	AverageHeartrate   float64 `json:"average_heartrate"`
	MaxHeartrate       float64 `json:"max_heartrate"`
	AveragePower       float64 `json:"average_power"`
	MaxPower           float64 `json:"max_power"`
	Kilojoules         float64 `json:"kilojoules"`
	DeviceWatts        bool    `json:"device_watts"`
	HasHeartrate       bool    `json:"has_heartrate"`
	ElevHigh           float64 `json:"elev_high"`
	ElevLow            float64 `json:"elev_low"`
	PRCount            int     `json:"pr_count"`
	KudosCount         int     `json:"kudos_count"`
	CommentCount       int     `json:"comment_count"`
	AthleteCount       int     `json:"athlete_count"`
	PhotoCount         int     `json:"photo_count"`
	Trainer            bool    `json:"trainer"`
	Commute            bool    `json:"commute"`
	Manual             bool    `json:"manual"`
	Private            bool    `json:"private"`
	Flagged            bool    `json:"flagged"`
	WorkoutType        int     `json:"workout_type"`
	AverageTemp        float64 `json:"average_temp"`
}

type StravaActivityDetail struct {
	StravaActivity
	Description         string                `json:"description"`
	Calories            float64               `json:"calories"`
	SegmentEfforts      []StravaSegmentEffort `json:"segment_efforts"`
	Splits              []StravaSplit         `json:"splits_metric"`
	Laps                []StravaLap           `json:"laps"`
	Gear                StravaGear            `json:"gear"`
	Photos              StravaPhotos          `json:"photos"`
	HighlightedKudosers []StravaAthlete       `json:"highlighted_kudosers"`
}

type StravaSegmentEffort struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	ElapsedTime      int     `json:"elapsed_time"`
	MovingTime       int     `json:"moving_time"`
	StartDate        string  `json:"start_date"`
	Distance         float64 `json:"distance"`
	AverageHeartrate float64 `json:"average_heartrate"`
	MaxHeartrate     float64 `json:"max_heartrate"`
	PRRank           int     `json:"pr_rank"`
	KOMRank          int     `json:"kom_rank"`
}

type StravaSplit struct {
	Distance            float64 `json:"distance"`
	ElapsedTime         int     `json:"elapsed_time"`
	ElevationDifference float64 `json:"elevation_difference"`
	MovingTime          int     `json:"moving_time"`
	Split               int     `json:"split"`
	AverageSpeed        float64 `json:"average_speed"`
	PaceZone            int     `json:"pace_zone"`
}

type StravaLap struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	ElapsedTime        int     `json:"elapsed_time"`
	MovingTime         int     `json:"moving_time"`
	StartDate          string  `json:"start_date"`
	Distance           float64 `json:"distance"`
	StartIndex         int     `json:"start_index"`
	EndIndex           int     `json:"end_index"`
	TotalElevationGain float64 `json:"total_elevation_gain"`
	AverageSpeed       float64 `json:"average_speed"`
	MaxSpeed           float64 `json:"max_speed"`
	AverageHeartrate   float64 `json:"average_heartrate"`
	MaxHeartrate       float64 `json:"max_heartrate"`
	AveragePower       float64 `json:"average_power"`
	MaxPower           float64 `json:"max_power"`
	LapIndex           int     `json:"lap_index"`
	Split              int     `json:"split"`
}

type StravaGear struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Distance    float64 `json:"distance"`
	BrandName   string  `json:"brand_name"`
	ModelName   string  `json:"model_name"`
	FrameType   int     `json:"frame_type"`
	Description string  `json:"description"`
}

type StravaPhotos struct {
	Primary StravaPhoto `json:"primary"`
	Count   int         `json:"count"`
}

type StravaPhoto struct {
	ID       int64             `json:"id"`
	Source   int               `json:"source"`
	UniqueID string            `json:"unique_id"`
	URLs     map[string]string `json:"urls"`
}

// StravaStreamData represents a single stream's data structure
type StravaStreamData struct {
	Data       interface{} `json:"data"`
	SeriesType string      `json:"series_type"`
}

// StravaStreamsResponse represents the raw response from Strava API when key_by_type=true
type StravaStreamsResponse map[string]StravaStreamData

// StravaStreams represents the parsed streams data in a more usable format
type StravaStreams struct {
	Time           []int       `json:"time,omitempty"`
	Distance       []float64   `json:"distance,omitempty"`
	Latlng         [][]float64 `json:"latlng,omitempty"`
	Altitude       []float64   `json:"altitude,omitempty"`
	VelocitySmooth []float64   `json:"velocity_smooth,omitempty"`
	Heartrate      []int       `json:"heartrate,omitempty"`
	Cadence        []int       `json:"cadence,omitempty"`
	Watts          []int       `json:"watts,omitempty"`
	Temp           []int       `json:"temp,omitempty"`
	Moving         []bool      `json:"moving,omitempty"`
	GradeSmooth    []float64   `json:"grade_smooth,omitempty"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// StravaError represents errors returned by the Strava API
type StravaError struct {
	Message string             `json:"message"`
	Errors  []StravaFieldError `json:"errors"`
	Code    string             `json:"code"`
}

type StravaFieldError struct {
	Resource string `json:"resource"`
	Field    string `json:"field"`
	Code     string `json:"code"`
}

func (e *StravaError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("Strava API error: %s", e.Message)
	}
	if len(e.Errors) > 0 {
		return fmt.Sprintf("Strava API error: %s", e.Errors[0].Code)
	}
	return "Unknown Strava API error"
}

// Custom error types for better error handling
var (
	ErrRateLimitExceeded  = errors.New("Strava API rate limit exceeded")
	ErrTokenExpired       = errors.New("Strava access token expired")
	ErrInvalidToken       = errors.New("Strava access token invalid")
	ErrActivityNotFound   = errors.New("Strava activity not found")
	ErrNetworkTimeout     = errors.New("Network request timed out")
	ErrServiceUnavailable = errors.New("Strava service temporarily unavailable")
)

type ActivityParams struct {
	Before  *time.Time
	After   *time.Time
	Page    int
	PerPage int
}

// StravaService handles all Strava API interactions
type StravaService interface {
	GetAthleteProfile(user *models.User) (*StravaAthlete, error)
	GetActivities(user *models.User, params ActivityParams) ([]*StravaActivity, error)
	GetActivityDetail(user *models.User, activityID int64) (*StravaActivityDetail, error)
	GetActivityStreams(user *models.User, activityID int64, streamTypes []string, resolution string) (*StravaStreams, error)
	RefreshToken(refreshToken string) (*TokenResponse, error)
}

// UserRepositoryInterface defines the interface for user repository operations needed by StravaService
type UserRepositoryInterface interface {
	Update(ctx context.Context, user *models.User) error
}

type stravaService struct {
	config      *config.Config
	httpClient  *http.Client
	rateLimiter *RateLimiter
	userRepo    UserRepositoryInterface
	makeRequest func(method, endpoint, accessToken string, params url.Values) ([]byte, error)
}

// RateLimiter implements a simple rate limiter for Strava API
type RateLimiter struct {
	requests    []time.Time
	maxRequests int
	window      time.Duration
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests:    make([]time.Time, 0),
		maxRequests: maxRequests,
		window:      window,
	}
}

func (rl *RateLimiter) Allow() bool {
	now := time.Now()

	// Remove old requests outside the window
	cutoff := now.Add(-rl.window)
	validRequests := make([]time.Time, 0)
	for _, req := range rl.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	rl.requests = validRequests

	// Check if we can make a new request
	if len(rl.requests) < rl.maxRequests {
		rl.requests = append(rl.requests, now)
		return true
	}

	return false
}

func NewStravaService(cfg *config.Config, userRepo UserRepositoryInterface) StravaService {
	// Strava API limits: 100 requests per 15 minutes, 1000 requests per day
	// We'll implement the 15-minute limit here
	rateLimiter := NewRateLimiter(100, 15*time.Minute)

	service := &stravaService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: rateLimiter,
		userRepo:    userRepo,
	}

	// Set the default makeRequest implementation
	service.makeRequest = service.defaultMakeRequest

	return service
}

// NewTestStravaService creates a Strava service for testing with a custom base URL
func NewTestStravaService(cfg *config.Config, baseURL string, userRepo UserRepositoryInterface) StravaService {
	service := &stravaService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: NewRateLimiter(100, 15*time.Minute),
		userRepo:    userRepo,
	}

	// Override makeRequest for testing
	service.makeRequest = func(method, endpoint, accessToken string, params url.Values) ([]byte, error) {
		fullURL := baseURL + endpoint
		if len(params) > 0 {
			fullURL += "?" + params.Encode()
		}

		req, err := http.NewRequest(method, fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := service.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}

		return body, nil
	}

	return service
}

// executeWithTokenRefresh wraps API calls with automatic token refresh on 401 errors
func (s *stravaService) executeWithTokenRefresh(user *models.User, apiCall func(string) (interface{}, error)) (interface{}, error) {
	// First attempt with current token
	result, err := apiCall(user.AccessToken)
	
	// If we get a token expired error, try to refresh
	if err != nil && (errors.Is(err, ErrTokenExpired) || errors.Is(err, ErrInvalidToken)) {
		log.Printf("Token expired for user %s, attempting refresh", user.ID)
		
		// Attempt to refresh the token
		tokenResp, refreshErr := s.RefreshToken(user.RefreshToken)
		if refreshErr != nil {
			log.Printf("Token refresh failed for user %s: %v", user.ID, refreshErr)
			return nil, fmt.Errorf("failed to refresh Strava token: %w", refreshErr)
		}
		
		// Update user with new tokens
		user.AccessToken = tokenResp.AccessToken
		user.RefreshToken = tokenResp.RefreshToken
		user.TokenExpiry = time.Unix(tokenResp.ExpiresAt, 0)
		
		// Save updated tokens to database
		ctx := context.Background()
		if updateErr := s.userRepo.Update(ctx, user); updateErr != nil {
			log.Printf("Failed to update user tokens in database for user %s: %v", user.ID, updateErr)
			return nil, fmt.Errorf("failed to save refreshed tokens: %w", updateErr)
		}
		
		log.Printf("Successfully refreshed tokens for user %s", user.ID)
		
		// Retry the original API call with new token
		result, err = apiCall(user.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("API call failed even after token refresh: %w", err)
		}
	}
	
	return result, err
}

func (s *stravaService) defaultMakeRequest(method, endpoint string, accessToken string, params url.Values) ([]byte, error) {
	// Check rate limit
	if !s.rateLimiter.Allow() {
		log.Printf("Strava API rate limit exceeded for endpoint: %s", endpoint)
		return nil, ErrRateLimitExceeded
	}

	baseURL := "https://www.strava.com/api/v3"
	fullURL := baseURL + endpoint

	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("Sending http request")
	fmt.Printf("full url: %s", fullURL)
	fmt.Printf("Token: %s", "Bearer "+accessToken)
	fmt.Printf("Method: %s", method)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		// Check for timeout errors
		if strings.Contains(err.Error(), "timeout") {
			return nil, ErrNetworkTimeout
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle different HTTP status codes
	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusUnauthorized:
		log.Printf("Strava API returned 401 for endpoint: %s", endpoint)
		return nil, ErrTokenExpired
	case http.StatusForbidden:
		log.Printf("Strava API returned 403 for endpoint: %s", endpoint)
		return nil, ErrInvalidToken
	case http.StatusNotFound:
		log.Printf("Strava API returned 404 for endpoint: %s", endpoint)
		return nil, ErrActivityNotFound
	case http.StatusTooManyRequests:
		log.Printf("Strava API rate limit hit for endpoint: %s", endpoint)
		return nil, ErrRateLimitExceeded
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		log.Printf("Strava API server error %d for endpoint: %s", resp.StatusCode, endpoint)
		return nil, ErrServiceUnavailable
	default:
		// Try to parse Strava error response
		var stravaErr StravaError
		if err := json.Unmarshal(body, &stravaErr); err == nil && stravaErr.Message != "" {
			log.Printf("Strava API error for endpoint %s: %s", endpoint, stravaErr.Message)
			return nil, &stravaErr
		}

		log.Printf("Strava API returned unexpected status %d for endpoint: %s, body: %s", resp.StatusCode, endpoint, string(body))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
}

func (s *stravaService) GetAthleteProfile(user *models.User) (*StravaAthlete, error) {
	apiCall := func(accessToken string) (interface{}, error) {
		body, err := s.makeRequest("GET", "/athlete", accessToken, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get athlete profile: %w", err)
		}

		// fmt.Println("body", )

		var athlete StravaAthlete
		if err := json.Unmarshal(body, &athlete); err != nil {
			return nil, fmt.Errorf("failed to parse athlete profile: %w", err)
		}

		fmt.Printf("----------body----------------\n%s", string(body))

		fmt.Printf("-----------xx----------------\nFTP: %d", athlete.FTP)

		return &athlete, nil
	}

	result, err := s.executeWithTokenRefresh(user, apiCall)
	if err != nil {
		return nil, err
	}

	return result.(*StravaAthlete), nil
}

func (s *stravaService) GetActivities(user *models.User, params ActivityParams) ([]*StravaActivity, error) {
	apiCall := func(accessToken string) (interface{}, error) {
		urlParams := url.Values{}

		if params.Before != nil {
			urlParams.Set("before", strconv.FormatInt(params.Before.Unix(), 10))
		}
		if params.After != nil {
			urlParams.Set("after", strconv.FormatInt(params.After.Unix(), 10))
		}
		if params.Page > 0 {
			urlParams.Set("page", strconv.Itoa(params.Page))
		}
		if params.PerPage > 0 {
			urlParams.Set("per_page", strconv.Itoa(params.PerPage))
		}

		body, err := s.makeRequest("GET", "/athlete/activities", accessToken, urlParams)
		if err != nil {
			return nil, fmt.Errorf("failed to get activities: %w", err)
		}

		var activities []*StravaActivity
		if err := json.Unmarshal(body, &activities); err != nil {
			return nil, fmt.Errorf("failed to parse activities: %w", err)
		}

		return activities, nil
	}

	result, err := s.executeWithTokenRefresh(user, apiCall)
	if err != nil {
		return nil, err
	}

	return result.([]*StravaActivity), nil
}

func (s *stravaService) GetActivityDetail(user *models.User, activityID int64) (*StravaActivityDetail, error) {
	apiCall := func(accessToken string) (interface{}, error) {
		endpoint := fmt.Sprintf("/activities/%d", activityID)

		body, err := s.makeRequest("GET", endpoint, accessToken, nil)

		if err != nil {
			return nil, fmt.Errorf("failed to get activity detail: %w", err)
		}

		var activity StravaActivityDetail
		if err := json.Unmarshal(body, &activity); err != nil {
			return nil, fmt.Errorf("failed to parse activity detail: %w", err)
		}

		return &activity, nil
	}

	result, err := s.executeWithTokenRefresh(user, apiCall)
	if err != nil {
		return nil, err
	}

	return result.(*StravaActivityDetail), nil
}

func (s *stravaService) GetActivityStreams(user *models.User, activityID int64, streamTypes []string, resolution string) (*StravaStreams, error) {
	apiCall := func(accessToken string) (interface{}, error) {
		endpoint := fmt.Sprintf("/activities/%d/streams", activityID)

		params := url.Values{}
		params.Set("keys", strings.Join(streamTypes, ","))
		if resolution != "" {
			params.Set("resolution", resolution)
		}
		params.Set("key_by_type", "true")

		body, err := s.makeRequest("GET", endpoint, accessToken, params)
		if err != nil {
			return nil, fmt.Errorf("failed to get activity streams: %w", err)
		}

		// First unmarshal into the raw response format
		var rawStreams StravaStreamsResponse
		if err := json.Unmarshal(body, &rawStreams); err != nil {
			return nil, fmt.Errorf("failed to parse activity streams response: %w", err)
		}

		// Convert to our structured format
		streams, err := parseStreamsResponse(rawStreams)
		if err != nil {
			return nil, fmt.Errorf("failed to convert activity streams: %w", err)
		}

		return streams, nil
	}

	result, err := s.executeWithTokenRefresh(user, apiCall)
	if err != nil {
		return nil, err
	}

	return result.(*StravaStreams), nil
}

func (s *stravaService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	tokenURL := "https://www.strava.com/oauth/token"

	data := url.Values{}
	data.Set("client_id", s.config.StravaClientID)
	data.Set("client_secret", s.config.StravaClientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token refresh response: %w", err)
	}

	return &tokenResp, nil
}

// parseStreamsResponse converts the raw Strava streams response to our structured format
func parseStreamsResponse(rawStreams StravaStreamsResponse) (*StravaStreams, error) {
	streams := &StravaStreams{}

	for streamType, streamData := range rawStreams {
		switch streamType {
		case "time":
			if data, ok := streamData.Data.([]interface{}); ok {
				streams.Time = make([]int, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.Time[i] = int(val)
					}
				}
			}
		case "distance":
			if data, ok := streamData.Data.([]interface{}); ok {
				streams.Distance = make([]float64, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.Distance[i] = val
					}
				}
			}
		case "latlng":
			if data, ok := streamData.Data.([]interface{}); ok {
				streams.Latlng = make([][]float64, len(data))
				for i, v := range data {
					if coords, ok := v.([]interface{}); ok && len(coords) == 2 {
						streams.Latlng[i] = make([]float64, 2)
						if lat, ok := coords[0].(float64); ok {
							streams.Latlng[i][0] = lat
						}
						if lng, ok := coords[1].(float64); ok {
							streams.Latlng[i][1] = lng
						}
					}
				}
			}
		case "altitude":
			if data, ok := streamData.Data.([]interface{}); ok {
				streams.Altitude = make([]float64, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.Altitude[i] = val
					}
				}
			}
		case "velocity_smooth":
			if data, ok := streamData.Data.([]interface{}); ok {
				streams.VelocitySmooth = make([]float64, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.VelocitySmooth[i] = val
					}
				}
			}
		case "heartrate":
			if data, ok := streamData.Data.([]interface{}); ok {
				streams.Heartrate = make([]int, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.Heartrate[i] = int(val)
					}
				}
			}
		case "cadence":
			if data, ok := streamData.Data.([]interface{}); ok {
				streams.Cadence = make([]int, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.Cadence[i] = int(val)
					}
				}
			}
		case "watts":
			if data, ok := streamData.Data.([]interface{}); ok {
				streams.Watts = make([]int, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.Watts[i] = int(val)
					}
				}
			}
		case "temp":
			if data, ok := streamData.Data.([]interface{}); ok {
				streams.Temp = make([]int, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.Temp[i] = int(val)
					}
				}
			}
		case "moving":
			if data, ok := streamData.Data.([]interface{}); ok {
				streams.Moving = make([]bool, len(data))
				for i, v := range data {
					if val, ok := v.(bool); ok {
						streams.Moving[i] = val
					}
				}
			}
		case "grade_smooth":
			if data, ok := streamData.Data.([]interface{}); ok {
				streams.GradeSmooth = make([]float64, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.GradeSmooth[i] = val
					}
				}
			}
		}
	}

	return streams, nil
}
