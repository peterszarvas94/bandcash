package utils

import (
	"context"
	"encoding/base64"
	"testing"
)

func TestGenerateCSRFToken(t *testing.T) {
	token, err := GenerateCSRFToken()
	if err != nil {
		t.Fatalf("GenerateCSRFToken returned error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	if len(token) != 43 {
		t.Fatalf("expected token length 43 for 32 random bytes, got %d", len(token))
	}

	if _, err := base64.RawURLEncoding.DecodeString(token); err != nil {
		t.Fatalf("expected valid base64url token, got decode error: %v", err)
	}
}

func TestContextWithCSRFTokenRoundTrip(t *testing.T) {
	ctx := ContextWithCSRFToken(context.Background(), "abc123")

	if got := CSRFTokenFromContext(ctx); got != "abc123" {
		t.Fatalf("expected csrf token abc123, got %q", got)
	}
}

func TestCSRFTokenFromContextMissingValues(t *testing.T) {
	if got := CSRFTokenFromContext(nil); got != "" {
		t.Fatalf("expected empty token for nil context, got %q", got)
	}

	if got := CSRFTokenFromContext(context.Background()); got != "" {
		t.Fatalf("expected empty token for context without csrf token, got %q", got)
	}
}
