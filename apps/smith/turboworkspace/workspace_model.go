package turboworkspace

import (
	"encoding/json"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
)

// Workspace is a Turborepo workspace as read from disk.
type Workspace struct {
	Root string
	// Turbo is the root turbo.json.
	Turbo TurboJSON
	// TurboVersion is the turbo dependency range of the root package.json.
	TurboVersion string
	// Manager is the package manager the workspace uses today.
	Manager Manager
	// RootPackage is the root package.json; Root holds what the conversion
	// reads from it besides.
	RootPackage nxworkspace.PackageJSON
	RootExtras  RootExtras
	// Pnpm is pnpm-workspace.yaml, when there is one.
	Pnpm *PnpmWorkspace
	// Globs are the workspace's package globs, wherever they were declared.
	Globs    []string
	Packages []*Package
	// Files are the workspace's files, workspace-relative.
	Files []string
}

// Package is one workspace package.
type Package struct {
	// Dir is the workspace-relative directory, forward slashes.
	Dir  string
	JSON nxworkspace.PackageJSON
	// Name is the package.json name, or the directory's name without one.
	Name string
	// Turbo is the package's own turbo.json (a package configuration), if any.
	Turbo *TurboJSON
	// Files are the package's files, workspace-relative.
	Files []string
}

// Manager is a JavaScript package manager.
type Manager string

// The package managers a Turborepo workspace uses.
const (
	NPM  Manager = "npm"
	PNPM Manager = "pnpm"
	Yarn Manager = "yarn"
	Bun  Manager = "bun"
)

// RootExtras are the root package.json fields that belong to a package
// manager and that the conversion moves or drops.
type RootExtras struct {
	// PackageManager is the `packageManager` field ("pnpm@9.1.0").
	PackageManager string `json:"packageManager"`
	// Resolutions is yarn's `resolutions`.
	Resolutions map[string]json.RawMessage `json:"resolutions"`
	// Pnpm is the `pnpm` field, key by key.
	Pnpm map[string]json.RawMessage `json:"pnpm"`
	// Engines is the `engines` field.
	Engines map[string]string `json:"engines"`
}

// PnpmWorkspace is pnpm-workspace.yaml.
type PnpmWorkspace struct {
	Packages []string `yaml:"packages"`
	// Catalog is the default catalog (`catalog:` specifiers); Catalogs the
	// named ones (`catalog:name`).
	Catalog  map[string]string            `yaml:"catalog"`
	Catalogs map[string]map[string]string `yaml:"catalogs"`
	// Overrides are pnpm 10's `overrides`, which moved here from package.json.
	Overrides map[string]string `yaml:"overrides"`
	// Keys are every top-level key of the file.
	Keys []string `yaml:"-"`
}

// PackageByName returns a package by its package.json name, or nil.
func (w *Workspace) PackageByName(name string) *Package {
	for _, p := range w.Packages {
		if p.Name == name {
			return p
		}
	}
	return nil
}
