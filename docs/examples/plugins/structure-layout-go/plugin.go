package main

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// layout is a structure plugin: it implements Contribute (tasks) and Doctor
// (health findings). It has no Generate, so it adds nothing to new projects.
type layout struct{ root string }

// Describe answers "initialize".
func (l *layout) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	l.root = p.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin: protocol.Peer{Name: "widget-layout", Version: "0.1.0"},
		Kind:   protocol.KindConvention,
		Roles:  []protocol.Role{protocol.RoleStructure},
		// The languages it applies to. A convention that names none applies to
		// every language.
		Languages: []string{"widget"},
	}, nil
}

// Contribute answers "contribute": a layout-check task per project, and the
// project's lint (when it has one) waits for it.
func (l *layout) Contribute(p protocol.ContributeParams) (protocol.ContributeResult, error) {
	exe, err := os.Executable()
	if err != nil {
		return protocol.ContributeResult{}, err
	}
	out := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, project := range p.Projects {
		out.Projects[project.Name] = protocol.Contribution{
			Tasks: map[string]protocol.Task{
				"layout-check": {
					Command: []string{exe, "layout-check"},
					// No inputs listed: every file of the project is one, so a new
					// or renamed folder is a cache miss.
					Cache:       true,
					Description: "fail on folders named util, helpers, common, shared or misc",
				},
			},
			// DependsOn only touches a task the project has once every earlier
			// plugin has contributed; for a project without lint it is ignored.
			// Do not decide from project.Tasks here: it holds only what the
			// language plugin discovered, not what a toolchain plugin adds.
			DependsOn: map[string][]string{"lint": {"layout-check"}},
		}
	}
	return out, nil
}

// Doctor answers "doctor": every project needs a README.md.
func (l *layout) Doctor(p protocol.DoctorParams) (protocol.DoctorResult, error) {
	result := protocol.DoctorResult{Findings: []protocol.Finding{}}
	for _, project := range p.Projects {
		readme := filepath.Join(l.root, filepath.FromSlash(project.Root), "README.md")
		_, err := os.Stat(readme)
		f := protocol.Finding{Check: project.Name + " has a README.md", OK: err == nil}
		if err != nil {
			f.Detail = project.Root + "/README.md does not exist"
			f.Remedy = "write " + project.Root + "/README.md: what the project is for"
			f.Warning = true // reported, but does not fail smith doctor
		}
		result.Findings = append(result.Findings, f)
	}
	return result, nil
}
