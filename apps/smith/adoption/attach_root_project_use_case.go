package adoption

import (
	"errors"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// attachRootProject looks for the one project a plain repository keeps at its
// root, which smith cannot manage there. When it finds one, the report carries
// it, with the plan for the suggested folder, so a dry run shows the move and
// the question. A root that is ambiguous becomes a warning that says why. It
// returns false when the root holds no project (the caller explains the empty
// list another way).
func attachRootProject(r *Report, files []string) bool {
	c, err := adoptmove.Detect(r.Root)
	if errors.Is(err, adoptmove.ErrNoRootProject) {
		return false
	}
	if err != nil {
		r.Warnings = append(r.Warnings, itemOfProblem(err))
		return true
	}
	r.RootProject = &c
	choices := adoptmove.Choices(c)
	plan, err := adoptmove.PlanMove(c, choices[adoptmove.DefaultChoice(choices)-1].Dir, files)
	if err != nil {
		r.Warnings = append(r.Warnings, itemOfProblem(err))
		return true
	}
	r.MovePlan = &plan
	return true
}

// itemOfProblem makes a finding of a problem: its sentence and each way out.
func itemOfProblem(err error) Item {
	var p *problem.Error
	if !errors.As(err, &p) {
		return Item{Message: err.Error()}
	}
	item := Item{Message: p.Problem, Steps: p.Remedy}
	if len(p.Remedy) > 0 {
		item.Step = p.Remedy[0]
	}
	return item
}

// withoutRootProgramWarning drops the warning that asks the person to move a
// root main package by hand, which the move offer replaces.
func withoutRootProgramWarning(items []Item) []Item {
	kept := items[:0]
	for _, it := range items {
		if !strings.HasPrefix(it.Message, "the root module's main package") {
			kept = append(kept, it)
		}
	}
	return kept
}
