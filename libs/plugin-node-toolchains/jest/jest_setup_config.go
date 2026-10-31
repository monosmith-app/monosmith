package jest

import (
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tooling"
	"github.com/monosmith-app/monosmith/libs/plugin-node/discovery"
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
	return &tooling.Toolchain{Name: Name, Version: version, Provides: []string{"test"}, Stacks: stacks, Follows: []string{discovery.ShapeTestRunner}, Setup: setup}
}

func setup(workspaceRoot string, params protocol.GenerateParams) tooling.Setup {
	devDependencies, transform := tsJest()
	if params.Shape[discovery.ShapeTestRunner] == discovery.ShapeSwcJest {
		devDependencies, transform = swcJest()
	}
	devDependencies["@jest/globals"] = "^30.5.2"
	devDependencies["jest"] = "^30.5.2"
	return tooling.Setup{
		DevDependencies: devDependencies,
		Files:           tooling.StarterSpec(workspaceRoot, params, "@jest/globals"),
		Scripts:         map[string]string{"test": "jest"},
		// Jest runs CommonJS: the compiler turns TypeScript into it, and a
		// relative import rewritten to .js finds the .ts source.
		Fields: map[string]any{"jest": map[string]any{
			"testEnvironment":  "node",
			"roots":            []string{"<rootDir>/src"},
			"moduleNameMapper": map[string]string{`^(\.{1,2}/.*)\.js$`: "$1"},
			"transform":        map[string]any{`^.+\.ts$`: transform},
		}},
	}
}

// tsJest is the default compiler of the tests: ts-jest.
func tsJest() (devDependencies map[string]string, transform []any) {
	return map[string]string{"ts-jest": "^29.4.15"}, []any{"ts-jest", map[string]any{"tsconfig": map[string]any{
		"module": "commonjs", "moduleResolution": "bundler", "verbatimModuleSyntax": false,
	}}}
}

// swcJest compiles the tests with @swc/jest, the compiler a workspace converted
// from Nx often has.
func swcJest() (devDependencies map[string]string, transform []any) {
	return map[string]string{"@swc/core": "^1.13.5", "@swc/jest": "^0.2.39"}, []any{discovery.ShapeSwcJest, map[string]any{
		"jsc":    map[string]any{"parser": map[string]any{"syntax": "typescript"}, "target": "es2022"},
		"module": map[string]any{"type": "commonjs"},
	}}
}
