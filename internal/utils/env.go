package utils

import (
	"errors"
	"log/slog"
	"strings"
	"sync"

	"github.com/caarlos0/env/v11"
)

type EnvConfig struct {
	Host             string
	Port             int
	DevGlobalDelayMS int
	LogLevel         slog.Level
	LogFolder        string
	LogPrefix        string
	DBPath           string
	URL              string
	AppEnv           string
	SuperadminEmail  string
	DisableSignup    bool
	DisableRateLimit bool
	SMTPHost         string
	SMTPPort         int
	SMTPUser         string
	SMTPPass         string
	EmailFrom        string
	BetterStackURI   string
}

var (
	envOnce sync.Once
	envCfg  *EnvConfig
)

type envVars struct {
	AppEnv           string `env:"APP_ENV" validate:"required,oneof=development production"`
	Host             string `env:"HOST" envDefault:"0.0.0.0"`
	Port             int    `env:"PORT" validate:"required,gte=1,lte=65535"`
	DevGlobalDelayMS int    `env:"DEV_GLOBAL_DELAY_MS" envDefault:"0" validate:"gte=0"`
	LogLevel         string `env:"LOG_LEVEL" validate:"required,oneof=debug info warn error"`
	LogFolder        string `env:"LOG_FOLDER" validate:"required"`
	LogPrefix        string `env:"LOG_PREFIX" validate:"required"`
	DBPath           string `env:"DB_PATH" validate:"required"`
	URL              string `env:"URL" validate:"required_if=AppEnv production"`
	SuperadminEmail  string `env:"SUPERADMIN_EMAIL" validate:"omitempty,email"`
	DisableSignup    bool   `env:"DISABLE_SIGNUP" envDefault:"false"`
	DisableRateLimit bool   `env:"DISABLE_RATE_LIMIT" envDefault:"false"`
	SMTPHost         string `env:"SMTP_HOST" validate:"required_if=AppEnv production"`
	SMTPPort         int    `env:"SMTP_PORT" validate:"required_if=AppEnv production,omitempty,gt=0"`
	SMTPUser         string `env:"SMTP_USERNAME" validate:"required_if=AppEnv production"`
	SMTPPass         string `env:"SMTP_PASSWORD" validate:"required_if=AppEnv production"`
	EmailFrom        string `env:"EMAIL_FROM" validate:"required_if=AppEnv production"`
	BetterStackURI   string `env:"BETTER_STACK_URI"`
}

func Env() *EnvConfig {
	envOnce.Do(func() {
		var parsed envVars

		err := env.Parse(&parsed)
		if err != nil {
			panic("invalid env vars: " + err.Error())
		}

		parsed.AppEnv = strings.ToLower(strings.TrimSpace(parsed.AppEnv))
		parsed.LogLevel = strings.ToLower(strings.TrimSpace(parsed.LogLevel))

		err = validate.Struct(parsed)
		if err != nil {
			panic("invalid env vars: " + err.Error())
		}

		logLevel, err := parseLogLevel(parsed.LogLevel)
		if err != nil {
			panic("invalid env vars: " + err.Error())
		}

		envCfg = &EnvConfig{
			Host:             parsed.Host,
			Port:             parsed.Port,
			DevGlobalDelayMS: parsed.DevGlobalDelayMS,
			LogLevel:         logLevel,
			LogFolder:        parsed.LogFolder,
			LogPrefix:        parsed.LogPrefix,
			DBPath:           parsed.DBPath,
			URL:              parsed.URL,
			AppEnv:           parsed.AppEnv,
			SuperadminEmail:  strings.ToLower(strings.TrimSpace(parsed.SuperadminEmail)),
			DisableSignup:    parsed.DisableSignup,
			DisableRateLimit: parsed.DisableRateLimit,
			SMTPHost:         parsed.SMTPHost,
			SMTPPort:         parsed.SMTPPort,
			SMTPUser:         parsed.SMTPUser,
			SMTPPass:         parsed.SMTPPass,
			EmailFrom:        parsed.EmailFrom,
			BetterStackURI:   parsed.BetterStackURI,
		}
	})
	return envCfg
}

func parseLogLevel(v string) (slog.Level, error) {
	switch v {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, errors.New("LOG_LEVEL must be one of: debug, info, warn, error")
	}
}
