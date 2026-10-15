package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all server configuration from the environment.
type Config struct {
	Port            string
	DatabaseURL     string
	RedisAddr       string
	SessionTTLHrs   int
	RateLimitPerMin int
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Load reads and validates environment configuration. It fails fast when
// required values are missing or malformed.
func Load() (Config, error) {
	c := Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		RedisAddr:       os.Getenv("REDIS_ADDR"),
		SessionTTLHrs:   168,
		RateLimitPerMin: 10,
	}
	if c.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if c.RedisAddr == "" {
		return Config{}, fmt.Errorf("REDIS_ADDR is required")
	}
	if v := os.Getenv("SESSION_TTL_HOURS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("SESSION_TTL_HOURS must be a positive integer")
		}
		c.SessionTTLHrs = n
	}
	if v := os.Getenv("RATE_LIMIT_PER_MIN"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("RATE_LIMIT_PER_MIN must be a positive integer")
		}
		c.RateLimitPerMin = n
	}
	return c, nil
}
