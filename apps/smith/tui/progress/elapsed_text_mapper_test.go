package progress

import (
	"testing"
	"time"
)

func TestElapsedIsSecondsThenMinutes(t *testing.T) {
	for d, want := range map[time.Duration]string{500 * time.Millisecond: "", 12 * time.Second: "12s", 75 * time.Second: "1m 15s", 605 * time.Second: "10m 05s"} {
		if got := Elapsed(d); got != want {
			t.Fatalf("%v: %q, want %q", d, got, want)
		}
	}
}

func TestWaitingAddsTheTimeOnlyOnceItIsNoticeable(t *testing.T) {
	if got := Waiting("Asking npm…", time.Second); got != "Asking npm…" {
		t.Fatalf("%q", got)
	}
	if got := Waiting("Asking npm…", 12*time.Second); got != "Asking npm… 12s" {
		t.Fatalf("%q", got)
	}
}
