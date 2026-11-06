package packageprompt

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/choicerow"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
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
			lines = append(lines, words.Say("packageprompt.registry.names"), words.Say("packageprompt.registry.wrong"), "")
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
		case strings.HasPrefix(l, words.Say("packageprompt.label.project")):
			switch {
			case p.focus == focusProject:
				l = th.Selected(l)
			case p.hover == focusProject:
				l = th.Hover(l)
			}
		case rowFocus(l) >= 0:
			id := rowFocus(l)
			switch {
			case p.focus == id:
				l = th.Selected(l)
			case p.hover == id:
				l = th.Hover(l)
			}
		case p.err != "" && l == p.err:
			l = th.Failure(l)
		case !p.remove && l == p.devLine():
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
	rows = append(rows, row(line))
	return pane.Frame(th, p.Title(), rows, r.W, r.H, true)
}
