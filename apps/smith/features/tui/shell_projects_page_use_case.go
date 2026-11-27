package tui

import (
	"path/filepath"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/palette"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/projectpage"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/settingsview"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The keys, the clicks and the palette entries of the Projects view: the
// filter, the sections of the detail, the Run… menu and the Open button.

// projectsKey moves in the list of projects. Enter opens the detail over the
// list on a narrow screen and moves the keyboard to it beside the list; / types
// the filter, and Esc clears it.
func (s *Shell) projectsKey(name string) tea.Cmd {
	switch commands.NavNamed(name) {
	case commands.NavUp:
		s.list = s.list.Move(-1)
	case commands.NavDown:
		s.list = s.list.Move(1)
	case commands.NavPageUp:
		s.list = s.list.Page(-1, s.listPage())
	case commands.NavPageDown:
		s.list = s.list.Page(1, s.listPage())
	case commands.NavHome:
		s.list = s.list.Home()
	case commands.NavEnd:
		s.list = s.list.End()
	case commands.NavConfirm:
		if _, ok := s.list.Selected(); !ok {
			return nil
		}
		if s.lay().Mode == layout.Narrow {
			s.detailOpen, s.detailCursor = true, 0
			return nil
		}
		s.focusSide()
	case commands.NavCancel:
		if s.list.Filter() != "" {
			s.list = s.list.WithFilter("")
		}
	case commands.NavNext:
		s.focusSide()
	default:
		switch name {
		case "right":
			s.focusSide()
		case "/":
			s.filtering = true
		}
	}
	return nil
}

// filterKey types the filter of the list: each letter narrows the list at
// once, Enter keeps the filter and gives the keys back to the list, Esc
// clears it, and the arrows still move in what is listed.
func (s *Shell) filterKey(msg tea.KeyMsg) tea.Cmd {
	text := s.list.Filter()
	switch msg.Type {
	case tea.KeyEnter:
		s.filtering = false
		return nil
	case tea.KeyEsc:
		s.filtering, text = false, ""
	case tea.KeyBackspace:
		if r := []rune(text); len(r) > 0 {
			text = string(r[:len(r)-1])
		}
	case tea.KeyCtrlU:
		text = ""
	case tea.KeyUp:
		s.list = s.list.Move(-1)
		return nil
	case tea.KeyDown:
		s.list = s.list.Move(1)
		return nil
	case tea.KeyRunes, tea.KeySpace:
		text += string(msg.Runes)
	default:
		return nil
	}
	s.list = s.list.WithFilter(text)
	return nil
}

// projectDetailKey is a key for the detail of a project while it has the
// keyboard: Tab, Shift+Tab and the side arrows change the section, Esc gives
// the keyboard back to the list (or closes the detail that fills the pane),
// and the other keys go to the open section. In the Run… menu the keys move
// between its actions, and Esc closes the menu.
func (s *Shell) projectDetailKey(msg tea.KeyMsg) tea.Cmd {
	name := msg.String()
	act := commands.NavNamed(name)
	if s.projMenu || !s.projectPage() {
		if s.projMenu && (act == commands.NavCancel || name == "backspace") {
			s.projMenu, s.detailTop = false, 0
			s.chooseDetail(0)
			return nil
		}
		return s.detailKey(name)
	}
	settings := s.projSection == projectpage.Settings
	if !settings || !s.settings.Modal() {
		switch {
		case name == "tab" || name == "right":
			s.showSection(s.projSection.Step(1))
			return nil
		case name == "shift+tab" || name == "left":
			s.showSection(s.projSection.Step(-1))
			return nil
		case act == commands.NavCancel || (name == "backspace" && !settings):
			if s.detailFull() {
				s.detailOpen, s.zoom = false, false
			} else {
				s.focusList()
			}
			return nil
		case name == "/":
			return s.activate("file.find")
		}
	}
	if settings {
		return s.projectSettingsKey(msg)
	}
	switch {
	case act == commands.NavPageUp:
		s.pageDetail(-1)
	case act == commands.NavPageDown:
		s.pageDetail(1)
	case s.projSection == projectpage.Dependencies && act == commands.NavUp:
		s.scrollDetail(-1)
	case s.projSection == projectpage.Dependencies && act == commands.NavDown:
		s.scrollDetail(1)
	case s.projSection == projectpage.Runs:
		rows, _ := s.projectRuns(s.selectedProject())
		switch act {
		case commands.NavUp:
			s.projRunAt = max(s.projRunAt-1, 0)
		case commands.NavDown:
			s.projRunAt = min(s.projRunAt+1, max(len(rows)-1, 0))
		case commands.NavConfirm:
			return s.openProjectRun()
		}
	case s.projSection == projectpage.Tasks:
		p, _ := s.list.Selected()
		switch act {
		case commands.NavUp:
			s.taskAt = max(s.taskAt-1, 0)
		case commands.NavDown:
			s.taskAt = min(s.taskAt+1, max(len(p.Tasks)-1, 0))
		case commands.NavHome:
			s.taskAt = 0
		case commands.NavEnd:
			s.taskAt = max(len(p.Tasks)-1, 0)
		case commands.NavConfirm:
			return s.runTaskAt()
		}
	}
	return nil
}

// showSection opens a section of the detail of a project, at its top.
func (s *Shell) showSection(sec projectpage.Section) {
	s.projSection, s.projMenu, s.detailTop = sec, false, 0
}

// runTaskAt runs the task the keyboard is on in the Tasks section, as its key
// does, unless a run is going.
func (s *Shell) runTaskAt() tea.Cmd {
	p, ok := s.list.Selected()
	if !ok || s.taskAt >= len(p.Tasks) {
		return nil
	}
	t := p.Tasks[s.taskAt]
	run, _ := commands.ByID("run.build")
	if reason := run.Unavailable(s.state()); reason != "" {
		s.say(words.Say("tui.keys.is_unavailable", "title", taskLabel(t.Name, p.Kind), "reason", reason))
		return nil
	}
	return s.activate(s.taskAction(t.Name, s.taskKeys(p)[t.Name], p.Kind).ID)
}

// openProjectRun opens the run the keyboard is on in the Runs section, in the
// run panel.
func (s *Shell) openProjectRun() tea.Cmd {
	_, at := s.projectRuns(s.selectedProject())
	if s.projRunAt >= len(at) {
		return nil
	}
	s.recentAt = len(s.runs) - 1 - at[s.projRunAt]
	return s.openRecent()
}

// openRunMenu shows the Run… menu in the detail: every action of the project,
// its tasks first. The keyboard moves to it.
func (s *Shell) openRunMenu() {
	if _, ok := s.list.Selected(); !ok {
		s.say(words.Say("tui.projects.reason.no_project"))
		return
	}
	s.projMenu, s.detailTop = true, 0
	s.chooseDetail(0)
	if s.lay().Mode == layout.Narrow {
		s.detailOpen = true
		return
	}
	s.focusSide()
}

// openProjectSettings shows the Settings section of the detail of the
// selected project: the editor of [projects.<name>] in smith.toml, which the
// Settings tab used to hold. The keyboard moves to it.
func (s *Shell) openProjectSettings() tea.Cmd {
	if _, ok := s.list.Selected(); !ok {
		return nil
	}
	s.showSection(projectpage.Settings)
	if s.lay().Mode == layout.Narrow {
		s.detailOpen = true
	} else {
		s.focusSide()
	}
	return s.syncProjectSettings()
}

// syncProjectSettings reads the settings of the selected project when its
// Settings section is on screen and shows another project, or none yet.
func (s *Shell) syncProjectSettings() tea.Cmd {
	if !s.projectPage() || s.projSection != projectpage.Settings || !s.detailVisible() {
		return nil
	}
	p, _ := s.list.Selected()
	if s.settings.ProjectOpen() == p.Name {
		return nil
	}
	var cmd tea.Cmd
	s.settings, cmd = s.settings.OpenProject(settingsview.ProjectRef{Name: p.Name, Root: p.Root, Language: p.Language, Kind: p.Kind, Version: p.Version})
	return cmd
}

// projectSettingsKey gives a key to the Settings section of the detail.
func (s *Shell) projectSettingsKey(msg tea.KeyMsg) tea.Cmd {
	if why := s.settingsRefusal(msg.String()); why != "" {
		s.say(why)
		return nil
	}
	var res settingsview.Result
	s.settings, res = s.settings.ProjectKey(msg)
	return s.settingsResult(res)
}

// projectHeadRows is how many rows the top of the detail of the selected
// project takes in w cells.
func (s Shell) projectHeadRows(w int) int {
	p, _ := s.list.Selected()
	head, _ := s.projectHead(p, w)
	return len(head)
}

// projectSettingsClick is a click on row y of the Settings section.
func (s *Shell) projectSettingsClick(y int, double bool) tea.Cmd {
	r := s.detailRect()
	h := max(r.H-s.toastRows(r.H)-s.projectHeadRows(r.W), 1)
	var res settingsview.Result
	s.settings, res = s.settings.ProjectClick(r.W, h, y, double)
	return s.settingsResult(res)
}

// scrollProjectDetail moves the open section of the detail by delta lines.
func (s *Shell) scrollProjectDetail(delta int) {
	if s.projSection == projectpage.Settings {
		s.settings = s.settings.Wheel(max(min(delta, 1), -1))
		return
	}
	p, _ := s.list.Selected()
	r := s.detailRect()
	body, _, _ := s.sectionBody(p, r.W)
	s.detailTop = pane.ClampTop(s.detailTop+delta, len(body), max(r.H-s.projectHeadRows(r.W), 1))
}

// openProjectFolder opens the folder of the selected project with the program
// the system chooses.
func (s *Shell) openProjectFolder() tea.Cmd {
	p, ok := s.list.Selected()
	open := s.sources.Open
	if !ok || open == nil {
		return nil
	}
	path := p.Root
	if s.wsRoot != "" && !filepath.IsAbs(path) {
		path = filepath.Join(s.wsRoot, path)
	}
	return func() tea.Msg { return openedMsg{path: path, folder: true, err: open(path)} }
}

// projectReasons are the keys of Projects that do nothing now, and why: Enter
// in a section with nothing to run or open, and Run… with no project.
func (s Shell) projectReasons() map[string]string {
	r := map[string]string{}
	if s.projectKeys() {
		p, _ := s.list.Selected()
		rows, _ := s.projectRuns(p.Name)
		switch {
		case s.projSection == projectpage.Dependencies:
			r["projects.detail-enter"] = words.Say("tui.projects.reason.nothing_to_run")
		case s.projSection == projectpage.Tasks && len(p.Tasks) == 0:
			r["projects.detail-enter"] = words.Say("tui.projects.reason.no_task")
		case s.projSection == projectpage.Runs && len(rows) == 0:
			r["projects.detail-enter"] = words.Say("tui.projects.reason.no_run")
		}
	}
	if _, ok := s.list.Selected(); !ok {
		r["projects.run-menu"] = words.Say("tui.projects.reason.no_project")
	}
	return r
}

// labels name, for the footer, what a key does now in place of its usual
// words: Enter on Home names its action, Enter in the sections of a project
// runs, opens or changes.
func (s Shell) labels() map[string]string {
	l := s.homeLabels()
	if !s.projectKeys() {
		return l
	}
	if l == nil {
		l = map[string]string{}
	}
	switch s.projSection {
	case projectpage.Runs:
		l["projects.detail-enter"] = words.Say("commands.projects.detail-enter.footer.open")
	case projectpage.Settings:
		l["projects.detail-enter"] = words.Say("commands.projects.detail-enter.footer.change")
	}
	return l
}

// withProjectEntries adds to the palette, on Projects, the actions of the
// selected project that no command of the palette does (explain, why, the
// forms about the project): the Run… menu lists them too. Its settings are
// a section of the detail, not an entry.
func (s Shell) withProjectEntries(entries []palette.Entry) []palette.Entry {
	if s.tab != commands.TabProjects || !s.inWorkspace {
		return entries
	}
	p, ok := s.list.Selected()
	if !ok {
		return entries
	}
	panel := s.projectPanel()
	var extra []palette.Entry
	for _, a := range append(append([]detail.Action(nil), panel.Actions...), panel.More...) {
		if !strings.HasPrefix(a.ID, "x:") || strings.HasPrefix(a.ID, "x:run:") || a.ID == "x:install" || a.ID == "x:uninstall" || a.ID == "x:settings" {
			continue
		}
		why := ""
		if a.Disabled {
			why = words.Say("tui.projects.reason.disabled")
		}
		extra = append(extra, palette.Entry{ID: a.ID, Section: commands.SectionProjects, Title: words.Say("tui.projects.palette_entry", "label", a.Label, "project", p.Name), Hint: a.Hint, Unavailable: why})
	}
	at := len(entries)
	for i, e := range entries {
		if e.Section == commands.SectionProjects {
			at = i + 1
		}
	}
	return slices.Insert(slices.Clone(entries), at, extra...)
}
