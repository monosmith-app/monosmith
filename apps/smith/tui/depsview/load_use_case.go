package depsview

import (
	"github.com/monosmith-app/monosmith/apps/smith/dependencies"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// Load asks the plugins for newer releases of what the projects
// declare, as `smith deps up` does, and applies the one-version policy when
// smith.toml asks for it. Nothing is written.
func Load(open func() (*session.Session, error)) Result {
	s, err := open()
	if err != nil {
		return Result{Err: err}
	}
	defer s.Close()
	report, err := dependencies.FindAvailable(s)
	if err != nil {
		return Result{Err: err}
	}
	var out Result
	for _, u := range report.Updates {
		out.Rows = append(out.Rows, Row{Name: u.Name, Plugin: u.Plugin, Current: u.Current, Latest: u.Latest, Level: u.Level, Projects: u.Projects()})
	}
	for _, f := range dependencies.CheckOneVersionPolicy(s) {
		out.Policy = append(out.Policy, PolicyLine{Check: f.Check, OK: f.OK, Warning: f.Warning, Detail: f.Detail, Remedy: f.Remedy})
	}
	return out
}
