package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bodda/internal/config"
	"bodda/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByStravaID(ctx context.Context, stravaID int64) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
}

type AuthService interface {
	HandleStravaOAuth(code string) (*models.User, error)
	ValidateToken(token string) (*models.User, error)
	RefreshStravaToken(user *models.User) error
	GenerateJWT(userID string) (string, error)
	GetStravaOAuthURL(state string) string
}

type authService struct {
	config      *config.Config
	userRepo    UserRepository
	oauthConfig *oauth2.Config
}

type StravaTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	Athlete      struct {
		ID        int64  `json:"id"`
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
	} `json:"athlete"`
}

func NewAuthService(cfg *config.Config, userRepo UserRepository) AuthService {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.StravaClientID,
		ClientSecret: cfg.StravaClientSecret,
		RedirectURL:  cfg.StravaRedirectURL,
		Scopes:       []string{"read,activity:read_all"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.strava.com/oauth/authorize",
			TokenURL: "https://www.strava.com/oauth/token",
		},
	}

	return &authService{
		config:      cfg,
		userRepo:    userRepo,
		oauthConfig: oauthConfig,
	}
}

func (s *authService) GetStravaOAuthURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *authService) HandleStravaOAuth(code string) (*models.User, error) {
	ctx := context.Background()
	
	// Exchange code for token
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get athlete info from Strava
	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.strava.com/api/v3/athlete")
	if err != nil {
		return nil, fmt.Errorf("failed to get athlete info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("strava API returned status %d", resp.StatusCode)
	}

	var athlete struct {
		ID        int64  `json:"id"`
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&athlete); err != nil {
		return nil, fmt.Errorf("failed to decode athlete response: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByStravaID(ctx, athlete.ID)
	if err == nil {
		// Update existing user's tokens
		existingUser.AccessToken = token.AccessToken
		existingUser.RefreshToken = token.RefreshToken
		existingUser.TokenExpiry = token.Expiry
		existingUser.FirstName = athlete.FirstName
		existingUser.LastName = athlete.LastName

		if err := s.userRepo.Update(ctx, existingUser); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		return existingUser, nil
	}

	// Create new user
	user := &models.User{
		StravaID:     athlete.ID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenExpiry:  token.Expiry,
		FirstName:    athlete.FirstName,
		LastName:     athlete.LastName,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *authService) GenerateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *authService) ValidateToken(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_id in token")
	}

	user, err := s.userRepo.GetByID(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

func (s *authService) RefreshStravaToken(user *models.User) error {
	ctx := context.Background()
	
	token := &oauth2.Token{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
		Expiry:       user.TokenExpiry,
	}

	tokenSource := s.oauthConfig.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update user with new token
	user.AccessToken = newToken.AccessToken
	user.RefreshToken = newToken.RefreshToken
	user.TokenExpiry = newToken.Expiry

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user tokens: %w", err)
	}

	return nil
}