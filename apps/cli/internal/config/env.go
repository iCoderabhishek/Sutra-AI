package config

import (
	"os"
)

// DefaultBackendURL is the server the CLI talks to when SUTRA_BACKEND_URL is
// not set. Release builds replace it with the hosted backend via -ldflags -X.
var DefaultBackendURL = "http://localhost:4000"

type EnvConfig struct {
	BackendURL      string
	OAuthCallback   string
	OAuthRoute      string
	LocalServerAddr string
}

func GetEnvConfig() EnvConfig {
	return EnvConfig{
		BackendURL:      getEnvOrDefault("SUTRA_BACKEND_URL", DefaultBackendURL),
		OAuthCallback:   getEnvOrDefault("SUTRA_OAUTH_CALLBACK", "http://localhost:8080/callback"),
		OAuthRoute:      getEnvOrDefault("SUTRA_OAUTH_ROUTE", "/api/v1/auth/google"),
		LocalServerAddr: getEnvOrDefault("SUTRA_LOCAL_SERVER_ADDR", "localhost:8080"),
	}
}

func getEnvOrDefault(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}
