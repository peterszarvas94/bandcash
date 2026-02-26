package utils

import (
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func LoadDotEnv(paths ...string) {
	existing := make([]string, 0, len(paths))

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existing = append(existing, path)
			continue
		} else if !os.IsNotExist(err) {
			slog.Warn("env.load: failed to stat dotenv file", "path", path, "err", err)
		}
	}

	if len(existing) == 0 {
		slog.Info("env.load: no dotenv files found", "paths", paths)
		return
	}

	if err := godotenv.Load(existing...); err != nil {
		slog.Warn("env.load: failed to load dotenv files", "paths", existing, "err", err)
		return
	}

	slog.Info("env.load: loaded dotenv files", "paths", existing)
}

func LoadAppDotEnv() {
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if appEnv == "production" {
		LoadDotEnv(".kamal/secrets")
		return
	}

	LoadDotEnv(".kamal/secrets.dev")
}
