package auth

import (
	"testing"
)

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{name: "normal", email: "user@example.com", want: "u***r@e***.com"},
		{name: "single local", email: "a@example.com", want: "a***@e***.com"},
		{name: "no at", email: "invalid", want: "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maskEmail(tt.email); got != tt.want {
				t.Fatalf("maskEmail(%q) = %q, want %q", tt.email, got, tt.want)
			}
		})
	}
}
