package scaffolding

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/functions"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectnaming"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

type templateData struct {
	Dir, Name, ID, Kind, Stack string
	// The test project's framework and packages, for the function apps that write their own.
	TestFramework, TestSdkVersion, XunitVersion, XunitRunnerVersion string
	// Data is what a function app (ADR 0072) adds; zero for any other project.
	functions.Data
}

// Generate writes a project and its tests project, and the workspace's
// Directory.Build.props the first time.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))
	if _, err := os.Stat(dir); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	templates := maps.Clone(sliceTemplates)
	maps.Copy(templates, projectTemplate)
	plan, isFunction := functions.PlanOf(params.Choices)
	id := cmp.Or(params.PackageName, projectnaming.Name(params.Workspace, projectnaming.DotNet, params.Name))
	switch {
	case isFunction && params.Kind != protocol.ProjectApp:
		return protocol.GenerateResult{}, fmt.Errorf("a function app is an app: smith add app %s --lang dotnet --with %s", params.Name, plan.Host)
	case isFunction:
		if plan.Framework == "" {
			plan.Runtime, plan.Framework = "dotnet10", TargetFramework
		}
		maps.Copy(templates, functions.Templates(plan.Host))
	case params.Kind == protocol.ProjectApp:
		maps.Copy(templates, appTemplates)
	}
	if _, err := os.Stat(filepath.Join(workspaceRoot, directoryBuildProps)); errors.Is(err, fs.ErrNotExist) {
		maps.Copy(templates, rootTemplates)
	} else if err != nil {
		return protocol.GenerateResult{}, err
	}
	data := templateData{Dir: params.Directory, Name: params.Name, ID: id, Kind: string(params.Kind), Stack: params.Stack,
		TestFramework: TargetFramework, TestSdkVersion: TestSdkVersion, XunitVersion: XunitVersion, XunitRunnerVersion: XunitRunnerVersion}
	if isFunction {
		data.Data = functions.DataOf(plan, params.Directory, id)
	}
	files, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	notes := []string{"Reference it from another project with <ProjectReference Include=\"<relative path>/" + data.ID + ".csproj\" />."}
	if isFunction {
		notes = append(notes, functionNotes(plan.Host)...)
	}
	return protocol.GenerateResult{Files: files, Notes: notes}, nil
}
