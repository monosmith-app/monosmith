package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/packageprompt"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/compose"
	"github.com/monosmith-app/monosmith/apps/smith/recipewizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/adoptwizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/cachewizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/composer"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dashboard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/fixwizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/helpdialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/logview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/palette"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pointer"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releasepick"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/resultdialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/selection"
	"github.com/monosmith-app/monosmith/apps/smith/tui/settingsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
	"github.com/monosmith-app/monosmith/apps/smith/tui/splitter"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toolsdialog"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// tabTitles are the names of the tabs, by tab id.
func tabTitles() map[string]string {
	return map[string]string{
		commands.TabHome: words.Say("tui.tab.home"), commands.TabProjects: words.Say("tui.tab.projects"), commands.TabDoctor: words.Say("tui.tab.doctor"),
		commands.TabDeps: words.Say("tui.tab.deps"), commands.TabRelease: words.Say("tui.tab.release"), commands.TabSettings: words.Say("tui.tab.settings"), commands.TabGuide: words.Say("tui.tab.guide"),
	}
}

// Config is what the shell is built from.
type Config struct {
	Version   string
	Workspace string
	Theme     theme.Theme
	// Splits are the sizes of the panels the person set earlier, by divider; SaveSplits
	// keeps them when they change (nil keeps none).
	Splits     map[string]int
	SaveSplits func(map[string]int)
	// Actions are the guided menu's entries.
	Actions []guide.Action
	// Commands are every smith command, for the guided forms.
	Commands []clicontract.CommandDescription
	// Sources load the views; nil outside a workspace.
	Sources *sources.Sources
	// Outside lets a shell that is not in a workspace run smith new and smith
	// convert in place (see sources.OutsideSources); nil runs them in the terminal.
	Outside *sources.Sources
	// EmptyFolder is true outside a workspace when the folder holds nothing yet.
	EmptyFolder bool
	// Adopt reads the folder for the adopt wizard; nil outside a folder to adopt.
	Adopt adoptwizard.Opener
	// Repair is the offer to fix a smith.toml that uses a renamed key; nil when
	// there is none to offer.
	Repair *sources.ConfigRepair
	// Start is a command to run as soon as the shell opens in a workspace.
	Start []string
	// Resume opens a guided form again with the answers of a try that failed.
	Resume *Resume
	// Clipboard receives the OSC 52 sequence that copies text; nil disables copying.
	Clipboard io.Writer
	// Notice is a quiet line shown once as a toast: smith was updated, or a
	// newer one exists.
	Notice string
	// Problem is why smith.toml could not be read (with its line and column);
	// "" when there is no such file or it reads well.
	Problem string
	// CommitOperations is true when smith.toml sets commit_operations, so every
	// change smith makes is already a commit.
	CommitOperations bool
	// CommitOperationsSet is true when smith.toml says anything about
	// commit_operations; CommitChoice is what the person answered once ("no" or
	// "ask"), and SaveCommitChoice keeps such an answer (nil keeps none).
	CommitOperationsSet bool
	CommitChoice        string
	SaveCommitChoice    func(choice string)
	// Environments are the environment names `[workspace] environments`
	// declares; the Add form starts its "names of my own" question on them.
	Environments []string
	// Hyperlinks is true when the terminal draws OSC 8 hyperlinks (see package hyperlink).
	Hyperlinks bool
}

// Shell is the interface of a bare `smith` in a terminal: a header with the
// workspace, the branch, the health and the view tabs, one content pane (the
// list, with its detail beside it from 100 columns), a row of key hints and a
// status bar that fill the terminal, usable with the keyboard (single keys,
// see package commands) and the mouse. Whatever it does is a smith
// command line, shown in the status bar and logged, run as a child process in
// the run panel of Projects, and confirmed first when it changes the workspace or publishes.
type Shell struct {
	th      theme.Theme
	version string
	header  string
	branch  string
	w, h    int

	inWorkspace bool
	emptyFolder bool
	// problem is why smith.toml cannot be read; commitOps mirrors commit_operations.
	problem string
	// repair is the fix offered for an unreadable smith.toml; the wizard that
	// shows it is s.fix.
	repair    *sources.ConfigRepair
	commitOps bool
	// commitSet, commitChoice and saveCommit decide whether smith still has to
	// ask how it treats undo (see commitUnasked); commitPending is the command
	// that waits for the answer, and commitPendingOK says it was confirmed.
	commitSet       bool
	commitChoice    string
	saveCommit      func(string)
	commitPending   []string
	commitPendingOK bool
	// environments are the names of `[workspace] environments`.
	environments []string
	sources      sources.Sources
	tabs         []string
	tab          string
	// detailOpen is true while the detail of the selected project fills the
	// pane, in the narrow layout (Enter opens it, Esc closes it).
	detailOpen bool

	list projectlist.List
	// projScroll is where the project list's window starts, apart from its
	// cursor, so the wheel scrolls without selecting.
	projScroll *pane.Scroll
	// doctorScroll is the same for the Doctor list.
	doctorScroll *pane.Scroll
	// loadedProjects is true once the project list arrived (or failed).
	loadedProjects bool
	projectsErr    error

	guide guide.Model
	// addForm is the add form as it ended, kept while its plan is on screen: Esc
	// on the plan goes back to its last question, with the answers filled in.
	addForm *guide.Model
	// formOrigin is the tab a guided form was opened from: cancelling the form,
	// or the confirmation it led to, shows that tab again.
	formOrigin string
	// settingsSince is when the Settings tab started the change it waits for.
	settingsSince time.Time
	pal           palette.Palette
	palOpen       bool
	doctor        doctorview.Model
	release       releaseview.Model
	deps          depsview.Model
	// depsLoading is true while the registries are asked for newer releases.
	depsLoading bool
	// settings is the Settings tab: plugins, style and the settings of smith.toml.
	settings settingsview.Model
	// pluginUndo is the plugin switch the status line offers to undo with u, until
	// another key is pressed or the tab is left.
	pluginUndo *settingsview.PluginToggle

	// completionWanted is true while Home lists "Tab completion is not set up".
	// It is read once at the start, and cleared by "Not now" and by a setup that
	// ends well.
	completionWanted bool
	// last is how the last command ended, for Home.
	last *dashboard.Run
	// justAdded is the project a finished `smith add` made, to be selected
	// once the project list has been read again.
	justAdded string
	// verifying are the checks a finished fix says it clears: the next doctor
	// result is read to see whether they pass now.
	verifying []string
	// rechecking is true while a doctor run the person asked for (Check again)
	// is out, so its result says in a toast what doctor found.
	rechecking bool
	// verifyingWhat is what that fix did, in words, for the "Fixed" toast.
	verifyingWhat string
	// verifiedFix and verifiedOutput are the wizard and the output of that fix,
	// kept so a fix whose check still fails can show what its steps printed.
	verifiedFix    fixwizard.Wizard
	verifiedOutput []string
	// run is the command the run panel shows: running or the last one.
	run  *runview.Run
	proc runner.Handle
	// stopping is set once Ctrl+C asked the run to stop.
	stopping bool
	// runSel is the task whose log is shown ("" for all output); runTop scrolls the task list.
	runSel string
	// runOpen is true while the right panel of Projects shows the run (a run
	// that starts opens it, a recent run is opened into it; Esc or Back closes it).
	runOpen bool
	runTop  int
	log     logview.Log
	lastArg []string
	history []string
	// runs are the runs of the session, oldest first; runAt is the one shown (run).
	runs  []*runview.Run
	runAt int
	// recentAt is the row chosen in the list of recent runs under the projects
	// (0 is the newest run); recentFocus is true while the keyboard moves in
	// that list and not in the list of projects; the list is unfolded upward to
	// show every run kept while it is true.
	recentAt    int
	recentFocus bool
	// sideFocus is true while the keyboard is on the right panel (the detail,
	// or the run) and not on the list beside it.
	sideFocus bool
	// homeSel is the button or run of Home the keyboard is on, counted in
	// reading order; homeTop is the first line of Home shown while it scrolls.
	homeSel, homeTop int

	confirm *dialog.Confirm
	// releaseWaiting is a release command line that waits for the plan to be
	// read, so its confirmation can list the packages and versions.
	releaseWaiting []string
	// releasing is the release the shell started: its packages and its remote,
	// for the steps drawn while it runs and for its result.
	releasing *releaseRun
	// releaseLoading is true while the release plan is read, so no second read
	// (the plan takes seconds of git) starts: the start reads it in the
	// background for Home, and the Release tab uses that read.
	releaseLoading bool
	// planWait is the plan a dialog that opened at once still waits for (see
	// openWaiting); nil when no dialog waits.
	planWait *planWait
	// planTicket counts the waits, so a late plan of a cancelled one is ignored.
	planTicket int
	// choices says whether the choices of the plugins are read; offers are
	// those that can be used, formPending names the form (add or swap) that waits
	// for them, and addBase and swapBase are the plain forms they are added to.
	choices     choicesState
	offers      []guide.ChoiceOffer
	formPending string
	addBase     *guide.Action
	swapBase    *guide.Action
	// recipeLevels are the plugins each level of a recipe can take, read once;
	// recipeBase is the plain Write a recipe form they are added to.
	recipeLevels *recipewizard.LevelCatalog
	recipeBase   *guide.Action
	// removeBase is the plain Remove form, which the project list turns into a pick list.
	removeBase *guide.Action
	// pkg asks which packages to install on, or remove from, the selected project.
	pkg *packageprompt.Prompt
	// relpick asks which packages of the release to hold back or bump, before the confirmation.
	relpick *releasepick.Model
	help    *helpdialog.Help
	// fix is the doctor fix wizard while it is open; fixing is the fix being
	// applied in the background.
	fix    *fixwizard.Wizard
	fixing *fixRun
	// depsRepair is set while the fix wizard walks a dependency repair.
	depsRepair *depsRepairWalk
	// adopt is the adopt wizard while it is open; adopting is the job it
	// has running; adopter reads the folder; exitNote is told on leaving.
	adopt    *adoptwizard.Model
	adopting *fixRun
	adopter  adoptwizard.Opener
	// compose is the "Compose a stack" screen while it is open; composeCat is
	// the catalogue it was read from, kept for the next time.
	compose    *composer.Model
	composeCat *compose.Catalogue
	// tools is the "Install the missing tools" dialog while it is open;
	// toolsProc is the install it streams and toolsArgv the command it came from.
	tools *toolsdialog.Model
	// cache is the remote cache setup while it is open; cacheRun is the engine
	// behind it, which runs in the background.
	cache    *cachewizard.Model
	cacheRun *cachewizard.Bridge
	// remote is the state of the remote cache, for the header and Home.
	remote remoteCache
	// cacheTierChosen is set while a cache clear chosen in the "which cache"
	// dialog goes on to its confirmation.
	cacheTierChosen bool
	// cacheSecrets is the clipboard and environment of its key screen; nil is this machine.
	cacheSecrets func() cachewizard.Secrets
	// result is the dialog that closes a run that changed files or could not be
	// checked; resultKept is the failure one, shown again after the tools install;
	// runLines is the output of the run that is going or ended last.
	result     *resultdialog.Model
	resultKept *resultdialog.Model
	runLines   []string
	toolsProc  runner.Handle
	toolsArgv  []string
	exitNote   string
	// reopen is true when the shell leaves only to start again inside the workspace it made.
	reopen bool
	// reopenStart is the command the reopened shell starts with; start is the one this shell starts with.
	reopenStart, startWith []string
	// reopenNote is the toast the reopened shell starts with.
	reopenNote string
	// reopenDir is the folder the reopened shell opens in; empty is this one.
	reopenDir string
	toast     toast.Toast
	// hint replaces the command in the status bar for hintTicks ticks.
	hint      string
	hintTicks int

	// waitFor and waitSince say which load the open tab waits for and since
	// when (see trackWait).
	waitFor   string
	waitSince time.Time
	tick      int
	ticking   bool
	mouseX    int
	// mouseShift is whether shift was held at the last mouse event.
	mouseShift bool
	mouseY     int
	// pressed is the action whose button the mouse holds down, "" for none; the
	// action runs when the button is let go over the same one. probing is set on
	// the copy of the Shell that looks up what is under the pointer, so drawing
	// that lookup does not start another.
	pressed string
	probing bool
	// hoverAt is a pointer motion with no button held that waits for the next
	// hover frame, and hoverWaits is set while that frame is on its way (see
	// deferHover). reuseFrame is set when the last event changed nothing on
	// screen, so View gives back frame, the screen it drew last.
	hoverAt    *tea.MouseEvent
	hoverWaits bool
	reuseFrame bool
	frame      *lastFrame
	// detailCursor is the highlighted action while the keyboard is in the detail,
	// and the chosen action of a split detail (see detail.Panel.Splits).
	// detailOnRun puts the keyboard on the run button of a split detail instead.
	detailCursor int
	detailOnRun  bool
	// zoom is set while the detail fills the pane on a wide screen (z).
	zoom bool
	// split is what the person did to the dividers between panels (see the splitter
	// package); saveSplits keeps the sizes for the next start, nil to keep none.
	split      splitter.State
	saveSplits func(map[string]int)
	// selection is the text the person selects with a drag (see the selection
	// package); selected is the text it held at the release, so a screen that
	// changes under the highlight drops it.
	selection selection.State
	// modalPress is where the left button went down inside a dialog, and
	// modalPressed says a drag from there may select text.
	modalPress   [2]int
	modalPressed bool
	selected     string
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
	// hyperlinks is true when the terminal draws OSC 8 hyperlinks.
	hyperlinks bool

	// Argv is a command that needs a terminal of its own (it asks questions or
	// takes over the screen): the shell leaves and smith runs it.
	Argv []string
	// Form and Answers say which guided form made Argv and what was answered,
	// so a command that fails after the shell left can open the form again
	// with the answers kept.
	Form    string
	Answers map[string][]string
	quit    bool
}

// NewShell builds the shell. Without sources (outside a workspace) it has
// only the Guide view and the palette.
func NewShell(cfg Config) Shell {
	sh := Shell{
		th: cfg.Theme, version: cfg.Version, header: cfg.Workspace, now: time.Now,
		guide: guide.NewModel(cfg.Actions), mouseX: -1, mouseY: -1, projScroll: &pane.Scroll{}, doctorScroll: &pane.Scroll{},
		list: projectlist.New(nil), clip: cfg.Clipboard, adopter: cfg.Adopt, startWith: cfg.Start,
		frame: &lastFrame{},
	}
	sh.problem, sh.commitOps, sh.environments = cfg.Problem, cfg.CommitOperations, cfg.Environments
	sh.hyperlinks = cfg.Hyperlinks
	sh.commitSet, sh.commitChoice, sh.saveCommit = cfg.CommitOperationsSet, cfg.CommitChoice, cfg.SaveCommitChoice
	if cfg.Repair != nil {
		sh.repair = cfg.Repair
		sh.sources.ShowFix, sh.sources.ApplyFix = cfg.Repair.Show, cfg.Repair.Apply
	}
	if sh.th.Icons.Spinner == nil {
		sh.th = theme.Default()
	}
	sh.pal = palette.New(nil)
	sh.split, sh.saveSplits = splitter.New(cfg.Splits), cfg.SaveSplits
	if cfg.Sources != nil {
		sh.inWorkspace = true
		sh.sources = *cfg.Sources
		sh.settings = settingsview.New(sh.sources.Settings, settingsForms(cfg.Actions))
		sh.applySplits()
		sh.tab = commands.TabHome
		// Doctor runs in the background from the start (see Init).
		sh.doctor.Loading = sh.sources.Doctor != nil
		// So does the release plan, for the release item of Home.
		sh.releaseLoading = sh.sources.Release != nil
		sh.completionWanted = sh.sources.CompletionWanted != nil && sh.sources.CompletionWanted()
		// And the remote cache's probe: the header says "checking" from the start.
		sh.remote.set = sh.sources.RemoteSet != nil && sh.sources.RemoteSet() && sh.sources.RemoteStatus != nil
		sh.remote.probing = sh.remote.set
	} else {
		sh.tab = commands.TabGuide
		sh.emptyFolder = cfg.EmptyFolder
		if cfg.Outside != nil {
			sh.sources = *cfg.Outside
		}
	}
	if r := cfg.Resume; r != nil && cfg.Sources == nil {
		sh.guide, _ = sh.guide.Resume(r.Form, r.Answers, r.Problem)
	}
	sh.tabs = sh.tabList()
	sh.refreshPalette()
	if cfg.Notice != "" {
		sh.notify(cfg.Notice, toast.Info)
	}
	return sh
}

// state is what decides which commands are available.
func (s Shell) state() commands.State {
	return commands.State{
		View: s.tab, Editing: s.editing(), Form: s.formOpen(), Narrow: s.lay().Mode == layout.Narrow, DetailOpen: s.detailFull(),
		InWorkspace: s.inWorkspace, Running: s.running(), RunOpen: s.runShown() && s.sideKeys(), RecentFocus: s.recentFocus && !s.sideKeys(), SideFocus: s.sideFocus && s.beside(), SideNext: s.sideNext(), RunFailed: s.runFailed(), TaskSelected: s.runSel != "", EmptyFolder: s.emptyFolder, BrokenConfig: s.problem != "", Reasons: s.viewReasons(),
		Project: s.selectedProject(), Projects: s.list.Len(), TaskProjects: s.taskProjects(), OtherTasks: s.otherTasks(), On: s.toggles(),
	}
}

// toggles are the toggle commands that are on now: the footer and the palette
// then name what the key does next ("ungroup", "collapse all").
func (s Shell) toggles() map[string]bool {
	return map[string]bool{
		"doctor.group":     s.doctor.Grouped,
		"doctor.foldall":   s.doctor.Grouped && s.doctor.AllOpen(),
		"run.recent-focus": s.recentFocus,
		"doctor.failed":    s.doctor.Shows(doctorview.FilterFailed),
		"doctor.warnings":  s.doctor.Shows(doctorview.FilterWarnings),
		"doctor.passed":    s.doctor.Shows(doctorview.FilterPassed),
	}
}

// otherTasks are the selected project's tasks that have no key of their own.
func (s Shell) otherTasks() []string {
	p, ok := s.list.Selected()
	if !ok {
		return nil
	}
	var names []string
	for _, t := range p.Tasks {
		switch t.Name {
		case "build", "test", "lint", "typecheck":
		default:
			names = append(names, t.Name)
		}
	}
	return names
}

// taskProjects counts, for each task, the projects that have it.
func (s Shell) taskProjects() map[string]int {
	counts := map[string]int{}
	for _, p := range s.list.Projects() {
		for _, t := range p.Tasks {
			counts[t.Name]++
		}
	}
	return counts
}

// viewReasons are the keys of the open view that cannot do anything now, and
// why: they are refused with the reason and leave the footer.
func (s Shell) viewReasons() map[string]string {
	r := map[string]string{}
	if s.doctor.Report == nil {
		r["ws.fix"] = words.Say("tui.shell.doctor_has_not_checked_the_workspace")
	} else if _, ok := s.doctor.FixAll(); !ok {
		r["ws.fix"] = words.Say("tui.shell.doctor_found_nothing_it_can_fix")
	}
	if s.list.Len() == 0 {
		r["projects.install-package"] = words.Say("tui.shell.the_workspace_has_no_project_yet")
		r["projects.remove-package"] = words.Say("tui.shell.the_workspace_has_no_project_yet")
	}
	switch s.tab {
	case commands.TabDoctor:
		if _, ok := s.doctor.SelectedFix(); !ok {
			r["doctor.fix"] = words.Say("tui.shell.this_check_has_no_fix_its")
		}
		if !s.doctor.Grouped {
			r["doctor.foldall"] = words.Say("tui.toggle.fold_needs_groups")
		}
	case commands.TabDeps:
		if !s.deps.Loaded {
			// Until the registries answer no key but reload has anything to act on.
			for _, id := range []string{"deps.mark", "deps.filter-all", "deps.filter-patch", "deps.filter-minor", "deps.filter-major", "deps.update", "deps.install", "deps.remove", "deps.align", "deps.one-version"} {
				r[id] = words.Say("tui.shell.deps_not_loaded")
			}
			break
		}
		if s.deps.UpArgv() == nil {
			r["deps.update"] = words.Say("tui.shell.no_dependency_to_update")
		}
		if _, _, ok := s.deps.RemoveTargets(); !ok {
			r["deps.remove"] = words.Say("tui.shell.no_dependency_is_selected")
		}
		if len(s.deps.Policy) > 0 {
			r["deps.one-version"] = words.Say("tui.shell.dependencies_one_version_is_already_on")
		}
		if s.deps.AlignArgv() == nil {
			r["deps.align"] = words.Say("tui.shell.the_one_version_policy_found_nothing")
		}
	case commands.TabRelease:
		if !s.release.CanRelease() {
			r["rel.notes"] = words.Say("tui.shell.there_is_no_release_to_show")
			r["rel.release"] = words.Say("tui.shell.there_is_nothing_to_release")
			r["release.open"] = words.Say("tui.shell.there_is_no_bump_to_show")
			r["release.expand"] = words.Say("tui.shell.there_is_no_bump_to_show")
		}
		if !s.release.Empty() {
			r["rel.why"] = words.Say("tui.release_reason.nothing_to_explain")
		}
		if !s.release.FirstRelease() {
			r["rel.first"] = words.Say("tui.release_reason.no_first")
		}
	case commands.TabSettings:
		for id, why := range s.settings.Reasons() {
			r[id] = why
		}
	case commands.TabHome:
		if s.homeCommand() == "" {
			r["home.next"] = words.Say("tui.shell.nothing_to_do_yet")
			r["home.move"] = words.Say("tui.shell.nothing_to_do_yet")
		}
		if !s.homeScrolls() {
			r["home.scroll"] = words.Say("tui.shell.the_page_fits")
		}
	}
	for id, why := range s.runReasons() {
		r[id] = why
	}
	return r
}

func (s Shell) running() bool { return s.run != nil && !s.run.Finished }

// runShown is true while the right panel of the open Projects view shows a run.
func (s Shell) runShown() bool {
	return s.runWanted() && (s.detailFull() || s.lay().HasDetail())
}

// runWanted is true while the open view is Projects and a run is open for it,
// whether or not the screen has room to draw it beside the list. It is the
// layout's own question, so it never asks the layout.
func (s Shell) runWanted() bool {
	return s.runOpen && s.run != nil && s.tab == commands.TabProjects && s.inWorkspace
}

// tabList is the tabs of the strip: the views of a workspace; outside one, only the guide.
func (s Shell) tabList() []string {
	if s.inWorkspace {
		return []string{commands.TabHome, commands.TabProjects, commands.TabDoctor, commands.TabDeps, commands.TabRelease, commands.TabSettings}
	}
	return []string{commands.TabGuide}
}

// refreshPalette rebuilds the palette's shell commands from the current
// state: what cannot work here is left out, and what is only unavailable for
// now carries its reason.
func (s *Shell) refreshPalette() {
	st := s.state()
	var entries []palette.Entry
	for _, e := range commands.PaletteEntries(st) {
		entries = append(entries, palette.Entry{ID: e.ID, Section: e.Section, Title: e.Title, Hint: e.Hint, Shortcut: e.Shortcut, Keywords: e.Keywords, Unavailable: e.Unavailable})
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
	if len(s.startWith) > 0 {
		argv := s.startWith
		cmds = append(cmds, func() tea.Msg { return startRunMsg{argv: argv} })
	}
	if load := s.sources.Projects; load != nil {
		cmds = append(cmds, func() tea.Msg { return load() })
	}
	if load := s.sources.Doctor; load != nil {
		cmds = append(cmds, func() tea.Msg { return load() })
	}
	if load := s.sources.Release; load != nil {
		// Home never waits for it: the release item shows when the plan arrives.
		cmds = append(cmds, func() tea.Msg { return load() })
	}
	if probe := s.sources.RemoteStatus; s.remote.set && probe != nil {
		// Init has a value receiver: the probe's state is set when it answers.
		cmds = append(cmds, func() tea.Msg { return remoteStatusMsg{status: probe()} })
	}
	return tea.Batch(cmds...)
}

// loadRelease reads the release plan in the background, unless a read is
// already going (the plan takes seconds of git) or this smith has no plan to read.
func (s *Shell) loadRelease() tea.Cmd {
	load := s.sources.Release
	if load == nil || s.releaseLoading {
		return nil
	}
	s.releaseLoading = true
	return func() tea.Msg { return load() }
}

// Messages the shell sends itself.
type (
	tickMsg struct{}
	// startRunMsg starts a command as the shell opens.
	startRunMsg struct{ argv []string }
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

// lay is where each area of the screen is drawn, for the size and the open view.
func (s Shell) lay() layout.Layout {
	return layout.DivideShared(s.w, s.h, s.hasDetail() && !s.zoomed(), s.detailWidth(), s.inlineHeader())
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

// everythingIsNotNews drops the "affected" mark when every project has it: with
// nothing to compare to (right after a conversion, or with no release tag yet)
// the mark says nothing, and 23 of them bury the rows that matter.
func everythingIsNotNews(projects []projectlist.Project) []projectlist.Project {
	if len(projects) < 2 {
		return projects
	}
	for _, p := range projects {
		if !p.Affected {
			return projects
		}
	}
	out := make([]projectlist.Project, len(projects))
	copy(out, projects)
	for i := range out {
		out[i].Affected = false
	}
	return out
}
