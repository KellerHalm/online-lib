package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	DBHost        string
	DBPort        int
	DBUser        string
	DBPassword    string
	DBName        string
	JWTSecret     string
	AllowedOrigin string
}

func Load() Config {
	return Config{
		Port:          getenv("PORT", "3000"),
		DBHost:        getenv("DB_HOST", "localhost"),
		DBPort:        getenvInt("DB_PORT", 5432),
		DBUser:        getenv("DB_USER", "lib"),
		DBPassword:    getenv("DB_PASSWORD", "lib_password"),
		DBName:        getenv("DB_NAME", "library"),
		JWTSecret:     getenv("JWT_SECRET", ""),
		AllowedOrigin: getenv("CORS_ALLOWED_ORIGIN", "http://localhost:5173"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
