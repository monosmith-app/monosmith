package tui

import (
	"context"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commandtext"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
	"github.com/monosmith-app/monosmith/libs/actions/catalog"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// What a captured command was asked for.
const (
	capUndo    = "undo"
	capAdd     = "add"
	capSwap    = "swap"
	capPlan    = "plan"
	capChoices = "choices"
	capSearch  = "plugin-search"

	capRecipeLevels = "recipe-levels"
)

// The states of the choices the Add form offers.
type choicesState int

const (
	// choicesNotAsked: smith choices has not run yet (or must run again).
	choicesNotAsked choicesState = iota
	// choicesAsked: it is running.
	choicesAsked
	// choicesKnown: the offers are read.
	choicesKnown
)

// addProjectTitle is the title of the guided form that adds a project.
func addProjectTitle() string { return words.Say("tui.capture.add_project_title") }

// capturedMsg is the end of a command the shell ran to read its text (a plan,
// a list, an error) and not to show in the Run tab.
type capturedMsg struct {
	purpose string
	argv    []string
	res     sources.Captured
	// ticket ties the end of a capture to the working dialog that waits for it.
	ticket int
}

// openedMsg says whether the system opened the report.
type openedMsg struct {
	path string
	// link is true when path is an address the failure dialog opened, not the report.
	link bool
	err  error
}

// capture runs a smith command line in the background and reports its text as
// a capturedMsg.
func (s *Shell) capture(purpose string, argv []string) tea.Cmd {
	run := s.sources.Capture
	if run == nil {
		s.notify(words.Say("tui.capture.this_smith_cannot_read_that_here", "line", commands.Line(argv)), toast.Warning)
		return nil
	}
	return func() tea.Msg { return capturedMsg{purpose: purpose, argv: argv, res: run(context.Background(), argv)} }
}

// captured does what a finished capture was for.
func (s *Shell) captured(msg capturedMsg) tea.Cmd {
	if waitsForTicket(msg.purpose) && !s.keepsPlan(msg) {
		// The person cancelled the dialog that waited for this plan.
		return nil
	}
	switch msg.purpose {
	case capUndo:
		return s.undoRead(msg)
	case capAdd:
		return s.addPlanRead(msg)
	case capSwap:
		return s.swapPlanRead(msg)
	case capPlan:
		return s.planRead(msg)
	case capChoices:
		return s.choicesRead(msg)
	case capRecipeLevels:
		return s.recipeLevelsRead(msg)
	case capCompose:
		return s.composeRead(msg)
	case capComposePlan:
		return s.composePlanRead(msg)
	case capTools:
		return s.toolsRead(msg)
	case capSearch:
		return s.searchRead(msg)
	}
	return nil
}

// planRead shows the plan a command printed with its plan flag (--dry-run or
// --check) as the details of the confirmation, and asks before the real
// command runs. The reason of a refused plan is shown as it is.
func (s *Shell) planRead(msg capturedMsg) tea.Cmd {
	argv := msg.argv[:len(msg.argv)-1]
	title, why := commands.Describe(argv)
	if !msg.res.OK {
		s.tell(words.Say("tui.capture.cannot", "title", strings.ToLower(title)), words.Say("tui.capture.smith_said"), commandtext.Failure(msg.res.Text, msg.res.Err))
		return nil
	}
	c := dialog.NewConfirm(title, why, catalog.WithConfirmation(argv)).
		WithDetails(capLines(commandtext.WithoutDryRunLine(commandtext.Lines(msg.res.Text)), max(s.h-18, 3))).Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &c
	return nil
}

// tell shows a notice dialog: a title, a sentence and the lines of a text.
func (s *Shell) tell(title, why, text string) {
	n := dialog.NewNotice(title, why, capLines(commandtext.Lines(text), max(s.h-12, 3))).Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &n
}

// openUndo reads what the last change did, before anything is reverted.
func (s *Shell) openUndo() tea.Cmd {
	return s.captureInDialog(capUndo, undoTitle(), undoWhy(), undoArgv, []string{"undo", "--dry-run"}, words.Say("tui.capture.reading_what_the_last_change_did"))
}

// undoTitle and undoWhy are the words of the undo confirmation, shown while it
// waits and when it is read.
func undoTitle() string { return words.Say("tui.capture.undo_title") }

func undoWhy() string { return words.Say("tui.capture.undo_why") }

// undoArgv is the command line that reverts the last change.
var undoArgv = []string{"undo", "--yes"}

// undoRead shows what smith undo would revert and asks; when there is nothing
// to revert, the reason smith gave is the notice.
func (s *Shell) undoRead(msg capturedMsg) tea.Cmd {
	if !msg.res.OK {
		s.tell(words.Say("tui.capture.nothing_to_undo"), words.Say("tui.capture.smith_said"), commandtext.Failure(msg.res.Text, msg.res.Err))
		return nil
	}
	c := dialog.NewConfirm(undoTitle(), undoWhy(), slices.Clone(undoArgv)).
		WithDetails(capLines(commandtext.WithoutDryRunLine(commandtext.Lines(msg.res.Text)), max(s.h-16, 3))).Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &c
	return nil
}

// addPreview asks smith what adding a project would do, before it does it.
func (s *Shell) addPreview(argv []string) tea.Cmd {
	if s.sources.Capture == nil {
		return s.execute(argv, false)
	}
	return s.captureInDialog(capAdd, addPlanTitle(argv), addPlanWhy(argv, s.list.Len(), s.offers), argv, append(slices.Clone(argv), "--dry-run"), words.Say("tui.capture.planning_the_new_project_nothing_is"))
}

// addPlanTitle is the question of the add confirmation.
func addPlanTitle(argv []string) string {
	name := words.Say("tui.capture.the_project")
	if len(argv) > 2 {
		name = argv[2]
	}
	return words.Say("tui.capture.add", "name", name)
}

// addPlanRead shows the plan of an add, with the sibling it follows, and asks
// before the real add runs. The error of a wrong combination is shown as it is.
func (s *Shell) addPlanRead(msg capturedMsg) tea.Cmd {
	argv := msg.argv[:len(msg.argv)-1]
	if !msg.res.OK {
		name := words.Say("tui.capture.the_project")
		if len(argv) > 2 {
			name = argv[2]
		}
		s.tell(words.Say("tui.capture.cannot_add_like_that", "name", name), words.Say("tui.capture.smith_said"), commandtext.Failure(msg.res.Text, msg.res.Err))
		return nil
	}
	c := dialog.NewConfirm(addPlanTitle(argv), addPlanWhy(argv, s.list.Len(), s.offers), argv).
		WithDetails(capLines(commandtext.WithoutDryRunLine(commandtext.Lines(msg.res.Text)), max(s.h-18, 3))).Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &c
	return nil
}

// swapTitle is the title of the guided form that swaps a choice of a project.
func swapTitle() string { return words.Say("tui.capture.swap_title") }

// swapPreview asks smith what a swap would do, before it does it.
func (s *Shell) swapPreview(argv []string) tea.Cmd {
	if s.sources.Capture == nil {
		return s.execute(append(slices.Clone(argv), "--yes"), false)
	}
	return s.captureInDialog(capSwap, swapPlanTitle(argv), swapPlanWhy(), append(slices.Clone(argv), "--yes"), append(slices.Clone(argv), "--dry-run"), words.Say("tui.capture.planning_the_swap_nothing_is_changed"))
}

// swapPlanWhy is the sentence of the swap confirmation.
func swapPlanWhy() string { return words.Say("tui.capture.swap_plan_why") }

// swapPlanTitle is the question of the swap confirmation.
func swapPlanTitle(argv []string) string {
	if len(argv) > 3 {
		return words.Say("tui.capture.swap_of_to", "argv", argv[2], "argv2", argv[1], "argv3", argv[3])
	}
	return words.Say("tui.capture.swap_the_choice")
}

// swapPlanRead shows the plan of a swap, with what it moves and what is left
// by hand, and asks before the real swap runs. The commit box is offered
// unticked. The reason of a refused swap is shown as it is.
func (s *Shell) swapPlanRead(msg capturedMsg) tea.Cmd {
	argv := msg.argv[:len(msg.argv)-1]
	if !msg.res.OK {
		s.tell(words.Say("tui.capture.cannot_swap_that"), words.Say("tui.capture.smith_said"), commandtext.Failure(msg.res.Text, msg.res.Err))
		return nil
	}
	c := dialog.NewConfirm(swapPlanTitle(argv), swapPlanWhy(), append(slices.Clone(argv), "--yes")).
		WithDetails(capLines(commandtext.WithoutDryRunLine(commandtext.Lines(msg.res.Text)), max(s.h-18, 3))).Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &c
	return nil
}

// openSwap opens the form that swaps a choice of a project, with the choices
// the plugins offer. The choices are read the first time, which takes a moment.
func (s *Shell) openSwap() tea.Cmd {
	return s.openChoiceForm(formSwap)
}

// showSwapForm opens the Swap a choice form.
func (s *Shell) showSwapForm() tea.Cmd {
	g := s.guide
	if s.swapBase == nil {
		if base, ok := g.ActionTitled(swapTitle()); ok {
			s.swapBase = &base
		}
	}
	if s.swapBase != nil {
		var projects []guide.SwapProject
		for _, p := range s.list.Projects() {
			projects = append(projects, guide.SwapProject{Name: p.Name, Language: p.Language})
		}
		g = g.WithAction(guide.SwapProjectAction(*s.swapBase, s.offers, projects))
	}
	if g, ok := g.WithDefault("project", s.selectedProject()).Pick(swapTitle()); ok {
		s.guide = g
	}
	return s.gotoTab(commands.TabGuide)
}

// The forms that wait for the choices of the plugins.
const (
	formAdd  = "add"
	formSwap = "swap"
)

// openAddProject opens the form that adds a project, with the choices the
// plugins offer. The choices are read the first time, which takes a moment.
func (s *Shell) openAddProject() tea.Cmd {
	return s.openChoiceForm(formAdd)
}

// openChoiceForm opens the form (formAdd or formSwap) that needs the choices
// of the plugins, reading them first when they are not known.
func (s *Shell) openChoiceForm(form string) tea.Cmd {
	switch {
	case s.sources.Capture == nil || s.choices == choicesKnown:
		return s.showChoiceForm(form)
	case s.choices == choicesAsked:
		s.formPending = form
		return nil
	}
	s.choices, s.formPending = choicesAsked, form
	title := addProjectTitle()
	if form == formSwap {
		title = swapTitle()
	}
	argv := []string{"choices", "--json"}
	// The form opens when the plugins have answered; until then a dialog says so.
	return s.captureWorking(capChoices, title, words.Say("tui.capture.the_form_offers_what_your_plugins"), argv, argv, words.Say("tui.capture.reading_the_choices_your_plugins_offer"), words.Say("tui.capture.cancelled_the_form_was_not_opened"))
}

// choicesRead keeps the choices and opens the form that waited for them. When
// they cannot be read, the form opens without them and says why.
func (s *Shell) choicesRead(msg capturedMsg) tea.Cmd {
	s.choices = choicesNotAsked
	if msg.res.OK {
		if offers, err := guide.ParseChoices(msg.res.Text); err == nil {
			s.offers, s.choices = offers, choicesKnown
		}
	}
	if s.choices != choicesKnown {
		s.notify(words.Say("tui.capture.could_not_read_the_choices_so", "firstLine", commandtext.FirstLine(commandtext.Failure(msg.res.Text, msg.res.Err))), toast.Warning)
	}
	if form := s.formPending; form != "" {
		s.formPending = ""
		return s.showChoiceForm(form)
	}
	return nil
}

// showChoiceForm opens the form that waited for the choices.
func (s *Shell) showChoiceForm(form string) tea.Cmd {
	if form == formSwap {
		return s.showSwapForm()
	}
	return s.showAddForm()
}

// showAddForm opens the Add a project form.
func (s *Shell) showAddForm() tea.Cmd {
	g := s.guide
	if s.addBase == nil {
		if base, ok := g.ActionTitled(addProjectTitle()); ok {
			s.addBase = &base
		}
	}
	if s.addBase != nil {
		g = g.WithAction(guide.WithWorkspaceEnvironments(guide.AddProjectAction(*s.addBase, s.offers, s.projectNames()), s.environments))
	}
	g = g.WithDefaultWhy("lang", s.mainLanguageWhy())
	if g, ok := g.WithDefault("lang", s.mainLanguage()).Pick(addProjectTitle()); ok {
		s.guide = g
	}
	return s.gotoTab(commands.TabGuide)
}

// recipeTitle is the title of the guided form that writes a recipe.
func recipeTitle() string { return words.Say("tui.capture.recipe_title") }

// openRecipeForm opens the form that writes a recipe, with a pick list or tick
// boxes for each level, read from `smith recipe new --list-levels --json` the
// first time. When they cannot be read, the plain form opens and says why.
func (s *Shell) openRecipeForm() tea.Cmd {
	if s.sources.Capture == nil || s.recipeLevels != nil {
		return s.showRecipeForm()
	}
	argv := []string{"recipe", "new", "--list-levels", "--json"}
	return s.captureWorking(capRecipeLevels, recipeTitle(), words.Say("tui.capture.the_form_lists_the_plugins_that"), argv, argv, words.Say("tui.capture.reading_the_plugins_a_recipe_can"), words.Say("tui.capture.cancelled_the_form_was_not_opened"))
}

// recipeLevelsRead keeps the levels and opens the form that waited for them.
func (s *Shell) recipeLevelsRead(msg capturedMsg) tea.Cmd {
	if msg.res.OK {
		if cat, err := guide.ParseRecipeLevels(msg.res.Text); err == nil {
			s.recipeLevels = &cat
		}
	}
	if s.recipeLevels == nil {
		s.notify(words.Say("tui.capture.could_not_read_the_plugins_so", "firstLine", commandtext.FirstLine(commandtext.Failure(msg.res.Text, msg.res.Err))), toast.Warning)
	}
	return s.showRecipeForm()
}

// showRecipeForm opens the Write a recipe form.
func (s *Shell) showRecipeForm() tea.Cmd {
	g := s.guide
	if s.recipeBase == nil {
		if base, ok := g.ActionTitled(recipeTitle()); ok {
			s.recipeBase = &base
		}
	}
	if s.recipeBase != nil && s.recipeLevels != nil {
		g = g.WithAction(guide.RecipeAction(*s.recipeBase, *s.recipeLevels))
	}
	if g, ok := g.Pick(recipeTitle()); ok {
		s.guide = g
	}
	return s.gotoTab(commands.TabGuide)
}

// removeProjectTitle is the title of the guided form that removes a project.
func removeProjectTitle() string { return words.Say("tui.capture.remove_project_title") }

// projectNames are the names of the workspace's projects, in list order.
func (s Shell) projectNames() []string {
	var names []string
	for _, p := range s.list.Projects() {
		names = append(names, p.Name)
	}
	return names
}

// webApps are the projects an environment can be added to: the apps written
// in Node. With none of them, every project is listed and smith refuses the
// one that is not a web app, with its reason.
func (s Shell) webApps() []string {
	var names []string
	for _, p := range s.list.Projects() {
		if p.Kind == "app" && p.Language == "node" {
			names = append(names, p.Name)
		}
	}
	if len(names) == 0 {
		return s.projectNames()
	}
	return names
}

// openEnvironments opens the form that adds environments to a web app or
// removes them, with the web apps as a pick list that starts on the selected
// project.
func (s *Shell) openEnvironments(title string) tea.Cmd {
	g := s.guide
	if base, ok := g.ActionTitled(title); ok {
		g = g.WithAction(guide.WithProjectPick(base, "project", s.webApps()))
	}
	if g, ok := g.WithDefault("project", s.selectedProject()).Pick(title); ok {
		s.guide = g
	}
	return s.gotoTab(commands.TabGuide)
}

// refreshProjectForms puts the project names into the Remove form of the menu,
// so a form opened from the Guide tab lists the projects too.
func (s *Shell) refreshProjectForms() {
	if s.removeBase == nil {
		base, ok := s.guide.ActionTitled(removeProjectTitle())
		if !ok {
			return
		}
		s.removeBase = &base
	}
	s.guide = s.guide.WithAction(guide.RemoveProjectAction(*s.removeBase, s.projectNames()))
}

// openRemoveProject opens the form that removes a project as a pick list that
// starts on the selected project.
func (s *Shell) openRemoveProject() tea.Cmd {
	g := s.guide
	if s.removeBase == nil {
		if base, ok := g.ActionTitled(removeProjectTitle()); ok {
			s.removeBase = &base
		}
	}
	if s.removeBase != nil {
		g = g.WithAction(guide.RemoveProjectAction(*s.removeBase, s.projectNames()))
	}
	if g, ok := g.WithDefault("name", s.selectedProject()).Pick(removeProjectTitle()); ok {
		s.guide = g
	}
	return s.gotoTab(commands.TabGuide)
}

// openPluginSearch searches the plugin marketplace for a term.
func (s *Shell) openPluginSearch(term string) tea.Cmd {
	argv := []string{"plugin", "search"}
	if term != "" {
		argv = append(argv, strings.Fields(term)...)
	}
	argv = append(argv, "--json")
	return s.captureWorking(capSearch, words.Say("tui.capture.search_the_plugin_marketplace"), words.Say("tui.capture.smith_asks_the_marketplace_for_the"), argv, argv, words.Say("tui.capture.searching_the_plugin_marketplace"), words.Say("tui.capture.cancelled_the_search_was_stopped"))
}

// searchRead lists the plugins found as a form; the one picked is installed
// after the usual confirmation.
func (s *Shell) searchRead(msg capturedMsg) tea.Cmd {
	if !msg.res.OK {
		s.tell(words.Say("tui.capture.could_not_search_the_marketplace"), words.Say("tui.capture.smith_said"), commandtext.Failure(msg.res.Text, msg.res.Err))
		return nil
	}
	hits, err := guide.ParsePluginHits(msg.res.Text)
	if err != nil {
		s.tell(words.Say("tui.capture.could_not_read_the_search"), err.Error(), "")
		return nil
	}
	if len(hits) == 0 {
		s.notify(words.Say("tui.capture.no_plugin_matches_search_with_fewer"), toast.Info)
		return nil
	}
	action := guide.PluginInstallAction(hits)
	g := s.guide.WithAction(action)
	if g, ok := g.Pick(action.Title); ok {
		s.guide = g
	}
	return s.gotoTab(commands.TabGuide)
}

// openReport writes the report to a temporary file; opening it follows when
// the command ends (see reportWritten).
func (s *Shell) openReport() tea.Cmd {
	return s.execute([]string{"report", "-o", commandtext.ReportPath(s.header)}, false)
}

// reportWritten opens the report a finished `smith report -o` wrote, and says
// where the file is so it can be opened again.
func (s *Shell) reportWritten(argv []string) tea.Cmd {
	path, ok := commandtext.ReportTarget(argv)
	if !ok {
		return nil
	}
	open := s.sources.Open
	if open == nil {
		s.notify(words.Say("tui.capture.report_written_to_open_it_in", "path", path), toast.Success)
		return nil
	}
	return func() tea.Msg { return openedMsg{path: path, err: open(path)} }
}

// opened tells the person where the report is.
func (s *Shell) opened(msg openedMsg) {
	if msg.link {
		if msg.err != nil {
			s.notify(words.Say("tui.capture.could_not_open_the_page_the"), toast.Warning)
			return
		}
		s.notify(words.Say("tui.capture.opened_in_your_browser_the_link", "path", msg.path), toast.Success)
		return
	}
	if msg.err != nil {
		s.notify(words.Say("tui.capture.report_written_to_but_the_browser", "path", msg.path, "err", msg.err), toast.Warning)
		return
	}
	s.notify(words.Say("tui.capture.report_opened_in_your_browser_the", "path", msg.path), toast.Success)
}

// undoHint is the sentence a toast adds to a change that can be undone.
func undoHint() string {
	return words.Say("tui.capture.it_can_be_undone_open_the", "shortcutOf", commands.ShortcutOf("file.palette"))
}

// notCommittedHint is the sentence a toast adds to a change that is not a commit.
func notCommittedHint() string {
	return words.Say("tui.capture.not_committed_review_with_git_status")
}

// changesFiles is true for the commands that write files and can commit them.
func changesFiles(argv []string) bool { return dialog.CanCommit(argv) }

// undoable is true for a command whose work `smith undo` can revert: one that
// was run with --commit, or in a workspace that sets commit_operations.
func undoable(argv []string, commitOps bool) bool {
	return dialog.CanCommit(argv) && (commitOps || slices.Contains(argv, "--commit"))
}
