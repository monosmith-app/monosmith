package projectpage

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Button is a button at the right of the meta line.
type Button struct {
	ID, Label string
	// Key is the key that does the same as a click; "" for none.
	Key string
	// Primary draws the button in the accent: the one action of the detail.
	Primary bool
}

// HeadInput is what the top of the detail is drawn from.
type HeadInput struct {
	Path, Kind, Language string
	// Stack is the stack the project is built with, with what it gives; "" when none.
	Stack string
	// Version is "current → next" while a release is pending; "" when none.
	Version string
	Changed bool
	Buttons []Button
	// Hover is the ID under the pointer and Pressed the one it holds down.
	Hover, Pressed string
	// Section is the open tab; Narrow names the tabs as a narrow screen does.
	Section Section
	Narrow  bool
}

// Span is a part [From, To) of a line that answers a click with ID.
type Span struct {
	From, To int
	ID       string
}

// Head is the top of the detail, which never scrolls: the muted line
// "path · kind · language · stack · version" with the changed mark in the accent and the
// buttons at the right (on their own line when the width is short), a blank
// line, the tabs of the sections with the open one in the accent, bold and
// underlined, and a rule.
//
// Parameters: th is the theme; in is what the top shows; w is the width.
// Returns: the lines, and the clickable parts of each line by line.
func Head(th theme.Theme, in HeadInput, w int) (lines []string, spans map[int][]Span) {
	spans = map[int][]Span{}
	meta := strings.Join(nonEmpty(in.Path, in.Kind, in.Language, in.Stack, in.Version), " · ")
	left := " " + th.Muted(meta)
	if in.Changed {
		left += "  " + th.Accent(words.Say("projectpage.changed", "dot", th.Icons.Dot))
	}
	buttons, bw := renderButtons(th, in)
	switch {
	case len(in.Buttons) == 0:
		lines = append(lines, pane.Fit(pane.CutEllipsis(left, w), w))
	case pane.Width(left)+2+bw <= w-1:
		from := w - 1 - bw
		lines = append(lines, pane.Fit(left+strings.Repeat(" ", from-pane.Width(left))+buttons, w))
		spans[0] = buttonSpans(th, in.Buttons, from)
	default:
		lines = append(lines, pane.Fit(pane.CutEllipsis(left, w), w))
		spans[1] = buttonSpans(th, in.Buttons, 1)
		lines = append(lines, pane.Fit(" "+buttons, w))
	}
	lines = append(lines, "")
	tabs, tabSpans := sectionTabs(th, in, w)
	spans[len(lines)] = tabSpans
	lines = append(lines, tabs, " "+th.Border(strings.Repeat("─", max(w-2, 1)), false))
	return lines, spans
}

// nonEmpty are the parts that hold text.
func nonEmpty(parts ...string) []string {
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// renderButtons draws the buttons with a gap of two cells, and their width.
func renderButtons(th theme.Theme, in HeadInput) (string, int) {
	var b strings.Builder
	for i, bt := range in.Buttons {
		if i > 0 {
			b.WriteString("  ")
		}
		state := button.Rest
		switch {
		case bt.ID == in.Hover && bt.ID == in.Pressed:
			state = button.Active
		case bt.Primary:
			state = button.Active
		case bt.ID == in.Hover:
			state = button.Hover
		}
		b.WriteString(button.Render(th, bt.Label, bt.Key, state))
	}
	return b.String(), pane.Width(b.String())
}

// buttonSpans are the clickable parts of the buttons drawn from column from.
func buttonSpans(th theme.Theme, buttons []Button, from int) []Span {
	var out []Span
	for _, bt := range buttons {
		n := button.Width(th, bt.Label, bt.Key)
		out = append(out, Span{From: from, To: from + n, ID: bt.ID})
		from += n + 2
	}
	return out
}

// sectionTabs is the line of tabs and the clickable part of each.
func sectionTabs(th theme.Theme, in HeadInput, w int) (string, []Span) {
	gap := 4
	if in.Narrow {
		gap = 3
	}
	var b strings.Builder
	var spans []Span
	at := 2
	b.WriteString("  ")
	for i, s := range Sections {
		if i > 0 {
			b.WriteString(strings.Repeat(" ", gap))
			at += gap
		}
		title := s.Title(in.Narrow)
		switch {
		case s == in.Section:
			b.WriteString(th.ActiveTab(title))
		case s.ID() == in.Hover:
			b.WriteString(th.RowHover(title))
		default:
			b.WriteString(th.Muted(title))
		}
		spans = append(spans, Span{From: at, To: at + pane.Width(title), ID: s.ID()})
		at += pane.Width(title)
	}
	return pane.Fit(b.String(), w), spans
}
