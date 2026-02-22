package utils

import (
	"crypto/rand"
	"regexp"
	"strings"
)

const alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var idPattern = regexp.MustCompile(`^[a-z]{3}_[A-Za-z0-9]{20}$`)

// GenerateID creates a new ID with the given 3-letter prefix followed by 20 random alphanumeric characters
// Format: "xxx_" + 20 random alphanumeric chars (no underscores or hyphens)
// Example: "evt_7x9KpQmN3vLwAbCdEfGh"
func GenerateID(prefix string) string {
	if len(prefix) != 3 {
		panic("ID prefix must be exactly 3 characters")
	}
	prefix = strings.ToLower(prefix)

	// Generate 20 random alphanumeric characters
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		panic("failed to generate random bytes: " + err.Error())
	}

	// Map to alphanumeric characters
	for i := range b {
		b[i] = alphanumeric[int(b[i])%len(alphanumeric)]
	}

	return prefix + "_" + string(b)
}

func IsValidID(id string, prefix string) bool {
	if !idPattern.MatchString(id) {
		return false
	}
	if len(prefix) != 3 {
		return false
	}
	return strings.HasPrefix(id, strings.ToLower(prefix)+"_")
}

// ID prefixes for different entity types
const (
	PrefixEvent       = "evt"
	PrefixMember      = "mem"
	PrefixParticipant = "par"
)
