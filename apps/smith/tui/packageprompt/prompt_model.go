package packageprompt

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Prompt asks which project and which packages to install on, or remove from,
// and ends as the command line that does it: `smith install -p <project> [-D]
// <packages>` or `smith uninstall -p <project> <packages>`. It shows that
// line as it will run, as the confirmation dialog does.
//
// Keys: typing fills the packages; Left and Right on the project choose
// another project; Tab and Shift+Tab move between the project, the packages, the development box and the buttons; Space or a click flips the
// box; Enter on the packages or on Install goes ahead; Esc cancels.
type Prompt struct {
	// margins are the rows above and below the dialog that stay free.
	margins layout.Margins
	project string
	// projects are the projects to choose from (project is one of them).
	projects []string
	remove   bool
	dev      bool
	input    textinput.Model
	// languages maps a project to its language id, for the examples.
	languages map[string]string
	// focus is the packages (0), the development box (1), the buttons (2 and 3)
	// and the project (4).
	focus int
	// hover is the control under the pointer, with the same numbers, or -1.
	hover int
	// err says why Install did not go ahead.
	err  string
	rect layout.Rect
	// libraries lists the libraries of the workspace a project may link; nil
	// means the dialog does not offer the choice (and Remove never does).
	libraries func(project string) []Library
	// fromWorkspace is the first answer: a library of this workspace (true) or a
	// package from the registry (false, the default).
	fromWorkspace bool
	// libAt and rangeAt are the library and the version range picked.
	libAt, rangeAt int
	// top is the first body line shown when the body is taller than the dialog.
	top int
	// sw and sh are the screen Place sized the dialog for, so it can size
	// itself again when its body grows or shrinks.
	sw, sh int
}

const (
	focusPackages = iota
	focusDev
	focusInstall
	focusCancel
	focusProject
	focusSource
	focusLibrary
	focusRange
)

// NewInstall asks for the packages to install on project, which can be
// changed to any of projects.
func NewInstall(project string, projects []string) Prompt { return newPrompt(project, projects, false) }

// NewRemove asks for the packages to remove from project, which can be
// changed to any of projects.
func NewRemove(project string, projects []string) Prompt { return newPrompt(project, projects, true) }

func newPrompt(project string, projects []string, remove bool) Prompt {
	in := textinput.New()
	in.Prompt = "› "
	in.Placeholder = ExampleNames("")
	in.CharLimit = 200
	in.Focus()
	return Prompt{project: project, projects: append([]string(nil), projects...), remove: remove, input: in, hover: -1}
}

// Title is the heading of the dialog.
func (p Prompt) Title() string {
	if p.remove {
		return words.Say("packageprompt.title.remove")
	}
	return words.Say("packageprompt.title.install")
}

// Project is the project the packages go to or come from.
func (p Prompt) Project() string { return p.project }

// choosable is true when there is a project to change to.
func (p Prompt) choosable() bool { return len(p.projects) > 1 }

// step moves the project to the next (or, with dir -1, the previous) one.
func (p Prompt) step(dir int) Prompt {
	if !p.choosable() {
		return p
	}
	at := 0
	for i, name := range p.projects {
		if name == p.project {
			at = i
		}
	}
	p.project = p.projects[(at+dir+len(p.projects))%len(p.projects)]
	p.libAt, p.rangeAt = 0, 0
	return p.examples()
}

// WithLanguages tells the prompt each project's language id, so the example
// package names are the ones that project's package manager knows.
func (p Prompt) WithLanguages(languages map[string]string) Prompt {
	p.languages = languages
	return p.examples()
}

// examples sets the placeholder to example names for the project's language.
func (p Prompt) examples() Prompt {
	p.input.Placeholder = ExampleNames(p.languages[p.project])
	return p
}

// WithLibraries lets the dialog first ask whether to add a library of this
// workspace or a package from the registry; libraries returns the libraries
// the project may link (not itself, not an app, same language, no cycle).
func (p Prompt) WithLibraries(libraries func(project string) []Library) Prompt {
	if !p.remove {
		p.libraries = libraries
		p.focus = focusSource
		p.input.Blur()
	}
	return p
}

// offersLibraries is true when the dialog asks the first question.
func (p Prompt) offersLibraries() bool { return !p.remove && p.libraries != nil }

// linking is true when the answer is a library of this workspace.
func (p Prompt) linking() bool { return p.offersLibraries() && p.fromWorkspace }

// valid are the libraries the selected project may link.
func (p Prompt) valid() []Library {
	if !p.offersLibraries() {
		return nil
	}
	return p.libraries(p.project)
}

// library is the library picked, and false when there is none to pick.
func (p Prompt) library() (Library, bool) {
	libs := p.valid()
	if len(libs) == 0 {
		return Library{}, false
	}
	return libs[min(p.libAt, len(libs)-1)], true
}

// rangeChoices are the version ranges for the library picked.
func (p Prompt) rangeChoices() []RangeChoice {
	lib, ok := p.library()
	if !ok {
		return nil
	}
	return RangeChoices(p.languages[p.project], lib)
}

// rangeChoice is the version range picked.
func (p Prompt) rangeChoice() RangeChoice {
	choices := p.rangeChoices()
	if len(choices) == 0 {
		return RangeChoice{}
	}
	return choices[min(p.rangeAt, len(choices)-1)]
}

// Argv is the command line, without the program name, for what is typed now;
// nil while no package is typed (or no library is picked).
func (p Prompt) Argv() []string {
	if p.linking() {
		lib, ok := p.library()
		if !ok {
			return nil
		}
		argv := []string{"install", flagvocabulary.Arg("project"), p.project, "workspace:" + lib.Name}
		if spec := p.rangeChoice().Spec; spec != "" {
			argv = append(argv, "--range", spec)
		}
		return argv
	}
	names, err := ParsePackages(p.input.Value())
	if err != nil {
		return nil
	}
	if p.remove {
		return append([]string{"uninstall", flagvocabulary.Arg("project"), p.project}, names...)
	}
	argv := []string{"install", flagvocabulary.Arg("project"), p.project}
	if p.dev {
		argv = append(argv, "-D")
	}
	if p.offersLibraries() {
		// The first answer was the registry: a name that is also a project of
		// the workspace is then not asked about again.
		argv = append(argv, "--registry")
	}
	return append(argv, names...)
}

// Removes reports whether the prompt removes packages instead of installing them.
func (p Prompt) Removes() bool { return p.remove }

// Dev reports whether the packages go to the development dependencies.
func (p Prompt) Dev() bool { return p.dev }

const minWidth = 46

// buttonKeys are the keys that do what a click on the go and cancel buttons does.
var buttonKeys = [2]string{"enter", "esc"} // words:literal

// buttonCancel is the label of the button that closes the dialog.
func buttonCancel() string { return words.Say("packageprompt.button.cancel") }

func (p Prompt) goLabel() string {
	if p.remove {
		return words.Say("packageprompt.button.remove")
	}
	return words.Say("packageprompt.button.install")
}

// choiceRows are the rows that take focus after the project row, each with the
// label its text starts with, so a row is found by what it says.
func choiceRows() []struct {
	prefix string
	id     int
} {
	return []struct {
		prefix string
		id     int
	}{
		{words.Say("packageprompt.label.add"), focusSource},
		{words.Say("packageprompt.label.library"), focusLibrary},
		{words.Say("packageprompt.label.range"), focusRange},
	}
}

// rowFocus is the focus id of the choice row that line l is, or -1.
func rowFocus(l string) int {
	for _, c := range choiceRows() {
		if strings.HasPrefix(l, c.prefix) {
			return c.id
		}
	}
	return -1
}

// Within keeps the dialog off the top rows (the header and the tab strip) and
// the bottom rows (the footer and the status bar) of the screen, as every
// dialog does: Place centres it in the rows between.
//
// Parameters: top and bottom are the rows to leave free.
// Returns: the dialog with the margins set.
func (p Prompt) Within(top, bottom int) Prompt {
	p.margins = layout.Margins{Top: max(top, 0), Bottom: max(bottom, 0)}
	return p
}

// Place centres the dialog on a screen of sw by sh cells: as tall as its
// body and buttons, at most the rows between the margins. A body taller than
// that scrolls above the buttons, and the window follows the focus.
func (p Prompt) Place(sw, sh int) Prompt {
	p.sw, p.sh = sw, sh
	w := min(max(sw-8, minWidth), min(max(sw-2, 10), 76))
	p.input.Width = max(w-8, 8)
	p.rect = p.margins.Centre(sw, sh, w, len(p.body(textW(w-4)))+4)
	return p.follow()
}

// room is the rows the scrolling body gets: the inside less the blank row and
// the buttons.
func (p Prompt) room() int { return pane.BodyRoom(p.rect.H-2, 2) }

// focusLine is the body line of the control with the focus, or -1 for the
// buttons, which never scroll.
func (p Prompt) focusLine(inner int) int {
	switch p.focus {
	case focusProject:
		return 0
	case focusPackages:
		if row := p.rowOf(words.Say("packageprompt.packages.label"), inner); row >= 0 {
			return row + 1
		}
	case focusDev:
		return p.rowOf("[", inner)
	}
	for _, c := range choiceRows() {
		if c.id == p.focus {
			return p.rowOf(c.prefix, inner)
		}
	}
	return -1
}

// follow scrolls the body so the control with the focus shows; an error,
// which says why the dialog did not go ahead, wins when both do not fit.
func (p Prompt) follow() Prompt {
	inner := textW(p.rect.W - 4)
	total := len(p.body(inner))
	if line := p.focusLine(inner); line >= 0 {
		p.top = pane.KeepVisible(p.top, line, total, p.room())
	}
	if p.err != "" {
		if row := p.rowOf(p.err, inner); row >= 0 {
			p.top = pane.KeepVisible(p.top, row, total, p.room())
		}
	}
	p.top = pane.ClampTop(p.top, total, p.room())
	return p
}

// Wheel scrolls the body by delta lines; the focus stays where it is.
//
// Parameters: delta is the lines to move, negative for up.
// Returns: the dialog with the body scrolled.
func (p Prompt) Wheel(delta int) Prompt {
	p.top = pane.ClampTop(p.top+delta, len(p.body(textW(p.rect.W-4))), p.room())
	return p
}

// refit sizes the dialog again for the screen Place last saw, so a body that
// grew (an error, a longer command, the library rows) or shrank keeps the
// dialog as tall as it needs, between the margins.
func (p Prompt) refit() Prompt {
	if p.sw == 0 && p.sh == 0 {
		return p
	}
	return p.Place(p.sw, p.sh)
}

// bodyRow is the body line drawn on screen row y, through the scroll, or -1.
func (p Prompt) bodyRow(y int) int {
	return pane.BodyLine(y-p.rect.Y-1, p.top, len(p.body(textW(p.rect.W-4))), p.room())
}

// Rect is where the dialog sits after Place.
func (p Prompt) Rect() layout.Rect { return p.rect }

// rowOf finds the line of the body that starts with prefix, or -1.
func (p Prompt) rowOf(prefix string, inner int) int {
	for i, l := range p.body(inner) {
		if strings.HasPrefix(strings.TrimLeft(l, " "), prefix) {
			return i
		}
	}
	return -1
}

// buttonLabels are the labels and keys of the Install and Cancel buttons.
func (p Prompt) buttonLabels() (labels, keys []string) {
	return []string{p.goLabel(), buttonCancel()}, buttonKeys[:]
}

// buttonsRow is the span (from the dialog's left edge) of each button.
func (p Prompt) buttonsRow(inner int) [][2]int {
	labels, keys := p.buttonLabels()
	return button.RowSpans(theme.Default(), labels, keys, inner)
}

// controlAt is the control at a cell: the development box, a button, or -1.
func (p Prompt) controlAt(x, y int) int {
	r := p.rect
	inner := r.W - 4
	if y == r.Y+r.H-2 {
		if i := button.At(p.buttonsRow(inner), r, x, y); i >= 0 {
			return focusInstall + i
		}
		return -1
	}
	line := p.bodyRow(y)
	inside := x > r.X && x < r.X+r.W-1
	if line < 0 || !inside {
		return -1
	}
	if line == 0 && p.choosable() {
		return focusProject
	}
	for _, c := range choiceRows() {
		if row := p.rowOf(c.prefix, textW(inner)); row >= 0 && line == row {
			return c.id
		}
	}
	if !p.remove && !p.linking() && line == p.rowOf("[", textW(inner)) {
		return focusDev
	}
	return -1
}

// Hover highlights the control under the pointer.
func (p Prompt) Hover(x, y int) Prompt { p.hover = p.controlAt(x, y); return p }

// Click handles a left press: the box flips, Install goes ahead, Cancel and a
// press outside the dialog cancel. The dialog sizes itself again after it.
func (p Prompt) Click(x, y int) (Prompt, dialog.Outcome) {
	p, out := p.click(x, y)
	return p.refit(), out
}

// click is Click before the dialog sizes itself again.
func (p Prompt) click(x, y int) (Prompt, dialog.Outcome) {
	switch p.controlAt(x, y) {
	case focusProject:
		p.focus = focusProject
		return p.step(1), dialog.Outcome{}
	case focusDev:
		p.dev, p.focus = !p.dev, focusDev
		return p, dialog.Outcome{}
	case focusSource, focusLibrary, focusRange:
		id := p.controlAt(x, y)
		p.focus = id
		return p.cycle(id, 1), dialog.Outcome{}
	case focusInstall:
		return p.submit()
	case focusCancel:
		return p, dialog.Outcome{Done: true}
	}
	if !p.rect.Contains(x, y) {
		return p, dialog.Outcome{Done: true}
	}
	return p, dialog.Outcome{}
}

// Key handles a key. The dialog sizes itself again after it, so a body that
// grew or shrank keeps it as tall as it needs, and the focus stays in view.
func (p Prompt) Key(msg tea.KeyMsg) (Prompt, dialog.Outcome) {
	switch {
	case msg.String() == "ctrl+u" || commands.NavNamed(msg.String()) == commands.NavPageUp:
		return p.Wheel(-pane.PageStep(p.room())), dialog.Outcome{}
	case msg.String() == "ctrl+d" || commands.NavNamed(msg.String()) == commands.NavPageDown:
		return p.Wheel(pane.PageStep(p.room())), dialog.Outcome{}
	}
	p, out := p.key(msg)
	return p.refit(), out
}

// key is Key before the dialog sizes itself again.
func (p Prompt) key(msg tea.KeyMsg) (Prompt, dialog.Outcome) {
	switch commands.NavNamed(msg.String()) {
	case commands.NavCancel:
		return p, dialog.Outcome{Done: true}
	case commands.NavNext, commands.NavDown:
		return p.move(1), dialog.Outcome{}
	case commands.NavPrev, commands.NavUp:
		return p.move(-1), dialog.Outcome{}
	case commands.NavConfirm:
		switch p.focus {
		case focusProject:
			p.focus = focusPackages
			p.input.Focus()
			return p, dialog.Outcome{}
		case focusSource:
			return p.move(1), dialog.Outcome{}
		case focusDev:
			p.dev = !p.dev
			return p, dialog.Outcome{}
		case focusCancel:
			return p, dialog.Outcome{Done: true}
		}
		return p.submit()
	}
	switch {
	case p.focus == focusPackages:
		p.input, _ = p.input.Update(msg)
		p.err = ""
	case p.focus == focusSource && !p.fromWorkspace && msg.Type == tea.KeyRunes && msg.String() != " ":
		// Typing while the answer is the registry (the default) is a package
		// name: go to the field instead of losing the keystrokes.
		p.focus = focusPackages
		p.input.Focus()
		p.input, _ = p.input.Update(msg)
		p.err = ""
	case (p.focus == focusSource || p.focus == focusLibrary || p.focus == focusRange) && (msg.String() == "left" || msg.String() == "right" || msg.String() == " " || msg.String() == "space"):
		dir := 1
		if msg.String() == "left" {
			dir = -1
		}
		p = p.cycle(p.focus, dir)
	case p.focus == focusProject && msg.String() == "left":
		p = p.step(-1)
	case p.focus == focusProject && (msg.String() == "right" || msg.String() == " " || msg.String() == "space"):
		p = p.step(1)
	case p.focus == focusDev && (msg.String() == " " || msg.String() == "space"):
		p.dev = !p.dev
	case (p.focus == focusInstall || p.focus == focusCancel) && (msg.String() == "left" || msg.String() == "right"):
		p.focus = focusInstall + focusCancel - p.focus
	}
	return p, dialog.Outcome{}
}

// move goes to the next (or, with dir -1, the previous) control, skipping the
// project when there is nothing to choose and the development box when it is
// not there.
func (p Prompt) move(dir int) Prompt {
	order := []int{}
	if p.choosable() {
		order = append(order, focusProject)
	}
	if p.offersLibraries() {
		order = append(order, focusSource)
	}
	switch {
	case p.linking():
		if _, ok := p.library(); ok {
			order = append(order, focusLibrary)
			if len(p.rangeChoices()) > 0 {
				order = append(order, focusRange)
			}
		}
	default:
		order = append(order, focusPackages)
		if !p.remove {
			order = append(order, focusDev)
		}
	}
	order = append(order, focusInstall, focusCancel)
	at := 0
	for i, f := range order {
		if f == p.focus {
			at = i
		}
	}
	p.focus = order[(at+dir+len(order))%len(order)]
	if p.focus == focusPackages {
		p.input.Focus()
	} else {
		p.input.Blur()
	}
	return p
}

// cycle moves the answer of the source, library or range row by dir.
func (p Prompt) cycle(id, dir int) Prompt {
	wrapAt := func(at, n int) int { return (at + dir + n) % n }
	switch id {
	case focusSource:
		p.fromWorkspace = !p.fromWorkspace
		p.libAt, p.rangeAt = 0, 0
	case focusLibrary:
		if n := len(p.valid()); n > 0 {
			p.libAt, p.rangeAt = wrapAt(p.libAt, n), 0
		}
	case focusRange:
		if n := len(p.rangeChoices()); n > 0 {
			p.rangeAt = wrapAt(p.rangeAt, n)
		}
	}
	return p
}

// submit goes ahead when the packages (or the library) are good, and says why
// not otherwise.
func (p Prompt) submit() (Prompt, dialog.Outcome) {
	if p.linking() {
		if _, ok := p.library(); !ok {
			p.err = words.Say("packageprompt.error.nolibrary")
			return p, dialog.Outcome{}
		}
		return p, dialog.Outcome{Done: true, Confirmed: true}
	}
	if _, err := ParsePackages(p.input.Value()); err != nil {
		p.err = err.Error()
		p.focus = focusPackages
		p.input.Focus()
		return p, dialog.Outcome{}
	}
	return p, dialog.Outcome{Done: true, Confirmed: true}
}
