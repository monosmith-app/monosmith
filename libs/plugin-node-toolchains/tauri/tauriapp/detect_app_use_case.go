package tauriapp

import (
	"os"
	"path/filepath"
	"regexp"
)

// Name is the tauri plugin's name in a stack's toolchain list and in smith.toml
// ([conventions.tauri], [projects.<p>.settings.tauri]).
const Name = "tauri"

// SrcDir is the folder of a Tauri app's Rust side, inside the Node project.
const SrcDir = "src-tauri"

// confFiles are the names Tauri reads its configuration from.
var confFiles = []string{"tauri.conf.json", "tauri.conf.json5", "Tauri.toml"}

// tauriDependency matches `tauri = ...` and `[dependencies.tauri]` in a Cargo.toml.
var tauriDependency = regexp.MustCompile(`(?m)^\s*tauri\s*=|^\s*\[(?:\w+-)?dependencies\.tauri\]`)

// Detect says whether the project at root (workspace-relative) holds a Tauri
// app: a src-tauri with a Tauri configuration, or a Cargo.toml that depends
// on tauri.
func Detect(workspaceRoot, root string) bool {
	src := filepath.Join(workspaceRoot, filepath.FromSlash(root), SrcDir)
	for _, name := range confFiles {
		if _, err := os.Stat(filepath.Join(src, name)); err == nil {
			return true
		}
	}
	cargo, err := os.ReadFile(filepath.Join(src, "Cargo.toml"))
	return err == nil && tauriDependency.Match(cargo)
}
