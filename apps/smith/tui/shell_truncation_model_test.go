package tui

import (
	"strings"
	"testing"
)

func TestTheTabStripAndTheFootersHelpAndQuitAreNeverCutAboveTheMinimumWidth(t *testing.T) {
	for w := minWidth; w <= 150; w += 3 {
		l := newLab(t, true, w, 24)
		lines := l.lines()
		// The strip is shortened to fit and, when it still does not, scrolls with a marker.
		if strip := lines[2]; !strings.Contains(strip, "Release") && !strings.Contains(strip, "›") {
			t.Fatalf("%d columns: the tab strip is cut without a marker: %q", w, strip)
		}
		if foot := lines[len(lines)-2]; !strings.Contains(foot, "? help") || !strings.Contains(foot, "q quit") {
			t.Fatalf("%d columns: the footer lost help or quit: %q", w, foot)
		}
	}
}

func TestBelowTheMinimumTheScreenIsOneHintLine(t *testing.T) {
	l := newLab(t, true, minWidth-1, 24)
	lines := l.lines()
	if !strings.Contains(lines[0], "too small") {
		t.Fatalf("no hint: %q", lines[0])
	}
	for _, rest := range lines[1:] {
		if strings.TrimSpace(rest) != "" {
			t.Fatalf("garbage under the hint: %q", rest)
		}
	}
	l = newLab(t, true, 120, minHeight-1)
	if !strings.Contains(l.lines()[0], "too small") {
		t.Fatalf("a short terminal gets no hint: %q", l.lines()[0])
	}
}
