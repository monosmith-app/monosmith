package adoptmove

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
)

// Apply runs the plan in the repository at root: git mv for each step, then the
// fix-up of the root .gitignore lines that pointed at moved folders. The moves
// stay staged and nothing is committed. The work tree must be clean (the caller
// checks); a failing git mv stops the run and the error names the step, and
// git restore --staged --worktree . takes back what ran.
func Apply(root string, p Plan, out io.Writer) error {
	git := gitclient.New(root)
	if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(p.Dir)), 0o755); err != nil {
		return err
	}
	moved := make([]string, 0, len(p.Moves))
	for _, m := range p.Moves {
		if _, err := git.Run("mv", m.From, m.To); err != nil {
			return fmt.Errorf("git mv %s %s: %w", m.From, m.To, err)
		}
		moved = append(moved, m.From)
	}
	fmt.Fprintf(out, "  moved %d entries into %s\n", len(p.Moves), p.Dir)
	path := filepath.Join(root, ".gitignore")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	fixed, changed := FixGitignore(string(raw), p.Dir, moved)
	if !changed {
		return nil
	}
	if err := os.WriteFile(path, []byte(fixed), 0o644); err != nil {
		return err
	}
	if _, err := git.Run("add", ".gitignore"); err != nil {
		return err
	}
	fmt.Fprintln(out, "  updated the paths in .gitignore that pointed at moved folders")
	return nil
}

// Commit commits what Apply staged, as its own commit, and returns its message.
// It runs the repository's hooks. smith undo does not cover this commit (it is
// made before the workspace exists): git revert HEAD takes it back.
func Commit(root string, c Candidate, p Plan) (string, error) {
	git := gitclient.New(root)
	subject := fmt.Sprintf("chore: move %s into %s", c.Name, p.Dir)
	body := "Made by smith adopt: smith needs one folder per project. git revert HEAD undoes this move."
	if _, err := git.Run("commit", "-m", subject, "-m", body); err != nil {
		return "", fmt.Errorf("%s", strings.SplitN(err.Error(), "\n", 2)[0])
	}
	return subject, nil
}
