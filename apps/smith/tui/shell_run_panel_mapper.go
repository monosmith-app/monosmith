package tui

import (
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/logview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/splitter"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The run panel is the right panel of Projects while a run goes, or after the
// user opened a past run. From the top it has the header (what runs and n/m),
// the buttons, the history header, the command line, the progress bar, the task
// rows, the rule of the log and the log.

// runGeom is where the parts of the run panel sit, in lines of its body.
type runGeom struct {
	headRow, ctlRows        int
	navRow, cmdRow, progRow int
	listTop, listRows       int
	sepRow                  int
	// infoRows are the rows under the rule that tell the task whose log is shown:
	// its command, and what its row had no room for.
	infoRows        int
	logTop, logRows int
}

// runGeom measures the run panel in w by h cells.
func (s Shell) runGeom(w, h int) runGeom {
	g := runGeom{}
	rows, _ := s.runControls(w)
	g.headRow, g.ctlRows = 0, len(rows)
	g.navRow = 1 + g.ctlRows
	g.cmdRow, g.progRow, g.listTop = g.navRow+1, g.navRow+2, g.navRow+3
	n := len(s.runIDs())
	avail := max(h-g.listTop-2, 1)
	g.listRows = min(n, max(min(avail, 2), avail*2/5))
	if given, ok := s.split.Size(splitter.RunListID); ok {
		g.listRows = splitter.RunListRows(given, avail)
	}
	g.sepRow = g.listTop + g.listRows
	g.logTop = g.sepRow + 1
	g.infoRows = len(s.taskInfo(w))
	g.logTop += g.infoRows
	g.logRows = max(h-g.logTop, 1)
	return g
}

// runRect is where the run panel is drawn: the right panel, or the whole pane
// when the panel fills it.
func (s Shell) runRect() (w, h int) {
	r := s.detailRect()
	return r.W, r.H
}

// logTotal is how many rows the log shown takes at the width of the run panel.
func (s Shell) logTotal() int {
	w, _ := s.runRect()
	return len(logview.Rows(s.th, s.logLines(), w))
}

// logHeight is the number of log lines visible in the run panel.
func (s Shell) logHeight() int {
	w, h := s.runRect()
	return s.runGeom(w, h).logRows
}

// keepRunVisible scrolls the task list so the selected task shows.
func (s *Shell) keepRunVisible() {
	if s.run == nil {
		return
	}
	w, h := s.runRect()
	g := s.runGeom(w, h)
	at := 0
	for i, id := range s.runIDs() {
		if id == s.runSel {
			at = i
		}
	}
	if at < s.runTop {
		s.runTop = at
	}
	if at >= s.runTop+g.listRows {
		s.runTop = at - g.listRows + 1
	}
	s.runTop = max(min(s.runTop, len(s.runIDs())-g.listRows), 0)
}

// runButton is a button of the run panel: its action, its label and its key.
type runButton struct {
	id, label, key string
	disabled       bool
}

// runButtons are the buttons of the run panel. Back and Stop are always there
// (Stop is faint once nothing runs); the others appear when they can be used,
// as the footer does for keys; Copy log is faint while the log is empty.
func (s Shell) runButtons() []runButton {
	out := []runButton{
		{"x:run-back", words.Say("tui.run_panel.back"), commands.ShortcutOf("run.all-output"), false},
		{"run.stop", words.Say("tui.run_panel.stop"), commands.ShortcutOf("run.stop"), !s.running()},
	}
	if !s.running() && s.repeatArgv() != nil {
		out = append(out, runButton{"x:rerun", words.Say("tui.run_panel.again"), commands.ShortcutOf("run.again"), false})
	}
	if !s.running() && s.failedArgv() != nil {
		out = append(out, runButton{"x:rerun-failed", words.Say("tui.run_panel.unfinished"), commands.ShortcutOf("run.again-failed"), false})
	}
	// Copy log stays when the log shown is empty (faint): a button that comes and
	// goes with the selected task moves the rows under the pointer.
	out = append(out, runButton{"x:copy-log", words.Say("tui.run_panel.copy"), commands.ShortcutOf("run.copy-log"), len(s.logLines()) == 0})
	return out
}

// runControls draws the buttons in rows of w cells, a button that does not fit
// starting a new row. Each row answers a click with the span of its buttons.
func (s Shell) runControls(w int) (rows []string, hits []hit) {
	if s.run == nil {
		return nil, nil
	}
	under := s.actionAt(s.mouseX, s.mouseY)
	line, col := " ", 1
	h := hit{kind: "action"}
	flush := func() {
		if len(h.spans) > 0 {
			rows, hits = append(rows, pane.Fit(line, w)), append(hits, h)
		}
		line, col, h = " ", 1, hit{kind: "action"}
	}
	for _, b := range s.runButtons() {
		bw := button.Width(s.th, b.label, b.key)
		if col > 1 && col+bw > w {
			flush()
		}
		state := s.buttonState(b.id, under)
		if b.disabled {
			state = button.Disabled
		}
		if col > 1 {
			line, col = line+" ", col+1
		}
		line += button.Render(s.th, b.label, b.key, state)
		h.spans = append(h.spans, hitSpan{from: col, to: col + bw, id: b.id})
		col += bw
	}
	flush()
	return rows, hits
}

// runHeader is the first row of the run panel: what runs, then the spinner (or
// the mark of how it ended) and n/m.
func (s Shell) runHeader(w int, c runview.Counts) string {
	title := commands.RunTitle(s.run.Command, s.list.Len())
	if title == "" {
		title = s.run.CommandLine()
	}
	mark := s.th.Info(s.th.Spin(s.tick))
	switch {
	case s.stopping && s.running():
		mark = s.th.Warning(s.th.Spin(s.tick))
	case s.running():
	case s.run.Cancelled:
		mark = s.th.Muted(s.th.Icons.Cancelled)
	case s.run.OK():
		mark = s.th.Success(s.th.Icons.Passed)
	default:
		mark = s.th.Failure(s.th.Icons.Failed)
	}
	tail := "  " + mark
	if c.Total > 0 {
		tail += " " + fmtCount(c.Done, c.Total)
	}
	room := max(w-1-pane.Width(tail), 4)
	return pane.Fit(" "+s.th.Title(pane.Fit(title, room))+tail, w)
}

// runPanelContent draws the run in w by h cells: header, buttons, history
// header, command line, progress, tasks and log, with the thing each row stands for.
func (s Shell) runPanelContent(w, h int) content {
	c := content{hits: map[int]hit{}}
	if s.run == nil {
		return c
	}
	g := s.runGeom(w, h)
	counts := s.run.Counts()
	now := s.now()
	took := shortDuration(s.run.RunTime(now))
	state := s.th.Info(words.Say("tui.main.running", "spin", s.th.Spin(s.tick)))
	switch {
	case s.stopping && s.running():
		state = s.th.Warning(words.Say("tui.main.stopping", "spin", s.th.Spin(s.tick)))
	case s.run.Cancelled:
		state = s.th.Muted(words.Say("tui.main.cancelled", "cancelled", s.th.Icons.Cancelled))
	case s.run.Finished && s.run.OK():
		state = s.th.Success(words.Say("tui.main.passed", "passed", s.th.Icons.Passed))
	case s.run.Finished:
		state = s.th.Failure(words.Say("tui.main.failed", "failed", s.th.Icons.Failed))
	}
	c.lines = append(c.lines, s.runHeader(w, counts))
	ctl, ctlHits := s.runControls(w)
	for i, line := range ctl {
		c.hits[len(c.lines)] = ctlHits[i]
		c.lines = append(c.lines, line)
	}
	c.hits[len(c.lines)] = hit{kind: "nav"}
	c.lines = append(c.lines, s.navLine(w))
	c.lines = append(c.lines, pane.Fit(" "+s.th.Faint("$")+" "+s.th.Bold(s.run.CommandLine())+"   "+state+"  "+s.th.Muted(took), w))
	if counts.Total > 0 {
		bar := progress.Draw(s.th, counts.Done, counts.Total, min(max(w/2, 20), 40), counts.Failed > 0)
		if s.run.Cancelled {
			// A stopped run is not complete: the bar counts what ended and what was left.
			bar = progress.DrawCancelled(s.th, counts.Passed+counts.Cached+counts.Failed, counts.Skipped+counts.Cancelled, counts.Total, min(max(w/2, 20), 40))
		}
		c.lines = append(c.lines, pane.Fit(" "+bar+"  "+s.countsText(counts), w))
	} else if name, ok := addedProject(s.run.Command); ok && s.run.Finished && s.run.OK() {
		c.lines = append(c.lines, pane.Fit(addedNextSteps(s.th, name), w))
	} else {
		c.lines = append(c.lines, "")
	}
	ids := s.runIDs()
	for i := s.runTop; i < len(ids) && i < s.runTop+g.listRows; i++ {
		c.hits[len(c.lines)] = hit{kind: "task", id: ids[i], index: i}
		c.lines = append(c.lines, s.taskLine(ids[i], w, now))
	}
	for len(c.lines) < g.sepRow {
		c.lines = append(c.lines, "")
	}
	copy(c.lines[g.listTop:g.sepRow], pane.Scrollbar(s.th, c.lines[g.listTop:g.sepRow], w, s.runTop, len(ids)))
	title := words.Say("tui.main.all_output")
	if s.runSel != "" {
		title = s.runSel
	}
	follow := "following"
	if !s.log.Following() {
		follow = "paused"
	}
	sep := words.Say("tui.main.log", "title", title, "follow", follow)
	c.lines = append(c.lines, s.th.Faint(pane.Fit(sep+strings.Repeat("─", max(w-pane.Width(sep), 0)), w)))
	c.lines = append(c.lines, s.taskInfo(w)...)
	c.lines = append(c.lines, s.log.Render(s.th, logview.Rows(s.th, s.logLines(), w), w, g.logRows)...)
	return c
}

// countsText is "✓ 4  ≈ 1  × 1  » 2  ● 2  ○ 3" without the zeros.
func (s Shell) countsText(c runview.Counts) string {
	var parts []string
	add := func(n int, st runview.Status) {
		if n > 0 {
			parts = append(parts, words.Say("tui.main.text_2", "statusIcon", s.statusIcon(st), "n", n, "st", st))
		}
	}
	add(c.Failed, runview.Failed)
	add(c.Running, runview.Running)
	add(c.Passed, runview.Passed)
	add(c.Cached, runview.Cached)
	add(c.Skipped, runview.Skipped)
	add(c.Cancelled, runview.Cancelled)
	add(c.Queued, runview.Queued)
	return strings.Join(parts, "  ")
}

// taskLine is one row of the task list.
func (s Shell) taskLine(id string, w int, now time.Time) string {
	if id == "" {
		line := words.Say("tui.main.all_output_2", "pointer", s.th.Icons.Pointer, "say", s.th.Muted(words.Say("tui.main.lines", "len", len(s.run.Lines))))
		if s.runSel == "" {
			return s.th.Selected(pane.Fit(pane.Strip(line), w))
		}
		return pane.Fit(line, w)
	}
	t := s.run.Task(id)
	if t == nil {
		return ""
	}
	line := " " + s.statusIcon(t.Status) + " " + pad(t.ID, taskNameW(w)) + " " + s.th.Muted(s.taskDetail(t, now))
	if id == s.runSel {
		line = s.th.Selected(pane.Fit(pane.Strip(line), w))
	}
	return pane.Fit(line, w)
}

// taskDetail is what a task row says after the name: the state, the time and the
// reason, and why a slow lint takes time.
func (s Shell) taskDetail(t *runview.Task, now time.Time) string {
	state, took, reason := s.taskFacts(t, now)
	detail := state
	if took != "" {
		detail += "  " + took
	}
	if reason != "" {
		detail += "  " + reason
	}
	if t.Status == runview.Running && strings.HasSuffix(t.ID, ":lint") && s.run.Elapsed(t, now) >= slowLintAfter {
		detail += words.Say("tui.main.lint_checks_every_file_it_is")
	}
	return detail
}

// taskInfo tells the task whose log is shown, under the rule of the log: the
// command it ran, and its state, time and reason when its row in the list was
// too narrow to say them. It is empty while all output shows.
func (s Shell) taskInfo(w int) []string {
	if s.run == nil || s.runSel == "" {
		return nil
	}
	t := s.run.Task(s.runSel)
	if t == nil {
		return nil
	}
	var out []string
	if t.Command != "" {
		out = append(out, pane.Fit(" "+s.th.Faint("$")+" "+s.th.Muted(t.Command), w))
	}
	now := s.now()
	detail := s.taskDetail(t, now)
	// The row is the space, the mark, the space, the name column, the space and
	// the detail; the last cell is the scrollbar's.
	if 1+pane.Width(s.statusIcon(t.Status))+1+taskNameW(w)+1+pane.Width(detail) > w-1 {
		for _, part := range pane.Wrap(s.th.Muted(detail), max(w-2, 4), 0) {
			out = append(out, pane.Fit(" "+part, w))
		}
	}
	return out
}

// taskNameW is the width of the name column of the task list in w cells.
func taskNameW(w int) int { return min(max(w/3, 12), 28) }

// taskFacts are the words a task row tells after its name: the state, the time
// when there is one, and the reason unless the task passed.
func (s Shell) taskFacts(t *runview.Task, now time.Time) (state, took, reason string) {
	state = t.Status.String()
	if d := s.run.Elapsed(t, now); d > 0 {
		took = shortDuration(d)
	}
	if t.Reason != "" && t.Status != runview.Passed {
		reason = t.Reason
	}
	return state, took, reason
}

// slowLintAfter is how long a lint task runs before the list says why it takes
// time.
const slowLintAfter = 5 * time.Second
