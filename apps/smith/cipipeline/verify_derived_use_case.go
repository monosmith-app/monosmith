package cipipeline

import (
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/derive"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// verifyDerived runs smith derive --check when the workspace turned derived
// files on ([derive] enabled): a derived file or smith.lock that is out of date
// fails the verify phase. It does nothing otherwise.
func verifyDerived(s *session.Session, out io.Writer) error {
	if !s.Config.Derive.Enabled {
		return nil
	}
	plan, _, err := derive.BuildPlan(s)
	if err != nil {
		return err
	}
	return derive.CheckPlan(out, plan)
}
