package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/dependencies"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// loadDependencies asks the plugins for newer releases of what the projects
// declare, as `smith deps up` does, and applies the one-version policy when
// smith.toml asks for it. Nothing is written.
func loadDependencies(open func() (*session.Session, error)) depsLoaded {
	s, err := open()
	if err != nil {
		return depsLoaded{err: err}
	}
	defer s.Close()
	report, err := dependencies.FindAvailable(s)
	if err != nil {
		return depsLoaded{err: err}
	}
	var out depsLoaded
	for _, u := range report.Updates {
		out.rows = append(out.rows, depRow{Name: u.Name, Plugin: u.Plugin, Current: u.Current, Latest: u.Latest, Level: u.Level, Projects: u.Projects()})
	}
	for _, f := range dependencies.CheckOneVersionPolicy(s) {
		out.policy = append(out.policy, policyLine{Check: f.Check, OK: f.OK, Warning: f.Warning, Detail: f.Detail, Remedy: f.Remedy})
	}
	return out
}
