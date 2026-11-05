package toolinstall

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// Need is a tool the workspace needs and who needs it.
type Need struct {
	protocol.ToolNeed
	// Dirs are the absolute folders where a pin file is looked up: the
	// workspace root first, then the projects that use the tool.
	Dirs []string `json:"-"`
	// Projects are the projects that use the tool; empty for a tool that no
	// project needs yet (named on the command line).
	Projects []string `json:"projects,omitempty"`
}

// State is what detection found.
type State string

// The states of a tool.
const (
	// Found means the tool is on PATH and its version fits the workspace.
	Found State = "ok"
	// Outdated means the tool is on PATH and its version is older than the workspace wants.
	Outdated State = "outdated"
	// Missing means no program of the tool runs.
	Missing State = "missing"
)

// Status is the result of detecting one tool.
type Status struct {
	Need
	State State `json:"state"`
	// Program is the program that answered; empty when missing.
	Program string `json:"program,omitempty"`
	// Version is the version the program printed; empty when missing or when it printed none.
	Version string `json:"version,omitempty"`
	// Wanted is the version range the workspace pins; the zero value means any version.
	Wanted Requirement `json:"wanted"`
	// WantedFrom names the file that gave Wanted.
	WantedFrom string `json:"wantedFrom,omitempty"`
}

// Problem says in plain words what is wrong with the tool, to follow its title:
// "is not installed", "is too old (1.2 is installed, 1.4 or newer is wanted)",
// or, for a tool with one wanted version, "is 2.13.0, not the 2.15.0 that CI
// pins". It is empty when the tool is found.
func (s Status) Problem() string {
	switch {
	case s.State == Missing:
		return "is not installed"
	case s.State == Outdated && s.Exact:
		return "is " + s.Version + ", not the " + s.Suggested + " that CI pins"
	case s.State == Outdated:
		return "is too old (" + s.Version + " is installed, " + s.Wanted.String() + " is wanted)"
	}
	return ""
}

// Lacking reports whether the tool needs an install: it is missing or outdated.
func (s Status) Lacking() bool { return s.State != Found }
