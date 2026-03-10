package config

import (
	"os"
	"strings"
	"time"
)

// Config holds all server configuration.
type Config struct {
	Port        string
	SessionTTL  time.Duration
	CORSOrigins []string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	port := getEnv("PORT", "8080")

	ttlStr := getEnv("SESSION_TTL", "10m")
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		ttl = 10 * time.Minute
	}

	originsStr := getEnv("CORS_ORIGINS", "*")
	origins := strings.Split(originsStr, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return Config{
		Port:        port,
		SessionTTL:  ttl,
		CORSOrigins: origins,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
