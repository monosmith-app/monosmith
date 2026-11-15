package nxworkspace

import "encoding/json"

// Workspace is an Nx workspace as read from disk.
type Workspace struct {
	Root string
	Nx   NxJSON
	// NxVersion is the nx devDependency range of the root package.json.
	NxVersion   string
	RootPackage PackageJSON
	// RootTargets are the targets of the workspace's root project, if any.
	RootTargets map[string]Target
	Projects    []*Project
}

// Legacy reports whether the old mnci (4.x) made the workspace.
func (w *Workspace) Legacy() bool { return w.Nx.Mnci != nil }

// Project is one Nx project.
type Project struct {
	// Name is the Nx project name.
	Name string
	// Root is the workspace-relative directory, forward slashes.
	Root string
	// Type is Nx's node type: app, lib or e2e.
	Type        string
	ProjectType string
	Tags        []string
	// Targets are the resolved targets, inferred ones included.
	Targets map[string]Target
	// Release is project.json's "release" section.
	Release map[string]json.RawMessage
	// HasProjectJSON is true when the project has a project.json.
	HasProjectJSON bool
	// Package is the project's package.json, if any.
	Package *PackageJSON
	// Files are the project's files, workspace-relative.
	Files []string
}

// HasTag reports whether the project carries a tag.
func (p *Project) HasTag(tag string) bool {
	for _, t := range p.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// HasFile reports whether the project has a file at a project-relative path.
func (p *Project) HasFile(rel string) bool {
	want := p.Root + "/" + rel
	for _, f := range p.Files {
		if f == want {
			return true
		}
	}
	return false
}

// PackageJSON is the part of a package.json a conversion reads.
type PackageJSON struct {
	Name                 string                     `json:"name"`
	Version              string                     `json:"version"`
	Private              bool                       `json:"private"`
	Scripts              map[string]string          `json:"scripts"`
	Dependencies         map[string]string          `json:"dependencies"`
	DevDependencies      map[string]string          `json:"devDependencies"`
	PeerDependencies     map[string]string          `json:"peerDependencies"`
	Workspaces           json.RawMessage            `json:"workspaces"`
	PublishConfig        json.RawMessage            `json:"publishConfig"`
	Nx                   json.RawMessage            `json:"nx"`
	Main                 string                     `json:"main"`
	Module               string                     `json:"module"`
	Types                string                     `json:"types"`
	Type                 string                     `json:"type"`
	OptionalDependencies map[string]string          `json:"optionalDependencies"`
	Overrides            map[string]json.RawMessage `json:"overrides"`
}

// WorkspaceGlobs are the npm/yarn workspaces globs.
func (p PackageJSON) WorkspaceGlobs() []string {
	var list []string
	if json.Unmarshal(p.Workspaces, &list) == nil {
		return list
	}
	var obj struct {
		Packages []string `json:"packages"`
	}
	_ = json.Unmarshal(p.Workspaces, &obj)
	return obj.Packages
}
