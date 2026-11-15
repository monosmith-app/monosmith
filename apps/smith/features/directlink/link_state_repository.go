package directlink

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// stateFile is the file in the smith home that remembers the link.
const stateFile = "link.json"

// LoadState reads what was linked; ok is false when nothing was.
//
// Parameters: home is the smith home.
// Returns: the state, and true when one was kept.
func LoadState(home string) (State, bool) {
	var s State
	raw, err := os.ReadFile(StatePath(home))
	if err != nil || json.Unmarshal(raw, &s) != nil || s.Version == "" {
		return State{}, false
	}
	return s, true
}

// saveState writes what was linked; a failure is returned so the caller can say so.
func saveState(home string, s State) error {
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(StatePath(home)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(StatePath(home), append(raw, '\n'), 0o644)
}

// forgetState removes the memory of the link.
func forgetState(home string) error {
	if err := os.Remove(StatePath(home)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
