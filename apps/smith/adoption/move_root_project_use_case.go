package adoption

import (
	"errors"
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/adoptmove"
)

// planForDryRun puts the move into dir in the report, for --dry-run --move.
func planForDryRun(r *Report, dir string) error {
	c, err := adoptmove.Detect(r.Root)
	if errors.Is(err, adoptmove.ErrNoRootProject) {
		return fmt.Errorf("--move has nothing to move: the repository root holds no project")
	}
	if err != nil {
		return err
	}
	plan, err := PlanFor(r.Root, c, dir)
	if err != nil {
		return err
	}
	r.RootProject, r.MovePlan = &c, &plan
	return nil
}

// MoveRootProject moves the project at the root into the plan's folder with
// git mv, as a commit of its own when commit is true, and reads the repository
// again so the adoption sees the project where it is now. The tree was clean
// before the move, so the report is clean after it: the move's own changes are
// not "uncommitted work" for the steps that follow.
func MoveRootProject(r Report, c adoptmove.Candidate, p adoptmove.Plan, commit bool, version string, log, out io.Writer) (Report, error) {
	if err := adoptmove.Apply(r.Root, p, out); err != nil {
		return r, err
	}
	if commit {
		subject, err := adoptmove.Commit(r.Root, c, p)
		if err != nil {
			return r, fmt.Errorf("the project moved but the commit failed (%w): commit the move yourself", err)
		}
		fmt.Fprintf(out, "  committed: %s (git revert HEAD takes it back)\n", subject)
	}
	for _, note := range p.Notes {
		fmt.Fprintf(out, "  note: %s\n", note)
	}
	next, err := Inspect(r.Root, version, log)
	if err != nil {
		return r, err
	}
	next.Clean = true
	return next, nil
}
