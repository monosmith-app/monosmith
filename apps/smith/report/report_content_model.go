package report

// Data is everything the page shows. Gather fills it from an open workspace;
// Render only reads it, so the same Data give the same page. It holds no time
// and no path of the machine, so a page can be compared and attached to a pull
// request.
type Data struct {
	// Workspace is the workspace name.
	Workspace string
	// Branch is the release branch.
	Branch string
	// Projects are the projects, sorted by name.
	Projects []Project
	// GraphSVG is the dependency graph as an inline SVG.
	GraphSVG string
	// Cycles list the import cycles, each as "a → b → a".
	Cycles []string
	// Health are the doctor results.
	Health Health
	// Drift is the state of the derived files.
	Drift Drift
	// Release is what a merge to the release branch would release.
	Release Release
}

// Project is one project with its stack and choices.
type Project struct {
	Name, Root, Kind, Language, Stack string
	// Choices are the choices of the project's stack.
	Choices []string
	// Conventions are the conventions that apply, such as vertical-slices.
	Conventions []string
	// Deps are the projects it depends on.
	Deps []string
}

// Status of a check or a derived file, as a word. The page prints the word,
// never only a colour.
const (
	StatusPass = "Pass"
	StatusWarn = "Warning"
	StatusFail = "Fail"
)

// Health is the doctor report.
type Health struct {
	Checks                 []Check
	Passed, Warned, Failed int
}

// Check is one doctor finding.
type Check struct {
	// Name is the check.
	Name string
	// Status is StatusPass, StatusWarn or StatusFail.
	Status string
	// Detail says what was found.
	Detail string
	// Remedy says what to do.
	Remedy string
}

// Drift is the derived files.
type Drift struct {
	// Enabled is false when [derive] is off in smith.toml.
	Enabled bool
	// Current counts the files that need nothing.
	Current int
	// Items are the files that need action, sorted by path.
	Items []DriftItem
	// Error says why the plan could not be built.
	Error string
}

// DriftItem is one derived file that needs action.
type DriftItem struct {
	Path, Renderer, Status, Owner, Detail string
}

// Release is the release plan of the current commit, as if it were merged.
type Release struct {
	// Items are the packages that would release.
	Items []ReleaseItem
	// Unchanged are the packages with nothing to release.
	Unchanged []string
	// Error says why no plan could be made.
	Error string
}

// ReleaseItem is one package in the plan.
type ReleaseItem struct {
	Project, Current, Next, Level, Reason string
}
