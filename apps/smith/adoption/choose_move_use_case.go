package adoption

import (
	"errors"
	"io"
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"golang.org/x/term"
)

// MoveOptions is what --init knows about moving the root project.
type MoveOptions struct {
	// Dir is --move: a folder, adoptmove.Leave, or "" when it was not given.
	Dir string
	// Yes is --yes: do not ask.
	Yes bool
	// In and Out are where a question is read and written.
	In  io.Reader
	Out io.Writer
}

// interactive says a person can answer a question: both ends are a terminal and
// CI is not set. SMITH_FORCE_INTERACTIVE=1 asks anyway, so a script can feed
// the answers on standard input.
func interactive() bool {
	if os.Getenv("SMITH_FORCE_INTERACTIVE") == "1" {
		return true
	}
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("CI") == ""
}

// ChooseMove decides, for --init, whether the root project moves and where. It
// returns a nil plan when nothing moves. The folder comes from --move; in a
// terminal (and without --yes) the person is asked; anything else is refused
// with the exact command, because smith never picks the folder. Nothing is
// changed. It returns the candidate with the plan.
func ChooseMove(r Report, opts MoveOptions) (*adoptmove.Candidate, *adoptmove.Plan, error) {
	if opts.Dir == adoptmove.Leave {
		return nil, nil, nil
	}
	if len(r.Projects) > 0 {
		if opts.Dir != "" {
			return nil, nil, problem.New("--move has nothing to move: smith already sees a project in a folder of its own", "drop --move")
		}
		return nil, nil, nil
	}
	c, err := adoptmove.Detect(r.Root)
	switch {
	case errors.Is(err, adoptmove.ErrNoRootProject):
		if opts.Dir != "" {
			return nil, nil, problem.New("--move has nothing to move: the repository root holds no project (package.json, go.mod, pyproject.toml or Cargo.toml)", "drop --move, or give a repository whose project is at the root")
		}
		return nil, nil, nil
	case err != nil:
		return nil, nil, err
	}
	dir := opts.Dir
	if dir == "" {
		if opts.Yes || !interactive() {
			return nil, nil, adoptmove.NeedsAnswer(c)
		}
		if dir, err = adoptmove.AskWhere(c, opts.In, opts.Out); err != nil {
			return nil, nil, err
		}
		if dir == adoptmove.Leave {
			return nil, nil, nil
		}
	}
	plan, err := PlanFor(r.Root, c, dir)
	if err != nil {
		return nil, nil, err
	}
	return &c, &plan, nil
}

// PlanFor is the move of c into dir (checked) against the repository's files.
func PlanFor(root string, c adoptmove.Candidate, dir string) (adoptmove.Plan, error) {
	clean, err := adoptmove.CleanFolder(dir)
	if err != nil {
		return adoptmove.Plan{}, err
	}
	files, err := gitclient.New(root).Files()
	if err != nil {
		return adoptmove.Plan{}, err
	}
	return adoptmove.PlanMove(c, clean, files)
}
