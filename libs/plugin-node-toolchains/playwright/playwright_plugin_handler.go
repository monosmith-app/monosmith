package playwright

import (
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tooling"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name in a stack's toolchain list and in smith.toml
// ([conventions.playwright]).
const Name = "playwright"

// Stack is the stack the plugin publishes.
const Stack = "node-vitest-playwright"

// Playwright is the playwright toolchain plugin.
type Playwright struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Playwright { return &Playwright{Version: version} }

// Describe implements serve.Plugin.
func (p *Playwright) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	p.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: p.Version},
		Kind:      protocol.KindConvention,
		Languages: []string{"node"},
		Roles:     []protocol.Role{protocol.RoleToolchain},
		Provides:  []string{TaskName},
		Stacks: []protocol.Stack{{
			Name:        Stack,
			Description: "TypeScript on Node, linted by eslint, unit-tested with Vitest, end-to-end tested with Playwright",
			Language:    "node",
			Scaffold:    "typescript",
			Toolchain:   []string{"typescript", "eslint", "vitest", Name},
		}},
	}, nil
}

// Contribute implements serve.Contributor.
func (p *Playwright) Contribute(params protocol.ContributeParams) (protocol.ContributeResult, error) {
	return Contribute(p.root, params), nil
}

// Doctor implements serve.DoctorContributor.
func (p *Playwright) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return Doctor(p.root, params), nil
}

// Generate implements serve.Generator: it runs after the node plugin has
// generated the project. Only a stack that names playwright gets a suite:
// as a workspace convention it runs the e2e task of projects that have a
// suite, and a new project with no UI must not get a browser test (#314).
func (p *Playwright) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Convention {
		return protocol.GenerateResult{}, nil
	}
	return tooling.Apply(p.root, params, setup(p.root, params))
}
