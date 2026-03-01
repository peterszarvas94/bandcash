package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
)

const CSRFCookieName = "_csrf"

type csrfContextKey struct{}

func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func ContextWithCSRFToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, csrfContextKey{}, token)
}

func CSRFTokenFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	token, _ := ctx.Value(csrfContextKey{}).(string)
	return token
}
