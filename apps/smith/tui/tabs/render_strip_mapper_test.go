package tabs

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func sample() []Tab {
	return []Tab{
		{ID: "projects", Label: "Projects", Short: "Proj"},
		{ID: "run", Label: "Run", Reserve: 3},
		{ID: "doctor", Label: "Doctor", Reserve: 3},
		{ID: "release", Label: "Release", Short: "Rel"},
		{ID: "settings", Label: "Settings", Short: "Sett"},
	}
}

func plain(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = pane.Strip(l)
	}
	return out
}

func props(width int, tall bool, active string) Props {
	return Props{Theme: theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode}), Tabs: sample(), Active: active, Width: width, Tall: tall,
		ShortenOrder: []string{"settings", "release", "projects"}}
}

func TestTheTallStripOpensTheActiveTabIntoThePane(t *testing.T) {
	s := Render(props(60, true, "run"))
	lines := plain(s.Lines)
	if len(lines) != 3 {
		t.Fatalf("a tall strip has three rows, got %d", len(lines))
	}
	for i, l := range lines {
		if pane.Width(l) != 60 {
			t.Fatalf("row %d is %d cells wide: %q", i, pane.Width(l), l)
		}
	}
	if !strings.HasPrefix(lines[0], "╭──────────╮╭─────╮") || !strings.Contains(lines[1], "│ Projects ││ Run │") {
		t.Fatalf("the boxes are wrong:\n%s", strings.Join(lines, "\n"))
	}
	// Under Projects the pane's edge is closed; under Run it is open.
	if !strings.HasPrefix(lines[2], "├──────────┴┘     └") {
		t.Fatalf("the active tab should be open into the pane:\n%s", strings.Join(lines, "\n"))
	}
	if !strings.HasSuffix(lines[2], "╮") {
		t.Fatalf("the edge ends in the pane's corner: %q", lines[2])
	}
}

func TestTheFirstTabOpensIntoTheLeftEdge(t *testing.T) {
	lines := plain(Render(props(60, true, "projects")).Lines)
	if !strings.HasPrefix(lines[2], "│          └") {
		t.Fatalf("the first tab, active, joins the pane's left edge:\n%s", strings.Join(lines, "\n"))
	}
}

func TestTheFlatStripIsOneRowOfNames(t *testing.T) {
	s := Render(props(80, false, "run"))
	if len(s.Lines) != 1 {
		t.Fatalf("rows: %d", len(s.Lines))
	}
	line := plain(s.Lines)[0]
	if pane.Width(line) != 80 || !strings.Contains(line, " Projects   Run   Doctor   Release   Settings ") {
		t.Fatalf("%q", line)
	}
}

func TestEverySpanFindsItsTab(t *testing.T) {
	for _, tall := range []bool{true, false} {
		s := Render(props(80, tall, "doctor"))
		for _, sp := range s.Spans {
			if got := s.At(sp.X0); got != sp.ID {
				t.Errorf("tall=%v: column %d is %q, want %q", tall, sp.X0, got, sp.ID)
			}
		}
		if s.At(79) != "" {
			t.Errorf("tall=%v: the far end belongs to no tab", tall)
		}
	}
}

func TestNamesAreShortenedBeforeTheStripScrolls(t *testing.T) {
	// Wide enough for the short names, not for the full ones.
	s := Render(props(52, true, "run"))
	text := strings.Join(plain(s.Lines), "\n")
	if strings.Contains(text, "‹") || strings.Contains(text, "›") {
		t.Fatalf("shortening is enough, no scrolling:\n%s", text)
	}
	if !strings.Contains(text, "Sett") || strings.Contains(text, "Settings") {
		t.Fatalf("Settings should be shortened first:\n%s", text)
	}
	if !strings.Contains(text, "Doctor") {
		t.Fatalf("Doctor is not shortenable:\n%s", text)
	}
}

func TestTheActiveTabIsAlwaysInViewWhenTheStripScrolls(t *testing.T) {
	for _, tall := range []bool{true, false} {
		for w := 24; w <= 90; w++ {
			for _, a := range sample() {
				s := Render(props(w, tall, a.ID))
				found := false
				for _, sp := range s.Spans {
					found = found || sp.ID == a.ID
				}
				for _, l := range s.Lines {
					if pane.Width(l) != w {
						t.Fatalf("tall=%v width %d: a row is %d cells", tall, w, pane.Width(l))
					}
				}
				if !found {
					t.Fatalf("tall=%v width %d: %s is not in view:\n%s", tall, w, a.ID, strings.Join(plain(s.Lines), "\n"))
				}
			}
		}
	}
}

func TestMarkersShowWhatIsHiddenAndAreClickable(t *testing.T) {
	s := Render(props(28, true, "doctor"))
	text := strings.Join(plain(s.Lines), "\n")
	if !strings.Contains(text, "‹") || !strings.Contains(text, "›") {
		t.Fatalf("tabs hidden on both sides need both markers:\n%s", text)
	}
	if s.At(0) != Prev {
		t.Fatalf("the left marker is clickable: %q", s.At(0))
	}
}

func TestABadgeSitsInsideTheTabAndNeverChangesTheNamesSpelledOut(t *testing.T) {
	without := props(52, true, "run")
	with := without
	with.Tabs = sample()
	with.Tabs[2].Badge = "✗2"
	a := strings.Join(plain(Render(without).Lines), "\n")
	b := strings.Join(plain(Render(with).Lines), "\n")
	if !strings.Contains(b, "Doctor ✗2") {
		t.Fatalf("the badge is in the tab:\n%s", b)
	}
	for _, name := range []string{"Sett", "Rel", "Proj"} {
		if strings.Contains(a, name) != strings.Contains(b, name) {
			t.Fatalf("%s was shortened differently with a badge", name)
		}
	}
}

func TestTheTitleSitsAtTheRightOfTheMiddleRowWhenItFits(t *testing.T) {
	p := props(80, true, "projects")
	p.Title = "Projects (4)"
	mid := plain(Render(p).Lines)[1]
	if !strings.Contains(mid, "Projects (4)") || pane.Width(mid) != 80 {
		t.Fatalf("%q", mid)
	}
	p.Width = 40
	if mid := plain(Render(p).Lines)[1]; strings.Contains(mid, "(4)") {
		t.Fatalf("no room for the title: %q", mid)
	}
}

func TestTinyWidthsDoNotPanic(t *testing.T) {
	for w := 0; w < 12; w++ {
		for _, tall := range []bool{true, false} {
			Render(props(w, tall, "run"))
		}
	}
}
