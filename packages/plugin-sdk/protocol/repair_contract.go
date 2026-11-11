package protocol

// RepairKind names one way to repair the installed dependencies of a
// workspace. The four kinds go from the lightest to the heaviest; a plugin
// offers only the kinds its tool has an honest equivalent for.
type RepairKind string

// The repair kinds, lightest first.
const (
	// RepairPrune removes installed packages that no manifest declares.
	RepairPrune RepairKind = "prune"
	// RepairReinstall deletes the installed dependencies and installs them
	// again from the lockfile: the versions stay.
	RepairReinstall RepairKind = "reinstall"
	// RepairLockfile deletes the installed dependencies and the lockfile, then
	// installs: the versions can change.
	RepairLockfile RepairKind = "lockfile"
	// RepairCache clears the download cache of the tool, then installs. The
	// cache usually belongs to the machine, not to the workspace.
	RepairCache RepairKind = "cache"
)

// RepairKinds lists every repair kind, lightest first: the order an interface
// offers them in.
func RepairKinds() []RepairKind {
	return []RepairKind{RepairPrune, RepairReinstall, RepairLockfile, RepairCache}
}

// RepairParams asks a plugin how it repairs the dependencies of its projects.
type RepairParams struct {
	// Projects are the plugin's projects the repair is for. The plan may also
	// name what the projects share (the workspace's one lockfile).
	Projects []Project `json:"projects"`
}

// RepairResult is one plan for each repair kind the plugin offers for the
// projects. A kind that is not here has no honest equivalent in the tool, or
// has nothing to act on (no lockfile to rebuild).
type RepairResult struct {
	Plans []RepairPlan `json:"plans"`
}

// RepairPlan says exactly what one repair kind does with the plugin's tool.
// The plugin only describes it: the host shows the plan, asks, deletes the
// paths, runs the commands and then installs, in that order.
type RepairPlan struct {
	// Kind is the repair this plan carries out.
	Kind RepairKind `json:"kind"`
	// Summary says in one line what the kind does with this tool ("delete
	// every node_modules folder, then npm install").
	Summary string `json:"summary"`
	// Delete are the files and folders the host deletes first, as
	// workspace-relative paths with forward slashes. Name only what exists. The
	// host deletes a link as a link and never follows it.
	Delete []string `json:"delete,omitempty"`
	// Commands run after the deletions, in order; the first that fails stops
	// the repair.
	Commands []RepairCommand `json:"commands,omitempty"`
	// Install asks the host to call dependencies/install for the projects once
	// the commands are done.
	Install bool `json:"install,omitempty"`
	// MachineWide is true when the repair changes something outside the
	// workspace that other workspaces on the machine use (a shared download
	// cache). The host says so before it asks.
	MachineWide bool `json:"machineWide,omitempty"`
	// ChangesVersions is true when the repair can install other versions than
	// the ones installed now (the lockfile is written again).
	ChangesVersions bool `json:"changesVersions,omitempty"`
	// Notes are remarks the host shows with the plan.
	Notes []string `json:"notes,omitempty"`
}

// RepairCommand is one command of a repair plan.
type RepairCommand struct {
	// Argv is the program and its arguments; no shell is involved.
	Argv []string `json:"argv"`
	// Dir is where it runs, workspace-relative; empty is the workspace root.
	Dir string `json:"dir,omitempty"`
}
