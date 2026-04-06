package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type assetSpec struct {
	Logical string
	Source  string
}

type importMapFile struct {
	Imports map[string]string `json:"imports"`
}

var coreAssetSpecs = []assetSpec{
	{Logical: "css/base.css", Source: "static/css/base.css"},
	{Logical: "css/components.css", Source: "static/css/components.css"},
	{Logical: "css/utilities.css", Source: "static/css/utilities.css"},
	{Logical: "css/style.css", Source: "static/css/style.css"},
	{Logical: "js/vendor/datastar.js", Source: "static/js/vendor/datastar.js"},
	{Logical: "js/notifications.js", Source: "static/js/notifications.js"},
	{Logical: "js/main.js", Source: "static/js/main.js"},
}

var devOnlyAssetSpecs = []assetSpec{}

func main() {
	appEnv := normalizedAppEnv()
	isDevelopment := appEnv == "development"

	token := ""
	if !isDevelopment {
		var err error
		token, err = generateToken()
		if err != nil {
			panic(err)
		}
	}

	sourceImportMap, err := loadSourceImportMap("static/importmap.json")
	if err != nil {
		panic(err)
	}

	assetSpecs := append([]assetSpec{}, coreAssetSpecs...)
	if isDevelopment {
		assetSpecs = append(assetSpecs, devOnlyAssetSpecs...)
	}

	if err := os.RemoveAll("static/gen"); err != nil {
		panic(err)
	}
	if err := os.MkdirAll("static/gen", 0o755); err != nil {
		panic(err)
	}

	manifest := map[string]string{}
	targetRelativePath := map[string]string{}
	for _, spec := range assetSpecs {
		if isDevelopment {
			targetRelativePath[spec.Logical] = spec.Logical
		} else {
			targetRelativePath[spec.Logical] = fingerprintedPath(spec.Logical, token)
		}
		manifest[spec.Logical] = "/static/gen/" + targetRelativePath[spec.Logical]
	}

	for _, spec := range assetSpecs {
		content, err := os.ReadFile(spec.Source)
		if err != nil {
			panic(err)
		}

		data := string(content)
		if strings.HasPrefix(spec.Logical, "css/") {
			for logical, url := range manifest {
				if !strings.HasPrefix(logical, "css/") {
					continue
				}
				data = strings.ReplaceAll(data, "/static/"+logical, url)
			}
		}

		target := filepath.Join("static/gen", filepath.FromSlash(targetRelativePath[spec.Logical]))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			panic(err)
		}
		if err := os.WriteFile(target, []byte(data), 0o644); err != nil {
			panic(err)
		}
	}

	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("static/gen/manifest.json", append(manifestJSON, '\n'), 0o644); err != nil {
		panic(err)
	}

	generatedImportMap := importMapFile{Imports: map[string]string{}}
	for module, sourceURL := range sourceImportMap.Imports {
		logical := strings.TrimPrefix(sourceURL, "/static/")
		if rewritten, ok := manifest[logical]; ok {
			generatedImportMap.Imports[module] = rewritten
			continue
		}
		generatedImportMap.Imports[module] = sourceURL
	}
	importMapJSON, err := json.MarshalIndent(generatedImportMap, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("static/gen/importmap.json", append(importMapJSON, '\n'), 0o644); err != nil {
		panic(err)
	}

	if isDevelopment {
		fmt.Println("generated development assets without fingerprint token")
		return
	}

	fmt.Printf("generated tokenized assets with token: %s\n", token)
}

func normalizedAppEnv() string {
	appEnv := strings.TrimSpace(strings.ToLower(os.Getenv("APP_ENV")))
	if appEnv == "" {
		return "development"
	}
	return appEnv
}

func generateToken() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func fingerprintedPath(logical string, token string) string {
	ext := filepath.Ext(logical)
	base := strings.TrimSuffix(logical, ext)
	return base + "-" + token + ext
}

func loadSourceImportMap(path string) (importMapFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return importMapFile{}, err
	}

	parsed := importMapFile{}
	if err := json.Unmarshal(content, &parsed); err != nil {
		return importMapFile{}, err
	}
	if parsed.Imports == nil {
		parsed.Imports = map[string]string{}
	}

	return parsed, nil
}
