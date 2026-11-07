package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/settingsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
)

// Update implements tea.Model.
func (s Shell) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := s.handle(msg)
	s.keepSelection()
	s.forgetDetailPlace()
	s.refreshPalette()
	s.trackWait()
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
	return s.running() || s.fixApplying() || s.adopting != nil || s.waitFor != "" || s.toast.Shown() || s.hintTicks > 0 || s.planWait != nil || s.settings.Busy() || s.backgroundWork()
}

func (s *Shell) handle(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case tea.WindowSizeMsg, tea.KeyMsg, tea.FocusMsg, tea.BlurMsg:
		// A hold on a divider needs a mouse in motion: these end it.
		s.letGoSplit()
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.resize(msg.Width, msg.Height)
	case tea.KeyMsg:
		if s.belowMinimum() {
			// Nothing is drawn but the hint: a key must not act on a screen the
			// person cannot see (Enter could start a command). Ctrl+C still leaves.
			if msg.String() == "ctrl+c" {
				s.quit = true
			}
			return nil
		}
		return s.key(msg)
	case tea.MouseMsg:
		if s.belowMinimum() {
			s.letGoSplit()
			return nil
		}
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
			if s.justAdded != "" {
				// The project a finished add made is the one the next keys act on.
				selected, s.justAdded = s.justAdded, ""
			}
			s.list = projectlist.New(everythingIsNotNews(msg.Projects)).SelectProject(selected)
			s.branch = msg.Branch
			s.refreshProjectForms()
		}
	case doctorview.Result:
		s.doctor = s.doctor.Replace(msg)
		s.verifyFix(msg)
	case settingsview.Msg:
		return s.settingsMsg(msg)
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
		s.deps, s.depsLoading = depsview.FromResult(msg), false
	case releaseview.Result:
		s.release, s.releaseLoading = releaseview.FromResult(msg), false
		if argv := s.releaseWaiting; argv != nil {
			// The release was asked before the plan was read: ask now, with the plan.
			s.releaseWaiting = nil
			s.closeWaiting()
			return s.execute(argv, false)
		}
	case firstCommitMsg:
		return s.firstCommitted(msg)
	case toolsStartedMsg:
		return s.toolsStarted(msg)
	case toolsOutputMsg:
		return s.toolsOutput(msg)
	case capturedMsg:
		return s.captured(msg)
	case openedMsg:
		s.opened(msg)
	case startRunMsg:
		return s.execute(msg.argv, true)
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

// dialogMargins are the rows a confirmation keeps clear: the header and the tab
// bar above the framed area, the status bar below it.
func (s Shell) dialogMargins() (top, bottom int) {
	l := s.lay()
	return l.Main.Y, max(s.h-l.Footer.Y, 0)
}

// resize re-lays the screen out; it runs on every WindowSizeMsg.
func (s *Shell) resize(w, h int) {
	s.w, s.h = w, h
	s.selection = s.selection.Clear()
	if s.relpick != nil {
		placed := s.relpick.Within(s.dialogMargins()).Place(w, h)
		s.relpick = &placed
	}
	if s.confirm != nil {
		// The dialog was placed for the old size: place it again, so a window
		// that settles after the dialog opened still shows the whole command.
		placed := s.confirm.Within(s.dialogMargins()).Place(w, h)
		s.confirm = &placed
	}
	if s.fix != nil {
		placed := s.fix.Place(w, h)
		s.fix = &placed
	}
	if s.adopt != nil {
		placed := s.adopt.Place(w, h)
		s.adopt = &placed
	}
	if s.tools != nil {
		placed := s.tools.Place(w, h)
		s.tools = &placed
	}
	if s.compose != nil {
		placed := s.compose.Place(w, h)
		s.compose = &placed
	}
	if s.result != nil {
		placed := s.result.Place(w, h)
		s.result = &placed
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
