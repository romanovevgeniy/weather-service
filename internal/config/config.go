package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort    string
	DefaultCity string
	JobInterval time.Duration
	DBHost      string
	DBPort      int
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
}

func Load() (Config, error) {
	cfg := Config{}

	cfg.HTTPPort = getEnv("HTTP_PORT", "3000")
	cfg.DefaultCity = getEnv("DEFAULT_CITY", "moscow")
	cfg.DBHost = getEnv("DB_HOST", "localhost")
	cfg.DBUser = getEnv("DB_USER", "admin")
	cfg.DBPassword = getEnv("DB_PASSWORD", "")
	cfg.DBName = getEnv("DB_NAME", "weather")
	cfg.DBSSLMode = getEnv("DB_SSL_MODE", "disable")

	dbPortStr := getEnv("DB_PORT", "5432")
	p, err := strconv.Atoi(dbPortStr)
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_PORT: %w", err)
	}
	cfg.DBPort = p

	interval := getEnv("JOB_INTERVAL", "10s")
	d, err := time.ParseDuration(interval)
	if err != nil {
		return Config{}, fmt.Errorf("invalid JOB_INTERVAL: %w", err)
	}
	cfg.JobInterval = d

	return cfg, nil
}

func (c Config) PGConnString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, urlEncode(c.DBPassword), c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func urlEncode(s string) string {
	r := make([]rune, 0, len(s))
	for _, ch := range s {
		if ch == '@' || ch == ':' || ch == '/' || ch == '?' || ch == '&' {
			r = append(r, '%')
			r = append(r, []rune(fmt.Sprintf("%X", ch))...)
		} else {
			r = append(r, ch)
		}
	}
	return string(r)
}
