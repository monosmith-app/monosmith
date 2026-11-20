package adoptmove

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/integrations/gitclient"
)

// Apply runs the plan in the repository at root: git mv for each step, a
// staged copy of each of the plan's Copies, then the
// fix-up of the root .gitignore lines for the project's new folder. The moves
// stay staged and nothing is committed. The work tree must be clean (the caller
// checks); a failing git mv stops the run and the error names the step, and
// git restore --staged --worktree . takes back what ran.
func Apply(root string, p Plan, out io.Writer) error {
	git := gitclient.New(root)
	if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(p.Dir)), 0o755); err != nil {
		return err
	}
	if len(p.Docs) > 0 {
		if err := os.MkdirAll(filepath.Join(root, DocsDir), 0o755); err != nil {
			return err
		}
	}
	moved := make([]string, 0, len(p.Moves))
	for _, m := range append(append([]Step{}, p.Moves...), p.Docs...) {
		if _, err := git.Run("mv", m.From, m.To); err != nil {
			return fmt.Errorf("git mv %s %s: %w", m.From, m.To, err)
		}
		moved = append(moved, m.From)
	}
	fmt.Fprintf(out, "  moved %d entries into %s\n", len(p.Moves), p.Dir)
	for _, c := range p.Copies {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(c.From)))
		if err != nil {
			return fmt.Errorf("copy %s: %w", c.From, err)
		}
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(c.To)), raw, 0o644); err != nil {
			return fmt.Errorf("copy %s: %w", c.From, err)
		}
		if _, err := git.Run("add", c.To); err != nil {
			return fmt.Errorf("git add %s: %w", c.To, err)
		}
		fmt.Fprintf(out, "  copied %s into %s, so the published package keeps it\n", c.From, p.Dir)
	}
	if len(p.Docs) > 0 {
		fmt.Fprintf(out, "  moved %d document(s) into %s\n", len(p.Docs), DocsDir)
	}
	path := filepath.Join(root, ".gitignore")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	fixed, changed := FixGitignore(string(raw), p.Dir, moved, p.Stays)
	if !changed {
		return nil
	}
	if err := os.WriteFile(path, []byte(fixed), 0o644); err != nil {
		return err
	}
	if _, err := git.Run("add", ".gitignore"); err != nil {
		return err
	}
	fmt.Fprintln(out, "  updated the paths in .gitignore for the project's new folder")
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
