package jest

import (
	"maps"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tooling"
	"github.com/monosmith-app/monosmith/libs/plugin-node/modulealiases"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name in a stack's toolchain list.
const Name = "jest"

// stacks are the stacks this plugin publishes.
var stacks = []protocol.Stack{{
	Name:        "node-jest",
	Description: "TypeScript on Node, linted by @monosmith/eslint-config, tested with Jest",
	Language:    "node",
	Scaffold:    "typescript",
	Toolchain:   []string{"typescript", "eslint", Name},
}}

// New returns the plugin; version is reported in the handshake.
func New(version string) *tooling.Toolchain {
	return &tooling.Toolchain{Name: Name, Version: version, Provides: []string{"test"}, Stacks: stacks, Setup: setup}
}

func setup(workspaceRoot string, params protocol.GenerateParams) tooling.Setup {
	// A module.aliases fact adds its aliases to the mapper and to the
	// compiler options ts-jest type-checks with; without one, neither changes.
	aliases := modulealiases.FromFacts(params.Facts)
	mapper := map[string]string{`^(\.{1,2}/.*)\.js$`: "$1"}
	maps.Copy(mapper, aliases.JestModuleNameMapper())
	compiler := map[string]any{"module": "commonjs", "moduleResolution": "bundler", "verbatimModuleSyntax": false}
	if paths := aliases.CompilerPaths(); paths != nil {
		compiler["paths"] = paths
	}
	return tooling.Setup{
		DevDependencies: map[string]string{
			"@jest/globals": "^30.5.2",
			"jest":          "^30.5.2",
			"ts-jest":       "^29.4.15",
		},
		Files:   tooling.StarterSpec(workspaceRoot, params, "@jest/globals"),
		Scripts: map[string]string{"test": "jest"},
		// Jest runs CommonJS: ts-jest compiles to it, and a relative import
		// rewritten to .js finds the .ts source.
		Fields: map[string]any{"jest": map[string]any{
			"testEnvironment":  "node",
			"roots":            []string{"<rootDir>/src"},
			"moduleNameMapper": mapper,
			"transform":        map[string]any{`^.+\.ts$`: []any{"ts-jest", map[string]any{"tsconfig": compiler}}},
		}},
	}
}
