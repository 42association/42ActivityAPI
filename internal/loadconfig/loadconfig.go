package loadconfig

import (
	"os"
	"errors"
)

type Config struct {
	UID         string
	Secret      string
	CallbackURL string
}

// Loading environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		UID:         os.Getenv("UID"),
		Secret:      os.Getenv("SECRET"),
		CallbackURL: os.Getenv("CALLBACK_URL"),
	}
	if config.UID == "" || config.Secret == "" || config.CallbackURL == "" {
		return nil, errors.New("one or more required environment variables are not set")
	}
	return config, nil
}