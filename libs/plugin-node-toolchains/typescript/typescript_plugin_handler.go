package typescript

import (
	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name in a stack.
const Name = "typescript"

// TypeScript is the typescript plugin.
type TypeScript struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *TypeScript { return &TypeScript{Version: version} }

// Describe implements serve.Plugin.
func (p *TypeScript) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	p.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: p.Version},
		Kind:      protocol.KindConvention,
		Languages: []string{"node"},
		Requires:  []protocol.PluginRef{protocol.RequiresLanguage("node")},
		Roles:     []protocol.Role{protocol.RoleScaffold, protocol.RoleToolchain},
		Provides:  []string{"build", "typecheck"},
		Stacks: []protocol.Stack{{
			Name:        "node-ts",
			Description: "TypeScript on Node, built and type-checked with tsc; no lint or tests",
			Language:    "node",
			Scaffold:    Name,
			Toolchain:   []string{Name},
		}},
	}, nil
}

// Generate implements serve.Generator: the TypeScript starter.
func (p *TypeScript) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.GenerateFrom(p.root, params, templates)
}
