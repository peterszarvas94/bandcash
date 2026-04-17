package utils

import (
	"errors"
	"log/slog"
	"strings"
	"sync"

	"github.com/caarlos0/env/v11"
)

type EnvConfig struct {
	Host               string
	Port               int
	LogLevel           slog.Level
	LogFolder          string
	LogPrefix          string
	DBPath             string
	URL                string
	AppEnv             string
	SuperadminEmail    string
	DisableRateLimit   bool
	EmailProvider      string
	ResendAPIKey       string
	MailtrapHost       string
	MailtrapPort       int
	MailtrapUsername   string
	MailtrapPassword   string
	EmailFrom          string
	LemonWebhookSecret string
	LemonHostedURL     string
}

const DefaultSuperadminEmail = "admin@bandcash.localhost"

var (
	envOnce sync.Once
	envCfg  *EnvConfig
)

type envVars struct {
	// AppEnv toggles environment-specific behavior and required fields.
	AppEnv string `env:"APP_ENV" envDefault:"development" validate:"required,oneof=development staging production"`
	// Host is the interface the HTTP server binds to.
	Host string `env:"HOST" envDefault:"0.0.0.0"`
	// Port is the HTTP port the app listens on.
	Port int `env:"PORT" envDefault:"2222" validate:"required,gte=1,lte=65535"`
	// LogLevel controls structured log verbosity.
	LogLevel string `env:"LOG_LEVEL" envDefault:"debug" validate:"required,oneof=debug info warn error"`
	// LogFolder is the directory where log files are stored.
	LogFolder string `env:"LOG_FOLDER" envDefault:"logs" validate:"required"`
	// LogPrefix is the base file name prefix for logs.
	LogPrefix string `env:"LOG_PREFIX" envDefault:"bandcash" validate:"required"`
	// DBPath is the SQLite database file path.
	DBPath string `env:"DB_PATH" envDefault:"sqlite.db" validate:"required"`
	// URL is the public base URL used for links and callbacks.
	URL string `env:"URL" envDefault:"http://localhost:2222" validate:"required_if=AppEnv production,required_if=AppEnv staging"`
	// SuperadminEmail is the bootstrap superadmin account email.
	SuperadminEmail string `env:"SUPERADMIN_EMAIL" envDefault:"admin@bandcash.localhost" validate:"required,email"`
	// DisableRateLimit disables request rate limiting (useful for local dev).
	DisableRateLimit bool `env:"DISABLE_RATE_LIMIT" envDefault:"true"`
	// EmailProvider controls which transport is used to deliver emails.
	EmailProvider string `env:"EMAIL_PROVIDER" validate:"required,oneof=resend mailtrap"`
	// ResendAPIKey stores the Resend API key.
	ResendAPIKey string `env:"RESEND_API_KEY" validate:"required_if=EmailProvider resend"`
	// MailtrapHost is the SMTP host for Mailtrap Sandbox.
	MailtrapHost string `env:"MAILTRAP_HOST" validate:"required_if=EmailProvider mailtrap"`
	// MailtrapPort is the SMTP port for Mailtrap Sandbox.
	MailtrapPort int `env:"MAILTRAP_PORT" envDefault:"2525" validate:"required_if=EmailProvider mailtrap,gte=1,lte=65535"`
	// MailtrapUsername is the SMTP username for Mailtrap Sandbox.
	MailtrapUsername string `env:"MAILTRAP_USERNAME" validate:"required_if=EmailProvider mailtrap"`
	// MailtrapPassword is the SMTP password for Mailtrap Sandbox.
	MailtrapPassword string `env:"MAILTRAP_PASSWORD" validate:"required_if=EmailProvider mailtrap"`
	// EmailFrom is the default From header for app emails.
	EmailFrom string `env:"EMAIL_FROM" envDefault:"BandCash <noreply@bandcash.localhost>" validate:"required_if=AppEnv production,required_if=AppEnv staging"`
	// LemonWebhookSecret is the endpoint secret used to verify webhook signatures.
	LemonWebhookSecret string `env:"LEMON_WEBHOOK_SECRET"`
	// LemonHostedURL is the hosted customer billing URL where users manage subscriptions.
	LemonHostedURL string `env:"LEMON_HOSTED_URL"`
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
		parsed.EmailProvider = strings.ToLower(strings.TrimSpace(parsed.EmailProvider))

		err = validate.Struct(parsed)
		if err != nil {
			panic("invalid env vars: " + err.Error())
		}

		superadminEmail := strings.ToLower(strings.TrimSpace(parsed.SuperadminEmail))
		if (parsed.AppEnv == "production" || parsed.AppEnv == "staging") && superadminEmail == DefaultSuperadminEmail {
			panic("invalid env vars: SUPERADMIN_EMAIL must be overridden in staging/production")
		}

		logLevel, err := parseLogLevel(parsed.LogLevel)
		if err != nil {
			panic("invalid env vars: " + err.Error())
		}

		envCfg = &EnvConfig{
			Host:               parsed.Host,
			Port:               parsed.Port,
			LogLevel:           logLevel,
			LogFolder:          parsed.LogFolder,
			LogPrefix:          parsed.LogPrefix,
			DBPath:             parsed.DBPath,
			URL:                parsed.URL,
			AppEnv:             parsed.AppEnv,
			SuperadminEmail:    superadminEmail,
			DisableRateLimit:   parsed.DisableRateLimit,
			EmailProvider:      parsed.EmailProvider,
			ResendAPIKey:       parsed.ResendAPIKey,
			MailtrapHost:       strings.TrimSpace(parsed.MailtrapHost),
			MailtrapPort:       parsed.MailtrapPort,
			MailtrapUsername:   strings.TrimSpace(parsed.MailtrapUsername),
			MailtrapPassword:   strings.TrimSpace(parsed.MailtrapPassword),
			EmailFrom:          parsed.EmailFrom,
			LemonWebhookSecret: strings.TrimSpace(parsed.LemonWebhookSecret),
			LemonHostedURL:     strings.TrimSpace(parsed.LemonHostedURL),
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
