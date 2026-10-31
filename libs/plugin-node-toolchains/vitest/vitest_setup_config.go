package vitest

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tooling"
	"github.com/monosmith-app/monosmith/libs/plugin-node/modulealiases"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name in a stack's toolchain list.
const Name = "vitest"

// stacks are the stacks this plugin publishes.
var stacks = []protocol.Stack{{
	Name:        "node-vitest",
	Description: "TypeScript on Node, linted by @monosmith/eslint-config, tested with Vitest",
	Language:    "node",
	Scaffold:    "typescript",
	Toolchain:   []string{"typescript", "eslint", Name},
}}

// New returns the plugin; version is reported in the handshake.
func New(version string) *tooling.Toolchain {
	return &tooling.Toolchain{Name: Name, Version: version, Provides: []string{"test"}, Stacks: stacks, Setup: setup}
}

func setup(workspaceRoot string, params protocol.GenerateParams) tooling.Setup {
	files := tooling.StarterSpec(workspaceRoot, params, "vitest")
	// A module.aliases fact reaches Vitest through a vitest.config.ts, unless
	// the project has a vite.config.ts: Vitest reads that, and the vite
	// plugin already wrote the aliases into it.
	if alias := modulealiases.FromFacts(params.Facts).ViteResolveAlias(); alias != "" && !hasViteConfig(workspaceRoot, params.Directory) {
		if files == nil {
			files = map[string]string{}
		}
		files["vitest.config.ts"] = "import { defineConfig } from 'vitest/config'\n\nexport default defineConfig({\n" + alias + "})\n"
	}
	return tooling.Setup{
		DevDependencies: map[string]string{"vitest": "^5.0.5"},
		Files:           files,
		Scripts:         map[string]string{"test": "vitest run"},
	}
}

// hasViteConfig reports whether the project directory already has a Vite config.
func hasViteConfig(workspaceRoot, dir string) bool {
	_, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(dir), "vite.config.ts"))
	return err == nil
}
