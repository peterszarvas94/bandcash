package utils

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
)

type EnvConfig struct {
	Port      int
	LogLevel  slog.Level
	LogFolder string
	LogPrefix string
	DBPath    string
	URL       string
	AppEnv    string
	SMTPHost  string
	SMTPPort  int
	SMTPUser  string
	SMTPPass  string
	EmailFrom string
}

var (
	envOnce sync.Once
	envCfg  *EnvConfig
)

func Env() *EnvConfig {
	envOnce.Do(func() {
		envCfg = &EnvConfig{
			Port:      getEnvInt("PORT"),
			LogLevel:  getEnvLogLevel("LOG_LEVEL"),
			LogFolder: getEnvString("LOG_FOLDER"),
			LogPrefix: getEnvString("LOG_PREFIX"),
			DBPath:    getEnvString("DB_PATH"),
			URL:       getEnvString("URL"),
			AppEnv:    getAppEnv("APP_ENV"),
			SMTPHost:  getEnvString("SMTP_HOST"),
			SMTPPort:  getEnvInt("SMTP_PORT"),
			SMTPUser:  getDevOptionalEnvString("SMTP_USERNAME", "APP_ENV"),
			SMTPPass:  getDevOptionalEnvString("SMTP_PASSWORD", "APP_ENV"),
			EmailFrom: getEnvString("EMAIL_FROM"),
		}
	})
	return envCfg
}

func getDevOptionalEnvString(key string, appEnvKey string) string {
	appEnv := getAppEnv(appEnvKey)
	if appEnv == "development" {
		return strings.TrimSpace(os.Getenv(key))
	}
	return getEnvString(key)
}

func getEnvString(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		panic(fmt.Sprintf("Missing required %s env var: %s", key, key))
	}
	return v
}

func getEnvInt(key string) int {
	v := getEnvString(key)
	i, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Sprintf("Invalid %s env var %s: %q. Must be an integer.", key, key, v))
	}
	return i
}

func getEnvLogLevel(key string) slog.Level {
	v := getEnvString(key)
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
		panic(fmt.Sprintf("Invalid env var %s: %q. Allowed values: debug, info, warn, error", key, v))
	}
}

func getAppEnv(key string) string {
	e := getEnvString(key)
	switch e {
	case "development", "DEVELOPMENT":
		return "development"
	case "production", "PRODUCTION":
		return "production"
	default:
		panic(fmt.Sprintf("Invalid env var %s: %q. Allowed values: development, production", key, e))
	}

}
