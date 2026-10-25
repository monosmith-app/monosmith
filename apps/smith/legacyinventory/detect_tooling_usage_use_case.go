package legacyinventory

import (
	"path"
	"strings"
)

// SharedLintConfigs are the shared ESLint configs known to be replaceable by
// smith's, which has the same call: config({ workspaceRoot }).
var SharedLintConfigs = []string{"@mnci/eslint-config"}

// rootLintConfigs are the names of a root ESLint flat config, in the order
// ESLint looks for them.
var rootLintConfigs = []string{"eslint.config.js", "eslint.config.mjs", "eslint.config.cjs", "eslint.config.ts", "eslint.config.mts", "eslint.config.cts"}

// RootLintConfig is the workspace's root ESLint flat config, "" when none.
func RootLintConfig(t *Tree) string {
	for _, name := range rootLintConfigs {
		if t.Has(name) {
			return name
		}
	}
	return ""
}

// Detect reads which lint and test tooling the workspace uses.
func Detect(t *Tree) Usage {
	var u Usage
	u.LintConfig = RootLintConfig(t)
	u.ESLint = u.LintConfig != ""
	if u.LintConfig != "" {
		text := t.Read(u.LintConfig)
		for _, shared := range SharedLintConfigs {
			if strings.Contains(text, "'"+shared+"'") || strings.Contains(text, `"`+shared+`"`) {
				u.SharedLintConfig = shared
			}
		}
	}
	for _, f := range t.Files() {
		if _, ok := fileFamily(f); !ok {
			continue
		}
		switch {
		case strings.HasPrefix(path.Base(f), "jest.config."):
			u.Jest = true
		case strings.HasPrefix(path.Base(f), "vitest.config."), strings.HasPrefix(path.Base(f), "vitest.workspace."):
			u.Vitest = true
		case strings.HasPrefix(path.Base(f), ".eslintrc"):
			u.ESLint = true
		}
	}
	for _, mf := range t.ManifestPaths() {
		m, ok := ReadManifest(t.Read(mf))
		if !ok {
			continue
		}
		for _, name := range m.Declared() {
			switch name {
			case "jest":
				u.Jest = true
			case "vitest":
				u.Vitest = true
			case "eslint":
				u.ESLint = true
			}
		}
	}
	return u
}
