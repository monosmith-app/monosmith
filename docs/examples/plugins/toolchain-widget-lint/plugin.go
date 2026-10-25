package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// pluginName is what users write in a stack's toolchain list and under
// [projects.<name>.settings.<plugin>].
const pluginName = "widget-lint"

// lint is a toolchain plugin: a convention-kind plugin with the toolchain
// role. The methods it has are its capabilities: Contribute and Generate.
type lint struct{ root string }

// Describe answers "initialize".
func (l *lint) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	l.root = p.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: pluginName, Version: "0.1.0"},
		Kind:      protocol.KindConvention,
		Roles:     []protocol.Role{protocol.RoleToolchain},
		Languages: []string{"widget"},                                        // the projects it applies to
		Provides:  []string{"lint"},                                          // the slot it fills: no other toolchain of a stack may
		Requires:  []protocol.PluginRef{protocol.RequiresLanguage("widget")}, // it is no use without the language
		// A ready-made stack: widget projects with this linter.
		Stacks: []protocol.Stack{{
			Name:        "widget-linted",
			Description: "widget projects that must describe and license themselves",
			Language:    "widget",
			Toolchain:   []string{pluginName},
		}},
	}, nil
}

// requiredFields reads the plugin's own settings for a project:
//
//	[projects.<name>.settings.widget-lint]
//	require = ["description", "license", "version"]
func requiredFields(p protocol.Project) []string {
	raw, ok := p.Settings[pluginName]["require"].([]any)
	if !ok {
		return defaultRequire
	}
	var fields []string
	for _, v := range raw {
		if s, ok := v.(string); ok && s != "" {
			fields = append(fields, s)
		}
	}
	if len(fields) == 0 {
		return defaultRequire
	}
	return fields
}

// Contribute answers "contribute" with the lint task of each project. A task
// the project already has is kept, so use Replace to take the slot over.
func (l *lint) Contribute(p protocol.ContributeParams) (protocol.ContributeResult, error) {
	exe, err := os.Executable()
	if err != nil {
		return protocol.ContributeResult{}, err
	}
	out := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, project := range p.Projects {
		out.Projects[project.Name] = protocol.Contribution{
			Tasks: map[string]protocol.Task{
				"lint": {
					Command:     []string{exe, "lint", strings.Join(requiredFields(project), ",")},
					Inputs:      []string{"widget.json"},
					Cache:       true,
					Description: "check that widget.json sets " + strings.Join(requiredFields(project), ", "),
				},
			},
		}
	}
	return out, nil
}

// Generate answers "generate": it runs after the language plugin wrote the
// project, and adds what the toolchain needs. Never overwrite: write a file
// only when it is absent.
func (l *lint) Generate(p protocol.GenerateParams) (protocol.GenerateResult, error) {
	if p.Convention {
		return protocol.GenerateResult{Files: []string{}}, nil // a convention adds only what every project needs
	}
	rel := path.Join(p.Directory, "WIDGET-LINT.md")
	full := filepath.Join(l.root, filepath.FromSlash(rel))
	if _, err := os.Stat(full); err == nil {
		return protocol.GenerateResult{Files: []string{}}, nil
	}
	body := fmt.Sprintf("# %s\n\nlinted by `%s`: `smith run lint -p %s`.\n", p.Name, pluginName, p.Name)
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		return protocol.GenerateResult{}, err
	}
	return protocol.GenerateResult{Files: []string{rel}}, nil
}
