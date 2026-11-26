package tui

import (
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/release"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/dashboard"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/releaseview"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/runhistory"
)

// maxRecentOnHome is how many runs of the history Home reads.
const maxRecentOnHome = 3

// homeContent is Home in a pane of w by h cells: the summary, the boxes, the
// list "Needs you now" and the buttons of Start something. A click on an item
// selects it, a click on a button runs its command and a click on the Last
// run box opens that run; a click anywhere else does nothing. A page taller
// than the pane scrolls (homeTop), and the first and last rows tell what is cut.
func (s Shell) homeContent(w, h int) content {
	ptr := dashboard.Away
	if in := s.mainInner(); in.Contains(s.mouseX, s.mouseY) {
		ptr = dashboard.Pointer{Col: s.mouseX - in.X, Row: s.mouseY - in.Y}
	}
	page := s.homePage(w, ptr)
	top := pane.ClampTop(s.homeTop, len(page.Lines), h)
	if top > 0 && ptr.Row >= 0 {
		// The pointer is over a window row; the page is drawn from the top row.
		ptr.Row += top
		page = s.homePage(w, ptr)
	}
	hits := map[int]hit{}
	for _, t := range page.Targets {
		h, ok := hits[t.Row]
		if !ok {
			h = hit{kind: homeHitKind(t)}
		}
		sp := hitSpan{from: t.From, to: t.To, id: t.Command, index: t.Item}
		if t.Kind == dashboard.TargetRun {
			sp.index = t.Run
		}
		h.spans = append(h.spans, sp)
		hits[t.Row] = h
	}
	rows := map[int]string{}
	for row := range hits {
		rows[row] = strconv.Itoa(row)
	}
	lines, shown := pane.WindowMarked(s.th, page.Lines, rows, top, h)
	c := content{lines: lines, hits: map[int]hit{}}
	for y, id := range shown {
		row, _ := strconv.Atoi(id)
		c.hits[y] = hits[row]
	}
	return c
}

// homeHitKind is what a row of Home holding target t answers a click with:
// "homeitem" for a line of an item of the list, "recent" for the Last run
// box and "action" for the buttons of Start something.
func homeHitKind(t dashboard.Target) string {
	switch {
	case t.Kind == dashboard.TargetRun:
		return "recent"
	case t.Item >= 0:
		return "homeitem"
	}
	return "action"
}

// homePage draws Home in w cells with the keyboard on the item homeSel of the
// list and its button homeBtn (no item while nothing can be pressed).
func (s Shell) homePage(w int, ptr dashboard.Pointer) dashboard.Rendered {
	facts := s.dashboardFacts()
	ch := dashboard.Choice{Item: -1}
	if dashboard.Primary(dashboard.Queue(facts)) != "" {
		ch = dashboard.Choice{Item: max(s.homeSel, 0), Button: max(s.homeBtn, 0)}
	}
	return dashboard.Render(s.th, facts, w, ptr, ch)
}

// homeCommand is the id of the command Enter runs on Home: the first button of
// the first item of the queue, "" while there is none.
func (s Shell) homeCommand() string { return dashboard.Primary(dashboard.Queue(s.dashboardFacts())) }

// homeChosen is the button Enter presses on Home: the button homeBtn of the
// item homeSel, both kept inside the queue. It is false while there is none.
func (s Shell) homeChosen() (dashboard.Button, bool) {
	q := dashboard.Queue(s.dashboardFacts())
	if dashboard.Primary(q) == "" {
		return dashboard.Button{}, false
	}
	it := q[min(max(s.homeSel, 0), len(q)-1)]
	if len(it.Buttons) == 0 {
		return dashboard.Button{}, false
	}
	return it.Buttons[min(max(s.homeBtn, 0), len(it.Buttons)-1)], true
}

// homeLabels name, for the footer, the action Enter does on Home: "fix".
func (s Shell) homeLabels() map[string]string {
	if s.tab != commands.TabHome {
		return nil
	}
	b, ok := s.homeChosen()
	if !ok {
		return nil
	}
	r := []rune(b.Label)
	return map[string]string{"home.next": strings.ToLower(string(r[:1])) + string(r[1:])}
}

// dashboardFacts gathers what the shell already knows; nothing is loaded here.
func (s Shell) dashboardFacts() dashboard.Facts {
	f := dashboard.Facts{ProjectsLoaded: s.loadedProjects && s.projectsErr == nil, Last: s.last, SettingsTip: s.settings.TipWanted(), CompletionOffer: s.completionWanted,
		Release: readyRelease(s.release), Remote: s.remoteFacts(), Workspace: s.header, Base: s.base, ReleaseLoading: s.releaseLoading, Tick: s.tick}
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
		f.DepsChecked, f.Outdated = true, len(s.deps.All)
	}
	now := s.now()
	for i := range s.runs {
		r := s.runs[len(s.runs)-1-i]
		if !r.Finished {
			continue
		}
		run := lastOf(r.Command, r.OK(), r.Cancelled, shortDuration(r.RunTime(now)))
		run.Index = i
		counts := r.Counts()
		run.Passed, run.Cached, run.Failed = counts.Passed, counts.Cached, counts.Failed
		if !r.Ended.IsZero() {
			run.Ago = runhistory.Ago(now.Sub(r.Ended))
		}
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
	ready := &dashboard.Ready{Packages: len(m.Plan.Items), Name: best.Project, From: best.Current, To: best.Next, Commits: commitsOf(best)}
	for _, it := range m.Plan.Items {
		switch it.Level {
		case "major":
			ready.Major++
		case "minor":
			ready.Minor++
		case "patch":
			ready.Patch++
		}
	}
	return ready
}

// commitsOf is how many commits release a package.
func commitsOf(it release.Item) int { return max(len(it.Commits), len(it.Changes)) }

// lastOf is how a command that ended is told on Home.
func lastOf(argv []string, ok, cancelled bool, took string) *dashboard.Run {
	return &dashboard.Run{Command: strings.TrimPrefix(commands.Line(argv), "smith "), OK: ok, Cancelled: cancelled, Took: took, Repeatable: repeatable(argv)} // words:literal
}
