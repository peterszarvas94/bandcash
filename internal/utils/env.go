package utils

import (
	"errors"
	"log/slog"
	"strings"
	"sync"

	"github.com/caarlos0/env/v11"
)

type EnvConfig struct {
	Host                string
	Port                int
	LogLevel            slog.Level
	LogFolder           string
	LogPrefix           string
	DBPath              string
	URL                 string
	AppEnv              string
	SuperadminEmail     string
	DisableRateLimit    bool
	SMTPHost            string
	SMTPPort            int
	SMTPUser            string
	SMTPPass            string
	EmailFrom           string
	PaddleEnv           string
	PaddleAPIKey        string
	PaddleAPIBaseURL    string
	PaddleClientToken   string
	PaddleWebhookSecret string
	PaddlePriceID       string
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
	// SMTPHost is the SMTP server host for outgoing email.
	SMTPHost string `env:"SMTP_HOST" envDefault:"localhost" validate:"required_if=AppEnv production,required_if=AppEnv staging"`
	// SMTPPort is the SMTP server port for outgoing email.
	SMTPPort int `env:"SMTP_PORT" envDefault:"1025" validate:"required_if=AppEnv production,required_if=AppEnv staging,omitempty,gt=0"`
	// SMTPUser is the SMTP username used for authentication.
	SMTPUser string `env:"SMTP_USERNAME" validate:"required_if=AppEnv production,required_if=AppEnv staging"`
	// SMTPPass is the SMTP password used for authentication.
	SMTPPass string `env:"SMTP_PASSWORD" validate:"required_if=AppEnv production,required_if=AppEnv staging"`
	// EmailFrom is the default From header for app emails.
	EmailFrom string `env:"EMAIL_FROM" envDefault:"BandCash <noreply@bandcash.localhost>" validate:"required_if=AppEnv production,required_if=AppEnv staging"`
	// PaddleEnv defines the Paddle environment used by client SDK and API calls.
	PaddleEnv string `env:"PADDLE_ENV" envDefault:"" validate:"omitempty,oneof=sandbox production"`
	// PaddleAPIKey is the Paddle server API key.
	PaddleAPIKey string `env:"PADDLE_API_KEY"`
	// PaddleAPIBaseURL is the Paddle API base URL (for example https://sandbox-api.paddle.com).
	PaddleAPIBaseURL string `env:"PADDLE_API_BASE_URL"`
	// PaddleClientToken is the client-side token used by Paddle.js.
	PaddleClientToken string `env:"PADDLE_CLIENT_TOKEN"`
	// PaddleWebhookSecret is the endpoint secret used to verify webhook signatures.
	PaddleWebhookSecret string `env:"PADDLE_WEBHOOK_SECRET"`
	// PaddlePriceID is the single monthly subscription slot price id.
	PaddlePriceID string `env:"PADDLE_PRICE_ID"`
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
		if (parsed.AppEnv == "production" || parsed.AppEnv == "staging") && superadminEmail == DefaultSuperadminEmail {
			panic("invalid env vars: SUPERADMIN_EMAIL must be overridden in staging/production")
		}

		logLevel, err := parseLogLevel(parsed.LogLevel)
		if err != nil {
			panic("invalid env vars: " + err.Error())
		}

		envCfg = &EnvConfig{
			Host:                parsed.Host,
			Port:                parsed.Port,
			LogLevel:            logLevel,
			LogFolder:           parsed.LogFolder,
			LogPrefix:           parsed.LogPrefix,
			DBPath:              parsed.DBPath,
			URL:                 parsed.URL,
			AppEnv:              parsed.AppEnv,
			SuperadminEmail:     superadminEmail,
			DisableRateLimit:    parsed.DisableRateLimit,
			SMTPHost:            parsed.SMTPHost,
			SMTPPort:            parsed.SMTPPort,
			SMTPUser:            parsed.SMTPUser,
			SMTPPass:            parsed.SMTPPass,
			EmailFrom:           parsed.EmailFrom,
			PaddleEnv:           strings.ToLower(strings.TrimSpace(parsed.PaddleEnv)),
			PaddleAPIKey:        strings.TrimSpace(parsed.PaddleAPIKey),
			PaddleAPIBaseURL:    strings.TrimSpace(parsed.PaddleAPIBaseURL),
			PaddleClientToken:   strings.TrimSpace(parsed.PaddleClientToken),
			PaddleWebhookSecret: strings.TrimSpace(parsed.PaddleWebhookSecret),
			PaddlePriceID:       strings.TrimSpace(parsed.PaddlePriceID),
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
