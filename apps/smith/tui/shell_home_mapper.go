package tui

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dashboard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

// homeContent is the Home page: the headline, the next action (a click on it
// does it) and the facts behind them.
func (s Shell) homeContent(w int) content {
	lines, row, id := dashboard.Lines(s.th, s.dashboardFacts(), w)
	c := content{lines: lines, hits: map[int]hit{}}
	if id == "" {
		return c
	}
	// The action may wrap onto several lines: all of them are the click target.
	for i := row; i < len(lines) && strings.TrimSpace(pane.Strip(lines[i])) != ""; i++ {
		c.hits[i] = hit{kind: "action", id: id}
	}
	return c
}

// nextAction is what Home suggests now.
func (s Shell) nextAction() dashboard.Next { return dashboard.NextAction(s.dashboardFacts()) }

// dashboardFacts gathers what the shell already knows; nothing is loaded here.
func (s Shell) dashboardFacts() dashboard.Facts {
	f := dashboard.Facts{ProjectsLoaded: s.loadedProjects && s.projectsErr == nil, Last: s.last, SettingsTip: s.settings.TipWanted()}
	if r := s.doctor.Report; r != nil {
		f.Checked, f.Failed, f.Warnings, f.Passed = true, r.Failed, r.Warnings, r.Passed
		f.Fixable = len(s.doctor.Plan())
	}
	f.Waited = s.waited()
	if s.doctor.Err != nil {
		f.CheckErr = s.doctor.Err.Error()
	}
	projects := s.list.Projects()
	f.Projects = len(projects)
	for _, p := range projects {
		if p.Affected {
			f.Affected++
		}
	}
	if s.deps.Loaded && s.deps.Err == nil {
		f.DepsChecked, f.Outdated = true, len(s.deps.Rows)
	}
	return f
}

// lastOf is how a command that ended is told on Home.
func lastOf(argv []string, ok, cancelled bool, took string) *dashboard.Run {
	return &dashboard.Run{Command: strings.TrimPrefix(commands.Line(argv), "smith "), OK: ok, Cancelled: cancelled, Took: took}
}
