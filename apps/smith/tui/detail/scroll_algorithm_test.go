package detail

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func numbered(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("line %d", i)
	}
	return out
}

func TestAPanelThatFitsIsNotWindowed(t *testing.T) {
	out, _ := Window(theme.Default(), numbered(5), nil, 3, 10)
	if len(out) != 5 || strings.Contains(strings.Join(out, ""), "more") {
		t.Fatalf("%v", out)
	}
}

func TestTheTopOfALongPanelSaysThereIsMoreBelow(t *testing.T) {
	out, _ := Window(theme.Default(), numbered(30), nil, 0, 10)
	if len(out) != 10 || out[0] != "line 0" || !strings.Contains(out[9], "↓ more") {
		t.Fatalf("%v", out)
	}
}

func TestTheMiddleSaysMoreOnBothSides(t *testing.T) {
	out, _ := Window(theme.Default(), numbered(30), nil, 10, 10)
	if !strings.Contains(out[0], "↑ more") || !strings.Contains(out[9], "↓ more") || out[1] != "line 11" {
		t.Fatalf("%v", out)
	}
}

func TestTheEndShowsTheLastLineAndNoDownMarker(t *testing.T) {
	out, _ := Window(theme.Default(), numbered(30), nil, 999, 10)
	if out[9] != "line 29" || !strings.Contains(out[0], "↑ more") {
		t.Fatalf("%v", out)
	}
}

func TestPagingNeverSkipsALine(t *testing.T) {
	const total, h = 40, 10
	seen := map[int]bool{}
	top := 0
	for i := 0; i < 20; i++ {
		for y := 0; y < h; y++ {
			seen[top+y] = true
		}
		top = ClampTop(top+PageStep(h), total, h)
	}
	for i := 0; i < total; i++ {
		if !seen[i] {
			t.Fatalf("line %d was never visible", i)
		}
	}
}

func TestHitsFollowTheWindow(t *testing.T) {
	_, hits := Window(theme.Default(), numbered(30), map[int]string{15: "a", 2: "b", 19: "c"}, 10, 10)
	if hits[5] != "a" || len(hits) != 1 {
		t.Fatalf("only the action in view, shifted, should remain (the one under the marker is dropped): %v", hits)
	}
}

func TestClampTop(t *testing.T) {
	if ClampTop(-3, 30, 10) != 0 || ClampTop(50, 30, 10) != 20 || ClampTop(4, 5, 10) != 0 {
		t.Fatal("clamp")
	}
}
