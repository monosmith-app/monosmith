package scaffolding

import (
	"cmp"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-python/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-python/functions"
	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectnaming"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolpin"
)

type templateData struct {
	Dir, Name, Distribution, Import, Kind, Stack string
	// Dependencies is the dependencies list of pyproject.toml, as TOML.
	Dependencies string
	// SourceRoot is the folder setuptools finds the packages in: "src", or
	// "." for a folder of sources without one (ManifestOnly).
	SourceRoot string
	// Data is what a function app (ADR 0072) adds; zero for any other project.
	functions.Data
}

// Sources are the source files a Python project holds, which the plugin
// recognises in a folder without a pyproject.toml (InitializeResult.Sources).
var Sources = []string{"*.py"}

// Generate scaffolds one Python project, creating the workspace venv and its
// tooling the first time, and installs the project in editable mode. With
// params.ManifestOnly it writes only pyproject.toml into the folder that holds
// the sources, and makes no venv and installs nothing.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	distribution := cmp.Or(params.PackageName, projectnaming.Name(params.Workspace, projectnaming.Python, params.Name))
	data := templateData{Dir: params.Directory, Name: params.Name, Distribution: distribution, Import: discovery.ImportName(distribution), Kind: string(params.Kind), Stack: params.Stack, Dependencies: "[]", SourceRoot: "src"}
	if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory), "src")); err != nil && params.ManifestOnly {
		// The sources stay where they are: at the top of the folder.
		data.SourceRoot = "."
	}
	plan, isFunction := functions.PlanOf(params.Choices)
	if isFunction {
		if params.Kind != protocol.ProjectApp {
			return protocol.GenerateResult{}, fmt.Errorf("a function app is an app: smith add app %s --lang python --with %s", params.Name, plan.Host)
		}
		if plan.PythonVersion == "" {
			plan.PythonVersion = toolpin.Version(workspaceRoot, "python")
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
	if params.ManifestOnly {
		files, err := templating.Write(workspaceRoot, templating.ManifestOnly(templates, Sources), data)
		return protocol.GenerateResult{Files: files}, err
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
