package functions

import (
	"fmt"
	"maps"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/typescript"
	"github.com/monosmith-app/monosmith/libs/plugin-node/choicedata"
	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name in a stack.
const Name = "functions"

// Functions is the functions plugin. A host choice writes the function app, a
// trigger and a runtime fill the slots the host's templates read.
type Functions struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Functions { return &Functions{Version: version} }

// Describe implements serve.Plugin.
func (p *Functions) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	p.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: p.Version},
		Kind:      protocol.KindConvention,
		Languages: []string{"node"},
		Requires:  []protocol.PluginRef{protocol.RequiresLanguage("node")},
		Roles:     []protocol.Role{protocol.RoleScaffold},
		Choices:   functionChoices.Choices(),
	}, nil
}

// Generate implements serve.Generator: the function app that the stack's
// choices describe. A function app is an app.
//
// It returns an error for a kind other than app, and for params that carry no
// choice of this plugin.
func (p *Functions) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Kind != protocol.ProjectApp {
		return protocol.GenerateResult{}, fmt.Errorf("a function app is an app: smith add app %s --with azure-functions", params.Name)
	}
	plan, err := functionChoices.Plan(params.Choices)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	if len(plan.Applied) == 0 {
		return protocol.GenerateResult{}, fmt.Errorf("the functions plugin writes projects from the choices azure-functions, aws-lambda and gcp-functions, not the stack %q: smith add app <name> --lang node --with azure-functions", params.Stack)
	}
	return scaffolding.GenerateFrom(p.root, params, templateSet(plan))
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
