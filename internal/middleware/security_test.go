package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestIsStateChangingMethod(t *testing.T) {
	tests := []struct {
		method string
		want   bool
	}{
		{method: http.MethodGet, want: false},
		{method: http.MethodHead, want: false},
		{method: http.MethodOptions, want: false},
		{method: http.MethodPost, want: true},
		{method: http.MethodPut, want: true},
		{method: http.MethodPatch, want: true},
		{method: http.MethodDelete, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			got := isStateChangingMethod(tt.method)
			if got != tt.want {
				t.Fatalf("isStateChangingMethod(%q) = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}

func TestIsSameOrigin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "https://app.example.com/dashboard", nil)
	req.Host = "app.example.com"
	c := e.NewContext(req, httptest.NewRecorder())

	if !isSameOrigin(c, "https://app.example.com/groups") {
		t.Fatal("expected same origin to pass")
	}

	if isSameOrigin(c, "https://evil.example.com/phish") {
		t.Fatal("expected different origin to fail")
	}

	if isSameOrigin(c, "://bad-url") {
		t.Fatal("expected malformed url to fail")
	}
}

func TestCSRFFromRequest(t *testing.T) {
	t.Run("header token has priority", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"csrf":"body-token"}`))
		r.Header.Set("X-CSRF-Token", "header-token")
		r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		got, err := csrfFromRequest(r)
		if err != nil {
			t.Fatalf("csrfFromRequest returned error: %v", err)
		}
		if got != "header-token" {
			t.Fatalf("expected header token, got %q", got)
		}
	})

	t.Run("form token", func(t *testing.T) {
		form := url.Values{"csrf": {"form-token"}}
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
		r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

		got, err := csrfFromRequest(r)
		if err != nil {
			t.Fatalf("csrfFromRequest returned error: %v", err)
		}
		if got != "form-token" {
			t.Fatalf("expected form token, got %q", got)
		}
	})

	t.Run("json csrf token", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"csrf":"json-token"}`))
		r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		got, err := csrfFromRequest(r)
		if err != nil {
			t.Fatalf("csrfFromRequest returned error: %v", err)
		}
		if got != "json-token" {
			t.Fatalf("expected json token, got %q", got)
		}
	})

	t.Run("json signals csrf token", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"signals":{"csrf":"signals-token"}}`))
		r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		got, err := csrfFromRequest(r)
		if err != nil {
			t.Fatalf("csrfFromRequest returned error: %v", err)
		}
		if got != "signals-token" {
			t.Fatalf("expected signals token, got %q", got)
		}
	})

	t.Run("malformed json returns empty token", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"csrf":`))
		r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		got, err := csrfFromRequest(r)
		if err != nil {
			t.Fatalf("csrfFromRequest returned error: %v", err)
		}
		if got != "" {
			t.Fatalf("expected empty token, got %q", got)
		}
	})

	t.Run("empty body returns empty token", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))

		got, err := csrfFromRequest(r)
		if err != nil {
			t.Fatalf("csrfFromRequest returned error: %v", err)
		}
		if got != "" {
			t.Fatalf("expected empty token, got %q", got)
		}
	})
}
