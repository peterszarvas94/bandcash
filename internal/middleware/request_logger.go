package middleware

import (
	"encoding/json"
	"log/slog"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func RequestLogger(next echo.HandlerFunc) echo.HandlerFunc {
	mw := echoMiddleware.RequestLoggerWithConfig(echoMiddleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    false,
		LogValuesFunc: func(c echo.Context, v echoMiddleware.RequestLoggerValues) error {
			req := c.Request()
			slog.Info("http.request.completed",
				"path", req.URL.Path,
				"query", buildQueryLog(req.URL.Query()),
				"method", req.Method,
				"status", v.Status,
			)
			return nil
		},
	})

	return mw(next)
}

func buildQueryLog(values url.Values) map[string]any {
	out := make(map[string]any, len(values))
	for key, items := range values {
		if len(items) == 0 {
			out[key] = ""
			continue
		}

		if key == "datastar" {
			if len(items) == 1 {
				out[key] = parseDatastarQuery(items[0])
				continue
			}

			parsed := make([]any, len(items))
			for i, item := range items {
				parsed[i] = parseDatastarQuery(item)
			}
			out[key] = parsed
			continue
		}

		if len(items) == 1 {
			out[key] = items[0]
			continue
		}

		vals := make([]string, len(items))
		copy(vals, items)
		out[key] = vals
	}

	return out
}

func parseDatastarQuery(raw string) any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	var payload any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return raw
	}

	return redactSensitive(payload)
}

func redactSensitive(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for key, value := range t {
			if isSensitiveKey(key) {
				out[key] = "[REDACTED]"
				continue
			}
			out[key] = redactSensitive(value)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, item := range t {
			out[i] = redactSensitive(item)
		}
		return out
	default:
		return v
	}
}

func isSensitiveKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "csrf", "token", "password", "secret", "authorization", "cookie":
		return true
	default:
		return false
	}
}
