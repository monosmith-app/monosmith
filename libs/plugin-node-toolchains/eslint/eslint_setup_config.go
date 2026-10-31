package eslint

import (
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tooling"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name in a stack's toolchain list.
const Name = "eslint"

const config = `// Lint AND formatting for every Node project: @monosmith/eslint-config carries the
// TypeScript rules, JavaScript Standard Style and the vertical-slice rules.
// Add your own blocks after the spread; later blocks win.
import monosmith from '@monosmith/eslint-config'

export default [
  ...monosmith(),
]
`

// frontEndConfig is the config of a React stack: the same, with the front-end slice
// roles that a .ts file does not get unasked (.hook .content .fixture .style .mock).
const frontEndConfig = `// Lint AND formatting for every Node project: @monosmith/eslint-config carries the
// TypeScript rules, JavaScript Standard Style and the vertical-slice rules.
// This is a front-end workspace, so the front-end slice roles are on.
// Add your own blocks after the spread; later blocks win.
import monosmith from '@monosmith/eslint-config'

export default [
  ...monosmith({ verticalSlices: { roles: ['hook', 'content', 'fixture', 'style', 'mock'] } }),
]
`

// configFor returns the root eslint config text for a stack: React stacks
// (react, react-p5, react-tauri) enable the front-end slice roles.
func configFor(stack string) string {
	if strings.HasPrefix(stack, "react") {
		return frontEndConfig
	}
	return config
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *tooling.Toolchain {
	return &tooling.Toolchain{Name: Name, Version: version, Provides: []string{"lint"}, Setup: setup}
}

func setup(_ string, params protocol.GenerateParams) tooling.Setup {
	return tooling.Setup{
		RootFiles: map[string]string{"eslint.config.mjs": configFor(params.Stack)},
		DevDependencies: map[string]string{
			"@monosmith/eslint-config": "^1.0.0",
			"eslint":                   "^10.12.0",
		},
		// eslint-plugin-jsx-a11y (a dependency of @monosmith/eslint-config)
		// still declares eslint 9 as its newest peer; it works with 10, and
		// without this npm prints a page of peer-conflict warnings on the
		// first install.
		RootFields: map[string]any{"overrides": map[string]any{"eslint-plugin-jsx-a11y": map[string]any{"eslint": "$eslint"}}},
		Scripts:    map[string]string{"lint": "eslint ."},
	}
}
