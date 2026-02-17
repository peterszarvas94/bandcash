package utils

import "encoding/json"

// JSString returns a JSON-encoded string literal for safe JS embedding.
func JSString(value string) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "\"\""
	}
	return string(encoded)
}
