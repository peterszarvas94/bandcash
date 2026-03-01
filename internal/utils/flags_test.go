package utils

import (
	"context"
	"path/filepath"
	"testing"

	"bandcash/internal/db"
)

func setupFlagsTestDB(t *testing.T) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "flags-test.db")
	if err := db.Init(dbPath); err != nil {
		t.Fatalf("db init failed: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("db migrate failed: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
		db.DB = nil
		db.Qry = nil
	})
}

func TestSignupFlag_DefaultFalseWhenMissing(t *testing.T) {
	setupFlagsTestDB(t)

	enabled, err := IsSignupEnabled(context.Background())
	if err != nil {
		t.Fatalf("IsSignupEnabled returned error: %v", err)
	}
	if enabled {
		t.Fatal("expected signup flag to default to false when missing")
	}
}

func TestSignupFlag_SetAndReadRoundTrip(t *testing.T) {
	setupFlagsTestDB(t)

	ctx := context.Background()

	if err := SetSignupEnabled(ctx, true); err != nil {
		t.Fatalf("SetSignupEnabled(true) returned error: %v", err)
	}
	enabled, err := IsSignupEnabled(ctx)
	if err != nil {
		t.Fatalf("IsSignupEnabled returned error after set true: %v", err)
	}
	if !enabled {
		t.Fatal("expected signup flag true after SetSignupEnabled(true)")
	}

	if err := SetSignupEnabled(ctx, false); err != nil {
		t.Fatalf("SetSignupEnabled(false) returned error: %v", err)
	}
	enabled, err = IsSignupEnabled(ctx)
	if err != nil {
		t.Fatalf("IsSignupEnabled returned error after set false: %v", err)
	}
	if enabled {
		t.Fatal("expected signup flag false after SetSignupEnabled(false)")
	}
}
