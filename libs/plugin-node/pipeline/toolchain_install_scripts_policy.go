package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
)

// toolchainInstallScripts are the packages with an install script that the
// default stack's tools bring in, by the root devDependency that brings them:
// jest through jest-haste-map (@parcel/watcher) and jest-resolve
// (unrs-resolver), @monosmith/eslint-config through its import resolver
// (unrs-resolver). They are native-binary downloads of the tools smith
// generated, so the opt-in node audit policy lets them run without a line in
// [ci] audit_allow_install_scripts; any other install script still fails.
var toolchainInstallScripts = map[string][]string{
	"jest":                     {"@parcel/watcher", "unrs-resolver"},
	"@monosmith/eslint-config": {"unrs-resolver"},
}

// AllowedInstallScripts is the allow-list the node audit policy passes to
// `smith npm-policy`: the workspace's own audit_allow_install_scripts, then the
// packages the root package.json's devDependencies bring with them.
func AllowedInstallScripts(workspaceRoot string, configured []string) []string {
	allowed := slices.Clone(configured)
	var manifest struct {
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if raw, err := os.ReadFile(filepath.Join(workspaceRoot, "package.json")); err == nil {
		_ = json.Unmarshal(raw, &manifest)
	}
	tools := make([]string, 0, len(manifest.DevDependencies))
	for tool := range manifest.DevDependencies {
		tools = append(tools, tool)
	}
	slices.Sort(tools)
	for _, tool := range tools {
		for _, name := range toolchainInstallScripts[tool] {
			if !slices.Contains(allowed, name) {
				allowed = append(allowed, name)
			}
		}
	}
	return allowed
}
