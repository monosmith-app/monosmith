package adoptflow

// Kind is what the folder is.
type Kind string

// The kinds. Nx and Turborepo workspaces are converted; a plain repository
// (npm workspaces, Go modules, Python projects) is adopted; the others have
// nothing to do yet.
const (
	KindNx        Kind = "nx"
	KindTurborepo Kind = "turborepo"
	KindPlain     Kind = "plain"
	KindSmith     Kind = "smith"
	KindNotARepo  Kind = "not-a-repo"
)

// Project is one project smith would see.
type Project struct {
	Name     string `json:"name"`
	Root     string `json:"root"`
	Language string `json:"language,omitempty"`
	Kind     string `json:"kind,omitempty"`
}

// Found is what a repository is and uses today: the first screen of the
// adopt wizard, and what decides which levels it offers.
type Found struct {
	Kind Kind `json:"kind"`
	// Source names the kind in words: "Nx workspace", "repository".
	Source string `json:"source"`
	// Name is what the workspace would be called.
	Name     string    `json:"name"`
	Projects []Project `json:"projects,omitempty"`
	// Tooling are the tools in use that the full level can retire or
	// replace: "Nx", "mnci", "ESLint", "Jest", "Vitest", "rollup", ...
	Tooling []string `json:"tooling,omitempty"`
	// Blockers stop the wizard before anything changes, each with the step
	// that clears it. Warnings are told and do not stop it.
	Blockers []string `json:"blockers,omitempty"`
	// NeedsInstall is true when the one blocker is that the packages are not
	// installed, or are not the ones the repository asks for: installing them
	// clears it, and the wizard offers to.
	NeedsInstall bool     `json:"needsInstall,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
	// Full is true when the full level is offered; NoFull says why not.
	Full   bool   `json:"full"`
	NoFull string `json:"noFull,omitempty"`
}

// Convertible is true for the kinds the conversion handles.
func (k Kind) Convertible() bool { return k == KindNx || k == KindTurborepo }
