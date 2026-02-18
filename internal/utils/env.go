package utils

import (
	"fmt"
	"os"
	"strings"
)

// ValidateRequiredEnv ensures required env vars are set.
func ValidateRequiredEnv(keys []string) error {
	missing := make([]string, 0, len(keys))
	for _, key := range keys {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	return nil
}

// ValidateEmailEnv checks SMTP configuration variables.
func ValidateEmailEnv() error {
	return ValidateRequiredEnv([]string{
		"SMTP_HOST",
		"SMTP_USERNAME",
		"SMTP_PASSWORD",
		"EMAIL_FROM",
	})
}
