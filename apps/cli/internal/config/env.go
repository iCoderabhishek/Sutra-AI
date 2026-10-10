package config

import (
	"os"
)

type EnvConfig struct {
	BackendURL      string
	OAuthCallback   string
	OAuthRoute      string
	LocalServerAddr string
}

func GetEnvConfig() EnvConfig {
	return EnvConfig{
		BackendURL:      getEnvOrDefault("SUTRA_BACKEND_URL", "http://localhost:4000"),
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
