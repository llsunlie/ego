package config

import "os"

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	WebPort     string
	JWTExpHours string
	LogLevel    string
	LogFormat   string
	LogOutput   string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://ego:ego@localhost:5432/ego?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		Port:        getEnv("PORT", "9443"),
		WebPort:     getEnv("WEB_PORT", "9080"),
		JWTExpHours: getEnv("JWT_EXP_HOURS", "720"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		LogFormat:   getEnv("LOG_FORMAT", "text"),
		LogOutput:   getEnv("LOG_OUTPUT", "stdout"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
