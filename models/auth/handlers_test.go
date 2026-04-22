package auth

import "testing"

func TestMaskEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal email keeps single ellipsis in domain",
			input: "cat@peter.hu",
			want:  "c...t@p...hu",
		},
		{
			name:  "single char local part",
			input: "c@peter.hu",
			want:  "c...@p...hu",
		},
		{
			name:  "single label domain",
			input: "cat@peter",
			want:  "c...t@p...",
		},
		{
			name:  "invalid email unchanged",
			input: "not-an-email",
			want:  "not-an-email",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := maskEmail(tt.input)
			if got != tt.want {
				t.Fatalf("maskEmail(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
