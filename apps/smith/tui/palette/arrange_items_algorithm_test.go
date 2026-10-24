package palette

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

const outside = "not in a workspace (create one with File > New workspace)"

func TestUnavailableCommandsAreHiddenUntilSearchedFor(t *testing.T) {
	blocked := []Entry{
		{ID: "ws.doctor", Title: "Doctor", Hint: "check", Unavailable: outside},
		{ID: "run.build", Title: "Build", Hint: "smith build", Shortcut: "Ctrl+B", Unavailable: outside},
		{ID: "file.new", Title: "New workspace…", Hint: "smith new", Shortcut: "Ctrl+N"},
	}
	p := New(blocked, nil)
	text := strings.Join(p.Lines(theme.Default(), 80, 12), "\n")
	if strings.Contains(text, "Doctor") || strings.Contains(text, "Build") || !strings.Contains(text, "New workspace…") {
		t.Fatalf("only the usable command should list:\n%s", text)
	}
	if !strings.Contains(text, "2 commands unusable here are hidden") {
		t.Fatalf("the footer should say what is hidden:\n%s", text)
	}
	found := strings.Join(typed(p, "doctor").Lines(theme.Default(), 80, 12), "\n")
	if !strings.Contains(found, "Doctor") || !strings.Contains(found, "unavailable") {
		t.Fatalf("searching should find it with the reason:\n%s", found)
	}
}

func TestAReasonNeverRunsIntoItsLabelOrTheMarker(t *testing.T) {
	blocked := []Entry{{ID: "ws.ci", Title: "Regenerate CI pipelines…", Unavailable: outside}}
	for _, w := range []int{40, 60, 90} {
		for _, line := range typed(New(blocked, nil), "regen").Lines(theme.Default(), w, 6) {
			plain := pane.Strip(line)
			if strings.Contains(plain, "unavailable") && !strings.Contains(plain, " unavailable") {
				t.Fatalf("w=%d needs a gap before the marker: %q", w, plain)
			}
			if pane.Width(line) > w {
				t.Fatalf("w=%d line too wide: %q", w, plain)
			}
		}
	}
}
