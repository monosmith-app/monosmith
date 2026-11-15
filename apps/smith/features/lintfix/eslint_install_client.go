package lintfix

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// InstalledESLint is the version of the ESLint under root's node_modules, ""
// when none is installed.
func InstalledESLint(root string) string {
	raw, err := os.ReadFile(filepath.Join(root, "node_modules", "eslint", "package.json"))
	if err != nil {
		return ""
	}
	var manifest struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(raw, &manifest) != nil {
		return ""
	}
	return manifest.Version
}
