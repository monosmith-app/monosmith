package cipipeline

import (
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/boundaries"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// verifyBoundaries fails the verify phase on a dependency between projects
// that breaks a rule of smith.toml [boundaries], naming each. Warnings are
// printed and let it pass.
func verifyBoundaries(s *session.Session, out io.Writer) error {
	violations := boundaries.Check(s.Graph, s.Config.Boundaries)
	if len(violations) == 0 {
		return nil
	}
	fmt.Fprint(out, boundaries.Text(violations))
	if errs := boundaries.Errors(violations); len(errs) > 0 {
		return fmt.Errorf("%d dependency(ies) between projects break the rules in smith.toml [boundaries]: %w", len(errs), boundaries.ErrViolations)
	}
	return nil
}
