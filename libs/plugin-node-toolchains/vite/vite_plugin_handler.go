package vite

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name in a stack.
const Name = "vite"

// apps are the starters, by stack name.
var apps = map[string]webApp{"react": react, "react-p5": reactP5, "react-tauri": react, "vue": vue, "svelte": svelte, "html": html}

var descriptions = map[string]string{
	"react":       "a React app in TypeScript, built with Vite, linted by eslint, tested with Vitest",
	"react-p5":    "a React app drawing with p5, built with Vite, linted by eslint, tested with Vitest",
	"react-tauri": "a React app in TypeScript, built with Vite, linted by eslint, tested with Vitest, packaged as a desktop app with Tauri",
	"vue":         "a Vue app in TypeScript, built with Vite, linted by eslint, tested with Vitest",
	"svelte":      "a Svelte app in TypeScript, built with Vite, linted by eslint, tested with Vitest",
	"html":        "a page in TypeScript with no framework, built with Vite, linted by eslint, tested with Vitest",
}

// Vite is the vite plugin.
type Vite struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Vite { return &Vite{Version: version} }

// Describe implements serve.Plugin.
func (p *Vite) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	p.root = params.WorkspaceRoot
	var stacks []protocol.Stack
	for _, name := range slices.Sorted(maps.Keys(apps)) {
		stacks = append(stacks, protocol.Stack{
			Name: name, Description: descriptions[name], Language: "node",
			Scaffold: Name, Toolchain: toolchain(name),
		})
	}
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: p.Version},
		Kind:      protocol.KindConvention,
		Languages: []string{"node"},
		Requires:  []protocol.PluginRef{protocol.RequiresLanguage("node")},
		Roles:     []protocol.Role{protocol.RoleScaffold, protocol.RoleToolchain},
		Provides:  []string{"build", "typecheck"},
		Stacks:    stacks,
	}, nil
}

// toolchain is what a stack's project is built with: tauri joins the stacks that
// package the app as a desktop app.
func toolchain(stack string) []string {
	tc := []string{Name, "eslint", "vitest"}
	if strings.HasSuffix(stack, "-tauri") {
		tc = append(tc, "tauri")
	}
	return tc
}

// Generate implements serve.Generator: the starter of the project's stack.
// A browser app is an app; its build is a static site, released by its tag.
func (p *Vite) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	app, ok := apps[params.Stack]
	if !ok {
		return protocol.GenerateResult{}, fmt.Errorf("the vite plugin scaffolds the %v stacks, not %q", slices.Sorted(maps.Keys(apps)), params.Stack)
	}
	if params.Kind != protocol.ProjectApp {
		return protocol.GenerateResult{}, fmt.Errorf("a %s project is an app: smith add app %s --stack %s", params.Stack, params.Name, params.Stack)
	}
	dev := maps.Clone(devDependencies)
	maps.Copy(dev, app.DevDependencies)
	return scaffolding.GenerateFrom(p.root, params, scaffolding.TemplateSet{Slice: app.templates(), DevDependencies: dev, Overrides: app.Overrides})
}
