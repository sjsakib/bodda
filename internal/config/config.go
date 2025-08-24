package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	
	// Strava OAuth
	StravaClientID     string
	StravaClientSecret string
	StravaRedirectURL  string
	
	// OpenAI
	OpenAIAPIKey string
	
	// Frontend URL for CORS
	FrontendURL string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/bodda?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		
		StravaClientID:     getEnv("STRAVA_CLIENT_ID", ""),
		StravaClientSecret: getEnv("STRAVA_CLIENT_SECRET", ""),
		StravaRedirectURL:  getEnv("STRAVA_REDIRECT_URL", "http://localhost:8080/auth/callback"),
		
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}