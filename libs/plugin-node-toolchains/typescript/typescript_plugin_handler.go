package typescript

import (
	"maps"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-node/discovery"
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
		Follows:   []string{discovery.ShapeBuildTool, discovery.ShapeExportsTarget, discovery.ShapeBuildOutput, discovery.ShapeExportsMap, discovery.ShapeTsconfigFiles},
		Stacks: []protocol.Stack{{
			Name:        "node-ts",
			Description: "TypeScript on Node, built and type-checked with tsc; no lint or tests",
			Language:    "node",
			Scaffold:    Name,
			Toolchain:   []string{Name},
		}},
	}, nil
}

// Contribute implements serve.Contributor: the plugin adds no tasks. It
// publishes the default source root of the projects it scaffolds as the shared
// fact source.roots, so a tool that needs it reads it from the facts.
func (p *TypeScript) Contribute(params protocol.ContributeParams) (protocol.ContributeResult, error) {
	res := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, project := range params.Projects {
		res.Projects[project.Name] = protocol.Contribution{Facts: map[string]any{protocol.FactSourceRoots: []string{"src"}}}
	}
	return res, nil
}

// Generate implements serve.Generator: the TypeScript starter.
func (p *TypeScript) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	set := setFor(params.Shape)
	if params.Shape[discovery.ShapeBuildTool] == "esbuild" {
		// A project that builds with esbuild needs it in the workspace.
		set.DevDependencies = maps.Clone(set.DevDependencies)
		set.DevDependencies["esbuild"] = esbuildVersion
	}
	return scaffolding.GenerateFrom(p.root, params, set)
}

// Derive implements serve.Deriver: the tsconfig.json of the projects the plugin
// scaffolded, rendered from its own template (ADR 0049). An app made from
// choices has a vite.config.ts and belongs to the vite plugin.
func (p *TypeScript) Derive(params protocol.DeriveParams) (protocol.DeriveResult, error) {
	return scaffolding.FileRenderer{
		ID: "typescript.tsconfig", Version: 1, File: "tsconfig.json",
		Templates: func(gp protocol.GenerateParams) (scaffolding.TemplateSet, bool) {
			_, err := os.Stat(filepath.Join(p.root, filepath.FromSlash(gp.Directory), "vite.config.ts"))
			return templates, err != nil
		},
	}.Derive(p.root, params)
}
