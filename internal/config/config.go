package config

import (
	"os"
	"strconv"
)

type Config struct {
	Env string `env:"APP_ENV" default:"development"`
	Port string `env:"PORT" default:"8080"`
	ShutdownSecs int
}

func Load() *Config {
	cfg := &Config{
		Env: getEnv("ENV", "development"),
		Port: getEnv("PORT", "8080"),
		ShutdownSecs: getEnvInt("SHUTDOWN_SECS", 10),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}