package tui

import (
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
)

// waitKind names the load the open tab is waiting for ("" when it waits for
// none): Home and Doctor wait for the same doctor run.
func (s Shell) waitKind() string {
	switch s.tab {
	case commands.TabHome, commands.TabDoctor:
		if s.doctor.Report == nil && s.doctor.Err == nil && s.doctor.Loading {
			return "doctor"
		}
	case commands.TabDeps:
		if !s.deps.Loaded && s.deps.Err == nil && s.sources.Deps != nil {
			return "deps"
		}
	case commands.TabRelease:
		if !s.release.Loaded && s.release.Err == nil && s.sources.Release != nil {
			return "release"
		}
	}
	return ""
}

// trackWait starts the clock when a wait begins and stops it when it ends, so
// the screen can say how long the user has been waiting.
func (s *Shell) trackWait() {
	kind := s.waitKind()
	if kind == "" {
		s.waitFor, s.waitSince = "", time.Time{}
		return
	}
	if kind != s.waitFor {
		s.waitFor, s.waitSince = kind, time.Now()
	}
}

// waited is how long the open tab has been waiting.
func (s Shell) waited() time.Duration {
	if s.waitFor == "" || s.waitSince.IsZero() {
		return 0
	}
	return time.Since(s.waitSince)
}
