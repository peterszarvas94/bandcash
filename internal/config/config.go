package config

import (
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	Port      int
	LogLevel  slog.Level
	LogFolder string
	LogPrefix string
	DBPath    string
	URL       string
}

func Load() *Config {
	return &Config{
		Port:      getEnvInt("PORT", 8080),
		LogLevel:  getEnvLogLevel("LOG_LEVEL", slog.LevelDebug),
		LogFolder: getEnv("LOG_FOLDER", "logs"),
		LogPrefix: getEnv("LOG_PREFIX", "bandcash"),
		DBPath:    getEnv("DB_PATH", "sqlite.db"),
		URL:       getEnv("URL", "http://localhost:8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvLogLevel(key string, fallback slog.Level) slog.Level {
	v := os.Getenv(key)
	switch v {
	case "debug", "DEBUG":
		return slog.LevelDebug
	case "info", "INFO":
		return slog.LevelInfo
	case "warn", "WARN":
		return slog.LevelWarn
	case "error", "ERROR":
		return slog.LevelError
	default:
		return fallback
	}
}
