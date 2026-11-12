package functions

import (
	"fmt"
	"maps"
	"slices"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/typescript"
	"github.com/monosmith-app/monosmith/libs/plugin-node/choicedata"
	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/libs/plugin-node/tsconfigchain"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
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
		Kinds:     []protocol.ProjectKind{protocol.ProjectApp},
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
	addAppTasks(plan, params.Directory)
	res, err := scaffolding.GenerateFrom(p.root, params, templateSet(plan, testModule(params.Choices)))
	if err != nil {
		return res, err
	}
	res.Notes = append(res.Notes, nextSteps(plan.Applied)...)
	return res, nil
}

// addAppTasks gives the plan the tasks that build, start and package the app.
// They run smith's own commands and the tool of the host, so the app holds no
// script of smith's.
func addAppTasks(plan choicedata.Plan, root string) {
	host := ""
	for _, name := range plan.Applied {
		if _, ok := functionhost.ToolOf(name); ok {
			host = name
		}
	}
	trigger := ""
	if t := plan.Slots["trigger"]; len(t) > 0 {
		trigger = t[0]
	}
	plan.Tasks["build"] = functionhost.AssembleTask(host)
	plan.Tasks["start"] = functionhost.StartTask(host, root, trigger)
	plan.Tasks["package"] = functionhost.PackageTask(host, root)
}

// nextSteps are printed after a function app is made: how to run it, what the
// local tool is, and for Azure the deploy pipeline smith has.
func nextSteps(applied []string) []string {
	notes := []string{"Run it locally with `smith run start`. `smith test` needs no cloud tool."}
	for _, name := range applied {
		tool, ok := functionhost.ToolOf(name)
		if !ok {
			continue
		}
		if tool.Binary != "" {
			notes = append(notes, "The local run needs "+tool.Name+" ("+tool.Binary+"); `smith doctor` checks it. To get it: "+tool.Install+".")
		}
		if name == functionhost.Azure {
			notes = append(notes, "Copy local.settings.sample.json to local.settings.json (git ignores it) for your settings.")
		}
	}
	return notes
}

// testModule is the module the starter's test imports describe, expect and it
// from: the one of the test runner the choices name. Vitest is the default.
func testModule(choices []string) string {
	if slices.Contains(choices, "jest") {
		return "@jest/globals"
	}
	return "vitest"
}

// templateSet is the starter that a plan of choices describes: the greeting
// slice and the files of the plan. module is the test runner's module, which
// the starter's test imports.
func templateSet(plan choicedata.Plan, module string) scaffolding.TemplateSet {
	files := maps.Clone(typescript.GreetingSlice)
	for path, body := range plan.Files {
		files["[[.Dir]]/"+path] = body
	}
	return scaffolding.TemplateSet{
		Slice: files, Root: tsconfigchain.NodeRoot(), DevDependencies: plan.DevDependencies, Overrides: plan.Overrides,
		Slots: withSlot(plan.Slots, "test.module", module), Scripts: plan.Tasks, Dependencies: plan.Dependencies,
	}
}

// withSlot returns a copy of slots with one more slot of one value.
func withSlot(slots map[string][]string, name, value string) map[string][]string {
	out := maps.Clone(slots)
	if out == nil {
		out = map[string][]string{}
	}
	out[name] = []string{value}
	return out
}
