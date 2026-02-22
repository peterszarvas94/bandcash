package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}

	goModPath := filepath.Join(root, "go.mod")
	misePath := filepath.Join(root, "mise.toml")

	goVersion, templModuleVersion, err := parseGoMod(goModPath)
	if err != nil {
		log.Fatalf("failed to parse go.mod: %v", err)
	}

	miseGoVersion, miseTemplVersion, err := parseMise(misePath)
	if err != nil {
		log.Fatalf("failed to parse mise.toml: %v", err)
	}

	if goVersion != miseGoVersion {
		log.Fatalf("Go version mismatch: go.mod=%s mise.toml=%s", goVersion, miseGoVersion)
	}

	if templModuleVersion != miseTemplVersion {
		log.Fatalf("templ version mismatch: go.mod=%s mise.toml=%s", templModuleVersion, miseTemplVersion)
	}

	fmt.Printf("Version parity OK (go=%s, templ=%s)\n", goVersion, templModuleVersion)
}

func parseGoMod(path string) (string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	var goVersion string
	var templVersion string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if goVersion == "" && strings.HasPrefix(line, "go ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				goVersion = parts[1]
			}
			continue
		}

		if templVersion == "" {
			parts := strings.Fields(line)
			if len(parts) >= 2 && parts[0] == "github.com/a-h/templ" {
				templVersion = parts[1]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", err
	}

	if goVersion == "" {
		return "", "", fmt.Errorf("missing go version")
	}
	if templVersion == "" {
		return "", "", fmt.Errorf("missing github.com/a-h/templ version")
	}

	return goVersion, templVersion, nil
}

func parseMise(path string) (string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	var goVersion string
	var templVersion string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if goVersion == "" && strings.HasPrefix(line, "go =") {
			goVersion = extractQuotedValue(line)
			continue
		}

		if templVersion == "" && strings.HasPrefix(line, "\"go:github.com/a-h/templ/cmd/templ\"") {
			templVersion = extractQuotedValue(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", err
	}

	if goVersion == "" {
		return "", "", fmt.Errorf("missing go tool version")
	}
	if templVersion == "" {
		return "", "", fmt.Errorf("missing templ CLI version")
	}

	return goVersion, templVersion, nil
}

func extractQuotedValue(line string) string {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return ""
	}
	right := strings.TrimSpace(parts[1])
	return strings.Trim(right, "\"")
}
