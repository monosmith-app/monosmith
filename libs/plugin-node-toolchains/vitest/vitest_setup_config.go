package vitest

import (
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tooling"
	"github.com/monosmith-app/monosmith/libs/plugin-node/discovery"
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
	return &tooling.Toolchain{Name: Name, Version: version, Provides: []string{"test"}, Stacks: stacks, Follows: []string{discovery.ShapeTestRunner}, Setup: setup}
}

func setup(workspaceRoot string, params protocol.GenerateParams) tooling.Setup {
	return tooling.Setup{
		DevDependencies: map[string]string{"vitest": "^5.0.5"},
		Files:           tooling.StarterSpec(workspaceRoot, params, "vitest"),
		Scripts:         map[string]string{"test": "vitest run"},
	}
}
