package angular

import (
	"fmt"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name in a stack.
const Name = "angular"

// Angular is the angular plugin.
type Angular struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Angular { return &Angular{Version: version} }

// Describe implements serve.Plugin.
func (p *Angular) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	p.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: p.Version},
		Kind:      protocol.KindConvention,
		Languages: []string{"node"},
		Requires:  []protocol.PluginRef{protocol.RequiresLanguage("node")},
		Roles:     []protocol.Role{protocol.RoleScaffold, protocol.RoleToolchain},
		Provides:  []string{"build", "typecheck"},
		Stacks: []protocol.Stack{{
			Name:        Name,
			Description: "an Angular app in TypeScript, built with the Angular CLI, linted by eslint, tested with Vitest",
			Language:    "node",
			Scaffold:    Name,
			Toolchain:   []string{Name, "eslint", "vitest"},
		}},
	}, nil
}

// Generate implements serve.Generator: the Angular starter, an app whose
// build is a static site released by its tag.
func (p *Angular) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Kind != protocol.ProjectApp {
		return protocol.GenerateResult{}, fmt.Errorf("an angular project is an app: smith add app %s --stack %s", params.Name, Name)
	}
	return scaffolding.GenerateFrom(p.root, params, scaffolding.TemplateSet{Slice: templates, DevDependencies: devDependencies})
}
