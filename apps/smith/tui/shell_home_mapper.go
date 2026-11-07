package tui

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dashboard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
)

// maxRecentOnHome is how many runs of the history Home reads.
const maxRecentOnHome = 3

// homeContent is the Home page: the attention queue, the last runs and the
// buttons that add a project. A click on a button runs its command and a click
// on a run opens it; a click anywhere else does nothing.
func (s Shell) homeContent(w int) content {
	ptr := dashboard.Away
	if in := s.mainInner(); in.Contains(s.mouseX, s.mouseY) {
		ptr = dashboard.Pointer{Col: s.mouseX - in.X, Row: s.mouseY - in.Y}
	}
	page := dashboard.Page(s.th, s.dashboardFacts(), w, ptr)
	c := content{lines: page.Lines, hits: map[int]hit{}}
	for _, t := range page.Targets {
		h, ok := c.hits[t.Row]
		if !ok {
			h = hit{kind: "action"}
			if t.Command == "" {
				h.kind = "recent"
			}
		}
		h.spans = append(h.spans, hitSpan{from: t.From, to: t.To, id: t.Command, index: t.Run})
		c.hits[t.Row] = h
	}
	return c
}

// homeCommand is the id of the command Enter runs on Home: the first button of
// the first item of the queue, "" while there is none.
func (s Shell) homeCommand() string { return dashboard.Primary(dashboard.Queue(s.dashboardFacts())) }

// dashboardFacts gathers what the shell already knows; nothing is loaded here.
func (s Shell) dashboardFacts() dashboard.Facts {
	f := dashboard.Facts{ProjectsLoaded: s.loadedProjects && s.projectsErr == nil, Last: s.last, SettingsTip: s.settings.TipWanted(),
		Workspace: s.header, Branch: s.branch, Release: readyRelease(s.release)}
	if r := s.doctor.Report; r != nil {
		f.Checked, f.Failed, f.Warnings, f.Passed = true, r.Failed, r.Warnings, r.Passed
		f.Fixable = len(s.doctor.Plan())
		for _, found := range r.Findings {
			if found.OK {
				continue
			}
			note := strings.TrimSpace(max2(found.Problem, found.Check))
			if found.Warning {
				f.WarningNotes = append(f.WarningNotes, note)
				continue
			}
			f.Problems = append(f.Problems, note)
		}
	}
	f.Waited = s.waited()
	if s.doctor.Err != nil {
		f.CheckErr = s.doctor.Err.Error()
	}
	projects := s.list.Projects()
	f.Projects = len(projects)
	for _, p := range projects {
		f.Languages = append(f.Languages, p.Language)
	}
	for _, p := range projects {
		if p.Affected {
			f.Affected++
		}
	}
	if s.deps.Loaded && s.deps.Err == nil {
		f.DepsChecked, f.Outdated = true, len(s.deps.Rows)
	}
	now := s.now()
	for i := range s.runs {
		r := s.runs[len(s.runs)-1-i]
		if !r.Finished {
			continue
		}
		run := lastOf(r.Command, r.OK(), r.Cancelled, shortDuration(r.RunTime(now)))
		run.Index = i
		f.Recent = append(f.Recent, *run)
		if len(f.Recent) == maxRecentOnHome {
			break
		}
	}
	return f
}

// max2 is the first text, or the second when the first is empty.
func max2(first, second string) string {
	if first != "" {
		return first
	}
	return second
}

// readyRelease is the plan of the Release tab as Home tells it: how many
// packages leave and the one that has the most commits. It is nil until the
// plan is read, when it failed, when something blocks it and when it is empty.
func readyRelease(m releaseview.Model) *dashboard.Ready {
	if !m.CanRelease() {
		return nil
	}
	best := m.Plan.Items[0]
	for _, it := range m.Plan.Items[1:] {
		if commitsOf(it) > commitsOf(best) {
			best = it
		}
	}
	return &dashboard.Ready{Packages: len(m.Plan.Items), Name: best.Project, From: best.Current, To: best.Next, Commits: commitsOf(best)}
}

// commitsOf is how many commits release a package.
func commitsOf(it release.Item) int { return max(len(it.Commits), len(it.Changes)) }

// lastOf is how a command that ended is told on Home.
func lastOf(argv []string, ok, cancelled bool, took string) *dashboard.Run {
	return &dashboard.Run{Command: strings.TrimPrefix(commands.Line(argv), "smith "), OK: ok, Cancelled: cancelled, Took: took, Repeatable: repeatable(argv)} // words:literal
}
