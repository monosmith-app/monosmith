package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
	"github.com/monosmith-app/monosmith/libs/actions/catalog"
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
const addProjectTitle = "Add a project"

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
		s.notify("This smith cannot read that here: run "+commands.Line(argv)+" in a terminal", toast.Warning)
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
		s.tell("Cannot "+strings.ToLower(title), "smith said:", failureText(msg.res))
		return nil
	}
	c := dialog.NewConfirm(title, why, catalog.WithConfirmation(argv)).
		WithDetails(capLines(withoutDryRunLine(textLines(msg.res.Text)), max(s.h-18, 3))).Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &c
	return nil
}

// dryRunLine ends the plan a command prints with --dry-run.
const dryRunLine = "Dry run: nothing was changed."

// withoutDryRunLine drops the closing "Dry run" line of a plan: the dialog
// above it already says that nothing is written yet, and a first-time user
// does not know the words "dry run".
//
// Parameters: lines are the lines of the plan.
// Returns: the lines without the closing "Dry run" line.
func withoutDryRunLine(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == dryRunLine {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// tell shows a notice dialog: a title, a sentence and the lines of a text.
func (s *Shell) tell(title, why, text string) {
	n := dialog.NewNotice(title, why, capLines(textLines(text), max(s.h-12, 3))).Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &n
}

// textLines splits a command's output into lines and drops the blank ones at
// the ends.
func textLines(text string) []string {
	return strings.Split(strings.Trim(text, "\n"), "\n")
}

// failureText is what to show for a command that failed: its output, or why
// it never ran.
func failureText(res sources.Captured) string {
	switch {
	case strings.TrimSpace(res.Text) != "":
		return res.Text
	case res.Err != nil:
		return res.Err.Error()
	}
	return "the command ended with an error and printed nothing"
}

// openUndo reads what the last change did, before anything is reverted.
func (s *Shell) openUndo() tea.Cmd {
	return s.captureInDialog(capUndo, undoTitle, undoWhy, undoArgv, []string{"undo", "--dry-run"}, "Reading what the last change did")
}

// The words of the undo confirmation, shown while it waits and when it is read.
const (
	undoTitle = "Undo the last change?"
	undoWhy   = "smith reverts it with a new commit, so your history stays. This is what it would revert:"
)

// undoArgv is the command line that reverts the last change.
var undoArgv = []string{"undo", "--yes"}

// undoRead shows what smith undo would revert and asks; when there is nothing
// to revert, the reason smith gave is the notice.
func (s *Shell) undoRead(msg capturedMsg) tea.Cmd {
	if !msg.res.OK {
		s.tell("Nothing to undo", "smith said:", failureText(msg.res))
		return nil
	}
	c := dialog.NewConfirm(undoTitle, undoWhy, slices.Clone(undoArgv)).
		WithDetails(capLines(withoutDryRunLine(textLines(msg.res.Text)), max(s.h-16, 3))).Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &c
	return nil
}

// addPreview asks smith what adding a project would do, before it does it.
func (s *Shell) addPreview(argv []string) tea.Cmd {
	if s.sources.Capture == nil {
		return s.execute(argv, false)
	}
	return s.captureInDialog(capAdd, addPlanTitle(argv), addPlanWhy(argv, s.list.Len(), s.offers), argv, append(slices.Clone(argv), "--dry-run"), "Planning the new project: nothing is written yet")
}

// addPlanTitle is the question of the add confirmation.
func addPlanTitle(argv []string) string {
	name := "the project"
	if len(argv) > 2 {
		name = argv[2]
	}
	return "Add " + name + "?"
}

// addPlanRead shows the plan of an add, with the sibling it follows, and asks
// before the real add runs. The error of a wrong combination is shown as it is.
func (s *Shell) addPlanRead(msg capturedMsg) tea.Cmd {
	argv := msg.argv[:len(msg.argv)-1]
	if !msg.res.OK {
		name := "the project"
		if len(argv) > 2 {
			name = argv[2]
		}
		s.tell("Cannot add "+name+" like that", "smith said:", failureText(msg.res))
		return nil
	}
	c := dialog.NewConfirm(addPlanTitle(argv), addPlanWhy(argv, s.list.Len(), s.offers), argv).
		WithDetails(capLines(withoutDryRunLine(textLines(msg.res.Text)), max(s.h-18, 3))).Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &c
	return nil
}

// swapTitle is the title of the guided form that swaps a choice of a project.
const swapTitle = "Swap a choice"

// swapPreview asks smith what a swap would do, before it does it.
func (s *Shell) swapPreview(argv []string) tea.Cmd {
	if s.sources.Capture == nil {
		return s.execute(append(slices.Clone(argv), "--yes"), false)
	}
	return s.captureInDialog(capSwap, swapPlanTitle(argv), swapPlanWhy, append(slices.Clone(argv), "--yes"), append(slices.Clone(argv), "--dry-run"), "Planning the swap: nothing is changed yet")
}

// swapPlanWhy is the sentence of the swap confirmation.
const swapPlanWhy = "This is the plan. smith moves what it can and lists what is left to do by hand."

// swapPlanTitle is the question of the swap confirmation.
func swapPlanTitle(argv []string) string {
	if len(argv) > 3 {
		return "Swap " + argv[2] + " of " + argv[1] + " to " + argv[3] + "?"
	}
	return "Swap the choice?"
}

// swapPlanRead shows the plan of a swap, with what it moves and what is left
// by hand, and asks before the real swap runs. The commit box is offered
// unticked. The reason of a refused swap is shown as it is.
func (s *Shell) swapPlanRead(msg capturedMsg) tea.Cmd {
	argv := msg.argv[:len(msg.argv)-1]
	if !msg.res.OK {
		s.tell("Cannot swap that", "smith said:", failureText(msg.res))
		return nil
	}
	c := dialog.NewConfirm(swapPlanTitle(argv), swapPlanWhy, append(slices.Clone(argv), "--yes")).
		WithDetails(capLines(withoutDryRunLine(textLines(msg.res.Text)), max(s.h-18, 3))).Within(s.dialogMargins()).Place(s.w, s.h)
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
		if base, ok := g.ActionTitled(swapTitle); ok {
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
	if g, ok := g.WithDefault("project", s.selectedProject()).Pick(swapTitle); ok {
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
	s.say("Reading the choices your plugins offer…")
	return s.capture(capChoices, []string{"choices", "--json"})
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
		s.notify("Could not read the choices, so the form has none: "+firstLine(failureText(msg.res)), toast.Warning)
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
		if base, ok := g.ActionTitled(addProjectTitle); ok {
			s.addBase = &base
		}
	}
	if s.addBase != nil {
		g = g.WithAction(guide.WithWorkspaceEnvironments(guide.AddProjectAction(*s.addBase, s.offers, s.projectNames()), s.environments))
	}
	g = g.WithDefaultWhy("lang", s.mainLanguageWhy())
	if g, ok := g.WithDefault("lang", s.mainLanguage()).Pick(addProjectTitle); ok {
		s.guide = g
	}
	return s.gotoTab(commands.TabGuide)
}

// recipeTitle is the title of the guided form that writes a recipe.
const recipeTitle = "Write a recipe"

// openRecipeForm opens the form that writes a recipe, with a pick list or tick
// boxes for each level, read from `smith recipe new --list-levels --json` the
// first time. When they cannot be read, the plain form opens and says why.
func (s *Shell) openRecipeForm() tea.Cmd {
	if s.sources.Capture == nil || s.recipeLevels != nil {
		return s.showRecipeForm()
	}
	s.say("Reading the plugins a recipe can fix…")
	return s.capture(capRecipeLevels, []string{"recipe", "new", "--list-levels", "--json"})
}

// recipeLevelsRead keeps the levels and opens the form that waited for them.
func (s *Shell) recipeLevelsRead(msg capturedMsg) tea.Cmd {
	if msg.res.OK {
		if cat, err := guide.ParseRecipeLevels(msg.res.Text); err == nil {
			s.recipeLevels = &cat
		}
	}
	if s.recipeLevels == nil {
		s.notify("Could not read the plugins, so the form asks for level=plugin text: "+firstLine(failureText(msg.res)), toast.Warning)
	}
	return s.showRecipeForm()
}

// showRecipeForm opens the Write a recipe form.
func (s *Shell) showRecipeForm() tea.Cmd {
	g := s.guide
	if s.recipeBase == nil {
		if base, ok := g.ActionTitled(recipeTitle); ok {
			s.recipeBase = &base
		}
	}
	if s.recipeBase != nil && s.recipeLevels != nil {
		g = g.WithAction(guide.RecipeAction(*s.recipeBase, *s.recipeLevels))
	}
	if g, ok := g.Pick(recipeTitle); ok {
		s.guide = g
	}
	return s.gotoTab(commands.TabGuide)
}

// removeProjectTitle is the title of the guided form that removes a project.
const removeProjectTitle = "Remove a project"

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
		base, ok := s.guide.ActionTitled(removeProjectTitle)
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
		if base, ok := g.ActionTitled(removeProjectTitle); ok {
			s.removeBase = &base
		}
	}
	if s.removeBase != nil {
		g = g.WithAction(guide.RemoveProjectAction(*s.removeBase, s.projectNames()))
	}
	if g, ok := g.WithDefault("name", s.selectedProject()).Pick(removeProjectTitle); ok {
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
	s.say("Searching the plugin marketplace…")
	return s.capture(capSearch, append(argv, "--json"))
}

// searchRead lists the plugins found as a form; the one picked is installed
// after the usual confirmation.
func (s *Shell) searchRead(msg capturedMsg) tea.Cmd {
	if !msg.res.OK {
		s.tell("Could not search the marketplace", "smith said:", failureText(msg.res))
		return nil
	}
	hits, err := guide.ParsePluginHits(msg.res.Text)
	if err != nil {
		s.tell("Could not read the search", err.Error(), "")
		return nil
	}
	if len(hits) == 0 {
		s.notify("No plugin matches. Search with fewer words, or leave the field empty to list them all.", toast.Info)
		return nil
	}
	action := guide.PluginInstallAction(hits)
	g := s.guide.WithAction(action)
	if g, ok := g.Pick(action.Title); ok {
		s.guide = g
	}
	return s.gotoTab(commands.TabGuide)
}

// reportPath is where the report of this workspace is written: a file in the
// temporary folder, named after the workspace.
func (s Shell) reportPath() string {
	name := strings.Map(func(r rune) rune {
		if r == '-' || r == '_' || (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return r
		}
		return '-'
	}, s.header)
	if strings.Trim(name, "-") == "" {
		name = "workspace"
	}
	return filepath.Join(os.TempDir(), "smith-report-"+name+".html")
}

// openReport writes the report to a temporary file; opening it follows when
// the command ends (see reportWritten).
func (s *Shell) openReport() tea.Cmd {
	return s.execute([]string{"report", "-o", s.reportPath()}, false)
}

// reportWritten opens the report a finished `smith report -o` wrote, and says
// where the file is so it can be opened again.
func (s *Shell) reportWritten(argv []string) tea.Cmd {
	path, ok := reportTarget(argv)
	if !ok {
		return nil
	}
	open := s.sources.Open
	if open == nil {
		s.notify("Report written to "+path+". Open it in a browser.", toast.Success)
		return nil
	}
	return func() tea.Msg { return openedMsg{path: path, err: open(path)} }
}

// opened tells the person where the report is.
func (s *Shell) opened(msg openedMsg) {
	if msg.link {
		if msg.err != nil {
			s.notify("Could not open the page. The link is copied: paste it in a browser.", toast.Warning)
			return
		}
		s.notify("Opened "+msg.path+" in your browser. The link is copied too.", toast.Success)
		return
	}
	if msg.err != nil {
		s.notify(fmt.Sprintf("Report written to %s, but the browser did not open (%v). Open that file yourself.", msg.path, msg.err), toast.Warning)
		return
	}
	s.notify("Report opened in your browser. The file is "+msg.path, toast.Success)
}

// reportTarget is the file a `smith report -o <file>` command line writes.
func reportTarget(argv []string) (string, bool) {
	if len(argv) == 0 || argv[0] != "report" {
		return "", false
	}
	for i, a := range argv {
		if (a == "-o" || a == "--output") && i+1 < len(argv) {
			return argv[i+1], true
		}
	}
	return "", false
}

// firstLine is the first non-blank line of a text.
func firstLine(text string) string {
	for _, l := range strings.Split(text, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			return l
		}
	}
	return ""
}

// undoHint is the sentence a toast adds to a change that can be undone.
func undoHint() string {
	return "It can be undone: open the palette (" + commands.ShortcutOf("file.palette") + ") and choose Undo the last change"
}

// notCommittedHint is the sentence a toast adds to a change that is not a commit.
func notCommittedHint() string {
	return "Not committed: review with git status"
}

// saidNotCommitted is true when a command asked to commit but printed that it
// did not (for example because the work tree had other changes): there is then
// nothing for smith undo to revert.
//
// Parameters: lines are the lines the command printed.
// Returns: whether one of them starts with "Not committed".
func saidNotCommitted(lines []string) bool {
	return slices.ContainsFunc(lines, func(l string) bool { return strings.HasPrefix(strings.TrimSpace(l), "Not committed") })
}

// changesFiles is true for the commands that write files and can commit them.
func changesFiles(argv []string) bool { return dialog.CanCommit(argv) }

// undoable is true for a command whose work `smith undo` can revert: one that
// was run with --commit, or in a workspace that sets commit_operations.
func undoable(argv []string, commitOps bool) bool {
	return dialog.CanCommit(argv) && (commitOps || slices.Contains(argv, "--commit"))
}
