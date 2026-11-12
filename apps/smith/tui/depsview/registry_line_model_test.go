package depsview

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func withPages() Model {
	return FromResult(Result{Rows: []Row{
		{Name: "zod", Current: "3.22.0", Latest: "3.23.1", Level: "minor", URL: "https://www.npmjs.com/package/zod"},
		{Name: "acme-internal", Current: "1.0.0", Latest: "1.1.0", Level: "minor"},
	}})
}

func TestTheSelectedRowShowsItsRegistryPageAndTheKeys(t *testing.T) {
	out := strings.Join(render(withPages(), 80, 24), "\n")
	for _, want := range []string{"Registry: https://www.npmjs.com/package/zod", "g opens it in your browser", "Y copies the link"} {
		if !strings.Contains(out, want) {
			t.Errorf("the screen lacks %q:\n%s", want, out)
		}
	}
}

func TestAPackageWithoutAPageSaysSoInsteadOfLinkingToAWrongOne(t *testing.T) {
	m := withPages()
	m.Cursor = 1
	out := strings.Join(render(m, 80, 24), "\n")
	if !strings.Contains(out, "no page known for acme-internal") || strings.Contains(out, "https://") {
		t.Errorf("screen:\n%s", out)
	}
}

func TestTheLinkIsAnOsc8HyperlinkOnlyWhereTheTerminalDrawsOne(t *testing.T) {
	m := withPages()
	plain, _ := m.View(theme.Default(), 80, 24)
	if strings.Contains(strings.Join(plain, ""), "\x1b]8;;") {
		t.Error("no hyperlink without terminal support")
	}
	m.Links = true
	lines, _ := m.View(theme.Default(), 80, 24)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "\x1b]8;;https://www.npmjs.com/package/zod\x1b\\") {
		t.Errorf("the link is an OSC 8 hyperlink:\n%q", joined)
	}
	for _, l := range lines {
		if pane.Width(l) > 80 {
			t.Errorf("line wider than the pane: %q", pane.Strip(l))
		}
	}
	if !strings.Contains(pane.Strip(joined), "Registry: https://www.npmjs.com/package/zod") {
		t.Error("the address is printed in full, not only linked")
	}
}

func TestALongAddressKeepsItsStartAndItsEndOnANarrowScreen(t *testing.T) {
	m := FromResult(Result{Rows: []Row{{Name: "x", Current: "1.0.0", Latest: "2.0.0", Level: "major", URL: "https://dev.azure.com/acme/web/_artifacts/feed/main/Npm/@acme%2Fui/overview"}}})
	out := strings.Join(render(m, 60, 24), "\n")
	if !strings.Contains(out, "https://dev.azure.com") || !strings.Contains(out, "overview") || !strings.Contains(out, "…") {
		t.Errorf("screen:\n%s", out)
	}
}

func TestTheSelectedDependencyHasTheCursorMarkAndTheBoxesStayInOneColumn(t *testing.T) {
	m := withPages()
	m.Cursor = 1
	box, rows := -1, 0
	for _, l := range render(m, 80, 24) {
		i := strings.Index(l, "[ ]")
		if i < 0 || !strings.Contains(l, " minor ") {
			continue
		}
		rows++
		i = pane.Width(l[:i])
		if box >= 0 && i != box {
			t.Fatalf("the box moved to column %d, want %d: %q", i, box, l)
		}
		box = i
		if strings.HasPrefix(l, "› ") != strings.Contains(l, "acme-internal") {
			t.Fatalf("only the selected row carries the cursor mark: %q", l)
		}
	}
	if rows != 2 {
		t.Fatalf("%d dependency rows, want 2", rows)
	}
}
