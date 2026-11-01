package convertplan

import (
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Severity orders findings from harmless to blocking.
type Severity string

// Severities. Manual and Unsupported findings fail the conversion under
// --strict; a Blocker always does.
const (
	// Mapped records what moved where.
	Mapped Severity = "mapped"
	// Dropped records Nx plumbing smith does not need.
	Dropped Severity = "dropped"
	// Changed records a behaviour change the user should know about.
	Changed Severity = "changed"
	// Manual is something the conversion could not do; the user must.
	Manual Severity = "manual"
	// Unsupported is a project or setting smith has no plugin for.
	Unsupported Severity = "unsupported"
	// Blocker prevents the conversion.
	Blocker Severity = "blocker"
)

// Finding is one line of the conversion report.
type Finding struct {
	Severity Severity `json:"severity"`
	// Project is the smith project name, or "" for the workspace.
	Project string `json:"project,omitempty"`
	// Subject is the target, file or setting the finding is about.
	Subject string `json:"subject"`
	Message string `json:"message"`
	// Decide marks a finding the reader should weigh before applying (CI is
	// replaced, release rules change), not one to skim.
	Decide bool `json:"decide,omitempty"`
}

// ProjectPlan is how one Nx project becomes a smith project.
type ProjectPlan struct {
	// Nx is the Nx project name; Name is the smith name (the directory name).
	Nx       string               `json:"nx"`
	Name     string               `json:"name"`
	Root     string               `json:"root"`
	Language string               `json:"language,omitempty"`
	Kind     protocol.ProjectKind `json:"kind,omitempty"`
	// KindMarker writes the kind into the manifest, because the directory
	// does not imply it.
	KindMarker bool `json:"kindMarker,omitempty"`
	// Scripts are package.json scripts to set (Node projects).
	Scripts map[string]string `json:"scripts,omitempty"`
	// Supported is false for a language smith has no plugin for.
	Supported bool `json:"supported"`
	// Declared is true for a project with no manifest of its own, declared in
	// smith.toml with a root and a language.
	Declared bool `json:"declared,omitempty"`
}

// The tools a plan converts from.
const (
	SourceNx        = "nx"
	SourceMnci      = "mnci"
	SourceTurborepo = "turborepo"
)

// Plan is everything a conversion will do.
type Plan struct {
	Root string `json:"root"`
	// Source is the tool converted from (SourceNx, SourceMnci or
	// SourceTurborepo).
	Source string `json:"source"`
	Legacy bool   `json:"legacy"`
	// Config is the smith.toml to write.
	Config   workspace.Config `json:"-"`
	Projects []*ProjectPlan   `json:"projects"`
	Findings []Finding        `json:"findings"`
	// RootScripts maps root package.json scripts to their new command; ""
	// removes the script.
	RootScripts map[string]string `json:"rootScripts,omitempty"`
	// RemoveDependencies are root devDependencies/dependencies to remove.
	RemoveDependencies []string `json:"removeDependencies,omitempty"`
	// RemoveOverrides are root npm overrides of removed packages.
	RemoveOverrides []string `json:"removeOverrides,omitempty"`
	// AddDevDependencies are root devDependencies to add, with their range.
	AddDevDependencies map[string]string `json:"addDevDependencies,omitempty"`
	// RemoveFields are root package.json fields to delete (packageManager,
	// the package manager's own settings).
	RemoveFields []string `json:"removeFields,omitempty"`
	// SetWorkspaces replaces the root package.json "workspaces" with these
	// globs when it is not nil.
	SetWorkspaces []string `json:"setWorkspaces,omitempty"`
	// SetOverrides are npm overrides to add to the root package.json.
	SetOverrides map[string]string `json:"setOverrides,omitempty"`
	// Specifiers rewrite dependency specifiers: the project root ("." for
	// the workspace root) maps each dependency name to its new range. A name
	// is rewritten in every dependency section that lists it.
	Specifiers map[string]map[string]string `json:"specifiers,omitempty"`
	// Delete lists workspace-relative files and directories to remove.
	Delete []string `json:"delete,omitempty"`
	// Write maps workspace-relative files to their new content.
	Write map[string]string `json:"-"`
	// Keep maps files the converted workspace needs to why, for the next
	// steps: they look like clutter, but deleting one breaks builds or tests.
	Keep map[string]string `json:"keep,omitempty"`
	// WriteFiles lists the keys of Write, for the report.
	WriteFiles []string `json:"write,omitempty"`
	// names maps each Nx project name to its smith name, so a target's
	// dependency on another project's target can name it.
	names map[string]string
	// tagChoices are the release tag prefixes planned for the projects smith
	// prefixes with their name; applyTagFormat looks for a format they share.
	tagChoices []tagChoice
}

// Add records a finding.
func (p *Plan) Add(severity Severity, project, subject, message string) {
	p.Findings = append(p.Findings, Finding{Severity: severity, Project: project, Subject: subject, Message: message})
}

// Decide records a behaviour change the reader should decide on before the
// conversion is applied; the report lists it first.
func (p *Plan) Decide(project, subject, message string) {
	p.Findings = append(p.Findings, Finding{Severity: Changed, Project: project, Subject: subject, Message: message, Decide: true})
}

// Decided reports whether a decision finding already names the subject, so
// one file is not announced twice.
func (p *Plan) Decided(subject string) bool {
	for _, f := range p.Findings {
		if f.Decide && f.Subject == subject {
			return true
		}
	}
	return false
}

// Count returns how many findings have each severity.
func (p *Plan) Count(s Severity) int {
	n := 0
	for _, f := range p.Findings {
		if f.Severity == s {
			n++
		}
	}
	return n
}

// Blocked reports whether the conversion must not run; strict also counts
// manual and unsupported findings.
func (p *Plan) Blocked(strict bool) bool {
	if p.Count(Blocker) > 0 {
		return true
	}
	return strict && (p.Count(Manual) > 0 || p.Count(Unsupported) > 0)
}

// Project returns a planned project by smith name.
func (p *Plan) Project(name string) *ProjectPlan {
	for _, pp := range p.Projects {
		if pp.Name == name {
			return pp
		}
	}
	return nil
}

// ProjectAt returns the planned project with a root, or nil.
func (p *Plan) ProjectAt(root string) *ProjectPlan {
	for _, pp := range p.Projects {
		if pp.Root == root {
			return pp
		}
	}
	return nil
}

// Override returns the smith.toml project section, creating it.
func (p *Plan) Override(project string) workspace.ProjectOverride {
	if p.Config.Projects == nil {
		p.Config.Projects = map[string]workspace.ProjectOverride{}
	}
	return p.Config.Projects[project]
}

// SetOverride stores a project section, dropping it when empty.
func (p *Plan) SetOverride(project string, o workspace.ProjectOverride) {
	if o.TagPrefix == "" && o.Release == nil && len(o.Tasks) == 0 && !o.Declared() {
		delete(p.Config.Projects, project)
		return
	}
	p.Config.Projects[project] = o
}

// SetTask stores one task override of a project.
func (p *Plan) SetTask(project, task string, t workspace.TaskOverride) {
	o := p.Override(project)
	if o.Tasks == nil {
		o.Tasks = map[string]workspace.TaskOverride{}
	}
	o.Tasks[task] = t
	p.SetOverride(project, o)
}

// SetSpecifier schedules a dependency's range to be rewritten in the manifest
// of the project at root ("." for the workspace root).
func (p *Plan) SetSpecifier(root, dependency, specifier string) {
	if p.Specifiers == nil {
		p.Specifiers = map[string]map[string]string{}
	}
	if p.Specifiers[root] == nil {
		p.Specifiers[root] = map[string]string{}
	}
	p.Specifiers[root][dependency] = specifier
}

// KeepFile records why the workspace needs a file convert writes.
func (p *Plan) KeepFile(path, why string) {
	if p.Keep == nil {
		p.Keep = map[string]string{}
	}
	p.Keep[path] = why
}

// WriteFile schedules a file to be written.
func (p *Plan) WriteFile(path, content string) {
	if p.Write == nil {
		p.Write = map[string]string{}
	}
	if _, ok := p.Write[path]; !ok {
		p.WriteFiles = append(p.WriteFiles, path)
		sort.Strings(p.WriteFiles)
	}
	p.Write[path] = content
}

// DeletePath schedules a file or directory for removal.
func (p *Plan) DeletePath(path string) {
	for _, d := range p.Delete {
		if d == path {
			return
		}
	}
	p.Delete = append(p.Delete, path)
	sort.Strings(p.Delete)
}
