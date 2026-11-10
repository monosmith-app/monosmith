package prompt

import (
	"sort"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// schemaVersion changes when the saved file changes shape; an older file reads
// as "no summary yet" until the next smith doctor run replaces it.
const schemaVersion = 1

// Project is one project as the prompt needs it: where it lives and which
// projects depend on it, so a changed file maps to its project without a plugin.
type Project struct {
	// Name is the project name.
	Name string `json:"name"`
	// Root is the workspace-relative folder, with forward slashes.
	Root string `json:"root"`
	// Dependents are the projects that reach this one, sorted; omitted when none.
	Dependents []string `json:"dependents,omitempty"`
}

// Summary is what a normal smith doctor run leaves for the prompt. It is keyed
// to one workspace: Workspace holds the root it was written for.
type Summary struct {
	// Schema is the version of the file's shape.
	Schema int `json:"schema"`
	// Workspace is the absolute workspace root the summary belongs to.
	Workspace string `json:"workspace"`
	// CheckedAt is when doctor ran.
	CheckedAt time.Time `json:"checkedAt"`
	// Failed is the number of failed checks.
	Failed int `json:"failed"`
	// Warnings is the number of warnings.
	Warnings int `json:"warnings"`
	// Projects are the workspace's projects.
	Projects []Project `json:"projects"`
}

// NewSummary builds the summary of a doctor run.
//
// root is the absolute workspace root, graph the project graph the run used,
// failed and warnings the counts of the report, and now the time of the run.
func NewSummary(root string, graph *projectgraph.Graph, failed, warnings int, now time.Time) Summary {
	projects := make([]Project, 0, len(graph.Projects))
	for _, p := range graph.Projects {
		dependents := graph.TransitiveDependents(p.Name)
		sort.Strings(dependents)
		projects = append(projects, Project{Name: p.Name, Root: p.Root, Dependents: dependents})
	}
	return Summary{Schema: schemaVersion, Workspace: root, CheckedAt: now.UTC(), Failed: failed, Warnings: warnings, Projects: projects}
}
