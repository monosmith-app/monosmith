package plugin

import (
	"slices"

	"github.com/monosmith-app/monosmith/libs/plugin-vertical-slices/languages"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the convention's name in smith.toml ([conventions.vertical-slices]).
const Name = "vertical-slices"

// VerticalSlices is the vertical-slices convention plugin.
type VerticalSlices struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *VerticalSlices { return &VerticalSlices{Version: version} }

// Describe implements serve.Plugin.
func (v *VerticalSlices) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	v.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin: protocol.Peer{Name: Name, Version: v.Version},
		Kind:   protocol.KindConvention,
		// Node is listed for doctor: its rules are @monosmith/eslint-config's,
		// not a task, so Contribute leaves it alone.
		Languages: slices.Concat(languages.Names(), nodeLanguages),
	}, nil
}

// Doctor implements serve.DoctorContributor: a Node project's slice rules run
// only when the root ESLint config loads @monosmith/eslint-config.
func (v *VerticalSlices) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return protocol.DoctorResult{Findings: append([]protocol.Finding{}, checkNodeESLint(v.root, params.Projects)...)}, nil
}

// Contribute implements serve.Contributor: a slice-check task for every
// project, run before lint.
func (v *VerticalSlices) Contribute(params protocol.ContributeParams) (protocol.ContributeResult, error) {
	res := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, p := range params.Projects {
		rules, ok := languages.For(p.Language)
		if !ok {
			continue
		}
		c := protocol.Contribution{Tasks: map[string]protocol.Task{TaskName: sliceCheck(rules)}}
		if _, ok := p.Tasks["lint"]; ok {
			c.DependsOn = map[string][]string{"lint": {TaskName}}
		}
		res.Projects[p.Name] = c
	}
	return res, nil
}
