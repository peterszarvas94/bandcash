package settings

import (
	"testing"

	appi18n "bandcash/internal/i18n"
)

func TestSettingsDataDefaultsAndBreadcrumb(t *testing.T) {
	s := New()
	data := s.Data(t.Context())
	if data.CurrentLang != appi18n.DefaultLocale {
		t.Fatalf("expected default locale %q, got %q", appi18n.DefaultLocale, data.CurrentLang)
	}
	if len(data.Breadcrumbs) != 1 {
		t.Fatalf("expected one breadcrumb, got %+v", data.Breadcrumbs)
	}
	if data.Breadcrumbs[0].Label == "" {
		t.Fatal("expected non-empty breadcrumb label")
	}
}
