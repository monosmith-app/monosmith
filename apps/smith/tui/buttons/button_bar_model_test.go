package buttons

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestAClickLandsOnTheButtonUnderIt(t *testing.T) {
	th := theme.Default()
	bar := Bar{{ID: "run", Label: "Run build", Icon: "▶", Shortcut: "Ctrl+B"}, {ID: "stop", Label: "Stop", Disabled: true}, {ID: "doctor", Label: "Doctor"}}
	// "[▶ Run build Ctrl+B]" is 20 cells wide, then a 2-cell gap.
	if id, ok := bar.At(th, 80, 0); id != "run" || !ok {
		t.Fatalf("%q %v", id, ok)
	}
	if id, ok := bar.At(th, 80, 19); id != "run" || !ok {
		t.Fatalf("%q %v", id, ok)
	}
	if id, _ := bar.At(th, 80, 20); id != "" {
		t.Fatalf("the gap is not a button: %q", id)
	}
	if id, ok := bar.At(th, 80, 23); id != "stop" || ok {
		t.Fatalf("a disabled button is found but not usable: %q %v", id, ok)
	}
}

func TestButtonsThatDoNotFitAreLeftOut(t *testing.T) {
	th := theme.Default()
	bar := Bar{{ID: "a", Label: "aaaaaaaaaa"}, {ID: "b", Label: "bbbbbbbbbb"}}
	if id, _ := bar.At(th, 15, 14); id != "" {
		t.Fatalf("%q", id)
	}
	if strings.Contains(bar.Render(th, 0, "", 15), "bbb") {
		t.Fatal("b should not be drawn")
	}
}

func TestABusyButtonShowsTheSpinner(t *testing.T) {
	th := theme.Default()
	bar := Bar{{ID: "run", Label: "Run", Icon: "▶", Busy: true}}
	if got := bar.Render(th, 1, "", 20); !strings.Contains(got, th.Icons.Spinner[1]) || strings.Contains(got, "▶") {
		t.Fatalf("%q", got)
	}
}
