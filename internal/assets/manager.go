package assets

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

const (
	manifestPath  = "static/gen/manifest.json"
	importMapPath = "static/gen/importmap.json"
	importMapSrc  = "static/importmap.json"
)

var (
	loadOnce   sync.Once
	assetPaths map[string]string
	importMap  string
)

func AssetPath(logical string) string {
	load()
	normalized := strings.TrimPrefix(logical, "/")
	if url, ok := assetPaths[normalized]; ok {
		return url
	}
	return "/static/" + normalized
}

func ImportMapJSON() string {
	load()
	if importMap == "" {
		return `{"imports":{}}`
	}
	return importMap
}

func load() {
	loadOnce.Do(func() {
		assetPaths = map[string]string{}
		manifestContent, err := os.ReadFile(manifestPath)
		if err == nil {
			_ = json.Unmarshal(manifestContent, &assetPaths)
		}

		importMapContent, err := os.ReadFile(importMapPath)
		if err == nil && json.Valid(importMapContent) {
			importMap = string(importMapContent)
			return
		}

		sourceImportMap, sourceErr := os.ReadFile(importMapSrc)
		if sourceErr == nil && json.Valid(sourceImportMap) {
			importMap = string(sourceImportMap)
		}
	})
}
