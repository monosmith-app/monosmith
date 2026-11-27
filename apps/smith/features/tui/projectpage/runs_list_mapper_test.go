package projectpage

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

func TestTheRunsListHasAMarkTheTaskTheResultTheTimeAndTheAge(t *testing.T) {
	rows := []RunRow{{Mark: Failed, Task: "test", Result: "2 failed", Time: "21.6s", Ago: "4 min ago"}, {Mark: Passed, Task: "build", Result: "passed", Time: "4.1s", Ago: "yesterday"}}
	lines, rowOf := RunList(plain, rows, 70, 1, true, 0)
	if first := pane.Strip(lines[0]); !strings.Contains(first, "× test") || !strings.Contains(first, "2 failed") || !strings.Contains(first, "4 min ago") {
		t.Fatalf("first run %q", first)
	}
	if !strings.Contains(lines[1], Pointer) || strings.Contains(lines[0], Pointer) || rowOf[1] != 1 {
		t.Fatalf("the cursor is on the second run: %q", lines)
	}
	if none, _ := RunList(plain, nil, 70, 0, true, 0); !strings.Contains(none[0], "No run") {
		t.Fatalf("no runs: %q", none)
	}
}
