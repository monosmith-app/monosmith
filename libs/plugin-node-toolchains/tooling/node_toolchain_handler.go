package tooling

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// Toolchain serves one Node toolchain plugin over the plugin protocol: when a
// Node project is generated, it applies the Setup the plugin describes.
type Toolchain struct {
	Name    string
	Version string
	// Provides are the tasks the plugin owns ("lint", "test").
	Provides []string
	// Stacks are the stacks the plugin publishes.
	Stacks []protocol.Stack
	// Follows are the shape keys the plugin applies (InitializeResult.Follows).
	Follows []string
	// Setup returns what to add to the project params names.
	Setup func(workspaceRoot string, params protocol.GenerateParams) Setup
	root  string
}

// Describe implements serve.Plugin.
func (c *Toolchain) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	c.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: c.Name, Version: c.Version},
		Kind:      protocol.KindConvention,
		Languages: []string{"node"},
		Requires:  []protocol.PluginRef{protocol.RequiresLanguage("node")},
		Roles:     []protocol.Role{protocol.RoleToolchain},
		Provides:  c.Provides,
		Stacks:    c.Stacks,
		Follows:   c.Follows,
	}, nil
}

// Generate implements serve.Generator: it runs after the node plugin has
// generated the project.
func (c *Toolchain) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return Apply(c.root, params, c.Setup(c.root, params))
}
