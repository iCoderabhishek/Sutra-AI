package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type SessionConfig struct {
	Cookie string `json:"cookie"`
	Token  string `json:"token,omitempty"`
}

func getSessionPath() (string, error) {
	configRoot, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not find user config dir: %w", err)
	}
	appConfigDir := filepath.Join(configRoot, "sutraai")

	// Ensure directory exists
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(appConfigDir, "session.json"), nil
}

func SaveSession(tokenValue string) error {
	sessionPath, err := getSessionPath()
	if err != nil {
		return err
	}

	session := SessionConfig{
		Token: tokenValue,
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(sessionPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

func LoadSession() (string, error) {
	sessionPath, err := getSessionPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read session file: %w", err)
	}

	var session SessionConfig
	if err := json.Unmarshal(data, &session); err != nil {
		return "", fmt.Errorf("failed to parse session file: %w", err)
	}

	if session.Token != "" {
		return session.Token, nil
	}
	return session.Cookie, nil
}

// ClearSession removes the saved session file. A missing file is not an error.
func ClearSession() error {
	sessionPath, err := getSessionPath()
	if err != nil {
		return err
	}
	if err := os.Remove(sessionPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove session file: %w", err)
	}
	return nil
}
