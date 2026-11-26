package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/muesli/termenv"
)

func TestOnAWideScreenTheWorkspaceAndBranchSitAtTheRightEndOfTheTabRow(t *testing.T) {
	l := newLab(t, true, 160, 30)
	first := l.lines()[0]
	for _, want := range []string{"Home (1)", "Settings (6)", "acme", "main"} {
		if !strings.Contains(first, want) {
			t.Fatalf("the first row lacks %q: %q", want, first)
		}
	}
	if strings.Index(first, "acme") < strings.Index(first, "Settings (6)") {
		t.Fatalf("the workspace must come after the tabs, at the right: %q", first)
	}
	if y := l.sh().lay().Main.Y; y != 1 {
		t.Fatalf("the content starts under the shared row, not %d", y)
	}
}

func TestOnTheTabRowAShortHeaderEndsAtTheRightEdge(t *testing.T) {
	l := newLab(t, true, 160, 30)
	first := pane.Strip(l.lines()[0])
	if !strings.HasSuffix(first, " smith   acme  main ") {
		t.Fatalf("the header (tag, workspace, branch) must end one column before the right edge, with no padding: %q", first)
	}
	if _, header, _ := strings.Cut(first, "Settings (6)"); strings.Contains(header, "×") {
		t.Fatalf("the doctor state is in the status bar, not the header: %q", first)
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
	for _, want := range []string{sh.th.ButtonActive(" smith "), sh.th.Bold("acme"), sh.th.Muted("main"), sh.th.Accent("Ctrl+P")} {
		if !strings.Contains(row, want) {
			t.Errorf("the title bar lacks %q: %q", want, row)
		}
	}
}

func TestBelowTheInlineWidthTheHeaderKeepsItsOwnRow(t *testing.T) {
	l := newLab(t, true, 110, 30)
	lines := l.lines()
	if !strings.Contains(lines[0], "acme") || strings.Contains(lines[0], "Home") {
		t.Fatalf("a narrow screen keeps the header alone on the first row: %q", lines[0])
	}
	if !strings.Contains(lines[1], "Home (1)") {
		t.Fatalf("and the tabs on the second: %q", lines[1])
	}
}

func TestTheHeaderOnItsOwnRowIsATitleBarWithThePaletteHintAndTheVersionAtTheRight(t *testing.T) {
	l := newLab(t, true, 110, 30)
	row := pane.Strip(l.sh().headerRow())
	if !strings.HasPrefix(row, "  smith   acme  main ") {
		t.Fatalf("the tag, the workspace and the branch belong at the left: %q", row)
	}
	if !strings.HasSuffix(row, "⌕ Type a command, project or task  Ctrl+P   smith 1.2.3 ") || pane.Width(row) != 110 {
		t.Fatalf("the palette hint and the version belong at the right end: %q", row)
	}
	narrow := pane.Strip(newLab(t, true, 60, 24).sh().headerRow())
	if !strings.HasSuffix(narrow, "Ctrl+P commands   smith 1.2.3 ") || strings.Contains(narrow, "Type a command") {
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
