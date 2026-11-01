package adoptmove

import "errors"

// The ecosystems of the project found at the root.
const (
	EcosystemNode   = "node"
	EcosystemGo     = "go"
	EcosystemPython = "python"
	EcosystemRust   = "rust"
)

// Leave is the --move value that keeps the project where it is.
const Leave = "none"

// ErrNoRootProject says the root holds no project manifest smith can move.
var ErrNoRootProject = errors.New("no project at the repository root")

// Candidate is the one project found at the repository root.
type Candidate struct {
	// Ecosystem is one of the Ecosystem constants.
	Ecosystem string `json:"ecosystem"`
	// Manifest is the file that shows it: "package.json", "go.mod", ...
	Manifest string `json:"manifest"`
	// Name is the folder name offered: the manifest name in kebab-case.
	Name string `json:"name"`
	// App is true when the project looks like a program.
	App bool `json:"app"`
}

// Choice is one answer of the question "where should it go?".
type Choice struct {
	// Key is "apps", "packages", "libs", "other" or "none".
	Key string
	// Dir is the folder, "" for "other" (the person types it) and "none".
	Dir string
	// Label is the option as the prompts print it, without the number.
	Label string
	// Default marks the suggested choice.
	Default bool
}

// Step is one git mv.
type Step struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Plan is what moving the project does.
type Plan struct {
	// Dir is the folder the project goes into.
	Dir string `json:"dir"`
	// Ecosystem is the candidate's.
	Ecosystem string `json:"ecosystem"`
	// Moves are the git mv commands, in order.
	Moves []Step `json:"moves"`
	// Stays are the top-level entries that stay at the root.
	Stays []string `json:"stays"`
	// Notes are things the person should know, one sentence each.
	Notes []string `json:"notes,omitempty"`
}
