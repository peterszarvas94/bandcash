package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"

	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
	shared "bandcash/models/shared"
	icons "bandcash/models/shared/icons"
)

func ErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		status := http.StatusInternalServerError
		if httpErr, ok := err.(*echo.HTTPError); ok {
			status = httpErr.Code
		}

		if !wantsHTMLErrorPage(c) {
			_ = c.NoContent(status)
			return
		}

		data := shared.ErrorPageData{
			Title:      errorPageTitle(c, status),
			StatusCode: status,
			IconName:   errorPageIcon(status),
			Heading:    errorPageTitle(c, status),
			Message:    errorPageBody(c, status),
			HomeLabel:  ctxi18n.T(c.Request().Context(), "error_pages.home_action"),
			HomeHref:   appi18n.LocalizedHomePath(c.Request().Context()),
		}

		if renderErr := utils.RenderPage(c, shared.ErrorPage(data)); renderErr != nil {
			slog.Warn("http.error: failed to render error page", "status", status, "err", renderErr)
			_ = c.NoContent(status)
		}
	}
}

func wantsHTMLErrorPage(c echo.Context) bool {
	request := c.Request()
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		return false
	}

	if request.URL.Path == "/sse" || strings.Contains(strings.ToLower(request.Header.Get("Accept")), "text/event-stream") {
		return false
	}

	accept := strings.ToLower(strings.TrimSpace(request.Header.Get("Accept")))
	if accept == "" {
		return true
	}

	return strings.Contains(accept, "text/html") || strings.Contains(accept, "*/*")
}

func errorPageTitle(c echo.Context, status int) string {
	switch status {
	case http.StatusBadRequest:
		return ctxi18n.T(c.Request().Context(), "error_pages.generic.bad_request_title")
	case http.StatusForbidden:
		return ctxi18n.T(c.Request().Context(), "error_pages.generic.forbidden_title")
	case http.StatusNotFound:
		return ctxi18n.T(c.Request().Context(), "error_pages.generic.not_found_title")
	case http.StatusTooManyRequests:
		return ctxi18n.T(c.Request().Context(), "error_pages.generic.too_many_requests_title")
	default:
		return ctxi18n.T(c.Request().Context(), "error_pages.generic.internal_title")
	}
}

func errorPageBody(c echo.Context, status int) string {
	switch status {
	case http.StatusBadRequest:
		return ctxi18n.T(c.Request().Context(), "error_pages.generic.bad_request_body")
	case http.StatusForbidden:
		return ctxi18n.T(c.Request().Context(), "error_pages.generic.forbidden_body")
	case http.StatusNotFound:
		return ctxi18n.T(c.Request().Context(), "error_pages.generic.not_found_body")
	case http.StatusTooManyRequests:
		return ctxi18n.T(c.Request().Context(), "error_pages.generic.too_many_requests_body")
	default:
		return ctxi18n.T(c.Request().Context(), "error_pages.generic.internal_body")
	}
}

func errorPageIcon(status int) icons.IconName {
	switch status {
	case http.StatusBadRequest:
		return icons.IconCircleX
	case http.StatusForbidden:
		return icons.IconShieldAlert
	case http.StatusNotFound:
		return icons.IconSearchX
	case http.StatusTooManyRequests:
		return icons.IconBan
	default:
		return icons.IconServerCrash
	}
}
