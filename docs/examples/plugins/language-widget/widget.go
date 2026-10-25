package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

// widget is the plugin. Every method it has beyond Describe is a capability
// serve.Main advertises: Discover, Generate, WriteVersion and Doctor.
type widget struct{ root string }

// Describe answers "initialize".
func (w *widget) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	w.root = p.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin: protocol.Peer{Name: "widget", Version: "0.1.0"},
		Kind:   protocol.KindLanguage,
		// A language plugin that also writes new projects plays the scaffold
		// level too; it does not need to be chosen as a toolchain.
		Roles:     []protocol.Role{protocol.RoleLanguage, protocol.RoleScaffold},
		Languages: []string{"widget"},
		Manifests: []string{manifestFile},
	}, nil
}

// self is the command the tasks run: this executable. A plugin that wraps a
// real tool would name the tool here instead.
func self() string {
	exe, err := os.Executable()
	if err != nil {
		return "smith-plugin-widget"
	}
	return exe
}

// Discover answers "discover": one project per widget.json that is ours.
func (w *widget) Discover(p protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, rel := range p.Manifests { // workspace-relative, forward slashes
		m, err := readManifest(filepath.Join(w.root, filepath.FromSlash(rel)))
		if err != nil {
			return result, err
		}
		if m.Name == "" {
			continue // not a widget manifest: leave it out
		}
		kind := protocol.ProjectKind(m.Kind)
		if kind == "" {
			kind = protocol.ProjectLib
		}
		root := path.Dir(rel)
		result.Projects = append(result.Projects, protocol.Project{
			Name:        m.Name,
			Root:        root,
			Language:    "widget",
			Kind:        kind,
			Version:     m.Version,
			Description: m.Description,
			Stack:       m.Stack,
			Exports:     []string{m.Name},
			Imports:     m.Uses, // the host keeps the ones another project exports
			Tasks: map[string]protocol.Task{
				"build": {
					Command:     []string{self(), "build", protocol.BuildDirFromProject(root)},
					Inputs:      []string{manifestFile},
					Outputs:     []string{protocol.BuildDirFromProject(root)},
					DependsOn:   []string{"^build"},
					Cache:       true,
					Description: "write widget.txt into the build directory",
				},
				"test": {
					Command:     []string{self(), "check"},
					Inputs:      []string{manifestFile},
					Cache:       true,
					Description: "check the manifest: a name and an x.y.z version",
				},
			},
		})
	}
	return result, nil
}

// Generate answers "generate": a new project's starter files.
func (w *widget) Generate(p protocol.GenerateParams) (protocol.GenerateResult, error) {
	data := map[string]string{
		"Dir": p.Directory, "Name": p.Name, "Kind": string(p.Kind), "Stack": p.Stack,
		"Workspace": p.Workspace.Name,
	}
	files, err := templating.Write(w.root, map[string]string{
		"[[.Dir]]/widget.json": `{
  "name": "[[.Name]]",
  "version": "0.0.0",
  "kind": "[[.Kind]]",
  "description": "A widget of the [[.Workspace]] workspace.",
  "license": "MIT"[[if .Stack]],
  "stack": "[[.Stack]]"[[end]]
}
`,
		"[[.Dir]]/README.md": "# [[.Name]]\n\nA widget. Edit widget.json.\n",
	}, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	return protocol.GenerateResult{Files: files}, nil
}

// WriteVersion answers "version/write": record the release version.
func (w *widget) WriteVersion(p protocol.VersionParams) (protocol.VersionResult, error) {
	file := path.Join(p.Project.Root, manifestFile)
	if !p.DryRun {
		if err := setField(filepath.Join(w.root, filepath.FromSlash(file)), "version", p.Version); err != nil {
			return protocol.VersionResult{}, err
		}
	}
	return protocol.VersionResult{Files: []string{file}}, nil
}

// Doctor answers "doctor": every failed finding names its remedy, and a fix
// when the plugin can write one.
func (w *widget) Doctor(p protocol.DoctorParams) (protocol.DoctorResult, error) {
	result := protocol.DoctorResult{Findings: []protocol.Finding{}}
	for _, project := range p.Projects {
		file := path.Join(project.Root, manifestFile)
		full := filepath.Join(w.root, filepath.FromSlash(file))
		m, err := readManifest(full)
		if err != nil {
			return result, err
		}
		finding := protocol.Finding{Check: project.Name + " names a license", OK: m.License != ""}
		if !finding.OK {
			finding.Detail = file + " has no license field"
			finding.Remedy = `add "license": "MIT" (or your license) to ` + file
			raw, err := os.ReadFile(full)
			if err != nil {
				return result, err
			}
			doc, err := jsondoc.ParseDocument(raw)
			if err != nil {
				return result, fmt.Errorf("%s: %w", file, err)
			}
			doc.SetString("license", "MIT")
			// Not safe: it picks a license for the team, so smith asks first.
			finding.Fix = &protocol.Fix{
				Description: "add an MIT license to widget.json",
				Safe:        false,
				AppliesTo:   []string{project.Name},
				Files:       []protocol.FileEdit{{Path: file, Content: string(doc.Bytes())}},
			}
		}
		result.Findings = append(result.Findings, finding)
	}
	return result, nil
}
