package config

import (
	"errors"
	"os"
)

type Config struct {
	Port          string
	DSN           string
	JWTSecret     string
	WeatherAPIKey string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:          getEnv("PORT", "8080"),
		DSN:           os.Getenv("DSN"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		WeatherAPIKey: os.Getenv("WEATHER_API_KEY"),
	}

	if cfg.DSN == "" {
		return nil, errors.New("DSN environment variable is required")
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET environment variable is required")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
