package tui

import (
	"strings"
	"testing"
)

func TestTheMenuBarAndTheTabStripAreNeverCutAboveTheMinimumWidth(t *testing.T) {
	for w := minWidth; w <= 150; w += 3 {
		l := newLab(t, true, w, 24)
		lines := l.lines()
		if !strings.Contains(lines[0], "Help") {
			t.Fatalf("%d columns: the menu bar is cut: %q", w, lines[0])
		}
		if !strings.Contains(lines[1], "Guide") {
			t.Fatalf("%d columns: the tab strip is cut: %q", w, lines[1])
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
