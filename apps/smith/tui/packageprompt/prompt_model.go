package packageprompt

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Prompt asks which project and which packages to install on, or remove from,
// and ends as the command line that does it: `smith install -w <project> [-D]
// <packages>` or `smith uninstall -w <project> <packages>`. It shows that
// line as it will run, as the confirmation dialog does.
//
// Keys: typing fills the packages; Left and Right on the project choose
// another project; Tab and Shift+Tab move between the project, the packages, the development box and the buttons; Space or a click flips the
// box; Enter on the packages or on Install goes ahead; Esc cancels.
type Prompt struct {
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
	rect dialog.Rect
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
	in.Placeholder = "zod  react@19  @types/node"
	in.CharLimit = 200
	in.Focus()
	return Prompt{project: project, projects: append([]string(nil), projects...), remove: remove, input: in, hover: -1}
}

// Title is the heading of the dialog.
func (p Prompt) Title() string {
	if p.remove {
		return "Remove packages"
	}
	return "Install packages"
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
		argv := []string{"install", "-w", p.project, "workspace:" + lib.Name}
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
		return append([]string{"uninstall", "-w", p.project}, names...)
	}
	argv := []string{"install", "-w", p.project}
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

// Dev reports whether the packages go to the development dependencies.
func (p Prompt) Dev() bool { return p.dev }

const (
	minWidth      = 46
	buttonInstall = "[ Install  Enter ]"
	buttonRemove  = "[ Remove  Enter ]"
	buttonCancel  = "[ Cancel  Esc ]"
)

func (p Prompt) goLabel() string {
	if p.remove {
		return buttonRemove
	}
	return buttonInstall
}

// Place centres the dialog on a screen of sw by sh cells.
func (p Prompt) Place(sw, sh int) Prompt {
	w := min(max(sw-8, minWidth), min(max(sw-2, 10), 76))
	p.input.Width = max(w-8, 8)
	h := min(len(p.body(w-4))+4, max(sh, 3))
	p.rect = dialog.Rect{X: max((sw-w)/2, 0), Y: max((sh-h)/2, 0), W: w, H: h}
	return p
}

// Rect is where the dialog sits after Place.
func (p Prompt) Rect() dialog.Rect { return p.rect }

// body is the dialog's lines between the borders, before the buttons.
func (p Prompt) body(inner int) []string {
	lines := []string{p.projectLine(), ""}
	if p.offersLibraries() {
		lines = append(lines, p.sourceLine(), "")
	}
	if p.linking() {
		lines = append(lines, p.libraryLines(inner)...)
	} else {
		lines = append(lines, "Packages, separated by spaces:", " "+p.input.View(), "")
		if !p.remove {
			lines = append(lines, p.devLine(), "")
			lines = append(lines, "Names come from the language's public registry, not from this workspace.", "If one is wrong, smith keeps nothing: fix the name and try again.", "")
		}
	}
	if p.err != "" {
		lines = append(lines, p.err, "")
	}
	lines = append(lines, "The command that will run:")
	command := "(type a package first)"
	if p.linking() {
		command = "(no library to link)"
	}
	if argv := p.Argv(); argv != nil {
		command = commands.Line(argv)
	}
	for _, l := range wrap(command, inner-4) {
		lines = append(lines, "  $ "+l)
	}
	return lines
}

func (p Prompt) projectLine() string {
	if p.choosable() {
		return "Project: ‹ " + p.project + " ›  (Left and Right change it)"
	}
	return "Project: " + p.project
}

// sourceLine is the first question: what is being added.
func (p Prompt) sourceLine() string {
	answer := "A package from the registry (default)"
	if p.fromWorkspace {
		answer = "A library of this workspace"
	}
	return "Add: ‹ " + answer + " ›  (Left, Right)"
}

// libraryLines are the library and range rows with what each choice means.
func (p Prompt) libraryLines(inner int) []string {
	lib, ok := p.library()
	if !ok {
		return append(wrap("No library of this workspace can be linked to "+p.project+": a library must be a lib or a package, in the same language, and must not already depend on "+p.project+".", inner), "")
	}
	libs := p.valid()
	line := "Library: ‹ " + lib.Name + " · " + lib.Kind + " · " + lib.Language + " ›"
	if len(libs) > 1 {
		line += "  " + strconv.Itoa(min(p.libAt, len(libs)-1)+1) + "/" + strconv.Itoa(len(libs)) + " (Left, Right)"
	}
	lines := []string{line}
	for _, l := range wrapWords(LibraryMeaning(p.project, lib), inner-2) {
		lines = append(lines, "  "+l)
	}
	lines = append(lines, "")
	if choices := p.rangeChoices(); len(choices) > 0 {
		current := p.rangeChoice()
		lines = append(lines, "Range: ‹ "+current.Label+" ›  (Left, Right)")
		for _, l := range wrapWords(current.Meaning, inner-2) {
			lines = append(lines, "  "+l)
		}
		lines = append(lines, "")
	}
	return lines
}

func (p Prompt) devLine() string {
	box := "[ ]"
	if p.dev {
		box = "[x]"
	}
	return " " + box + " a development dependency (-D)"
}

// rowOf finds the line of the body that starts with prefix, or -1.
func (p Prompt) rowOf(prefix string, inner int) int {
	for i, l := range p.body(inner) {
		if strings.HasPrefix(strings.TrimLeft(l, " "), prefix) {
			return i
		}
	}
	return -1
}

// Lines draws the dialog: Rect().H lines of Rect().W cells.
func (p Prompt) Lines(th theme.Theme) []string {
	r := p.rect
	if r.W < 6 || r.H < 3 {
		return nil
	}
	inner := r.W - 4
	rows := make([]string, 0, r.H-2)
	row := func(content string) string { return " " + pane.Fit(content, inner) }
	for _, l := range p.body(inner) {
		if len(rows) >= r.H-4 {
			break
		}
		switch {
		case strings.HasPrefix(l, "Project:"):
			switch {
			case p.focus == focusProject:
				l = th.Selected(l)
			case p.hover == focusProject:
				l = th.Hover(l)
			}
		case strings.HasPrefix(l, "Add:"), strings.HasPrefix(l, "Library:"), strings.HasPrefix(l, "Range:"):
			id := map[string]int{"Add:": focusSource, "Library:": focusLibrary, "Range:": focusRange}[l[:strings.Index(l, ":")+1]]
			switch {
			case p.focus == id:
				l = th.Selected(l)
			case p.hover == id:
				l = th.Hover(l)
			}
		case p.err != "" && l == p.err:
			l = th.Failure(l)
		case strings.HasPrefix(strings.TrimLeft(l, " "), "[") && !p.remove && strings.Contains(l, "development"):
			switch {
			case p.focus == focusDev:
				l = th.Selected(l)
			case p.hover == focusDev:
				l = th.Hover(l)
			}
		case strings.HasPrefix(l, "  $ "):
			l = th.Bold(l)
		}
		rows = append(rows, row(l))
	}
	rows = append(rows, row(""))
	line := strings.Repeat(" ", p.buttonsRow(inner)[0][0]-2)
	for i, label := range []string{p.goLabel(), buttonCancel} {
		text := th.Muted(label)
		switch id := focusInstall + i; {
		case p.focus == id:
			text = th.Selected(label)
		case p.hover == id:
			text = th.Hover(label)
		}
		line += text
		if i == 0 {
			line += "  "
		}
	}
	rows = append(rows, row(line))
	return pane.Frame(th, p.Title(), rows, r.W, r.H, true)
}

// buttonsRow is the span (from the dialog's left edge) of each button.
func (p Prompt) buttonsRow(inner int) [2][2]int {
	run, stop := ansi.StringWidth(p.goLabel()), ansi.StringWidth(buttonCancel)
	start := 2 + max((inner-(run+2+stop))/2, 0)
	return [2][2]int{{start, start + run}, {start + run + 2, start + run + 2 + stop}}
}

// controlAt is the control at a cell: the development box, a button, or -1.
func (p Prompt) controlAt(x, y int) int {
	r := p.rect
	inner := r.W - 4
	if y == r.Y+r.H-2 {
		for i, s := range p.buttonsRow(inner) {
			if x-r.X >= s[0] && x-r.X < s[1] {
				return focusInstall + i
			}
		}
		return -1
	}
	if y-r.Y-1 == 0 && p.choosable() && x > r.X && x < r.X+r.W-1 {
		return focusProject
	}
	inside := x > r.X && x < r.X+r.W-1
	for _, c := range []struct {
		prefix string
		id     int
	}{{"Add:", focusSource}, {"Library:", focusLibrary}, {"Range:", focusRange}} {
		if row := p.rowOf(c.prefix, inner); row >= 0 && y-r.Y-1 == row && inside {
			return c.id
		}
	}
	if !p.remove && !p.linking() && y-r.Y-1 == p.rowOf("[", inner) && inside {
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
	switch msg.String() {
	case "esc", "ctrl+c":
		return p, dialog.Outcome{Done: true}
	case "tab", "down":
		return p.move(1), dialog.Outcome{}
	case "shift+tab", "up":
		return p.move(-1), dialog.Outcome{}
	case "enter":
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
			p.err = "There is no library to link: choose a package from the registry, or press Esc."
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

// wrap breaks s into lines of at most w cells, cutting words that are longer.
func wrap(s string, w int) []string {
	if w < 1 || ansi.StringWidth(s) <= w {
		return []string{s}
	}
	var out []string
	for ansi.StringWidth(s) > w {
		out = append(out, ansi.Truncate(s, w, ""))
		s = ansi.TruncateLeft(s, w, "")
	}
	return append(out, s)
}
