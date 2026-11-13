package pane

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestDialogBodyPadsABodyThatFitsAndKeepsTheFootAtTheBottom(t *testing.T) {
	got := DialogBody(theme.Default(), numbered(2), []string{"keys"}, 20, 5, 0)
	want := []string{"line 0", "line 1", "", "", "keys"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("got %q", got)
	}
}

func TestDialogBodyScrollsALongBodyAboveThePinnedFoot(t *testing.T) {
	th := theme.Default()
	body := numbered(20)
	got := DialogBody(th, body, []string{"keys"}, 20, 6, 7)
	if len(got) != 6 || got[5] != "keys" {
		t.Fatalf("the foot must stay on the last row: %q", got)
	}
	if !strings.Contains(Strip(got[0]), "more") || !strings.Contains(Strip(got[4]), "more") {
		t.Errorf("the hidden lines must be marked above and below: %q", got)
	}
	for y := range 5 {
		line := BodyLine(y, 7, len(body), 5)
		if line < 0 {
			continue
		}
		if !strings.HasPrefix(Strip(got[y]), body[line]) {
			t.Errorf("row %d draws %q but maps to %q", y, Strip(got[y]), body[line])
		}
	}
}

func TestBodyLineSkipsTheMoreMarkersAndTheRowsPastTheBody(t *testing.T) {
	cases := []struct {
		name             string
		y, top, total, h int
		want             int
	}{
		{"fits", 2, 0, 4, 6, 2},
		{"past a short body", 5, 0, 4, 6, -1},
		{"marker above", 0, 3, 20, 6, -1},
		{"marker below", 5, 3, 20, 6, -1},
		{"inside", 2, 3, 20, 6, 5},
		{"first row with nothing above", 0, 0, 20, 6, 0},
		{"outside", 6, 0, 20, 6, -1},
	}
	for _, c := range cases {
		if got := BodyLine(c.y, c.top, c.total, c.h); got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, got, c.want)
		}
	}
}

func TestFollowLinesShowsEveryLineOfAWrappedChoice(t *testing.T) {
	const total, h = 30, 8
	for top := 0; top <= total-h; top++ {
		for from := 0; from+2 < total; from++ {
			got := FollowLines(top, from, from+2, total, h)
			for line := from; line < from+2; line++ {
				if y := line - got; BodyLine(y, got, total, h) != line {
					t.Fatalf("top %d choice %d-%d: window %d hides line %d", top, from, from+2, got, line)
				}
			}
		}
	}
}
