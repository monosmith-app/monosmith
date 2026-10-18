package astro

import (
	"fmt"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name in a stack.
const Name = "astro"

// DocsSite is the stack the plugin publishes.
const DocsSite = "docs-site"

// Astro is the astro plugin.
type Astro struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Astro { return &Astro{Version: version} }

// Describe implements serve.Plugin.
func (p *Astro) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	p.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: p.Version},
		Kind:      protocol.KindConvention,
		Languages: []string{"node"},
		Requires:  []protocol.PluginRef{protocol.RequiresLanguage("node")},
		Roles:     []protocol.Role{protocol.RoleScaffold, protocol.RoleToolchain},
		Provides:  []string{"build"},
		Stacks: []protocol.Stack{{
			Name:        DocsSite,
			Description: "a documentation site with Astro Starlight, checked and built by Astro, linted by eslint",
			Language:    "node",
			Scaffold:    Name,
			Toolchain:   []string{Name, "eslint"},
		}},
	}, nil
}

// Generate implements serve.Generator: the Starlight starter, an app whose
// build is a static site released by its tag.
func (p *Astro) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Kind != protocol.ProjectApp {
		return protocol.GenerateResult{}, fmt.Errorf("a docs site is an app: smith add app %s --stack %s", params.Name, DocsSite)
	}
	return scaffolding.GenerateFrom(p.root, params, scaffolding.TemplateSet{Slice: templates, DevDependencies: devDependencies})
}
