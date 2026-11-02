package electron

import (
	"fmt"
	"maps"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/typescript"
	"github.com/monosmith-app/monosmith/libs/plugin-node/choicedata"
	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name in a stack.
const Name = "electron"

// Electron is the electron plugin.
type Electron struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Electron { return &Electron{Version: version} }

// Describe implements serve.Plugin.
func (p *Electron) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	p.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: p.Version},
		Kind:      protocol.KindConvention,
		Languages: []string{"node"},
		Requires:  []protocol.PluginRef{protocol.RequiresLanguage("node")},
		Roles:     []protocol.Role{protocol.RoleScaffold},
		Choices:   electronChoices.Choices(),
	}, nil
}

// Generate implements serve.Generator: the Electron app that the stack's
// choices describe. A desktop app is an app.
//
// It returns an error for a kind other than app, and for params that carry no
// choice of this plugin.
func (p *Electron) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Kind != protocol.ProjectApp {
		return protocol.GenerateResult{}, fmt.Errorf("an Electron project is an app: smith add app %s --lang node --with electron", params.Name)
	}
	plan, err := electronChoices.Plan(params.Choices)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	if len(plan.Applied) == 0 {
		return protocol.GenerateResult{}, fmt.Errorf("the electron plugin writes projects from the choice electron, not the stack %q: smith add app <name> --lang node --with electron", params.Stack)
	}
	res, err := scaffolding.GenerateFrom(p.root, params, templateSet(plan))
	if err != nil {
		return res, err
	}
	res.Notes = append(res.Notes,
		"Run it with `smith run start` (it opens a window). `smith run package` makes the app folder with Electron Forge; it does not sign it.",
		"The first install downloads the Electron binary, so it needs a network.")
	return res, nil
}

// templateSet is the starter that a plan of choices describes: the greeting
// slice and the files of the plan.
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
