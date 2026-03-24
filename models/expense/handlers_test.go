package expense

import "testing"

func TestNormalizePaidAtInput(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "trim and keep date", in: " 2026-04-12 ", want: "2026-04-12"},
		{name: "invalid kept trimmed", in: " not-a-date ", want: "not-a-date"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePaidAtInput(tt.in)
			if got != tt.want {
				t.Fatalf("normalizePaidAtInput(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestPaidAtArg(t *testing.T) {
	t.Run("unpaid always null", func(t *testing.T) {
		got := paidAtArg(false, "2026-04-12")
		if got.Valid {
			t.Fatalf("expected invalid null string for unpaid, got %+v", got)
		}
	})

	t.Run("paid with empty keeps explicit empty", func(t *testing.T) {
		got := paidAtArg(true, "")
		if !got.Valid || got.String != "" {
			t.Fatalf("expected valid empty string, got %+v", got)
		}
	})

	t.Run("paid with date uses normalized date", func(t *testing.T) {
		got := paidAtArg(true, " 2026-04-12 ")
		if !got.Valid || got.String != "2026-04-12" {
			t.Fatalf("expected valid normalized date, got %+v", got)
		}
	})
}
