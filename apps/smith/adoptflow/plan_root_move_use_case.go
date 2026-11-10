package adoptflow

import (
	"errors"

	"github.com/monosmith-app/monosmith/apps/smith/adoption"
	"github.com/monosmith-app/monosmith/apps/smith/adoptmove"
)

// MovePlan is the move of the project at the repository root into dir, with
// the development dependencies and the ESLint setup smith cannot place by
// itself in Unsure and Lint: the questions the wizard asks. It reads the
// repository and changes nothing.
//
// Parameters: dir is the folder the person chose.
// Returns: the plan, or an error when the repository has no project at its
// root or the move cannot be planned.
func (s *Session) MovePlan(dir string) (adoptmove.Plan, error) {
	if s.report.RootProject == nil {
		return adoptmove.Plan{}, errors.New("the repository has no project at its root")
	}
	return adoption.PlanFor(s.Root, *s.report.RootProject, dir)
}
