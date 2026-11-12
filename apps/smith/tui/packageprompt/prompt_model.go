package packageprompt

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
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

// Place centres the dialog on a screen of sw by sh cells.
func (p Prompt) Place(sw, sh int) Prompt {
	w := min(max(sw-8, minWidth), min(max(sw-2, 10), 76))
	p.input.Width = max(w-8, 8)
	h := min(len(p.body(textW(w-4)))+4, max(p.margins.Room(sh), 3))
	p.rect = p.margins.Centre(sw, sh, w, h)
	return p
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
	if y-r.Y-1 == 0 && p.choosable() && x > r.X && x < r.X+r.W-1 {
		return focusProject
	}
	inside := x > r.X && x < r.X+r.W-1
	for _, c := range choiceRows() {
		if row := p.rowOf(c.prefix, textW(inner)); row >= 0 && y-r.Y-1 == row && inside {
			return c.id
		}
	}
	if !p.remove && !p.linking() && y-r.Y-1 == p.rowOf("[", textW(inner)) && inside {
		return focusDev
	}
	return -1
}

// Hover highlights the control under the pointer.
func (p Prompt) Hover(x, y int) Prompt { p.hover = p.controlAt(x, y); return p }

// Click handles a left press: the box flips, Install goes ahead, Cancel and a
// press outside the dialog cancel.
func (p Prompt) Click(x, y int) (Prompt, dialog.Outcome) {
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

// Key handles a key.
func (p Prompt) Key(msg tea.KeyMsg) (Prompt, dialog.Outcome) {
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
