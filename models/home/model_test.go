package home

import (
	"testing"

	appi18n "bandcash/internal/i18n"
)

func TestHomeDataDefaults(t *testing.T) {
	data := Data(t.Context())
	if data.CurrentLang != appi18n.DefaultLocale {
		t.Fatalf("expected default locale %q, got %q", appi18n.DefaultLocale, data.CurrentLang)
	}
	if len(data.Breadcrumbs) != 0 {
		t.Fatalf("expected empty breadcrumbs, got %+v", data.Breadcrumbs)
	}
}
