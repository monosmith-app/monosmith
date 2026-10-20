package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/logview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/menubar"
	"github.com/monosmith-app/monosmith/apps/smith/tui/navigator"
	"github.com/monosmith-app/monosmith/apps/smith/tui/palette"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pointer"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
)

// tabNavigator is the navigator shown as a tab below 80 columns.
const tabNavigator = "workspace"

var tabTitles = map[string]string{
	tabNavigator: "Workspace", commands.TabProjects: "Projects", commands.TabRun: "Run", commands.TabDoctor: "Doctor",
	commands.TabRelease: "Release", commands.TabPlugins: "Plugins", commands.TabDeps: "Deps", commands.TabSettings: "Settings", commands.TabGuide: "Guide",
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
	Sources *Sources
}

// Shell is the interface of a bare `smith` in a terminal: a menu bar, a
// navigator, a main view with tabs, an inspector and a status bar that fill the
// terminal, usable with the keyboard and the mouse. Whatever it does is a smith
// command line, shown in the status bar and logged, run as a child process in
// the Run tab, and confirmed first when it changes the workspace or publishes.
type Shell struct {
	th      theme.Theme
	version string
	header  string
	branch  string
	w, h    int
	lay     layout.Layout

	inWorkspace bool
	sources     Sources
	tabs        []string
	tab         string
	focus       layout.Area

	bar   menubar.Bar
	popup menubar.Popup
	tree  navigator.Tree
	// loadedProjects is true once the project list arrived (or failed).
	loadedProjects bool
	projectsErr    error

	guide    guide.Model
	pal      palette.Palette
	palOpen  bool
	doctor   doctorView
	release  releaseView
	settings settingsView
	plugins  pluginsView
	deps     depsView

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

	confirm *dialog.Confirm
	help    *dialog.Help
	toast   toast.Toast
	// hint replaces the command in the status bar for hintTicks ticks.
	hint      string
	hintTicks int

	tick    int
	ticking bool
	mouseX  int
	mouseY  int
	// inspCursor is the highlighted action in the inspector.
	inspCursor int
	lastClick  pointer.Click
	now        func() time.Time

	// Argv is a command that needs a terminal of its own (it asks questions or
	// takes over the screen): the shell leaves and smith runs it.
	Argv []string
	quit bool
}

// NewShell builds the shell. Without sources (outside a workspace) it has
// only the Guide tab, the palette and the menus.
func NewShell(cfg Config) Shell {
	sh := Shell{
		th: cfg.Theme, version: cfg.Version, header: cfg.Workspace, now: time.Now,
		guide: guide.NewModel(cfg.Actions), focus: layout.Main, mouseX: -1, mouseY: -1,
		tree: navigator.New(nil),
	}
	if sh.th.Icons.Spinner == nil {
		sh.th = theme.Default()
	}
	sh.pal = palette.New(nil, cfg.Commands)
	if cfg.Sources != nil {
		sh.inWorkspace = true
		sh.sources = *cfg.Sources
		sh.tab = commands.TabProjects
	} else {
		sh.tab = commands.TabGuide
	}
	sh.bar = menubar.New(commands.Menus(sh.state()), 80)
	sh.tabs = sh.tabList()
	sh.refreshMenus()
	return sh
}

// state is what decides which commands are available.
func (s Shell) state() commands.State {
	sel := s.tree.Selected()
	return commands.State{InWorkspace: s.inWorkspace, Running: s.running(), HasSelection: sel.Valid && sel.Kind == navigator.TaskRow}
}

func (s Shell) running() bool { return s.run != nil && !s.run.Finished }

// tabList is the tabs in this layout and workspace.
func (s Shell) tabList() []string {
	var tabs []string
	if s.lay.Mode == layout.Narrow && s.inWorkspace {
		tabs = append(tabs, tabNavigator)
	}
	if s.inWorkspace {
		tabs = append(tabs, commands.TabProjects, commands.TabRun, commands.TabDoctor, commands.TabRelease, commands.TabPlugins, commands.TabDeps, commands.TabSettings)
	}
	return append(tabs, commands.TabGuide)
}

// refreshMenus rebuilds the menu bar and the palette's shell commands from
// the current state, so every unavailable item carries its reason.
func (s *Shell) refreshMenus() {
	st := s.state()
	s.bar = s.bar.SetMenus(commands.Menus(st)).SetWidth(s.w)
	var entries []palette.Entry
	for _, c := range commands.All() {
		if c.Visible(st) {
			entries = append(entries, palette.Entry{ID: c.ID, Title: c.Title, Hint: c.Hint, Shortcut: c.Shortcut, Unavailable: c.Unavailable(st)})
		}
	}
	s.pal = s.pal.WithEntries(entries)
}

// Init implements tea.Model: it loads the project list in the background.
func (s Shell) Init() tea.Cmd {
	if s.inWorkspace && s.sources.Projects != nil {
		load := s.sources.Projects
		return func() tea.Msg { return load() }
	}
	return nil
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
	doctorFailed struct{ err error }
)

// Update implements tea.Model.
func (s Shell) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := s.handle(msg)
	s.refreshMenus()
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
	return s.running() || s.toast.Shown() || s.hintTicks > 0
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
	case projectsLoaded:
		s.loadedProjects, s.projectsErr = true, msg.err
		if msg.err == nil {
			selected := s.tree.Selected()
			s.tree = navigator.New(msg.projects).SetHeight(max(s.lay.Navigator.H-2, 1))
			switch {
			case selected.Project != "":
				s.tree = s.tree.SelectProject(selected.Project)
			case len(s.tree.Ordered()) > 0:
				s.tree = s.tree.SelectProject(s.tree.Ordered()[0].Name)
			}
			s.branch = msg.branch
		}
	case doctorLoaded:
		s.doctor = doctorView{report: &msg.report}
	case doctorFailed:
		s.doctor = doctorView{err: msg.err}
	case pluginsLoaded:
		s.plugins = pluginsView{loaded: true, rows: msg.rows, err: msg.err, cursor: s.plugins.cursor}
		if s.plugins.cursor >= len(msg.rows) {
			s.plugins.cursor = 0
		}
	case depsLoaded:
		s.deps = depsView{loaded: true, rows: msg.rows, policy: msg.policy, err: msg.err}
	case settingsLoaded:
		s.settings = settingsView{loaded: true, settings: msg.settings, err: msg.err}
	case releaseLoaded:
		s.release = releaseView{loaded: true, plan: msg.plan, blockers: msg.blockers, warnings: msg.warnings, err: msg.err}
	case startedMsg:
		return s.started(msg)
	case plannedMsg:
		if s.run != nil && !s.run.Finished && s.proc == msg.proc {
			s.run.Plan(msg.ids)
		}
	case outputMsg:
		return s.output(msg)
	}
	return nil
}

// resize re-lays the screen out; it runs on every WindowSizeMsg.
func (s *Shell) resize(w, h int) {
	s.w, s.h = w, h
	s.lay = layout.Divide(w, h)
	s.tabs = s.tabList()
	if !contains(s.tabs, s.tab) {
		s.tab = commands.TabProjects
		if !s.inWorkspace {
			s.tab = commands.TabGuide
		}
	}
	if (!s.lay.HasNavigator() && s.focus == layout.Navigator) || (!s.lay.HasInspector() && s.focus == layout.Inspector) {
		s.focus = layout.Main
	}
	s.tree = s.tree.SetHeight(max(s.lay.Navigator.H-2, 1))
	s.bar = s.bar.SetWidth(w)
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
