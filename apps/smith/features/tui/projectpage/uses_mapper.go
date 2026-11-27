package projectpage

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// boxesFrom is the width from which Uses and Used by are two boxes side by side.
const boxesFrom = 60

// Uses draws what a project uses and what uses it. From 60 cells they are two
// rounded boxes side by side, titled "Uses" and "Used by"; narrower, two rows
// with the label in muted text. A project that others use says that a change
// to it tests them too.
//
// Parameters: th is the theme; uses are the projects it imports; usedBy are
// the projects that import it; w is the width.
// Returns: the lines.
func Uses(th theme.Theme, uses, usedBy []string, w int) []string {
	usesText, usedText := names(th, uses), names(th, usedBy)
	note := ""
	if len(usedBy) > 0 {
		note = words.Say("projectpage.deps.tests_them")
	}
	if w < boxesFrom {
		labelW := max(pane.Width(words.Say("projectpage.deps.uses")), pane.Width(words.Say("projectpage.deps.used_by"))) + 3
		lines := []string{pane.Fit(" "+th.Muted(pane.Cell(words.Say("projectpage.deps.uses"), labelW))+usesText, w)}
		used := " " + th.Muted(pane.Cell(words.Say("projectpage.deps.used_by"), labelW)) + usedText
		if note != "" {
			used += "  " + th.Muted("("+note+")")
		}
		return append(lines, pane.Fit(pane.CutEllipsis(used, w), w))
	}
	boxW := (w - 3) / 2
	inner := boxW - 4
	left := pane.Frame(th, words.Say("projectpage.deps.uses"), box(usesText, "", inner), boxW, 5, false)
	right := pane.Frame(th, words.Say("projectpage.deps.used_by"), box(usedText, th.Muted(note), inner), boxW, 5, false)
	out := make([]string, len(left))
	for i := range left {
		out[i] = pane.Fit(" "+left[i]+" "+right[i], w)
	}
	return out
}

// box is the body of one box: a blank line, the names, the note, a blank line.
func box(text, note string, w int) []string {
	return []string{"", " " + pane.CutEllipsis(text, w) + " ", " " + pane.CutEllipsis(note, w) + " "}
}

// names are project names joined with a dot, or a muted "none".
func names(th theme.Theme, list []string) string {
	if len(list) == 0 {
		return th.Muted(words.Say("projectpage.deps.none"))
	}
	return strings.Join(list, " · ")
}
