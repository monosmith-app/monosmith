package main

import (
	"fmt"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

// stackName is the stack the plugin publishes and the name users give to
// `smith add app shop --stack widget-service`.
const stackName = "widget-service"

// defaultPort is used when the recipe (or nobody) sets no port.
const defaultPort = 8080

// scaffold implements Generate only: that is what a scaffold is.
type scaffold struct{ root string }

// Describe answers "initialize".
func (s *scaffold) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	s.root = p.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: "widget-service", Version: "0.1.0"},
		Kind:      protocol.KindConvention,
		Roles:     []protocol.Role{protocol.RoleScaffold},
		Languages: []string{"widget"},
		Requires:  []protocol.PluginRef{protocol.RequiresLanguage("widget")},
		Stacks: []protocol.Stack{{
			Name:        stackName,
			Description: "a widget service: an app with a port and a handler",
			Language:    "widget",
			Scaffold:    "widget-service",
			Toolchain:   []string{"widget-lint"},
		}},
	}, nil
}

// portOf reads the plugin's own settings for the new project, which smith
// takes from the recipe's settings.widget-service table:
//
//	[settings.widget-service]
//	port = 9000
func portOf(settings map[string]any) (int, error) {
	raw, ok := settings["port"]
	if !ok {
		return defaultPort, nil
	}
	// JSON numbers arrive as float64.
	port, ok := raw.(float64)
	if !ok || port < 1 || port > 65535 || port != float64(int(port)) {
		return 0, fmt.Errorf("settings.widget-service.port must be a whole number from 1 to 65535, not %v", raw)
	}
	return int(port), nil
}

// templates are the starter files. Targets and bodies are text/template
// with [[ ]] delimiters, so a body can contain {{ }} untouched.
var templates = map[string]string{
	"[[.Dir]]/widget.json": `{
  "name": "[[.Name]]",
  "version": "0.0.0",
  "kind": "app",
  "description": "The [[.Name]] service of the [[.Workspace]] workspace.",
  "license": "MIT",
  "stack": "[[.Stack]]"
}
`,
	"[[.Dir]]/service.json": `{
  "port": [[.Port]],
  "handler": "src/handler.txt"
}
`,
	"[[.Dir]]/src/handler.txt": "hello from [[.Name]]\n",
	"[[.Dir]]/README.md":       "# [[.Name]]\n\nA widget service on port [[.Port]].\n",
}

// Generate answers "generate". The scaffold takes the language plugin's
// place, so it must write the manifest, and record params.Stack in it.
func (s *scaffold) Generate(p protocol.GenerateParams) (protocol.GenerateResult, error) {
	if p.Kind != protocol.ProjectApp {
		return protocol.GenerateResult{}, fmt.Errorf("a widget service is an app: smith add app %s --stack %s", p.Name, stackName)
	}
	port, err := portOf(p.Settings)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	files, err := templating.Write(s.root, templates, map[string]any{
		"Dir": p.Directory, "Name": p.Name, "Stack": p.Stack, "Port": port,
		"Workspace": p.Workspace.Name,
	})
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	return protocol.GenerateResult{Files: files, Notes: []string{fmt.Sprintf("%s listens on port %d: edit %s/service.json", p.Name, port, p.Directory)}}, nil
}
