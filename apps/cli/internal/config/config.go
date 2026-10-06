package config

type SessionConfig struct {
	Cookie string `json:"cookie"`
	Token  string `json:"token"`
}

func SaveSession(cookieValue string) error {}

func LoadSession() (string, error) {}
