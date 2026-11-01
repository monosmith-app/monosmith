package jest

import (
	"encoding/json"
	"maps"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tooling"
	"github.com/monosmith-app/monosmith/libs/plugin-node/discovery"
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
	return &tooling.Toolchain{Name: Name, Version: version, Provides: []string{"test"}, Stacks: stacks, Choices: choices.Choices(), Follows: []string{discovery.ShapeTestRunner, discovery.ShapeJestConfig}, Setup: setup}
}

func setup(workspaceRoot string, params protocol.GenerateParams) tooling.Setup {
	// A module.aliases fact adds its aliases to the mapper and, for ts-jest, to
	// the compiler options it type-checks with; without one, neither changes.
	aliases := modulealiases.FromFacts(params.Facts)
	devDependencies, transform := tsJest(aliases)
	if params.Shape[discovery.ShapeTestRunner] == discovery.ShapeSwcJest {
		devDependencies, transform = swcJest()
	}
	devDependencies["@jest/globals"] = "^30.5.2"
	devDependencies["jest"] = "^30.5.2"
	mapper := map[string]string{`^(\.{1,2}/.*)\.js$`: "$1"}
	maps.Copy(mapper, aliases.JestModuleNameMapper())
	options := map[string]any{
		"testEnvironment":  "node",
		"roots":            []string{"<rootDir>/src"},
		"moduleNameMapper": mapper,
		"transform":        map[string]any{`^.+\.ts$`: transform},
	}
	setup := tooling.Setup{
		DevDependencies: devDependencies,
		Files:           tooling.StarterSpec(workspaceRoot, params, "@jest/globals"),
		Scripts:         map[string]string{"test": "jest"},
		// Jest runs CommonJS: the compiler turns TypeScript into it, and a
		// relative import rewritten to .js finds the .ts source.
		Fields: map[string]any{"jest": options},
	}
	if params.Shape[discovery.ShapeJestConfig] == discovery.ShapeConfigFile {
		// The siblings keep the configuration in a jest.config.cts, which Jest
		// loads through ts-node.
		devDependencies["ts-node"] = "^10.9.2"
		setup.Fields = nil
		if setup.Files == nil {
			setup.Files = map[string]string{}
		}
		maps.Copy(setup.Files, configFiles(params.Name, options, transform))
	}
	return setup
}

// configFiles are the jest.config.cts of a project and, for a transform that
// takes options (@swc/jest), the .spec.swcrc it reads them from.
func configFiles(name string, options map[string]any, transform []any) map[string]string {
	rest := maps.Clone(options)
	rest["displayName"] = name
	if transform[0] != discovery.ShapeSwcJest {
		body, _ := json.MarshalIndent(rest, "", "  ")
		return map[string]string{"jest.config.cts": "module.exports = " + string(body) + "\n"}
	}
	delete(rest, "transform")
	body, _ := json.MarshalIndent(rest, "", "  ")
	swc, _ := json.MarshalIndent(transform[1], "", "  ")
	return map[string]string{
		".spec.swcrc": string(swc) + "\n",
		"jest.config.cts": `/* eslint-disable */
const { readFileSync } = require('fs')

// Reading the SWC compilation config for the spec files
const swcJestConfig = JSON.parse(readFileSync(` + "`${__dirname}/.spec.swcrc`" + `, 'utf-8'))

// Disable .swcrc look-up by SWC core because we're passing in swcJestConfig ourselves
swcJestConfig.swcrc = false

module.exports = ` + string(body[:len(body)-2]) + `,
  "transform": { "^.+\\.ts$": ["@swc/jest", swcJestConfig] }
}
`,
	}
}

// tsJest is the default compiler of the tests: ts-jest.
func tsJest(aliases modulealiases.Aliases) (devDependencies map[string]string, transform []any) {
	compiler := map[string]any{"module": "commonjs", "moduleResolution": "bundler", "verbatimModuleSyntax": false}
	if paths := aliases.CompilerPaths(); paths != nil {
		compiler["paths"] = paths
	}
	return map[string]string{"ts-jest": "^29.4.15"}, []any{"ts-jest", map[string]any{"tsconfig": compiler}}
}

// swcJest compiles the tests with @swc/jest, the compiler a workspace converted
// from Nx often has.
func swcJest() (devDependencies map[string]string, transform []any) {
	return map[string]string{"@swc/core": "^1.13.5", "@swc/jest": "^0.2.39"}, []any{discovery.ShapeSwcJest, map[string]any{
		"jsc":    map[string]any{"parser": map[string]any{"syntax": "typescript"}, "target": "es2022"},
		"module": map[string]any{"type": "commonjs"},
	}}
}
