package vite

import (
	"fmt"
	"maps"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/typescript"
	"github.com/monosmith-app/monosmith/libs/plugin-node/choicedata"
	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name in a stack.
const Name = "vite"

// Vite is the vite plugin. It builds browser apps from choices: a framework
// writes the app, a style adds to it. The choices are data (see
// web_choices_config.go and ADR 0045).
type Vite struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Vite { return &Vite{Version: version} }

// Describe implements serve.Plugin.
func (p *Vite) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	p.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: p.Version},
		Kind:      protocol.KindConvention,
		Languages: []string{"node"},
		Requires:  []protocol.PluginRef{protocol.RequiresLanguage("node")},
		Roles:     []protocol.Role{protocol.RoleScaffold, protocol.RoleToolchain},
		Provides:  []string{"build", "typecheck"},
		Stacks:    webChoices.Stacks(),
		Choices:   webChoices.Choices(),
	}, nil
}

// Generate implements serve.Generator: the starter that the stack's choices
// describe. A browser app is an app; its build is a static site, released by
// its tag.
//
// It returns an error for a kind other than app, and for params that carry no
// choice of this plugin.
func (p *Vite) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Kind != protocol.ProjectApp {
		return protocol.GenerateResult{}, fmt.Errorf("a %s project is an app: smith add app %s --stack %s", params.Stack, params.Name, params.Stack)
	}
	plan, err := webChoices.Plan(params.Choices)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	if len(plan.Applied) == 0 {
		var offered []string
		for _, c := range webChoices.Choices() {
			if c.Plays(protocol.RoleScaffold) {
				offered = append(offered, c.Name)
			}
		}
		return protocol.GenerateResult{}, fmt.Errorf("the vite plugin writes projects from the choices %v, not the stack %q: smith add app <name> --with react", offered, params.Stack)
	}
	return scaffolding.GenerateFrom(p.root, params, templateSet(plan))
}

// templateSet is the starter that a plan of choices describes.
func templateSet(plan choicedata.Plan) scaffolding.TemplateSet {
	files := maps.Clone(typescript.GreetingSlice)
	for path, body := range plan.Files {
		files["[[.Dir]]/"+path] = body
	}
	return scaffolding.TemplateSet{
		Slice: files, DevDependencies: plan.DevDependencies, Overrides: plan.Overrides,
		Slots: plan.Slots, Scripts: plan.Tasks, Dependencies: plan.Dependencies,
	}
}

// Derive implements serve.Deriver: the tsconfig.json of the apps that the
// plugin made from choices, rendered from the same choices (ADR 0049).
func (p *Vite) Derive(params protocol.DeriveParams) (protocol.DeriveResult, error) {
	return scaffolding.FileRenderer{
		ID: "vite.tsconfig", Version: 1, File: "tsconfig.json",
		Templates: func(gp protocol.GenerateParams) (scaffolding.TemplateSet, bool) {
			choices := scaffolding.StackChoices(p.root, gp.Stack)
			for _, st := range webChoices.Stacks() {
				if st.Name == gp.Stack {
					choices = st.Choices
				}
			}
			plan, err := webChoices.Plan(choices)
			if err != nil || len(plan.Applied) == 0 {
				return scaffolding.TemplateSet{}, false
			}
			return templateSet(plan), true
		},
	}.Derive(p.root, params)
}
