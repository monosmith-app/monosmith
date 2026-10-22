package protocol

// ContributeParams hands a plugin the projects it applies to: those whose
// stack includes it, or, for a workspace-wide convention, those of the
// languages it supports that smith.toml enables it for.
type ContributeParams struct {
	Projects []Project `json:"projects"`
}

// ContributeResult is what a convention adds, by project name.
type ContributeResult struct {
	Projects map[string]Contribution `json:"projects,omitempty"`
}

// Contribution is a convention's addition to one project.
type Contribution struct {
	// Tasks are new tasks. A task the project already has is kept as it is.
	Tasks map[string]Task `json:"tasks,omitempty"`
	// Replace are tasks that replace the project's task of the same name,
	// or are added when it has none.
	Replace map[string]Task `json:"replace,omitempty"`
	// DependsOn adds dependencies to tasks the project already has, by task
	// name ({"lint": ["slice-check"]}).
	DependsOn map[string][]string `json:"dependsOn,omitempty"`
	// Consumes are workspace-relative globs of the shared schema files the
	// project consumes, so it depends on their owner. A task named generate
	// or generate-<name> that reads a schema through a {workspaceRoot}/
	// input declares it the same way without listing it here.
	Consumes []string `json:"consumes,omitempty"`
}
