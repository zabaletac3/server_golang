package config

import (
	"os"
	"strconv"
)

type Config struct {
	Env string `env:"APP_ENV" default:"development"`
	// Port string `env:"PORT" default:"8080"`
	Port string `env:"PORT"`
	ShutdownSecs int

	ReadHeaderTimeoutSecs int
	ReadTimeoutSecs       int
	WriteTimeoutSecs      int
	IdleTimeoutSecs       int
	MaxHeaderBytes        int
}

func Load() *Config {
	cfg := &Config{
		Env: getEnv("APP_ENV", "development"),
		Port: getEnv("PORT", ""),
		ShutdownSecs: getEnvInt("SHUTDOWN_SECS", 10),

		ReadHeaderTimeoutSecs: getEnvInt("READ_HEADER_TIMEOUT_SECS", 5),
		ReadTimeoutSecs:       getEnvInt("READ_TIMEOUT_SECS", 15),
		WriteTimeoutSecs:      getEnvInt("WRITE_TIMEOUT_SECS", 15),
		IdleTimeoutSecs:       getEnvInt("IDLE_TIMEOUT_SECS", 60),
		MaxHeaderBytes:        getEnvInt("MAX_HEADER_BYTES", 1<<20),
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