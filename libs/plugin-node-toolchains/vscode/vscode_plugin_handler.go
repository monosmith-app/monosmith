package vscode

import (
	"fmt"
	"maps"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name in a stack.
const Name = "vscode"

// Stack is the stack the plugin publishes.
const Stack = "vscode-extension"

// VSCode is the vscode plugin.
type VSCode struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *VSCode { return &VSCode{Version: version} }

// Describe implements serve.Plugin.
func (p *VSCode) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	p.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: p.Version},
		Kind:      protocol.KindConvention,
		Languages: []string{"node"},
		Requires:  []protocol.PluginRef{protocol.RequiresLanguage("node")},
		Roles:     []protocol.Role{protocol.RoleScaffold, protocol.RoleToolchain, protocol.RoleDeploy},
		Provides:  []string{"build", "typecheck"},
		Stacks: []protocol.Stack{{
			Name:        Stack,
			Description: "a VS Code extension in TypeScript, bundled by esbuild, linted by eslint, tested with Vitest, packed and published with vsce",
			Language:    "node",
			Scaffold:    Name,
			Toolchain:   []string{Name, "eslint", "vitest"},
		}},
	}, nil
}

// Generate implements serve.Generator: the extension starter. A package is
// published to the Marketplace; an app is an extension kept to the
// workspace's own .vsix. The publisher is the workspace's npm scope, else
// its name.
func (p *VSCode) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Kind == protocol.ProjectLib {
		return protocol.GenerateResult{}, fmt.Errorf("an extension is a package (published to the Marketplace) or an app: smith add package %s --stack %s", params.Name, Stack)
	}
	publisher := strings.TrimPrefix(params.Workspace.Scope, "@")
	if publisher == "" {
		publisher = params.Workspace.Name
	}
	files := maps.Clone(templates)
	manifest := "[[.Dir]]/package.json"
	files[manifest] = strings.NewReplacer(publisherToken, publisher, kindToken, string(params.Kind)).Replace(files[manifest])
	return scaffolding.GenerateFrom(p.root, params, scaffolding.TemplateSet{Slice: files, DevDependencies: devDependencies})
}

// Contribute implements serve.Contributor: the package task of an extension app.
func (p *VSCode) Contribute(params protocol.ContributeParams) (protocol.ContributeResult, error) {
	return contributeExtensions(p.root, params), nil
}

// Publish implements serve.Publisher: a release of an extension app.
func (p *VSCode) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return publishExtension(p.root, params)
}

// CI implements serve.CIContributor: the tokens an extension app's release reads.
func (p *VSCode) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return ciExtensions(p.root, params), nil
}
