package operation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
)

// Uncommitted is the last operation that left its files uncommitted: its words
// and the files it changed, so smith commit can take those and leave the
// person's own changes out.
type Uncommitted struct {
	// Command is the human text of the operation: "add app web".
	Command string `json:"command"`
	// Files are the workspace-relative paths the operation changed.
	Files []string `json:"files"`
}

func pendingPath(root string) string {
	return filepath.Join(root, ".smith", "uncommitted-operation.json")
}

// WritePending records the last uncommitted operation.
//
// Parameters: root is the workspace root; u is the operation.
// Returns: an error when the file cannot be written.
func WritePending(root string, u Uncommitted) error {
	raw, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(pendingPath(root)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(pendingPath(root), raw, 0o644)
}

// ReadPending reads the last uncommitted operation.
//
// Parameters: root is the workspace root.
// Returns: the operation, and false when none is recorded.
func ReadPending(root string) (Uncommitted, bool) {
	raw, err := os.ReadFile(pendingPath(root))
	if err != nil {
		return Uncommitted{}, false
	}
	var u Uncommitted
	if json.Unmarshal(raw, &u) != nil || u.Command == "" {
		return Uncommitted{}, false
	}
	return u, true
}

// clearPending forgets the recorded operation, once it is committed.
func clearPending(root string) { _ = os.Remove(pendingPath(root)) }

// changedSet is the paths the work tree has changed now, smith's own state left out.
func changedSet(root string) map[string]bool {
	lines, err := changedLines(gitclient.New(root))
	out := map[string]bool{}
	if err != nil {
		return out
	}
	for _, l := range lines {
		out[l.Path] = true
	}
	return out
}

// newSince are the paths that changed after an operation and not before it, sorted.
func newSince(before, after map[string]bool) []string {
	var out []string
	for p := range after {
		if !before[p] {
			out = append(out, p)
		}
	}
	slices.Sort(out)
	return out
}

// FinishOnly commits only the listed files as one recorded operation, and leaves
// every other change of the work tree, staged or not, where it is.
//
// Parameters: command is the human text of the operation; files are the paths to commit.
// Returns: the commit hash, or "" when none of the files changed; an error when git failed.
func (p *Pending) FinishOnly(command string, files []string) (string, error) {
	if len(files) == 0 {
		return "", nil
	}
	if _, err := p.git.Run(append([]string{"add", "-A", "--"}, files...)...); err != nil {
		return "", err
	}
	body := "Made by smith: " + command + ".\n\nsmith undo reverts this commit."
	if _, err := p.git.Run(append([]string{"commit", "-m", "chore: " + command, "-m", body, "--only", "--"}, files...)...); err != nil {
		return "", err
	}
	hash, err := p.git.Run("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return hash, recordOperation(p.root, command, hash, strings.Join(files, "\n"))
}
