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
			AppEnv:    getEnvString("APP_ENV"),
			SMTPHost:  getEnvString("SMTP_HOST"),
			SMTPPort:  getEnvInt("SMTP_PORT"),
			SMTPUser:  getEnvString("SMTP_USERNAME"),
			SMTPPass:  getEnvString("SMTP_PASSWORD"),
			EmailFrom: getEnvString("EMAIL_FROM"),
		}
	})
	return envCfg
}

func getEnvString(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		panic(fmt.Sprintf("missing required env var: %s", key))
	}
	return v
}

func getEnvInt(key string) int {
	v := getEnvString(key)
	i, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Sprintf("invalid integer env var %s: %q", key, v))
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
		panic(fmt.Sprintf("invalid log level env var %s: %q", key, v))
	}
}
