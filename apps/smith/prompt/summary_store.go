package prompt

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// summaryPath is where the summary lives, beside the other small state files
// of smith (operations.json) and outside the task cache, which prune empties.
func summaryPath(root string) string {
	return filepath.Join(root, ".smith", "prompt.json")
}

// Save writes the summary for the workspace at root. The write is atomic, so a
// prompt that reads at the same moment never sees half a file.
//
// It returns the error of the first step that fails; callers may ignore it,
// because the prompt works without a summary.
func Save(root string, s Summary) error {
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	path := summaryPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load reads the summary saved for the workspace at root. The second result is
// false when there is none, when it is unreadable, when it has another shape,
// or when it was written for another workspace (a copied folder).
func Load(root string) (Summary, bool) {
	raw, err := os.ReadFile(summaryPath(root))
	if err != nil {
		return Summary{}, false
	}
	var s Summary
	if json.Unmarshal(raw, &s) != nil || s.Schema != schemaVersion || s.Workspace != root {
		return Summary{}, false
	}
	return s, true
}
