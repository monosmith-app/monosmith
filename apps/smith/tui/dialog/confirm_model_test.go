package dialog

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func release() Confirm {
	return NewConfirm("Release", "This versions, publishes and tags every package with releasable commits.", []string{"release", "--yes"}).Place(100, 30)
}

func TestTheDialogShowsTheExactCommandFirst(t *testing.T) {
	text := strings.Join(release().Lines(theme.Default()), "\n")
	for _, want := range []string{"Release", "The command that will run:", "$ smith release --yes", "Run  Enter", "Cancel  Esc", "╭", "╯"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in\n%s", want, text)
		}
	}
}

func TestTheDialogIsCentredAndExactlyItsSize(t *testing.T) {
	c := release()
	r := c.Rect()
	if r.X != (100-r.W)/2 || r.Y != (30-r.H)/2 {
		t.Fatalf("%+v", r)
	}
	lines := c.Lines(theme.Default())
	if len(lines) != r.H {
		t.Fatalf("%d lines for height %d", len(lines), r.H)
	}
	for _, l := range lines {
		if pane.Width(l) != r.W {
			t.Fatalf("width %d != %d: %q", pane.Width(l), r.W, l)
		}
	}
}

func TestEnterAndYConfirmEscAndNCancel(t *testing.T) {
	c := release()
	if _, o := c.Key("enter"); !o.Done || !o.Confirmed {
		t.Fatalf("%+v", o)
	}
	if _, o := c.Key("y"); !o.Done || !o.Confirmed {
		t.Fatalf("%+v", o)
	}
	for _, k := range []string{"esc", "n", "ctrl+c"} {
		if _, o := c.Key(k); !o.Done || o.Confirmed {
			t.Fatalf("%s: %+v", k, o)
		}
	}
}

func TestTabMovesFocusToCancelSoEnterCancels(t *testing.T) {
	c, o := release().Key("tab")
	if o.Done {
		t.Fatal("tab only moves")
	}
	if _, o := c.Key("enter"); !o.Done || o.Confirmed {
		t.Fatalf("%+v", o)
	}
}

func TestClickingTheButtonsRunsOrCancels(t *testing.T) {
	c := release()
	r := c.Rect()
	y := r.Y + r.H - 2
	spans := c.buttonsRow(r.W - 4)
	if _, o := c.Click(r.X+spans[0][0]+2, y); !o.Done || !o.Confirmed {
		t.Fatalf("run: %+v", o)
	}
	if _, o := c.Click(r.X+spans[1][0]+2, y); !o.Done || o.Confirmed {
		t.Fatalf("cancel: %+v", o)
	}
	if _, o := c.Click(r.X+3, r.Y+2); o.Done {
		t.Fatal("a click on the text does nothing")
	}
	if _, o := c.Click(0, 0); !o.Done || o.Confirmed {
		t.Fatalf("a click outside cancels: %+v", o)
	}
}

func TestHoveringAButtonHighlightsIt(t *testing.T) {
	c := release()
	r := c.Rect()
	spans := c.buttonsRow(r.W - 4)
	if got := c.Hover(r.X+spans[1][0]+1, r.Y+r.H-2).hover; got != 1 {
		t.Fatalf("hover %d", got)
	}
	if got := c.Hover(0, 0).hover; got != -1 {
		t.Fatalf("hover %d", got)
	}
}

func TestALongCommandWrapsAndStillFitsASmallScreen(t *testing.T) {
	argv := []string{"add", "package", "a-very-long-package-name-indeed", "--lang", "python", "--stack", "team-python-with-everything"}
	for _, size := range [][2]int{{80, 24}, {40, 12}, {20, 8}, {1, 1}} {
		c := NewConfirm("Add project", "This creates files in the workspace.", argv).Place(size[0], size[1])
		for _, l := range c.Lines(theme.Default()) {
			if pane.Width(l) != c.Rect().W {
				t.Fatalf("%v: width %d != %d", size, pane.Width(l), c.Rect().W)
			}
		}
		if r := c.Rect(); r.X < 0 || r.Y < 0 {
			t.Fatalf("%v: %+v", size, r)
		}
	}
}
