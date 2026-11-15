package taskcache

import "time"

// EntrySchema versions the action-cache entry format.
const EntrySchema = 1

// Entry is an action-cache record: what a successful run produced.
type Entry struct {
	Schema  int          `json:"schema"`
	Task    string       `json:"task"`
	Created time.Time    `json:"created"`
	Outputs []OutputFile `json:"outputs"`
	// Log is the CAS digest of the task's combined output.
	Log string `json:"log"`
	// Duration is how long the original run took, shown as time saved.
	Duration time.Duration `json:"duration"`
	// Inputs are what the key was computed from, for smith why-miss. Entries
	// written before smith 0.2 have none.
	Inputs *KeyInputs `json:"inputs,omitempty"`
}

// OutputFile is one file restored on a hit.
type OutputFile struct {
	// Path is workspace-relative, forward slashes.
	Path   string `json:"path"`
	Digest string `json:"digest"`
	Mode   uint32 `json:"mode"`
}
