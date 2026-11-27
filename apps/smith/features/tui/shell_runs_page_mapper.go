package tui

import (
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/logview"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/runhistory"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/runpage"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/runview"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/splitter"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The Runs view (key 3) shows the run going now, or the one opened from the
// runs of the session. From the top it has the header (what runs, the time, the
// buttons, the runs of the session to step through, the progress bar and the
// counts), a rule, then on a wide screen the tree of the tasks grouped by
// project at the left and the log of the selected task at the right, divided by
// a line the mouse drags. A narrow screen shows the tree or the log, switched
// with Tab. Before anything ran it says how to start a run.

// runsGeom is where the parts of the Runs view sit, in rows and cells of its body.
type runsGeom struct {
	// narrow is true when the tree and the log take turns (Tab); onLog when the
	// log is the one shown.
	narrow, onLog bool
	// headRows are the rows above the rule; ruleRow is the rule's row.
	headRows, ruleRow int
	// bodyTop and bodyRows are the rows under the rule.
	bodyTop, bodyRows int
	// treeW is the cells of the tree (all of them on a narrow screen); the
	// divider is in column treeW on a wide one.
	treeW int
	// treeRows are the rows of the tree window, above the runs of the session.
	treeRows int
	// logX and logW are the column and the cells of the log; logTop is its
	// first row and logRows its rows, under its title and the task's facts.
	logX, logW, logTop, logRows int
}

// runsNarrow is true when the Runs view shows the tree or the log, one at a
// time: a narrow screen. The full-screen run has the room for both.
func (s Shell) runsNarrow() bool { return s.lay().Mode == layout.Narrow && !s.fullRun() }

// runsTreeWidth is the cells the tree takes beside the log in w cells: what the
// person dragged it to, else a third of the width between 26 and 44.
func (s Shell) runsTreeWidth(w int) int {
	lo, hi := runsTreeMin, max(w-runsLogMin, runsTreeMin)
	if n, ok := s.split.Size(splitter.RunsTreeID); ok {
		return pane.Clamp(n, lo, hi)
	}
	return pane.Clamp(w/3, lo, min(44, hi))
}

// The least cells of the tree and of the log beside it.
const (
	runsTreeMin = 24
	runsLogMin  = 30
)

// runsGeom measures the Runs view in w by h cells.
func (s Shell) runsGeom(w, h int) runsGeom {
	g := runsGeom{narrow: s.runsNarrow()}
	// Only the count of the header's lines counts here: the copy draws it without
	// asking what is under the pointer, which would draw the view again.
	probe := s
	probe.probing = true
	head, _ := probe.runsHead(w)
	g.headRows = min(len(head), max(h-3, 1))
	g.ruleRow = g.headRows
	g.bodyTop = g.ruleRow + 1
	g.bodyRows = max(h-g.bodyTop, 1)
	session := len(s.sessionLines(w))
	g.treeW, g.logX, g.logW = w, 0, w
	g.onLog = g.narrow && (s.runOnLog || s.runHasNoTasks())
	if !g.narrow {
		g.treeW = s.runsTreeWidth(w)
		g.logX = g.treeW + 2
		g.logW = max(w-g.logX, 1)
	}
	if session > g.bodyRows/2 {
		session = 0
	}
	g.treeRows = max(g.bodyRows-session, 1)
	info := len(s.taskInfo(g.logW))
	g.logTop = g.bodyTop + 1 + info
	g.logRows = max(h-g.logTop, 1)
	return g
}

// runHasNoTasks is true for a run whose command runs no task (smith add): its
// output is all there is to show.
func (s Shell) runHasNoTasks() bool { return s.run == nil || len(s.run.Tasks) == 0 }

// runRect is the size of the body the Runs view is drawn in.
func (s Shell) runRect() (w, h int) {
	in := s.mainInner()
	return in.W, in.H - s.toastRows(in.H)
}

// logTotal is how many rows the log shown takes at the width of its column.
func (s Shell) logTotal() int {
	w, h := s.runRect()
	return len(logview.Rows(s.th, s.logLines(), s.runsGeom(w, h).logW-logPad))
}

// logHeight is the number of log lines visible.
func (s Shell) logHeight() int {
	w, h := s.runRect()
	return s.runsGeom(w, h).logRows
}

// runTree is the tree of the run shown, the failed tasks only while Errors
// only is on.
func (s Shell) runTree() []runpage.Row {
	if s.run == nil {
		return nil
	}
	// The tasks in the order they were first seen (the plan's order), so a
	// project keeps its place in the tree while its tasks change state.
	var refs []runpage.TaskRef
	for _, t := range s.run.Tasks {
		refs = append(refs, runpage.TaskRef{ID: t.ID, Failed: t.Status == runview.Failed})
	}
	return runpage.Tree(refs, s.runErrors)
}

// treeLine is one line of the tree's window: "All output" (all), a heading or a
// task (row), or a note that says why the tree is empty.
type treeLine struct {
	all  bool
	row  runpage.Row
	note string
}

// treeLines are the lines of the tree: All output first, then the projects
// and their tasks, or a note when there is no task to list.
func (s Shell) treeLines() []treeLine {
	out := []treeLine{{all: true}}
	rows := s.runTree()
	for _, r := range rows {
		out = append(out, treeLine{row: r})
	}
	switch {
	case len(rows) > 0:
	case s.runErrors && !s.runHasNoTasks():
		out = append(out, treeLine{note: words.Say("runpage.tree.no_errors")})
	default:
		out = append(out, treeLine{note: words.Say("runpage.tree.no_tasks")})
	}
	return out
}

// keepRunVisible scrolls the tree so the selected task shows, with the heading
// of its project when there is room.
func (s *Shell) keepRunVisible() {
	if s.run == nil {
		return
	}
	w, h := s.runRect()
	g := s.runsGeom(w, h)
	lines := s.treeLines()
	at := 0
	for i, l := range lines {
		if !l.all && l.note == "" && l.row.TaskID == s.runSel && s.runSel != "" {
			at = i
		}
	}
	first := at
	if at > 0 && lines[at-1].row.Heading() && lines[at-1].note == "" && !lines[at-1].all {
		first = at - 1
	}
	if first < s.runTop {
		s.runTop = first
	}
	if at >= s.runTop+g.treeRows {
		s.runTop = at - g.treeRows + 1
	}
	s.runTop = max(min(s.runTop, len(lines)-g.treeRows), 0)
}

// runButton is a button of the Runs view: its action, its label and its key.
type runButton struct {
	id, label, key string
	disabled       bool
}

// runButtons are the buttons of the Runs view. Back and Stop are always there
// (Stop is faint once nothing runs); the others appear when they can be used,
// as the footer does for keys; Copy log is faint while the log is empty. The
// last one makes the output fill the terminal; the full-screen view has its own
// button in the top bar instead. A narrow screen keeps only what answers the
// run now (Stop while it goes, Run again and Run unfinished after it): the
// footer has the other keys.
func (s Shell) runButtons() []runButton {
	all := s.allRunButtons()
	if !s.runsNarrow() {
		return all
	}
	var out []runButton
	for _, b := range all {
		switch {
		case b.id == "run.stop" && s.running(), b.id == "x:rerun", b.id == "x:rerun-failed":
			out = append(out, b)
		}
	}
	return out
}

// allRunButtons are the buttons of a wide Runs view.
func (s Shell) allRunButtons() []runButton {
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
	if !s.fullRun() {
		out = append(out, runButton{"x:run-fullscreen", words.Say("tui.run_full.enlarge"), commands.ShortcutOf("run.fullscreen"), false})
	}
	return out
}

// runButtonRow draws the buttons on one line, one cell between two: the line,
// the span of each button from column 0 and the cells it takes.
func (s Shell) runButtonRow() (string, []hitSpan, int) {
	under := s.actionAt(s.mouseX, s.mouseY)
	var b strings.Builder
	var spans []hitSpan
	col := 0
	for i, btn := range s.runButtons() {
		if i > 0 {
			b.WriteString(" ")
			col++
		}
		state := s.buttonState(btn.id, under)
		if btn.disabled {
			state = button.Disabled
		}
		bw := button.Width(s.th, btn.label, btn.key)
		b.WriteString(button.Render(s.th, btn.label, btn.key, state))
		spans = append(spans, hitSpan{from: col, to: col + bw, id: btn.id})
		col += bw
	}
	return b.String(), spans, col
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

// runState is the first words of the header: the mark (the spinner while the
// run goes) and what it does ("Running 18 tasks", "Failed 3 tasks").
func (s Shell) runState(c runview.Counts) string {
	what := commands.RunTitle(s.run.Command, s.list.Len())
	if c.Total > 0 {
		what = words.Count("runpage.state.tasks", c.Total)
	} else if what == "" {
		what = s.run.CommandLine()
	}
	switch {
	case s.stopping && s.running():
		return s.th.Warning(s.th.Spin(s.tick)) + " " + s.th.Title(words.Say("runpage.state.stopping", "what", what))
	case s.running():
		return s.th.Info(s.th.Spin(s.tick)) + " " + s.th.Title(words.Say("runpage.state.running", "what", what))
	case s.run.Cancelled:
		return s.th.Muted(s.th.Icons.Cancelled) + " " + s.th.Title(words.Say("runpage.state.stopped", "what", what))
	case s.run.OK():
		return s.th.Success(s.th.Icons.Passed) + " " + s.th.Title(words.Say("runpage.state.passed", "what", what))
	}
	return s.th.Failure(s.th.Icons.Failed) + " " + s.th.Title(words.Say("runpage.state.failed", "what", what))
}

// runsHead is the header of the Runs view in w cells, with what a click on each
// line means: the state and the time, with the buttons at the right when they
// fit (else on the lines under it), the runs of the session to step through,
// the progress bar and the counts, a blank line and, on a narrow screen, the
// switch between the tasks and the log.
func (s Shell) runsHead(w int) ([]string, map[int]hit) {
	var lines []string
	hits := map[int]hit{}
	if s.run == nil {
		return nil, hits
	}
	c := s.run.Counts()
	state := " " + s.runState(c)
	if took := progress.Elapsed(s.run.RunTime(s.now())); took != "" {
		state += "  " + s.th.Muted(took)
	}
	if left, ok := s.timeLeft(); ok {
		state += s.th.Muted(words.Say("runpage.estimate", "time", left))
	}
	var ctl []string
	var ctlHits []hit
	if row, spans, bw := s.runButtonRow(); pane.Width(state)+2+bw <= w {
		// The buttons fit beside the state: they go at its right end.
		at := w - bw
		h := hit{kind: "action"}
		for _, sp := range spans {
			h.spans = append(h.spans, hitSpan{from: sp.from + at, to: sp.to + at, id: sp.id})
		}
		if len(h.spans) > 0 {
			hits[0] = h
		}
		lines = append(lines, pane.Fit(pane.Fit(state, at)+row, w))
	} else {
		lines = append(lines, pane.Fit(state, w))
		ctl, ctlHits = s.runControls(w)
	}
	for i, line := range ctl {
		hits[len(lines)] = ctlHits[i]
		lines = append(lines, line)
	}
	if len(s.runs) > 1 {
		hits[len(lines)] = hit{kind: "nav"}
		lines = append(lines, s.navLine(w))
	}
	lines = append(lines, "")
	switch {
	case c.Total > 0 && s.run.Cancelled:
		// A stopped run is not complete: the bar counts what ended and what was left.
		bar := progress.DrawCancelled(s.th, c.Passed+c.Cached+c.Failed, c.Skipped+c.Cancelled, c.Total, min(w-2, 96))
		lines = append(lines, pane.Fit(" "+bar, w), pane.Fit(" "+s.runCounts(c), w))
	case c.Total > 0:
		lines = append(lines, " "+runpage.Bar(s.th, c.Done, c.Total, c.Failed > 0, min(w-2, 96)))
		lines = append(lines, pane.Fit(" "+s.runCounts(c), w))
	default:
		if name, ok := addedProject(s.run.Command); ok && s.run.Finished && s.run.OK() {
			lines = append(lines, pane.Fit(addedNextSteps(s.th, name), w))
		}
	}
	if s.runsNarrow() {
		lines = append(lines, "")
		line, spans := runpage.Switch(s.th, s.runSel, s.runOnLog || s.runHasNoTasks(), w)
		h := hit{kind: "action"}
		for _, sp := range spans {
			h.spans = append(h.spans, hitSpan{from: sp.From, to: sp.To, id: sp.ID})
		}
		hits[len(lines)] = h
		lines = append(lines, line)
	}
	return lines, hits
}

// runCounts is "× 1 failed   ✓ 9 passed   ≈ 2 cached   ● 2 running   ○ 4
// queued" without the zeros, each count with its mark.
func (s Shell) runCounts(c runview.Counts) string {
	var parts []string
	add := func(n int, st runview.Status, key string) {
		if n > 0 {
			parts = append(parts, words.Say(key, "mark", s.statusIcon(st), "n", n))
		}
	}
	add(c.Failed, runview.Failed, "runpage.counts.failed")
	add(c.Passed, runview.Passed, "runpage.counts.passed")
	add(c.Cached, runview.Cached, "runpage.counts.cached")
	add(c.Running, runview.Running, "runpage.counts.running")
	add(c.Queued, runview.Queued, "runpage.counts.queued")
	add(c.Skipped, runview.Skipped, "runpage.counts.skipped")
	add(c.Cancelled, runview.Cancelled, "runpage.counts.cancelled")
	return strings.Join(parts, "   ")
}

// runsContent draws the Runs view in w by h cells, with the thing each row
// stands for.
func (s Shell) runsContent(w, h int) content {
	c := content{hits: map[int]hit{}}
	if s.run == nil {
		c.lines = runpage.Empty(s.th, runpage.Keys{
			Build: commands.ShortcutOf("run.build"), Test: commands.ShortcutOf("run.test"),
			Palette: commands.ShortcutOf("file.palette"), Projects: commands.ShortcutOf("view.projects"),
		}, w)
		return c
	}
	g := s.runsGeom(w, h)
	head, hits := s.runsHead(w)
	for i := 0; i < g.headRows && i < len(head); i++ {
		c.lines = append(c.lines, head[i])
		if hh, ok := hits[i]; ok {
			c.hits[i] = hh
		}
	}
	for len(c.lines) < g.ruleRow {
		c.lines = append(c.lines, "")
	}
	c.lines = append(c.lines, s.th.Border(strings.Repeat("─", w), s.mainFocused()))
	left, leftHits := s.treeColumn(g, w)
	if g.narrow {
		col := left
		if g.onLog {
			col, leftHits = s.logColumn(g, w), nil
		}
		for i := 0; i < g.bodyRows; i++ {
			line := ""
			if i < len(col) {
				line = col[i]
			}
			if hh, ok := leftHits[i]; ok {
				c.hits[g.bodyTop+i] = hh
			}
			c.lines = append(c.lines, pane.Fit(line, w))
		}
		return c
	}
	right := s.logColumn(g, g.logW)
	divider := s.th.Border("│", true)
	for i := 0; i < g.bodyRows; i++ {
		l, r := "", ""
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		if hh, ok := leftHits[i]; ok {
			c.hits[g.bodyTop+i] = hh
		}
		c.lines = append(c.lines, pane.Fit(l, g.treeW)+divider+" "+pane.Fit(r, g.logW))
	}
	return c
}

// treeColumn is the tree's column, w cells wide on a narrow screen and treeW on
// a wide one: the window of the tree, then the runs of the session. A click on
// a task shows its log; one on a run of the session opens that run.
func (s Shell) treeColumn(g runsGeom, w int) ([]string, map[int]hit) {
	tw := g.treeW
	if g.narrow {
		tw = w
	}
	lines := s.treeLines()
	now := s.now()
	var out []string
	hits := map[int]hit{}
	for i := s.runTop; i < len(lines) && len(out) < g.treeRows; i++ {
		l := lines[i]
		row := len(out)
		switch {
		case l.all:
			out = append(out, runpage.AllOutputRow(s.th, len(s.run.Lines), s.runSel == "", tw))
			hits[row] = hit{kind: "task", spans: []hitSpan{{from: 0, to: tw, id: ""}}}
		case l.note != "":
			for _, part := range pane.Wrap(s.th.Muted(l.note), max(tw-2, 4), 0) {
				out = append(out, pane.Fit("  "+part, tw))
			}
		case l.row.Heading():
			out = append(out, runpage.HeadingLine(s.th, l.row.Project, tw))
		default:
			t := s.run.Task(l.row.TaskID)
			if t == nil {
				continue
			}
			out = append(out, runpage.TaskRow(s.th, l.row, s.treeTaskLine(t, now), tw))
			hits[row] = hit{kind: "task", spans: []hitSpan{{from: 0, to: tw, id: t.ID}}}
		}
	}
	out = out[:min(len(out), g.treeRows)]
	if len(lines) > g.treeRows {
		out = pane.Scrollbar(s.th, pane.Block(out, tw, g.treeRows), tw, s.runTop, len(lines))
	}
	for len(out) < g.treeRows {
		out = append(out, "")
	}
	session := s.sessionLines(tw)
	if g.treeRows+len(session) <= g.bodyRows {
		first := len(out)
		out = append(out, session...)
		for i, at := range s.sessionRuns() {
			hits[first+2+i] = hit{kind: "recent", index: len(s.runs) - 1 - at}
		}
	}
	return out, hits
}

// treeTaskLine is what the row of task t tells: its mark, its time and whether
// the cache gave its result.
func (s Shell) treeTaskLine(t *runview.Task, now time.Time) runpage.TaskLine {
	l := runpage.TaskLine{Mark: s.statusIcon(t.Status), Selected: t.ID == s.runSel}
	if d := s.run.Elapsed(t, now); d > 0 {
		l.Took = shortDuration(d)
	}
	if t.Status == runview.Cached {
		l.Cache = words.Say("runpage.cache.hit")
		if t.Remote {
			l.Cache = words.Say("runpage.cache.remote")
		}
	}
	return l
}

// maxSessionRows is the most runs of the session the tree's column lists.
const maxSessionRows = 5

// sessionRuns are the positions in s.runs of the runs the column lists, the
// newest first; none while the session has only the one run.
func (s Shell) sessionRuns() []int {
	if len(s.runs) < 2 {
		return nil
	}
	var at []int
	for i := len(s.runs) - 1; i >= 0 && len(at) < maxSessionRows; i-- {
		at = append(at, i)
	}
	return at
}

// sessionLines are the runs of the session under the tree: a blank line, the
// heading, then one row each (mark and command, result, time, how long ago),
// the one shown with the cursor mark.
func (s Shell) sessionLines(w int) []string {
	runs := s.sessionRuns()
	if len(runs) == 0 {
		return nil
	}
	out := []string{"", pane.Fit(" "+s.th.Bold(words.Say("runpage.tree.session")), w)}
	now := s.now()
	for _, at := range runs {
		r := s.runs[at]
		st := runOutcome(r)
		ago := words.Say("tui.projects.ago.now")
		if r.Finished {
			ago = runhistory.Ago(now.Sub(r.Ended))
		}
		out = append(out, runpage.SessionRow(s.th, runpage.SessionLine{
			Mark: s.statusMark(st), Command: r.CommandLine(), Ago: ago, Selected: at == s.runAt,
		}, w))
	}
	return out
}

// runOutcome is the state a whole run is in, for its row in the list.
func runOutcome(r *runview.Run) runview.Status {
	switch {
	case !r.Finished:
		return runview.Running
	case r.Cancelled:
		return runview.Cancelled
	case r.OK():
		return runview.Passed
	}
	return runview.Failed
}

// statusMark is the plain symbol of a state, for text that is cut to a width.
func (s Shell) statusMark(st runview.Status) string {
	switch st {
	case runview.Running:
		return s.th.Icons.Running
	case runview.Passed:
		return s.th.Icons.Passed
	case runview.Cancelled:
		return s.th.Icons.Cancelled
	}
	return s.th.Icons.Failed
}

// logColumn is the log's column in w cells: its title (the task, or all output,
// and whether it follows), the command and the facts of the task, then the log
// with the one cell of padding of every pane.
func (s Shell) logColumn(g runsGeom, w int) []string {
	title := words.Say("runpage.tree.all_output")
	if s.runSel != "" {
		title = s.runSel
	}
	follow := words.Say("runpage.log.following")
	if !s.log.Following() {
		follow = words.Say("runpage.log.paused")
	}
	head := " " + s.th.Title(title) + "  " + s.th.Muted(follow)
	rows := logview.Rows(s.th, s.logLines(), w-logPad)
	if search := s.searchLine(rows); search != "" {
		head += "  " + search
	}
	out := []string{pane.Fit(head, w)}
	out = append(out, s.taskInfo(w)...)
	for _, line := range s.log.Render(s.th, s.litRows(rows), w-logPad, g.logRows) {
		out = append(out, strings.Repeat(" ", logPad)+line)
	}
	return out
}

// taskDetail is what the log's column says of a task after its command: the
// state, the time and the reason, and why a slow lint takes time.
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

// taskInfo tells the task whose log is shown, under the title of the log: the
// command it ran, then its state, time and reason, wrapped to w cells. It is
// empty while all output shows.
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
	for _, part := range pane.Wrap(s.th.Muted(s.taskDetail(t, s.now())), max(w-2, 4), 0) {
		out = append(out, pane.Fit(" "+part, w))
	}
	if t.Miss != "" && t.Status != runview.Cached {
		// Why the cache did not answer, as smith why-miss would say it.
		for _, part := range pane.Wrap(s.th.Muted(words.Say("runpage.log.miss", "why", t.Miss)), max(w-2, 4), 0) {
			out = append(out, pane.Fit(" "+part, w))
		}
	}
	return out
}

// logPad is the blank cells before each line of the log.
const logPad = 1

// taskFacts are the words the log's column tells of a task: the state, the time
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

// slowLintAfter is how long a lint task runs before the log's column says why
// it takes time.
const slowLintAfter = 5 * time.Second
