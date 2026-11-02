package scaffolding

import (
	"fmt"
	"maps"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-python/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-python/functions"
	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

type templateData struct {
	Dir, Name, Distribution, Import, Kind, Stack string
	// Dependencies is the dependencies list of pyproject.toml, as TOML.
	Dependencies string
	// Data is what a function app (ADR 0072) adds; zero for any other project.
	functions.Data
}

// Generate scaffolds one Python project, creating the workspace venv and its
// tooling the first time, and installs the project in editable mode.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	distribution := distributionName(params.Workspace.Name, params.Name)
	data := templateData{Dir: params.Directory, Name: params.Name, Distribution: distribution, Import: discovery.ImportName(distribution), Kind: string(params.Kind), Stack: params.Stack, Dependencies: "[]"}
	plan, isFunction := functions.PlanOf(params.Choices)
	if isFunction {
		if params.Kind != protocol.ProjectApp {
			return protocol.GenerateResult{}, fmt.Errorf("a function app is an app: smith add app %s --lang python --with %s", params.Name, plan.Host)
		}
		if plan.PythonVersion == "" {
			plan.PythonVersion = venv.PythonVersion
		}
		data.Data, data.Dependencies = functions.DataOf(plan, protocol.BuildDirFromProject(params.Directory)), functions.Dependencies(plan.Host)
	}

	templates := maps.Clone(projectTemplates)
	switch params.Kind {
	case protocol.ProjectApp:
		if isFunction {
			maps.Copy(templates, functions.Templates(plan.Host))
		} else {
			maps.Copy(templates, appTemplates)
		}
	case protocol.ProjectPackage:
		maps.Copy(templates, packageTemplates)
	}
	files, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	created, err := venv.Ensure(workspaceRoot)
	if err != nil {
		return protocol.GenerateResult{Files: files}, err
	}
	files = append(files, created...)
	if err := venv.InstallEditable(workspaceRoot, params.Directory); err != nil {
		return protocol.GenerateResult{Files: files}, err
	}
	notes := []string{"Import it from another Python project as `import " + data.Import + "`."}
	var shared []string
	for _, f := range files {
		if f == venv.RequirementsFile {
			shared = append(shared, f)
		}
	}
	if len(shared) > 0 {
		notes = append(notes, "Created at the workspace root, shared by every Python project: "+strings.Join(shared, ", ")+" and .venv/ (the environment, ignored by git).")
	}
	if isFunction {
		notes = append(notes, functionNotes(plan.Host)...)
	}
	return protocol.GenerateResult{Files: files, Notes: notes}, nil
}

// distributionIdentifier lowercases a workspace name and turns each run of
// characters other than letters and digits (the dots of jato.auto.packages)
// into one dash.
func distributionIdentifier(name string) string {
	var b strings.Builder
	for _, c := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			b.WriteRune(c)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "-"):
			b.WriteByte('-')
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// distributionName prefixes the workspace name so packages are unique on a
// shared index: workspace "acme" and project "sdk" give "acme-sdk".
func distributionName(workspace, project string) string {
	ws := distributionIdentifier(workspace)
	if ws == "" || strings.HasPrefix(project, ws+"-") {
		return project
	}
	return ws + "-" + project
}
