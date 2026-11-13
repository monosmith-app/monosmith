package packageprompt

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/choicerow"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// body is the dialog's lines between the borders, before the buttons.
func (p Prompt) body(inner int) []string {
	lines := []string{p.projectLine(), ""}
	if p.offersLibraries() {
		lines = append(lines, p.sourceLine(), "")
	}
	if p.linking() {
		lines = append(lines, p.libraryLines(inner)...)
	} else {
		lines = append(lines, words.Say("packageprompt.packages.label"), " "+p.input.View(), "")
		if !p.remove {
			lines = append(lines, p.devLine(), "")
			// The notes wrap rather than run past the border.
			lines = append(lines, pane.WrapWords(words.Say("packageprompt.registry.names"), inner)...)
			lines = append(lines, pane.WrapWords(words.Say("packageprompt.registry.wrong"), inner)...)
			lines = append(lines, "")
		}
	}
	if p.err != "" {
		lines = append(lines, p.err, "")
	}
	lines = append(lines, words.Say("packageprompt.command.label"))
	command := words.Say("packageprompt.command.empty")
	if p.linking() {
		command = words.Say("packageprompt.command.nolibrary")
	}
	if argv := p.Argv(); argv != nil {
		command = commands.Line(argv)
	}
	for _, l := range pane.WrapHard(command, inner-4) {
		lines = append(lines, "  $ "+l)
	}
	return lines
}

func (p Prompt) projectLine() string {
	if p.choosable() {
		return words.Say("packageprompt.project.choosable", "project", p.project)
	}
	return words.Say("packageprompt.project.fixed", "project", p.project)
}

// sourceLine is the first question: what is being added.
func (p Prompt) sourceLine() string {
	answer := words.Say("packageprompt.source.registry")
	if p.fromWorkspace {
		answer = words.Say("packageprompt.source.workspace")
	}
	return words.Say("packageprompt.source.line", "answer", answer)
}

// libraryLines are the library and range rows with what each choice means.
func (p Prompt) libraryLines(inner int) []string {
	lib, ok := p.library()
	if !ok {
		return append(pane.WrapHard(words.Say("packageprompt.library.none", "project", p.project), inner), "")
	}
	libs := p.valid()
	line := words.Say("packageprompt.library.line", "name", lib.Name, "kind", lib.Kind, "language", lib.Language)
	if len(libs) > 1 {
		line += words.Say("packageprompt.library.position", "at", min(p.libAt, len(libs)-1)+1, "total", len(libs))
	}
	lines := []string{line}
	for _, l := range pane.WrapWords(LibraryMeaning(p.project, lib), inner-2) {
		lines = append(lines, "  "+l)
	}
	lines = append(lines, "")
	if choices := p.rangeChoices(); len(choices) > 0 {
		current := p.rangeChoice()
		lines = append(lines, words.Say("packageprompt.range.line", "choice", current.Label))
		for _, l := range pane.WrapWords(current.Meaning, inner-2) {
			lines = append(lines, "  "+l)
		}
		lines = append(lines, "")
	}
	return lines
}

func (p Prompt) devLine() string {
	return words.Say("packageprompt.dev.line", "box", choicerow.Box(p.dev, false))
}

// textW is the width of the body's text in a dialog of inner cells: the
// column of the cursor mark comes first.
func textW(inner int) int {
	// One cell less: the last cell is the scrollbar's when the body scrolls.
	return inner - pane.Width(choicerow.Cursor(theme.Default(), false)) - 1
}

// Lines draws the dialog: Rect().H lines of Rect().W cells.
func (p Prompt) Lines(th theme.Theme) []string {
	r := p.rect
	if r.W < 6 || r.H < 3 {
		return nil
	}
	inner := r.W - 4
	rows := make([]string, 0, r.H-2)
	row := func(content string) string { return " " + pane.Fit(content, pane.DialogRowWidth(r.W)) }
	// A row that takes the focus has the look of every pick list: the cursor
	// mark in the column in front of the text, and an accent label.
	focusable := func(l string, id int) string {
		switch {
		case p.focus == id:
			return th.Accent(choicerow.Cursor(th, true) + l)
		case p.hover == id:
			return th.RowHover(pane.Fit(choicerow.Cursor(th, false)+l, inner))
		}
		return choicerow.Cursor(th, false) + l
	}
	var styled []string
	for _, l := range p.body(textW(inner)) {
		switch {
		case strings.HasPrefix(l, words.Say("packageprompt.label.project")):
			l = focusable(l, focusProject)
		case rowFocus(l) >= 0:
			l = focusable(l, rowFocus(l))
		case p.err != "" && l == p.err:
			l = choicerow.Cursor(th, false) + th.Failure(l)
		case !p.remove && l == p.devLine():
			l = focusable(l, focusDev)
		case strings.HasPrefix(l, "  $ "):
			l = choicerow.Cursor(th, false) + th.Bold(l)
		default:
			l = choicerow.Cursor(th, false) + l
		}
		styled = append(styled, l)
	}
	labels, keys := p.buttonLabels()
	line := strings.Repeat(" ", button.RowSpans(th, labels, keys, inner)[0][0]-2)
	for i, label := range labels {
		state := button.Rest
		switch id := focusInstall + i; {
		case p.hover == id:
			state = button.Hover
		case p.focus == id:
			state = button.Active
		}
		line += button.Render(th, label, keys[i], state)
		if i == 0 {
			line += " "
		}
	}
	// A body taller than the dialog scrolls above the buttons, which stay.
	for _, l := range pane.DialogBody(th, styled, []string{"", line}, pane.DialogRowWidth(r.W), r.H-2, p.top) {
		rows = append(rows, row(l))
	}
	return pane.Frame(th, p.Title(), rows, r.W, r.H, true)
}
