package depsview

import (
	"strings"
	"testing"
	"time"
)

func TestWaitingForTheRegistriesShowsHowLongItHasTaken(t *testing.T) {
	got := strings.Join(render(Model{Waited: 12 * time.Second}, 100, 20), "\n")
	if !strings.Contains(got, "newer releases of every dependency. 12s") {
		t.Fatalf("no elapsed time:\n%s", got)
	}
	if quick := strings.Join(render(Model{Waited: time.Second}, 100, 20), "\n"); strings.Contains(quick, "1s") {
		t.Fatalf("a wait of one second is not worth a number:\n%s", quick)
	}
}
