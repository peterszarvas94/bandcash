package email

import (
	"context"
	"strings"
	"testing"
)

func TestJoinBilingualText(t *testing.T) {
	got := joinBilingualText("  Szia  ", "  Hello  ")

	t.Run("includes intro line", func(t *testing.T) {
		if !strings.HasPrefix(got, "English follows below.") {
			t.Fatalf("expected intro prefix, got %q", got)
		}
	})

	t.Run("uses separator between language blocks", func(t *testing.T) {
		if !strings.Contains(got, "\n\n---\n\n") {
			t.Fatalf("expected separator between languages, got %q", got)
		}
	})

	t.Run("contains both localized bodies", func(t *testing.T) {
		if !strings.Contains(got, "Szia") || !strings.Contains(got, "Hello") {
			t.Fatalf("expected both language blocks in output, got %q", got)
		}
	})
}

func TestJoinBilingualHTML(t *testing.T) {
	got, err := joinBilingualHTML(context.Background(), "  <p>HU body</p>  ", "  <p>EN body</p>  ")
	if err != nil {
		t.Fatalf("joinBilingualHTML returned error: %v", err)
	}

	t.Run("returns trimmed html", func(t *testing.T) {
		if got != strings.TrimSpace(got) {
			t.Fatalf("expected trimmed html output, got %q", got)
		}
	})

	t.Run("includes bilingual hint text", func(t *testing.T) {
		if !strings.Contains(got, "English follows below.") {
			t.Fatalf("expected bilingual hint text in html output, got %q", got)
		}
	})

	t.Run("contains both html language blocks", func(t *testing.T) {
		if !strings.Contains(got, "<p>HU body</p>") || !strings.Contains(got, "<p>EN body</p>") {
			t.Fatalf("expected both language html blocks in output, got %q", got)
		}
	})
}
