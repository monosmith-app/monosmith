package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/muesli/termenv"
)

func TestOnAWideScreenTheTitleBarIsTheFirstRowAndTheRailListsTheViews(t *testing.T) {
	l := newLab(t, true, 160, 30)
	first := pane.Strip(l.lines()[0])
	for _, want := range []string{"acme", "main", "Ctrl+K", "smith 1.2.3"} {
		if !strings.Contains(first, want) {
			t.Fatalf("the title bar lacks %q: %q", want, first)
		}
	}
	if strings.Contains(first, "Home") || strings.Contains(first, "Settings") {
		t.Fatalf("the views are in the rail, not on the title bar: %q", first)
	}
	lay := l.sh().lay()
	if lay.Main.Y != 2 || lay.Rail.W == 0 || lay.Tabs.W != 0 {
		t.Fatalf("a blank row, then the rail and the frames: %+v", lay)
	}
	rail := strings.Join(l.lines()[lay.Rail.Y:lay.Rail.Y+lay.Rail.H], "\n")
	for _, want := range []string{"╭─ Views", "Home", "Projects", "Health", "Dependencies", "Release", "Settings"} {
		if !strings.Contains(rail, want) {
			t.Fatalf("the rail lacks %q:\n%s", want, rail)
		}
	}
}

func TestTheTitleBarHoldsNoDoctorState(t *testing.T) {
	l := newLab(t, true, 160, 30)
	if first := pane.Strip(l.lines()[0]); strings.Contains(first, "×") {
		t.Fatalf("the doctor state is in the status bar, not the title bar: %q", first)
	}
	lines := l.lines()
	if bar := lines[len(lines)-1]; !strings.HasSuffix(bar, "main  ·  1.2.3  ·  ×2 ") {
		t.Fatalf("the status bar ends with the branch, the version and the doctor state: %q", bar)
	}
}

func TestTheHeaderTagIsTheAccentAndTheWorkspaceIsBold(t *testing.T) {
	l := newLab(t, true, 110, 30)
	sh := l.sh()
	sh.th = theme.New(theme.Choice{Tokens: theme.Dark, Icons: theme.Unicode})
	defer lipgloss.SetColorProfile(lipgloss.ColorProfile())
	lipgloss.SetColorProfile(termenv.TrueColor)
	row := sh.headerRow()
	for _, want := range []string{sh.th.ButtonActive(" smith "), sh.th.Bold("acme"), sh.th.Muted("main"), sh.th.Accent("Ctrl+K")} {
		if !strings.Contains(row, want) {
			t.Errorf("the title bar lacks %q: %q", want, row)
		}
	}
}

func TestOnANarrowScreenTheStripOfViewsIsTheSecondRow(t *testing.T) {
	l := newLab(t, true, 80, 30)
	lines := l.lines()
	if !strings.Contains(lines[0], "acme") || strings.Contains(lines[0], "Home") {
		t.Fatalf("the title bar is alone on the first row: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "1 Home  2 Projects") || !strings.Contains(lines[1], "4 Deps") || strings.Contains(lines[1], "Dependencies") {
		t.Fatalf("and the strip of views on the second, with Deps short: %q", lines[1])
	}
	if pane.Width(lines[1]) != 80 || strings.Contains(lines[1], "›") {
		t.Fatalf("the strip fits 80 columns without scrolling: %q", lines[1])
	}
}

func TestTheTitleBarHasThePaletteHintAndTheVersionAtTheRightAtEveryWidth(t *testing.T) {
	l := newLab(t, true, 110, 30)
	row := pane.Strip(l.sh().headerRow())
	if !strings.HasPrefix(row, "  smith   acme  main ") {
		t.Fatalf("the tag, the workspace and the branch belong at the left: %q", row)
	}
	if !strings.HasSuffix(row, "⌕ Type a command, project or task  Ctrl+K   smith 1.2.3 ") || pane.Width(row) != 110 {
		t.Fatalf("the palette hint and the version belong at the right end: %q", row)
	}
	for _, w := range []int{160, 80, 60} {
		row := pane.Strip(newLab(t, true, w, 24).sh().headerRow())
		if !strings.Contains(row, "Ctrl+K") || !strings.HasSuffix(row, "smith 1.2.3 ") {
			t.Fatalf("at %d columns the title bar keeps the hint and the version: %q", w, row)
		}
	}
	narrow := pane.Strip(newLab(t, true, 60, 24).sh().headerRow())
	if !strings.HasSuffix(narrow, "Ctrl+K commands   smith 1.2.3 ") || strings.Contains(narrow, "Type a command") {
		t.Fatalf("a narrow title bar has the short hint: %q", narrow)
	}
}

func TestAHeaderTooLongForItsRowEndsWithAnEllipsis(t *testing.T) {
	l := newLab(t, false, 50, 24)
	row := strings.TrimRight(pane.Strip(l.sh().headerRow()), " ")
	if !strings.HasSuffix(row, "…") {
		t.Fatalf("the cut header must end with an ellipsis: %q", row)
	}
}

func TestTheWorkspaceButtonsOfferToClearTheCache(t *testing.T) {
	l := newLab(t, true, 160, 30)
	var ids []string
	for _, row := range l.sh().workspaceButtons(200, "") {
		for _, sp := range row.hit.spans {
			ids = append(ids, sp.id)
		}
	}
	if len(ids) == 0 || ids[len(ids)-1] != "ws.cache-clear" {
		t.Fatalf("the last workspace button deletes the local cache: %v", ids)
	}
}
