package middleware

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"

	"bandcash/internal/utils"
)

func FetchSiteProtection() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !isStateChangingMethod(c.Request().Method) {
				return next(c)
			}

			fetchSite := strings.ToLower(strings.TrimSpace(c.Request().Header.Get("Sec-Fetch-Site")))
			if fetchSite != "" && fetchSite != "same-origin" && fetchSite != "same-site" && fetchSite != "none" {
				return c.NoContent(http.StatusForbidden)
			}

			return next(c)
		}
	}
}

func OriginProtection() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !isStateChangingMethod(c.Request().Method) {
				return next(c)
			}

			origin := c.Request().Header.Get(echo.HeaderOrigin)
			if origin != "" {
				if !isSameOrigin(c, origin) {
					return c.NoContent(http.StatusForbidden)
				}

				return next(c)
			}

			referer := c.Request().Header.Get("Referer")
			if referer == "" || !isSameOrigin(c, referer) {
				return c.NoContent(http.StatusForbidden)
			}

			return next(c)
		}
	}
}

func CSRFToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := ""
			cookie, err := c.Cookie(utils.CSRFCookieName)
			if err == nil {
				token = cookie.Value
			}

			if token == "" {
				env := utils.Env()
				token, err = utils.GenerateCSRFToken()
				if err != nil {
					return c.NoContent(http.StatusInternalServerError)
				}

				c.SetCookie(&http.Cookie{
					Name:     utils.CSRFCookieName,
					Value:    token,
					Path:     "/",
					MaxAge:   86400 * 30,
					HttpOnly: true,
					Secure:   env.AppEnv == "production",
					SameSite: http.SameSiteLaxMode,
				})
			}

			ctx := utils.ContextWithCSRFToken(c.Request().Context(), token)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func CSRFProtection() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !isStateChangingMethod(c.Request().Method) {
				return next(c)
			}

			expected := utils.CSRFTokenFromContext(c.Request().Context())
			if expected == "" {
				return c.NoContent(http.StatusForbidden)
			}

			provided, err := csrfFromRequest(c.Request())
			if err != nil || provided == "" {
				return c.NoContent(http.StatusForbidden)
			}

			if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
				return c.NoContent(http.StatusForbidden)
			}

			return next(c)
		}
	}
}

func csrfFromRequest(r *http.Request) (string, error) {
	headerToken := strings.TrimSpace(r.Header.Get("X-CSRF-Token"))
	if headerToken != "" {
		return headerToken, nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	if len(bytes.TrimSpace(body)) == 0 {
		return "", nil
	}

	contentType := strings.ToLower(r.Header.Get(echo.HeaderContentType))
	if strings.Contains(contentType, echo.MIMEApplicationForm) {
		if err := r.ParseForm(); err != nil {
			return "", err
		}
		return strings.TrimSpace(r.FormValue("csrf")), nil
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", nil
	}

	if token, ok := payload["csrf"].(string); ok {
		return strings.TrimSpace(token), nil
	}

	if signals, ok := payload["signals"].(map[string]any); ok {
		if token, ok := signals["csrf"].(string); ok {
			return strings.TrimSpace(token), nil
		}
	}

	return "", nil
}

func isSameOrigin(c echo.Context, rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}

	requestOrigin := strings.ToLower(c.Scheme() + "://" + c.Request().Host)
	valueOrigin := strings.ToLower(parsed.Scheme + "://" + parsed.Host)

	return requestOrigin == valueOrigin
}

func isStateChangingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
