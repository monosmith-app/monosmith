package tui

import (
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/tui/adoptwizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dashboard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/fixwizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/logview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/palette"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pointer"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
)

var tabTitles = map[string]string{
	commands.TabHome: "Home", commands.TabProjects: "Projects", commands.TabRun: "Run", commands.TabDoctor: "Doctor",
	commands.TabDeps: "Deps", commands.TabRelease: "Release", commands.TabGuide: "Guide",
}

// Config is what the shell is built from.
type Config struct {
	Version   string
	Workspace string
	Theme     theme.Theme
	// Actions are the guided menu's entries.
	Actions []guide.Action
	// Commands are every smith command, for the palette.
	Commands []clicontract.CommandDescription
	// Sources load the views; nil outside a workspace.
	Sources *sources.Sources
	// EmptyFolder is true outside a workspace when the folder holds nothing yet.
	EmptyFolder bool
	// Adopt reads the folder for the adopt wizard; nil outside a folder to adopt.
	Adopt adoptwizard.Opener
	// Clipboard receives the OSC 52 sequence that copies text; nil disables copying.
	Clipboard io.Writer
}

// Shell is the interface of a bare `smith` in a terminal: a header with the
// workspace, the branch, the health and the view tabs, one content pane (the
// list, with its detail beside it from 100 columns), a row of key hints and a
// status bar that fill the terminal, usable with the keyboard (single keys,
// see package commands) and the mouse. Whatever it does is a smith
// command line, shown in the status bar and logged, run as a child process in
// the Run tab, and confirmed first when it changes the workspace or publishes.
type Shell struct {
	th      theme.Theme
	version string
	header  string
	branch  string
	w, h    int

	inWorkspace bool
	emptyFolder bool
	sources     sources.Sources
	tabs        []string
	tab         string
	// detailOpen is true while the detail of the selected project fills the
	// pane, in the narrow layout (Enter opens it, Esc closes it).
	detailOpen bool

	list projectlist.List
	// loadedProjects is true once the project list arrived (or failed).
	loadedProjects bool
	projectsErr    error

	guide   guide.Model
	pal     palette.Palette
	palOpen bool
	doctor  doctorview.Model
	release releaseview.Model
	deps    depsview.Model

	// last is how the last command ended, for Home.
	last *dashboard.Run
	// run is the command in the Run tab: running or the last one.
	run  *runview.Run
	proc runner.Handle
	// stopping is set once Ctrl+C asked the run to stop.
	stopping bool
	// runSel is the task whose log is shown ("" for all output); runTop scrolls the task list.
	runSel  string
	runTop  int
	log     logview.Log
	lastArg []string
	history []string
	// runs are the task runs that used the terminal, oldest first.
	runs []inlineRun

	confirm *dialog.Confirm
	help    *dialog.Help
	// fix is the doctor fix wizard while it is open; fixing is the fix being
	// applied in the background.
	fix    *fixwizard.Wizard
	fixing *fixRun
	// adopt is the adopt wizard while it is open; adopting is the job it
	// has running; adopter reads the folder; exitNote is told on leaving.
	adopt    *adoptwizard.Model
	adopting *fixRun
	adopter  adoptwizard.Opener
	exitNote string
	toast    toast.Toast
	// hint replaces the command in the status bar for hintTicks ticks.
	hint      string
	hintTicks int

	tick    int
	ticking bool
	mouseX  int
	mouseY  int
	// detailCursor is the highlighted action while the detail fills the pane.
	detailCursor int
	// zoom is set while the detail fills the pane on a wide screen (z).
	zoom bool
	// detailW is the width the user gave the detail pane ([ and ]); 0 is automatic.
	detailW int
	// detailTop is the first visible line of the detail; detailAll shows every
	// line of a folded list; detailFor is what the detail told last, so the
	// place starts over for another selection.
	detailTop int
	detailAll bool
	detailFor string
	lastClick pointer.Click
	now       func() time.Time

	// mouseOff is set while `m` has turned mouse reporting off.
	mouseOff bool
	// clip is where the OSC 52 sequence of `y` is written; nil copies nothing.
	clip io.Writer

	// Argv is a command that needs a terminal of its own (it asks questions or
	// takes over the screen): the shell leaves and smith runs it.
	Argv []string
	quit bool
}

// NewShell builds the shell. Without sources (outside a workspace) it has
// only the Guide view and the palette.
func NewShell(cfg Config) Shell {
	sh := Shell{
		th: cfg.Theme, version: cfg.Version, header: cfg.Workspace, now: time.Now,
		guide: guide.NewModel(cfg.Actions), mouseX: -1, mouseY: -1,
		list: projectlist.New(nil), clip: cfg.Clipboard, adopter: cfg.Adopt,
	}
	if sh.th.Icons.Spinner == nil {
		sh.th = theme.Default()
	}
	sh.pal = palette.New(nil, cfg.Commands)
	if cfg.Sources != nil {
		sh.inWorkspace = true
		sh.sources = *cfg.Sources
		sh.tab = commands.TabHome
		// Doctor runs in the background from the start (see Init).
		sh.doctor.Loading = sh.sources.Doctor != nil
	} else {
		sh.tab = commands.TabGuide
		sh.emptyFolder = cfg.EmptyFolder
	}
	sh.tabs = sh.tabList()
	sh.refreshPalette()
	return sh
}

// state is what decides which commands are available.
func (s Shell) state() commands.State {
	return commands.State{
		View: s.tab, Editing: s.editing(), Form: s.formOpen(), Narrow: s.lay().Mode == layout.Narrow, DetailOpen: s.detailFull(),
		InWorkspace: s.inWorkspace, Running: s.running(), EmptyFolder: s.emptyFolder, Reasons: s.viewReasons(),
	}
}

// viewReasons are the keys of the open view that cannot do anything now, and
// why: they are refused with the reason and leave the footer.
func (s Shell) viewReasons() map[string]string {
	r := map[string]string{}
	if s.doctor.Report == nil {
		r["ws.fix"] = "doctor has not checked the workspace yet"
	} else if _, ok := s.doctor.FixAll(); !ok {
		r["ws.fix"] = "doctor found nothing it can fix"
	}
	switch s.tab {
	case commands.TabDoctor:
		if _, ok := s.doctor.SelectedFix(); !ok {
			r["doctor.fix"] = "this check has no fix; its remedy says what to do"
		}
	case commands.TabDeps:
		if s.deps.UpArgv() == nil {
			r["deps.update"] = "no dependency to update"
		}
		if s.deps.AlignArgv() == nil {
			r["deps.align"] = "the one-version policy found nothing to align"
		}
	case commands.TabRelease:
		if !s.release.CanRelease() {
			r["rel.notes"] = "there is no release to show the notes of"
			r["rel.release"] = "there is nothing to release"
		}
	case commands.TabHome:
		if s.nextAction().CommandID == "" {
			r["home.next"] = "nothing to do yet"
		}
	case commands.TabRun:
		if s.run == nil {
			// The output of a run stays in the terminal's scrollback: there is no log to follow, no tasks to pick.
			r["run.follow"] = "the output is in your terminal's scrollback"
			r["run.scroll"] = "the output is in your terminal's scrollback"
			if s.lastArg == nil {
				r["run.again"] = "nothing has run yet"
			}
		}
	}
	return r
}

func (s Shell) running() bool { return s.run != nil && !s.run.Finished }

// tabList is the tabs of the strip: the views of a workspace; outside one, only the guide.
func (s Shell) tabList() []string {
	if s.inWorkspace {
		return []string{commands.TabHome, commands.TabProjects, commands.TabRun, commands.TabDoctor, commands.TabDeps, commands.TabRelease}
	}
	return []string{commands.TabGuide}
}

// refreshPalette rebuilds the palette's shell commands from the current
// state: what cannot work here is left out, and what is only unavailable for
// now carries its reason.
func (s *Shell) refreshPalette() {
	st := s.state()
	var entries []palette.Entry
	for _, c := range commands.All() {
		if c.Visible(st) && (st.InWorkspace || !c.NeedsWorkspace) {
			entries = append(entries, palette.Entry{ID: c.ID, Title: c.Title, Hint: c.Hint, Shortcut: c.Shortcut, Unavailable: c.Unavailable(st)})
		}
	}
	s.pal = s.pal.WithEntries(entries)
}

// Init implements tea.Model: it loads the project list in the background.
//
// It also runs doctor in the background, so the health in the header and on
// Home is known without opening the Doctor view.
func (s Shell) Init() tea.Cmd {
	if !s.inWorkspace {
		return nil
	}
	var cmds []tea.Cmd
	if load := s.sources.Projects; load != nil {
		cmds = append(cmds, func() tea.Msg { return load() })
	}
	if load := s.sources.Doctor; load != nil {
		cmds = append(cmds, func() tea.Msg { return load() })
	}
	return tea.Batch(cmds...)
}

// Messages the shell sends itself.
type (
	tickMsg struct{}
	// startedMsg reports that a command's process started (or why it could not).
	startedMsg struct {
		proc runner.Handle
		argv []string
		err  error
	}
	// outputMsg is a batch of output lines from a process, and its exit at the end.
	outputMsg struct {
		proc  runner.Handle
		lines []string
		exit  *runner.Exit
	}
	// plannedMsg lists the tasks a run will start.
	plannedMsg struct {
		proc runner.Handle
		ids  []string
	}
)

// Update implements tea.Model.
func (s Shell) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := s.handle(msg)
	s.forgetDetailPlace()
	s.refreshPalette()
	if s.quit {
		return s, tea.Quit
	}
	if s.animating() && !s.ticking {
		s.ticking = true
		cmd = tea.Batch(cmd, tickCmd())
	}
	return s, cmd
}

// tickInterval is the time between animation frames; tests shorten it.
var tickInterval = 100 * time.Millisecond

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

// animating is true while something moves: a run, a toast, a hint.
func (s Shell) animating() bool {
	return s.running() || s.fixApplying() || s.adopting != nil || s.toast.Shown() || s.hintTicks > 0
}

func (s *Shell) handle(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.resize(msg.Width, msg.Height)
	case tea.KeyMsg:
		return s.key(msg)
	case tea.MouseMsg:
		return s.mouse(msg)
	case tickMsg:
		s.tick++
		s.toast = s.toast.Tick()
		if s.hintTicks > 0 {
			if s.hintTicks--; s.hintTicks == 0 {
				s.hint = ""
			}
		}
		if s.animating() {
			return tickCmd()
		}
		s.ticking = false
	case sources.ProjectsLoaded:
		s.loadedProjects, s.projectsErr = true, msg.Err
		if msg.Err == nil {
			selected := s.list.SelectedName()
			s.list = projectlist.New(msg.Projects).SelectProject(selected)
			s.branch = msg.Branch
		}
	case doctorview.Result:
		s.doctor = s.doctor.Replace(msg)
	case fixProgressMsg:
		if s.adopting != nil && msg.run == s.adopting {
			return s.adoptProgress(msg)
		}
		return s.fixProgress(msg)
	case fixDoneMsg:
		if s.adopting != nil && msg.run == s.adopting {
			return s.adoptDone(msg)
		}
		return s.fixDone(msg)
	case depsview.Result:
		s.deps = depsview.FromResult(msg)
	case releaseview.Result:
		s.release = releaseview.FromResult(msg)
	case startedMsg:
		return s.started(msg)
	case plannedMsg:
		if s.run != nil && !s.run.Finished && s.proc == msg.proc {
			s.run.Plan(msg.ids)
		}
	case outputMsg:
		return s.output(msg)
	case inlineDoneMsg:
		return s.inlineDone(msg)
	}
	return nil
}

// resize re-lays the screen out; it runs on every WindowSizeMsg.
func (s *Shell) resize(w, h int) {
	s.w, s.h = w, h
	if s.fix != nil {
		placed := s.fix.Place(w, h)
		s.fix = &placed
	}
	if s.adopt != nil {
		placed := s.adopt.Place(w, h)
		s.adopt = &placed
	}
	s.tabs = s.tabList()
	if !s.canShow(s.tab) {
		s.tab = commands.TabHome
		if !s.inWorkspace {
			s.tab = commands.TabGuide
		}
	}
	if s.lay().Mode == layout.Wide {
		// The detail is beside the list now; there is nothing to close.
		s.detailOpen = false
	}
}

// lay is where each area of the screen is drawn, for the size and the open view.
func (s Shell) lay() layout.Layout {
	return layout.DivideWith(s.w, s.h, s.hasDetail() && !s.zoomed(), s.detailW)
}

// canShow is true for a view the shell can open: a tab of the strip, or the
// guided form inside a workspace (a form is not a tab, a command opens it).
func (s Shell) canShow(id string) bool {
	return contains(s.tabs, id) || (id == commands.TabGuide && s.inWorkspace)
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// say shows a message in the status bar for a few seconds.
func (s *Shell) say(text string) {
	s.hint, s.hintTicks = text, 40
}

// notify shows a toast.
func (s *Shell) notify(text string, level toast.Level) {
	s.toast = toast.New(text, level)
}
