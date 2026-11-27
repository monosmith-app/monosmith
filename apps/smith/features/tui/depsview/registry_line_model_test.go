package depsview

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
)

func withPages() Model {
	return FromResult(Result{Rows: []Row{
		{Name: "zod", Current: "3.22.0", Latest: "3.23.1", Level: "minor", URL: "https://www.npmjs.com/package/zod"},
		{Name: "acme-internal", Current: "1.0.0", Latest: "1.1.0", Level: "minor"},
	}})
}

func TestTheSelectedRowShowsItsRegistryPageAndTheKeys(t *testing.T) {
	out := detailText(withPages(), 76)
	for _, want := range []string{"Registry\nhttps://www.npmjs.com/package/zod", "g opens it in your browser", "Y copies the link"} {
		if !strings.Contains(out, want) {
			t.Errorf("the detail lacks %q:\n%s", want, out)
		}
	}
}

func TestAPackageWithoutAPageSaysSoInsteadOfLinkingToAWrongOne(t *testing.T) {
	m := withPages()
	m.Cursor = 1
	out := detailText(m, 76)
	if !strings.Contains(out, "no page known for acme-internal") || strings.Contains(out, "https://") {
		t.Errorf("detail:\n%s", out)
	}
}

func TestTheLinkIsAnOsc8HyperlinkOnlyWhereTheTerminalDrawsOne(t *testing.T) {
	m := withPages()
	join := func(m Model) string {
		var out []string
		for _, s := range m.Detail(theme.Default(), 76, nil) {
			if s.Heading == "Registry" {
				out = append(out, s.Items...)
			}
		}
		return strings.Join(out, "\n")
	}
	if strings.Contains(join(m), "\x1b]8;;") {
		t.Error("no hyperlink without terminal support")
	}
	m.Links = true
	joined := join(m)
	if !strings.Contains(joined, "\x1b]8;;https://www.npmjs.com/package/zod\x1b\\") {
		t.Errorf("the link is an OSC 8 hyperlink:\n%q", joined)
	}
	for _, l := range strings.Split(joined, "\n") {
		if pane.Width(l) > 76 {
			t.Errorf("line wider than the detail: %q", pane.Strip(l))
		}
	}
	if !strings.Contains(pane.Strip(joined), "https://www.npmjs.com/package/zod") {
		t.Error("the address is printed in full, not only linked")
	}
}

func TestALongAddressKeepsItsStartAndItsEndOnANarrowScreen(t *testing.T) {
	m := FromResult(Result{Rows: []Row{{Name: "x", Current: "1.0.0", Latest: "2.0.0", Level: "major", URL: "https://dev.azure.com/acme/web/_artifacts/feed/main/Npm/@acme%2Fui/overview"}}})
	out := detailText(m, 56)
	if !strings.Contains(out, "https://dev.azure.com") || !strings.Contains(out, "overview") || !strings.Contains(out, "…") {
		t.Errorf("detail:\n%s", out)
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
