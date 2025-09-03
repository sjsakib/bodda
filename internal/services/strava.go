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
	ResourceState       int                     `json:"resource_state"`
	Description         string                  `json:"description"`
	Calories            float64                 `json:"calories"`
	SegmentEfforts      []StravaSegmentEffort   `json:"segment_efforts"`
	Splits              []StravaSplit           `json:"splits_metric"`
	SplitsStandard      []StravaSplitStandard   `json:"splits_standard"`
	BestEfforts         []StravaBestEffort      `json:"best_efforts"`
	Laps                []StravaLap             `json:"laps"`
	Gear                StravaGear              `json:"gear"`
	Photos              StravaPhotos            `json:"photos"`
	HighlightedKudosers []StravaAthlete         `json:"highlighted_kudosers"`
	SimilarActivities   StravaSimilarActivities `json:"similar_activities"`
	AvailableZones      []string                `json:"available_zones"`

	// Enhanced athlete information
	Athlete StravaAthleteRef `json:"athlete"`

	// Location data
	StartLatlng     []float64 `json:"start_latlng"`
	EndLatlng       []float64 `json:"end_latlng"`
	LocationCity    string    `json:"location_city"`
	LocationState   string    `json:"location_state"`
	LocationCountry string    `json:"location_country"`

	// Achievement and social metrics
	AchievementCount int  `json:"achievement_count"`
	TotalPhotoCount  int  `json:"total_photo_count"`
	HasKudoed        bool `json:"has_kudoed"`

	// Enhanced cadence metrics
	AverageCadence float64 `json:"average_cadence"`

	// Enhanced power metrics
	WeightedAverageWatts float64 `json:"weighted_average_watts"`

	// Enhanced temperature metrics
	AverageTemp float64 `json:"average_temp"`

	// Strava-specific metrics
	SufferScore             float64 `json:"suffer_score"`
	PerceivedExertion       int     `json:"perceived_exertion"`
	PreferPerceivedExertion bool    `json:"prefer_perceived_exertion"`

	// Privacy and display options
	HeartrateOptOut            bool `json:"heartrate_opt_out"`
	DisplayHideHeartrateOption bool `json:"display_hide_heartrate_option"`
	HideFromHome               bool `json:"hide_from_home"`

	// Upload and external tracking
	UploadID        int64  `json:"upload_id"`
	UploadIDStr     string `json:"upload_id_str"`
	ExternalID      string `json:"external_id"`
	FromAcceptedTag bool   `json:"from_accepted_tag"`

	// Device information
	DeviceName string `json:"device_name"`
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
	ID                 int64             `json:"id"`
	ResourceState      int               `json:"resource_state"`
	Name               string            `json:"name"`
	Activity           StravaActivityRef `json:"activity"`
	Athlete            StravaAthleteRef  `json:"athlete"`
	ElapsedTime        int               `json:"elapsed_time"`
	MovingTime         int               `json:"moving_time"`
	StartDate          string            `json:"start_date"`
	StartDateLocal     string            `json:"start_date_local"`
	Distance           float64           `json:"distance"`
	StartIndex         int               `json:"start_index"`
	EndIndex           int               `json:"end_index"`
	TotalElevationGain float64           `json:"total_elevation_gain"`
	AverageSpeed       float64           `json:"average_speed"`
	MaxSpeed           float64           `json:"max_speed"`
	AverageHeartrate   float64           `json:"average_heartrate"`
	MaxHeartrate       float64           `json:"max_heartrate"`
	AveragePower       float64           `json:"average_power"`
	MaxPower           float64           `json:"max_power"`
	AverageCadence     float64           `json:"average_cadence"`
	DeviceWatts        bool              `json:"device_watts"`
	AverageWatts       float64           `json:"average_watts"`
	LapIndex           int               `json:"lap_index"`
	Split              int               `json:"split"`
	PaceZone           int               `json:"pace_zone"`
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

// StravaSplitStandard represents standard splits (typically mile or kilometer splits)
type StravaSplitStandard struct {
	Distance                  float64 `json:"distance"`
	ElapsedTime               int     `json:"elapsed_time"`
	ElevationDifference       float64 `json:"elevation_difference"`
	MovingTime                int     `json:"moving_time"`
	Split                     int     `json:"split"`
	AverageSpeed              float64 `json:"average_speed"`
	AverageGradeAdjustedSpeed float64 `json:"average_grade_adjusted_speed"`
	AverageHeartrate          float64 `json:"average_heartrate"`
	AveragePower              float64 `json:"average_power"`
	AverageCadence            float64 `json:"average_cadence"`
	PaceZone                  int     `json:"pace_zone"`
}

// StravaBestEffort represents best effort segments within an activity
type StravaBestEffort struct {
	ID               int64               `json:"id"`
	ResourceState    int                 `json:"resource_state"`
	Name             string              `json:"name"`
	Activity         StravaActivityRef   `json:"activity"`
	Athlete          StravaAthleteRef    `json:"athlete"`
	ElapsedTime      int                 `json:"elapsed_time"`
	MovingTime       int                 `json:"moving_time"`
	StartDate        string              `json:"start_date"`
	StartDateLocal   string              `json:"start_date_local"`
	Distance         float64             `json:"distance"`
	StartIndex       int                 `json:"start_index"`
	EndIndex         int                 `json:"end_index"`
	AverageHeartrate float64             `json:"average_heartrate"`
	MaxHeartrate     float64             `json:"max_heartrate"`
	AveragePower     float64             `json:"average_power"`
	MaxPower         float64             `json:"max_power"`
	AverageCadence   float64             `json:"average_cadence"`
	PRRank           int                 `json:"pr_rank"`
	Achievements     []StravaAchievement `json:"achievements"`
}

// StravaAchievement represents achievements earned during best efforts
type StravaAchievement struct {
	TypeID int    `json:"type_id"`
	Type   string `json:"type"`
	Rank   int    `json:"rank"`
}

// StravaSimilarActivities represents similar activities for comparison
type StravaSimilarActivities struct {
	EffortCount        int              `json:"effort_count"`
	AverageSpeed       float64          `json:"average_speed"`
	MinAverageSpeed    float64          `json:"min_average_speed"`
	MidAverageSpeed    float64          `json:"mid_average_speed"`
	MaxAverageSpeed    float64          `json:"max_average_speed"`
	PRRank             int              `json:"pr_rank"`
	FrequencyMilestone string           `json:"frequency_milestone"`
	TrendStats         StravaTrendStats `json:"trend"`
	ResourceState      int              `json:"resource_state"`
}

// StravaTrendStats represents trend statistics for similar activities
type StravaTrendStats struct {
	Speeds               []float64 `json:"speeds"`
	CurrentActivityIndex int       `json:"current_activity_index"`
	MinSpeed             float64   `json:"min_speed"`
	MidSpeed             float64   `json:"mid_speed"`
	MaxSpeed             float64   `json:"max_speed"`
	Direction            int       `json:"direction"`
}

// StravaAthleteRef represents a reference to an athlete in activity details
type StravaAthleteRef struct {
	ID            int64 `json:"id"`
	ResourceState int   `json:"resource_state"`
}

// StravaActivityRef represents a reference to an activity
type StravaActivityRef struct {
	ID            int64 `json:"id"`
	ResourceState int   `json:"resource_state"`
}

// StravaAthleteWithZones represents an athlete profile integrated with training zones
type StravaAthleteWithZones struct {
	*StravaAthlete
	Zones *StravaAthleteZones `json:"zones,omitempty"`
}

// StravaAthleteZones represents an athlete's configured training zones
type StravaAthleteZones struct {
	HeartRate *StravaZoneSet `json:"heart_rate,omitempty"`
	Power     *StravaZoneSet `json:"power,omitempty"`
	Pace      *StravaZoneSet `json:"pace,omitempty"`
}

// StravaZoneSet represents a set of training zones for a specific metric
type StravaZoneSet struct {
	CustomZones   bool         `json:"custom_zones"`
	Zones         []StravaZone `json:"zones"`
	ResourceState int          `json:"resource_state"`
}

// StravaZone represents a single training zone with min/max boundaries
type StravaZone struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// StravaActivityZones represents training zone data for an activity
type StravaActivityZones struct {
	HeartRate *StravaZoneDistribution `json:"heart_rate,omitempty"`
	Power     *StravaZoneDistribution `json:"power,omitempty"`
	Pace      *StravaZoneDistribution `json:"pace,omitempty"`
}

// StravaActivityDetailWithZones represents an activity detail integrated with zone distribution data
type StravaActivityDetailWithZones struct {
	*StravaActivityDetail
	Zones *StravaActivityZones `json:"zones,omitempty"`
}

// StravaZoneDistribution represents time spent in each training zone
type StravaZoneDistribution struct {
	CustomZones   bool             `json:"custom_zones"`
	Zones         []StravaZoneData `json:"distribution_buckets"`
	Type          string           `json:"type"`
	ResourceState int              `json:"resource_state"`
	SensorBased   bool             `json:"sensor_based"`
}

// StravaZoneData represents data for a specific training zone
type StravaZoneData struct {
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
	Time float64 `json:"time"`
}

// StravaStreamData represents a single stream's data structure
type StravaStreamData struct {
	Data       any    `json:"data"`
	SeriesType string `json:"series_type"`
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
	GetAthleteProfile(user *models.User) (*StravaAthleteWithZones, error)
	GetAthleteZones(user *models.User) (*StravaAthleteZones, error)
	GetActivities(user *models.User, params ActivityParams) ([]*StravaActivity, error)
	GetActivityDetail(user *models.User, activityID int64) (*StravaActivityDetail, error)
	GetActivityDetailWithZones(user *models.User, activityID int64) (*StravaActivityDetailWithZones, error)
	GetActivityStreams(user *models.User, activityID int64, streamTypes []string, resolution string) (*StravaStreams, error)
	GetActivityZones(user *models.User, activityID int64) (*StravaActivityZones, error)
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
func (s *stravaService) executeWithTokenRefresh(user *models.User, apiCall func(string) (any, error)) (any, error) {
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

func (s *stravaService) GetAthleteProfile(user *models.User) (*StravaAthleteWithZones, error) {
	// First get the basic athlete profile
	apiCall := func(accessToken string) (any, error) {
		body, err := s.makeRequest("GET", "/athlete", accessToken, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get athlete profile: %w", err)
		}

		var athlete StravaAthlete
		if err := json.Unmarshal(body, &athlete); err != nil {
			return nil, fmt.Errorf("failed to parse athlete profile: %w", err)
		}

		return &athlete, nil
	}

	result, err := s.executeWithTokenRefresh(user, apiCall)
	if err != nil {
		return nil, err
	}

	athlete := result.(*StravaAthlete)

	// Create the integrated profile with zones
	profileWithZones := &StravaAthleteWithZones{
		StravaAthlete: athlete,
	}

	// Attempt to fetch zone data - this is optional and may not be available
	zones, err := s.GetAthleteZones(user)
	if err != nil {
		// Log the error but don't fail the entire request
		// Zones may not be configured or accessible
		log.Printf("Failed to fetch athlete zones for user %s: %v", user.ID, err)
	} else {
		profileWithZones.Zones = zones
	}

	return profileWithZones, nil
}

func (s *stravaService) GetAthleteZones(user *models.User) (*StravaAthleteZones, error) {
	apiCall := func(accessToken string) (any, error) {
		body, err := s.makeRequest("GET", "/athlete/zones", accessToken, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get athlete zones: %w", err)
		}

		// Strava returns an array of zone sets
		var zones StravaAthleteZones
		if err := json.Unmarshal(body, &zones); err != nil {
			return nil, fmt.Errorf("failed to parse athlete zones: %w", err)
		}

		return &zones, nil
	}

	result, err := s.executeWithTokenRefresh(user, apiCall)
	if err != nil {
		return nil, err
	}

	return result.(*StravaAthleteZones), nil
}

func (s *stravaService) GetActivities(user *models.User, params ActivityParams) ([]*StravaActivity, error) {
	apiCall := func(accessToken string) (any, error) {
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
	apiCall := func(accessToken string) (any, error) {
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

func (s *stravaService) GetActivityDetailWithZones(user *models.User, activityID int64) (*StravaActivityDetailWithZones, error) {
	// First get the basic activity detail
	activityDetail, err := s.GetActivityDetail(user, activityID)
	if err != nil {
		return nil, err
	}

	// Create the integrated activity detail with zones
	activityWithZones := &StravaActivityDetailWithZones{
		StravaActivityDetail: activityDetail,
	}

	// Attempt to fetch zone data if available - this is optional and may not be available
	if len(activityDetail.AvailableZones) > 0 {
		zones, err := s.GetActivityZones(user, activityID)
		if err != nil {
			// Log the error but don't fail the entire request
			// Zone data may not be available for all activities
			log.Printf("Failed to fetch activity zones for activity %d: %v", activityID, err)
		} else {
			activityWithZones.Zones = zones
		}
	}

	return activityWithZones, nil
}

func (s *stravaService) GetActivityStreams(user *models.User, activityID int64, streamTypes []string, resolution string) (*StravaStreams, error) {
	apiCall := func(accessToken string) (any, error) {
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

func (s *stravaService) GetActivityZones(user *models.User, activityID int64) (*StravaActivityZones, error) {
	apiCall := func(accessToken string) (any, error) {
		endpoint := fmt.Sprintf("/activities/%d/zones", activityID)

		body, err := s.makeRequest("GET", endpoint, accessToken, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get activity zones: %w", err)
		}

		// Strava returns an array of zone distributions
		var zoneDistributions []StravaZoneDistribution
		if err := json.Unmarshal(body, &zoneDistributions); err != nil {
			return nil, fmt.Errorf("failed to parse activity zones: %w", err)
		}

		// Convert to our structured format
		zones := &StravaActivityZones{}
		for _, dist := range zoneDistributions {
			switch strings.ToLower(dist.Type) {
			case "heartrate", "heart_rate":
				zones.HeartRate = &dist
			case "power":
				zones.Power = &dist
			case "pace":
				zones.Pace = &dist
			}
		}

		return zones, nil
	}

	result, err := s.executeWithTokenRefresh(user, apiCall)
	if err != nil {
		return nil, err
	}

	return result.(*StravaActivityZones), nil
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
			if data, ok := streamData.Data.([]any); ok {
				streams.Time = make([]int, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.Time[i] = int(val)
					}
				}
			}
		case "distance":
			if data, ok := streamData.Data.([]any); ok {
				streams.Distance = make([]float64, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.Distance[i] = val
					}
				}
			}
		case "latlng":
			if data, ok := streamData.Data.([]any); ok {
				streams.Latlng = make([][]float64, len(data))
				for i, v := range data {
					if coords, ok := v.([]any); ok && len(coords) == 2 {
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
			if data, ok := streamData.Data.([]any); ok {
				streams.Altitude = make([]float64, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.Altitude[i] = val
					}
				}
			}
		case "velocity_smooth":
			if data, ok := streamData.Data.([]any); ok {
				streams.VelocitySmooth = make([]float64, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.VelocitySmooth[i] = val
					}
				}
			}
		case "heartrate":
			if data, ok := streamData.Data.([]any); ok {
				streams.Heartrate = make([]int, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.Heartrate[i] = int(val)
					}
				}
			}
		case "cadence":
			if data, ok := streamData.Data.([]any); ok {
				streams.Cadence = make([]int, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.Cadence[i] = int(val)
					}
				}
			}
		case "watts":
			if data, ok := streamData.Data.([]any); ok {
				streams.Watts = make([]int, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.Watts[i] = int(val)
					}
				}
			}
		case "temp":
			if data, ok := streamData.Data.([]any); ok {
				streams.Temp = make([]int, len(data))
				for i, v := range data {
					if val, ok := v.(float64); ok {
						streams.Temp[i] = int(val)
					}
				}
			}
		case "moving":
			if data, ok := streamData.Data.([]any); ok {
				streams.Moving = make([]bool, len(data))
				for i, v := range data {
					if val, ok := v.(bool); ok {
						streams.Moving[i] = val
					}
				}
			}
		case "grade_smooth":
			if data, ok := streamData.Data.([]any); ok {
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
