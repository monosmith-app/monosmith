package adoption

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"golang.org/x/term"
)

// MoveOptions is what --init knows about moving the root project.
type MoveOptions struct {
	// Dir is --move: a folder, adoptmove.Leave, or "" when it was not given.
	Dir string
	// Hoist is --hoist: development dependencies the person sends to the
	// workspace without being asked.
	Hoist []string
	// EslintSetup is --eslint-setup: "shared", "keep", or "" when not given.
	EslintSetup string
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
	if err := placeDependencies(&plan, opts); err != nil {
		return nil, nil, err
	}
	return &c, &plan, nil
}

// placeDependencies settles the development dependencies smith cannot place
// itself: the ones named by --hoist go to the workspace, a terminal asks about
// each of the rest, and without a terminal (or with --yes) they stay in the
// project, which the output says, since that changes the least.
func placeDependencies(p *adoptmove.Plan, opts MoveOptions) error {
	var workspace []string
	var open []string
	for _, name := range p.Unsure {
		if slices.Contains(opts.Hoist, name) {
			workspace = append(workspace, name)
		} else {
			open = append(open, name)
		}
	}
	if len(open) > 0 && !opts.Yes && interactive() {
		for _, name := range open {
			toWorkspace, err := adoptmove.AskWorkspace(name, opts.In, opts.Out)
			if err != nil {
				return err
			}
			if toWorkspace {
				workspace = append(workspace, name)
			}
		}
	} else if len(open) > 0 && opts.Out != nil {
		fmt.Fprintf(opts.Out, "  stays in the project (nobody was asked; --hoist <name> sends one to the workspace): %s\n", strings.Join(open, ", "))
	}
	p.Settle(workspace)
	if !p.Lint {
		return nil
	}
	switch {
	case opts.EslintSetup == "shared":
		p.SettleLint(false)
	case opts.EslintSetup == "keep":
		p.SettleLint(true)
	case !opts.Yes && interactive():
		keep, err := adoptmove.AskLint(opts.In, opts.Out)
		if err != nil {
			return err
		}
		p.SettleLint(keep)
	default:
		if opts.Out != nil {
			fmt.Fprintln(opts.Out, "  keeps your own ESLint setup in the project (nobody was asked; --eslint-setup shared uses the shared config)")
		}
		p.SettleLint(true)
	}
	return nil
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
	plan, err := adoptmove.PlanMove(c, clean, files)
	if err != nil {
		return plan, err
	}
	return plan, adoptmove.SortDependencies(root, c, &plan)
}
