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
	DisableRateLimit bool
	SMTPHost         string
	SMTPPort         int
	SMTPUser         string
	SMTPPass         string
	EmailFrom        string
}

var (
	envOnce sync.Once
	envCfg  *EnvConfig
)

type envVars struct {
	// AppEnv toggles environment-specific behavior and required fields.
	AppEnv string `env:"APP_ENV" envDefault:"development" validate:"required,oneof=development production"`
	// Host is the interface the HTTP server binds to.
	Host string `env:"HOST" envDefault:"0.0.0.0"`
	// Port is the HTTP port the app listens on.
	Port int `env:"PORT" envDefault:"2222" validate:"required,gte=1,lte=65535"`
	// DevGlobalDelayMS adds an artificial delay to responses in development.
	DevGlobalDelayMS int `env:"DEV_GLOBAL_DELAY_MS" envDefault:"0" validate:"gte=0"`
	// LogLevel controls structured log verbosity.
	LogLevel string `env:"LOG_LEVEL" envDefault:"debug" validate:"required,oneof=debug info warn error"`
	// LogFolder is the directory where log files are stored.
	LogFolder string `env:"LOG_FOLDER" envDefault:"logs" validate:"required"`
	// LogPrefix is the base file name prefix for logs.
	LogPrefix string `env:"LOG_PREFIX" envDefault:"bandcash" validate:"required"`
	// DBPath is the SQLite database file path.
	DBPath string `env:"DB_PATH" envDefault:"sqlite.db" validate:"required"`
	// URL is the public base URL used for links and callbacks.
	URL string `env:"URL" envDefault:"http://bandcash.localhost:9080" validate:"required_if=AppEnv production"`
	// SuperadminEmail is the bootstrap superadmin account email.
	SuperadminEmail string `env:"SUPERADMIN_EMAIL" validate:"required,email"`
	// DisableRateLimit disables request rate limiting (useful for local dev).
	DisableRateLimit bool `env:"DISABLE_RATE_LIMIT" envDefault:"true"`
	// SMTPHost is the SMTP server host for outgoing email.
	SMTPHost string `env:"SMTP_HOST" envDefault:"localhost" validate:"required_if=AppEnv production"`
	// SMTPPort is the SMTP server port for outgoing email.
	SMTPPort int `env:"SMTP_PORT" envDefault:"1025" validate:"required_if=AppEnv production,omitempty,gt=0"`
	// SMTPUser is the SMTP username used for authentication.
	SMTPUser string `env:"SMTP_USERNAME" validate:"required_if=AppEnv production"`
	// SMTPPass is the SMTP password used for authentication.
	SMTPPass string `env:"SMTP_PASSWORD" validate:"required_if=AppEnv production"`
	// EmailFrom is the default From header for app emails.
	EmailFrom string `env:"EMAIL_FROM" envDefault:"BandCash <noreply@bandcash.localhost>" validate:"required_if=AppEnv production"`
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

		superadminEmail := strings.ToLower(strings.TrimSpace(parsed.SuperadminEmail))

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
			SuperadminEmail:  superadminEmail,
			DisableRateLimit: parsed.DisableRateLimit,
			SMTPHost:         parsed.SMTPHost,
			SMTPPort:         parsed.SMTPPort,
			SMTPUser:         parsed.SMTPUser,
			SMTPPass:         parsed.SMTPPass,
			EmailFrom:        parsed.EmailFrom,
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
