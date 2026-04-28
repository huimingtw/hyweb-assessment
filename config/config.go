package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Port          string
	DSN           string
	JWTSecret     string
	WeatherAPIKey string
}

func Load() (*Config, error) {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbHost := getEnv("DB_HOST", "localhost:3306")

	if dbUser == "" {
		return nil, errors.New("DB_USER environment variable is required")
	}
	if dbPass == "" {
		return nil, errors.New("DB_PASS environment variable is required")
	}
	if dbName == "" {
		return nil, errors.New("DB_NAME environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET environment variable is required")
	}

	return &Config{
		Port:          getEnv("PORT", "8080"),
		DSN:           fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&loc=UTC", dbUser, dbPass, dbHost, dbName),
		JWTSecret:     jwtSecret,
		WeatherAPIKey: os.Getenv("WEATHER_API_KEY"),
	}, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
