package adoptwizard

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
)

func TestTheCheckedTextSaysWhatWasProvedAndHowLongItTook(t *testing.T) {
	got := checkedText([]checkedStep{
		{Name: "install the packages and refresh package-lock.json", Took: 30 * time.Second},
		{Name: removalplan.StepBuild, Took: 9800 * time.Millisecond},
		{Name: removalplan.StepTest, Took: 31 * time.Second},
		{Name: removalplan.StepDoctor, Took: 125 * time.Second},
	})
	want := "everything builds (9.8s), every test passes (31s), doctor finds nothing wrong (2m05s)"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if checkedText(nil) != "" {
		t.Fatal("no check ran, so there is nothing to say")
	}
}

func TestTheStepTimerTimesEachStepFromItsLinesAndPassesThemOn(t *testing.T) {
	var out bytes.Buffer
	timer := newStepTimer(&out)
	clock := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	timer.now = func() time.Time { return clock }
	write := func(line string, advance time.Duration) {
		clock = clock.Add(advance)
		if _, err := timer.Write([]byte(line)); err != nil {
			t.Fatal(err)
		}
	}
	write("… build every project\n", 0)
	write("✔ build every project\n", 9*time.Second)
	write("… run every project's tests\n", time.Second)
	write("✔ run every pro", 10*time.Second) // a line can arrive in pieces
	write("ject's tests\n", 10*time.Second)
	steps := timer.Steps()
	if len(steps) != 2 || steps[0].Took != 9*time.Second || steps[1].Took != 20*time.Second {
		t.Fatalf("steps %+v", steps)
	}
	if !strings.Contains(out.String(), "… build every project\n") {
		t.Fatalf("the shell must still see every line: %q", out.String())
	}
}
