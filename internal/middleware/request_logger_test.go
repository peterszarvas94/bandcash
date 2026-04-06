package middleware

import (
	"net/url"
	"reflect"
	"testing"
)

func TestBuildQueryLog(t *testing.T) {
	t.Run("keeps regular query params readable", func(t *testing.T) {
		query := url.Values{
			"tab_id": {"tab_123"},
			"tag":    {"one", "two"},
		}

		got := buildQueryLog(query)

		want := map[string]any{
			"tab_id": "tab_123",
			"tag":    []string{"one", "two"},
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("buildQueryLog() = %#v, want %#v", got, want)
		}
	})

	t.Run("parses datastar json and redacts sensitive keys", func(t *testing.T) {
		query := url.Values{
			"datastar": {`{"mode":"single","csrf":"abc","nested":{"token":"secret-token"},"rows":[{"password":"pw"}]}`},
		}

		got := buildQueryLog(query)

		datastar, ok := got["datastar"].(map[string]any)
		if !ok {
			t.Fatalf("datastar should be parsed object, got %T", got["datastar"])
		}

		if datastar["mode"] != "single" {
			t.Fatalf("mode = %v, want single", datastar["mode"])
		}

		if datastar["csrf"] != "[REDACTED]" {
			t.Fatalf("csrf = %v, want [REDACTED]", datastar["csrf"])
		}

		nested, ok := datastar["nested"].(map[string]any)
		if !ok {
			t.Fatalf("nested should be object, got %T", datastar["nested"])
		}
		if nested["token"] != "[REDACTED]" {
			t.Fatalf("nested.token = %v, want [REDACTED]", nested["token"])
		}

		rows, ok := datastar["rows"].([]any)
		if !ok || len(rows) != 1 {
			t.Fatalf("rows should be 1-item array, got %#v", datastar["rows"])
		}
		firstRow, ok := rows[0].(map[string]any)
		if !ok {
			t.Fatalf("first row should be object, got %T", rows[0])
		}
		if firstRow["password"] != "[REDACTED]" {
			t.Fatalf("rows[0].password = %v, want [REDACTED]", firstRow["password"])
		}
	})

	t.Run("keeps datastar string when json is invalid", func(t *testing.T) {
		query := url.Values{
			"datastar": {"{not-json"},
		}

		got := buildQueryLog(query)

		if got["datastar"] != "{not-json" {
			t.Fatalf("datastar = %v, want original string", got["datastar"])
		}
	})
}
